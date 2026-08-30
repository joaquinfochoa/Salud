import { Encabezado } from "@/componentes/encabezado";
import { Pie } from "@/componentes/pie";

export default function LayoutPublico({ children }: LayoutProps<"/">) {
  return (
    <>
      <Encabezado />
      {children}
      <Pie />
    </>
  );
}
