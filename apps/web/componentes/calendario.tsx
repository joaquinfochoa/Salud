import Link from "next/link";
import type { Dia } from "@/lib/dias";

/**
 * La tira de días, y debajo lo que cae en el día elegido.
 *
 * Acepta links o callbacks porque las dos pantallas que la usan eligen
 * distinto: el perfil público navega —es un Server Component que tiene que
 * renderizar sin JavaScript y quedar indexado— y las pantallas con sesión
 * seleccionan en memoria, porque navegar ahí perdería lo que la persona está
 * haciendo.
 *
 * Qué se dibuja abajo lo decide quien la usa: un hueco es un chip de hora, un
 * turno es una fila con paciente y motivo.
 */
export function Calendario<T extends { inicio: string }>({
  dias,
  diaElegido,
  hrefDelDia,
  onDia,
  children,
}: {
  dias: Dia<T>[];
  diaElegido: string;
  hrefDelDia?: (fecha: string) => string;
  onDia?: (fecha: string) => void;
  children: (items: T[]) => React.ReactNode;
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
              {/* El punto dice "acá hay algo" sin depender solo del color. */}
              <span
                aria-hidden="true"
                className={`mt-0.5 h-1 w-1 rounded-full ${
                  dia.items.length > 0 ? "bg-current" : "bg-transparent"
                }`}
              />
            </>
          );

          return (
            <li key={dia.fecha} className="snap-start">
              {dia.items.length === 0 ? (
                // Un día vacío no es un control: no hay nada a dónde ir, y un
                // botón que no hace nada enseña a desconfiar de los que sí
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
                  // Sin esto un lector de pantalla lee "Lun31", que en una
                  // tira de catorce botones no dice de qué día se trata.
                  ariaLabel={dia.largo}
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
            dice de qué día es lo de abajo tiene que estar al lado de lo de
            abajo. */}
        {elegido && (
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-tinta-suave">
            {elegido.largo}
          </h3>
        )}
        {children(elegido?.items ?? [])}
      </div>
    </div>
  );
}

function Accion({
  href,
  onClick,
  ariaCurrent,
  ariaLabel,
  className,
  children,
}: {
  href?: string;
  onClick?: () => void;
  ariaCurrent?: boolean;
  ariaLabel?: string;
  className: string;
  children: React.ReactNode;
}) {
  if (href) {
    return (
      <Link
        href={href}
        aria-current={ariaCurrent ? "true" : undefined}
        aria-label={ariaLabel}
        className={className}
      >
        {children}
      </Link>
    );
  }
  return (
    <button
      type="button"
      onClick={onClick}
      aria-current={ariaCurrent ? "true" : undefined}
      aria-label={ariaLabel}
      className={className}
    >
      {children}
    </button>
  );
}
