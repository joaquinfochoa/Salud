import { Encabezado } from "@/componentes/encabezado";

// Sin pie: es una pantalla de gestión, no de descubrimiento. Alguien que entra
// a ver sus turnos ya está adentro, y no hay nada que ofrecerle abajo.
export default function LayoutPaciente({ children }: LayoutProps<"/">) {
  return (
    <>
      <Encabezado />
      {children}
    </>
  );
}
