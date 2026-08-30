# Plan de implementación — Panel del profesional

> **Para agentes:** SUB-SKILL REQUERIDA: usar `superpowers:subagent-driven-development`
> (recomendado) o `superpowers:executing-plans` para implementar tarea por tarea.
> Los pasos usan checkbox (`- [ ]`) para el seguimiento.

**Objetivo:** que un profesional entre, cargue su semana, vea su agenda y edite
su perfil — y que un paciente pueda encontrarlo por eso.

**Arquitectura:** una sola app Next con tres grupos de rutas y layouts
distintos: `(publico)`, `(paciente)` y `(panel)`. Comparten tokens, cliente de
API y sesión. El login es uno solo y redirige según `perfilProfesionalId`.

**Stack:** el de la etapa anterior. **Cero dependencias nuevas.**

**Spec:** [2026-08-30-panel-profesional-design.md](../specs/2026-08-30-panel-profesional-design.md)
**Base:** rama `refactor-gian`, commit `3231b2a`

---

## Restricciones globales

- **Español en todo lo que se escribe.** Inglés solo para las palabras del
  framework (`page`, `layout`) y de la plataforma (`fetch`, `useState`).
- **Cero dependencias nuevas.** Ninguna librería de calendario, de gráficos ni
  de formularios.
- **Sin gestor de estado global.** El único contexto que entra es el del panel,
  que lleva un objeto y existe para no pedir `/usuarios/yo` en cada pantalla.
- **Los tipos salen de `lib/contrato.ts`.** Ninguno se escribe a mano.
- **El acento (`--color-accion`) solo en la acción principal de cada pantalla.**
- **El color nunca es la única señal.**
- **4.5:1 de contraste como piso.** Medido, no estimado.
- **El motivo de consulta aparece solo en `/panel` y `/panel/agenda`.** Nunca en
  una página pública ni en un `<title>`.
- **Nada de datos inventados.** Sin ingresos, sin ausentismo, sin fotos.
- Comandos desde `apps/web/` salvo que se indique otra cosa.

---

## Pre-flight

| Hallazgo | Consecuencia |
|---|---|
| **Los grupos de rutas no cambian URLs.** `app/(publico)/page.tsx` sirve `/` igual que `app/page.tsx`. | El E2E de la etapa anterior **tiene que seguir pasando sin tocarlo**. Es la verificación de que la reorganización salió bien. |
| **`/usuarios/yo` devuelve `id`, `email`, `nombre`, `apellido`, `creadoEn` y `perfilProfesionalId`.** No trae especialidad. | El saludo de `/panel` necesita **dos** llamadas: `/usuarios/yo` y después `GET /profesionales/{id}`. Se hacen una vez en el layout del panel, no en cada pantalla. |
| **Cinco `page.tsx` se mueven**, ninguno cambia de URL. | Task 1 es mecánica y verificable. |
| **No hay endpoint que simule un cambio de horario.** | Ver la decisión de abajo. |
| **`armarDias` y `<Calendario>` ya existen**, del refactor de la agenda del paciente. Agrupan por `inicio.slice(0, 10)`, que es lo mismo que necesita `/panel/agenda` con turnos. | No se escribe `tira-de-dias.tsx`. Se generaliza `armarDias` a cualquier cosa con `inicio` (Task 4, Paso 1). |
| **`PUT /horarios` devuelve `ListaHorarios`**, que trae `turnosCancelados` además de los horarios. `POST /bloqueos` devuelve lo mismo en su forma. | El número real está disponible después de guardar sin una llamada extra. |
| **`HorarioSemanal.desde`/`hasta` son horas de reloj (`"09:00"`), no instantes.** `diaSemana` es `"lunes" … "domingo"`, en español y sin acento. | El editor de semana trabaja con strings `HH:MM`. No entra ninguna conversión de zona. |

### La decisión que el spec dejaba abierta

El spec pide avisar **antes de guardar** cuántos turnos quedarían huérfanos. La
API no tiene un modo simulación: `PUT /horarios` te dice cuántos canceló
**después** de cancelarlos.

Se evaluaron tres caminos:

| | Por qué no / por qué sí |
|---|---|
| Endpoint de simulación en el back | Correcto, pero es trabajo de dominio y el spec dice que esta etapa no toca la API. |
| Aviso genérico sin número | Gratis y honesto, pero el spec promete el número y "algunos turnos se pueden cancelar" no ayuda a decidir. |
| **Calcularlo en el front** | Es duplicar ~15 líneas de `huecosDelBloque`. **Se elige, con dos mitigaciones.** |

**Las dos mitigaciones importan más que el cálculo:**

1. **Se listan los turnos, no se cuentan.** El diálogo muestra *cuáles* se
   cancelarían —día, hora y paciente— así que aunque el cálculo estuviera mal,
   la persona ve los turnos concretos y puede juzgar.
2. **Después de guardar se informa el número real** que devolvió la API. Si no
   coincide con lo previsto, se ve.

Queda anotado con `ponytail:` en el código, nombrando el arreglo: un
`PUT /horarios?simular=true` que devuelva los turnos afectados sin escribir.

---

## Estructura de archivos

```
app/
  layout.tsx                       raíz: fuentes y tokens. Sin header.
  (publico)/
    layout.tsx                     header de marketing + footer
    page.tsx                       /  landing: hero = buscador
    profesionales/page.tsx         landing de captación
    perfiles/[slug]/page.tsx       ← se mueve
    perfiles/[slug]/reservar/      ← se mueve
  (paciente)/
    layout.tsx                     header simple
    turnos/page.tsx                ← se mueve
  (panel)/
    layout.tsx                     trae usuario y perfil; tabs / barra lateral
    panel/page.tsx                 hoy
    panel/agenda/page.tsx
    panel/horarios/page.tsx
    panel/perfil/page.tsx
  entrar/page.tsx                  ← se queda en la raíz: es de los dos lados

componentes/
  encabezado.tsx                   ← el de hoy, pasa a ser el del grupo público
  navegacion-panel.tsx             tabs abajo en móvil, barra lateral arriba
  nudge.tsx                        qué le falta al perfil para recibir turnos
  calendario.tsx                   ← ya existe: la tira de días y lo del día
  fila-turno-profesional.tsx       un turno visto desde el lado del profesional
  editor-semana.tsx                los siete días con sus bloques
  chips.tsx                        obras sociales y modalidades

lib/
  dias.ts                          ← ya existe: armarDias, se generaliza
  panel.ts                         contexto del panel: usePanel()
  semana.ts                        qué turnos quedan fuera de un horario nuevo
  ocupacion.ts                     el KPI
```

---

## Task 1: grupos de rutas

Mecánica y verificable: ninguna URL cambia, así que el E2E existente es la
prueba.

**Archivos:**
- Crear: `app/(publico)/layout.tsx`, `app/(paciente)/layout.tsx`
- Mover: los cinco `page.tsx` a sus grupos
- Modificar: `app/layout.tsx`

- [ ] **Paso 1: mover los archivos**

```bash
cd apps/web/app
mkdir -p "(publico)/perfiles" "(paciente)"
git mv page.tsx "(publico)/page.tsx"
git mv perfiles "(publico)/perfiles"
git mv turnos "(paciente)/turnos"
```

`entrar/` se queda donde está: es de los dos lados y no pertenece a ningún
grupo.

- [ ] **Paso 2: sacar el header del layout raíz**

En `app/layout.tsx`, quitar `<Encabezado />` y su import. El raíz queda solo con
las fuentes, los tokens y `<html lang="es">`: cada grupo trae su propia
navegación, y eso es lo que hace que se sientan dos productos.

- [ ] **Paso 3: el layout del grupo público**

`app/(publico)/layout.tsx`:

```tsx
import { Encabezado } from "@/componentes/encabezado";

export default function LayoutPublico({ children }: LayoutProps<"/">) {
  return (
    <>
      <Encabezado />
      {children}
      <PieDePagina />
    </>
  );
}
```

El `<PieDePagina />` lleva el link a `/profesionales`, que es la landing de
captación: alguien que llega buscando turno y resulta ser profesional tiene que
encontrar su camino.

- [ ] **Paso 4: el layout del grupo paciente**

`app/(paciente)/layout.tsx`: el mismo `<Encabezado />`, sin footer. Es una
pantalla de gestión, no de descubrimiento.

- [ ] **Paso 5: verificar que nada se movió de URL**

```bash
pnpm build
pnpm e2e
```

Esperado: build en verde con las mismas rutas de antes, y **los 3 E2E pasando
sin haber tocado el archivo de tests**. Si alguno falla, una URL cambió.

- [ ] **Paso 6: commit**

```bash
git add -A
git commit -m "refactor(web): separar las rutas en grupos publico y paciente"
```

---

## Task 2: el login que redirige y el contexto del panel

**Archivos:**
- Crear: `lib/panel.ts`, `app/(panel)/layout.tsx`,
  `componentes/navegacion-panel.tsx`, `app/(panel)/panel/page.tsx` (mínima)
- Modificar: `app/entrar/page.tsx`

**Interfaces:**
- Produce:
  - `usePanel(): { usuario: UsuarioActual; perfil: Profesional }`
  - `<NavegacionPanel />`

- [ ] **Paso 1: escribir el test que falla**

`lib/destino.test.ts` gana el caso nuevo:

```ts
import { destinoDespuesDeEntrar } from "./destino";

describe("destinoDespuesDeEntrar", () => {
  it("respeta el volver por encima de todo", () => {
    // Alguien que venía de reservar tiene que volver a reservar, tenga o no
    // perfil profesional.
    expect(destinoDespuesDeEntrar("/perfiles/x/reservar", "abc-123")).toBe(
      "/perfiles/x/reservar",
    );
  });

  it("manda al panel si tiene perfil profesional", () => {
    expect(destinoDespuesDeEntrar(null, "abc-123")).toBe("/panel");
  });

  it("manda a mis turnos si no tiene", () => {
    expect(destinoDespuesDeEntrar(null, null)).toBe("/turnos");
  });

  it("sigue rechazando un volver externo", () => {
    expect(destinoDespuesDeEntrar("https://sitio-falso.com", "abc-123")).toBe("/panel");
  });
});
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
pnpm vitest run lib/destino.test.ts
```

Esperado: FAIL — `destinoDespuesDeEntrar` no existe.

- [ ] **Paso 3: implementar**

En `lib/destino.ts`, junto a `destinoSeguro`:

```ts
/**
 * A dónde va alguien después de entrar.
 *
 * El `volver` gana siempre: si venía de reservar un turno, vuelve a reservar
 * aunque además sea profesional. La redirección por perfil es el default, no
 * una regla.
 *
 * Que el destino salga de `perfilProfesionalId` y no de un campo de rol es
 * consecuencia directa de la sección 3.1 de la spec de autenticación: no hay
 * enum `Rol`, así que el destino se deriva del estado real en vez de un dato
 * que se puede desincronizar.
 */
export function destinoDespuesDeEntrar(
  volver: string | null,
  perfilProfesionalId: string | null,
): string {
  const seguro = destinoSeguro(volver);
  if (seguro !== "/turnos") return seguro; // vino un volver válido
  return perfilProfesionalId ? "/panel" : "/turnos";
}
```

- [ ] **Paso 4: usarlo en `/entrar`**

Después del `POST /api/v1/sesiones`, pedir `/api/v1/usuarios/yo` y redirigir con
`destinoDespuesDeEntrar(parametros.get("volver"), yo.perfilProfesionalId)`.

- [ ] **Paso 5: el contexto del panel**

`lib/panel.ts`. Existe por una razón concreta: **las cuatro pantallas necesitan
el `perfilProfesionalId` y el perfil, y pedirlos en cada una son ocho llamadas
por navegación.** Es un objeto y un provider, no un gestor de estado.

```tsx
"use client";

import { createContext, useContext } from "react";
import type { Profesional, UsuarioActual } from "./api";

type Panel = { usuario: UsuarioActual; perfil: Profesional };

const Contexto = createContext<Panel | null>(null);

export const ProveedorPanel = Contexto.Provider;

export function usePanel(): Panel {
  const valor = useContext(Contexto);
  if (!valor) {
    // Si esto explota, una pantalla del panel quedó fuera de (panel)/layout.
    throw new Error("usePanel() solo funciona dentro del layout del panel");
  }
  return valor;
}
```

- [ ] **Paso 6: el layout del panel**

`app/(panel)/layout.tsx`, cliente. Trae `/usuarios/yo` y, con el id, el perfil.
Tres estados:

| Estado | Qué hace |
|---|---|
| 401 | `router.replace("/entrar?volver=/panel")` |
| Sin `perfilProfesionalId` | `router.replace("/panel/perfil")` — es el modo de alta |
| Con perfil | Renderiza `<ProveedorPanel>` con `<NavegacionPanel />` |

**`/panel/perfil` tiene que quedar accesible sin perfil**, porque es donde se
crea. Se resuelve dejando pasar esa ruta con `usePathname()`.

- [ ] **Paso 7: la navegación**

`componentes/navegacion-panel.tsx`: Hoy · Agenda · Horarios · Perfil.

```tsx
// Tabs abajo en móvil, barra lateral en escritorio: son dos usos reales.
// Mirar la agenda del día se hace en el teléfono entre pacientes; cargar la
// semana entera se hace sentado.
//
// Una sola lista de rutas, dos presentaciones. Duplicarla es cómo se agrega un
// link en un lado y no en el otro.
```

`aria-current="page"` en el activo: sin eso, un lector de pantalla no dice en
qué sección estás.

- [ ] **Paso 8: verificar**

```bash
pnpm build
```

Y a mano, con la API corriendo: entrar con `martin.gonzalez@ejemplo.com` /
`desarrollo123` tiene que llevar a `/panel`; entrar con una cuenta de paciente,
a `/turnos`.

- [ ] **Paso 9: commit**

```bash
git add -A
git commit -m "feat(web): login que redirige segun el perfil y layout del panel"
```

---

## Task 3: `/panel` — hoy

**Archivos:**
- Modificar: `app/(panel)/panel/page.tsx`
- Crear: `lib/ocupacion.ts`, `lib/ocupacion.test.ts`, `componentes/nudge.tsx`,
  `componentes/fila-turno-profesional.tsx`

**Interfaces:**
- Produce:
  - `calcularOcupacion(turnos: Turno[], huecosLibres: number): { tomados: number; total: number; porcentaje: number }`
  - `<Nudge perfil={Profesional} tieneHorarios={boolean} />`
  - `<FilaTurnoProfesional turno={Turno} />`

- [ ] **Paso 1: escribir el test de ocupación**

```ts
import { describe, expect, it } from "vitest";
import { calcularOcupacion } from "./ocupacion";

const activo = (inicio: string) => ({ inicio, estado: "reservado" }) as never;
const cancelado = (inicio: string) => ({ inicio, estado: "cancelado" }) as never;

describe("calcularOcupacion", () => {
  it("cuenta los tomados sobre el total ofrecido", () => {
    // 3 tomados + 7 libres = 10 ofrecidos.
    const r = calcularOcupacion([activo("a"), activo("b"), activo("c")], 7);
    expect(r).toEqual({ tomados: 3, total: 10, porcentaje: 30 });
  });

  it("no cuenta los cancelados: su hueco volvió a estar libre", () => {
    const r = calcularOcupacion([activo("a"), cancelado("b")], 9);
    expect(r).toEqual({ tomados: 1, total: 10, porcentaje: 10 });
  });

  it("sin agenda cargada no divide por cero", () => {
    // Un profesional recién registrado. 0/0 sería NaN en la pantalla.
    expect(calcularOcupacion([], 0)).toEqual({ tomados: 0, total: 0, porcentaje: 0 });
  });

  it("redondea a entero", () => {
    expect(calcularOcupacion([activo("a")], 2).porcentaje).toBe(33);
  });
});
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
pnpm vitest run lib/ocupacion.test.ts
```

- [ ] **Paso 3: implementar**

```ts
import type { Turno } from "./api";

/**
 * Cuántos de los turnos que el profesional ofrece están tomados.
 *
 * Es el KPI que reemplaza a los dos del prototipo —"cobrado hoy" y
 * "completados"— que necesitan pagos y `atendido`/`ausente`, y no existen. Este
 * sale de datos que sí tenemos y dice algo que un profesional independiente
 * mira de verdad: si le sobra agenda o le falta.
 *
 * `huecosLibres` son los que quedan sin reservar. El total ofrecido es la suma:
 * un turno tomado ya no aparece como hueco.
 */
export function calcularOcupacion(turnos: Turno[], huecosLibres: number) {
  const tomados = turnos.filter((t) => t.estado === "reservado").length;
  const total = tomados + huecosLibres;
  return {
    tomados,
    total,
    // Sin agenda cargada, 0/0 pondría NaN% en la pantalla.
    porcentaje: total === 0 ? 0 : Math.round((tomados / total) * 100),
  };
}
```

- [ ] **Paso 4: el nudge**

`componentes/nudge.tsx`. Tres estados, y **el tercero no lleva link a propósito**:

```tsx
// Un profesional sin horarios cargados no aparece con disponibilidad en la
// búsqueda: lo descubrimos cuando el seed dejaba a los cuatro sin agenda y el
// listado mostraba cuatro tarjetas vacías. Este banner cierra el círculo entre
// registrarse y recibir turnos.
//
// El de verificación es informativo y sin botón: hoy nada mueve ese estado
// —REFEPS es su propia etapa— y un botón que no hace nada es peor que ninguno.
```

| Estado | Texto | Acción |
|---|---|---|
| Sin horarios | Todavía no cargaste tus horarios, así que nadie puede reservarte un turno. | **Configurar** → `/panel/horarios` |
| Sin bio | Tu perfil no tiene descripción. Es lo primero que lee un paciente. | **Completar** → `/panel/perfil` |
| `verificacion: pendiente` | Tu matrícula está pendiente de verificación. | sin link |

- [ ] **Paso 5: la pantalla**

Saludo con nombre y especialidad, el nudge, los dos KPIs y los turnos de hoy.
Los que ya pasaron, en gris.

Cada turno muestra hora, nombre del paciente, motivo y modalidad. **El motivo se
muestra acá y en `/panel/agenda`, y en ningún otro lado.**

> El nombre viene en el propio turno: `GET /profesionales/{id}/turnos` devuelve
> `ListaTurnosDeProfesional`, y cada elemento es un `TurnoConPaciente` con
> `paciente: { id, nombre, apellido }`. **No trae el email, a propósito.** Esto
> se resolvió en la etapa 6a, que existió justamente porque este plan lo
> detectó como hallazgo mientras se escribía.

- [ ] **Paso 6: correr y verificar**

```bash
pnpm vitest run && pnpm build
```

Y a mano: entrar como Martín, ver los turnos de hoy y los dos KPIs.

- [ ] **Paso 7: commit**

```bash
git add -A
git commit -m "feat(web): pantalla de hoy con ocupacion y el nudge de perfil"
```

---

## Task 4: `/panel/agenda`

La tira de días ya existe. Esta tarea la generaliza para que sirva con turnos y
construye la pantalla encima.

**Archivos:**
- Modificar: `lib/dias.ts`, `componentes/calendario.tsx`, `lib/dias.test.ts`
- Crear: `app/(panel)/panel/agenda/page.tsx`, `componentes/dialogo-bloqueo.tsx`
- Usa: `componentes/fila-turno-profesional.tsx` (viene de la Task 3)

**Interfaces:**
- Consume: `usePanel()` (Task 2), `<FilaTurnoProfesional>` (Task 3)
- Produce:
  - `armarDias<T extends { inicio: string }>(items: T[], cantidad?: number, hoy?: Date): Dia<T>[]`
  - `<Calendario<T> dias diaElegido onDia>{(items) => …}</Calendario>`, ahora genérico

- [ ] **Paso 1: escribir el test que falla**

En `lib/dias.test.ts`, agregar al final:

```ts
// La agenda del profesional agrupa turnos, no huecos. Es la misma operación
// —partir por día— sobre otra entidad, y duplicarla es cómo se arreglan bugs
// de zona horaria en un lado y no en el otro.
it("agrupa cualquier cosa que tenga inicio", () => {
  const turnos = [
    { inicio: "2026-08-31T09:00:00-03:00", motivo: "control" },
    { inicio: "2026-09-01T09:00:00-03:00", motivo: "primera vez" },
  ];

  const dias = armarDias(turnos, 2, LUNES);

  expect(dias[0].items[0].motivo).toBe("control");
  expect(dias[1].items[0].motivo).toBe("primera vez");
});
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
pnpm vitest run lib/dias.test.ts
```

Esperado: FAIL de TypeScript — `armarDias` toma `Hueco[]` y el objeto no lo es.

- [ ] **Paso 3: generalizar**

En `lib/dias.ts`, la firma y el tipo. El campo se renombra de `huecos` a
`items`, y `Dia` se vuelve genérico:

```ts
export type Dia<T = Hueco> = {
  fecha: string;
  etiqueta: string;
  numero: string;
  largo: string;
  /** Lo que cae ese día: huecos en el perfil público, turnos en el panel. */
  items: T[];
};

export function armarDias<T extends { inicio: string }>(
  items: T[],
  cantidad = 14,
  hoy = new Date(),
): Dia<T>[] {
```

El cuerpo no cambia salvo el nombre del campo. `primerDiaConHuecos` pasa a
`primerDiaConItems<T>(dias: Dia<T>[]): string`.

Renombrar los tres usos: `componentes/calendario.tsx`,
`app/(publico)/perfiles/[slug]/page.tsx` y `componentes/reservar.tsx`.

`<Calendario>` también se vuelve genérico, y lo de abajo de la tira pasa a ser
un render prop: un hueco se dibuja como un chip de hora, un turno como una fila
con paciente y motivo.

```tsx
export function Calendario<T extends { inicio: string }>({
  dias,
  diaElegido,
  hrefDelDia,
  onDia,
  children,
}: {
  dias: Dia<T>[];
  diaElegido: string;
  hrefDelDia?: (fecha: string) => string;
  onDia?: (fecha: string) => void;
  /** Qué se dibuja debajo de la tira con lo que cae ese día. */
  children: (items: T[]) => React.ReactNode;
}) {
```

Los chips de hora y el mensaje de día vacío se mudan al `children` de cada call
site. `hrefDelHueco` y `onHueco` desaparecen: el call site ya decide qué dibuja.

- [ ] **Paso 4: correr los tests**

```bash
pnpm vitest run && pnpm exec playwright test
```

Esperado: los 28 unit tests y los 4 E2E en verde. **Los E2E son la verificación
de que generalizar no rompió la pantalla del paciente**, que es la que ya
funciona. En particular el de "sin JavaScript": si el render prop se convirtió
en algo que necesita cliente, ese test lo agarra.

- [ ] **Paso 5: commit del refactor, solo**

Va separado de la pantalla nueva: si algo se rompió del lado del paciente, tiene
que poder revertirse sin perder el panel.

```bash
git add -A
git commit -m "refactor(web): armarDias y Calendario sirven para turnos tambien"
```

- [ ] **Paso 6: la pantalla**

`app/(panel)/panel/agenda/page.tsx`, cliente. Trae
`GET /api/v1/profesionales/{id}/turnos?desde=…&hasta=…` con catorce días y lo
pasa por `armarDias`.

```tsx
<Calendario dias={dias} diaElegido={dia} onDia={setDia}>
  {(turnos) =>
    turnos.length === 0 ? (
      <p className="py-2 text-sm text-tinta-suave">No tenés turnos este día.</p>
    ) : (
      <ul className="grid gap-2">
        {turnos.map((t) => (
          <FilaTurnoProfesional key={t.id} turno={t} />
        ))}
      </ul>
    )
  }
</Calendario>
```

**Los cancelados van tachados y dicen quién canceló.** Es el dato que la etapa 4
guardó y que nada leía todavía:

```tsx
// canceladoPor es un id de Usuario. Si coincide con pacienteId canceló el
// paciente; si no, fue el propio profesional desde otro dispositivo. Es la
// diferencia entre "se me cayó un paciente" y "yo lo cancelé", y sin decirlo el
// profesional no sabe cuál de las dos pasó.
const loCancelo = turno.canceladoPor === turno.pacienteId ? "el paciente" : "vos";
```

- [ ] **Paso 7: el botón de bloquear**

Un botón de acción sobre el día que se está mirando. Abre un diálogo con desde,
hasta y motivo, y hace `POST /api/v1/profesionales/{id}/bloqueos`.

La respuesta trae `turnosCancelados`. **Se informa siempre, incluso en cero**,
porque "0 turnos cancelados" es la confirmación de que no rompiste nada:

> Bloqueo creado. Se cancelaron **2 turnos** que caían adentro.

`<dialog>` nativo, no una librería: es la restricción de cero dependencias
nuevas, y `showModal()` ya trae foco atrapado, cierre con Escape y fondo inerte.

El borrado de un bloqueo va en la misma pantalla, con
`DELETE /api/v1/profesionales/{id}/bloqueos/{bloqueoId}`.

- [ ] **Paso 8: verificar**

```bash
pnpm build
```

Y a mano, con la API corriendo: entrar como Martín, ver la agenda, cambiar de
día, crear un bloqueo sobre un día con turnos y confirmar que el número que
informa coincide con los turnos que desaparecen de la lista.

- [ ] **Paso 9: commit**

```bash
git add -A
git commit -m "feat(web): agenda del profesional con bloqueos"
```

---

## Task 5: `/panel/horarios`

La pantalla donde el profesional se sienta a trabajar, y la única de la etapa
que puede destruirle algo a un paciente.

**Archivos:**
- Crear: `lib/semana.ts`, `lib/semana.test.ts`,
  `app/(panel)/panel/horarios/page.tsx`, `componentes/editor-semana.tsx`

**Interfaces:**
- Consume: `usePanel()`
- Produce:
  - `turnosHuerfanos(turnos: TurnoConPaciente[], horarios: HorarioSemanal[]): TurnoConPaciente[]`
  - `<EditorSemana horarios onCambiar />`

- [ ] **Paso 1: escribir el test que falla**

`lib/semana.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import type { HorarioSemanal } from "./api";
import { turnosHuerfanos } from "./semana";

const bloque = (
  diaSemana: HorarioSemanal["diaSemana"],
  desde: string,
  hasta: string,
): HorarioSemanal => ({ diaSemana, desde, hasta, duracionMin: 50, modalidad: "presencial" });

// 2026-08-31 es lunes; 2026-09-01, martes.
const turno = (inicio: string, fin: string) =>
  ({ id: inicio, inicio, fin, estado: "reservado" }) as never;

describe("turnosHuerfanos", () => {
  it("un turno que entra en el horario nuevo no es huérfano", () => {
    const t = turno("2026-08-31T09:00:00-03:00", "2026-08-31T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "13:00")])).toEqual([]);
  });

  it("un turno que quedó fuera del bloque acortado sí lo es", () => {
    // Atendía hasta las 13; el bloque nuevo termina a las 10.
    const t = turno("2026-08-31T12:00:00-03:00", "2026-08-31T12:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "10:00")])).toHaveLength(1);
  });

  it("un turno de un día que se borró entero es huérfano", () => {
    const t = turno("2026-09-01T09:00:00-03:00", "2026-09-01T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "13:00")])).toHaveLength(1);
  });

  // El turno termina 09:50 y el bloque a las 09:30: entra el inicio y no el
  // final. La API lo cancela, así que la pantalla también tiene que verlo.
  it("un turno que empieza adentro pero termina afuera es huérfano", () => {
    const t = turno("2026-08-31T09:00:00-03:00", "2026-08-31T09:50:00-03:00");
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "09:30")])).toHaveLength(1);
  });

  it("un turno ya cancelado no se cuenta dos veces", () => {
    const t = {
      ...turno("2026-08-31T12:00:00-03:00", "2026-08-31T12:50:00-03:00"),
      estado: "cancelado",
    } as never;
    expect(turnosHuerfanos([t], [bloque("lunes", "09:00", "10:00")])).toEqual([]);
  });
});
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
pnpm vitest run lib/semana.test.ts
```

Esperado: FAIL — `turnosHuerfanos` no existe.

- [ ] **Paso 3: implementar**

`lib/semana.ts`:

```ts
import type { HorarioSemanal, TurnoConPaciente } from "./api";

const ZONA = "America/Argentina/Buenos_Aires";

const DIAS: HorarioSemanal["diaSemana"][] = [
  "domingo", "lunes", "martes", "miercoles", "jueves", "viernes", "sabado",
];

/**
 * Los turnos que el horario nuevo dejaría afuera, y que `PUT /horarios`
 * cancelaría al guardar.
 *
 * ponytail: esto duplica en el front la regla que el back ya tiene, porque la
 * API no ofrece modo simulación — `PUT /horarios` informa cuántos canceló
 * DESPUÉS de cancelarlos. El arreglo real es un `PUT /horarios?simular=true`
 * que devuelva los afectados sin escribir; se hace cuando alguien reporte que
 * el número previsto y el real no coinciden.
 *
 * Las dos mitigaciones importan más que el cálculo: la pantalla LISTA los
 * turnos en vez de contarlos, así que aunque esto se equivoque la persona ve
 * turnos concretos y juzga; y después de guardar se informa el número real que
 * devolvió la API.
 */
export function turnosHuerfanos(
  turnos: TurnoConPaciente[],
  horarios: HorarioSemanal[],
): TurnoConPaciente[] {
  return turnos.filter((turno) => {
    // Un turno ya cancelado no se puede cancelar de nuevo.
    if (turno.estado !== "reservado") return false;

    const dia = DIAS[new Date(turno.inicio).getDay()];

    // El turno tiene que entrar ENTERO en algún bloque de ese día: el back
    // cancela el que se pasa del final aunque haya empezado adentro.
    return !horarios.some(
      (h) =>
        h.diaSemana === dia &&
        reloj(turno.inicio) >= h.desde &&
        reloj(turno.fin) <= h.hasta,
    );
  });
}

/** `"09:50"`, para comparar contra las horas de reloj del horario. */
function reloj(iso: string): string {
  return new Intl.DateTimeFormat("es-AR", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: ZONA,
  }).format(new Date(iso));
}
```

> **Ojo con `getDay()`:** devuelve el día en la zona del proceso, no en la de
> Argentina. Esta pantalla corre en el browser del profesional, que está en
> Argentina, así que coincide. **Si algún día corre en un servidor en UTC, un
> turno de las 21:00 se cuenta como del día siguiente.** Se deja anotado al lado
> de la línea; el arreglo es derivar el día de `turno.inicio.slice(0, 10)`, que
> ya viene con el offset aplicado.

- [ ] **Paso 4: correr y verificar que pasa**

```bash
pnpm vitest run lib/semana.test.ts
```

- [ ] **Paso 5: mutación**

Sacar `&& reloj(turno.fin) <= h.hasta` de la condición y correr de nuevo. Tiene
que fallar el caso "empieza adentro pero termina afuera". Restaurar y volver a
verde.

Es el caso que más fácil se escribe mal, y el que más caro sale: cancelar un
turno sin haberlo avisado.

- [ ] **Paso 6: el editor**

`componentes/editor-semana.tsx`: siete días, cada uno con sus bloques. Cada
bloque tiene desde, hasta, duración y modalidad, y un botón de borrar.
`<input type="time">` y `<select>` nativos — la restricción de cero
dependencias, y el picker del sistema operativo es mejor que cualquier
reimplementación.

```
Móvil                          Escritorio
┌──────────────────┐           ┌────┬────┬────┬────┬────┬────┬────┐
│ Lunes         +  │           │Lun │Mar │Mié │Jue │Vie │Sáb │Dom │
│ 09:00–13:00 50m  │           ├────┼────┼────┼────┼────┼────┼────┤
│ 15:00–19:00 50m  │           │9-13│    │9-13│    │9-13│    │    │
├──────────────────┤           │15-19    │15-19    │    │    │    │
│ Martes        +  │           │ +  │ +  │ +  │ +  │ +  │ +  │ +  │
```

Una sola lista de días, dos presentaciones con CSS. Duplicar la lista es cómo se
agrega un día en un lado y no en el otro.

Cada botón de agregar necesita nombre accesible propio: `aria-label="Agregar
bloque a Sábado"`. Siete botones que dicen todos "+" son siete botones que un
lector de pantalla no puede distinguir — y además es el locator del E2E.

- [ ] **Paso 7: la confirmación antes de guardar**

Al tocar Guardar, **antes** del `PUT`:

1. `GET /api/v1/profesionales/{id}/turnos` desde hoy.
2. `turnosHuerfanos(turnos, horariosNuevos)`.
3. Si hay alguno, `<dialog>` que los **lista** —día, hora y paciente— con el
   número en el encabezado:

> Con este cambio se cancelan **3 turnos** ya reservados. Los pacientes lo van a
> ver en la app.
> - lunes 7, 12:00 — Lucía Fernández
> - lunes 14, 12:00 — Diego Paz
> - lunes 21, 12:00 — Lucía Fernández
>
> [Cancelar] [Guardar igual]

El botón destructivo **no** lleva `--color-accion`: el acento es de la acción
principal, y acá la acción principal es no romper nada.

- [ ] **Paso 8: informar el número real**

Después del `PUT`, `ListaHorarios.turnosCancelados` trae lo que pasó de verdad.
Se muestra siempre, en un `role="alert"`:

> Horarios guardados. Se cancelaron **3 turnos**.

Si no coincide con lo previsto, se ve. Es la segunda mitigación del cálculo
duplicado.

- [ ] **Paso 9: verificar de punta a punta**

Con la API corriendo, y esto es el criterio de aceptación 6:

1. Entrar como Martín y cargar un bloque en un día vacío.
2. Abrir `/perfiles/martin-gonzalez` **en otra pestaña** y confirmar que los
   huecos nuevos aparecen.
3. Volver, acortar el bloque de un día con turnos, y confirmar que el diálogo
   los lista antes de guardar.

- [ ] **Paso 10: commit**

```bash
git add -A
git commit -m "feat(web): editor de la semana que avisa que turnos cancelaria"
```

---

## Task 6: `/panel/perfil`, en sus dos modos

**Archivos:**
- Crear: `app/(panel)/panel/perfil/page.tsx`, `componentes/chips.tsx`
- Modificar: `lib/formato.ts`, `lib/formato.test.ts`

**Interfaces:**
- Consume: `usePanel()`
- Produce:
  - `enCentavos(pesos: string): number | null`
  - `<Chips opciones seleccionadas onCambiar />`

- [ ] **Paso 1: el test del precio**

Es el campo donde más fácil se cuela un error de dos órdenes de magnitud, así
que la conversión se prueba antes de escribirla. En `lib/formato.test.ts`:

```ts
describe("enCentavos", () => {
  it("pesos enteros", () => expect(enCentavos("12000")).toBe(1200000));
  it("con separador de miles", () => expect(enCentavos("12.000")).toBe(1200000));
  it("con centavos", () => expect(enCentavos("12000,50")).toBe(1200050));
  it("con el símbolo", () => expect(enCentavos("$12.000")).toBe(1200000));
  it("vacío es null, no cero", () => expect(enCentavos("")).toBe(null));
  it("texto es null", () => expect(enCentavos("gratis")).toBe(null));
  it("negativo es null", () => expect(enCentavos("-100")).toBe(null));

  // El campo muestra el valor formateado mientras se escribe, así que lo que
  // sale de formatearPrecio vuelve a entrar por enCentavos en cada tecla. Si el
  // viaje de ida y vuelta pierde precisión, el precio se degrada solo.
  it("da la vuelta completa", () => {
    expect(enCentavos(formatearPrecio(1200000))).toBe(1200000);
    expect(enCentavos(formatearPrecio(1200050))).toBe(1200050);
  });
});
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
pnpm vitest run lib/formato.test.ts
```

- [ ] **Paso 3: implementar**

```ts
/**
 * Pesos escritos a mano a centavos. La inversa de `formatearPrecio`.
 *
 * Devuelve `null` y no `0` para lo que no es número: un precio vacío es "no
 * completó el campo", y guardar $0 como si lo hubiera elegido es peor que
 * pedirle que lo complete.
 */
export function enCentavos(pesos: string): number | null {
  // Se sacan el símbolo, los espacios y los puntos de miles; la coma decimal
  // pasa a punto. Es el formato es-AR, que es el que ve el profesional.
  const limpio = pesos.replace(/[$\s.]/g, "").replace(",", ".");
  if (limpio === "") return null;

  const numero = Number(limpio);
  if (!Number.isFinite(numero) || numero < 0) return null;
  return Math.round(numero * 100);
}
```

- [ ] **Paso 4: correr y verificar que pasa**

```bash
pnpm vitest run lib/formato.test.ts
```

- [ ] **Paso 5: los chips**

`componentes/chips.tsx`, para obras sociales y modalidades. Botones que se
prenden y apagan.

```tsx
// aria-pressed y no solo una clase: sin él un lector de pantalla lee "OSDE,
// botón" tanto si está elegida como si no. Y el estado elegido lleva además un
// check, porque el color nunca es la única señal.
```

Las obras sociales son texto libre en el contrato (`obrasSociales?: string[]`),
así que las opciones salen de un array local con las del mercado argentino y hay
un campo para agregar una que no esté. **No se inventa un catálogo con ids** que
la API no tiene.

- [ ] **Paso 6: la pantalla, modo edición**

Bio, precio, zona, modalidades, obras sociales y anticipación mínima.
`PUT /api/v1/profesionales/{id}`. Con "Ver mi perfil público" →
`/perfiles/{slug}`.

Matrícula y especialidad se muestran **de solo lectura**, con una línea que
explica por qué: cambiarlas resetea la verificación.

- [ ] **Paso 7: la pantalla, modo alta**

Si `usuario.perfilProfesionalId` es `null`, el mismo formulario con matrícula y
especialidad **editables**, y el botón dice "Crear mi perfil" en vez de "Guardar
cambios". Hace `POST /api/v1/profesionales` y después va a `/panel`.

```tsx
// Dos modos de un formulario, no dos pantallas: los campos son los mismos y el
// alta solo agrega dos. Duplicarlo es cómo se agrega un campo en uno y no en el
// otro.
const alta = usuario.perfilProfesionalId === null;
```

Los 422 se muestran debajo de su campo con `aria-invalid` y `aria-describedby`,
igual que en `reservar.tsx`. El 409 de matrícula en uso es su propio mensaje:
*"Ya hay un perfil con esa matrícula."*

- [ ] **Paso 8: verificar**

```bash
pnpm vitest run && pnpm build
```

Y a mano, que es el criterio de aceptación 9: cambiar la bio y el precio como
Martín, abrir `/perfiles/martin-gonzalez` y ver los dos cambios. Después, con
una cuenta de paciente, entrar a `/panel/perfil` y crear un perfil.

- [ ] **Paso 9: commit**

```bash
git add -A
git commit -m "feat(web): editar y crear el perfil profesional"
```

---

## Task 7: las dos landings

**Archivos:**
- Modificar: `app/(publico)/page.tsx`, `componentes/encabezado.tsx`
- Crear: `app/(publico)/profesionales/page.tsx`, `componentes/pie.tsx`

- [ ] **Paso 1: `/` — el hero es el buscador**

Se agrega **debajo** de lo que ya está, sin tocar el buscador ni el listado: la
propuesta de valor y cómo funciona, en tres pasos.

```tsx
// El hero es el buscador y no un texto de marketing. `/` es la página con más
// autoridad de SEO del sitio, y gastarla en un hero sin contenido indexable es
// tirarla: en un marketplace de salud la búsqueda orgánica es el canal de
// adquisición. Es lo que hacen Doctoralia y Zocdoc.
```

- [ ] **Paso 2: `/profesionales` — la landing de captación**

Server Component. Su propio mensaje, y el CTA va a
`/entrar?volver=/panel/perfil`: alguien que ya tiene cuenta entra y cae en el
alta; alguien que no, se registra desde ahí.

**Sin números inventados.** Nada de "más de 500 profesionales" ni "10.000
pacientes": no hay de dónde sacarlos, y en salud un número falso es lo que rompe
la confianza que la página vino a construir.

- [ ] **Paso 3: el pie**

`componentes/pie.tsx`, solo en el grupo público: link a `/profesionales` y el
aviso de que la plataforma no reemplaza una consulta de urgencia.

- [ ] **Paso 4: verificar sin JavaScript**

Es el criterio de aceptación 10:

```bash
curl -s http://localhost:3000/ | grep -c "Cómo funciona"
curl -s http://localhost:3000/profesionales | grep -c "Crear mi perfil"
```

Esperado: 1 en los dos. Si da 0, la sección se está renderizando en el cliente.

- [ ] **Paso 5: commit**

```bash
git add -A
git commit -m "feat(web): landing publica y landing de captacion"
```

---

## Task 8: el E2E del circuito profesional y el cierre

La etapa no está terminada hasta que el circuito completo esté cubierto por un
test que corra sin nadie mirando.

**Archivos:**
- Crear: `e2e/panel.spec.ts`

- [ ] **Paso 1: el test del circuito**

`e2e/panel.spec.ts`, contra la API real igual que el otro:

```ts
// El circuito que justifica la etapa entera: un profesional carga su semana y
// esos huecos aparecen en su perfil público. Si esto pasa, las dos mitades del
// marketplace están conectadas.
test("un profesional carga su semana y aparece en su perfil público", async ({ page }) => {
  await page.goto("/entrar");
  await page.getByLabel("Email").fill("martin.gonzalez@ejemplo.com");
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByRole("button", { name: "Entrar" }).click();

  // El login redirige solo al panel: es el criterio de aceptación 2.
  await expect(page).toHaveURL("/panel");

  await page.getByRole("link", { name: "Horarios" }).click();
  // Un día que el seed deja vacío, para no pelear con turnos existentes.
  await page.getByRole("button", { name: "Agregar bloque a Sábado" }).click();
  await page.getByLabel("Desde").last().fill("10:00");
  await page.getByLabel("Hasta").last().fill("12:00");
  await page.getByRole("button", { name: "Guardar" }).click();
  await expect(page.getByRole("alert")).toContainText("Horarios guardados");

  // Y del otro lado del marketplace.
  await page.goto(`/perfiles/martin-gonzalez?dia=${proximoSabado()}`);
  await expect(page.getByRole("link", { name: "10:00" })).toBeVisible();
});
```

`proximoSabado()` se calcula en el test: sin el `?dia=` la página abre en el
primer día con huecos, que no es el sábado.

- [ ] **Paso 2: el test de la redirección de paciente**

```ts
test("una cuenta sin perfil profesional va a sus turnos", async ({ page }) => {
  // La cuenta se crea en el test, para no depender del seed.
  // …registrarse desde una reserva o desde /entrar…
  await expect(page).toHaveURL("/turnos");
});
```

- [ ] **Paso 3: correr todo**

```bash
pnpm exec playwright test
```

Esperado: los 4 de la etapa anterior **sin tocarlos** y los 2 nuevos.

- [ ] **Paso 4: el recorrido con teclado**

Criterio de aceptación 11, y se hace a mano porque es lo único que lo verifica
de verdad: abrir `/panel` y recorrer las cuatro pantallas solo con Tab, Shift+Tab
y Enter. Cada elemento enfocado tiene que verse.

Lo que se busca: un `<div onClick>` que Tab no alcanza, un chip sin
`aria-pressed`, un `<dialog>` que no atrapa el foco.

- [ ] **Paso 5: contraste**

**Medido, no estimado** — es exactamente el error que ya se cometió una vez en
esta app, cuando `--color-libre` salió con 2.26:1 después de haber escrito
"contraste medido, no estimado" en el mismo documento.

Cada color nuevo de esta etapa contra su fondo, con un contador de contraste.
Piso 4.5:1.

- [ ] **Paso 6: ponytail-audit**

```
/ponytail-audit apps/web
```

Buscando lo que la etapa anterior encontró: componentes sin importadores,
helpers con un solo llamador, duplicación entre las cuatro pantallas del panel.

- [ ] **Paso 7: verificación final**

```bash
pnpm lint && pnpm test && pnpm build && pnpm exec playwright test
```

- [ ] **Paso 8: commit**

```bash
git add -A
git commit -m "test(web): e2e del circuito profesional"
```

---

## Auto-revisión del plan

**Cobertura de la spec:**

| Sección de la spec | Tarea |
|---|---|
| 2. Tres grupos de rutas | Task 1 |
| 3. Login único | Task 2 |
| 4. Layout móvil / escritorio | Task 2 (navegación), Task 5 (grilla de la semana) |
| 5. `/panel` | Task 3 |
| 5. `/panel/agenda` | Task 4 |
| 5. `/panel/horarios` | Task 5 |
| 5. `/panel/perfil`, dos modos | Task 6 |
| 6. El nudge | Task 3 |
| 7. Las landings | Task 7 |
| 8. Fuera de alcance | Nada de eso aparece en ninguna tarea |

**Los doce criterios de aceptación:**

| # | Dónde se verifica |
|---|---|
| 1 | Task 8, Paso 7 |
| 2 | Task 2 (Paso 8) y Task 8 (Pasos 1 y 2) |
| 3 | Task 3, Paso 6 |
| 4 | Task 3, Pasos 1–3 (test) y Paso 6 (a mano) |
| 5 | Task 3, Paso 4 |
| 6 | Task 5, Paso 9 |
| 7 | Task 5, Pasos 1–5 y 7 |
| 8 | Task 4, Pasos 7 y 8 |
| 9 | Task 6, Paso 8 |
| 10 | Task 7, Paso 4 |
| 11 | Task 8, Paso 4 |
| 12 | Tasks 3 y 4 lo muestran; ninguna otra pantalla lo recibe |

**Consistencia de tipos:** `armarDias` cambia de firma en Task 4 y sus tres
llamadores se renombran en el mismo paso; `Dia.huecos` pasa a `Dia.items` ahí
mismo, y `primerDiaConHuecos` a `primerDiaConItems`. `usePanel()` se define en
Task 2 y lo consumen las tareas 3, 4, 5 y 6 con la misma forma.
`TurnoConPaciente` es lo que devuelve `GET /profesionales/{id}/turnos`, y es lo
que consumen `turnosHuerfanos` y `<FilaTurnoProfesional>`.

**Alcance:** ocho tareas, cada una con su commit y su verificación. Las tareas 4
y 5 son las grandes; la 4 se parte en dos commits a propósito, para que el
refactor del lado del paciente pueda revertirse solo.
