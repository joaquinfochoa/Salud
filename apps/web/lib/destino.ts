/**
 * A dónde mandar a alguien después de que entra, sin dejar que la URL lo
 * decida por nosotros.
 *
 * Solo se aceptan rutas internas. Una redirección abierta
 * —`/entrar?volver=https://sitio-falso.com`— es cómo se arma un phishing que
 * empieza en tu propio dominio: el link se ve legítimo porque el dominio lo es.
 *
 * `//otro-sitio.com` se rechaza explícitamente: es una URL absoluta con
 * protocolo relativo, el browser la sigue igual que una con `https://`, y
 * empieza con barra, así que una comprobación ingenua la deja pasar.
 */
export function destinoSeguro(volver: string | null): string {
  if (!volver || !volver.startsWith("/") || volver.startsWith("//")) {
    return "/turnos";
  }
  return volver;
}
