import { describe, expect, it } from "vitest";
import { destinoDespuesDeEntrar, destinoSeguro } from "./destino";

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

describe("destinoDespuesDeEntrar", () => {
  // El volver gana siempre: alguien que venía de reservar tiene que volver a
  // reservar, tenga o no perfil profesional. La redirección por perfil es el
  // default, no una regla.
  it("respeta el volver por encima de todo", () => {
    expect(destinoDespuesDeEntrar("/perfiles/x/reservar", "abc-123")).toBe(
      "/perfiles/x/reservar",
    );
  });

  it("manda al panel si tiene perfil profesional", () => {
    expect(destinoDespuesDeEntrar(null, "abc-123")).toBe("/panel");
  });

  it("manda a mis turnos si no tiene", () => {
    expect(destinoDespuesDeEntrar(null, null)).toBe("/turnos");
  });

  // La comprobación de destinoSeguro sigue mandando: un volver externo se
  // descarta y recién ahí decide el perfil.
  it("sigue rechazando un volver externo", () => {
    expect(destinoDespuesDeEntrar("https://sitio-falso.com", "abc-123")).toBe("/panel");
    expect(destinoDespuesDeEntrar("//sitio-falso.com", null)).toBe("/turnos");
  });

  // El caso que rompe la implementación ingenua: si destinoSeguro devuelve
  // "/turnos" porque no vino nada, no hay forma de distinguirlo de un volver
  // que decía "/turnos" de verdad. Un profesional que pidió ir a sus turnos
  // tiene que ir a sus turnos, no al panel.
  it("un volver que dice /turnos no termina en el panel", () => {
    expect(destinoDespuesDeEntrar("/turnos", "abc-123")).toBe("/turnos");
  });
});
