import Link from "next/link";
import type { Hueco, Profesional } from "@/lib/api";
import { formatearDia, formatearPrecio } from "@/lib/formato";
import { ESPECIALIDADES } from "@/lib/especialidades";
import { Hora } from "./hora";

const MODALIDADES: Record<string, string> = {
  telemedicina: "Videollamada",
  presencial: "En consultorio",
  domicilio: "A domicilio",
};

/**
 * Un resultado de la búsqueda.
 *
 * No muestra foto, rating ni reseñas: la API no los tiene y el front no los
 * inventa. Inventar prueba social es exactamente lo que un producto de salud no
 * puede hacer.
 */
export function TarjetaProfesional({
  profesional,
  proximo,
}: {
  profesional: Profesional;
  proximo: Hueco | null;
}) {
  const nombre = `${profesional.nombre} ${profesional.apellido}`;

  return (
    <li>
      <Link
        href={`/perfiles/${profesional.slug}`}
        className="block rounded-xl border border-borde bg-superficie p-5 transition-colors hover:border-accion/40"
      >
        <div className="flex items-start justify-between gap-5">
          <div className="flex min-w-0 items-start gap-4">
            <Iniciales nombre={profesional.nombre} apellido={profesional.apellido} />
            <div className="min-w-0">
              {/* h3 y no h2: la tarjeta vive dentro de una sección que ya tiene su
                  h2, y dos niveles iguales le dicen a un lector de pantalla que
                  el nombre de una persona es un apartado de la página. */}
              <h3 className="truncate text-lg font-bold tracking-tight">{nombre}</h3>
              <p className="text-sm text-tinta-suave">
                {ESPECIALIDADES[profesional.especialidad] ?? profesional.especialidad}
                {" · "}
                {profesional.zona}
              </p>
              <p className="mt-1 text-sm text-tinta-suave">
                {profesional.modalidades
                  .map((m) => MODALIDADES[m] ?? m)
                  .join(" · ")}
              </p>
            </div>
          </div>

          <p className="shrink-0 text-sm font-semibold tabular-nums">
            {formatearPrecio(profesional.precioConsultaCentavos)}
          </p>
        </div>

        <div className="mt-4 border-t border-borde pt-4">
          {proximo ? (
            <div className="flex items-baseline gap-3">
              <Hora inicio={proximo.inicio} />
              <span className="text-sm text-tinta-suave">
                {formatearDia(proximo.inicio)}
              </span>
            </div>
          ) : (
            // Un vacío es una invitación a hacer otra cosa, no un espacio en
            // blanco. Acá lo honesto es decir que todavía no publicó horarios.
            <p className="text-sm text-tinta-suave">Todavía no publicó horarios</p>
          )}
        </div>
      </Link>
    </li>
  );
}

// Sin fotos: la API no tiene el campo. Las iniciales no fingen nada y siguen
// dando algo que distinga una tarjeta de otra a la velocidad de un scroll.
function Iniciales({ nombre, apellido }: { nombre: string; apellido: string }) {
  return (
    <span
      aria-hidden="true"
      className="grid size-11 shrink-0 place-items-center rounded-full bg-accent text-sm font-bold text-accion"
    >
      {nombre[0]}
      {apellido[0]}
    </span>
  );
}
