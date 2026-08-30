import type { Profesional } from "./api";

/**
 * Los campos que acepta `POST` y `PUT /profesionales`, y solo esos.
 *
 * La API decodifica con `DisallowUnknownFields`, así que mandarle el
 * `Profesional` que devolvió —con `id`, `slug`, `estado`, `verificacion` y las
 * fechas— la hace responder 400. Es exactamente el error que cometí en dos
 * pantallas a la vez escribiendo `{ ...perfil, bio }`: se ve razonable, compila,
 * y falla recién contra el servidor.
 *
 * `cambios` pisa lo que venía del perfil.
 */
export function aPeticionProfesional(
  perfil: Profesional,
  cambios: Partial<Profesional> = {},
) {
  const p = { ...perfil, ...cambios };

  return {
    nombre: p.nombre,
    apellido: p.apellido,
    matricula: p.matricula,
    especialidad: p.especialidad,
    bio: p.bio,
    precioConsultaCentavos: p.precioConsultaCentavos,
    modalidades: p.modalidades,
    zona: p.zona,
    obrasSociales: p.obrasSociales,
    anticipacionMinimaMin: p.anticipacionMinimaMin,
    horizonteDias: p.horizonteDias,
  };
}
