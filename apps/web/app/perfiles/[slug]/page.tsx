import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Hora } from "@/componentes/hora";
import { ESPECIALIDADES } from "@/componentes/tarjeta-profesional";
import { ErrorAPI, pedir, type Profesional } from "@/lib/api";
import { formatearDia, formatearPrecio } from "@/lib/formato";
import { huecosPorDia } from "@/lib/huecos";

const MODALIDADES: Record<string, string> = {
  telemedicina: "Videollamada",
  presencial: "En consultorio",
  domicilio: "A domicilio",
};

async function traerPerfil(slug: string): Promise<Profesional> {
  try {
    return await pedir<Profesional>(`/api/v1/perfiles/${slug}`);
  } catch (e) {
    if (e instanceof ErrorAPI && e.estado === 404) notFound();
    throw e;
  }
}

// Estos metadatos son el producto de haber elegido Next. Un perfil sin title ni
// description propios no se distingue de ningún otro en un resultado de
// búsqueda, y la búsqueda orgánica es el canal por el que llegan los pacientes.
//
// La especialidad va en el title a propósito, y no contradice la regla de no
// mostrar datos de salud: acá es del PROFESIONAL, es información que él
// publica, y es lo que hace que la página se encuentre. Lo que nunca sale es el
// motivo de consulta de un paciente.
export async function generateMetadata({
  params,
}: PageProps<"/perfiles/[slug]">): Promise<Metadata> {
  const { slug } = await params;
  const p = await traerPerfil(slug);
  const especialidad = ESPECIALIDADES[p.especialidad] ?? p.especialidad;

  return {
    title: `${p.nombre} ${p.apellido} — ${especialidad} en ${p.zona}`,
    description: p.bio.slice(0, 155),
    alternates: { canonical: `/perfiles/${slug}` },
  };
}

// Server Component, y tiene que seguir siéndolo.
export default async function Perfil({ params }: PageProps<"/perfiles/[slug]">) {
  const { slug } = await params;
  const p = await traerPerfil(slug);
  const porDia = await huecosPorDia(p.id);

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <Link href="/" className="text-sm text-tinta-suave hover:text-accion">
        ← Volver a la búsqueda
      </Link>

      <header className="mt-6 rounded-xl border border-borde bg-superficie p-6">
        <h1 className="text-3xl font-black tracking-tight">
          {p.nombre} {p.apellido}
        </h1>
        <p className="mt-1 text-tinta-suave">
          {ESPECIALIDADES[p.especialidad] ?? p.especialidad} · {p.zona}
        </p>

        <p className="mt-3 inline-flex items-center gap-2 rounded-md border border-borde px-2.5 py-1">
          <span className="text-sm font-semibold tabular-nums">{p.matricula}</span>
          <span className="text-xs uppercase tracking-wide text-tinta-suave">
            matrícula
          </span>
        </p>

        {p.bio && <p className="mt-5 max-w-prose leading-relaxed">{p.bio}</p>}

        <dl className="mt-6 grid gap-4 border-t border-borde pt-5 sm:grid-cols-3">
          <Dato etiqueta="Consulta">{formatearPrecio(p.precioConsultaCentavos)}</Dato>
          <Dato etiqueta="Modalidad">
            {p.modalidades.map((m) => MODALIDADES[m] ?? m).join(" · ")}
          </Dato>
          <Dato etiqueta="Obras sociales">
            {p.obrasSociales.length > 0 ? p.obrasSociales.join(" · ") : "Solo particular"}
          </Dato>
        </dl>
      </header>

      <section className="mt-8">
        <h2 className="text-xl font-bold tracking-tight">Elegí un horario</h2>

        {porDia.size === 0 ? (
          <p className="mt-4 rounded-xl border border-borde bg-superficie p-8 text-center text-tinta-suave">
            {p.estado === "activo"
              ? "No tiene horarios disponibles en las próximas dos semanas."
              : "No está atendiendo en este momento."}
          </p>
        ) : (
          <div className="mt-4 grid gap-4">
            {[...porDia.entries()].map(([dia, huecos]) => (
              <div key={dia} className="rounded-xl border border-borde bg-superficie p-4">
                <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-tinta-suave">
                  {formatearDia(huecos[0].inicio)}
                </h3>
                <ul className="flex flex-wrap gap-2">
                  {huecos.map((hueco) => (
                    <li key={hueco.inicio}>
                      {/* Un link y no un botón: funciona sin JavaScript, se
                          puede abrir en otra pestaña, y el horario elegido
                          viaja en la URL en vez de en estado. */}
                      <Link
                        href={`/perfiles/${slug}/reservar?inicio=${encodeURIComponent(hueco.inicio)}`}
                        className="block rounded-lg border border-borde px-3 py-2 transition-colors hover:border-accion hover:bg-accent"
                      >
                        <Hora inicio={hueco.inicio} />
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

function Dato({ etiqueta, children }: { etiqueta: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        {etiqueta}
      </dt>
      <dd className="mt-1 text-sm font-medium">{children}</dd>
    </div>
  );
}
