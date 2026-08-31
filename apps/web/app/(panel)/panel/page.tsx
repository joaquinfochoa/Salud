"use client";

import { useEffect, useState } from "react";
import { DialogoCancelarTurno } from "@/componentes/dialogo-cancelar-turno";
import { FilaTurnoProfesional } from "@/componentes/fila-turno-profesional";
import { Nudge } from "@/componentes/nudge";
import { ESPECIALIDADES } from "@/lib/especialidades";
import {
  pedir,
  type ListaHuecos,
  type ListaTurnosDeProfesional,
  type TurnoConPaciente,
} from "@/lib/api";
import { comoFecha } from "@/lib/formato";
import { calcularOcupacion } from "@/lib/ocupacion";
import { usePanel } from "@/lib/panel";

const DIAS_DE_VENTANA = 7;

type Datos = {
  hoy: TurnoConPaciente[];
  semana: TurnoConPaciente[];
  huecosLibres: number;
};

export default function Hoy() {
  const { usuario, perfil } = usePanel();
  const [datos, setDatos] = useState<Datos | null>(null);
  const [aCancelar, setACancelar] = useState<TurnoConPaciente | null>(null);
  const [cancelando, setCancelando] = useState(false);
  const [recarga, setRecarga] = useState(0);

  useEffect(() => {
    let vigente = true;

    (async () => {
      const desde = new Date();
      const hasta = new Date(desde);
      hasta.setDate(desde.getDate() + DIAS_DE_VENTANA);
      const rango = `desde=${comoFecha(desde)}&hasta=${comoFecha(hasta)}`;

      // Las dos llamadas en paralelo: son independientes, y encadenarlas
      // duplicaría la espera de la pantalla que más se abre.
      const [turnos, huecos] = await Promise.all([
        pedir<ListaTurnosDeProfesional>(
          `/api/v1/profesionales/${perfil.id}/turnos?${rango}`,
        ),
        pedir<ListaHuecos>(`/api/v1/profesionales/${perfil.id}/huecos?${rango}`),
      ]);

      const fechaDeHoy = comoFecha(desde);
      if (vigente) {
        setDatos({
          semana: turnos.datos,
          hoy: turnos.datos.filter((t) => t.inicio.slice(0, 10) === fechaDeHoy),
          huecosLibres: huecos.datos.length,
        });
      }
    })();

    return () => {
      vigente = false;
    };
  }, [perfil.id, recarga]);

  async function cancelarTurno() {
    if (!aCancelar) return;
    setCancelando(true);
    try {
      await pedir(`/api/v1/turnos/${aCancelar.id}`, { method: "DELETE" });
      setACancelar(null);
      // Recargar y no parchear el estado: cancelar cambia los dos KPI además
      // de la fila, y recalcularlos a mano acá es cómo se desincronizan.
      setRecarga((n) => n + 1);
    } catch {
      // ponytail: sin aviso propio. Esta pantalla es de vistazo y el diálogo
      // ya se cerró; si falló, el turno sigue ahí y se ve. El aviso completo
      // está en /panel/agenda, que es donde se gestiona.
    } finally {
      setCancelando(false);
    }
  }

  const ocupacion = calcularOcupacion(datos?.semana ?? [], datos?.huecosLibres ?? 0);
  const activosDeLaSemana = (datos?.semana ?? []).filter((t) => t.estado === "reservado");
  const ahora = new Date().toISOString();

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <h1 className="text-2xl font-black tracking-tight">Hola, {usuario.nombre}</h1>
      <p className="mt-1 text-tinta-suave">
        {ESPECIALIDADES[perfil.especialidad] ?? perfil.especialidad} · {perfil.zona}
      </p>

      {/* El nudge va arriba de todo: si el perfil no puede recibir turnos, es
          lo único que importa de esta pantalla. */}
      {datos && (
        <div className="mt-6">
          <Nudge perfil={perfil} tieneHorarios={datos.huecosLibres > 0} />
        </div>
      )}

      <dl className="mt-6 grid grid-cols-2 gap-4">
        <Kpi etiqueta="Turnos esta semana" valor={datos ? `${activosDeLaSemana.length}` : "—"} />
        <Kpi
          etiqueta="Ocupación"
          valor={datos ? `${ocupacion.porcentaje}%` : "—"}
          detalle={datos ? `${ocupacion.tomados} de ${ocupacion.total} horarios` : undefined}
        />
      </dl>

      <h2 className="mt-10 text-lg font-bold tracking-tight">Hoy</h2>

      {!datos ? (
        <p className="mt-4 text-tinta-suave">Cargando…</p>
      ) : datos.hoy.length === 0 ? (
        <p className="mt-4 rounded-xl border border-borde bg-superficie p-8 text-center text-tinta-suave">
          No tenés turnos hoy.
        </p>
      ) : (
        <ul className="mt-4 grid gap-2">
          {datos.hoy.map((turno) => (
            <FilaTurnoProfesional
              key={turno.id}
              turno={turno}
              // Los que ya pasaron van en gris, como en el prototipo: lo que
              // importa de esta pantalla es lo que falta, no lo que fue.
              pasado={turno.fin < ahora}
              onCancelar={setACancelar}
            />
          ))}
        </ul>
      )}
      <DialogoCancelarTurno
        turno={aCancelar}
        enviando={cancelando}
        onCerrar={() => setACancelar(null)}
        onConfirmar={cancelarTurno}
      />
    </main>
  );
}

function Kpi({
  etiqueta,
  valor,
  detalle,
}: {
  etiqueta: string;
  valor: string;
  detalle?: string;
}) {
  return (
    <div className="rounded-xl border border-borde bg-superficie p-4">
      <dt className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        {etiqueta}
      </dt>
      <dd className="mt-1 text-3xl font-black tabular-nums tracking-tight">{valor}</dd>
      {detalle && <p className="mt-1 text-xs text-tinta-suave">{detalle}</p>}
    </div>
  );
}
