import { describe, expect, it } from "vitest";
import { calcularOcupacion } from "./ocupacion";

const activo = (inicio: string) => ({ inicio, estado: "reservado" }) as never;
const cancelado = (inicio: string) => ({ inicio, estado: "cancelado" }) as never;

describe("calcularOcupacion", () => {
  it("cuenta los tomados sobre el total ofrecido", () => {
    // 3 tomados + 7 libres = 10 ofrecidos.
    expect(calcularOcupacion([activo("a"), activo("b"), activo("c")], 7)).toEqual({
      tomados: 3,
      total: 10,
      porcentaje: 30,
    });
  });

  // Un turno cancelado devolvió su hueco a la lista de libres, así que contarlo
  // como tomado lo contaría dos veces.
  it("no cuenta los cancelados", () => {
    expect(calcularOcupacion([activo("a"), cancelado("b")], 9)).toEqual({
      tomados: 1,
      total: 10,
      porcentaje: 10,
    });
  });

  // Un profesional recién registrado. 0/0 pondría "NaN%" en la pantalla.
  it("sin agenda cargada no divide por cero", () => {
    expect(calcularOcupacion([], 0)).toEqual({ tomados: 0, total: 0, porcentaje: 0 });
  });

  it("redondea a entero", () => {
    expect(calcularOcupacion([activo("a")], 2).porcentaje).toBe(33);
  });

  it("la agenda llena es 100", () => {
    expect(calcularOcupacion([activo("a"), activo("b")], 0).porcentaje).toBe(100);
  });
});
