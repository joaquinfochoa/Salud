"use client";

import type { HorarioSemanal } from "@/lib/api";

// Una sola lista, dos presentaciones. Duplicarla para móvil y escritorio es
// cómo se agrega un día en un lado y no en el otro.
const DIAS: { valor: HorarioSemanal["diaSemana"]; texto: string }[] = [
  { valor: "lunes", texto: "Lunes" },
  { valor: "martes", texto: "Martes" },
  { valor: "miercoles", texto: "Miércoles" },
  { valor: "jueves", texto: "Jueves" },
  { valor: "viernes", texto: "Viernes" },
  { valor: "sabado", texto: "Sábado" },
  { valor: "domingo", texto: "Domingo" },
];

const MODALIDADES: { valor: HorarioSemanal["modalidad"]; texto: string }[] = [
  { valor: "presencial", texto: "En consultorio" },
  { valor: "telemedicina", texto: "Videollamada" },
  { valor: "domicilio", texto: "A domicilio" },
];

const DURACIONES = [20, 30, 40, 45, 50, 60, 90];

/**
 * Los siete días con sus bloques.
 *
 * `<input type="time">` y `<select>` nativos: es la restricción de cero
 * dependencias, y el picker del sistema operativo funciona mejor que cualquier
 * reimplementación —sobre todo en el teléfono, que es donde esta pantalla se
 * usa menos pero se usa.
 */
export function EditorSemana({
  horarios,
  onCambiar,
}: {
  horarios: HorarioSemanal[];
  onCambiar: (horarios: HorarioSemanal[]) => void;
}) {
  function actualizar(indice: number, cambio: Partial<HorarioSemanal>) {
    onCambiar(horarios.map((h, i) => (i === indice ? { ...h, ...cambio } : h)));
  }

  return (
    <div className="grid gap-3">
      {DIAS.map(({ valor, texto }) => {
        // Se guardan los índices sobre el arreglo completo: sin eso, editar el
        // primer bloque del miércoles editaría el primero del lunes.
        const delDia = horarios
          .map((h, indice) => ({ h, indice }))
          .filter(({ h }) => h.diaSemana === valor);

        return (
          <section
            key={valor}
            className="rounded-xl border border-borde bg-superficie p-4"
          >
            <div className="flex items-center justify-between gap-3">
              <h2 className="font-bold">{texto}</h2>
              <button
                type="button"
                // Siete botones que dicen "+" son siete botones que un lector
                // de pantalla no puede distinguir.
                aria-label={`Agregar bloque a ${texto}`}
                onClick={() =>
                  onCambiar([
                    ...horarios,
                    {
                      diaSemana: valor,
                      desde: "09:00",
                      hasta: "13:00",
                      duracionMin: 50,
                      modalidad: "presencial",
                    },
                  ])
                }
                className="rounded-lg border border-borde px-3 py-1.5 text-sm font-semibold transition-colors hover:border-accion hover:bg-accent"
              >
                Agregar bloque
              </button>
            </div>

            {delDia.length === 0 ? (
              <p className="mt-3 text-sm text-tinta-suave">No atendés este día.</p>
            ) : (
              <ul className="mt-3 grid gap-3">
                {delDia.map(({ h, indice }) => (
                  <li
                    key={indice}
                    className="grid gap-3 rounded-lg border border-borde p-3 sm:grid-cols-[auto_auto_1fr_auto] sm:items-end"
                  >
                    <Campo
                      etiqueta="Desde"
                      id={`desde-${indice}`}
                      tipo="time"
                      valor={h.desde}
                      onCambiar={(desde) => actualizar(indice, { desde })}
                    />
                    <Campo
                      etiqueta="Hasta"
                      id={`hasta-${indice}`}
                      tipo="time"
                      valor={h.hasta}
                      onCambiar={(hasta) => actualizar(indice, { hasta })}
                    />

                    <div className="grid gap-3 sm:grid-cols-2">
                      <Select
                        etiqueta="Cada"
                        id={`duracion-${indice}`}
                        valor={String(h.duracionMin)}
                        opciones={DURACIONES.map((d) => ({ valor: String(d), texto: `${d} min` }))}
                        onCambiar={(v) => actualizar(indice, { duracionMin: Number(v) })}
                      />
                      <Select
                        etiqueta="Modalidad"
                        id={`modalidad-${indice}`}
                        valor={h.modalidad}
                        opciones={MODALIDADES}
                        onCambiar={(v) =>
                          actualizar(indice, { modalidad: v as HorarioSemanal["modalidad"] })
                        }
                      />
                    </div>

                    <button
                      type="button"
                      aria-label={`Quitar el bloque de ${texto} de ${h.desde} a ${h.hasta}`}
                      onClick={() => onCambiar(horarios.filter((_, i) => i !== indice))}
                      className="h-11 rounded-lg px-3 text-sm font-semibold text-tinta-suave underline hover:text-tinta"
                    >
                      Quitar
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>
        );
      })}
    </div>
  );
}

function Campo({
  etiqueta,
  id,
  tipo,
  valor,
  onCambiar,
}: {
  etiqueta: string;
  id: string;
  tipo: string;
  valor: string;
  onCambiar: (valor: string) => void;
}) {
  return (
    <div className="grid gap-1.5">
      <label htmlFor={id} className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        {etiqueta}
      </label>
      <input
        id={id}
        type={tipo}
        value={valor}
        onChange={(e) => onCambiar(e.target.value)}
        className="h-11 rounded-lg border border-borde px-3 tabular-nums"
      />
    </div>
  );
}

function Select({
  etiqueta,
  id,
  valor,
  opciones,
  onCambiar,
}: {
  etiqueta: string;
  id: string;
  valor: string;
  opciones: { valor: string; texto: string }[];
  onCambiar: (valor: string) => void;
}) {
  return (
    <div className="grid gap-1.5">
      <label htmlFor={id} className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        {etiqueta}
      </label>
      <select
        id={id}
        value={valor}
        onChange={(e) => onCambiar(e.target.value)}
        className="h-11 rounded-lg border border-borde bg-superficie px-3"
      >
        {opciones.map((o) => (
          <option key={o.valor} value={o.valor}>
            {o.texto}
          </option>
        ))}
      </select>
    </div>
  );
}
