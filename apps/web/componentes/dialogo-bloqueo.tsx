"use client";

import { useEffect, useRef, useState } from "react";
import { pedir, type BloqueoCreado } from "@/lib/api";

/**
 * Bloquear un rango del día que se está mirando.
 *
 * `<dialog>` nativo y no una librería: es la restricción de cero dependencias
 * nuevas, y `showModal()` ya trae foco atrapado, cierre con Escape y fondo
 * inerte — que es justamente lo que las librerías de modal implementan mal.
 */
export function DialogoBloqueo({
  profesionalID,
  dia,
  onCreado,
}: {
  profesionalID: string;
  /** `2026-09-07`, el día abierto en la agenda. */
  dia: string;
  onCreado: (mensaje: string) => void;
}) {
  const dialogo = useRef<HTMLDialogElement>(null);
  const [abierto, setAbierto] = useState(false);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // showModal() no se puede llamar en el render: hay que esperar a que el
  // elemento exista en el DOM.
  useEffect(() => {
    if (abierto) dialogo.current?.showModal();
    else dialogo.current?.close();
  }, [abierto]);

  async function crear(formulario: FormData) {
    setEnviando(true);
    setError(null);

    const desde = String(formulario.get("desde") ?? "");
    const hasta = String(formulario.get("hasta") ?? "");

    if (hasta <= desde) {
      setEnviando(false);
      setError("La hora de fin tiene que ser posterior a la de inicio.");
      return;
    }

    try {
      // El offset se arma acá: la API espera instantes, y el formulario da
      // horas de reloj. -03:00 es fijo — Argentina no cambia de horario desde
      // 2009, y el día que vuelva a cambiar esto se rompe visiblemente en vez
      // de correr los turnos una hora en silencio.
      const respuesta = await pedir<BloqueoCreado>(
        `/api/v1/profesionales/${profesionalID}/bloqueos`,
        {
          method: "POST",
          body: JSON.stringify({
            desde: `${dia}T${desde}:00-03:00`,
            hasta: `${dia}T${hasta}:00-03:00`,
            motivo: String(formulario.get("motivo") ?? ""),
          }),
        },
      );

      setAbierto(false);
      // Se informa siempre, incluso en cero: "0 turnos cancelados" es la
      // confirmación de que no rompiste nada, y sin ella hay que ir a mirar.
      onCreado(
        respuesta.turnosCancelados === 0
          ? "Bloqueo creado. No se canceló ningún turno."
          : `Bloqueo creado. Se cancelaron ${respuesta.turnosCancelados} ${
              respuesta.turnosCancelados === 1 ? "turno que caía" : "turnos que caían"
            } adentro.`,
      );
    } catch {
      setError("No pudimos crear el bloqueo. Probá de nuevo.");
    } finally {
      setEnviando(false);
    }
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setAbierto(true)}
        className="rounded-lg border border-borde px-4 py-2 text-sm font-semibold transition-colors hover:border-accion hover:bg-accent"
      >
        Bloquear un rato
      </button>

      <dialog
        ref={dialogo}
        onClose={() => setAbierto(false)}
        className="m-auto w-[min(24rem,calc(100vw-2rem))] rounded-xl border border-borde bg-superficie p-6 text-tinta backdrop:bg-tinta/40"
      >
        <h2 className="text-lg font-bold tracking-tight">Bloquear un rato</h2>
        <p className="mt-1 text-sm text-tinta-suave">
          Nadie va a poder reservar en ese rango.
        </p>

        <form action={crear} className="mt-5 grid gap-4">
          {error && (
            <p role="alert" className="rounded-lg border border-destructive px-3 py-2 text-sm text-destructive">
              {error}
            </p>
          )}

          <div className="grid grid-cols-2 gap-4">
            <Campo nombre="desde" etiqueta="Desde" tipo="time" valor="09:00" />
            <Campo nombre="hasta" etiqueta="Hasta" tipo="time" valor="13:00" />
          </div>
          <Campo nombre="motivo" etiqueta="Motivo" tipo="text" requerido={false} />

          <div className="mt-2 flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setAbierto(false)}
              className="rounded-lg px-4 py-2 text-sm font-semibold text-tinta-suave hover:text-tinta"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={enviando}
              className="rounded-lg bg-accion px-4 py-2 text-sm font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
            >
              {enviando ? "Bloqueando…" : "Bloquear"}
            </button>
          </div>
        </form>
      </dialog>
    </>
  );
}

function Campo({
  nombre,
  etiqueta,
  tipo,
  valor,
  requerido = true,
}: {
  nombre: string;
  etiqueta: string;
  tipo: string;
  valor?: string;
  requerido?: boolean;
}) {
  return (
    <div className="grid gap-1.5">
      <label htmlFor={nombre} className="text-sm font-semibold">
        {etiqueta}
      </label>
      <input
        id={nombre}
        name={nombre}
        type={tipo}
        required={requerido}
        defaultValue={valor}
        className="h-11 rounded-lg border border-borde px-3"
      />
    </div>
  );
}
