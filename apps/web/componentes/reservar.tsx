"use client";

import Link from "next/link";
import { useState } from "react";
import { Calendario } from "./calendario";
import { ErrorAPI, pedir, type Hueco, type Profesional, type Turno } from "@/lib/api";
import { formatearDia, formatearHora, formatearPrecio } from "@/lib/formato";
import { armarDias, primerDiaConItems } from "@/lib/dias";
import { huecosDe } from "@/lib/huecos";
import { Hora } from "./hora";

type Paso = "elegir" | "confirmar" | "listo";

type Aviso = {
  mensaje: string;
  accion?: { texto: string; href: string };
};

export function Reservar({
  slug,
  profesional,
  huecosIniciales,
  inicioPreseleccionado,
}: {
  slug: string;
  profesional: Profesional;
  huecosIniciales: Hueco[];
  inicioPreseleccionado: string | null;
}) {
  const preseleccionado =
    huecosIniciales.find((h) => h.inicio === inicioPreseleccionado) ?? null;

  const [huecos, setHuecos] = useState(huecosIniciales);
  const [elegido, setElegido] = useState<Hueco | null>(preseleccionado);
  const [paso, setPaso] = useState<Paso>(preseleccionado ? "confirmar" : "elegir");
  const [turno, setTurno] = useState<Turno | null>(null);
  const [aviso, setAviso] = useState<Aviso | null>(null);
  const [errorDeCampo, setErrorDeCampo] = useState<ErrorAPI | null>(null);
  const [enviando, setEnviando] = useState(false);

  const rutaActual = `/perfiles/${slug}/reservar${elegido ? `?inicio=${encodeURIComponent(elegido.inicio)}` : ""}`;

  async function confirmar(formulario: FormData) {
    if (!elegido) return;

    setEnviando(true);
    setAviso(null);
    setErrorDeCampo(null);

    const datos = {
      email: String(formulario.get("email") ?? ""),
      contrasena: String(formulario.get("contrasena") ?? ""),
      nombre: String(formulario.get("nombre") ?? ""),
      apellido: String(formulario.get("apellido") ?? ""),
      motivo: String(formulario.get("motivo") ?? ""),
    };

    // 1. Registrarse. El back deja la sesión abierta en la misma respuesta, así
    //    que no hace falta un login después.
    try {
      await pedir("/api/v1/usuarios", {
        method: "POST",
        body: JSON.stringify({
          email: datos.email,
          contrasena: datos.contrasena,
          nombre: datos.nombre,
          apellido: datos.apellido,
        }),
      });
    } catch (e) {
      setEnviando(false);
      if (e instanceof ErrorAPI && e.estado === 409) {
        // El caso más probable de los dos: alguien que ya se registró antes y
        // no se acuerda. Un link, no un login acá adentro: conservar el hueco a
        // través de un segundo flujo son ochenta líneas para ahorrar dos
        // clicks.
        setAviso({
          mensaje: "Ya tenés una cuenta con ese email.",
          accion: { texto: "Entrar", href: `/entrar?volver=${encodeURIComponent(rutaActual)}` },
        });
        return;
      }
      if (e instanceof ErrorAPI && e.estado === 422) {
        setErrorDeCampo(e);
        return;
      }
      throw e;
    }

    // 2. Reservar. Ya hay sesión.
    try {
      const creado = await pedir<Turno>(
        `/api/v1/profesionales/${profesional.id}/turnos`,
        {
          method: "POST",
          body: JSON.stringify({ inicio: elegido.inicio, motivo: datos.motivo }),
        },
      );
      setTurno(creado);
      setPaso("listo");
    } catch (e) {
      setEnviando(false);
      if (e instanceof ErrorAPI && e.estado === 409) {
        // Alguien tomó el hueco mientras completaba el formulario. La cuenta ya
        // existe y la sesión está abierta, así que volver a elegir es un click:
        // no se pierde nada de lo que hizo.
        setHuecos(await huecosDe(profesional.id));
        setElegido(null);
        setPaso("elegir");
        setAviso({ mensaje: "Ese horario se tomó recién. Elegí otro." });
        return;
      }
      if (e instanceof ErrorAPI && e.estado === 422) {
        setErrorDeCampo(e);
        return;
      }
      throw e;
    } finally {
      setEnviando(false);
    }
  }

  if (paso === "listo" && turno) {
    return (
      <Marco slug={slug}>
        <div className="rounded-xl border border-borde bg-superficie p-8 text-center">
          <p className="text-sm font-semibold uppercase tracking-wide text-libre">
            Turno reservado
          </p>
          <p className="mt-4 font-horas text-5xl font-black tabular-nums tracking-tight">
            {formatearHora(turno.inicio)}
          </p>
          <p className="mt-1 text-tinta-suave">{formatearDia(turno.inicio)}</p>
          <p className="mt-6 text-sm">
            Con {profesional.nombre} {profesional.apellido}
          </p>
          <Link
            href="/turnos"
            className="mt-8 inline-block rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva"
          >
            Ver mis turnos
          </Link>
        </div>
      </Marco>
    );
  }

  return (
    <Marco slug={slug}>
      <h1 className="text-2xl font-black tracking-tight">
        Turno con {profesional.nombre} {profesional.apellido}
      </h1>

      {aviso && (
        // role="alert" para que un lector de pantalla lo anuncie: si el hueco
        // se perdió, alguien que no ve la pantalla tiene que enterarse igual.
        <div
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso.mensaje}{" "}
          {aviso.accion && (
            <Link href={aviso.accion.href} className="font-bold text-accion underline">
              {aviso.accion.texto}
            </Link>
          )}
        </div>
      )}

      {paso === "elegir" ? (
        <ElegirHorario
          huecos={huecos}
          onElegir={(hueco) => {
            setElegido(hueco);
            setPaso("confirmar");
            setAviso(null);
          }}
        />
      ) : (
        elegido && (
          <Formulario
            elegido={elegido}
            profesional={profesional}
            enviando={enviando}
            errorDeCampo={errorDeCampo}
            onVolver={() => {
              setPaso("elegir");
              setAviso(null);
              setErrorDeCampo(null);
            }}
            onConfirmar={confirmar}
          />
        )
      )}
    </Marco>
  );
}

function Marco({ slug, children }: { slug: string; children: React.ReactNode }) {
  return (
    <main className="mx-auto w-full max-w-xl px-4 py-10 sm:px-6 sm:py-14">
      <Link href={`/perfiles/${slug}`} className="text-sm text-tinta-suave hover:text-accion">
        ← Volver al perfil
      </Link>
      <div className="mt-6">{children}</div>
    </main>
  );
}

function ElegirHorario({
  huecos,
  onElegir,
}: {
  huecos: Hueco[];
  onElegir: (hueco: Hueco) => void;
}) {
  const dias = armarDias(huecos);
  // Acá el día elegido sí es estado: navegar para cambiarlo desmontaría el
  // componente y se perdería el formulario a medio llenar. Se deriva y no se
  // guarda crudo porque después de un 409 los huecos se recargan, y el día que
  // estaba abierto puede haberse quedado sin ninguno.
  const [pedido, setDia] = useState("");
  const dia = dias.some((d) => d.fecha === pedido && d.items.length > 0)
    ? pedido
    : primerDiaConItems(dias);

  if (!dias.some((d) => d.items.length > 0)) {
    return (
      <p className="mt-6 rounded-xl border border-borde bg-superficie p-8 text-center text-tinta-suave">
        No quedan horarios disponibles en las próximas dos semanas.
      </p>
    );
  }

  return (
    <div className="mt-6">
      <Calendario dias={dias} diaElegido={dia} onDia={setDia}>
        {(huecos) => (
          <ul className="flex flex-wrap gap-2">
            {huecos.map((hueco) => (
              <li key={hueco.inicio}>
                <button
                  type="button"
                  onClick={() => onElegir(hueco)}
                  className="block rounded-lg border border-borde px-4 py-2.5 transition-colors hover:border-accion hover:bg-accent"
                >
                  <Hora inicio={hueco.inicio} />
                </button>
              </li>
            ))}
          </ul>
        )}
      </Calendario>
    </div>
  );
}

function Formulario({
  elegido,
  profesional,
  enviando,
  errorDeCampo,
  onVolver,
  onConfirmar,
}: {
  elegido: Hueco;
  profesional: Profesional;
  enviando: boolean;
  errorDeCampo: ErrorAPI | null;
  onVolver: () => void;
  onConfirmar: (formulario: FormData) => void;
}) {
  return (
    <>
      <div className="mt-6 rounded-xl border border-accion bg-accent p-5">
        <div className="flex items-baseline gap-3">
          <Hora inicio={elegido.inicio} />
          <span className="text-sm">{formatearDia(elegido.inicio)}</span>
        </div>
        <p className="mt-2 text-sm text-tinta-suave">
          {formatearPrecio(profesional.precioConsultaCentavos)} · se paga en la consulta
        </p>
        <button
          type="button"
          onClick={onVolver}
          className="mt-3 text-sm font-semibold text-accion underline"
        >
          Cambiar horario
        </button>
      </div>

      <form
        action={onConfirmar}
        className="mt-6 grid gap-4 rounded-xl border border-borde bg-superficie p-5"
      >
        <p className="text-sm text-tinta-suave">
          Creamos tu cuenta con estos datos para que puedas ver y cancelar tus
          turnos.
        </p>

        <Campo nombre="email" etiqueta="Email" tipo="email" error={errorDeCampo} />
        <Campo
          nombre="contrasena"
          etiqueta="Contraseña"
          tipo="password"
          ayuda="Al menos 8 caracteres"
          error={errorDeCampo}
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <Campo nombre="nombre" etiqueta="Nombre" error={errorDeCampo} />
          <Campo nombre="apellido" etiqueta="Apellido" error={errorDeCampo} />
        </div>
        <Campo
          nombre="motivo"
          etiqueta="Motivo de la consulta"
          requerido={false}
          ayuda="Opcional. Solo lo ve el profesional."
          error={errorDeCampo}
        />

        <button
          type="submit"
          disabled={enviando}
          className="mt-2 rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
        >
          {enviando ? "Reservando…" : `Reservar ${formatearHora(elegido.inicio)}`}
        </button>
      </form>
    </>
  );
}

function Campo({
  nombre,
  etiqueta,
  tipo = "text",
  ayuda,
  requerido = true,
  error,
}: {
  nombre: string;
  etiqueta: string;
  tipo?: string;
  ayuda?: string;
  requerido?: boolean;
  error: ErrorAPI | null;
}) {
  const mensaje = error?.porCampo(nombre);
  const idAyuda = `${nombre}-ayuda`;
  const idError = `${nombre}-error`;

  return (
    // min-w-0: un <input> tiene ancho intrínseco propio, y en CSS Grid un item
    // no se encoge por debajo del contenido. Sin esto, dos campos en
    // sm:grid-cols-2 se desbordan de la tarjeta que los contiene.
    <div className="grid min-w-0 gap-1.5">
      <label htmlFor={nombre} className="text-sm font-semibold">
        {etiqueta}
      </label>
      <input
        id={nombre}
        name={nombre}
        type={tipo}
        required={requerido}
        // Sin aria-invalid y aria-describedby, un lector de pantalla anuncia el
        // campo sin decir que está mal ni por qué.
        aria-invalid={mensaje ? true : undefined}
        aria-describedby={mensaje ? idError : ayuda ? idAyuda : undefined}
        className={`h-11 w-full rounded-lg border px-3 ${
          mensaje ? "border-destructive" : "border-borde"
        }`}
      />
      {/* El mensaje va debajo de SU campo, no en un cartel arriba: un error de
          formulario lejos del campo obliga a adivinar cuál falló. */}
      {mensaje ? (
        <p id={idError} className="text-sm text-destructive">
          {mensaje}
        </p>
      ) : (
        ayuda && (
          <p id={idAyuda} className="text-sm text-tinta-suave">
            {ayuda}
          </p>
        )
      )}
    </div>
  );
}
