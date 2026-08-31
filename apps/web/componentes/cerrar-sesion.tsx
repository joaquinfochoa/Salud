"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { pedir } from "@/lib/api";

/**
 * Cerrar la sesión.
 *
 * `DELETE /api/v1/sesiones/actual` existía desde la etapa de autenticación y no
 * lo llamaba nadie: se podía entrar y no salir nunca. En una computadora
 * compartida eso deja los turnos y el motivo de consulta de una persona a la
 * vista de la siguiente, que en una app de salud es lo peor que puede pasar.
 *
 * Va a `/` y no a `/entrar`: cerrar sesión es irse, no volver a empezar.
 */
export function CerrarSesion({ className }: { className?: string }) {
  const router = useRouter();
  const [saliendo, setSaliendo] = useState(false);

  async function salir() {
    setSaliendo(true);
    try {
      await pedir("/api/v1/sesiones/actual", { method: "DELETE" });
    } catch {
      // Si la sesión ya no existía del lado del servidor, el objetivo igual se
      // cumplió: la cookie se borró y acá no hay nada que reintentar.
    }
    router.push("/");
    router.refresh();
  }

  return (
    <button
      type="button"
      onClick={salir}
      disabled={saliendo}
      className={className}
    >
      {saliendo ? "Saliendo…" : "Cerrar sesión"}
    </button>
  );
}
