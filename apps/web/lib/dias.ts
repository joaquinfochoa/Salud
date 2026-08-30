import { comoFecha } from "./formato";

const ZONA = "America/Argentina/Buenos_Aires";

export type Dia<T> = {
  /** `2026-09-07`, que es también el valor del parámetro `dia` en la URL. */
  fecha: string;
  /** `Lun` */
  etiqueta: string;
  /** `31` */
  numero: string;
  /** `lunes, 31 de agosto` */
  largo: string;
  /** Lo que cae ese día: huecos en el perfil público, turnos en el panel. */
  items: T[];
};

/**
 * Los próximos `cantidad` días a partir de hoy, con lo que cae en cada uno.
 *
 * Sirve para cualquier cosa que tenga `inicio` —huecos y turnos— porque partir
 * por día es la misma operación sobre las dos, y duplicarla es cómo se arreglan
 * los bugs de zona horaria en un lado y no en el otro.
 *
 * Devuelve **todos** los días, incluidos los vacíos. Saltearlos evitaría toques
 * muertos, pero rompe la metáfora: una tira de días salteados no se lee como un
 * calendario. Un día apagado además dice algo útil —"no atiende los martes"—
 * que un día ausente no dice.
 */
export function armarDias<T extends { inicio: string }>(
  items: T[],
  cantidad = 14,
  hoy = new Date(),
): Dia<T>[] {
  const porFecha = new Map<string, T[]>();
  for (const item of items) {
    const fecha = item.inicio.slice(0, 10);
    porFecha.set(fecha, [...(porFecha.get(fecha) ?? []), item]);
  }

  const formato = new Intl.DateTimeFormat("es-AR", {
    weekday: "short",
    day: "numeric",
    timeZone: ZONA,
  });
  const formatoLargo = new Intl.DateTimeFormat("es-AR", {
    weekday: "long",
    day: "numeric",
    month: "long",
    timeZone: ZONA,
  });

  const dias: Dia<T>[] = [];

  for (let i = 0; i < cantidad; i++) {
    const d = new Date(hoy);
    d.setDate(hoy.getDate() + i);

    const partes = formato.formatToParts(d);
    const diaDeSemana = partes.find((p) => p.type === "weekday")?.value ?? "";
    const fecha = comoFecha(d);

    dias.push({
      fecha,
      // es-AR abrevia como "lun." —minúscula y con punto—, que en una tira de
      // columnas angostas se lee peor que "Lun".
      etiqueta: diaDeSemana.replace(".", "").replace(/^./, (c) => c.toUpperCase()),
      numero: partes.find((p) => p.type === "day")?.value ?? "",
      largo: formatoLargo.format(d),
      items: porFecha.get(fecha) ?? [],
    });
  }

  return dias;
}

/** El primer día con algo, o el primero de todos si están todos vacíos. */
export function primerDiaConItems<T>(dias: Dia<T>[]): string {
  return (dias.find((d) => d.items.length > 0) ?? dias[0])?.fecha ?? "";
}
