# Diseño — CRUD de Professional en Go

- **Fecha:** 2026-08-21
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de plan de implementación

---

## 1. Contexto

El repositorio contiene hoy dos cosas distintas bajo un único commit inicial:

- `APP salud/`: un laboratorio de investigación en Python (3.432 líneas) inspirado en
  `karpathy/autoresearch`, con el relevamiento del mercado de salud argentino,
  normativa, competidores y decisiones candidatas. Es investigación, no producto.
- `src/`: un prototipo React + Vite de la aplicación (2.191 líneas), enteramente
  mockeado — sin backend, sin persistencia, con autenticación falsa y datos
  hardcodeados.

Se decidió migrar de stack. Este documento especifica **la primera etapa de esa
migración**: un backend en Go con un CRUD de una sola entidad, sin base de datos,
que establece la arquitectura sobre la que se construye todo lo demás.

### Restricciones que vienen de fuera de lo técnico

| Restricción | Origen | Consecuencia |
|---|---|---|
| Los datos viven en Argentina | Decisión del fundador, Ley 25.326 | Nada de PaaS administrado (Supabase, Vercel, Neon, Firebase, Clerk). Todo auto-hospedable en contenedores. |
| Datos de salud son dato sensible | Ley 25.326, Ley 26.529 | Los registros no se destruyen a la ligera; hace falta rastro de auditoría a futuro. |
| El profesional verificado es el eje de confianza | `research/program.md`, relevamiento REFEPS | El estado de verificación existe en el modelo desde el día 1, aunque la integración no. |
| Alcance Etapa 1: agenda + pago + comprobante | `DEC-20260520-003` | Sin receta electrónica. ReNaPDiS queda fuera hasta la consulta legal. |

---

## 2. Decisiones tomadas

### Stack

| Área | Elección | Motivo |
|---|---|---|
| Backend | Go | Binario único, bajo consumo, deploy trivial en fierro propio argentino |
| Frontend | Next.js (etapa posterior) | El perfil público del profesional necesita SSR: es el canal de adquisición orgánica |
| Base de datos | Ninguna ahora, PostgreSQL después | Definir el dominio antes que el esquema |
| Router HTTP | `net/http` de la stdlib | Desde Go 1.22 el `ServeMux` resuelve method + path params |
| Contrato de API | OpenAPI escrito a mano | Go y TS no comparten tipos; el YAML es el único puente |
| Repositorio | Monorepo | Un solo lugar para back, front y contrato |

**Dependencias externas del backend: una.** `github.com/google/uuid`.
Todo lo demás sale de la biblioteca estándar: `net/http`, `log/slog`, `testing`,
`net/http/httptest`, `sync`, `context`.

### Idioma del código

Inglés para todo lo técnico. Español para los términos del dominio que no tienen
traducción fiel, y para todo lo que ve el usuario final.

| En español | Por qué no se traduce |
|---|---|
| `ObraSocial` | No es *health plan*, ni *insurance*, ni *HMO*. Es una figura del derecho argentino sin equivalente. |
| `Coseguro` | *Copay* es parecido pero no igual. |
| `Matricula` | *License number* pierde la carga institucional (colegio, jurisdicción, habilitación). |
| `Especialidad`, `Modalidad`, `Zona` | Forman parte del mismo vocabulario del negocio. |

El resto —`Professional`, `Appointment`, `Patient`, `Create`, `Repository`— en
inglés, que es lo idiomático en Go y lo que espera cualquier desarrollador que se
sume.

### Estructura del monorepo

```
salud/
├── apps/
│   ├── api/                  # backend Go — lo único que se construye en esta etapa
│   └── web/                  # Next.js — etapa posterior, no se crea todavía
├── packages/
│   ├── ui/                   # design system — etapa posterior
│   └── api-client/           # cliente TS generado del OpenAPI — etapa posterior
├── docs/superpowers/specs/
├── research/                 # era "APP salud/"
└── legacy/prototype/         # era "src/" + config de Vite
```

Tres movimientos de archivos, ninguno reescribe código:

1. `APP salud/` → `research/`. No es producto. El espacio en el nombre rompe
   scripts en cualquier shell.
2. `src/`, `index.html`, `vite.config.js`, `eslint.config.js`, `package.json`,
   `package-lock.json`, `public/` → `legacy/prototype/`. Es la especificación
   visual de los flujos. Sigue siendo ejecutable. Se elimina cuando `apps/web`
   lo reemplace.
3. Se crea `apps/api/`.

`packages/` y `apps/web/` quedan documentados como estructura acordada pero **no
se crean vacíos**. Se crean el día que se llenen.

### Arquitectura de un solo frontend

Cuando llegue la etapa del front: **una sola aplicación Next.js**, con las
audiencias separadas por *route groups* (`(public)`, `(paciente)`,
`(profesional)`, `(admin)`), no por aplicaciones distintas.

Cada route group tiene su propio layout, middleware de protección y límites de
error. Next hace code-splitting por ruta, así que el paciente nunca descarga el
bundle del profesional. Con un equipo de 2-3 personas, un build y un deploy es
una ventaja operativa real.

Para que separarlas después salga barato, lo compartido vive en `packages/ui` y
`packages/api-client`, nunca dentro de la app.

**Una sola API en Go para todas las audiencias, sin BFF por audiencia.** Los
Server Components de Next ya cumplen ese rol cuando una pantalla necesita
componer varias llamadas. Lo que se diferencia entre audiencias no es la API sino
los permisos, y eso se resuelve con autorización en el backend.

---

## 3. Arquitectura del backend

```
apps/api/
├── cmd/api/main.go              # composition root
├── internal/
│   ├── domain/                  # entidades, value objects, reglas, errores
│   │   ├── professional.go
│   │   ├── matricula.go
│   │   ├── money.go
│   │   ├── slug.go
│   │   └── errors.go
│   ├── repository/
│   │   ├── professional.go      # la INTERFAZ
│   │   └── memory/
│   │       ├── professional.go  # implementación en memoria
│   │       └── seed.go
│   ├── service/
│   │   └── professional.go      # casos de uso
│   └── handler/
│       ├── router.go
│       ├── professional.go      # controllers
│       ├── dto.go
│       ├── problem.go           # errores de dominio → HTTP
│       └── middleware.go
├── api/openapi.yaml
├── Dockerfile
├── Makefile
├── .golangci.yml
└── go.mod
```

### Dirección de las dependencias

```
handler ──▶ service ──▶ repository (interfaz)
   │           │              ▲
   └───────────┴──────────────┤
               ▼              │
            domain      repository/memory
```

`domain` no importa nada del proyecto. Si algún día lo hace, la arquitectura se
rompió y el compilador lo dice.

### Equivalencias con Spring / ASP.NET

| Spring / ASP.NET | Acá |
|---|---|
| `@RestController` | `internal/handler/` |
| `@Service` | `internal/service/` |
| `@Repository` | `internal/repository/memory/` |
| Entidad / Value Object | `internal/domain/` |
| Contenedor de DI | `cmd/api/main.go`, cableado explícito |

No hay anotaciones ni contenedor de inyección de dependencias. El cableado es
explícito y se lee de arriba abajo:

```go
repo := memory.NewProfessional()
svc  := service.NewProfessional(repo)
h    := handler.NewProfessional(svc)
```

### El punto de cambio a PostgreSQL

La interfaz `repository.Professional` y una línea de `main.go`:

```go
repo := memory.NewProfessional()              // hoy
// repo := postgres.NewProfessional(db)       // mañana
```

**Todos los métodos del repositorio reciben `context.Context` desde el día 1**,
aunque la implementación en memoria lo ignore. Agregarlo después obliga a tocar
todas las firmas.

```go
package repository

type Professional interface {
    Create(ctx context.Context, p domain.Professional) error
    GetByID(ctx context.Context, id uuid.UUID) (domain.Professional, error)
    GetBySlug(ctx context.Context, slug string) (domain.Professional, error)
    GetByMatricula(ctx context.Context, m domain.Matricula) (domain.Professional, error)
    List(ctx context.Context, f Filter) ([]domain.Professional, int, error)
    Update(ctx context.Context, p domain.Professional) error
}

type Filter struct {
    Especialidad *domain.Especialidad
    Zona         *string
    Status       *domain.Status
    Query        *string   // busca en nombre y apellido
    Limit        int
    Offset       int
}
```

No hay `Delete` en la interfaz: la baja es un `Update` que cambia el estado.
Ver sección 6.

---

## 4. Modelo de dominio

**Invariante central: no se puede construir un `Professional` inválido.** No hay
setters públicos ni structs armados a mano fuera del paquete `domain`. Hay un
constructor que valida y devuelve la entidad o un error.

```go
type Professional struct {
    ID            uuid.UUID
    Slug          string
    FirstName     string
    LastName      string
    Matricula     Matricula
    Especialidad  Especialidad
    Bio           string
    ConsultaPrice Money
    Modalidades   []Modalidad
    Zona          string
    ObrasSociales []string
    Status        Status
    Verification  VerificationStatus
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeactivatedAt *time.Time
}

func NewProfessional(in NewProfessionalInput) (Professional, error)
```

### Value objects

```go
type Money int64   // centavos, SIEMPRE. Money(1200000) == $12.000,00

func (m Money) String() string   // "$12.000,00" — formato argentino
```

Un `float64` no representa 0,10 exactamente. En un sistema que va a cobrar
consultas y liquidar honorarios eso no se negocia. El tipo propio impide sumar un
precio con una cantidad por accidente.

```go
type MatriculaTipo string   // MN (nacional) | MP (provincial)

type Matricula struct {
    Tipo   MatriculaTipo
    Numero string
}

func ParseMatricula(s string) (Matricula, error)
```

Validación **laxa y con normalización**. Las matrículas argentinas varían por
jurisdicción y profesión: `MN 98.234`, `MP 12345`, `M.N. 45321`, `mn98234`. Una
expresión regular estricta rechaza profesionales reales, que es el peor error
posible en este dominio. Se valida el tipo (`MN`/`MP`) y que el número tenga entre
1 y 10 dígitos; se normaliza a la forma canónica `MN 98234`. La verificación seria
llega con la integración a REFEPS.

```go
type Especialidad string  // psicologia | kinesiologia | odontologia
type Modalidad    string  // telemedicina | presencial | domicilio
type Status       string  // active | inactive
type VerificationStatus string  // pending | verified | rejected
```

Los tres verticales del enum salen de `research/data/vertical_scores.csv`. Agregar
un cuarto es una línea y recompilar.

`Status` y `Verification` son **dos ejes distintos y no se mezclan**: uno dice si
la matrícula está verificada en el mundo real, el otro si el profesional opera hoy
en la plataforma. Un profesional puede estar verificado y de licencia, o recién
anotado y sin verificar.

### Slug

Función pura que deriva la URL pública del profesional a partir del nombre.
Tiene que aguantar el español real:

| Entrada | Salida |
|---|---|
| `Martín González` | `martin-gonzalez` |
| `Íñigo Muñoz Ríos` | `inigo-munoz-rios` |
| `José  de  la  Cruz` | `jose-de-la-cruz` |

Normaliza acentos y eñes, pasa a minúsculas, colapsa espacios, une con guiones y
descarta todo lo que no sea `[a-z0-9-]`. La unicidad no es responsabilidad del
dominio (ver sección 5).

La normalización de base (`domain.Normalize`) se reutiliza en el filtro `q` del
listado, para que buscar `gonzalez` encuentre a `González`. Una sola función, dos
usos, un solo lugar donde arreglarla.

### Reglas de validación

Todas viven en `domain`, ninguna en el handler.

| Campo | Regla |
|---|---|
| `FirstName`, `LastName` | No vacíos tras recortar espacios, máximo 100 caracteres |
| `Matricula` | Parseable por `ParseMatricula` |
| `Especialidad` | Dentro del enum |
| `Bio` | Máximo 2000 caracteres, puede estar vacía |
| `ConsultaPrice` | Mayor o igual a cero |
| `Modalidades` | Al menos una, todas válidas, sin repetidas |
| `Zona` | No vacía, máximo 100 caracteres |
| `ObrasSociales` | Puede estar vacía, sin repetidas |

### Errores

```go
var (
    ErrNotFound       = errors.New("professional not found")
    ErrMatriculaTaken = errors.New("matricula already registered")
)

type FieldError struct {
    Field   string
    Message string
}

type ValidationError struct {
    Fields []FieldError
}

func (e ValidationError) Error() string
```

`ValidationError` acumula **todos** los campos inválidos en una pasada. Devolver
el primer error obliga al cliente a corregir de a uno.

---

## 5. Servicio

Tres reglas que necesitan consultar el repositorio, y por eso no pueden vivir en
el dominio:

1. **La matrícula es única.** Es la única identidad real de una persona en este
   sistema. Colisión → `ErrMatriculaTaken` → HTTP 409.
2. **El slug es único.** Si `martin-gonzalez` ya existe, el siguiente es
   `martin-gonzalez-2`, después `martin-gonzalez-3`. El sufijo se resuelve en el
   servicio, no en el dominio. Una colisión de slug **nunca es un error para el
   cliente**: se resuelve sola incrementando el sufijo hasta encontrar uno libre.
3. **Cambiar la matrícula o la especialidad devuelve `Verification` a `pending`.**
   Sale directo de `research/program.md`: *"toda orientación, agenda o cobro
   debería depender de un profesional verificado"*. Si se edita el dato sobre el
   que se apoya la verificación, la verificación deja de valer.

```go
type ProfessionalService struct {
    repo repository.Professional
}

func (s *ProfessionalService) Create(ctx context.Context, in CreateInput) (domain.Professional, error)
func (s *ProfessionalService) GetByID(ctx context.Context, id uuid.UUID) (domain.Professional, error)
func (s *ProfessionalService) GetBySlug(ctx context.Context, slug string) (domain.Professional, error)
func (s *ProfessionalService) List(ctx context.Context, f repository.Filter) ([]domain.Professional, int, error)
func (s *ProfessionalService) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (domain.Professional, error)
func (s *ProfessionalService) Deactivate(ctx context.Context, id uuid.UUID) error
func (s *ProfessionalService) Reactivate(ctx context.Context, id uuid.UUID) (domain.Professional, error)
```

`CreateInput` y `UpdateInput` son structs propios, **no la entidad**. Es la
frontera que impide que un cliente mande `CreatedAt`, `Verification` o `Status` y
se los guarde.

`Update` es reemplazo total (semántica de `PUT`) de los campos editables:
nombre, apellido, matrícula, especialidad, bio, precio, modalidades, zona y obras
sociales. `ID`, `Slug`, `Status`, `Verification` y las marcas de tiempo no son
editables por esta vía. El slug **no** se regenera al cambiar el nombre: es una
URL pública y romperla rompe enlaces y posicionamiento.

---

## 6. La API

Prefijo `/api/v1`.

| Método | Ruta | Qué hace |
|---|---|---|
| `GET` | `/professionals` | Lista con filtros y paginación |
| `POST` | `/professionals` | Crea |
| `GET` | `/professionals/{id}` | Trae por ID |
| `GET` | `/professionals/by-slug/{slug}` | Trae por slug — es lo que usará `/p/:slug` en el front |
| `PUT` | `/professionals/{id}` | Reemplaza los campos editables |
| `DELETE` | `/professionals/{id}` | Baja lógica |
| `POST` | `/professionals/{id}/reactivate` | Revierte la baja |
| `GET` | `/healthz` | Sin prefijo de versión |

Parámetros del listado: `especialidad`, `zona`, `status`, `q` (busca en nombre y
apellido), `limit` (por defecto 20, máximo 100), `offset` (por defecto 0).

El parámetro `q` compara **sin distinguir mayúsculas ni acentos**, reutilizando el
mismo normalizador del slug: buscar `gonzalez` tiene que encontrar a `González`.
En un producto argentino, lo contrario es una búsqueda rota.

Por defecto el listado devuelve **solo los activos**. `?status=inactive` muestra
los dados de baja.

**Comportamiento sobre profesionales inactivos**, para que no quede librado a la
implementación:

- `GET` por ID o por slug: `200` con `status: "inactive"`. El recurso existe.
- `PUT`: permitido. Editar los datos de alguien dado de baja no tiene por qué
  bloquearse, y no cambia su estado.
- `DELETE` sobre alguien ya inactivo: `204`, idempotente. No es un error.
- `POST /reactivate`: `200` con el profesional activo y `deactivatedAt` en null.
  Sobre alguien ya activo, `200` sin cambios.

### La baja lógica, y por qué no es un borrado

Se evaluaron los motivos reales por los que un profesional deja de estar
disponible:

| Situación | Qué tiene que pasar |
|---|---|
| Se va de la plataforma | Desaparece de las búsquedas. Turnos, comprobantes y pagos históricos **siguen existiendo**. |
| Se le vence o suspende la matrícula | Deja de recibir turnos de inmediato. El registro queda, con fecha y motivo. |
| Alta duplicada o mal cargada | Único caso donde conviene eliminar de verdad. |
| Ejercicio del derecho de supresión (Ley 25.326 art. 16) | Borrado real, con límites legales de retención. Es un trámite, no un endpoint. |
| Fraude o expulsión | Baja, nunca borrado: el rastro es justamente lo que se necesita. |

De cinco casos, en uno solo conviene eliminar el registro, y es un error de carga.
Además, un turno pasado apunta a un profesional: si el registro se evapora, ese
turno queda huérfano, y con él el comprobante que el paciente presentó para pedir
un reintegro.

Por eso `DELETE /professionals/{id}` conserva ese nombre —es lo que espera
cualquiera que lea un CRUD— pero por dentro pone `Status = inactive` y sella
`DeactivatedAt`. Devuelve `204`. Un `GET` por ID de un profesional inactivo
devuelve `200` con su estado: el recurso existe, simplemente no opera.

### Códigos de estado

| Situación | Código |
|---|---|
| Creado | `201` + header `Location` |
| Lectura o actualización correcta | `200` |
| Baja correcta | `204` |
| JSON malformado | `400` |
| JSON válido, datos inválidos | `422` con detalle por campo |
| Matrícula ya registrada | `409` |
| No existe | `404` |
| Panic recuperado | `500` |

Separar `400` de `422` importa: *"no entiendo lo que mandaste"* y *"entendí
perfecto y está mal"* son problemas distintos, y el front los maneja distinto.

### Convenciones del JSON

- **camelCase.** El único consumidor va a ser TypeScript; así no hace falta capa
  de traducción de nombres.
- **`consultaPriceCents: 1200000`**, entero de centavos, con el nombre explícito.
  Un `12000.50` viajando como número de JSON se convierte en `float64` del otro
  lado, y ahí empiezan los centavos que no cierran.
- Fechas en RFC 3339 UTC.
- La matrícula viaja como string canónica (`"MN 98234"`) y se parsea al entrar.

Ejemplo de respuesta:

```json
{
  "id": "3f2a1b4c-...",
  "slug": "martin-gonzalez",
  "firstName": "Martín",
  "lastName": "González",
  "matricula": "MN 98234",
  "especialidad": "psicologia",
  "bio": "Psicólogo clínico con orientación cognitivo-conductual.",
  "consultaPriceCents": 1200000,
  "modalidades": ["telemedicina", "presencial"],
  "zona": "CABA",
  "obrasSociales": ["OSDE", "Swiss Medical"],
  "status": "active",
  "verification": "pending",
  "createdAt": "2026-08-21T14:02:11Z",
  "updatedAt": "2026-08-21T14:02:11Z",
  "deactivatedAt": null
}
```

Listado:

```json
{
  "data": [ ... ],
  "pagination": { "total": 42, "limit": 20, "offset": 0 }
}
```

### Errores en `application/problem+json` (RFC 7807)

```json
{
  "type": "https://salud.app/errors/validation",
  "title": "Datos inválidos",
  "status": 422,
  "detail": "El profesional no pudo ser creado",
  "errors": [
    { "field": "matricula",   "message": "formato inválido" },
    { "field": "modalidades", "message": "se requiere al menos una" }
  ]
}
```

Es un estándar ya documentado, OpenAPI lo describe bien, y cuesta las mismas
líneas que inventar un formato propio.

### OpenAPI

`apps/api/api/openapi.yaml`, **escrito a mano y antes que el código Go**. Es la
fuente de verdad del contrato. Sirve para tres cosas: documentación navegable con
Swagger UI (la única forma de probar la API mientras no exista el front), generar
después el cliente TypeScript, y obligar a decidir la forma de la API antes de
codearla.

Se valida en CI. Si el YAML y los handlers se separan, el build falla.

---

## 7. Repositorio en memoria

```go
type Professional struct {
    mu   sync.RWMutex
    data map[uuid.UUID]domain.Professional
}
```

**Se guardan y se devuelven copias, nunca punteros.** Y como `Professional`
contiene slices (`Modalidades`, `ObrasSociales`), una copia superficial comparte
el array subyacente: quien reciba el profesional puede mutar el store desde afuera
sin enterarse. Los slices se copian explícitamente, y hay un test que lo verifica.
Es el bug número uno de este patrón.

El filtrado del listado es un scan lineal, marcado en el código:

```go
// ponytail: scan O(n), correcto para un store en memoria. La implementación
// Postgres resuelve esto con índices sobre especialidad, zona y status.
```

El **seed** son los cuatro profesionales de
`legacy/prototype/src/data/profesionales.js`, adaptados al modelo nuevo. Carga
únicamente con `APP_ENV=development`. Nunca en producción.

---

## 8. Testing

**No hay un solo mock en este proyecto.** El repositorio en memoria *es* el doble
de test: la implementación real ya es rápida y determinista. Regla que queda
escrita: si un test necesita un mock, la frontera está mal dibujada.

| Capa | Contra qué corre | Qué prueba |
|---|---|---|
| `domain` | Nada, funciones puras | Validaciones campo por campo, `ParseMatricula` con formatos reales, normalización de slug y búsqueda con acentos y eñes, formato de `Money` |
| `repository/memory` | Sí mismo | Que las copias sean copias de verdad (mutar lo devuelto no toca el store), filtros, paginación |
| `service` | El repositorio en memoria real | Matrícula duplicada → `ErrMatriculaTaken`; unicidad de slug con sufijo; cambio de matrícula → `Verification` vuelve a `pending`; baja → sale del listado por defecto; baja y reactivación → vuelve a aparecer |
| `handler` | El stack completo por `httptest` | Códigos de estado, forma del `problem+json`, header `Location`, JSON malformado → 400 |
| Concurrencia | `go test -race` | N goroutines escribiendo el store en paralelo |

Los tests de dominio son table-driven, que es lo idiomático en Go y lo que hace
que agregar un caso sea agregar una línea.

**No se fija un porcentaje de cobertura.** Se prueban las reglas, no las líneas.

---

## 9. Operación

- **Configuración por variables de entorno**: `PORT` (8080), `APP_ENV`
  (`development`), `LOG_LEVEL` (`info`), `SHUTDOWN_TIMEOUT` (10s). Un struct que
  se lee al arrancar y falla rápido si algo está mal. Sin librería de config.
- **Logs con `log/slog`** en JSON, con request ID propagado por `context`.
- **Middleware**: request ID, logging de acceso, recuperación de panics, timeout.
  **CORS todavía no**: no hay navegador que lo necesite hasta que exista
  `apps/web`.
- **Apagado gracioso**: `signal.NotifyContext` + `srv.Shutdown`. Evita cortar un
  request a la mitad en cada deploy.
- **Timeouts del servidor**: `ReadTimeout`, `WriteTimeout`, `IdleTimeout`
  configurados explícitamente. Los valores por defecto de `http.Server` son cero,
  es decir, sin límite.
- **Dockerfile multi-etapa** sobre `distroless`, usuario no-root. Imagen final
  alrededor de 15 MB.
- **Makefile**: `run`, `test`, `test-race`, `lint`, `openapi-lint`, `build`,
  `docker-build`.
- **CI en GitHub Actions**: `golangci-lint`, `go test -race ./...`, validación del
  OpenAPI, build de la imagen.

---

## 10. Fuera de alcance

Lo que **no** entra en esta etapa, con el disparador que lo va a activar:

| Pendiente | Cuándo se activa |
|---|---|
| Autenticación y autorización | Spec propia. Obligatorio antes de exponer la API a internet. |
| Entidades `Patient` y `Appointment` | Después de que este esqueleto esté cerrado |
| Migración a PostgreSQL | Cuando el modelo de dominio deje de moverse |
| Integración real con REFEPS | Spec propia. Hoy `Verification` queda en `pending`. |
| `StatusSuspended` | Cuando exista REFEPS y algo pueda disparar una suspensión |
| Endpoint de purga (Ley 25.326 art. 16) | Cuando el abogado sanitario defina qué se está obligado a conservar |
| Rating, reseñas, horarios, coseguros | Son entidades propias, no campos de `Professional` |
| CORS | Cuando exista `apps/web` |
| `apps/web`, `packages/ui`, `packages/api-client` | Etapa del frontend |
| Auditoría de accesos a datos sensibles | Cuando haya datos de pacientes y autenticación |

### Deuda que este diseño hereda del prototipo

Cosas que están en `legacy/prototype/` y **no deben portarse** cuando se
construya el front:

- El formulario de tarjeta de `Reserva.jsx`: los datos de tarjeta no pueden pasar
  por código propio (PCI-DSS). Va gateway hospedado.
- El texto *"El reintegro se gestiona automáticamente"*: `research/program.md`
  lo prohíbe explícitamente — no prometer reintegro sin convenio real.
- Las validaciones simuladas de RENAPER y "convenios de la Red": no existen.

---

## 11. Criterios de aceptación

La etapa está terminada cuando:

1. `make test-race` pasa en verde.
2. `make lint` pasa sin advertencias.
3. Los seis endpoints responden según el `openapi.yaml`, verificable desde
   Swagger UI.
4. `docker build` produce una imagen que arranca y responde `GET /healthz`.
5. Cambiar `memory.NewProfessional()` por otra implementación de
   `repository.Professional` en `main.go` es el **único** cambio necesario para
   migrar de almacenamiento.
6. `internal/domain` no importa ningún otro paquete del proyecto.
