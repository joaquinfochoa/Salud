"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { ErrorAPI, pedir } from "@/lib/api";
import { destinoDespuesDeEntrar, destinoSeguro } from "@/lib/destino";

/**
 * Crear una cuenta.
 *
 * Existía un agujero: la única alta de usuario vivía dentro del formulario de
 * reservar un turno, así que un profesional que llegaba desde la landing de
 * captación caía en `/entrar` sin forma de registrarse, y el texto lo mandaba a
 * reservarse un turno con otro profesional. El E2E no lo agarró porque creaba
 * la cuenta con un fetch directo a la API en vez de pasar por la pantalla.
 */
function FormularioCrearCuenta() {
  const router = useRouter();
  const parametros = useSearchParams();
  const volver = parametros.get("volver");

  const [aviso, setAviso] = useState<{ texto: string; href?: string } | null>(null);
  const [error, setError] = useState<ErrorAPI | null>(null);
  const [enviando, setEnviando] = useState(false);

  async function crear(formulario: FormData) {
    setEnviando(true);
    setAviso(null);
    setError(null);

    try {
      // El back deja la sesión abierta en la misma respuesta, así que no hace
      // falta un login después.
      await pedir("/api/v1/usuarios", {
        method: "POST",
        body: JSON.stringify({
          email: String(formulario.get("email") ?? ""),
          contrasena: String(formulario.get("contrasena") ?? ""),
          nombre: String(formulario.get("nombre") ?? ""),
          apellido: String(formulario.get("apellido") ?? ""),
        }),
      });

      // Una cuenta recién creada nunca tiene perfil profesional, así que sin
      // `volver` va a sus turnos. Con `volver` —el caso de la landing de
      // captación— va a donde pidió: el alta del perfil.
      router.push(destinoDespuesDeEntrar(volver, null));
      router.refresh();
    } catch (e) {
      setEnviando(false);
      if (e instanceof ErrorAPI && e.estado === 409) {
        setAviso({
          texto: "Ya tenés una cuenta con ese email.",
          href: `/entrar${volver ? `?volver=${encodeURIComponent(destinoSeguro(volver))}` : ""}`,
        });
        return;
      }
      if (e instanceof ErrorAPI && e.estado === 422) {
        setError(e);
        return;
      }
      throw e;
    }
  }

  return (
    <main className="mx-auto w-full max-w-sm px-4 py-16 sm:px-6">
      <h1 className="text-2xl font-black tracking-tight">Crear cuenta</h1>
      <p className="mt-2 text-sm text-tinta-suave">
        Con la misma cuenta reservás turnos y, si sos profesional, publicás tu
        agenda.
      </p>

      <form
        action={crear}
        className="mt-6 grid gap-4 rounded-xl border border-borde bg-superficie p-5"
      >
        {aviso && (
          <p
            role="alert"
            className="rounded-lg border border-accion bg-accent px-3 py-2 text-sm"
          >
            {aviso.texto}{" "}
            {aviso.href && (
              <Link href={aviso.href} className="font-bold text-accion underline">
                Entrar
              </Link>
            )}
          </p>
        )}

        <div className="grid gap-4 sm:grid-cols-2">
          <Campo nombre="nombre" etiqueta="Nombre" autoComplete="given-name" error={error} />
          <Campo nombre="apellido" etiqueta="Apellido" autoComplete="family-name" error={error} />
        </div>
        <Campo
          nombre="email"
          etiqueta="Email"
          tipo="email"
          autoComplete="email"
          error={error}
        />
        <Campo
          nombre="contrasena"
          etiqueta="Contraseña"
          tipo="password"
          autoComplete="new-password"
          ayuda="Al menos 8 caracteres"
          error={error}
        />

        <button
          type="submit"
          disabled={enviando}
          className="mt-2 rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
        >
          {enviando ? "Creando…" : "Crear cuenta"}
        </button>
      </form>

      <p className="mt-5 text-center text-sm text-tinta-suave">
        ¿Ya tenés cuenta?{" "}
        <Link
          href={`/entrar${volver ? `?volver=${encodeURIComponent(destinoSeguro(volver))}` : ""}`}
          className="font-semibold text-accion underline"
        >
          Entrar
        </Link>
      </p>
    </main>
  );
}

function Campo({
  nombre,
  etiqueta,
  tipo = "text",
  autoComplete,
  ayuda,
  error,
}: {
  nombre: string;
  etiqueta: string;
  tipo?: string;
  autoComplete?: string;
  ayuda?: string;
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
        required
        autoComplete={autoComplete}
        // Sin aria-invalid y aria-describedby, un lector de pantalla anuncia el
        // campo sin decir que está mal ni por qué.
        aria-invalid={mensaje ? true : undefined}
        aria-describedby={mensaje ? idError : ayuda ? idAyuda : undefined}
        className={`h-11 w-full rounded-lg border px-3 ${
          mensaje ? "border-destructive" : "border-borde"
        }`}
      />
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

// useSearchParams necesita un límite de Suspense: sin él, Next no puede
// prerenderizar nada de esta ruta.
export default function CrearCuenta() {
  return (
    <Suspense>
      <FormularioCrearCuenta />
    </Suspense>
  );
}
