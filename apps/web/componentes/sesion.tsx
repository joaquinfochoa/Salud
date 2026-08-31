"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { pedir, type UsuarioActual } from "@/lib/api";
import { CerrarSesion } from "./cerrar-sesion";

const enlace =
  "rounded-lg px-3 py-2 text-sm font-semibold text-tinta-suave transition-colors hover:bg-muted hover:text-tinta";

/**
 * La parte del encabezado que depende de quién sos.
 *
 * Es cliente porque el encabezado es un Server Component y la cookie de sesión
 * viaja desde el browser: es la misma línea que separa las páginas públicas
 * —que se renderizan enteras en el servidor para que las indexe Google— de
 * todo lo que está detrás de una sesión.
 *
 * Antes el encabezado mostraba "Mis turnos" y "Entrar" siempre, a todo el
 * mundo. Le ofrecía entrar a alguien que ya había entrado, y sus turnos a
 * alguien anónimo que terminaba rebotando al login.
 */
export function Sesion() {
  const ruta = usePathname();
  const [usuario, setUsuario] = useState<UsuarioActual | null>(null);

  useEffect(() => {
    let vigente = true;
    pedir<UsuarioActual>("/api/v1/usuarios/yo")
      .then((yo) => vigente && setUsuario(yo))
      .catch(() => vigente && setUsuario(null));
    // La ruta entra como dependencia para que entrar o salir se refleje sin
    // recargar: las dos cosas navegan.
    return () => {
      vigente = false;
    };
  }, [ruta]);

  // Sin sesión —y mientras se resuelve— se muestra "Entrar". Es el estado
  // correcto para la enorme mayoría de las visitas a una página pública, y es
  // lo que se renderiza en el servidor, así que sin JavaScript el visitante
  // anónimo ve exactamente lo que le corresponde.
  if (!usuario) {
    return (
      <Link
        href="/entrar"
        className="rounded-lg border border-borde px-3 py-2 text-sm font-semibold transition-colors hover:border-accion hover:bg-accent"
      >
        Entrar
      </Link>
    );
  }

  return (
    <>
      <Link href={usuario.perfilProfesionalId ? "/panel" : "/turnos"} className={enlace}>
        {usuario.perfilProfesionalId ? "Mi panel" : "Mis turnos"}
      </Link>
      <CerrarSesion className={enlace} />
    </>
  );
}
