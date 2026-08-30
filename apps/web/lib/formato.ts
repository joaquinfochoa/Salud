// La API manda los montos en centavos y los instantes en ISO con la zona de
// Argentina. Estas tres funciones son el único lugar donde eso se traduce a
// algo legible: formatear inline en cada pantalla es cómo terminan apareciendo
// precios con dos órdenes de magnitud de diferencia entre una vista y otra.

const ZONA = "America/Argentina/Buenos_Aires";

/** Centavos a pesos. `formatearPrecio(1_200_000)` es `"$12.000"`. */
export function formatearPrecio(centavos: number): string {
  const conCentavos = centavos % 100 !== 0;

  return new Intl.NumberFormat("es-AR", {
    style: "currency",
    currency: "ARS",
    minimumFractionDigits: conCentavos ? 2 : 0,
    maximumFractionDigits: conCentavos ? 2 : 0,
  })
    .format(centavos / 100)
    // Intl mete un espacio duro entre el símbolo y el número; en una columna de
    // precios se ve como un error de alineación.
    .replace(/ /g, "");
}

/**
 * La hora del profesional, en 24 h.
 *
 * Siempre en la zona de Argentina, sin importar dónde esté el browser: un turno
 * a las 09:50 es a las 09:50 para las dos personas que lo acordaron. Dejarlo en
 * la zona local mostraría otra hora a un paciente que viaja.
 */
export function formatearHora(iso: string): string {
  return new Intl.DateTimeFormat("es-AR", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: ZONA,
  }).format(new Date(iso));
}

/** `"lunes, 7 de septiembre"`. Mismo argumento de zona que `formatearHora`. */
export function formatearDia(iso: string): string {
  return new Intl.DateTimeFormat("es-AR", {
    weekday: "long",
    day: "numeric",
    month: "long",
    timeZone: ZONA,
  }).format(new Date(iso));
}

/** `"2026-09-07"`, que es lo que esperan los parámetros `desde` y `hasta`. */
export function comoFecha(fecha: Date): string {
  return new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    timeZone: ZONA,
  }).format(fecha);
}
