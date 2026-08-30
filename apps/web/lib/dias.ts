import type { Hueco } from "./api";
import { comoFecha } from "./formato";

const ZONA = "America/Argentina/Buenos_Aires";

export type Dia = {
  /** `2026-09-07`, que es también el valor del parámetro `dia` en la URL. */
  fecha: string;
  /** `Lun` */
  etiqueta: string;
  /** `31` */
  numero: string;
  /** `lunes, 31 de agosto` */
  largo: string;
  huecos: Hueco[];
};

/**
 * Los próximos `cantidad` días a partir de hoy, con los huecos de cada uno.
 *
 * Devuelve **todos** los días, incluidos los que no tienen horarios. Mostrar
 * solo los días con disponibilidad evitaría toques muertos, pero rompe la
 * metáfora: una tira de días salteados no se lee como un calendario. Un día
 * apagado además dice algo útil —"no atiende los martes"— que un día ausente no
 * dice.
 */
export function armarDias(huecos: Hueco[], cantidad = 14, hoy = new Date()): Dia[] {
  const porFecha = new Map<string, Hueco[]>();
  for (const hueco of huecos) {
    const fecha = hueco.inicio.slice(0, 10);
    porFecha.set(fecha, [...(porFecha.get(fecha) ?? []), hueco]);
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

  const dias: Dia[] = [];

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
      huecos: porFecha.get(fecha) ?? [],
    });
  }

  return dias;
}

/** El primer día con horarios, o el primero de todos si no hay ninguno. */
export function primerDiaConHuecos(dias: Dia[]): string {
  return (dias.find((d) => d.huecos.length > 0) ?? dias[0])?.fecha ?? "";
}
