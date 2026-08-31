/**
 * El panel del profesional, dibujado.
 *
 * Es una maqueta y no una captura: una imagen quedaría desactualizada la
 * primera vez que cambie el panel, pesaría más que toda la página y no se
 * podría leer en un lector de pantalla. Dibujada con los mismos tokens que el
 * producto real, cambia con él.
 *
 * Los datos son de ejemplo y se nota que lo son: barras grises donde iría
 * texto. Lo único escrito son las horas, que es lo que la página quiere que se
 * entienda.
 *
 * `aria-hidden` porque es una ilustración: lo que dice está en el texto de al
 * lado, y un lector de pantalla no tiene por qué recorrer una agenda falsa.
 */
export function MaquetaPanel() {
  return (
    <div
      aria-hidden="true"
      className="overflow-hidden rounded-2xl border border-borde bg-superficie shadow-[0_1px_2px_rgba(25,21,64,0.04),0_24px_60px_-24px_rgba(25,21,64,0.25)]"
    >
      <div className="flex">
        {/* La barra lateral */}
        <div className="hidden w-36 shrink-0 border-r border-borde p-3 sm:block">
          <div className="px-2 py-1.5 text-sm tracking-tight">
            <span className="text-tinta-suave">app </span>
            <span className="font-black">salud</span>
          </div>
          <div className="mt-4 grid gap-1">
            {["Hoy", "Agenda", "Horarios", "Perfil"].map((s, i) => (
              <div
                key={s}
                className={`rounded-lg px-2 py-1.5 text-xs font-semibold ${
                  i === 0 ? "bg-accent text-accion" : "text-tinta-suave"
                }`}
              >
                {s}
              </div>
            ))}
          </div>
        </div>

        <div className="min-w-0 flex-1 p-5">
          <Barra ancho="w-28" alto="h-4" />

          <div className="mt-4 grid grid-cols-2 gap-3">
            {[
              { etiqueta: "Turnos esta semana", valor: "12" },
              { etiqueta: "Ocupación", valor: "68%" },
            ].map((k) => (
              <div key={k.etiqueta} className="rounded-lg border border-borde p-3">
                <p className="text-[10px] font-semibold uppercase tracking-wide text-tinta-suave">
                  {k.etiqueta}
                </p>
                <p className="mt-0.5 text-xl font-black tabular-nums tracking-tight">
                  {k.valor}
                </p>
              </div>
            ))}
          </div>

          <div className="mt-4 grid gap-2">
            {["09:00", "09:50", "10:40"].map((hora) => (
              <div
                key={hora}
                className="flex items-center gap-3 rounded-lg border border-borde p-3"
              >
                <span className="font-horas text-lg font-black tabular-nums tracking-tight text-libre">
                  {hora}
                </span>
                <Barra ancho="w-20" />
                <Barra ancho="w-14" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * El perfil como lo ve un paciente, dibujado. Mismo criterio que la maqueta del
 * panel: los horarios se leen, el resto son barras.
 */
export function MaquetaPerfil() {
  return (
    <div
      aria-hidden="true"
      className="overflow-hidden rounded-2xl border border-borde bg-superficie shadow-[0_1px_2px_rgba(25,21,64,0.04),0_24px_60px_-24px_rgba(25,21,64,0.25)]"
    >
      <div className="p-5">
        <div className="flex items-center gap-3">
          <span className="size-10 shrink-0 rounded-full bg-accent" />
          <div className="grid gap-1.5">
            <Barra ancho="w-28" alto="h-3" />
            <Barra ancho="w-20" />
          </div>
        </div>
        <div className="mt-4 grid gap-1.5">
          <Barra ancho="w-full" />
          <Barra ancho="w-4/5" />
        </div>
      </div>

      <div className="flex items-center gap-4 border-y border-borde bg-fondo px-5 py-3">
        <Barra ancho="w-16" />
        <Barra ancho="w-20" />
        <span className="ml-auto text-xs font-semibold tabular-nums">$12.000</span>
      </div>

      <div className="p-5">
        <Barra ancho="w-24" />
        <div className="mt-3 flex flex-wrap gap-2">
          {["15:30", "16:30", "17:30"].map((hora) => (
            <span
              key={hora}
              className="rounded-lg border border-borde px-3 py-1.5 font-horas text-base font-black tabular-nums tracking-tight text-libre"
            >
              {hora}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

/**
 * La semana del profesional, dibujada.
 *
 * Es la tercera maqueta y no una repetición de la del panel: cada sección de la
 * landing muestra una parte distinta del producto. Repetir el mismo dibujo dos
 * veces era decir dos veces lo mismo.
 */
export function MaquetaSemana() {
  // Bloques de ejemplo: hora de inicio y alto relativo. Un día vacío también es
  // información —"no atiendo los martes"— así que hay dos.
  const semana: { dia: string; bloques: { desde: string; alto: string }[] }[] = [
    { dia: "Lun", bloques: [{ desde: "09:00", alto: "h-16" }, { desde: "15:00", alto: "h-12" }] },
    { dia: "Mar", bloques: [] },
    { dia: "Mié", bloques: [{ desde: "14:00", alto: "h-20" }] },
    { dia: "Jue", bloques: [{ desde: "09:00", alto: "h-12" }] },
    { dia: "Vie", bloques: [{ desde: "09:00", alto: "h-16" }, { desde: "16:00", alto: "h-8" }] },
  ];

  return (
    <div
      aria-hidden="true"
      className="overflow-hidden rounded-2xl border border-borde bg-superficie p-5 shadow-[0_1px_2px_rgba(25,21,64,0.04),0_24px_60px_-24px_rgba(25,21,64,0.25)]"
    >
      <div className="grid grid-cols-5 gap-2">
        {semana.map(({ dia, bloques }) => (
          <div key={dia}>
            <p className="mb-2 text-center text-[10px] font-semibold uppercase tracking-wide text-tinta-suave">
              {dia}
            </p>
            <div className="grid gap-2">
              {bloques.length === 0 ? (
                <span className="block h-16 rounded-lg border border-dashed border-borde" />
              ) : (
                bloques.map((b) => (
                  <span
                    key={b.desde}
                    className={`flex items-start justify-center rounded-lg bg-accent pt-1.5 text-[10px] font-bold tabular-nums text-accion ${b.alto}`}
                  >
                    {b.desde}
                  </span>
                ))
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function Barra({ ancho, alto = "h-2" }: { ancho: string; alto?: string }) {
  return <span className={`block rounded-full bg-muted ${ancho} ${alto}`} />;
}
