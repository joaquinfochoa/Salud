"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Logo } from "./logo";

/**
 * Una sola lista de rutas, dos presentaciones: tabs abajo en móvil, barra
 * lateral en escritorio.
 *
 * Son dos usos reales y los dos importan: mirar la agenda del día se hace en el
 * teléfono entre pacientes, y cargar la semana entera se hace sentado. Duplicar
 * la lista para las dos presentaciones es cómo se agrega un link en un lado y
 * no en el otro.
 */
const RUTAS = [
  { href: "/panel", texto: "Hoy" },
  { href: "/panel/agenda", texto: "Agenda" },
  { href: "/panel/horarios", texto: "Horarios" },
  { href: "/panel/perfil", texto: "Perfil" },
] as const;

export function NavegacionPanel() {
  const ruta = usePathname();

  return (
    <nav
      aria-label="Panel"
      className="fixed inset-x-0 bottom-0 z-10 border-t border-borde bg-superficie sm:static sm:w-52 sm:shrink-0 sm:border-r sm:border-t-0"
    >
      {/* El logo solo en escritorio: en móvil esto es la barra de tabs de
          abajo, y ahí no hay lugar ni motivo para la marca. Era el único lugar
          de la app donde no aparecía por ningún lado. */}
      <div className="hidden border-b border-borde px-6 py-4 sm:block">
        <Logo />
      </div>

      <ul className="flex sm:flex-col sm:gap-1 sm:p-3">
        {RUTAS.map(({ href, texto }) => {
          // Coincidencia exacta y no startsWith: con startsWith, "/panel"
          // quedaría activo en las cuatro pantallas.
          const activo = ruta === href;

          return (
            <li key={href} className="flex-1 sm:flex-none">
              <Link
                href={href}
                // Sin esto un lector de pantalla no dice en qué sección estás.
                aria-current={activo ? "page" : undefined}
                className={`block px-3 py-3 text-center text-sm font-semibold transition-colors sm:rounded-lg sm:text-left ${
                  activo
                    ? "text-accion sm:bg-accent"
                    : "text-tinta-suave hover:text-tinta sm:hover:bg-muted"
                }`}
              >
                {texto}
                {/* En móvil la barra abajo marca el activo además del color:
                    el color nunca es la única señal. */}
                <span
                  aria-hidden="true"
                  className={`mx-auto mt-1.5 block h-0.5 w-6 rounded-full sm:hidden ${
                    activo ? "bg-accion" : "bg-transparent"
                  }`}
                />
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
