"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { NavegacionPanel } from "@/componentes/navegacion-panel";
import { ErrorAPI, pedir, type Profesional, type UsuarioActual } from "@/lib/api";
import { ProveedorPanel, type Panel } from "@/lib/panel";

/**
 * El layout del panel: resuelve quién sos antes de dejarte pasar.
 *
 * Es cliente y no servidor porque la cookie de sesión viaja desde el browser:
 * es la misma línea que separa las páginas públicas —que se renderizan enteras
 * en el servidor para que las indexe Google— de todo lo que está detrás de una
 * sesión.
 */
export default function LayoutPanel({ children }: LayoutProps<"/">) {
  const router = useRouter();
  const ruta = usePathname();
  const [panel, setPanel] = useState<Panel | null>(null);

  useEffect(() => {
    let vigente = true;

    (async () => {
      let usuario: UsuarioActual;
      try {
        usuario = await pedir<UsuarioActual>("/api/v1/usuarios/yo");
      } catch (e) {
        if (e instanceof ErrorAPI && e.estado === 401) {
          router.replace(`/entrar?volver=${encodeURIComponent(ruta)}`);
          return;
        }
        throw e;
      }
      if (!vigente) return;

      // Sin perfil no hay panel que mostrar: el alta vive en /empezar, que es
      // un flujo paso a paso fuera de este grupo de rutas.
      if (!usuario.perfilProfesionalId) {
        router.replace("/empezar");
        return;
      }

      const perfil = await pedir<Profesional>(
        `/api/v1/profesionales/${usuario.perfilProfesionalId}`,
      );
      if (vigente) setPanel({ usuario, perfil });
    })();

    return () => {
      vigente = false;
    };
  }, [router, ruta]);

  if (!panel) {
    return (
      <main className="mx-auto w-full max-w-3xl px-4 py-20 text-center sm:px-6">
        <p className="text-tinta-suave">Cargando…</p>
      </main>
    );
  }

  return (
    <ProveedorPanel value={panel}>
      {/* flex-1 y no min-h-full: el body es el flex column que tiene la altura,
          y sin crecer acá la barra lateral termina donde termina el contenido. */}
      <div className="flex flex-1 flex-col sm:flex-row">
        <NavegacionPanel />
        {/* pb-20 en móvil: las tabs están fijas abajo y taparían el final del
            contenido. */}
        <div className="flex-1 pb-20 sm:pb-0">{children}</div>
      </div>
    </ProveedorPanel>
  );
}
