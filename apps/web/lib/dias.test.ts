import { describe, expect, it } from "vitest";
import type { Hueco } from "./api";
import { armarDias, primerDiaConItems } from "./dias";

// Un lunes al mediodía en Buenos Aires. Fijo para que los tests no cambien de
// resultado según el día que se corran.
const LUNES = new Date("2026-08-31T15:00:00Z");

function hueco(inicio: string): Hueco {
  return { inicio, fin: inicio, modalidad: "presencial" } as Hueco;
}

describe("armarDias", () => {
  it("devuelve días consecutivos desde hoy", () => {
    const dias = armarDias([], 3, LUNES);

    expect(dias.map((d) => d.fecha)).toEqual(["2026-08-31", "2026-09-01", "2026-09-02"]);
    expect(dias.map((d) => d.etiqueta)).toEqual(["Lun", "Mar", "Mié"]);
    expect(dias.map((d) => d.numero)).toEqual(["31", "1", "2"]);
    expect(dias[0].largo).toBe("lunes, 31 de agosto");
  });

  // Un día sin horarios se muestra apagado, no se saltea: una tira de días
  // salteados no se lee como un calendario.
  it("incluye los días vacíos", () => {
    const dias = armarDias([hueco("2026-09-02T09:00:00-03:00")], 3, LUNES);

    expect(dias.map((d) => d.items.length)).toEqual([0, 0, 1]);
  });

  it("agrupa los huecos por día", () => {
    const dias = armarDias(
      [
        hueco("2026-08-31T09:00:00-03:00"),
        hueco("2026-08-31T10:00:00-03:00"),
        hueco("2026-09-01T09:00:00-03:00"),
      ],
      2,
      LUNES,
    );

    expect(dias[0].items.map((h) => h.inicio)).toEqual([
      "2026-08-31T09:00:00-03:00",
      "2026-08-31T10:00:00-03:00",
    ]);
    expect(dias[1].items).toHaveLength(1);
  });

  // Los huecos vienen con el offset de Argentina, así que la fecha del ISO y la
  // del día tienen que coincidir sin importar en qué zona corra el proceso.
  it("no corre de día un hueco de la noche", () => {
    const dias = armarDias([hueco("2026-08-31T21:00:00-03:00")], 2, LUNES);

    expect(dias[0].items).toHaveLength(1);
    expect(dias[1].items).toHaveLength(0);
  });
});

describe("primerDiaConItems", () => {
  it("saltea los días vacíos", () => {
    const dias = armarDias([hueco("2026-09-02T09:00:00-03:00")], 5, LUNES);

    expect(primerDiaConItems(dias)).toBe("2026-09-02");
  });

  it("sin ningún hueco cae en el primer día", () => {
    expect(primerDiaConItems(armarDias([], 5, LUNES))).toBe("2026-08-31");
  });
});

// La agenda del profesional agrupa turnos, no huecos. Es la misma operación
// —partir por día— sobre otra entidad, y duplicarla es cómo se arreglan bugs de
// zona horaria en un lado y no en el otro.
describe("armarDias con cualquier cosa que tenga inicio", () => {
  it("agrupa turnos igual que huecos", () => {
    const turnos = [
      { inicio: "2026-08-31T09:00:00-03:00", motivo: "control" },
      { inicio: "2026-09-01T09:00:00-03:00", motivo: "primera vez" },
    ];

    const dias = armarDias(turnos, 2, LUNES);

    expect(dias[0].items[0].motivo).toBe("control");
    expect(dias[1].items[0].motivo).toBe("primera vez");
  });
});
