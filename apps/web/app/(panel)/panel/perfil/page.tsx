"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Chips } from "@/componentes/chips";
import { ESPECIALIDADES } from "@/componentes/tarjeta-profesional";
import { Campo } from "@/componentes/campo";
import { ErrorAPI, pedir, type Profesional } from "@/lib/api";
import { enCentavos, formatearPrecio } from "@/lib/formato";
import { usePanel } from "@/lib/panel";
import { aPeticionProfesional } from "@/lib/perfil";

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

/**
 * Editar el perfil. El alta ya no vive acá: es `/empezar`, paso a paso.
 *
 * Esta pantalla tenía dos modos y el de creación se llevaba puesto medio
 * archivo —el contexto opcional, una `key` para reiniciar el estado y una
 * excepción en el layout del panel— para exhibir dos campos más.
 */
export default function Perfil() {
  const { perfil } = usePanel();
  const router = useRouter();

  const [bio, setBio] = useState(perfil.bio);
  const [precio, setPrecio] = useState(formatearPrecio(perfil.precioConsultaCentavos));
  const [zona, setZona] = useState(perfil.zona);
  const [modalidades, setModalidades] = useState<string[]>(perfil.modalidades);
  const [obrasSociales, setObrasSociales] = useState<string[]>(perfil.obrasSociales);

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
    try {
      await pedir(`/api/v1/profesionales/${perfil.id}`, {
        method: "PUT",
        body: JSON.stringify(
          aPeticionProfesional(perfil, {
            bio,
            precioConsultaCentavos: centavos,
            modalidades: modalidades as Profesional["modalidades"],
            zona,
            obrasSociales,
          }),
        ),
      });
      setAviso("Perfil guardado. Los cambios ya se ven en tu perfil público.");
      router.refresh();
    } catch (e) {
      if (e instanceof ErrorAPI && e.estado === 422) {
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
        <h1 className="text-2xl font-black tracking-tight">Perfil</h1>
        <Link
          href={`/perfiles/${perfil.slug}`}
          className="text-sm font-semibold text-tinta-suave underline hover:text-accion"
        >
          Ver mi perfil público
        </Link>
      </div>

      {aviso && (
        <p
          role="alert"
          className="mt-4 rounded-lg border border-accion bg-accent px-4 py-3 text-sm"
        >
          {aviso}
        </p>
      )}

      <div className="mt-6 grid gap-5 rounded-xl border border-borde bg-superficie p-5">
        {/* De solo lectura, y se dice por qué: cambiarlas resetea la
            verificación, así que no es un campo que se edite al pasar. */}
        <div className="grid gap-1.5">
          <p className="text-sm font-semibold">Matrícula y especialidad</p>
          <p className="text-sm">
            {perfil.matricula} ·{" "}
            {ESPECIALIDADES[perfil.especialidad] ?? perfil.especialidad}
          </p>
          <p className="text-sm text-tinta-suave">
            No se editan acá: cambiarlas vuelve a poner tu matrícula en
            verificación. Escribinos si hay un error.
          </p>
        </div>

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
          {guardando ? "Guardando…" : "Guardar cambios"}
        </button>
      </div>
    </main>
  );
}
