"use client";

import { useEffect, useState } from "react";
import { Campo } from "@/componentes/campo";
import { ErrorAPI, pedir, type UsuarioActual } from "@/lib/api";

/**
 * Los datos de la cuenta.
 *
 * Está en el área del paciente pero la usan los dos: un profesional también es
 * un `Usuario`, y su nombre, su email y su teléfono viven acá. Lo que edita en
 * `/panel/perfil` es otra cosa —su perfil público— y por eso son dos pantallas
 * y no una.
 *
 * Sin contraseña: cambiarla necesita pedir la actual, y meterla en el mismo
 * formulario que el teléfono haría que un descuido reemplace una credencial.
 */
export default function Cuenta() {
  const [usuario, setUsuario] = useState<UsuarioActual | null>(null);
  const [nombre, setNombre] = useState("");
  const [apellido, setApellido] = useState("");
  const [email, setEmail] = useState("");
  const [telefono, setTelefono] = useState("");

  const [aviso, setAviso] = useState<string | null>(null);
  const [error, setError] = useState<ErrorAPI | null>(null);
  const [guardando, setGuardando] = useState(false);

  useEffect(() => {
    let vigente = true;
    pedir<UsuarioActual>("/api/v1/usuarios/yo").then((yo) => {
      if (!vigente) return;
      setUsuario(yo);
      setNombre(yo.nombre);
      setApellido(yo.apellido);
      setEmail(yo.email);
      setTelefono(yo.telefono);
    });
    return () => {
      vigente = false;
    };
  }, []);

  const cambioElEmail = Boolean(usuario) && email.trim() !== usuario?.email;

  async function guardar() {
    setAviso(null);
    setError(null);
    setGuardando(true);

    try {
      const actualizado = await pedir<UsuarioActual>("/api/v1/usuarios/yo", {
        method: "PUT",
        body: JSON.stringify({ email, nombre, apellido, telefono }),
      });
      setUsuario(actualizado);
      // El teléfono vuelve normalizado, así que se pisa lo tipeado con lo que
      // quedó guardado: si escribió "011 15 1234-5678" tiene que ver en qué se
      // convirtió, no seguir viendo su borrador.
      setTelefono(actualizado.telefono);
      setEmail(actualizado.email);
      setAviso(
        cambioElEmail
          ? "Datos guardados. De ahora en más entrás con tu email nuevo."
          : "Datos guardados.",
      );
    } catch (e) {
      if (e instanceof ErrorAPI && e.estado === 409) {
        setAviso("Ya hay otra cuenta con ese email.");
      } else if (e instanceof ErrorAPI && e.estado === 422) {
        setError(e);
      } else {
        setAviso("No pudimos guardar los cambios. Probá de nuevo.");
      }
    } finally {
      setGuardando(false);
    }
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <h1 className="text-2xl font-black tracking-tight">Mi cuenta</h1>

      {aviso && (
        <p
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso}
        </p>
      )}

      {!usuario ? (
        <p className="mt-6 text-tinta-suave">Cargando…</p>
      ) : (
        <>
          <div className="mt-6 grid gap-5 rounded-xl border border-borde bg-superficie p-5">
            <div className="grid gap-5 sm:grid-cols-2">
              <Campo
                nombre="nombre"
                etiqueta="Nombre"
                valor={nombre}
                onCambiar={setNombre}
                autoComplete="given-name"
                error={error}
              />
              <Campo
                nombre="apellido"
                etiqueta="Apellido"
                valor={apellido}
                onCambiar={setApellido}
                autoComplete="family-name"
                error={error}
              />
            </div>

            <Campo
              nombre="email"
              etiqueta="Email"
              tipo="email"
              valor={email}
              onCambiar={setEmail}
              autoComplete="email"
              // Se avisa antes y no después: cambiar el email cambia con qué se
              // entra, y enterarse recién en el próximo login es tarde.
              ayuda={
                cambioElEmail
                  ? "Vas a entrar con este email a partir de ahora."
                  : "Con este email entrás a tu cuenta."
              }
              error={error}
            />

            <Campo
              nombre="telefono"
              etiqueta="Celular"
              tipo="tel"
              valor={telefono}
              onCambiar={setTelefono}
              autoComplete="tel"
              ayuda="Para avisarte si hay un cambio en tu turno. No se muestra en ningún lado."
              error={error}
            />

            <button
              type="button"
              onClick={guardar}
              disabled={guardando}
              className="mt-1 justify-self-start rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
            >
              {guardando ? "Guardando…" : "Guardar cambios"}
            </button>
          </div>

          {/* La contraseña no se edita acá, y se dice por qué en vez de dejar
              a alguien buscándola. */}
          {/* Sin "cerrar sesión" acá: el encabezado del área del paciente ya
              lo tiene, en escritorio y en móvil. Dos botones para lo mismo en
              la misma pantalla es ruido. */}
          <p className="mt-6 text-sm text-tinta-suave">
            La contraseña se cambia por separado, pidiéndote la actual. Todavía
            no está: si necesitás cambiarla, escribinos.
          </p>
        </>
      )}
    </main>
  );
}
