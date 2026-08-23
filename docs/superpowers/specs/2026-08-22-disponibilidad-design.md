# Diseño — Disponibilidad del profesional

- **Fecha:** 2026-08-22
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de plan de implementación
- **Depende de:** [CRUD de Profesional](2026-08-21-professional-crud-go-design.md)

---

## 1. Contexto

El backend tiene una sola entidad: `Profesional`, un directorio público con
perfil, matrícula, especialidad y precio. Funciona de punta a punta.

Lo que le falta para ser útil es **cuándo se puede consultar a esa persona**. El
prototipo de `legacy/prototype` ya muestra "Próximo disponible: Mañana 09:00" y
una grilla de horarios en el perfil, con datos inventados y nada detrás. Esta
etapa los reemplaza por datos reales.

### Por qué esto y no otra cosa

Se evaluaron cuatro caminos: `Paciente`, `Disponibilidad`, migrar a PostgreSQL, y
`Turno` con autenticación primero.

`Disponibilidad` gana por tres razones:

1. **Completa algo que ya existe.** El perfil público está construido y hoy
   miente sobre los horarios. `Paciente`, en cambio, sería una entidad sin
   ningún consumidor hasta que exista `Turno`.
2. **Sigue siendo pública.** La agenda disponible la ve cualquiera, así que no
   necesita autenticación. Es la última pieza grande que cumple esa condición.
3. **Es el prerrequisito de `Turno`.** No se puede reservar un hueco que el
   sistema no sabe calcular.

### La autenticación se difiere, deliberadamente

Se diseñó una spec de autenticación —sesiones opacas en cookie `HttpOnly`, por
revocabilidad— y se decidió posponerla para construir dominio primero.

El razonamiento, para que quede registrado:

| Qué | Costo de diferir |
|---|---|
| El mecanismo de auth | **Ninguno.** Es autocontenido; cuesta lo mismo hoy que en seis meses. |
| El modelo de datos | **Casi ninguno.** `Usuario` se engancha con una referencia opcional, y no hay datos en producción que migrar. |
| Las firmas del servicio | **Alto, y proporcional al dominio de usuario que se construya en el medio.** |
| El riesgo | **Cero mientras no esté expuesto. Sin techo el día que lo esté.** |

La frontera es esta: **mientras se construyan cosas públicas o de catálogo,
diferir no cuesta.** `Profesional` y `Disponibilidad` son públicas: "listar
profesionales" y "ver los huecos de alguien" significan lo mismo sin importar
quién pregunte.

`Turno` es lo contrario. "Listar turnos" no significa nada hasta saber quién
pregunta —mis turnos, los turnos conmigo, o todos— y eso no es un filtro que se
agrega después: cambia la firma del método, cada handler que lo llama y cada
test.

**Cuando se llegue a `Turno`, autenticación va primero.**

---

## 2. Decisiones tomadas

### Alcance: regla semanal más excepciones

El profesional carga su horario habitual y puede bloquear rangos concretos.

Se descartó la versión mínima (solo la regla semanal) porque una agenda que no
soporta un feriado ni unas vacaciones produce pacientes reservando turnos que no
existen — y el relevamiento dice que la carga administrativa es justamente lo
que espanta a los profesionales.

Se descartó también la versión sin recurrencia (cargar los horarios semana a
semana) por el mismo motivo: es más simple de modelar y mucho más trabajo para
el usuario.

### Los horarios son hora de reloj, no instantes

**Esta es la decisión que más se rompe en sistemas de agenda.**

Cuando un profesional dice "atiendo de 9 a 13", eso significa las 9 de su reloj,
para siempre. No es un instante: es una hora local que se repite. Guardarlo en
UTC significa guardar "12:00 UTC", y esa equivalencia solo vale mientras el país
no cambie de huso.

Argentina no tiene horario de verano desde 2009, pero lo tuvo entre 2007 y 2009
y cada tanto vuelve a proponerse. El día que vuelva, una agenda guardada en UTC
mueve todos los turnos una hora sin que nadie toque nada.

Entonces:

- `HorarioSemanal` guarda día de la semana más hora de reloj.
- `Bloqueo` guarda fechas y horas locales.
- Los **huecos calculados** sí son instantes, y salen en RFC 3339 con su offset
  (`2026-08-25T09:00:00-03:00`).

### La zona horaria es una constante del sistema

`America/Argentina/Buenos_Aires`, fija. No un campo del profesional.

Argentina tiene un huso único y el producto arranca en AMBA, así que un campo
por profesional sería configurabilidad que nadie usa. Queda marcado en el código
con su condición de salida:

```go
// ponytail: zona fija. El producto es Argentina, que tiene un huso único.
// El día que haya profesionales fuera del país, esto pasa a ser un campo de
// Profesional — el resto del modelo ya trata las horas como hora de reloj,
// así que el cambio es local.
```

Que sea constante **no** es lo mismo que guardar UTC. Las horas siguen siendo de
reloj; lo único fijo es a qué reloj se refieren.

### La duración va en el bloque

El prototipo tiene al psicólogo con sesiones de 50 minutos y al kinesiólogo con
60. Pero un mismo profesional puede querer 50 minutos por videollamada y 60
presenciales, porque el presencial incluye que la persona entre y salga.

En el bloque sale gratis; en el profesional habría que mudarla después.

### Los huecos se calculan, no se guardan

No hay tabla de huecos. `HuecosLibres(profesionalID, desde, hasta)` es una
función de las reglas y el rango pedido.

Materializarlos es lo que se hace cuando hace falta reservar uno de forma
atómica —un problema de `Turno`, no de acá— y hacerlo antes obliga a mantener
sincronizada una tabla derivada con las reglas que la generan.

Cuando llegue `Turno`, esta misma función le resta los turnos tomados **sin
cambiar de firma**.

### Anticipación y horizonte: configurables por profesional

Cada profesional define con cuánta anticipación mínima acepta reservas y cuánto
de su agenda expone hacia adelante.

Un psicólogo por videollamada puede aceptar con una hora; un kinesiólogo a
domicilio necesita medio día. Un control odontológico se agenda con más tiempo
que una sesión de terapia.

**El horizonte lleva un tope del sistema encima.** Sin él, alguien configura
3.650 días, un cliente pide el rango entero y el servidor calcula cientos de
miles de huecos por pedido.

---

## 3. Modelo de dominio

### Entidades nuevas

```go
// HorarioSemanal es un bloque del horario habitual. Un profesional que atiende
// mañana y tarde tiene dos filas para el mismo día.
//
// No tiene ID a propósito: la semana se reemplaza entera, así que ningún
// endpoint direcciona un bloque suelto y nada lo referencia. Un ID sería peso
// muerto que además cambiaría en cada guardado, confundiendo a cualquier
// cliente que lo guarde. Un horario es un valor, no una entidad con identidad.
// Bloqueo sí conserva el suyo, porque se borra de a uno.
type HorarioSemanal struct {
    ProfesionalID uuid.UUID
    DiaSemana     DiaSemana
    Desde         HoraDelDia
    Hasta         HoraDelDia
    DuracionMin   int
    Modalidad     Modalidad
}

// Bloqueo es lo que rompe la regla: vacaciones, un feriado, cerrar temprano un
// martes. Guarda fechas y horas locales, no instantes UTC.
type Bloqueo struct {
    ID            uuid.UUID
    ProfesionalID uuid.UUID
    Desde         time.Time
    Hasta         time.Time
    Motivo        string
    CreadoEn      time.Time
}
```

**Se llama `Bloqueo`, no `Excepcion`.** Se consideró un
`Excepcion{Tipo: bloqueo | horario_extra}`, pero un enum con un solo valor es una
abstracción especulativa esperando un segundo caso que quizás nunca llegue.
Bloquear cubre vacaciones, feriados y cerrar temprano. Si algún día un
profesional quiere abrir un sábado suelto, eso es un concepto distinto y merece
su propio nombre.

### Campos nuevos en `Profesional`

```go
AnticipacionMinimaMin int  // minutos. Por defecto 120.
HorizonteDias         int  // por defecto 60. Tope del sistema: 180.
```

**Por qué ahí y no en una entidad aparte:** `PrecioConsulta` y `Modalidades` ya
viven en `Profesional` y tampoco son identidad — son cómo esa persona ejerce.
Estos dos son de la misma familia. Una entidad `ConfiguracionAgenda` para dos
campos es ceremonia.

La consecuencia es que se toca una entidad ya construida, probada y publicada en
el contrato. Ambos campos tienen valor por defecto, así que una petición que no
los mande sigue siendo válida.

### Value objects nuevos

```go
// HoraDelDia es una hora de reloj sin fecha: "las 9 de la mañana".
//
// time.Time no sirve acá porque siempre carga una fecha, y guardar una fecha
// arbitraria para después ignorarla es la clase de convención que alguien
// rompe.
type HoraDelDia struct {
    Minutos int  // desde medianoche. 09:00 es 540.
}

func ParsearHoraDelDia(s string) (HoraDelDia, error)  // "09:00"
func (h HoraDelDia) String() string
```

### Structs de entrada

Como `EntradaProfesional`: tipos primitivos, para que todo el parseo y toda la
validación ocurran dentro del dominio y no repartidos por los handlers.

```go
type EntradaHorarioSemanal struct {
    DiaSemana   string
    Desde       string  // "09:00"
    Hasta       string
    DuracionMin int
    Modalidad   string
}

type EntradaBloqueo struct {
    Desde  time.Time
    Hasta  time.Time
    Motivo string
}
```

```go
// DiaSemana sigue el vocabulario del resto de los enums del proyecto: valores
// en español, sin acentos, como Especialidad y Modalidad.
type DiaSemana string

const (
    DiaLunes     DiaSemana = "lunes"
    DiaMartes    DiaSemana = "martes"
    DiaMiercoles DiaSemana = "miercoles"
    DiaJueves    DiaSemana = "jueves"
    DiaViernes   DiaSemana = "viernes"
    DiaSabado    DiaSemana = "sabado"
    DiaDomingo   DiaSemana = "domingo"
)

func (d DiaSemana) EsValido() bool
func (d DiaSemana) AWeekday() time.Weekday
func DiaSemanaDe(w time.Weekday) DiaSemana
```

El enum propio en vez de `time.Weekday` directo mantiene la frontera: el dominio
habla español y no expone una numeración donde el domingo es cero, que es una
convención de C que nadie recuerda.

### Reglas de validación

| Regla | Por qué |
|---|---|
| `Desde` < `Hasta` en un bloque | Un bloque que termina antes de empezar no es nada |
| Dos bloques del mismo día no se solapan | "Lunes 9-13" y "Lunes 12-15" es ambiguo: casi siempre un error de carga |
| `DuracionMin` entre 10 y 480 | Menos de diez minutos no es una consulta; más de ocho horas tampoco |
| Un bloque tiene que entrar al menos un hueco | "Lunes 9:00 a 9:30" con sesiones de 50 minutos no produce nada, y el profesional debería enterarse al cargarlo y no al no recibir turnos |
| `DiaSemana` y `Modalidad` dentro de su enum | |
| `Desde` < `Hasta` en un bloqueo | |
| `Hasta` de un bloqueo posterior al momento de crearlo | Bloquear el pasado no hace nada. Se evalúa contra el `ahora` que recibe el constructor, no contra el reloj del sistema, para que el test sea determinista. |
| `Motivo` hasta 200 caracteres, puede estar vacío | |
| `AnticipacionMinimaMin` entre 0 y 10.080 (una semana) | |
| `HorizonteDias` entre 1 y 180 | El tope protege el cálculo |

La cuarta es la más valiosa del conjunto: es una validación que le ahorra al
profesional descubrir por silencio que su agenda no genera turnos.

Los errores se acumulan en un `ErrorValidacion`, igual que en `Profesional`.

---

## 4. El cálculo de huecos

Toda la matemática de calendario vive en el dominio, como una **función pura
sobre entradas explícitas**. Sin repositorios, sin reloj del sistema, sin
contexto.

```go
type Hueco struct {
    Inicio    time.Time
    Fin       time.Time
    Modalidad Modalidad
}

type CalculoHuecos struct {
    Horarios              []HorarioSemanal
    Bloqueos              []Bloqueo
    Desde, Hasta          time.Time
    AnticipacionMinimaMin int
    Ahora                 time.Time
}

func (c CalculoHuecos) Generar() []Hueco
```

El servicio carga los datos y arma el struct; el dominio hace la cuenta. Esa
separación es la que permite probar "un bloqueo que empieza exactamente cuando
termina un hueco" sin levantar nada.

### El algoritmo

`Desde` y `Hasta` son días completos y **ambos entran**: pedir del 25 al 27
incluye los tres. El servicio los construye a partir de las fechas de la query,
`Desde` a las 00:00 y `Hasta` al final del día, en la zona del sistema.

Para cada día del rango:

1. Tomar los bloques cuyo `DiaSemana` coincida con ese día.
2. Desde `bloque.Desde`, avanzar de a `DuracionMin` mientras
   `inicio + duración <= bloque.Hasta`.
3. Convertir cada hueco a un instante en la zona del sistema.
4. Descartar los que caen antes de `Ahora` más `AnticipacionMinimaMin`.
5. Descartar los que se solapan con algún bloqueo.

Ordenar por `Inicio`.

Un bloque de 9:00 a 13:00 con sesiones de 50 minutos da cuatro huecos: 9:00,
9:50, 10:40 y 11:30. El de 12:20 terminaría 13:10 y se descarta. Coincide con lo
que el prototipo mostraba a mano.

### Los intervalos son semiabiertos

`[inicio, fin)`. Un hueco está bloqueado si:

```go
hueco.Inicio.Before(bloqueo.Hasta) && hueco.Fin.After(bloqueo.Desde)
```

Con esa definición, un hueco que termina 9:50 y un bloqueo que empieza 9:50 no
se tocan. Es el off-by-one clásico de las agendas y tiene su propio test.

### Un profesional inactivo no tiene huecos

Sin importar lo que digan sus reglas. La verificación vive en el servicio, que
es quien conoce el estado del profesional.

---

## 5. Repositorio

Dos interfaces nuevas, en `internal/repository`, con la misma forma que la de
`Profesional`: `context.Context` primero, implementación en memoria ahora,
PostgreSQL después.

```go
type HorarioSemanal interface {
    ReemplazarDeProfesional(ctx context.Context, profesionalID uuid.UUID, horarios []domain.HorarioSemanal) error
    ListarDeProfesional(ctx context.Context, profesionalID uuid.UUID) ([]domain.HorarioSemanal, error)
}

type Bloqueo interface {
    Crear(ctx context.Context, b domain.Bloqueo) error
    ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Bloqueo, error)
    Eliminar(ctx context.Context, id uuid.UUID) error
    ListarDeProfesional(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Bloqueo, error)
}
```

**`ReemplazarDeProfesional` es una sola operación a propósito.** Borra y escribe
bajo el mismo lock, igual que un `DELETE` más `INSERT` dentro de una transacción
cuando llegue PostgreSQL. Si fueran dos llamadas, entre una y otra el profesional
queda sin horarios y alguien puede leer ese estado.

Las implementaciones en memoria mantienen la disciplina de `Profesional`: mutex,
copias profundas de entrada y de salida, orden total antes de devolver.

---

## 6. Servicio

```go
type Agenda struct {
    profesionales repository.Profesional
    horarios      repository.HorarioSemanal
    bloqueos      repository.Bloqueo
    ahora         func() time.Time
}

func (s *Agenda) ReemplazarHorarios(ctx context.Context, profesionalID uuid.UUID, entradas []domain.EntradaHorarioSemanal) ([]domain.HorarioSemanal, error)
func (s *Agenda) ListarHorarios(ctx context.Context, profesionalID uuid.UUID) ([]domain.HorarioSemanal, error)
func (s *Agenda) CrearBloqueo(ctx context.Context, profesionalID uuid.UUID, entrada domain.EntradaBloqueo) (domain.Bloqueo, error)
func (s *Agenda) ListarBloqueos(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Bloqueo, error)
func (s *Agenda) EliminarBloqueo(ctx context.Context, profesionalID, bloqueoID uuid.UUID) error
func (s *Agenda) HuecosLibres(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Hueco, error)
```

Reglas que necesitan más de una entidad, y por eso viven acá:

1. **Todas las operaciones verifican que el profesional exista.** Cargar
   horarios para un ID inexistente devuelve `ErrNoEncontrado`, no un éxito
   silencioso.
2. **La modalidad de cada bloque tiene que estar entre las que el profesional
   declara.** Cargar un bloque presencial cuando el perfil dice que solo hace
   telemedicina es incoherente, y produce huecos que el paciente ve y no puede
   usar. Es una regla entre dos entidades, por eso no puede vivir en el dominio:
   el bloque solo no sabe qué ofrece su profesional.

   Devuelve `ErrorValidacion` sobre el campo `modalidad`, nombrando cuáles sí
   están disponibles.

3. **`EliminarBloqueo` verifica que el bloqueo sea de ese profesional.** Sin
   auth cualquiera puede llamarlo, pero al menos la ruta y el recurso tienen que
   ser coherentes: borrar el bloqueo de otro profesional desde la ruta de este
   es un 404.
4. **`HuecosLibres` recorta el rango al horizonte del profesional, contado
   desde hoy.** Desde hoy y no desde la fecha pedida: el horizonte es cuánto de
   su agenda el profesional expone hacia adelante, y contarlo desde `desde`
   dejaría que un cliente pidiera septiembre de 2099 y reservara turnos a tres
   años vista.

   Recorta, no rechaza: pedir noventa días a alguien que expone sesenta
   devuelve sesenta y lo informa, igual que `paginacion.limite` en el listado de
   profesionales. Un rango que cae entero más allá del horizonte devuelve una
   lista vacía, no un error. Lo único que sí rechaza es un rango invertido, que
   no es una preferencia sino un error.

   Como el horizonte de cada profesional ya está acotado a 180 días por la
   validación de la sección 3, el recorte alcanza para proteger el cálculo: no
   hace falta un segundo tope en la API.

5. **`ListarBloqueos` sin rango devuelve los vigentes y futuros**, es decir
   aquellos cuyo `Hasta` es posterior a `ahora`. El handler traduce la ausencia
   de parámetros a ese rango antes de llamar al repositorio, así la interfaz del
   repositorio no necesita punteros ni valores centinela.
6. **Un profesional inactivo devuelve una lista vacía**, no un error: el recurso
   existe, simplemente no tiene disponibilidad.

---

## 7. La API

Prefijo `/api/v1`. Todas las rutas cuelgan del profesional, porque una
disponibilidad no existe sin dueño.

| Método | Ruta | Qué hace |
|---|---|---|
| `GET` | `/profesionales/{id}/horarios` | Lista los bloques de la semana |
| `PUT` | `/profesionales/{id}/horarios` | Reemplaza la semana entera |
| `GET` | `/profesionales/{id}/bloqueos` | Lista los bloqueos, con rango opcional |
| `POST` | `/profesionales/{id}/bloqueos` | Crea un bloqueo |
| `DELETE` | `/profesionales/{id}/bloqueos/{bloqueoID}` | Elimina un bloqueo |
| `GET` | `/profesionales/{id}/huecos` | Calcula los huecos libres |

**Por qué el horario se reemplaza entero en vez de tener alta y baja por
bloque:** una semana se edita como una unidad. Nadie agrega un bloque suelto;
configura su semana. Y hacerlo así vuelve trivial la validación de solapamiento
—se valida el conjunto y se guarda— mientras que con altas individuales habría
que validar cada bloque contra los existentes, y un cliente que mueve dos bloques
pasa por un estado intermedio inválido que la API tendría que rechazar aunque el
resultado final sea correcto.

Los bloqueos sí son individuales: se agregan unas vacaciones, después se borran.

### Parámetros

`GET /huecos` exige `desde` y `hasta` como fechas locales (`2026-08-25`). Sin
ellas, 400.

`GET /bloqueos` los acepta como opcionales; sin ellos devuelve los vigentes y
futuros.

### Códigos de estado

| Situación | Código |
|---|---|
| Lectura correcta | `200` |
| Bloqueo creado | `201` + `Location` |
| Horario reemplazado | `200` con la semana resultante |
| Bloqueo eliminado | `204` |
| Parámetro con formato inválido, o rango invertido | `400` |
| JSON válido con datos inválidos | `422` con detalle por campo |
| Profesional o bloqueo inexistente | `404` |

Se mantiene la distinción entre `400` y `422` ya establecida.

### Forma del JSON

Español y camelCase, como el resto.

```json
{
  "horarios": [
    {
      "diaSemana": "lunes",
      "desde": "09:00",
      "hasta": "13:00",
      "duracionMin": 50,
      "modalidad": "telemedicina"
    }
  ]
}
```

```json
{
  "id": "…",
  "desde": "2026-09-10T00:00:00-03:00",
  "hasta": "2026-09-20T23:59:59-03:00",
  "motivo": "Vacaciones",
  "creadoEn": "2026-08-22T14:02:11-03:00"
}
```

```json
{
  "datos": [
    { "inicio": "2026-08-25T09:00:00-03:00", "fin": "2026-08-25T09:50:00-03:00", "modalidad": "telemedicina" },
    { "inicio": "2026-08-25T09:50:00-03:00", "fin": "2026-08-25T10:40:00-03:00", "modalidad": "telemedicina" }
  ],
  "rango": { "desde": "2026-08-25", "hasta": "2026-09-01" }
}
```

El listado de huecos no lleva `paginacion`: el rango ya lo acota y devolver una
página de huecos partiría un día por la mitad, que no le sirve a ninguna
interfaz. `rango` informa el rango efectivamente usado, que puede ser menor al
pedido si el horizonte del profesional lo recortó — mismo principio que
`paginacion.limite` en el listado de profesionales.

Los errores siguen en `application/problem+json`.

### El contrato se escribe primero

`apps/api/api/openapi.yaml` se extiende **antes** que los handlers, como en la
etapa anterior.

---

## 8. Testing

Sin mocks. Los repositorios en memoria son los dobles de test.

| Capa | Contra qué | Qué prueba |
|---|---|---|
| `domain` value objects | Nada | `ParsearHoraDelDia` con formatos válidos e inválidos, forma canónica, `DiaSemana` ida y vuelta contra `time.Weekday` |
| `domain` validación | Nada | Cada regla de la sección 3, solapamiento entre bloques, bloque que no entra ningún hueco |
| `domain` cálculo | Nada, función pura | Conteo y horarios exactos, bloque corto sin huecos, dos bloques el mismo día, bloqueo total, bloqueo parcial, **el borde exacto**, anticipación mínima, límites del rango |
| `repository/memory` | Sí mismo | Que el reemplazo sea atómico, que el filtro por rango incluya los bloqueos que se solapan parcialmente, copias profundas |
| `service` | Los repositorios reales | Profesional inexistente, profesional inactivo, horizonte recortando el rango, bloqueo de otro profesional |
| `handler` | El stack por `httptest` | Códigos de estado, `desde`/`hasta` faltantes o mal formados, rango invertido, y que un rango mayor al horizonte se recorte en vez de fallar |

**El test del borde** es el que más importa: un hueco `[09:00, 09:50)` y un
bloqueo que empieza exactamente `09:50` no se solapan. Una implementación que
use `<=` en vez de `<` lo descarta, y nadie lo nota hasta que un profesional
pierde un turno por día.

**El test del bloque corto** protege la regla de validación: cargar "lunes 9:00
a 9:30" con sesiones de 50 minutos tiene que ser un 422, no una semana guardada
que no produce nada.

---

## 9. Fuera de alcance

| Pendiente | Cuándo se activa |
|---|---|
| Autenticación y autorización | Antes de `Turno`, y antes de exponer la API |
| `Turno` | Después de auth |
| `Paciente` | Con `Turno` |
| Restar turnos tomados al calcular huecos | Cuando exista `Turno`. `CalculoHuecos` recibe un campo más; la firma pública no cambia. |
| Horarios extra puntuales (abrir un sábado suelto) | Cuando alguien lo pida. Es un concepto distinto de `Bloqueo`. |
| Zona horaria por profesional | Si alguna vez hay profesionales fuera de Argentina |
| Feriados nacionales automáticos | Hoy el profesional los bloquea a mano. Automatizarlo necesita una fuente de feriados y decidir si se aplican por defecto. |
| Duración variable dentro de un bloque | Un bloque tiene una sola duración. Un profesional que hace primeras consultas de 80 minutos y seguimientos de 50 necesita dos bloques. |
| PostgreSQL | Cuando el modelo deje de moverse |

---

## 10. Deuda que esta etapa agranda

**La ausencia de autenticación deja de ser solo destructiva y pasa a ser
manipulable.** Hoy cualquiera con la URL puede dar de baja a un profesional. Con
esta etapa, además puede vaciarle la agenda o llenársela de bloqueos.

No es una regresión —es la misma ausencia ya registrada— pero **crece**, y
conviene decirlo explícito: cada etapa pública que se suma antes de auth aumenta
la superficie de lo que un desconocido puede romper. Es una razón más para que
`Turno` no empiece sin auth.

---

## 11. Criterios de aceptación

La etapa está terminada cuando:

1. `make test-race` y `make lint` pasan en verde.
2. Un profesional puede cargar su semana, y cargar un bloque que no entra ningún
   hueco devuelve 422 nombrando el campo.
3. `GET /huecos` devuelve los huecos correctos para un rango, con los bloqueos
   restados y la anticipación mínima aplicada.
4. Un hueco que termina exactamente cuando empieza un bloqueo **no** se
   descarta, y hay un test que lo prueba.
5. Un profesional inactivo devuelve una lista vacía de huecos, no un error.
6. El rango pedido se recorta al horizonte del profesional, y la respuesta
   informa el rango que se usó de verdad.
7. `internal/domain` sigue sin importar ningún otro paquete del proyecto.
8. El contrato OpenAPI describe las seis rutas nuevas y `redocly lint` pasa sin
   errores.
