import type { components } from "./contrato";

type Esquemas = components["schemas"];

// Los tipos salen del contrato, no de acá. `pnpm contrato` los regenera desde
// apps/api/api/openapi.yaml, y un cambio en el contrato rompe la compilación
// del front — que es exactamente cuando conviene enterarse.
export type Profesional = Esquemas["Profesional"];
export type ListaProfesionales = Esquemas["ListaProfesionales"];
export type Hueco = Esquemas["Hueco"];
export type ListaHuecos = Esquemas["ListaHuecos"];
export type Turno = Esquemas["Turno"];
export type ListaTurnos = Esquemas["ListaTurnos"];
export type Usuario = Esquemas["Usuario"];
export type UsuarioActual = Esquemas["UsuarioActual"];
export type Problema = Esquemas["Problema"];
export type Especialidad = Esquemas["Especialidad"];

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

/**
 * Un error que la API devolvió como `application/problem+json`.
 *
 * Conserva el estado HTTP además del mensaje porque las pantallas deciden con
 * él: la de reserva muestra cosas distintas ante un 409 y un 422, y con solo el
 * texto no puede distinguirlos.
 */
export class ErrorAPI extends Error {
  constructor(
    readonly estado: number,
    readonly problema: Problema,
  ) {
    super(problema.detail ?? problema.title);
    this.name = "ErrorAPI";
  }

  /**
   * El mensaje de un campo puntual de un 422.
   *
   * Existe para que el formulario los muestre debajo de cada input: un error de
   * validación que no está al lado de su campo obliga a la persona a adivinar
   * cuál falló.
   */
  porCampo(campo: string): string | undefined {
    return this.problema.errores?.find((e) => e.campo === campo)?.mensaje;
  }
}

/**
 * Todo el cliente de la API.
 *
 * No hay capa de recursos ni cliente generado: son cientos de líneas para
 * envolver `fetch`, y el contrato ya da los tipos. Cuando una pantalla necesita
 * algo, llama a esto con la ruta.
 */
export async function pedir<T>(ruta: string, opciones: RequestInit = {}): Promise<T> {
  const respuesta = await fetch(`${BASE}${ruta}`, {
    ...opciones,
    // Sin esto la cookie de sesión no viaja: el front y la API están en
    // orígenes distintos, y por eso el back tiene el middleware CORS.
    credentials: "include",
    headers: {
      // La API responde 415 a cualquier cuerpo que no llegue como JSON. Es
      // media defensa contra CSRF —la otra media es SameSite=Lax— no una
      // formalidad.
      ...(opciones.body ? { "Content-Type": "application/json" } : {}),
      ...opciones.headers,
    },
  });

  // 204 sin cuerpo: DELETE /sesiones/actual y DELETE /turnos/{id}.
  if (respuesta.status === 204) {
    return undefined as T;
  }

  const cuerpo = await respuesta.json();
  if (!respuesta.ok) {
    throw new ErrorAPI(respuesta.status, cuerpo as Problema);
  }
  return cuerpo as T;
}
