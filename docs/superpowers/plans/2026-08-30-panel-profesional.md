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
**Base:** rama `refactor-gian`, commit `3382026`

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
  tira-de-dias.tsx                 la semana con punto en los días con turnos
  fila-turno-profesional.tsx       un turno visto desde el lado del profesional
  editor-semana.tsx                los siete días con sus bloques
  chips.tsx                        obras sociales y modalidades

lib/
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

> El nombre del paciente sale de `Turno.pacienteId`, y la API **no devuelve el
> nombre**. Para esta etapa se muestra el motivo y el horario, y donde iría el
> nombre va "Paciente". Traerlo necesita expandir la respuesta de
> `GET /profesionales/{id}/turnos` con el nombre del paciente, que es back y no
> entra acá. **Queda como el primer hallazgo de esta etapa.**

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
