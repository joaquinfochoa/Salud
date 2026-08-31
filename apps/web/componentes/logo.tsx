import Link from "next/link";

/**
 * La marca.
 *
 * Un texto por ahora, no una imagen: cuando exista un logo de verdad se cambia
 * acá y aparece en los tres lugares que lo usan.
 *
 * Dos pesos y no uno. "salud" es lo que se recuerda y "app" lo acompaña, así
 * que la marca se lee como una sola palabra con jerarquía en vez de dos
 * sueltas.
 */
export function Logo() {
  return (
    <Link href="/" className="text-lg tracking-tight">
      <span className="font-normal text-tinta-suave">app </span>
      <span className="font-black">salud</span>
    </Link>
  );
}
