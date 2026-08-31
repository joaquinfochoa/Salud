import Link from "next/link";
import { Hora } from "@/componentes/hora";
import { ESPECIALIDADES, TarjetaProfesional } from "@/componentes/tarjeta-profesional";
import { pedir, type Hueco, type ListaProfesionales, type Profesional } from "@/lib/api";
import { armarDias } from "@/lib/dias";
import { formatearDia } from "@/lib/formato";
import { huecosDe, proximoHueco } from "@/lib/huecos";

const EN_PORTADA = 4;

// Sin esto Next prerenderiza la página en el build, y son dos problemas a la
// vez: la disponibilidad que muestra la portada quedaría congelada en el
// momento de compilar, y el build de CI —donde la API no corre— fallaría al
// intentar traerla. Antes esta ruta era dinámica de casualidad, porque el
// buscador leía searchParams.
export const dynamic = "force-dynamic";

/**
 * La landing.
 *
 * Server Component, y tiene que seguir siéndolo: sigue siendo la página con más
 * autoridad de SEO del sitio. Por eso el mensaje convive con contenido real
 * —profesionales, especialidades, horarios de verdad— en vez de reemplazarlo.
 * El buscador completo, con filtros, vive en `/buscar`.
 *
 * La pieza central no es una ilustración: son **horarios reales traídos de la
 * API**. La promesa del producto es que la hora que ves es la que reservás, así
 * que la prueba tiene que ser la hora misma, con la tipografía que ya es la
 * firma de la app.
 */
export default async function Inicio() {
  const lista = await pedir<ListaProfesionales>("/api/v1/profesionales");
  const destacados = lista.datos.slice(0, EN_PORTADA);

  const proximos = await Promise.all(destacados.map((p) => proximoHueco(p.id)));

  // Para la muestra de disponibilidad: el primero que efectivamente tenga
  // horarios. Si ninguno tiene, la sección no se dibuja — antes que inventarla.
  const indice = proximos.findIndex((h) => h !== null);
  const enVivo = indice >= 0 ? destacados[indice] : null;
  const huecos = enVivo ? (await huecosDe(enVivo.id, 7)).slice(0, 30) : [];

  return (
    <main>
      <Portada profesional={enVivo} huecos={huecos} />

      <Seccion titulo="Cómo funciona">
        {/* Numerados porque acá el orden SÍ es información: son tres cosas que
            pasan en ese orden, no tres características. */}
        <ol className="grid gap-10 sm:grid-cols-3 sm:gap-8">
          <Paso numero="1" titulo="Buscá">
            Filtrá por especialidad y zona. Cada perfil muestra la matrícula, el
            precio y las obras sociales que acepta.
          </Paso>
          <Paso numero="2" titulo="Elegí el horario">
            Los horarios que ves son los que el profesional tiene libres ahora.
            No hay lista de espera ni llamada de confirmación.
          </Paso>
          <Paso numero="3" titulo="Reservá">
            Queda confirmado en el momento. Lo ves y lo cancelás desde{" "}
            <span className="whitespace-nowrap">Mis turnos</span>.
          </Paso>
        </ol>
      </Seccion>

      <Seccion
        titulo="Profesionales"
        accion={{ texto: "Ver todos", href: "/buscar" }}
      >
        <ul className="grid gap-3 lg:grid-cols-2">
          {destacados.map((profesional, i) => (
            <TarjetaProfesional
              key={profesional.id}
              profesional={profesional}
              proximo={proximos[i]}
            />
          ))}
        </ul>
      </Seccion>

      <ParaProfesionales />
    </main>
  );
}

/**
 * El hero: la promesa a la izquierda, la prueba a la derecha.
 *
 * En escritorio son dos columnas porque la frase y los horarios se leen juntos:
 * el texto dice "el horario que ves es el que reservás" y al lado hay horarios
 * que se pueden tocar. En móvil se apilan y la prueba queda debajo de la
 * promesa, que es el orden en que se lee de todas formas.
 */
function Portada({
  profesional,
  huecos,
}: {
  profesional: Profesional | null;
  huecos: Hueco[];
}) {
  return (
    <section className="border-b border-borde bg-superficie">
      <div className="mx-auto grid w-full max-w-6xl gap-12 px-4 py-14 sm:px-6 sm:py-20 lg:grid-cols-[1.1fr_1fr] lg:items-center lg:gap-16">
        <div className="surgir">
          <h1 className="text-4xl font-black leading-[1.05] tracking-tight text-balance sm:text-5xl lg:text-6xl">
            El horario que ves es el que reservás.
          </h1>
          <p className="mt-6 max-w-prose text-lg leading-relaxed text-tinta-suave">
            Profesionales con matrícula verificada publican su agenda real.
            Elegís una hora y el turno queda confirmado en el momento: sin
            llamar, sin mensajes, sin esperar que te confirmen.
          </p>

          <div className="mt-9 flex flex-wrap items-center gap-x-6 gap-y-4">
            <Link
              href="/buscar"
              className="rounded-lg bg-accion px-7 py-3.5 font-bold text-white transition-colors hover:bg-accion-viva"
            >
              Buscar un turno
            </Link>
            {/* La segunda puerta. Va como link y no como segundo botón: son dos
                públicos distintos, no dos acciones del mismo peso. */}
            <Link
              href="/profesionales"
              className="font-semibold text-tinta-suave underline decoration-borde underline-offset-4 transition-colors hover:text-accion hover:decoration-accion"
            >
              Soy profesional de la salud
            </Link>
          </div>
        </div>

        {profesional && huecos.length > 0 && (
          <Disponibilidad profesional={profesional} huecos={huecos} />
        )}
      </div>
    </section>
  );
}

/**
 * Disponibilidad de verdad, traída de la API.
 *
 * Es el elemento firma de la página, y es honesto: si el turno se toma, mañana
 * esta tarjeta muestra otro. No hay foto, ni rating, ni reseñas inventadas —
 * solo lo que existe.
 */
const DIAS_EN_PORTADA = 2;
const HORAS_POR_DIA = 4;

function Disponibilidad({
  profesional,
  huecos,
}: {
  profesional: Profesional;
  huecos: Hueco[];
}) {
  // Dos días y no uno: con un solo día la tarjeta queda flaca cuando ese día
  // tiene pocos horarios libres, y sobre todo no se lee como una agenda, que es
  // exactamente lo que el producto es.
  const conHuecos = armarDias(huecos, 7)
    .filter((d) => d.items.length > 0)
    .slice(0, DIAS_EN_PORTADA);

  return (
    <div className="surgir surgir-tarde">
      <div className="rounded-2xl border border-borde bg-fondo p-6 shadow-[0_1px_2px_rgba(25,21,64,0.04),0_12px_32px_-12px_rgba(25,21,64,0.12)]">
        <div className="flex items-baseline justify-between gap-4">
          <div className="min-w-0">
            <p className="truncate font-bold tracking-tight">
              {profesional.nombre} {profesional.apellido}
            </p>
            <p className="text-sm text-tinta-suave">
              {ESPECIALIDADES[profesional.especialidad] ?? profesional.especialidad}
              {" · "}
              {profesional.zona}
            </p>
          </div>
          <p className="shrink-0 text-xs font-semibold uppercase tracking-wide text-tinta-suave">
            {profesional.matricula}
          </p>
        </div>

        {conHuecos.map((dia) => (
          <div key={dia.fecha} className="mt-6">
            <p className="text-sm font-bold uppercase tracking-wide text-tinta-suave">
              {formatearDia(dia.items[0].inicio)}
            </p>
            <ul className="mt-3 flex flex-wrap gap-2">
              {dia.items.slice(0, HORAS_POR_DIA).map((hueco) => (
                <li key={hueco.inicio}>
                  <Link
                    href={`/perfiles/${profesional.slug}/reservar?inicio=${encodeURIComponent(hueco.inicio)}`}
                    className="block rounded-lg border border-borde bg-superficie px-4 py-2.5 transition-colors hover:border-accion hover:bg-accent"
                  >
                    <Hora inicio={hueco.inicio} />
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}

        <Link
          href={`/perfiles/${profesional.slug}`}
          className="mt-6 inline-block border-t border-borde pt-5 text-sm font-semibold text-accion underline underline-offset-4"
        >
          Ver la agenda completa
        </Link>
      </div>

      <p className="mt-4 flex items-center justify-center gap-2 text-sm text-tinta-suave">
        {/* El punto no es decoración: dice que esto está vivo, y viene con su
            texto al lado porque el color nunca es la única señal. */}
        <span aria-hidden="true" className="size-1.5 rounded-full bg-libre" />
        Horarios reales. Tocá uno y reservás.
      </p>
    </div>
  );
}

function ParaProfesionales() {
  return (
    // Fondo distinto al del resto: es la otra mitad del marketplace, y tiene
    // que leerse como que le habla a otra persona.
    <section className="border-t border-borde bg-superficie">
      <div className="mx-auto w-full max-w-6xl px-4 py-14 sm:px-6 sm:py-16">
        <div className="grid gap-8 lg:grid-cols-[1fr_auto] lg:items-end">
          <div>
            <h2 className="text-2xl font-black tracking-tight sm:text-3xl">
              ¿Atendés pacientes?
            </h2>
            <p className="mt-3 max-w-prose text-tinta-suave">
              Publicá los días y horas en que atendés y tu perfil empieza a
              mostrar horarios reservables. Los turnos entran solos, y los ves
              todos en un lugar.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-x-6 gap-y-3">
            <Link
              href="/empezar"
              className="rounded-lg bg-accion px-7 py-3.5 font-bold text-white transition-colors hover:bg-accion-viva"
            >
              Publicar mi agenda
            </Link>
            <Link
              href="/entrar"
              className="font-semibold text-tinta-suave underline decoration-borde underline-offset-4 transition-colors hover:text-accion hover:decoration-accion"
            >
              Ya tengo cuenta
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}

function Seccion({
  titulo,
  accion,
  children,
}: {
  titulo: string;
  accion?: { texto: string; href: string };
  children: React.ReactNode;
}) {
  return (
    <section className="mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 sm:py-16">
      <div className="mb-8 flex items-baseline justify-between gap-4">
        <h2 className="text-2xl font-black tracking-tight sm:text-3xl">{titulo}</h2>
        {accion && (
          <Link
            href={accion.href}
            className="shrink-0 font-semibold text-accion underline underline-offset-4"
          >
            {accion.texto}
          </Link>
        )}
      </div>
      {children}
    </section>
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
    // El número en su propia columna y no encima: así queda atado al título en
    // vez de flotando, y las tres columnas alinean por la misma línea de base.
    <li className="grid grid-cols-[auto_1fr] gap-x-4">
      <span
        aria-hidden="true"
        className="font-horas text-2xl font-black leading-none tabular-nums text-apagado"
      >
        {numero}
      </span>
      <h3 className="text-lg font-bold leading-none tracking-tight">{titulo}</h3>
      <p className="col-start-2 mt-3 leading-relaxed text-tinta-suave">{children}</p>
    </li>
  );
}
