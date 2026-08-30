"use client";

import { usePanel } from "@/lib/panel";

export default function Hoy() {
  const { usuario } = usePanel();

  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 sm:py-14">
      <h1 className="text-2xl font-black tracking-tight">Hola, {usuario.nombre}</h1>
    </main>
  );
}
