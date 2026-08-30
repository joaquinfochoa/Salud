import { describe, expect, it } from "vitest";
import { formatearDia, formatearHora, formatearPrecio } from "./formato";

describe("formatearPrecio", () => {
  it("convierte centavos a pesos", () => {
    // La API manda 1_200_000 centavos. Mostrarlo tal cual sería un error de dos
    // órdenes de magnitud en la pantalla donde alguien decide si puede pagarlo.
    expect(formatearPrecio(1_200_000)).toBe("$12.000");
  });

  it("no muestra decimales cuando son cero", () => {
    expect(formatearPrecio(950_000)).toBe("$9.500");
  });

  it("muestra los centavos cuando no son cero", () => {
    expect(formatearPrecio(1_200_050)).toBe("$12.000,50");
  });

  it("maneja el cero", () => {
    // Cero es un precio válido en el dominio: el back usa un puntero
    // justamente para distinguirlo de "no lo mandaron".
    expect(formatearPrecio(0)).toBe("$0");
  });
});

describe("formatearHora", () => {
  it("devuelve la hora del profesional", () => {
    expect(formatearHora("2026-09-07T09:50:00-03:00")).toBe("09:50");
  });

  it("no corre la hora cuando el instante viene en UTC", () => {
    // Las 12:50 UTC son las 09:50 en Argentina. Si esto devolviera 12:50, cada
    // horario de la app estaría tres horas corrido.
    expect(formatearHora("2026-09-07T12:50:00Z")).toBe("09:50");
  });
});

describe("formatearDia", () => {
  // Intl mete la coma ("lunes, 7 de septiembre"). Se acepta tal cual: es la
  // data de locale y no vale un replace propio para discutirle una coma.
  it("escribe el día en español", () => {
    expect(formatearDia("2026-09-07T09:50:00-03:00")).toBe("lunes, 7 de septiembre");
  });

  it("usa la zona de Argentina para decidir qué día es", () => {
    // Las 02:00 UTC del 8 son las 23:00 del 7 en Argentina. Sin la zona, el
    // turno aparecería un día después del que es.
    expect(formatearDia("2026-09-08T02:00:00Z")).toBe("lunes, 7 de septiembre");
  });
});
