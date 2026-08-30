/**
 * Si una URL de `?volver=` se puede seguir.
 *
 * Solo rutas internas. Una redirección abierta
 * —`/entrar?volver=https://sitio-falso.com`— es cómo se arma un phishing que
 * empieza en tu propio dominio: el link se ve legítimo porque el dominio lo es.
 *
 * `//otro-sitio.com` se rechaza explícitamente: es una URL absoluta con
 * protocolo relativo, el browser la sigue igual que una con `https://`, y
 * empieza con barra, así que una comprobación ingenua la deja pasar.
 */
function esInterna(volver: string | null): volver is string {
  return Boolean(volver) && volver!.startsWith("/") && !volver!.startsWith("//");
}

/** A dónde mandar a alguien, sin dejar que la URL lo decida por nosotros. */
export function destinoSeguro(volver: string | null): string {
  return esInterna(volver) ? volver : "/turnos";
}

/**
 * A dónde va alguien después de entrar.
 *
 * El `volver` gana siempre: si venía de reservar un turno, vuelve a reservar
 * aunque además sea profesional. La redirección por perfil es el default, no
 * una regla.
 *
 * Que el destino salga de `perfilProfesionalId` y no de un campo de rol es
 * consecuencia directa de no haber puesto un enum `Rol` en el dominio: se
 * deriva del estado real en vez de un dato que se puede desincronizar. La misma
 * persona puede ser profesional acá y paciente de otro profesional.
 *
 * La pregunta es si el `volver` era válido, y no si `destinoSeguro` devolvió
 * `/turnos`: con `volver=/turnos` las dos cosas coinciden, y un profesional que
 * pidió ir a sus turnos terminaría en el panel.
 */
export function destinoDespuesDeEntrar(
  volver: string | null,
  perfilProfesionalId: string | null,
): string {
  if (esInterna(volver)) return volver;
  return perfilProfesionalId ? "/panel" : "/turnos";
}
