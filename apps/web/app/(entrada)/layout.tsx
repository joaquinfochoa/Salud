import { Encabezado } from "@/componentes/encabezado";

/**
 * Entrar, crear cuenta y el alta del profesional.
 *
 * Las tres estaban en la raíz, así que no tenían ningún encabezado: alguien a
 * mitad del alta no tenía forma de volver a ninguna parte.
 *
 * El encabezado va compacto, solo con el logo. Son tres flujos con una sola
 * cosa para hacer, y ofrecer "Buscar" o "Mis turnos" en el medio invita a
 * abandonarlos justo cuando la persona estaba por terminar.
 */
export default function LayoutEntrada({ children }: LayoutProps<"/">) {
  return (
    <>
      <Encabezado compacto />
      {children}
    </>
  );
}
