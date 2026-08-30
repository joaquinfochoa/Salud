# Diseño — Turno

- **Fecha:** 2026-08-29
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de plan de implementación
- **Depende de:** [CRUD de Profesional](2026-08-21-professional-crud-go-design.md), [Disponibilidad](2026-08-22-disponibilidad-design.md), [Identidad y autenticación](2026-08-29-autenticacion-design.md)

---

## 1. Contexto

Las tres etapas anteriores construyeron un directorio, una agenda y una
identidad. Ninguna cierra el circuito: hoy se puede ver *cuándo* atiende
alguien, pero no *tomar* ese turno.

`Turno` es la pieza que lo cierra, y es la primera entidad que toca el negocio:
una reserva **es** la transacción del producto.

Llega ahora y no antes porque la spec de disponibilidad lo dejó escrito —*"antes
de `Turno`, autenticación"*— y esa deuda está pagada. La consecuencia es que
esta etapa no necesita inventar al paciente: **el paciente es un `Usuario`**.

### Sin pagos, a propósito

`Turno` empuja la pregunta de quién le paga a quién, y esa pregunta no está
contestada: comisión por turno y suscripción mensual son modelos de datos
distintos —comisión necesita split, conciliación y estado del cobro;
suscripción necesita plan, ciclo y morosidad—. Modelar el pago antes de saber
cuál es apostar a ciegas en la entidad más cara de cambiar.

Reservar y cancelar no dependen de esa respuesta. Se construyen ahora; el cobro
se enchufa cuando haya evidencia.

---

## 2. La entidad

| Campo | Tipo | Notas |
|---|---|---|
| `ID` | `uuid.UUID` | |
| `ProfesionalID` | `uuid.UUID` | |
| `PacienteID` | `uuid.UUID` | referencia a `Usuario` |
| `Inicio`, `Fin` | `time.Time` | los dos salen del hueco, no del cliente |
| `Modalidad` | `Modalidad` | copia del bloque horario al reservar |
| `Estado` | `EstadoTurno` | `reservado` \| `cancelado` |
| `Motivo` | `string` | opcional, lo escribe el paciente |
| `CreadoEn` | `time.Time` | |
| `CanceladoEn` | `*time.Time` | |
| `CanceladoPor` | `*uuid.UUID` | |

**No hay entidad `Paciente`.** `Usuario` ya es la identidad y agregar una
entidad que solo tendría un `UsuarioID` y nada más es ceremonia. Aparece el día
que un paciente necesite datos que un usuario no tiene —obra social, número de
afiliado, fecha de nacimiento— y entonces se enganchará al `Usuario` con una
referencia, sin tocar `Turno`.

**`CanceladoPor` es un `UsuarioID`, no un enum `paciente|profesional`.** Guarda
más información por el mismo precio, y quién es cada uno se deduce comparando
contra `PacienteID`. El enum obligaría además a inventar un tercer valor el día
que cancele un administrador.

**`Modalidad` se copia al reservar, no se lee del horario al mostrar.** Si el
profesional convierte los martes de presencial a telemedicina, el turno ya
tomado tiene que seguir diciendo lo que se pactó. Un turno es un acuerdo entre
dos personas en un momento; leerlo del horario actual lo reescribiría hacia
atrás.

### Estados

Solo `reservado` y `cancelado`. Un turno pasado es un turno cuya fecha ya pasó:
eso lo dice el reloj, no un campo.

`atendido` y `ausente` se evaluaron y se difieren. Hoy nada consume esa
información —no hay cobro, ni reputación, ni política de ausencias— y agregarlas
después cuesta una constante en el enum, un método de transición y un endpoint,
sin ninguna migración: `Estado` es una columna de texto que acepta un valor más.
`confirmado` se difiere por otra razón: el estado es media hora, pero solo sirve
con notificaciones, y eso es una etapa entera con proveedor externo.

### Lo que sí se guarda ahora porque después no se recupera

`CanceladoEn` y `CanceladoPor`. Si cancelar solo cambiara el estado, el día que
exista una política de ausencias —*"cancelar con menos de 24 horas cuenta como
ausencia"*— faltaría un dato que no se puede reconstruir. Es el mismo argumento
que sostiene `DadoDeBajaEn` en `Profesional`.

**Lo que deliberadamente no se guarda:** el precio de la consulta al momento de
reservar. Sin pagos no lo lee nadie, y cuando existan, el precio va a venir del
pago y no del turno.

---

## 3. Reservar

```
POST /api/v1/profesionales/{id}/turnos     requiere sesión
```

El paciente sale de la sesión, nunca del cuerpo. El cuerpo manda **solo**
`inicio` y `motivo`.

**No manda `fin` ni `modalidad`**: los dos salen del hueco. Aceptarlos del
cliente sería dejar que se invente un turno de tres horas en una agenda de 50
minutos.

### La regla, que es una sola

**El `inicio` pedido tiene que coincidir exactamente con un hueco libre
calculado en ese momento.**

Es la decisión que hace barata toda la etapa. Arrastra gratis, sin reimplementar
nada: el horario semanal, los bloqueos, la anticipación mínima, el horizonte, el
profesional inactivo y los turnos ya tomados. Todo eso ya lo decide
`CalculoHuecos`.

Un `inicio` que no cae en un hueco es **422**, y el mensaje nombra el campo.
La alternativa —validar cada regla otra vez en el alta del turno— duplicaría
seis reglas en un segundo lugar que se desincroniza del primero.

### Una segunda regla: el paciente no se solapa consigo mismo

Un paciente no puede tener dos turnos **activos** que se pisen en el tiempo,
aunque sean con profesionales distintos. Devuelve **409**.

Sale casi gratis —para reservar ya hay que leer los turnos del paciente— y
cubre el error honesto, que es reservar dos cosas a la misma hora sin darse
cuenta. Los turnos cancelados no cuentan: cancelar libera al paciente igual que
libera el hueco.

**Lo que esta regla no cubre:** acaparar. Sin pagos no hay ninguna fricción
contra tomar los doce huecos de un profesional en una tarde. Se evaluó un tope
de turnos futuros por profesional y se descartó: el número sería inventado y
rompería un caso real —un tratamiento kinesiológico de diez sesiones son diez
turnos con la misma persona—. Va a la lista de **antes de exponer la API**,
junto al rate limiting del login, que es la herramienta que corresponde.

### La carrera

Dos pacientes pidiendo el mismo hueco al mismo tiempo se resuelve en la
escritura del repositorio, bajo el lock, igual que la matrícula: el chequeo del
servicio es el camino rápido que da el error lindo, y la garantía la da
`Crear`, que rechaza con `ErrHuecoTomado` → **409**.

---

## 4. Cancelar

```
DELETE /api/v1/turnos/{id}     requiere sesión
```

Lo pueden hacer **el paciente o el profesional**. Por eso la ruta no cuelga de
`/profesionales/{id}`: no es una operación del dueño de la agenda, es de
cualquiera de las dos partes. Un tercero recibe **403**.

Sin ventana mínima: se cancela hasta que el turno empieza. Una ventana es una
política de negocio que hoy no tiene consecuencia —no hay nada que cobrar ni
penalizar— y agregarla después es un campo en `Profesional`, al lado de
`AnticipacionMinimaMin`.

Cancelar un turno ya empezado o ya cancelado es **422**. No es idempotente a
propósito: a diferencia del logout, acá repetir la operación probablemente
significa que el cliente cree algo distinto de lo que pasó.

---

## 5. Listar

```
GET /api/v1/turnos                        mis turnos, como paciente
GET /api/v1/profesionales/{id}/turnos     la agenda del profesional
```

Los dos requieren sesión. El primero responde con los turnos del usuario de la
sesión; el segundo, solo si es el dueño del perfil.

**El segundo es privado, a diferencia de `/horarios` y `/bloqueos`.** La
diferencia importa: los huecos libres son información de oferta y son públicos,
pero la agenda *ocupada* dice quién es paciente de quién, y eso es dato de salud
bajo Ley 25.326. Un listado público de turnos filtraría exactamente lo que la
ley protege.

Por la misma razón, `GET /api/v1/turnos` nunca acepta un `pacienteId` como
filtro: el paciente es siempre el de la sesión.

Los dos aceptan `desde` y `hasta` opcionales, con el mismo contrato que
`GET /bloqueos`: sin rango, se devuelven los vigentes y futuros. Reusar la regla
que ya existe evita que el cliente tenga que aprender dos convenciones de fecha
en la misma API.

Ningún listado excluye los cancelados: un turno cancelado es parte de tu
historial y del del profesional. Lo que no aparece es en `GET /huecos`, que es
otra pregunta.

### Códigos de los turnos que no son tuyos

Un `id` que no existe da **404**; uno que existe pero es de otras dos personas
da **403**. La distinción no es un oráculo acá: los ids son UUID v4, así que no
se pueden enumerar, y responder 404 a todo escondería un error real del cliente
detrás de "no existe".

---

## 6. Los huecos restan turnos

`CalculoHuecos` gana el campo `Turnos []Turno`, que su propio comentario ya
anticipa —*"Cuando exista Turno, se le suma un campo con los turnos ya tomados y
se restan igual que los bloqueos. La firma pública no cambia."*—.

Se restan igual que los bloqueos, con una diferencia: **los turnos `cancelado`
no se restan.** Cancelar libera el hueco, que es el punto de cancelar.

---

## 7. El bloqueo que pisa turnos

`CrearBloqueo` pasa a **cancelar** los turnos que solapa, con `CanceladoPor` =
el profesional. Lo mismo `ReemplazarHorarios` cuando saca un bloque que tenía
turnos.

**Solo se cancelan los turnos que todavía no empezaron.** Un bloqueo puede
cubrir fechas pasadas —nada lo prohíbe, y cargar "estuve de licencia la semana
pasada" es un caso real— pero un turno que ya ocurrió no se puede cancelar: ya
pasó. Reescribirlo hacia atrás perdería el registro de una consulta que
efectivamente sucedió. Es la misma regla que la sección 4 le aplica a una
cancelación manual, y por eso vive en un solo lugar del dominio.

Se evaluaron las tres salidas:

| | Por qué no |
|---|---|
| Rechazar el bloqueo con 422 | Irse de viaje con la agenda llena se vuelve una tarea de doce pasos, y el profesional termina no bloqueando nada. |
| Dejar los turnos vivos | El profesional cree que esa semana está libre y se entera el día del turno. |
| **Cancelarlos** | Lo que hacen las plataformas reales, y lo único que deja al profesional y al sistema diciendo lo mismo. |

La respuesta de las dos operaciones informa **cuántos turnos canceló**, para que
el profesional lo sepa en el momento en que lo hace y no después.

### Deuda que esta etapa crea

**El paciente se entera entrando a la app.** Sin notificaciones, cancelar
automáticamente sigue siendo mejor que dejar turnos fantasma —el sistema y el
profesional al menos coinciden— pero no es aceptable con usuarios reales.

Va a la misma lista que el rate limiting del login: **antes de exponer la API**.
No es una regresión, es una deuda nueva y se registra como tal.

---

## 8. Decisiones menores, resueltas por default

Se anotan porque son decisiones, no descuidos.

- **`Motivo` es opcional, máximo 500 caracteres.** Un motivo de consulta es una
  línea, no una historia clínica. `Bio` tiene 2000 porque es un texto de
  marketing; esto no.
- **Un usuario que además es profesional puede reservar consigo mismo.** No
  daña nada y prohibirlo es código para un caso que nadie va a intentar en
  serio.
- **`POST /bloqueos` y `PUT /horarios` cambian su respuesta** para informar
  cuántos turnos cancelaron. Es un cambio de contrato en dos operaciones que
  ya existen, y hay que reflejarlo en `openapi.yaml`.
- **Los listados incluyen los cancelados.** Filtrarlos sería esconderle al
  paciente que le cancelaron un turno, que es justo lo que necesita ver.

---

## 9. Fuera de alcance

| Pendiente | Cuándo se activa |
|---|---|
| Pagos y comisión | Cuando haya evidencia de cuál es el modelo. Ver sección 1. |
| Notificaciones | **Antes de exponer la API.** La deuda de la sección 7 depende de esto. |
| `atendido` / `ausente` | Cuando algo consuma el dato: cobro, reputación o política de ausencias |
| `confirmado` | Con notificaciones, no antes |
| Reprogramar | Hoy es cancelar y reservar de nuevo. Se vuelve su propia operación si hace falta conservar la identidad del turno. |
| Ventana de cancelación | Cuando cancelar tenga una consecuencia. Es un campo en `Profesional`. |
| Turno cargado por el profesional | Cuando alguien lo pida. Necesita crear un `Usuario` en nombre de otro, que es su propio problema. |
| Turnos recurrentes | Sin fecha. Un tratamiento kinesiológico de diez sesiones es el caso real. |
| Entidad `Paciente` | Cuando el paciente necesite datos que un `Usuario` no tiene |
| PostgreSQL | Cuando el modelo deje de moverse |

---

## 10. Criterios de aceptación

1. `make test-race` y `make lint` en verde.
2. Un paciente logueado reserva un hueco de `GET /huecos`, y ese hueco
   desaparece de la siguiente llamada a `GET /huecos`.
3. Reservar un `inicio` que no coincide con ningún hueco da **422** nombrando el
   campo.
4. Dos reservas del mismo hueco: la primera 201, la segunda **409**.
5. Un paciente que ya tiene un turno a las 10:00 con A no puede reservar las
   10:00 con B: **409**.
6. El paciente y el profesional pueden cancelar; un tercero recibe **403**.
7. Un turno cancelado libera el hueco: vuelve a aparecer en `GET /huecos`.
8. `GET /api/v1/turnos` devuelve solo los turnos del usuario de la sesión, y sin
   sesión da **401**.
9. `GET /api/v1/profesionales/{id}/turnos` solo lo responde el dueño; otro
   usuario recibe **403**.
10. Un bloqueo que pisa dos turnos los deja cancelados con `CanceladoPor` = el
   profesional, y la respuesta informa que canceló dos.
11. Los huecos de un profesional inactivo siguen siendo lista vacía, y no se
    puede reservar contra él.
