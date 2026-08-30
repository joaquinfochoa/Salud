import type { TurnoConPaciente } from "@/lib/api";
import { formatearHora } from "@/lib/formato";

const MODALIDADES: Record<string, string> = {
  telemedicina: "Videollamada",
  presencial: "En consultorio",
  domicilio: "A domicilio",
};

/**
 * Un turno visto desde el lado del profesional: quién viene y a qué.
 *
 * El motivo de consulta se muestra acá y en la agenda, las dos detrás de sesión
 * y solo para el dueño del perfil. En ninguna página pública y en ningún
 * `<title>`: es dato de salud bajo Ley 25.326.
 */
export function FilaTurnoProfesional({
  turno,
  pasado = false,
}: {
  turno: TurnoConPaciente;
  pasado?: boolean;
}) {
  const cancelado = turno.estado === "cancelado";

  return (
    <li
      className={`flex flex-wrap items-baseline gap-x-4 gap-y-1 rounded-xl border border-borde bg-superficie p-4 ${
        pasado ? "opacity-55" : ""
      }`}
    >
      <span
        className={`font-horas text-2xl font-black tabular-nums tracking-tight ${
          cancelado ? "text-apagado line-through decoration-2" : "text-libre"
        }`}
      >
        {formatearHora(turno.inicio)}
      </span>

      <span className={`font-semibold ${cancelado ? "line-through" : ""}`}>
        {turno.paciente.nombre} {turno.paciente.apellido}
      </span>

      <span className="text-sm text-tinta-suave">
        {MODALIDADES[turno.modalidad] ?? turno.modalidad}
      </span>

      {turno.motivo && (
        <p className="w-full text-sm text-tinta-suave">{turno.motivo}</p>
      )}

      {cancelado && (
        // canceladoPor es un id de Usuario. Si coincide con pacienteId canceló
        // el paciente; si no, fue el propio profesional desde otro dispositivo.
        // Es la diferencia entre "se me cayó un paciente" y "yo lo cancelé", y
        // sin decirlo el profesional no sabe cuál de las dos pasó.
        <p className="w-full text-sm font-semibold text-destructive">
          Cancelado por {turno.canceladoPor === turno.pacienteId ? "el paciente" : "vos"}
        </p>
      )}
    </li>
  );
}
