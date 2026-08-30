# Diseño — `apps/web`, lado paciente

- **Fecha:** 2026-08-30
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de plan de implementación
- **Depende de:** las cuatro etapas de `apps/api`. Ninguna entidad nueva.

---

## 1. Contexto

`apps/api` tiene 22 operaciones, un contrato de mil líneas y cuatro etapas de
dominio. **Nunca tuvo un consumidor.**

CORS lo demostró: no existió hasta ayer, y si alguien hubiera intentado un
`fetch` desde un browser una sola vez en cuatro meses, lo habría descubierto en
cinco minutos. Cada etapa que se agregue sin consumidor es otro contrato
diseñado a ciegas.

Esta etapa **no agrega dominio.** Hace dos cosas: valida el contrato contra
alguien del otro lado, y convierte 74 commits en algo que se puede mostrar.

### Por qué Next y no seguir con Vite

En un marketplace de salud el SEO de los perfiles no es un extra: **es el canal
de adquisición.** "psicólogo palermo osde" es literalmente cómo un paciente
encuentra a alguien, y `/api/v1/perfiles/{slug}` con slugs legibles existe para
eso. Renderizado en el cliente, Google ve una página vacía.

Es la única razón. Si el SEO no importara, Vite alcanzaría.

---

## 2. Stack

| Pieza | Elección |
|---|---|
| Framework | Next 15, App Router, TypeScript |
| Paquetes | pnpm con workspaces |
| Estilos | Tailwind |
| Componentes | shadcn/ui sobre Radix |
| Tipos de la API | `openapi-typescript`, generados de `api/openapi.yaml` |
| E2E | Playwright |

**shadcn no es una dependencia**, es un generador: copia los componentes al
repositorio y el proyecto es dueño del código. Debajo hay Radix, que resuelve
foco, teclado y roles. En una app de salud eso no es opcional.

### Lo que deliberadamente no entra

- **Gestor de estado global.** Lo público lo trae el servidor; lo privado es un
  `fetch` y el estado local de una pantalla. Redux o Zustand resuelven un
  problema que todavía no existe.
- **Librería de formularios.** El formulario más complejo de esta etapa tiene
  cinco campos.
- **Cliente de API generado.** Se generan los **tipos**, no el cliente. El
  cliente es una función `pedir()` que hace `fetch` con
  `credentials: "include"` y traduce un `Problema` de RFC 7807 a un error
  tipado. Un cliente generado son cientos de líneas para envolver `fetch`.

Los tipos sí se generan, y es la única concesión a la generación de código: el
contrato ya es la fuente de verdad y está validado en CI, así que copiar sus
formas a mano es duplicación que deriva. Generándolos, un cambio en
`openapi.yaml` rompe la compilación del front, que es exactamente cuando
conviene enterarse.

---

## 3. La línea del híbrido

Next puede llamar a la API desde el servidor o dejar que la llame el browser.
Esta etapa hace las dos cosas, y **la línea cae exactamente donde cae la
autenticación**:

| Se renderiza en el servidor | Sale directo del browser |
|---|---|
| búsqueda, perfil público, huecos | reservar, registro, login, mis turnos |
| **no necesitan sesión** | la cookie viaja sola |

Esa coincidencia es lo que hace barato al híbrido: **nunca hay que reenviar una
cookie a mano**, porque lo que se renderiza en el servidor es justo lo que es
público. Las acciones privadas usan CORS, que ya está.

Un BFF completo —todo por Next— daría el mismo SEO pero obligaría a reenviar la
cookie en cada llamada y a escribir un espejo de cada endpoint. Más trabajo por
un beneficio que el híbrido ya da.

---

## 4. Cinco pantallas

```
/                            búsqueda con filtros; cada resultado muestra sus próximos huecos
/perfiles/[slug]             perfil público — la que indexa Google
/perfiles/[slug]/reservar    elegir hueco y confirmar
/entrar                      login
/turnos                      mis turnos
```

Las dos primeras son Server Components. Las tres últimas son de cliente, y
`/turnos` sin sesión redirige a `/entrar?volver=/turnos`.

**Los filtros de `/`** son los que la API ya soporta: especialidad, zona y
búsqueda por nombre. Nada más — inventar un filtro que el back no tiene sería
prometer algo que no existe.

### El primer hallazgo de tener un consumidor

Cada resultado de `/` muestra sus próximos huecos, porque eso es la tesis de
Grilla. Pero **el listado no devuelve huecos**: `GET /profesionales` da los
perfiles, y los huecos salen de `GET /profesionales/{id}/huecos`, uno por
profesional. Con veinte resultados son veintiuna llamadas.

Es exactamente el tipo de cosa que aparece la primera vez que alguien consume un
contrato diseñado sin consumidor, y vale la pena registrarlo como tal.

**Decisión para esta etapa:** el Server Component pide los huecos de los
resultados de la página en paralelo. Contra una API en memoria y con veinte
resultados por página, es rápido, y preserva la tesis del diseño.

> **ponytail:** N+1 contra la API. El arreglo real es un campo
> `proximoDisponible` en el listado, calculado del lado del back, y es una
> etapa chica. Se hace cuando el listado se sienta lento o cuando exista
> PostgreSQL, lo que llegue primero.

**`/` y `/perfiles/[slug]` tienen que verse completas sin JavaScript.** Es el
criterio que hace que la decisión de usar Next signifique algo, y se verifica
apagando JS en el browser, no leyendo el código.

El prototipo tiene además `Consulta.jsx` y `PostConsulta.jsx`. Quedan afuera:
son la videollamada y lo que sigue, y no hay ni video ni pagos.

---

## 5. Reservar sin cuenta

El paciente busca, entra al perfil y elige el horario **sin sesión**. Recién al
confirmar aparece el formulario: email, contraseña, nombre, apellido y motivo
opcional.

Al enviar son **dos llamadas**: `POST /api/v1/usuarios` —que ya deja la sesión
abierta— y después `POST /api/v1/profesionales/{id}/turnos`.

Entre las dos hay tiempo, y de ahí salen dos ramas de error. **No son
complejidad agregada: el back devuelve esos códigos y el front tiene que hacer
algo con ellos.** La decisión es solamente si ese algo dice la verdad.

| Qué pasa | Qué ve el paciente |
|---|---|
| **409 al reservar** — alguien tomó el hueco mientras se registraba | Vuelve a la lista de horarios con *"Ese horario se tomó recién. Elegí otro."* Ya quedó logueado, así que el segundo intento es un click. |
| **409 al registrar** — el email ya existe | *"Ya tenés una cuenta con ese email"* y un link a `/entrar?volver=<esta página>`. |

Nada más. Se evaluó y **se descartó** un login en línea dentro de la pantalla de
reserva que conservara el hueco elegido: son unas ochenta líneas y una segunda
máquina de estados para ahorrar dos clicks en un caso que un link resuelve.

También se evaluó **pedir la cuenta antes de mostrar los horarios**. Elimina las
dos ramas de raíz, pero Google llegaría al muro de login — y el SEO era la razón
de haber elegido Next.

> La primera rama, además, hoy casi no puede ocurrir: necesita dos pacientes,
> mismo profesional, mismo horario, dentro de los diez segundos que lleva
> completar un formulario. Se escribe igual porque la alternativa no es
> ahorrarse código, es mentirle al usuario.

---

## 6. Sistema visual

Estructura **Grilla** —el horario es el objeto principal de la interfaz, no un
detalle del profesional— con el ritmo de **Calma**: aire, una decisión por
pantalla, nada de densidad que obligue a comparar.

```
--fondo    #FAFAFC        --libre    #00C08B     estado
--tinta    #191540        --accion   #4B3FE4     acción
--borde    #E6E5EF        --apagado  #C8C6D6
```

Dos colores para dos trabajos distintos. En una pantalla que es casi toda
horarios, si "libre" y "reservar" comparten color no hay jerarquía posible: todo
compite.

**Archivo 900** para las horas, con `font-variant-numeric: tabular-nums`.
**Public Sans** para el resto. La hora es el elemento de mayor tamaño de cada
fila, porque es lo que la persona vino a buscar.

### Cuatro reglas, no sugerencias

1. **El acento se gana, no se reparte.** El violeta aparece en la acción
   principal de cada pantalla y en ningún otro lado.
2. **Ningún gris puro.** Todos los neutros corridos hacia el índigo. Es lo que
   hace que la pantalla se lea como un sistema y no como componentes sueltos.
3. **El color nunca es la única señal.** "Ocupado" lleva tachado y borde
   punteado además del gris. Ocho de cada cien varones no distinguen rojo de
   verde.
4. **Contraste medido, no estimado.** 4.5:1 de piso para texto. El violeta
   saturado sobre blanco no llega, y por eso se usa como fondo de botón con
   texto blanco, nunca como texto.

### Lo que no se muestra nunca

El **motivo de consulta** no aparece en listados, previsualizaciones ni títulos
de página. La **especialidad** no sale en ninguna notificación. Es Ley 25.326,
no una preferencia de diseño, y es la misma restricción que va a condicionar las
plantillas cuando existan notificaciones.

---

## 7. Testing

Un E2E con Playwright **contra la API de Go real, sin mocks** — el equivalente
de front a la convención que el back ya tiene escrita.

Un solo camino, el completo: buscar → entrar al perfil → elegir un hueco →
registrarse → reservar → verlo en `/turnos`.

Es exactamente el tipo de test que habría encontrado lo de CORS. No se agregan
tests de componente: en un front que recién arranca, la mayoría prueban JSX, y
el JSX cambia todos los días.

---

## 8. Fuera de alcance

| Pendiente | Cuándo se activa |
|---|---|
| Lado profesional (cargar horarios, ver mi agenda) | Etapa siguiente. Sin profesionales cargando su semana no hay nada que buscar. |
| Videoconsulta y post-consulta | Cuando exista video. Hoy `Consulta.jsx` no tiene nada detrás. |
| Pagos | Con la evidencia de cuál es el modelo. Ver la spec de Turno. |
| Tema oscuro | Cuando alguien lo pida. La paleta se define en tokens, así que agregarlo es un segundo bloque de valores. |
| PWA, notificaciones push | Con las notificaciones del back |
| i18n | Solo si hay usuarios fuera de Argentina |
| Borrar `legacy/prototype` | Cuando estas cinco pantallas lo reemplacen. Hasta entonces es la única descripción que existe de los flujos. |

---

## 9. Criterios de aceptación

1. `pnpm build` y `pnpm lint` en verde, y el E2E de Playwright pasa contra la
   API real levantada con el seed.
2. `/` y `/perfiles/[slug]` **se ven completas con JavaScript apagado**, con el
   nombre, la especialidad, el precio y los horarios en el HTML.
3. Un visitante sin cuenta puede buscar, entrar a un perfil y elegir un horario
   sin que nada le pida registrarse.
4. Al confirmar, se crea la cuenta y el turno, y el turno aparece en `/turnos`.
5. Si el hueco se tomó en el medio, la pantalla vuelve a los horarios con el
   mensaje correcto y la sesión ya iniciada.
6. Si el email ya existe, el mensaje lo dice y ofrece entrar.
7. El horario reservado desaparece de la lista de huecos del profesional.
8. Toda la interfaz se recorre con teclado, con foco visible en cada control.
9. Ningún texto de la interfaz baja de 4.5:1 de contraste.
10. El motivo de consulta no aparece en ningún listado ni en ningún `<title>`.
