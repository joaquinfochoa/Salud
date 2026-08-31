import { ESPECIALIDADES } from "@/lib/especialidades";

/**
 * Los tres filtros que la API soporta: especialidad, zona y búsqueda por
 * nombre. Ni uno más — inventar un cuarto sería prometer algo que no existe.
 *
 * Es un `<form>` con method GET, sin JavaScript. Los filtros viajan en la query
 * string, así que una búsqueda se comparte por link, sobrevive un refresh, y
 * esta página sigue siendo un Server Component.
 */
export function Filtros({
  valores,
}: {
  valores: { especialidad?: string; zona?: string; busqueda?: string };
}) {
  return (
    <form
      method="GET"
      action="/"
      className="grid gap-3 rounded-xl border border-borde bg-superficie p-4 sm:grid-cols-[1fr_1fr_1fr_auto]"
    >
      <Campo etiqueta="Especialidad">
        <select
          name="especialidad"
          defaultValue={valores.especialidad ?? ""}
          className="h-10 w-full rounded-lg border border-borde bg-superficie px-3 text-sm"
        >
          <option value="">Todas</option>
          {Object.entries(ESPECIALIDADES).map(([valor, texto]) => (
            <option key={valor} value={valor}>
              {texto}
            </option>
          ))}
        </select>
      </Campo>

      <Campo etiqueta="Zona">
        <input
          name="zona"
          defaultValue={valores.zona ?? ""}
          placeholder="CABA, GBA Norte…"
          className="h-10 w-full rounded-lg border border-borde bg-superficie px-3 text-sm"
        />
      </Campo>

      <Campo etiqueta="Nombre">
        <input
          name="busqueda"
          defaultValue={valores.busqueda ?? ""}
          placeholder="Buscar por nombre"
          className="h-10 w-full rounded-lg border border-borde bg-superficie px-3 text-sm"
        />
      </Campo>

      <div className="flex items-end">
        <button
          type="submit"
          className="h-10 w-full rounded-lg bg-accion px-5 text-sm font-bold text-white transition-colors hover:bg-accion-viva sm:w-auto"
        >
          Buscar
        </button>
      </div>
    </form>
  );
}

function Campo({ etiqueta, children }: { etiqueta: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        {etiqueta}
      </span>
      {children}
    </label>
  );
}
