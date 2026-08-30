"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Chips } from "@/componentes/chips";
import { ESPECIALIDADES } from "@/componentes/tarjeta-profesional";
import {
  ErrorAPI,
  pedir,
  type Especialidad,
  type Profesional,
  type UsuarioActual,
} from "@/lib/api";
import { enCentavos, formatearPrecio } from "@/lib/formato";
import { usePanelOpcional } from "@/lib/panel";

const MODALIDADES = ["presencial", "telemedicina", "domicilio"] as const;

const NOMBRE_MODALIDAD: Record<string, string> = {
  presencial: "En consultorio",
  telemedicina: "Videollamada",
  domicilio: "A domicilio",
};

// Texto libre en el contrato, así que esto es una lista de sugerencias y no un
// catálogo: se puede agregar cualquier otra. Inventar ids que la API no tiene
// sería fabricar un modelo de datos del lado del front.
const OBRAS_SOCIALES = [
  "OSDE",
  "Swiss Medical",
  "Galeno",
  "Medifé",
  "OMINT",
  "IOMA",
  "PAMI",
  "Unión Personal",
] as const;

export default function Perfil() {
  const panel = usePanelOpcional();
  // La key hace que el formulario se reinicie cuando el perfil pasa de null a
  // existir: sin ella, React reusa el estado del alta y los campos quedan con
  // lo que se tipeó en vez de lo que se guardó.
  return <Formulario key={panel?.perfil.id ?? "alta"} perfil={panel?.perfil ?? null} />;
}

/**
 * El mismo formulario en dos modos.
 *
 * Sin perfil, `usePanel()` no tiene nada que dar, así que el layout deja pasar
 * esta ruta sin proveedor y la pantalla arranca vacía. Dos modos y no dos
 * pantallas: los campos son los mismos y el alta solo agrega dos. Duplicarlo es
 * cómo se agrega un campo en uno y no en el otro.
 */
function Formulario({ perfil }: { perfil: Profesional | null }) {
  const router = useRouter();
  const alta = perfil === null;

  const [bio, setBio] = useState(perfil?.bio ?? "");
  const [precio, setPrecio] = useState(
    perfil ? formatearPrecio(perfil.precioConsultaCentavos) : "",
  );
  const [zona, setZona] = useState(perfil?.zona ?? "");
  const [modalidades, setModalidades] = useState<string[]>(perfil?.modalidades ?? []);
  const [obrasSociales, setObrasSociales] = useState<string[]>(perfil?.obrasSociales ?? []);
  const [matricula, setMatricula] = useState("");
  const [especialidad, setEspecialidad] = useState<Especialidad>("psicologia");

  const [aviso, setAviso] = useState<string | null>(null);
  const [error, setError] = useState<ErrorAPI | null>(null);
  const [guardando, setGuardando] = useState(false);

  async function guardar() {
    setAviso(null);
    setError(null);

    const centavos = enCentavos(precio);
    if (centavos === null) {
      setAviso("Escribí el precio de la consulta, en pesos.");
      return;
    }
    if (modalidades.length === 0) {
      setAviso("Elegí al menos una modalidad de atención.");
      return;
    }

    setGuardando(true);
    const cuerpo = {
      nombre: perfil?.nombre ?? "",
      apellido: perfil?.apellido ?? "",
      matricula: perfil?.matricula ?? matricula,
      especialidad: perfil?.especialidad ?? especialidad,
      bio,
      precioConsultaCentavos: centavos,
      modalidades,
      zona,
      obrasSociales,
      anticipacionMinimaMin: perfil?.anticipacionMinimaMin ?? 120,
      horizonteDias: perfil?.horizonteDias ?? 60,
    };

    try {
      if (alta) {
        // El nombre del alta sale de la cuenta: el perfil profesional es de la
        // misma persona, y pedirlo de nuevo invita a que no coincidan.
        const yo = await pedir<UsuarioActual>("/api/v1/usuarios/yo");
        await pedir("/api/v1/profesionales", {
          method: "POST",
          body: JSON.stringify({ ...cuerpo, nombre: yo.nombre, apellido: yo.apellido }),
        });
        router.push("/panel");
        router.refresh();
        return;
      }

      await pedir(`/api/v1/profesionales/${perfil.id}`, {
        method: "PUT",
        body: JSON.stringify(cuerpo),
      });
      setAviso("Perfil guardado. Los cambios ya se ven en tu perfil público.");
      router.refresh();
    } catch (e) {
      if (e instanceof ErrorAPI && e.estado === 409) {
        setAviso("Ya hay un perfil con esa matrícula.");
      } else if (e instanceof ErrorAPI && e.estado === 422) {
        setError(e);
      } else {
        setAviso("No pudimos guardar el perfil. Probá de nuevo.");
      }
    } finally {
      setGuardando(false);
    }
  }

  return (
    <main className="mx-auto w-full max-w-2xl px-4 py-10 sm:px-6 sm:py-14">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-black tracking-tight">
          {alta ? "Creá tu perfil" : "Perfil"}
        </h1>
        {perfil && (
          <Link
            href={`/perfiles/${perfil.slug}`}
            className="text-sm font-semibold text-tinta-suave underline hover:text-accion"
          >
            Ver mi perfil público
          </Link>
        )}
      </div>

      {alta && (
        <p className="mt-2 text-tinta-suave">
          Con esto ya podés cargar tus horarios y empezar a recibir turnos.
        </p>
      )}

      {aviso && (
        <p
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso}
        </p>
      )}

      <div className="mt-6 grid gap-5 rounded-xl border border-borde bg-superficie p-5">
        {alta ? (
          <>
            <Campo
              nombre="matricula"
              etiqueta="Matrícula"
              valor={matricula}
              onCambiar={setMatricula}
              ayuda="Como figura en tu credencial: MN 98234, M.P. 12345."
              error={error}
            />
            <div className="grid gap-1.5">
              <label htmlFor="especialidad" className="text-sm font-semibold">
                Especialidad
              </label>
              <select
                id="especialidad"
                value={especialidad}
                onChange={(e) => setEspecialidad(e.target.value as Especialidad)}
                className="h-11 rounded-lg border border-borde bg-superficie px-3"
              >
                {Object.entries(ESPECIALIDADES).map(([valor, texto]) => (
                  <option key={valor} value={valor}>
                    {texto}
                  </option>
                ))}
              </select>
            </div>
          </>
        ) : (
          // De solo lectura, y se dice por qué: cambiarlas resetea la
          // verificación, así que no es un campo que se edite al pasar.
          <div className="grid gap-1.5">
            <p className="text-sm font-semibold">Matrícula y especialidad</p>
            <p className="text-sm">
              {perfil.matricula} · {ESPECIALIDADES[perfil.especialidad] ?? perfil.especialidad}
            </p>
            <p className="text-sm text-tinta-suave">
              No se editan acá: cambiarlas vuelve a poner tu matrícula en verificación.
              Escribinos si hay un error.
            </p>
          </div>
        )}

        <div className="grid gap-1.5">
          <label htmlFor="bio" className="text-sm font-semibold">
            Descripción
          </label>
          <textarea
            id="bio"
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            rows={4}
            className="rounded-lg border border-borde p-3"
          />
          <p className="text-sm text-tinta-suave">
            Es lo primero que lee un paciente. Contá qué atendés y cómo trabajás.
          </p>
        </div>

        <div className="grid gap-5 sm:grid-cols-2">
          <Campo
            nombre="precio"
            etiqueta="Precio de la consulta"
            valor={precio}
            onCambiar={setPrecio}
            // Se formatea al salir del campo y no en cada tecla: reformatear
            // mientras se escribe mueve el cursor de lugar.
            onSalir={() => {
              const c = enCentavos(precio);
              if (c !== null) setPrecio(formatearPrecio(c));
            }}
            ayuda="En pesos."
            error={error}
          />
          <Campo
            nombre="zona"
            etiqueta="Zona"
            valor={zona}
            onCambiar={setZona}
            ayuda="Dónde atendés: CABA, Rosario, Palermo."
            error={error}
          />
        </div>

        <Chips
          etiqueta="Modalidades"
          opciones={MODALIDADES.map((m) => NOMBRE_MODALIDAD[m])}
          elegidas={modalidades.map((m) => NOMBRE_MODALIDAD[m] ?? m)}
          onCambiar={(elegidas) =>
            setModalidades(
              MODALIDADES.filter((m) => elegidas.includes(NOMBRE_MODALIDAD[m])),
            )
          }
        />

        <Chips
          etiqueta="Obras sociales"
          opciones={OBRAS_SOCIALES}
          elegidas={obrasSociales}
          onCambiar={setObrasSociales}
          libre
        />

        <button
          type="button"
          onClick={guardar}
          disabled={guardando}
          className="mt-1 rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva disabled:opacity-60"
        >
          {guardando ? "Guardando…" : alta ? "Crear mi perfil" : "Guardar cambios"}
        </button>
      </div>
    </main>
  );
}

function Campo({
  nombre,
  etiqueta,
  valor,
  onCambiar,
  onSalir,
  ayuda,
  error,
}: {
  nombre: string;
  etiqueta: string;
  valor: string;
  onCambiar: (valor: string) => void;
  onSalir?: () => void;
  ayuda?: string;
  error: ErrorAPI | null;
}) {
  const mensaje = error?.porCampo(nombre);
  const idAyuda = `${nombre}-ayuda`;
  const idError = `${nombre}-error`;

  return (
    <div className="grid gap-1.5">
      <label htmlFor={nombre} className="text-sm font-semibold">
        {etiqueta}
      </label>
      <input
        id={nombre}
        value={valor}
        onChange={(e) => onCambiar(e.target.value)}
        onBlur={onSalir}
        aria-invalid={mensaje ? true : undefined}
        aria-describedby={mensaje ? idError : ayuda ? idAyuda : undefined}
        className={`h-11 rounded-lg border px-3 ${
          mensaje ? "border-destructive" : "border-borde"
        }`}
      />
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
