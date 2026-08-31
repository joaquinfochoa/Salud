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
 * lo que estoy escribiendo. Cada paso resalta —y agranda— la parte que está
 * llenando, así que la relación entre el campo y el resultado se ve en vez de
 * explicarse.
 *
 * **Solo en escritorio.** En un teléfono, ponerla debajo del formulario
 * significa que nunca se ve mientras se escribe, que es justo cuando sirve: el
 * teclado tapa media pantalla y la tarjeta queda fuera de cuadro. Un elemento
 * que solo aparece cuando ya no hace falta es peso, no ayuda.
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
  const modalidades = borrador.modalidades
    .map((m) => NOMBRE_MODALIDAD[m] ?? m)
    .join(" · ");

  return (
    <aside
      aria-label="Vista previa de tu perfil"
      // surgir: entra con la página en vez de aparecer de golpe. El bloque de
      // prefers-reduced-motion de globals.css ya la desactiva.
      className="surgir surgir-tarde @container hidden md:sticky md:top-8 md:block"
    >
      <p className="text-xs font-semibold uppercase tracking-wide text-tinta-suave">
        Así te va a ver un paciente
      </p>

      <div className="mt-3 overflow-hidden rounded-2xl border border-borde bg-superficie shadow-[0_1px_2px_rgba(25,21,64,0.04),0_16px_40px_-16px_rgba(25,21,64,0.14)]">
        <div className="p-5">
          <Parte activa={foco === "nombre"}>
            <div className="flex items-center gap-3.5">
              <Redondel nombre={borrador.nombre} apellido={borrador.apellido} />
              <div className="min-w-0">
                <p
                  className={`text-base font-bold leading-tight tracking-tight text-balance ${
                    nombre ? "" : "text-apagado"
                  }`}
                >
                  {nombre || "Tu nombre"}
                </p>
                <p className="mt-0.5 text-sm text-tinta-suave">
                  <span className={borrador.especialidad ? "" : "text-apagado"}>
                    {borrador.especialidad
                      ? (ESPECIALIDADES[borrador.especialidad] ?? borrador.especialidad)
                      : "Tu especialidad"}
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
        </div>

        <div
          className={`flex items-center gap-4 border-y border-borde px-5 py-3.5 transition-colors duration-300 ${
            foco === "atencion" ? "bg-accent" : "bg-fondo"
          }`}
        >
          <div className="flex min-w-0 flex-1 items-center gap-2.5">
            <Pin />
            <p className={`truncate text-sm font-medium ${borrador.zona ? "" : "text-apagado"}`}>
              {borrador.zona || "Tu zona"}
            </p>
          </div>
          <Mapita />
        </div>

        <dl className="grid gap-4 p-5 @sm:grid-cols-2">
          <Parte activa={foco === "precio"}>
            <Dato etiqueta="Consulta">
              <span className={centavos === null ? "text-apagado" : ""}>
                {centavos === null ? "A definir" : formatearPrecio(centavos)}
              </span>
            </Dato>
          </Parte>

          <Parte activa={foco === "agenda"}>
            <Dato etiqueta="Atiende">
              <span className={dias.length ? "" : "text-apagado"}>
                {dias.length ? dias.map((d) => NOMBRE_DIA[d]).join(", ") : "Sin horarios"}
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

          <Parte activa={foco === "atencion"}>
            <Dato etiqueta="Modalidad">
              <span className={modalidades ? "" : "text-apagado"}>
                {modalidades || "A definir"}
              </span>
            </Dato>
          </Parte>
        </dl>
      </div>

    </aside>
  );
}

/**
 * El redondel del profesional: iniciales cuando hay nombre, un ícono cuando
 * todavía no.
 *
 * El ícono es un SVG acá adentro y no una librería: es un dibujo, no una
 * dependencia. Cambia de color cuando el nombre aparece, que es el primer
 * momento del alta en que la tarjeta deja de estar vacía.
 */
function Redondel({ nombre, apellido }: { nombre: string; apellido: string }) {
  const iniciales = `${nombre.trim()[0] ?? ""}${apellido.trim()[0] ?? ""}`.toUpperCase();

  return (
    <span
      aria-hidden="true"
      className={`grid size-12 shrink-0 place-items-center rounded-full text-sm font-bold transition-colors duration-300 ${
        iniciales ? "bg-accion text-white" : "bg-muted text-apagado"
      }`}
    >
      {iniciales || (
        <svg viewBox="0 0 24 24" fill="none" className="size-6">
          <circle cx="12" cy="8.5" r="3.5" fill="currentColor" />
          <path
            d="M4.5 20a7.5 7.5 0 0 1 15 0"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
          />
        </svg>
      )}
    </span>
  );
}

/**
 * Un mapa genérico, dibujado acá adentro.
 *
 * Es decoración, no un dato: la API guarda una zona escrita a mano —"CABA",
 * "Vicente López"— y no una dirección ni coordenadas, así que estas calles no
 * son las de ningún lado. Deliberadamente abstracto: sin nombres, sin marcador
 * sobre un punto, sin norte. Lo que dice dónde atendés es el texto de al lado.
 *
 * SVG en el archivo y no una imagen ni un proveedor de mapas: son 12 líneas,
 * cero dependencias y cero pedidos a un tercero desde una pantalla de salud.
 */
function Mapita() {
  return (
    <svg
      viewBox="0 0 64 44"
      aria-hidden="true"
      className="h-11 w-16 shrink-0 overflow-hidden rounded-lg border border-borde"
    >
      {/* La tierra primero y las calles encima en blanco: al revés se lee como
          una grilla de fichas, no como un mapa. */}
      <rect width="64" height="44" className="fill-muted" />
      <rect x="26" y="24" width="13" height="13" rx="1" className="fill-libre/25" />
      <path
        d="M0 20h64M24 0v44M44 0v44M0 34h64"
        className="stroke-superficie"
        strokeWidth="3"
      />
      {/* La diagonal: en cualquier ciudad hay una, y es lo que termina de
          separar el dibujo de una grilla. */}
      <path d="M-4 46 46-4" className="stroke-superficie" strokeWidth="4" />
    </svg>
  );
}

function Pin() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
      className="size-4 shrink-0 text-accion"
    >
      <path
        d="M12 21s7-5.5 7-11a7 7 0 1 0-14 0c0 5.5 7 11 7 11Z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      <circle cx="12" cy="10" r="2.5" fill="currentColor" />
    </svg>
  );
}

/**
 * Resalta y agranda la parte que el paso actual está llenando.
 *
 * Tres señales a la vez —fondo, borde y tamaño— y no solo color: es la regla
 * que vale en toda la app. `origin-left` hace que crezca desde su borde
 * izquierdo en vez de empujar el contenido hacia los dos costados.
 */
function Parte({ activa, children }: { activa: boolean; children: React.ReactNode }) {
  return (
    <div
      // scale-[1.02] y no más: con 1.03 el borde derecho del resaltado llegaba
      // justo al borde de la tarjeta. Medido, no estimado.
      className={`-mx-2 origin-left rounded-lg px-2 py-1 transition-all duration-300 ease-out ${
        activa ? "scale-[1.02] bg-accent ring-1 ring-accion/25" : "scale-100"
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
