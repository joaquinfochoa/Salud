import { formatearHora } from "@/lib/formato";

/**
 * La hora de un turno. Es el átomo de toda la interfaz: en la estructura que
 * elegimos, el horario es lo que la persona vino a buscar y por eso es el
 * elemento de mayor tamaño de cada fila.
 *
 * El estado se distingue por tres cosas a la vez —color, tachado y opacidad—
 * nunca solo por color. Ocho de cada cien varones no distinguen rojo de verde,
 * y esto es una app de salud.
 */
export function Hora({
  inicio,
  estado = "libre",
}: {
  /** El ISO que devuelve la API, sin convertir. */
  inicio: string;
  estado?: "libre" | "ocupado";
}) {
  const base = "font-horas text-3xl font-black tabular-nums tracking-tight";

  return (
    <span
      className={
        estado === "libre"
          ? `${base} text-libre`
          : `${base} text-apagado line-through decoration-2`
      }
    >
      {formatearHora(inicio)}
    </span>
  );
}
