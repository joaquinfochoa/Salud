import type { Metadata } from "next";
import Link from "next/link";
import { MaquetaPanel, MaquetaPerfil, MaquetaSemana } from "@/componentes/maqueta-panel";

export const metadata: Metadata = {
  title: "Publicá tu agenda",
  description:
    "Publicá tus horarios y recibí turnos sin coordinar por WhatsApp. Sin comisión por turno y sin permanencia.",
  alternates: { canonical: "/profesionales" },
};

/**
 * La landing de captación de profesionales.
 *
 * Server Component, como todo el grupo público: es una página que tiene que
 * indexarse.
 *
 * **Sin números inventados.** Nada de "reducí un 40% las llamadas" ni "el 30%
 * confirma solo": no hay de dónde sacarlos, y en salud un número falso es
 * exactamente lo que rompe la confianza que la página vino a construir.
 *
 * **Sin fotos de profesionales.** Una imagen de banco en una página que dice
 * "publicá tu agenda" sugiere colegas que todavía no existen. Lo que se muestra
 * es el producto, dibujado con sus propios componentes.
 */
export default function Profesionales() {
  return (
    <main>
      <Portada />

      <Seccion titulo="Qué resuelve" fondo>
        <div className="grid gap-10 lg:grid-cols-[1fr_1.1fr] lg:items-center lg:gap-16">
          <div className="grid gap-7">
            <Punto titulo="Se acabó coordinar por mensaje">
              Hoy cada turno son cuatro mensajes: cuándo podés, yo puedo tal día,
              dale, confirmado. Acá el paciente ve tus horarios libres y elige
              uno. Vos te enterás cuando ya está reservado.
            </Punto>
            <Punto titulo="Tu agenda, tus reglas">
              Definís los bloques de cada día, cuánto dura una sesión, con cuánta
              anticipación aceptás reservas y qué modalidades ofrecés. Bloqueás
              un rato cuando lo necesitás.
            </Punto>
            <Punto titulo="Si algo cambia, se ve de los dos lados">
              Cancelás un turno y el paciente lo ve cancelado. Te cancelan y lo
              ves vos, con quién era. El horario vuelve a quedar libre solo.
            </Punto>
          </div>
          <MaquetaSemana />
        </div>
      </Seccion>

      <Seccion titulo="Tu perfil, del lado del paciente">
        <div className="grid gap-10 lg:grid-cols-[1.1fr_1fr] lg:items-center lg:gap-16">
          <MaquetaPerfil />
          <div className="grid gap-7">
            <Punto titulo="Con tu matrícula a la vista">
              Se publica en tu perfil. Es lo que le permite a alguien que no te
              conoce confirmar que sos quien decís ser, y es la razón por la que
              la plataforma funciona.
            </Punto>
            <Punto titulo="Tu especialidad y tu zona, buscables">
              Un paciente filtra por lo que necesita y por dónde está. Tu perfil
              es una página propia, pensada para aparecer cuando alguien busca lo
              que hacés.
            </Punto>
            <Punto titulo="Los horarios que ve son los que tenés">
              No hay lista de espera ni “te confirmo”: lo que muestra tu perfil
              sale de los bloques que cargaste, menos lo que ya está tomado.
            </Punto>
          </div>
        </div>
      </Seccion>

      <Seccion titulo="Cómo empezás" fondo>
        <ol className="grid gap-10 sm:grid-cols-3 sm:gap-8">
          <Paso numero="1" titulo="Creás tu cuenta">
            Nombre, email y celular. Un par de minutos.
          </Paso>
          <Paso numero="2" titulo="Cargás tu perfil">
            Matrícula, especialidad, zona, precio y las obras sociales que
            aceptás. Paso a paso, viendo cómo va quedando.
          </Paso>
          <Paso numero="3" titulo="Publicás tu semana">
            Los días y horas en que atendés. Desde ese momento te pueden
            reservar.
          </Paso>
        </ol>
      </Seccion>

      <Seccion titulo="Lo que conviene saber">
        <dl className="grid gap-8 sm:grid-cols-2">
          <Item titulo="¿Cuánto cuesta?">
            Hoy no cobramos comisión por turno ni pedimos permanencia. Si eso
            cambia, te lo vamos a decir antes: no lo vas a descubrir en un
            resumen.
          </Item>
          <Item titulo="¿Qué necesito para empezar?">
            Tu número de matrícula. Nada más: no hace falta instalar nada ni
            firmar nada.
          </Item>
          <Item titulo="¿Los pacientes pagan por acá?">
            Todavía no. El turno se acuerda acá y se cobra como lo hacés hoy, en
            la consulta.
          </Item>
          <Item titulo="¿Puedo dejar de usarlo?">
            Sí, cuando quieras. Borrás tus horarios y dejás de recibir turnos;
            los que ya tenías siguen en tu agenda.
          </Item>
        </dl>
      </Seccion>

      <Cierre />
    </main>
  );
}

function Portada() {
  return (
    <section className="border-b border-borde bg-superficie">
      <div className="mx-auto grid w-full max-w-6xl gap-12 px-4 py-14 sm:px-6 sm:py-20 lg:grid-cols-[1.05fr_1fr] lg:items-center lg:gap-16">
        <div className="surgir">
          <h1 className="text-4xl font-black leading-[1.05] tracking-tight text-balance sm:text-5xl lg:text-6xl">
            Tu agenda, publicada. Los turnos entran solos.
          </h1>
          <p className="mt-6 max-w-prose text-lg leading-relaxed text-tinta-suave">
            Cargás los días y horas en que atendés, y tu perfil empieza a mostrar
            horarios reservables. Sin coordinar por mensaje, sin confirmar uno
            por uno.
          </p>

          <div className="mt-9 flex flex-wrap items-center gap-x-6 gap-y-4">
            <Link
              href="/empezar"
              className="rounded-lg bg-accion px-7 py-3.5 font-bold text-white transition-colors hover:bg-accion-viva"
            >
              Publicar mi agenda
            </Link>
            <span className="text-sm text-tinta-suave">
              Es gratis y lleva unos minutos.{" "}
              <Link
                href="/entrar?volver=%2Fempezar"
                className="font-semibold underline hover:text-accion"
              >
                Ya tengo cuenta
              </Link>
              .
            </span>
          </div>
        </div>

        <div className="surgir surgir-tarde">
          <MaquetaPanel />
        </div>
      </div>
    </section>
  );
}

function Cierre() {
  return (
    <section className="border-t border-borde bg-superficie">
      <div className="mx-auto w-full max-w-6xl px-4 py-16 text-center sm:px-6 sm:py-20">
        <h2 className="text-2xl font-black tracking-tight text-balance sm:text-3xl">
          Vas a necesitar tu matrícula. Nada más.
        </h2>
        <p className="mx-auto mt-3 max-w-prose text-tinta-suave">
          Se publica en tu perfil: es lo que le permite a un paciente confirmar
          que sos quien decís ser.
        </p>
        <Link
          href="/empezar"
          className="mt-8 inline-block rounded-lg bg-accion px-7 py-3.5 font-bold text-white transition-colors hover:bg-accion-viva"
        >
          Publicar mi agenda
        </Link>
      </div>
    </section>
  );
}

function Seccion({
  titulo,
  fondo = false,
  children,
}: {
  titulo: string;
  /** Alterna el fondo, para que las secciones se separen sin líneas. */
  fondo?: boolean;
  children: React.ReactNode;
}) {
  return (
    <section className={fondo ? "bg-superficie" : ""}>
      <div className="mx-auto w-full max-w-6xl px-4 py-14 sm:px-6 sm:py-20">
        <h2 className="mb-10 text-2xl font-black tracking-tight sm:text-3xl">
          {titulo}
        </h2>
        {children}
      </div>
    </section>
  );
}

function Punto({ titulo, children }: { titulo: string; children: React.ReactNode }) {
  return (
    <div>
      <h3 className="text-lg font-bold tracking-tight">{titulo}</h3>
      <p className="mt-2 max-w-prose leading-relaxed text-tinta-suave">{children}</p>
    </div>
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
    // El número en su propia columna, atado al título. Numerados porque acá el
    // orden es información: son tres cosas que pasan en ese orden.
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

function Item({ titulo, children }: { titulo: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="font-bold">{titulo}</dt>
      <dd className="mt-2 max-w-prose leading-relaxed text-tinta-suave">{children}</dd>
    </div>
  );
}
