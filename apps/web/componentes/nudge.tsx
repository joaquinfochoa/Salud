import Link from "next/link";
import type { Profesional } from "@/lib/api";

/**
 * Qué le falta al perfil para poder recibir turnos.
 *
 * Existe porque **un profesional sin horarios cargados no aparece con
 * disponibilidad en la búsqueda**, y nada se lo dice. Lo descubrimos cuando el
 * seed dejaba a los cuatro sin agenda y el listado mostraba cuatro tarjetas
 * vacías: el mismo agujero que tendría alguien que se registra de verdad.
 *
 * Se muestra uno solo, el más bloqueante: tres banners apilados son ruido, y el
 * que importa es el que impide cobrar.
 */
export function Nudge({
  perfil,
  tieneHorarios,
}: {
  perfil: Profesional;
  tieneHorarios: boolean;
}) {
  if (!tieneHorarios) {
    return (
      <Aviso
        texto="Todavía no cargaste tus horarios, así que nadie puede reservarte un turno."
        accion={{ texto: "Configurar", href: "/panel/horarios" }}
      />
    );
  }

  if (!perfil.bio.trim()) {
    return (
      <Aviso
        texto="Tu perfil no tiene descripción. Es lo primero que lee un paciente."
        accion={{ texto: "Completar", href: "/panel/perfil" }}
      />
    );
  }

  if (perfil.verificacion === "pendiente") {
    // Informativo y sin botón, a propósito: hoy nada mueve ese estado —la
    // integración con el registro de matrículas es su propia etapa— y un botón
    // que no hace nada es peor que ninguno.
    return <Aviso texto="Tu matrícula está pendiente de verificación." />;
  }

  return null;
}

function Aviso({
  texto,
  accion,
}: {
  texto: string;
  accion?: { texto: string; href: string };
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-accion bg-accent px-4 py-3">
      <p className="text-sm">{texto}</p>
      {accion && (
        <Link
          href={accion.href}
          className="shrink-0 rounded-lg bg-accion px-4 py-2 text-sm font-bold text-white transition-colors hover:bg-accion-viva"
        >
          {accion.texto}
        </Link>
      )}
    </div>
  );
}
