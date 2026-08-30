import { Filtros } from "@/componentes/filtros";
import { TarjetaProfesional } from "@/componentes/tarjeta-profesional";
import { pedir, type ListaProfesionales } from "@/lib/api";
import { proximoHueco } from "@/lib/huecos";

// Server Component, y tiene que seguir siéndolo: es una de las dos páginas que
// indexa Google, que es la razón por la que el front es Next y no Vite. Un
// `useState` acá la convierte en componente de cliente sin que nada falle, y el
// SEO desaparece en silencio.
export default async function Buscar({ searchParams }: PageProps<"/">) {
  const filtros = await searchParams;

  const consulta = new URLSearchParams();
  for (const clave of ["especialidad", "zona", "busqueda"] as const) {
    const valor = filtros[clave];
    if (typeof valor === "string" && valor) consulta.set(clave, valor);
  }

  const lista = await pedir<ListaProfesionales>(
    `/api/v1/profesionales?${consulta}`,
  );

  // En paralelo, no en serie: veinte llamadas secuenciales serían veinte veces
  // la latencia. Con Promise.all es una sola espera.
  const proximos = await Promise.all(
    lista.datos.map((profesional) => proximoHueco(profesional.id)),
  );

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <header className="mb-8">
        <h1 className="text-3xl font-black tracking-tight sm:text-4xl">
          Encontrá tu próximo turno
        </h1>
        <p className="mt-2 max-w-prose text-tinta-suave">
          Profesionales con matrícula verificada. Elegí el horario que te queda
          bien y reservá sin llamar a nadie.
        </p>
      </header>

      <Filtros
        valores={{
          especialidad: typeof filtros.especialidad === "string" ? filtros.especialidad : undefined,
          zona: typeof filtros.zona === "string" ? filtros.zona : undefined,
          busqueda: typeof filtros.busqueda === "string" ? filtros.busqueda : undefined,
        }}
      />

      <p className="mt-8 mb-4 text-sm text-tinta-suave">
        {lista.paginacion.total === 0
          ? "Sin resultados"
          : `${lista.paginacion.total} ${lista.paginacion.total === 1 ? "profesional" : "profesionales"}`}
      </p>

      {lista.datos.length === 0 ? (
        <p className="rounded-xl border border-borde bg-superficie p-8 text-center text-tinta-suave">
          No encontramos a nadie con esos filtros. Probá con otra zona o sin
          filtrar por especialidad.
        </p>
      ) : (
        <ul className="grid gap-3">
          {lista.datos.map((profesional, i) => (
            <TarjetaProfesional
              key={profesional.id}
              profesional={profesional}
              proximo={proximos[i]}
            />
          ))}
        </ul>
      )}

      {/* Debajo del buscador y de los resultados, nunca arriba. El hero de esta
          página es el buscador: `/` es la que más autoridad de SEO tiene, y
          gastarla en un hero de marketing sin contenido indexable es tirarla.
          En un marketplace de salud la búsqueda orgánica es el canal de
          adquisición, y es lo que hacen Doctoralia y Zocdoc. */}
      <section className="mt-16 border-t border-borde pt-10">
        <h2 className="text-xl font-bold tracking-tight">Cómo funciona</h2>

        {/* Numerados porque acá el orden SÍ es información: son tres pasos que
            pasan en ese orden, no tres características. */}
        <ol className="mt-5 grid gap-5 sm:grid-cols-3">
          <Paso numero="1" titulo="Buscá">
            Filtrá por especialidad y zona. Cada perfil muestra la matrícula, el
            precio y los horarios libres de verdad.
          </Paso>
          <Paso numero="2" titulo="Elegí el horario">
            Los horarios que ves son los que el profesional tiene disponibles
            ahora. No hay lista de espera ni llamada de confirmación.
          </Paso>
          <Paso numero="3" titulo="Reservá">
            Queda confirmado en el momento. Lo podés ver y cancelar desde{" "}
            <span className="whitespace-nowrap">Mis turnos</span>.
          </Paso>
        </ol>

        <p className="mt-8 text-sm text-tinta-suave">
          Todos los profesionales publican su número de matrícula. Salud no
          reemplaza una consulta de urgencia.
        </p>
      </section>
    </main>
  );
}

function Paso({
  numero,
  titulo,
  children,
}: {
  numero: string;
  titulo: string;
  children: React.ReactNode;
}) {
  return (
    <li>
      <span className="font-horas text-2xl font-black tabular-nums text-apagado">
        {numero}
      </span>
      <h3 className="mt-1 font-bold">{titulo}</h3>
      <p className="mt-1 text-sm text-tinta-suave">{children}</p>
    </li>
  );
}
