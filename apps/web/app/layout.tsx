import type { Metadata } from "next";
import { Archivo, Public_Sans } from "next/font/google";
import "./globals.css";

// Archivo solo para las horas. Es el objeto principal de la interfaz —lo que la
// persona vino a buscar— y el único elemento con tipografía propia.
const archivo = Archivo({
  variable: "--fuente-horas",
  subsets: ["latin"],
  weight: ["700", "900"],
});

// Public Sans para todo lo demás. `shadcn init` había agregado Geist como
// --font-sans; se sacó para no tener dos tipografías de texto compitiendo por
// el mismo rol.
const publicSans = Public_Sans({
  variable: "--fuente-texto",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "Salud — turnos con profesionales verificados",
    template: "%s · Salud",
  },
  description:
    "Encontrá psicólogos, kinesiólogos y odontólogos con matrícula verificada, y reservá un turno en el horario que te queda bien.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    // lang="es" no es cosmético: sin él un lector de pantalla lee el español
    // con fonética inglesa y no se entiende nada.
    <html
      lang="es"
      className={`${archivo.variable} ${publicSans.variable} h-full antialiased`}
    >
      {/* Sin navegación acá: cada grupo de rutas trae la suya. El público y
          el del paciente comparten el mismo encabezado; el panel tiene tabs
          abajo en móvil y barra lateral en escritorio, y es eso lo que hace
          que se sientan dos productos distintos dentro de una sola app. */}
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
