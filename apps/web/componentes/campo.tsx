"use client";

import type { ErrorAPI } from "@/lib/api";

/**
 * Un campo de formulario controlado, con su etiqueta, su ayuda y su error.
 *
 * Existe porque iba por su cuarta copia. Es la variante controlada —el valor
 * vive en el estado de quien lo usa— que necesitan las pantallas donde el dato
 * sobrevive a la navegación. Los formularios que envían con `FormData` y no
 * guardan estado siguen con su input propio: unificarlos agregaría ramas sin
 * ganar nada.
 */
export function Campo({
  nombre,
  etiqueta,
  valor,
  onCambiar,
  onSalir,
  tipo = "text",
  autoComplete,
  ayuda,
  error,
}: {
  nombre: string;
  etiqueta: string;
  valor: string;
  onCambiar: (valor: string) => void;
  onSalir?: () => void;
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
        value={valor}
        autoComplete={autoComplete}
        onChange={(e) => onCambiar(e.target.value)}
        onBlur={onSalir}
        // Sin aria-invalid y aria-describedby, un lector de pantalla anuncia el
        // campo sin decir que está mal ni por qué.
        aria-invalid={mensaje ? true : undefined}
        aria-describedby={mensaje ? idError : ayuda ? idAyuda : undefined}
        className={`h-11 w-full rounded-lg border px-3 ${
          mensaje ? "border-destructive" : "border-borde"
        }`}
      />
      {/* El mensaje va debajo de SU campo, no en un cartel arriba: un error de
          formulario lejos del campo obliga a adivinar cuál falló. */}
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
