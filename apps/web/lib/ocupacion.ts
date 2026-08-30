import type { Turno } from "./api";

/**
 * Cuántos de los turnos que el profesional ofrece están tomados.
 *
 * Es el KPI que reemplaza a los dos del prototipo —"cobrado hoy" y
 * "completados"—, que necesitan pagos y `atendido`/`ausente` y no existen.
 * Este sale de datos que sí tenemos y dice algo que un profesional
 * independiente mira de verdad: si le sobra agenda o le falta.
 *
 * `huecosLibres` son los que quedaron sin reservar. El total ofrecido es la
 * suma de los dos: un turno tomado ya no aparece como hueco.
 */
export function calcularOcupacion(turnos: Turno[], huecosLibres: number) {
  // Un turno cancelado devolvió su hueco a la lista de libres. Contarlo como
  // tomado lo contaría dos veces.
  const tomados = turnos.filter((t) => t.estado === "reservado").length;
  const total = tomados + huecosLibres;

  return {
    tomados,
    total,
    // Sin agenda cargada, 0/0 pondría "NaN%" en la pantalla.
    porcentaje: total === 0 ? 0 : Math.round((tomados / total) * 100),
  };
}
