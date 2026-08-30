import { pedir, type Hueco, type ListaHuecos } from "./api";
import { comoFecha } from "./formato";

// Siete días, no el horizonte completo del profesional. GET /huecos exige un
// rango y devuelve todos los del período: pedir sesenta días para mostrar uno
// solo son cincuenta huecos de JSON por profesional, multiplicado por los
// resultados de la página.
const DIAS_DE_VENTANA = 7;

/**
 * El primer horario libre de un profesional en la próxima semana, o `null` si
 * no tiene ninguno.
 *
 * ponytail: esto es un N+1 contra la API. `GET /profesionales` no devuelve
 * huecos, así que el listado hace una llamada por resultado. El arreglo real es
 * un campo `proximoDisponible` calculado del lado del back, y es una etapa
 * chica: se hace cuando el listado se sienta lento o cuando exista PostgreSQL,
 * lo que llegue primero.
 */
export async function proximoHueco(profesionalID: string): Promise<Hueco | null> {
  const hoy = new Date();
  const hasta = new Date(hoy);
  hasta.setDate(hoy.getDate() + DIAS_DE_VENTANA);

  try {
    const respuesta = await pedir<ListaHuecos>(
      `/api/v1/profesionales/${profesionalID}/huecos` +
        `?desde=${comoFecha(hoy)}&hasta=${comoFecha(hasta)}`,
    );
    // Vienen ordenados por inicio desde el servidor.
    return respuesta.datos[0] ?? null;
  } catch {
    // Un profesional que todavía no cargó su horario no es un error de la
    // búsqueda: es una tarjeta sin horarios. Que un fallo suyo rompa el listado
    // entero sería mucho peor que no mostrar su disponibilidad.
    return null;
  }
}

/** Los huecos de un profesional en un rango, agrupados por día. */
export async function huecosPorDia(
  profesionalID: string,
  dias = 14,
): Promise<Map<string, Hueco[]>> {
  const hoy = new Date();
  const hasta = new Date(hoy);
  hasta.setDate(hoy.getDate() + dias);

  const respuesta = await pedir<ListaHuecos>(
    `/api/v1/profesionales/${profesionalID}/huecos` +
      `?desde=${comoFecha(hoy)}&hasta=${comoFecha(hasta)}`,
  );

  const porDia = new Map<string, Hueco[]>();
  for (const hueco of respuesta.datos) {
    const dia = hueco.inicio.slice(0, 10);
    porDia.set(dia, [...(porDia.get(dia) ?? []), hueco]);
  }
  return porDia;
}
