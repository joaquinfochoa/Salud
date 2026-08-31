"use client";

/**
 * El marco de un paso del onboarding: dónde estás, qué se te pide, y cómo
 * seguir.
 *
 * Todos los pasos comparten esta forma a propósito. Un onboarding donde cada
 * pantalla se ve distinta obliga a reorientarse en cada una; acá lo único que
 * cambia es el contenido del medio, así que el botón de avanzar está siempre en
 * el mismo lugar.
 */
export function Paso({
  numero,
  total,
  titulo,
  ayuda,
  aviso,
  avanzar,
  textoAvanzar = "Continuar",
  puedeAvanzar,
  enviando = false,
  volver,
  masTarde,
  children,
}: {
  numero: number;
  total: number;
  titulo: string;
  ayuda?: string;
  aviso?: React.ReactNode;
  avanzar: () => void;
  textoAvanzar?: string;
  puedeAvanzar: boolean;
  enviando?: boolean;
  volver?: () => void;
  /** Solo en los pasos opcionales. */
  masTarde?: () => void;
  children: React.ReactNode;
}) {
  return (
    // Sin <main> ni centrado propio: el marco de dos columnas —pasos a la
    // izquierda, vista previa a la derecha— lo pone la pantalla que lo usa.
    <div>
      <div className="flex items-center justify-between gap-4">
        <p className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
          Paso {numero} de {total}
        </p>
        {masTarde && (
          <button
            type="button"
            onClick={masTarde}
            className="text-sm font-semibold text-tinta-suave underline hover:text-tinta"
          >
            Completar más tarde
          </button>
        )}
      </div>

      {/* La barra dice cuánto falta. Un onboarding sin progreso visible se
          siente infinito, y es donde la gente abandona. */}
      <div
        role="progressbar"
        aria-valuenow={numero}
        aria-valuemin={1}
        aria-valuemax={total}
        aria-label={`Paso ${numero} de ${total}`}
        className="mt-2 h-1 w-full overflow-hidden rounded-full bg-borde"
      >
        <div
          className="h-full rounded-full bg-accion transition-[width] duration-300"
          style={{ width: `${(numero / total) * 100}%` }}
        />
      </div>

      <h1 className="mt-6 text-2xl font-black tracking-tight">{titulo}</h1>
      {ayuda && <p className="mt-2 text-tinta-suave">{ayuda}</p>}

      {aviso && (
        <div
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso}
        </div>
      )}

      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (puedeAvanzar && !enviando) avanzar();
        }}
        className="mt-6 grid gap-5 rounded-xl border border-borde bg-superficie p-5"
      >
        {children}

        <button
          type="submit"
          // Deshabilitado y no oculto: que el botón esté ahí, apagado, dice que
          // falta algo. Esconderlo deja a la persona buscando cómo seguir.
          disabled={!puedeAvanzar || enviando}
          className="mt-1 rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
        >
          {enviando ? "Guardando…" : textoAvanzar}
        </button>
      </form>

      {volver && (
        <button
          type="button"
          onClick={volver}
          className="mt-5 text-sm font-semibold text-tinta-suave hover:text-accion"
        >
          ← Volver
        </button>
      )}
    </div>
  );
}
