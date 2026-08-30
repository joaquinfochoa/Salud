import { afterEach, describe, expect, it, vi } from "vitest";
import { ErrorAPI, pedir } from "./api";

afterEach(() => vi.unstubAllGlobals());

function respuestaFalsa(estado: number, cuerpo: unknown) {
  return vi.fn().mockResolvedValue({
    ok: estado >= 200 && estado < 300,
    status: estado,
    json: async () => cuerpo,
  });
}

describe("pedir", () => {
  it("devuelve el cuerpo cuando sale bien", async () => {
    vi.stubGlobal("fetch", respuestaFalsa(200, { datos: [] }));
    await expect(pedir("/api/v1/profesionales")).resolves.toEqual({ datos: [] });
  });

  it("manda las credenciales, o la cookie de sesion no viaja", async () => {
    const espia = respuestaFalsa(200, {});
    vi.stubGlobal("fetch", espia);

    await pedir("/api/v1/turnos");

    // El front y la API están en orígenes distintos. Sin credentials la cookie
    // no se manda y todo lo privado da 401 sin ninguna pista de por qué.
    expect(espia.mock.calls[0][1]).toMatchObject({ credentials: "include" });
  });

  it("pone Content-Type solo cuando hay cuerpo", async () => {
    const espia = respuestaFalsa(201, {});
    vi.stubGlobal("fetch", espia);

    // La API responde 415 a cualquier cuerpo que no llegue como JSON: es parte
    // de su defensa contra CSRF, no una formalidad.
    await pedir("/api/v1/sesiones", { method: "POST", body: "{}" });
    expect(espia.mock.calls[0][1].headers).toMatchObject({
      "Content-Type": "application/json",
    });

    await pedir("/api/v1/turnos");
    expect(espia.mock.calls[1][1].headers).not.toHaveProperty("Content-Type");
  });

  it("convierte un problema en ErrorAPI conservando el estado", async () => {
    vi.stubGlobal(
      "fetch",
      respuestaFalsa(409, {
        type: "https://salud.app/errors/conflict",
        title: "El turno ya fue tomado",
        status: 409,
        detail: "Alguien reservó ese horario mientras elegías. Probá con otro.",
      }),
    );

    // El número tiene que sobrevivir intacto: la pantalla de reserva muestra
    // cosas distintas ante un 409 y un 422, y sin el estado no puede
    // distinguirlos.
    await expect(pedir("/api/v1/turnos")).rejects.toMatchObject({
      estado: 409,
      problema: { title: "El turno ya fue tomado" },
    });
  });

  it("usa el detail como mensaje del error", async () => {
    vi.stubGlobal(
      "fetch",
      respuestaFalsa(422, {
        type: "https://salud.app/errors/validation",
        title: "Datos inválidos",
        status: 422,
        detail: "Uno o más campos no cumplen las reglas del sistema",
        errores: [{ campo: "contrasena", mensaje: "tiene que tener al menos 8 caracteres" }],
      }),
    );

    await expect(pedir("/api/v1/usuarios", { method: "POST", body: "{}" })).rejects.toThrow(
      "Uno o más campos no cumplen las reglas del sistema",
    );
  });

  it("no intenta parsear el cuerpo de un 204", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 204,
        json: async () => {
          throw new Error("no hay cuerpo que parsear");
        },
      }),
    );

    // DELETE /sesiones/actual y DELETE /turnos/{id} devuelven 204 sin cuerpo.
    await expect(
      pedir("/api/v1/sesiones/actual", { method: "DELETE" }),
    ).resolves.toBeUndefined();
  });
});

describe("ErrorAPI", () => {
  it("expone los errores por campo de un 422", () => {
    const e = new ErrorAPI(422, {
      type: "https://salud.app/errors/validation",
      title: "Datos inválidos",
      status: 422,
      errores: [{ campo: "email", mensaje: "no tiene un formato válido" }],
    });

    // El formulario los muestra debajo de cada campo, no en un cartel arriba.
    expect(e.porCampo("email")).toBe("no tiene un formato válido");
    expect(e.porCampo("nombre")).toBeUndefined();
  });
});
