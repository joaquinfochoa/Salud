"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { ErrorAPI, pedir, type UsuarioActual } from "@/lib/api";
import { destinoDespuesDeEntrar } from "@/lib/destino";

function FormularioEntrar() {
  const router = useRouter();
  const parametros = useSearchParams();
  const volver = parametros.get("volver");

  const [error, setError] = useState<string | null>(null);
  const [enviando, setEnviando] = useState(false);

  async function entrar(formulario: FormData) {
    setEnviando(true);
    setError(null);

    try {
      await pedir("/api/v1/sesiones", {
        method: "POST",
        body: JSON.stringify({
          email: String(formulario.get("email") ?? ""),
          contrasena: String(formulario.get("contrasena") ?? ""),
        }),
      });

      // Segunda llamada, y hace falta: la sesión no dice si sos profesional.
      // `perfilProfesionalId` es lo único que lo sabe, y sale de acá.
      const yo = await pedir<UsuarioActual>("/api/v1/usuarios/yo");

      router.push(destinoDespuesDeEntrar(parametros.get("volver"), yo.perfilProfesionalId));
      router.refresh();
    } catch (e) {
      setEnviando(false);
      if (e instanceof ErrorAPI && e.estado === 401) {
        // El mismo texto que manda el back, que a propósito no distingue si
        // falló el email o la contraseña: distinguirlos convertiría el login en
        // un oráculo de qué direcciones están registradas.
        setError("El email o la contraseña no son correctos.");
        return;
      }
      throw e;
    }
  }

  return (
    <main className="mx-auto w-full max-w-sm px-4 py-16 sm:px-6">
      <h1 className="text-2xl font-black tracking-tight">Entrar</h1>

      <form
        action={entrar}
        className="mt-6 grid gap-4 rounded-xl border border-borde bg-superficie p-5"
      >
        {error && (
          <p role="alert" className="rounded-lg border border-destructive px-3 py-2 text-sm text-destructive">
            {error}
          </p>
        )}

        <div className="grid gap-1.5">
          <label htmlFor="email" className="text-sm font-semibold">
            Email
          </label>
          <input
            id="email"
            name="email"
            type="email"
            required
            autoComplete="email"
            className="h-11 rounded-lg border border-borde px-3"
          />
        </div>

        <div className="grid gap-1.5">
          <label htmlFor="contrasena" className="text-sm font-semibold">
            Contraseña
          </label>
          <input
            id="contrasena"
            name="contrasena"
            type="password"
            required
            autoComplete="current-password"
            className="h-11 rounded-lg border border-borde px-3"
          />
        </div>

        <button
          type="submit"
          disabled={enviando}
          className="mt-2 rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
        >
          {enviando ? "Entrando…" : "Entrar"}
        </button>
      </form>

      {/* Antes acá decía que la cuenta "se crea sola cuando reservás tu primer
          turno", y era el único camino que existía: alguien que llegaba desde
          la landing de captación quedaba sin forma de registrarse, y el texto
          lo mandaba a reservarse un turno con otro profesional. */}
      <p className="mt-5 text-center text-sm text-tinta-suave">
        ¿No tenés cuenta?{" "}
        <Link
          href={`/crear-cuenta${volver ? `?volver=${encodeURIComponent(volver)}` : ""}`}
          className="font-semibold text-accion underline"
        >
          Crear una
        </Link>
      </p>
    </main>
  );
}

// useSearchParams necesita un límite de Suspense: sin él, Next no puede
// prerenderizar nada de esta ruta.
export default function Entrar() {
  return (
    <Suspense>
      <FormularioEntrar />
    </Suspense>
  );
}
