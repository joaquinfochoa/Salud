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

/**
 * Pesos escritos a mano a centavos. La inversa de `formatearPrecio`.
 *
 * Es el campo donde más fácil se cuela un error de dos órdenes de magnitud, y
 * por eso la conversión vive acá y no inline en el formulario: hay un test del
 * viaje de ida y vuelta, porque el campo muestra el valor formateado mientras
 * se escribe y la salida de una función vuelve a entrar por la otra.
 *
 * Devuelve `null` y no `0` para lo que no es un número: un precio vacío es "no
 * completó el campo", y guardar $0 como si lo hubiera elegido es peor que
 * pedirle que lo complete.
 */
export function enCentavos(pesos: string): number | null {
  // Se sacan el símbolo, los espacios y los puntos de miles; la coma decimal
  // pasa a punto. Es el formato es-AR, que es el que ve el profesional.
  const limpio = pesos.replace(/[$\s.]/g, "").replace(",", ".");
  if (limpio === "") return null;

  const numero = Number(limpio);
  if (!Number.isFinite(numero) || numero < 0) return null;
  return Math.round(numero * 100);
}
