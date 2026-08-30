"use client";

import { createContext, useContext } from "react";
import type { Profesional, UsuarioActual } from "./api";

export type Panel = { usuario: UsuarioActual; perfil: Profesional };

const Contexto = createContext<Panel | null>(null);

export const ProveedorPanel = Contexto.Provider;

/**
 * El usuario y su perfil profesional, resueltos una sola vez en el layout.
 *
 * Existe por una razón concreta: las cuatro pantallas del panel necesitan el
 * perfil, y `/usuarios/yo` no trae ni la especialidad ni el slug —hay que ir a
 * `GET /profesionales/{id}` con el id que devuelve—, así que pedirlos en cada
 * pantalla son ocho llamadas por navegación.
 *
 * Es un objeto y un provider, no un gestor de estado.
 */
export function usePanel(): Panel {
  const valor = useContext(Contexto);
  if (!valor) {
    // Si esto explota, una pantalla del panel quedó fuera de (panel)/layout.
    throw new Error("usePanel() solo funciona dentro del layout del panel");
  }
  return valor;
}
