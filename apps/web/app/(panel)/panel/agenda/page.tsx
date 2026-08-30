"use client";

import { useCallback, useEffect, useState } from "react";
import { Calendario } from "@/componentes/calendario";
import { DialogoBloqueo } from "@/componentes/dialogo-bloqueo";
import { FilaTurnoProfesional } from "@/componentes/fila-turno-profesional";
import {
  pedir,
  type ListaBloqueos,
  type ListaTurnosDeProfesional,
  type TurnoConPaciente,
} from "@/lib/api";
import { armarDias, primerDiaConItems } from "@/lib/dias";
import { comoFecha, formatearHora } from "@/lib/formato";
import { usePanel } from "@/lib/panel";

const DIAS_DE_VENTANA = 14;

export default function Agenda() {
  const { perfil } = usePanel();
  const [turnos, setTurnos] = useState<TurnoConPaciente[] | null>(null);
  const [bloqueos, setBloqueos] = useState<ListaBloqueos["datos"]>([]);
  const [pedido, setDia] = useState("");
  const [aviso, setAviso] = useState<string | null>(null);

  const traer = useCallback(async () => {
    const desde = new Date();
    const hasta = new Date(desde);
    hasta.setDate(desde.getDate() + DIAS_DE_VENTANA);
    const rango = `desde=${comoFecha(desde)}&hasta=${comoFecha(hasta)}`;

    const [t, b] = await Promise.all([
      pedir<ListaTurnosDeProfesional>(`/api/v1/profesionales/${perfil.id}/turnos?${rango}`),
      pedir<ListaBloqueos>(`/api/v1/profesionales/${perfil.id}/bloqueos?${rango}`),
    ]);
    return { turnos: t.datos, bloqueos: b.datos };
  }, [perfil.id]);

  useEffect(() => {
    let vigente = true;
    traer().then((d) => {
      if (!vigente) return;
      setTurnos(d.turnos);
      setBloqueos(d.bloqueos);
    });
    return () => {
      vigente = false;
    };
  }, [traer]);

  const dias = armarDias(turnos ?? [], DIAS_DE_VENTANA);
  // El día elegido se deriva: después de crear un bloqueo los turnos se
  // recargan, y el día que estaba abierto puede quedarse sin ninguno.
  const dia = dias.some((d) => d.fecha === pedido) ? pedido : primerDiaConItems(dias);
  const bloqueosDelDia = bloqueos.filter((b) => b.desde.slice(0, 10) === dia);

  async function recargar(mensaje: string) {
    setAviso(mensaje);
    const d = await traer();
    setTurnos(d.turnos);
    setBloqueos(d.bloqueos);
  }

  async function borrarBloqueo(id: string) {
    await pedir(`/api/v1/profesionales/${perfil.id}/bloqueos/${id}`, { method: "DELETE" });
    await recargar("Bloqueo borrado. Los horarios vuelven a estar disponibles.");
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-black tracking-tight">Agenda</h1>
        {dia && (
          <DialogoBloqueo profesionalID={perfil.id} dia={dia} onCreado={recargar} />
        )}
      </div>

      {aviso && (
        // role="alert" para que un lector de pantalla lo anuncie: si el bloqueo
        // canceló turnos, alguien que no ve la pantalla tiene que enterarse.
        <p
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso}
        </p>
      )}

      {!turnos ? (
        <p className="mt-6 text-tinta-suave">Cargando…</p>
      ) : (
        <div className="mt-6">
          <Calendario dias={dias} diaElegido={dia} onDia={setDia}>
            {(delDia) => (
              <>
                {delDia.length === 0 ? (
                  <p className="py-2 text-sm text-tinta-suave">No tenés turnos este día.</p>
                ) : (
                  <ul className="grid gap-2">
                    {delDia.map((turno) => (
                      <FilaTurnoProfesional key={turno.id} turno={turno} />
                    ))}
                  </ul>
                )}

                {bloqueosDelDia.length > 0 && (
                  <ul className="mt-3 grid gap-2 border-t border-borde pt-3">
                    {bloqueosDelDia.map((b) => (
                      <li
                        key={b.id}
                        className="flex flex-wrap items-baseline justify-between gap-2 rounded-lg border border-dashed border-borde px-3 py-2 text-sm"
                      >
                        <span className="text-tinta-suave">
                          Bloqueado {formatearHora(b.desde)}–{formatearHora(b.hasta)}
                          {b.motivo && ` · ${b.motivo}`}
                        </span>
                        <button
                          type="button"
                          onClick={() => borrarBloqueo(b.id)}
                          className="font-semibold text-tinta-suave underline hover:text-tinta"
                        >
                          Quitar
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </>
            )}
          </Calendario>
        </div>
      )}
    </main>
  );
}
