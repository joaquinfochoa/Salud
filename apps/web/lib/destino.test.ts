import { describe, expect, it } from "vitest";
import { destinoSeguro } from "./destino";

describe("destinoSeguro", () => {
  it("acepta una ruta interna", () => {
    expect(destinoSeguro("/perfiles/martin-gonzalez/reservar")).toBe(
      "/perfiles/martin-gonzalez/reservar",
    );
  });

  it("conserva la query", () => {
    expect(destinoSeguro("/perfiles/x/reservar?inicio=2026-09-07T09%3A00%3A00-03%3A00")).toBe(
      "/perfiles/x/reservar?inicio=2026-09-07T09%3A00%3A00-03%3A00",
    );
  });

  it("cae en /turnos si no vino nada", () => {
    expect(destinoSeguro(null)).toBe("/turnos");
    expect(destinoSeguro("")).toBe("/turnos");
  });

  // Una redirección abierta es cómo se arma un phishing que empieza en tu
  // propio dominio: el link se ve legítimo porque el dominio lo es.
  it("rechaza una URL absoluta", () => {
    expect(destinoSeguro("https://sitio-falso.com")).toBe("/turnos");
    expect(destinoSeguro("http://sitio-falso.com")).toBe("/turnos");
  });

  // El caso que se olvida siempre: protocolo relativo. El browser lo sigue
  // igual que uno con https://, y empieza con barra, así que una comprobación
  // ingenua lo deja pasar.
  it("rechaza el protocolo relativo", () => {
    expect(destinoSeguro("//sitio-falso.com")).toBe("/turnos");
    expect(destinoSeguro("//sitio-falso.com/algo")).toBe("/turnos");
  });

  it("rechaza una ruta relativa sin barra", () => {
    expect(destinoSeguro("turnos")).toBe("/turnos");
  });
});
