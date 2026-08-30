import { describe, expect, it } from "vitest";
import type { Profesional } from "./api";
import { aPeticionProfesional } from "./perfil";

const perfil = {
  id: "abc-123",
  slug: "martin-gonzalez",
  nombre: "Martín",
  apellido: "González",
  matricula: "MN 98234",
  especialidad: "psicologia",
  bio: "Psicólogo clínico.",
  precioConsultaCentavos: 1200000,
  modalidades: ["telemedicina"],
  zona: "CABA",
  obrasSociales: ["OSDE"],
  estado: "activo",
  verificacion: "pendiente",
  creadoEn: "2026-08-01T10:00:00-03:00",
  actualizadoEn: "2026-08-20T10:00:00-03:00",
  dadoDeBajaEn: null,
  anticipacionMinimaMin: 120,
  horizonteDias: 60,
} as unknown as Profesional;

describe("aPeticionProfesional", () => {
  // El test que importa. La API decodifica con DisallowUnknownFields: un campo
  // de más y responde 400. Un `{ ...perfil }` compila igual y falla recién
  // contra el servidor.
  it("no manda ningún campo que la API no acepte", () => {
    expect(Object.keys(aPeticionProfesional(perfil)).sort()).toEqual([
      "anticipacionMinimaMin",
      "apellido",
      "bio",
      "especialidad",
      "horizonteDias",
      "matricula",
      "modalidades",
      "nombre",
      "obrasSociales",
      "precioConsultaCentavos",
      "zona",
    ]);
  });

  it("conserva los valores del perfil", () => {
    const p = aPeticionProfesional(perfil);
    expect(p.matricula).toBe("MN 98234");
    expect(p.precioConsultaCentavos).toBe(1200000);
    expect(p.obrasSociales).toEqual(["OSDE"]);
  });

  it("los cambios pisan lo que venía", () => {
    const p = aPeticionProfesional(perfil, { bio: "Nueva bio.", zona: "Rosario" });
    expect(p.bio).toBe("Nueva bio.");
    expect(p.zona).toBe("Rosario");
    // Y lo que no se cambió sigue igual.
    expect(p.nombre).toBe("Martín");
  });

  it("un cambio a un campo que no viaja no lo agrega", () => {
    const p = aPeticionProfesional(perfil, { slug: "otro-slug" } as Partial<Profesional>);
    expect(p).not.toHaveProperty("slug");
  });
});
