import { notFound } from "next/navigation";
import { Reservar } from "@/componentes/reservar";
import { ErrorAPI, pedir, type Profesional } from "@/lib/api";
import { huecosDe } from "@/lib/huecos";

// La página es un Server Component que trae los datos y se los pasa al
// componente de cliente. Así el primer render llega completo desde el servidor
// y solo la interacción —elegir, completar, enviar— corre en el browser.
export default async function PaginaReservar({
  params,
  searchParams,
}: PageProps<"/perfiles/[slug]/reservar">) {
  const { slug } = await params;
  const { inicio } = await searchParams;

  let profesional: Profesional;
  try {
    profesional = await pedir<Profesional>(`/api/v1/perfiles/${slug}`);
  } catch (e) {
    if (e instanceof ErrorAPI && e.estado === 404) notFound();
    throw e;
  }

  const huecos = await huecosDe(profesional.id);

  return (
    <Reservar
      slug={slug}
      profesional={profesional}
      huecosIniciales={huecos}
      inicioPreseleccionado={typeof inicio === "string" ? inicio : null}
    />
  );
}
