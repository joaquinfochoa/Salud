"use client";

import { useEffect, useRef } from "react";
import type { TurnoConPaciente } from "@/lib/api";
import { formatearDia, formatearHora } from "@/lib/formato";

/**
 * Confirmar antes de cancelarle el turno a un paciente.
 *
 * No es una formalidad: del otro lado hay alguien que reservó, que se organizó
 * y que lo va a ver en su app. Un click sin confirmación en una fila de una
 * lista es demasiado fácil de dar sin querer.
 *
 * `<dialog>` nativo, igual que los otros dos de la app: `showModal()` ya trae
 * foco atrapado, cierre con Escape y fondo inerte.
 */
export function DialogoCancelarTurno({
  turno,
  enviando,
  onCerrar,
  onConfirmar,
}: {
  turno: TurnoConPaciente | null;
  enviando: boolean;
  onCerrar: () => void;
  onConfirmar: () => void;
}) {
  const dialogo = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    if (turno) dialogo.current?.showModal();
    else dialogo.current?.close();
  }, [turno]);

  return (
    <dialog
      ref={dialogo}
      onClose={onCerrar}
      className="m-auto w-[min(26rem,calc(100vw-2rem))] rounded-xl border border-borde bg-superficie p-6 text-tinta backdrop:bg-tinta/40"
    >
      {turno && (
        <>
          <h2 className="text-lg font-bold tracking-tight">
            ¿Cancelar el turno de {turno.paciente.nombre}?
          </h2>
          <p className="mt-3 text-sm">
            <span className="font-semibold tabular-nums">
              {formatearHora(turno.inicio)}
            </span>{" "}
            <span className="text-tinta-suave">{formatearDia(turno.inicio)}</span>
          </p>
          <p className="mt-3 text-sm text-tinta-suave">
            {turno.paciente.nombre} {turno.paciente.apellido} lo va a ver
            cancelado en su app. El horario vuelve a quedar libre para que lo
            tome otra persona.
          </p>

          <div className="mt-6 flex justify-end gap-2">
            {/* El acento va en no cancelar: la acción principal acá es no
                romperle el turno a nadie. */}
            <button
              type="button"
              onClick={onCerrar}
              className="rounded-lg bg-accion px-4 py-2 text-sm font-bold text-white transition-colors hover:bg-accion-viva"
            >
              No, dejarlo
            </button>
            <button
              type="button"
              onClick={onConfirmar}
              disabled={enviando}
              className="rounded-lg border border-destructive px-4 py-2 text-sm font-bold text-destructive transition-colors hover:bg-destructive/10 disabled:opacity-60"
            >
              {enviando ? "Cancelando…" : "Sí, cancelar"}
            </button>
          </div>
        </>
      )}
    </dialog>
  );
}
