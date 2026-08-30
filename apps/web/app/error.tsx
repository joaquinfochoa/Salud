"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTransition } from "react";

/**
 * Lo que se ve cuando algo revienta del lado del servidor.
 *
 * Casi siempre es la API sin responder: las pantallas públicas la consultan
 * desde el Server Component, y si el fetch falla la página entera tira. Sin
 * este archivo eso es la pantalla pelada de Next —en producción, "Application
 * error"— que no dice qué pasó ni deja hacer nada.
 *
 * Un solo error.tsx en la raíz cubre todas las rutas de app/. No hay
 * global-error.tsx: ese solo entra si revienta el layout raíz, que no consulta
 * nada y no tiene de qué fallar.
 */
export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const router = useRouter();
  const [reintentando, empezar] = useTransition();

  // reset() solo limpia el error boundary del cliente: el render del servidor
  // que falló queda cacheado en el router, así que sin el refresh el botón
  // vuelve a mostrar el mismo error aunque la API ya haya vuelto. Un botón que
  // dice "Reintentar" y no reintenta es peor que no tenerlo.
  function reintentar() {
    empezar(() => {
      router.refresh();
      reset();
    });
  }

  return (
    <main className="mx-auto w-full max-w-md px-4 py-20 text-center sm:px-6">
      <h1 className="text-2xl font-black tracking-tight">No pudimos cargar esto</h1>
      <p className="mt-3 text-tinta-suave">
        Fue un problema nuestro, no tuyo. Probá de nuevo en un momento.
      </p>

      {/* El mensaje del error NO se muestra: en producción Next ya lo reemplaza
          por un digest, y en desarrollo largarlo en pantalla es filtrar rutas
          internas. El digest sí, que es lo que permite encontrarlo en los logs
          si alguien lo reporta. */}

      <div className="mt-8 flex flex-col items-center gap-3">
        <button
          type="button"
          onClick={reintentar}
          disabled={reintentando}
          className="rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
        >
          {reintentando ? "Reintentando…" : "Reintentar"}
        </button>
        <Link href="/" className="text-sm text-tinta-suave hover:text-accion">
          Volver a la búsqueda
        </Link>
      </div>

      {error.digest && (
        <p className="mt-10 font-mono text-xs text-apagado">Error {error.digest}</p>
      )}
    </main>
  );
}
