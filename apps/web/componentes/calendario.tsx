import Link from "next/link";
import type { Hueco } from "@/lib/api";
import type { Dia } from "@/lib/dias";
import { Hora } from "./hora";

/**
 * La tira de días y los horarios del día elegido.
 *
 * Acepta links o callbacks porque las dos pantallas que lo usan eligen
 * distinto: el perfil navega —es un Server Component que tiene que renderizar
 * sin JavaScript y quedar indexado— y la reserva selecciona en memoria, porque
 * navegar ahí perdería el formulario a medio llenar.
 */
export function Calendario({
  dias,
  diaElegido,
  hrefDelDia,
  onDia,
  hrefDelHueco,
  onHueco,
}: {
  dias: Dia[];
  diaElegido: string;
  hrefDelDia?: (fecha: string) => string;
  onDia?: (fecha: string) => void;
  hrefDelHueco?: (hueco: Hueco) => string;
  onHueco?: (hueco: Hueco) => void;
}) {
  const elegido = dias.find((d) => d.fecha === diaElegido) ?? dias[0];
  const base =
    "flex w-16 shrink-0 flex-col items-center gap-0.5 rounded-lg border px-2 py-2.5 text-center transition-colors";

  return (
    <div className="rounded-xl border border-borde bg-superficie">
      {/* La tira scrollea sola en pantallas angostas y entra completa en las
          anchas. `snap` hace que un día quede alineado al borde en vez de
          cortado a la mitad. */}
      <ul className="flex snap-x snap-mandatory gap-2 overflow-x-auto p-3">
        {dias.map((dia) => {
          const activo = dia.fecha === elegido?.fecha;
          const etiquetas = (
            <>
              <span className="text-[11px] font-semibold uppercase tracking-wide opacity-80">
                {dia.etiqueta}
              </span>
              <span className="text-lg font-bold tabular-nums leading-none">
                {dia.numero}
              </span>
              {/* El punto dice "acá hay turnos" sin depender solo del color. */}
              <span
                aria-hidden="true"
                className={`mt-0.5 h-1 w-1 rounded-full ${
                  dia.huecos.length > 0 ? "bg-current" : "bg-transparent"
                }`}
              />
            </>
          );

          return (
            <li key={dia.fecha} className="snap-start">
              {dia.huecos.length === 0 ? (
                // Un día sin horarios no es un control: no hay nada a dónde ir,
                // y un botón que no hace nada enseña a desconfiar de los que sí
                // hacen. Se muestra igual porque "no atiende los martes" es
                // información.
                <span aria-disabled="true" className={`${base} border-borde opacity-40`}>
                  {etiquetas}
                </span>
              ) : (
                <Accion
                  href={hrefDelDia?.(dia.fecha)}
                  onClick={onDia && (() => onDia(dia.fecha))}
                  ariaCurrent={activo}
                  className={
                    activo
                      ? `${base} border-accion bg-accion text-white`
                      : `${base} border-borde hover:border-accion hover:bg-accent`
                  }
                >
                  {etiquetas}
                </Accion>
              )}
            </li>
          );
        })}
      </ul>

      <div className="border-t border-borde p-4">
        {/* El día completo va acá y no arriba del componente: la etiqueta que
            dice de qué día son los horarios tiene que estar al lado de los
            horarios. */}
        {elegido && (
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-tinta-suave">
            {elegido.largo}
          </h3>
        )}

        {elegido && elegido.huecos.length > 0 ? (
          <ul className="flex flex-wrap gap-2">
            {elegido.huecos.map((hueco) => (
              <li key={hueco.inicio}>
                <Accion
                  href={hrefDelHueco?.(hueco)}
                  onClick={onHueco && (() => onHueco(hueco))}
                  className="block rounded-lg border border-borde px-4 py-2.5 transition-colors hover:border-accion hover:bg-accent"
                >
                  <Hora inicio={hueco.inicio} />
                </Accion>
              </li>
            ))}
          </ul>
        ) : (
          // Un vacío es una invitación a hacer otra cosa, no un espacio en
          // blanco: dice qué pasó y qué hacer.
          <p className="py-2 text-sm text-tinta-suave">
            No atiende este día. Elegí otro de la tira.
          </p>
        )}
      </div>
    </div>
  );
}

function Accion({
  href,
  onClick,
  ariaCurrent,
  className,
  children,
}: {
  href?: string;
  onClick?: () => void;
  ariaCurrent?: boolean;
  className: string;
  children: React.ReactNode;
}) {
  if (href) {
    return (
      <Link href={href} aria-current={ariaCurrent ? "true" : undefined} className={className}>
        {children}
      </Link>
    );
  }
  return (
    <button
      type="button"
      onClick={onClick}
      aria-current={ariaCurrent ? "true" : undefined}
      className={className}
    >
      {children}
    </button>
  );
}
