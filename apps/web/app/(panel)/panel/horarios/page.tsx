"use client";

import { useEffect, useRef, useState } from "react";
import { EditorSemana } from "@/componentes/editor-semana";
import {
  ErrorAPI,
  pedir,
  type HorarioSemanal,
  type ListaHorarios,
  type ListaTurnosDeProfesional,
  type TurnoConPaciente,
} from "@/lib/api";
import { formatearDia, formatearHora } from "@/lib/formato";
import { usePanel } from "@/lib/panel";
import { turnosHuerfanos } from "@/lib/semana";

/** Cuántos minutos dura un bloque. Las horas son de reloj, no instantes. */
function minutos(h: HorarioSemanal): number {
  const aMinutos = (hhmm: string) => {
    const [horas, mins] = hhmm.split(":").map(Number);
    return horas * 60 + mins;
  };
  return aMinutos(h.hasta) - aMinutos(h.desde);
}

export default function Horarios() {
  const { perfil } = usePanel();
  const [horarios, setHorarios] = useState<HorarioSemanal[] | null>(null);
  const [huerfanos, setHuerfanos] = useState<TurnoConPaciente[] | null>(null);
  const [aviso, setAviso] = useState<string | null>(null);
  const [guardando, setGuardando] = useState(false);
  const dialogo = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    let vigente = true;
    pedir<ListaHorarios>(`/api/v1/profesionales/${perfil.id}/horarios`).then((r) => {
      if (vigente) setHorarios(r.horarios);
    });
    return () => {
      vigente = false;
    };
  }, [perfil.id]);

  useEffect(() => {
    if (huerfanos) dialogo.current?.showModal();
    else dialogo.current?.close();
  }, [huerfanos]);

  /**
   * Antes de guardar, mirar qué turnos quedarían afuera.
   *
   * Sin esto, un profesional que acorta un bloque descubre que canceló seis
   * turnos recién cuando ve el número en la respuesta — cuando los pacientes ya
   * lo vieron en su app.
   */
  async function intentarGuardar() {
    if (!horarios) return;
    setAviso(null);

    // Primero lo que se puede saber sin la API. Pedirle a alguien que
    // sacrifique un turno para un guardado que va a fallar igual es la peor
    // versión de este diálogo.
    const invalido = horarios.find((h) => h.hasta <= h.desde || minutos(h) < h.duracionMin);
    if (invalido) {
      setAviso(
        invalido.hasta <= invalido.desde
          ? `El bloque de ${invalido.desde} a ${invalido.hasta} termina antes de empezar.`
          : `En el bloque de ${invalido.desde} a ${invalido.hasta} no entra ninguna sesión de ${invalido.duracionMin} minutos.`,
      );
      return;
    }

    const turnos = await pedir<ListaTurnosDeProfesional>(
      `/api/v1/profesionales/${perfil.id}/turnos`,
    );
    const afectados = turnosHuerfanos(turnos.datos, horarios);

    if (afectados.length > 0) {
      setHuerfanos(afectados);
      return;
    }
    await guardar();
  }

  async function guardar() {
    if (!horarios) return;
    setGuardando(true);
    setHuerfanos(null);

    try {
      const respuesta = await pedir<ListaHorarios>(
        `/api/v1/profesionales/${perfil.id}/horarios`,
        { method: "PUT", body: JSON.stringify({ horarios }) },
      );
      setHorarios(respuesta.horarios);

      // El número REAL que devolvió la API, no el que calculamos. Si no
      // coinciden, se ve: es la segunda mitigación de haber duplicado la regla
      // en el front.
      setAviso(
        respuesta.turnosCancelados === 0
          ? "Horarios guardados. No se canceló ningún turno."
          : `Horarios guardados. Se cancelaron ${respuesta.turnosCancelados} ${
              respuesta.turnosCancelados === 1 ? "turno" : "turnos"
            }.`,
      );
    } catch (e) {
      // El mensaje de la API y no uno genérico: "no pudimos guardar" no le
      // dice al profesional qué bloque arreglar, y es lo único que necesita.
      const detalle =
        e instanceof ErrorAPI && e.problema.errores?.length
          ? e.problema.errores.map((x) => x.mensaje).join(". ")
          : null;
      setAviso(detalle ? `No se guardó: ${detalle}.` : "No pudimos guardar los horarios. Probá de nuevo.");
    } finally {
      setGuardando(false);
    }
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <h1 className="text-2xl font-black tracking-tight">Horarios</h1>
      <p className="mt-1 text-tinta-suave">
        Los huecos que ve un paciente salen de acá.
      </p>

      {aviso && (
        <p
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso}
        </p>
      )}

      {!horarios ? (
        <p className="mt-6 text-tinta-suave">Cargando…</p>
      ) : (
        <>
          <div className="mt-6">
            <EditorSemana horarios={horarios} onCambiar={setHorarios} />
          </div>

          <div className="mt-6 flex justify-end">
            <button
              type="button"
              onClick={intentarGuardar}
              disabled={guardando}
              className="rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
            >
              {guardando ? "Guardando…" : "Guardar"}
            </button>
          </div>
        </>
      )}

      <dialog
        ref={dialogo}
        onClose={() => setHuerfanos(null)}
        className="m-auto w-[min(32rem,calc(100vw-2rem))] rounded-xl border border-borde bg-superficie p-6 text-tinta backdrop:bg-tinta/40"
      >
        {huerfanos && (
          <>
            <h2 className="text-lg font-bold tracking-tight">
              {huerfanos.length === 1
                ? "Con este cambio se cancela 1 turno ya reservado"
                : `Con este cambio se cancelan ${huerfanos.length} turnos ya reservados`}
            </h2>
            <p className="mt-1 text-sm text-tinta-suave">
              Los pacientes lo van a ver en la app.
            </p>

            {/* Se listan y no se cuentan: si el cálculo estuviera mal, la
                persona ve turnos concretos y puede juzgar. */}
            <ul className="mt-4 max-h-64 overflow-y-auto rounded-lg border border-borde">
              {huerfanos.map((t) => (
                <li
                  key={t.id}
                  className="flex flex-wrap items-baseline gap-x-3 border-b border-borde px-3 py-2 text-sm last:border-b-0"
                >
                  <span className="font-semibold tabular-nums">{formatearHora(t.inicio)}</span>
                  <span className="text-tinta-suave">{formatearDia(t.inicio)}</span>
                  <span>
                    {t.paciente.nombre} {t.paciente.apellido}
                  </span>
                </li>
              ))}
            </ul>

            <div className="mt-6 flex justify-end gap-2">
              {/* El botón destructivo no lleva el color de acento: el acento es
                  de la acción principal, y acá la principal es no romper
                  nada. */}
              <button
                type="button"
                onClick={() => setHuerfanos(null)}
                className="rounded-lg bg-accion px-4 py-2 text-sm font-bold text-white transition-colors hover:bg-accion-viva"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={guardar}
                className="rounded-lg border border-destructive px-4 py-2 text-sm font-bold text-destructive transition-colors hover:bg-destructive/10"
              >
                Guardar igual
              </button>
            </div>
          </>
        )}
      </dialog>
    </main>
  );
}
