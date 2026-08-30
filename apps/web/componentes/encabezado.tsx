import Link from "next/link";

/**
 * La navegación de toda la app.
 *
 * Existe porque sin ella no había forma de llegar a "mis turnos" desde ninguna
 * pantalla: el único link estaba en la confirmación de la reserva, así que
 * alguien que reservaba y volvía al día siguiente no podía encontrar su turno.
 * Los tests no lo encontraron porque navegan por links que ya conocen.
 *
 * Deliberadamente sin el color de acento: el acento es para la acción principal
 * de cada pantalla, y una barra de navegación no lo es en ninguna.
 */
export function Encabezado() {
  return (
    <header className="border-b border-borde bg-superficie">
      <nav
        aria-label="Principal"
        className="mx-auto flex w-full max-w-3xl items-center justify-between gap-4 px-4 py-3 sm:px-6"
      >
        <Link href="/" className="text-lg font-black tracking-tight">
          Salud
        </Link>

        <Link
          href="/turnos"
          className="rounded-lg px-3 py-2 text-sm font-semibold transition-colors hover:bg-muted"
        >
          Mis turnos
        </Link>
      </nav>
    </header>
  );
}
