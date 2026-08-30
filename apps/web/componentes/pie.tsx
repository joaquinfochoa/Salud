import Link from "next/link";

/**
 * El pie del grupo público.
 *
 * Lleva el link a la landing de captación porque alguien que llega buscando un
 * turno y resulta ser profesional tiene que encontrar su camino: es el único
 * lugar de las páginas públicas donde ese camino aparece sin competir con la
 * búsqueda, que es lo que la home vino a hacer.
 */
export function Pie() {
  return (
    <footer className="mt-auto border-t border-borde bg-superficie">
      <div className="mx-auto w-full max-w-3xl px-4 py-8 sm:px-6">
        <p className="text-sm">
          <Link href="/profesionales" className="font-semibold hover:text-accion">
            ¿Sos profesional de la salud?
          </Link>{" "}
          <span className="text-tinta-suave">Publicá tu agenda y recibí turnos.</span>
        </p>

        {/* No es una formalidad legal: alguien con un cuadro agudo que entra a
            buscar turno tiene que leer esto antes de esperar tres días. */}
        <p className="mt-6 border-t border-borde pt-6 text-xs text-tinta-suave">
          Salud no reemplaza una consulta de urgencia. Ante una emergencia, llamá
          al 107 o acercate a una guardia.
        </p>
      </div>
    </footer>
  );
}
