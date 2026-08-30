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
  const [sinPerfil, setSinPerfil] = useState(false);

  // /panel/perfil tiene que quedar accesible sin perfil: es la pantalla donde
  // se crea. Sin esta excepción, alguien que llega desde la landing de
  // captación rebota para siempre entre el layout y el alta.
  const esAlta = ruta === "/panel/perfil";

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

      if (!usuario.perfilProfesionalId) {
        if (esAlta) setSinPerfil(true);
        else router.replace("/panel/perfil");
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
  }, [router, ruta, esAlta]);

  // Sin perfil todavía: el alta se dibuja sin navegación, porque no hay a dónde
  // navegar hasta que exista el perfil.
  //
  // Se exige `esAlta` además de `sinPerfil`: apenas se crea el perfil, la
  // pantalla navega a /panel y este layout todavía tiene el estado viejo. Sin
  // esa condición, /panel se dibujaba sin proveedor y usePanel() explotaba
  // antes de que el efecto llegara a traer el perfil recién creado.
  if (sinPerfil && esAlta) return <>{children}</>;

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
