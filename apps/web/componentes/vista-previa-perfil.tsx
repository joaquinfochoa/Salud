"use client";

import type { Especialidad, HorarioSemanal } from "@/lib/api";
import { enCentavos, formatearPrecio } from "@/lib/formato";
import { ESPECIALIDADES } from "./tarjeta-profesional";

const NOMBRE_MODALIDAD: Record<string, string> = {
  presencial: "En consultorio",
  telemedicina: "Videollamada",
  domicilio: "A domicilio",
};

const NOMBRE_DIA: Record<HorarioSemanal["diaSemana"], string> = {
  lunes: "lunes",
  martes: "martes",
  miercoles: "miércoles",
  jueves: "jueves",
  viernes: "viernes",
  sabado: "sábado",
  domingo: "domingo",
};

const ORDEN: HorarioSemanal["diaSemana"][] = [
  "lunes",
  "martes",
  "miercoles",
  "jueves",
  "viernes",
  "sabado",
  "domingo",
];

/** Qué parte se está editando ahora. Sirve para resaltarla. */
export type Foco = "nombre" | "matricula" | "atencion" | "precio" | "bio" | "agenda";

export type Borrador = {
  nombre: string;
  apellido: string;
  matricula: string;
  /** `null` hasta que el paso que la pide exista: mostrar el valor por defecto
   *  del select antes de llegar ahí sería atribuirle una especialidad que
   *  todavía no eligió. */
  especialidad: Especialidad | null;
  modalidades: string[];
  zona: string;
  precio: string;
  obrasSociales: string[];
  bio: string;
  horarios: HorarioSemanal[];
};

/**
 * Cómo va quedando el perfil mientras se completa el alta.
 *
 * Contesta la pregunta que un formulario largo nunca contesta: para qué sirve
 * lo que estoy escribiendo. Cada paso resalta la parte que está llenando, así
 * que la relación entre el campo y el resultado se ve, no se explica.
 *
 * **Sin datos inventados.** Lo que falta se muestra como hueco —"Tu nombre",
 * "Tu especialidad"— en color apagado, nunca como si ya existiera. Y no hay
 * estrellas ni reseñas: la API no las tiene, y mostrarle a un profesional una
 * calificación que no existe es prometerle algo que no vamos a cumplir.
 */
export function VistaPreviaPerfil({ borrador, foco }: { borrador: Borrador; foco: Foco }) {
  const nombre = [borrador.nombre.trim(), borrador.apellido.trim()]
    .filter(Boolean)
    .join(" ");
  const centavos = enCentavos(borrador.precio);

  const dias = ORDEN.filter((d) => borrador.horarios.some((h) => h.diaSemana === d));

  return (
    <aside aria-label="Vista previa de tu perfil" className="lg:sticky lg:top-8">
      <p className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        Así te va a ver un paciente
      </p>

      <div className="mt-3 rounded-2xl border border-borde bg-superficie p-6 shadow-[0_1px_2px_rgba(25,21,64,0.04),0_12px_32px_-12px_rgba(25,21,64,0.10)]">
        <Parte activa={foco === "nombre"}>
          <div className="flex items-start gap-4">
            {/* Iniciales y no una foto: la API no tiene el campo, y un
                marcador de posición con silueta sugiere que después vas a poder
                subir una. */}
            <span
              aria-hidden="true"
              className={`grid size-12 shrink-0 place-items-center rounded-full text-sm font-bold ${
                nombre ? "bg-accent text-accion" : "bg-muted text-apagado"
              }`}
            >
              {nombre ? `${borrador.nombre[0] ?? ""}${borrador.apellido[0] ?? ""}` : "—"}
            </span>
            <div className="min-w-0">
              <p
                className={`truncate text-lg font-bold tracking-tight ${
                  nombre ? "" : "text-apagado"
                }`}
              >
                {nombre || "Tu nombre"}
              </p>
              <p className="text-sm text-tinta-suave">
                <span className={borrador.especialidad ? "" : "text-apagado"}>
                  {borrador.especialidad
                    ? (ESPECIALIDADES[borrador.especialidad] ?? borrador.especialidad)
                    : "Tu especialidad"}
                </span>
                {" · "}
                <span className={borrador.zona ? "" : "text-apagado"}>
                  {borrador.zona || "Tu zona"}
                </span>
              </p>
            </div>
          </div>
        </Parte>

        <Parte activa={foco === "matricula"}>
          <p className="mt-4 inline-flex items-center gap-2 rounded-md border border-borde px-2.5 py-1">
            <span
              className={`text-sm font-semibold tabular-nums ${
                borrador.matricula ? "" : "text-apagado"
              }`}
            >
              {borrador.matricula || "Tu matrícula"}
            </span>
            <span className="text-xs uppercase tracking-wide text-tinta-suave">
              matrícula
            </span>
          </p>
        </Parte>

        <Parte activa={foco === "bio"}>
          <p
            className={`mt-4 text-sm leading-relaxed ${
              borrador.bio.trim() ? "" : "text-apagado"
            }`}
          >
            {borrador.bio.trim() ||
              "Acá va tu descripción: qué atendés y cómo trabajás."}
          </p>
        </Parte>

        <dl className="mt-5 grid gap-4 border-t border-borde pt-4 sm:grid-cols-2">
          <Parte activa={foco === "precio"}>
            <Dato etiqueta="Consulta">
              <span className={centavos === null ? "text-apagado" : ""}>
                {centavos === null ? "A definir" : formatearPrecio(centavos)}
              </span>
            </Dato>
          </Parte>

          <Parte activa={foco === "atencion"}>
            <Dato etiqueta="Modalidad">
              <span className={borrador.modalidades.length ? "" : "text-apagado"}>
                {borrador.modalidades.length
                  ? borrador.modalidades.map((m) => NOMBRE_MODALIDAD[m] ?? m).join(" · ")
                  : "A definir"}
              </span>
            </Dato>
          </Parte>

          <Parte activa={foco === "precio"}>
            <Dato etiqueta="Obras sociales">
              <span className={borrador.obrasSociales.length ? "" : "text-apagado"}>
                {borrador.obrasSociales.length
                  ? borrador.obrasSociales.join(" · ")
                  : "Solo particular"}
              </span>
            </Dato>
          </Parte>

          <Parte activa={foco === "agenda"}>
            <Dato etiqueta="Atiende">
              <span className={dias.length ? "" : "text-apagado"}>
                {dias.length
                  ? dias.map((d) => NOMBRE_DIA[d]).join(", ")
                  : "Todavía sin horarios"}
              </span>
            </Dato>
          </Parte>
        </dl>
      </div>

      <p className="mt-3 text-sm text-tinta-suave">
        {foco === "agenda"
          ? "Los horarios exactos salen de los bloques que cargues."
          : "Se actualiza mientras completás."}
      </p>
    </aside>
  );
}

/**
 * Resalta la parte que el paso actual está llenando.
 *
 * El resaltado no es solo color: el fondo cambia Y la parte se separa del resto
 * con un borde propio, porque el color nunca es la única señal.
 */
function Parte({ activa, children }: { activa: boolean; children: React.ReactNode }) {
  return (
    <div
      className={`-mx-2 rounded-lg px-2 py-1 transition-colors duration-300 ${
        activa ? "bg-accent ring-1 ring-accion/25" : ""
      }`}
    >
      {children}
    </div>
  );
}

function Dato({ etiqueta, children }: { etiqueta: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        {etiqueta}
      </dt>
      <dd className="mt-1 text-sm font-medium">{children}</dd>
    </div>
  );
}
