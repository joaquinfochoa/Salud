"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

/**
 * Las secciones del paciente.
 *
 * Tabs arriba y no barra lateral como el panel del profesional, a propósito:
 * son dos usos distintos. El profesional trabaja acá adentro todos los días y
 * necesita una barra siempre presente; el paciente entra cada varios meses, a
 * ver o cancelar un turno, y una barra lateral para dos secciones es
 * andamiaje.
 */
const SECCIONES = [
  { href: "/turnos", texto: "Mis turnos" },
  { href: "/cuenta", texto: "Mi cuenta" },
] as const;

export function NavegacionPaciente() {
  const ruta = usePathname();

  return (
    <nav aria-label="Mi cuenta" className="border-b border-borde">
      <ul className="mx-auto flex w-full max-w-3xl gap-1 px-4 sm:px-6">
        {SECCIONES.map(({ href, texto }) => {
          const activo = ruta === href;
          return (
            <li key={href}>
              <Link
                href={href}
                aria-current={activo ? "page" : undefined}
                // El subrayado y no solo el color: el color nunca es la única
                // señal, y acá además marca dónde estás.
                className={`-mb-px block border-b-2 px-3 py-3 text-sm font-semibold transition-colors ${
                  activo
                    ? "border-accion text-accion"
                    : "border-transparent text-tinta-suave hover:text-tinta"
                }`}
              >
                {texto}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
