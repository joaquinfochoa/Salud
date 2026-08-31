import { Encabezado } from "@/componentes/encabezado";
import { NavegacionPaciente } from "@/componentes/navegacion-paciente";

/**
 * El área del paciente: sus turnos y su cuenta.
 *
 * Tiene navegación propia, con otra forma que la del panel del profesional. No
 * es un capricho de estilo: el profesional vive acá adentro todos los días y el
 * paciente entra cada varios meses. Dos usos distintos no se sirven con la
 * misma estructura.
 */
export default function LayoutPaciente({ children }: LayoutProps<"/">) {
  return (
    <>
      <Encabezado />
      <NavegacionPaciente />
      {children}
    </>
  );
}
