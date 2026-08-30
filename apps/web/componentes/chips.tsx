"use client";

import { useState } from "react";

/**
 * Opciones que se prenden y se apagan: modalidades y obras sociales.
 *
 * `aria-pressed` y no solo una clase: sin él, un lector de pantalla lee "OSDE,
 * botón" tanto si está elegida como si no. Y el estado elegido lleva además un
 * check, porque el color nunca es la única señal.
 */
export function Chips({
  etiqueta,
  opciones,
  elegidas,
  onCambiar,
  /** Deja agregar una opción que no está en la lista. */
  libre = false,
}: {
  etiqueta: string;
  opciones: readonly string[];
  elegidas: string[];
  onCambiar: (elegidas: string[]) => void;
  libre?: boolean;
}) {
  const [nueva, setNueva] = useState("");

  // Las elegidas que no están en la lista sugerida se muestran igual: alguien
  // que agregó "Swiss Medical" tiene que poder verla y sacarla.
  const todas = [...opciones, ...elegidas.filter((e) => !opciones.includes(e))];

  function alternar(opcion: string) {
    onCambiar(
      elegidas.includes(opcion)
        ? elegidas.filter((e) => e !== opcion)
        : [...elegidas, opcion],
    );
  }

  function agregar() {
    const limpia = nueva.trim();
    if (limpia && !elegidas.includes(limpia)) onCambiar([...elegidas, limpia]);
    setNueva("");
  }

  return (
    <fieldset className="grid gap-2">
      <legend className="text-sm font-semibold">{etiqueta}</legend>

      <ul className="flex flex-wrap gap-2">
        {todas.map((opcion) => {
          const activa = elegidas.includes(opcion);
          return (
            <li key={opcion}>
              <button
                type="button"
                aria-pressed={activa}
                onClick={() => alternar(opcion)}
                className={`rounded-full border px-3 py-1.5 text-sm font-semibold transition-colors ${
                  activa
                    ? "border-accion bg-accion text-white"
                    : "border-borde hover:border-accion hover:bg-accent"
                }`}
              >
                {activa && <span aria-hidden="true">✓ </span>}
                {opcion}
              </button>
            </li>
          );
        })}
      </ul>

      {libre && (
        <div className="mt-1 flex gap-2">
          <input
            value={nueva}
            onChange={(e) => setNueva(e.target.value)}
            onKeyDown={(e) => {
              // Enter agrega en vez de enviar el formulario entero: escribir
              // una obra social y perder todo lo demás sería el peor final.
              if (e.key === "Enter") {
                e.preventDefault();
                agregar();
              }
            }}
            placeholder="Agregar otra"
            aria-label={`Agregar a ${etiqueta.toLowerCase()}`}
            className="h-10 flex-1 rounded-lg border border-borde px-3 text-sm"
          />
          <button
            type="button"
            onClick={agregar}
            className="rounded-lg border border-borde px-4 text-sm font-semibold transition-colors hover:border-accion hover:bg-accent"
          >
            Agregar
          </button>
        </div>
      )}
    </fieldset>
  );
}
