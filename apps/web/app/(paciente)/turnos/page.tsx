"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Hora } from "@/componentes/hora";
import { ESPECIALIDADES } from "@/lib/especialidades";
import {
  ErrorAPI,
  pedir,
  type ListaTurnosDePaciente,
  type TurnoConProfesional,
} from "@/lib/api";
import { formatearDia, formatearHora } from "@/lib/formato";

export default function MisTurnos() {
  const router = useRouter();
  const [turnos, setTurnos] = useState<TurnoConProfesional[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  const alFallar = useCallback(
    (e: unknown) => {
      if (e instanceof ErrorAPI && e.estado === 401) {
        // Sin sesión no hay nada que mostrar. El `volver` hace que después de
        // entrar vuelva acá y no a la home.
        router.replace("/entrar?volver=/turnos");
        return;
      }
      throw e;
    },
    [router],
  );

  const cargar = useCallback(async () => {
    try {
      const lista = await pedir<ListaTurnosDePaciente>("/api/v1/turnos");
      setTurnos(lista.datos);
    } catch (e) {
      alFallar(e);
    }
  }, [alFallar]);

  useEffect(() => {
    // El setState va en el callback de la promesa, no en el cuerpo del efecto:
    // así el render no cascadea, que es lo que pide la regla de React.
    //
    // Y `vigente` arregla algo que no es de estilo: sin él, una respuesta que
    // llega después de que la persona se fue de la pantalla llama a setState
    // sobre un componente desmontado.
    let vigente = true;

    pedir<ListaTurnosDePaciente>("/api/v1/turnos")
      .then((lista) => {
        if (vigente) setTurnos(lista.datos);
      })
      .catch((e) => {
        if (vigente) alFallar(e);
      });

    return () => {
      vigente = false;
    };
  }, [alFallar]);

  async function cancelar(turno: TurnoConProfesional) {
    // Cancelar un turno no se deshace, y del otro lado hay una persona que
    // reservó ese horario. Preguntar es más barato que un turno perdido.
    if (!confirm(`¿Cancelar el turno del ${formatearDia(turno.inicio)} a las ${formatearHora(turno.inicio)}?`)) {
      return;
    }

    setError(null);
    try {
      await pedir(`/api/v1/turnos/${turno.id}`, { method: "DELETE" });
      await cargar();
    } catch (e) {
      if (e instanceof ErrorAPI) {
        setError(e.message);
        return;
      }
      throw e;
    }
  }

  if (turnos === null) {
    return (
      <main className="mx-auto w-full max-w-xl px-4 py-16 sm:px-6">
        <p className="text-tinta-suave">Cargando tus turnos…</p>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-xl px-4 py-10 sm:px-6 sm:py-14">
      <h1 className="text-2xl font-black tracking-tight">Mis turnos</h1>

      {error && (
        <p role="alert" className="mt-4 rounded-lg border border-destructive px-4 py-3 text-sm text-destructive">
          {error}
        </p>
      )}

      {turnos.length === 0 ? (
        <div className="mt-6 rounded-xl border border-borde bg-superficie p-8 text-center">
          <p className="text-tinta-suave">Todavía no reservaste ningún turno.</p>
          <Link
            href="/"
            className="mt-5 inline-block rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva"
          >
            Buscar un profesional
          </Link>
        </div>
      ) : (
        <ul className="mt-6 grid gap-3">
          {turnos.map((turno) => (
            <FilaTurno key={turno.id} turno={turno} onCancelar={() => cancelar(turno)} />
          ))}
        </ul>
      )}
    </main>
  );
}

function FilaTurno({
  turno,
  onCancelar,
}: {
  turno: TurnoConProfesional;
  onCancelar: () => void;
}) {
  const cancelado = turno.estado === "cancelado";
  const yaPaso = new Date(turno.inicio) <= new Date();

  return (
    <li className="rounded-xl border border-borde bg-superficie p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-baseline gap-3">
            <Hora inicio={turno.inicio} estado={cancelado ? "ocupado" : "libre"} />
            <span className="text-sm text-tinta-suave">{formatearDia(turno.inicio)}</span>
          </div>

          {/* Con quién es el turno. Hasta esta etapa la API solo mandaba el id,
              así que esta pantalla mostraba una hora y nada más — que es la
              mitad de lo que un paciente necesita saber de acá. */}
          <p className="mt-2 font-semibold">
            <Link href={`/perfiles/${turno.profesional.slug}`} className="hover:text-accion">
              {turno.profesional.nombre} {turno.profesional.apellido}
            </Link>
          </p>
          <p className="text-sm text-tinta-suave">
            {ESPECIALIDADES[turno.profesional.especialidad] ?? turno.profesional.especialidad}
          </p>

          {turno.motivo && <p className="mt-2 text-sm text-tinta-suave">{turno.motivo}</p>}
        </div>

        {/* Los cancelados se muestran, marcados. Hasta que existan
            notificaciones, esta pantalla es el único lugar donde el paciente se
            entera de que el profesional le canceló un turno: esconderlos sería
            esconderle justo lo que necesita ver. */}
        {cancelado ? (
          <span className="shrink-0 rounded-md border border-apagado px-2.5 py-1 text-xs font-semibold uppercase tracking-wide text-tinta-suave">
            Cancelado
          </span>
        ) : (
          !yaPaso && (
            <button
              type="button"
              onClick={onCancelar}
              className="shrink-0 rounded-lg border border-borde px-3 py-2 text-sm font-semibold transition-colors hover:border-destructive hover:text-destructive"
            >
              Cancelar
            </button>
          )
        )}
      </div>
    </li>
  );
}
