import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Publicá tu agenda",
  description:
    "Publicá tus horarios y recibí turnos sin coordinar por WhatsApp. Sin comisión por turno y sin permanencia.",
  alternates: { canonical: "/profesionales" },
};

/**
 * La landing de captación.
 *
 * Server Component, como todo el grupo público: es una página que tiene que
 * indexarse. El CTA lleva a `/entrar?volver=/panel/perfil`, así que quien ya
 * tiene cuenta entra y cae directo en el alta del perfil.
 *
 * Sin números inventados. Nada de "más de 500 profesionales" ni "10.000
 * pacientes": no hay de dónde sacarlos, y en salud un número falso es
 * exactamente lo que rompe la confianza que la página vino a construir.
 */
export default function Profesionales() {
  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-14 sm:px-6 sm:py-20">
      <h1 className="max-w-2xl text-3xl font-black tracking-tight sm:text-4xl">
        Tu agenda, publicada. Los turnos entran solos.
      </h1>
      <p className="mt-4 max-w-prose text-lg text-tinta-suave">
        Cargás los días y horas en que atendés, y tu perfil empieza a mostrar
        horarios reservables. Sin coordinar por mensaje, sin confirmar uno por
        uno.
      </p>

      <div className="mt-8 flex flex-wrap items-center gap-4">
        <Link
          href="/entrar?volver=%2Fpanel%2Fperfil"
          className="rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva"
        >
          Crear mi perfil
        </Link>
        <span className="text-sm text-tinta-suave">Es gratis y lleva unos minutos.</span>
      </div>

      <section className="mt-16 border-t border-borde pt-10">
        <h2 className="text-xl font-bold tracking-tight">Qué incluye</h2>
        <dl className="mt-5 grid gap-6 sm:grid-cols-2">
          <Item titulo="Tu agenda, tus reglas">
            Definís los bloques de cada día, cuánto dura una sesión y con cuánta
            anticipación aceptás reservas. Bloqueás un rato cuando lo necesitás.
          </Item>
          <Item titulo="Un perfil que se encuentra">
            Tu especialidad, tu zona, tu matrícula y tus horarios libres, en una
            página pensada para aparecer cuando alguien busca lo que hacés.
          </Item>
          <Item titulo="Los turnos, en un lugar">
            Quién viene, a qué hora y por qué motivo. Si un paciente cancela, lo
            ves en tu agenda con el horario liberado.
          </Item>
          <Item titulo="Nada que instalar">
            Entrás desde el teléfono entre pacientes o desde la computadora para
            cargar la semana entera.
          </Item>
        </dl>
      </section>

      <section className="mt-14 border-t border-borde pt-10">
        <h2 className="text-xl font-bold tracking-tight">Antes de empezar</h2>
        <p className="mt-4 max-w-prose text-tinta-suave">
          Vas a necesitar tu número de matrícula. Se publica en tu perfil: es lo
          que le permite a un paciente confirmar que sos quien decís ser, y es la
          razón por la que la plataforma funciona.
        </p>
        <p className="mt-4 max-w-prose text-tinta-suave">
          Todavía no cobramos comisión por turno ni pedimos permanencia. Si eso
          cambia, te lo vamos a decir antes.
        </p>

        <Link
          href="/entrar?volver=%2Fpanel%2Fperfil"
          className="mt-8 inline-block rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva"
        >
          Crear mi perfil
        </Link>
      </section>
    </main>
  );
}

function Item({ titulo, children }: { titulo: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="font-bold">{titulo}</dt>
      <dd className="mt-1 text-sm text-tinta-suave">{children}</dd>
    </div>
  );
}
