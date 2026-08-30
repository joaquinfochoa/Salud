import { describe, expect, it } from "vitest";
import type { HorarioSemanal, TurnoConPaciente } from "./api";
import { turnosHuerfanos } from "./semana";

const bloque = (
  diaSemana: HorarioSemanal["diaSemana"],
  desde: string,
  hasta: string,
): HorarioSemanal => ({
  diaSemana,
  desde,
  hasta,
  duracionMin: 50,
  modalidad: "presencial",
});

// 2026-08-31 es lunes; 2026-09-01, martes. Solo se completan los campos que
// turnosHuerfanos mira: un turno entero acá sería ruido que oculta qué importa.
const turno = (
  inicio: string,
  fin: string,
  estado: TurnoConPaciente["estado"] = "reservado",
) => ({ id: inicio, inicio, fin, estado }) as TurnoConPaciente;

describe("turnosHuerfanos", () => {
  it("un turno que entra en el horario nuevo no es huérfano", () => {
    const t = turno("2026-08-31T09:00:00-03:00", "2026-08-31T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "13:00")])).toEqual([]);
  });

  it("un turno que quedó fuera del bloque acortado sí lo es", () => {
    // Atendía hasta las 13; el bloque nuevo termina a las 10.
    const t = turno("2026-08-31T12:00:00-03:00", "2026-08-31T12:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "10:00")])).toHaveLength(1);
  });

  it("un turno de un día que se borró entero es huérfano", () => {
    const t = turno("2026-09-01T09:00:00-03:00", "2026-09-01T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "13:00")])).toHaveLength(1);
  });

  // El turno termina 09:50 y el bloque a las 09:30: entra el inicio y no el
  // final. La API lo cancela, así que la pantalla también tiene que verlo. Es
  // el caso que más fácil se escribe mal y el que más caro sale.
  it("un turno que empieza adentro pero termina afuera es huérfano", () => {
    const t = turno("2026-08-31T09:00:00-03:00", "2026-08-31T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "09:30")])).toHaveLength(1);
  });

  it("un turno ya cancelado no se cuenta dos veces", () => {
    const t = turno("2026-08-31T12:00:00-03:00", "2026-08-31T12:50:00-03:00", "cancelado");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "10:00")])).toEqual([]);
  });

  // Un día puede tener mañana y tarde. El turno de la tarde no es huérfano
  // porque el bloque de la mañana no lo contenga: alcanza con que ALGÚN bloque
  // de ese día lo contenga.
  it("basta con que entre en alguno de los bloques del día", () => {
    const t = turno("2026-08-31T16:00:00-03:00", "2026-08-31T16:50:00-03:00");
    const semana = [bloque("lunes", "09:00", "13:00"), bloque("lunes", "15:00", "19:00")];
    expect(turnosHuerfanos([t], semana)).toEqual([]);
  });

  it("sin ningún bloque, todos los activos son huérfanos", () => {
    const t = turno("2026-08-31T09:00:00-03:00", "2026-08-31T09:50:00-03:00");
    expect(turnosHuerfanos([t], [])).toHaveLength(1);
  });

  // El domingo es el índice 0 de getDay(), que es el error clásico de
  // desplazamiento en un arreglo de días.
  it("no confunde el domingo con el lunes", () => {
    // 2026-09-06 es domingo.
    const t = turno("2026-09-06T09:00:00-03:00", "2026-09-06T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("domingo", "09:00", "13:00")])).toEqual([]);
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "13:00")])).toHaveLength(1);
  });
});
