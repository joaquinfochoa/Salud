# CRUD de Profesional en Go — Plan de Implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Construir el backend en Go con un CRUD completo de `Profesional`, sin base de datos, con arquitectura en capas y el punto de cambio a PostgreSQL aislado en una interfaz.

**Architecture:** Cuatro capas — `handler → service → repository → domain` — donde `domain` no importa nada del proyecto. El repositorio en memoria implementa una interfaz que después implementará PostgreSQL sin tocar el resto del código. El cableado de dependencias es explícito en `cmd/api/main.go`, sin contenedor de inyección.

**Tech Stack:** Go 1.24+, `net/http` de la stdlib (ServeMux con patrones de Go 1.22+), `log/slog`, `testing` + `net/http/httptest`. Una única dependencia externa: `github.com/google/uuid`.

## Global Constraints

- **Module path:** `github.com/joaquinfochoa/Salud/apps/api`
- **Go mínimo:** 1.24 (el `ServeMux` con `"GET /path/{id}"` requiere 1.22+)
- **Dependencias externas permitidas:** únicamente `github.com/google/uuid`. Cualquier otro `go get` es un error del plan — preguntar antes.
- **Idioma del código:** todo lo que escribimos nosotros va en español — tipos, funciones, campos, constantes, comentarios, mensajes y los nombres de campo del JSON. Los paquetes quedan en inglés (`domain`, `repository`, `memory`, `service`, `handler`, `config`), porque nombran patrones arquitectónicos, no el negocio. Quedan en inglés por contrato externo: `String()` y `Error()` (interfaces de Go), las variables de entorno y sus valores, y las claves `type`/`title`/`status`/`detail` del RFC 7807. Sin híbridos.
- **Dinero:** `int64` en centavos, tipo `domain.Dinero`. Nunca `float64`, nunca en ningún lugar.
- **JSON:** camelCase. El precio viaja como `precioConsultaCentavos` (entero).
- **Errores HTTP:** `application/problem+json` (RFC 7807).
- **Sin mocks.** El repositorio en memoria es el doble de test. Si un test parece necesitar un mock, la frontera está mal dibujada — parar y preguntar.
- **`internal/domain` no importa ningún otro paquete del proyecto.** Es verificable y es criterio de aceptación.
- **Comentarios:** en español, explicando el *por qué*, no el *qué*. No comentar lo obvio.

---

## Estructura de archivos

| Archivo | Responsabilidad |
|---|---|
| `apps/api/go.mod` | Módulo y dependencias |
| `apps/api/internal/domain/dinero.go` | Tipo `Dinero` y su formato |
| `apps/api/internal/domain/texto.go` | `Normalizar` y `GenerarSlug` — compartidos por slug y búsqueda |
| `apps/api/internal/domain/enums.go` | `Especialidad`, `Modalidad`, `Estado`, `EstadoVerificacion` |
| `apps/api/internal/domain/matricula.go` | Value object `Matricula` y su parser |
| `apps/api/internal/domain/errores.go` | Errores centinela y `ErrorValidacion` |
| `apps/api/internal/domain/profesional.go` | La entidad, su constructor y sus transiciones |
| `apps/api/internal/repository/profesional.go` | La interfaz y el `Filtro` |
| `apps/api/internal/repository/memory/profesional.go` | Implementación en memoria |
| `apps/api/internal/repository/memory/semilla.go` | Datos de desarrollo |
| `apps/api/internal/service/profesional.go` | Casos de uso |
| `apps/api/internal/handler/problema.go` | Errores de dominio → HTTP |
| `apps/api/internal/handler/dto.go` | Structs de request y response |
| `apps/api/internal/handler/profesional.go` | Controllers |
| `apps/api/internal/handler/middleware.go` | `IDPeticion`, logging, recover |
| `apps/api/internal/handler/router.go` | Tabla de rutas |
| `apps/api/internal/config/config.go` | Configuración por variables de entorno |
| `apps/api/cmd/api/main.go` | Composition root |
| `apps/api/api/openapi.yaml` | El contrato |

---

## Task 1: Reestructurar el monorepo e inicializar el módulo Go

**Files:**
- Move: `APP salud/` → `research/`
- Move: `src/`, `public/`, `index.html`, `vite.config.js`, `eslint.config.js`, `package.json`, `package-lock.json` → `legacy/prototype/`
- Create: `apps/api/go.mod`
- Create: `apps/api/cmd/api/main.go`
- Create: `README.md` (reemplaza el del template de Vite)

**Interfaces:**
- Consumes: nada
- Produces: el módulo `github.com/joaquinfochoa/Salud/apps/api` compilable

- [ ] **Step 1: Mover el lab de research**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git mv "APP salud" research
```

- [ ] **Step 2: Mover el prototipo React**

```bash
mkdir -p legacy/prototype
git mv src legacy/prototype/src
git mv public legacy/prototype/public
git mv index.html legacy/prototype/index.html
git mv vite.config.js legacy/prototype/vite.config.js
git mv eslint.config.js legacy/prototype/eslint.config.js
git mv package.json legacy/prototype/package.json
git mv package-lock.json legacy/prototype/package-lock.json
```

- [ ] **Step 3: Verificar que el prototipo sigue siendo ejecutable**

Run: `cd legacy/prototype && npm install && npm run build`
Expected: el build de Vite termina con `✓ built in ...`. Si falla, el movimiento rompió una ruta relativa — arreglarla antes de seguir.

- [ ] **Step 4: Inicializar el módulo Go**

```bash
cd "c:/Users/gianl/Desktop/Salud"
mkdir -p apps/api/cmd/api
cd apps/api
go mod init github.com/joaquinfochoa/Salud/apps/api
go get github.com/google/uuid
```

- [ ] **Step 5: Escribir un `main.go` mínimo que compile**

Archivo `apps/api/cmd/api/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("salud api")
}
```

- [ ] **Step 6: Verificar que compila y corre**

Run: `cd apps/api && go build ./... && go run ./cmd/api`
Expected: imprime `salud api` y termina con código 0.

- [ ] **Step 7: Escribir el README de la raíz**

Archivo `README.md` (reemplaza completo el contenido actual, que es el template de Vite):

```markdown
# Salud

Plataforma de salud digital. Monorepo.

## Estructura

| Carpeta | Qué es |
|---|---|
| `apps/api` | Backend en Go. Ver su README. |
| `apps/web` | Frontend en Next.js. Todavía no existe. |
| `research/` | Lab de investigación en Python y el relevamiento de mercado. No es producto. |
| `legacy/prototype/` | Prototipo React + Vite, mockeado. Es la especificación visual de los flujos. Se elimina cuando `apps/web` lo reemplace. |
| `docs/superpowers/` | Specs de diseño y planes de implementación. |

## Backend

```bash
cd apps/api
go run ./cmd/api
```

## Prototipo (referencia visual)

```bash
cd legacy/prototype
npm install && npm run dev
```
```

- [ ] **Step 8: Actualizar el `.gitignore` de la raíz**

Agregar al final de `.gitignore`:

```gitignore

# Go
apps/api/bin/
*.exe
coverage.out
```

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add -A
git commit -m "refactor: reestructurar como monorepo e inicializar el módulo Go

- APP salud/ -> research/ (el espacio en el nombre rompía scripts)
- src/ y config de Vite -> legacy/prototype/ (sigue ejecutable, es la
  especificación visual de los flujos)
- apps/api/ con el módulo Go inicializado

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 2: `Dinero` y normalización de texto

Dos piezas puras, sin dependencias, que el resto del dominio usa.

**Files:**
- Create: `apps/api/internal/domain/dinero.go`
- Create: `apps/api/internal/domain/texto.go`
- Test: `apps/api/internal/domain/dinero_test.go`
- Test: `apps/api/internal/domain/texto_test.go`

**Interfaces:**
- Consumes: nada
- Produces:
  - `domain.Dinero` (`int64`), `func (Dinero) String() string`
  - `func domain.Normalizar(string) string`
  - `func domain.GenerarSlug(string) string`

- [ ] **Step 1: Escribir los tests de `Dinero`**

Archivo `apps/api/internal/domain/dinero_test.go`:

```go
package domain

import "testing"

func TestDineroString(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  Dinero
		esperado string
	}{
		{"cero", 0, "$0,00"},
		{"un peso", 100, "$1,00"},
		{"centavos sueltos", 5, "$0,05"},
		{"menos de un peso", 999, "$9,99"},
		{"miles", 1200000, "$12.000,00"},
		{"millones", 123456789, "$1.234.567,89"},
		{"tres digitos", 99900, "$999,00"},
		{"negativo", -50, "-$0,50"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := caso.entrada.String(); obtenido != caso.esperado {
				t.Errorf("Dinero(%d).String() = %q, se esperaba %q", int64(caso.entrada), obtenido, caso.esperado)
			}
		})
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run TestDineroString -v`
Expected: FAIL con `undefined: Dinero`

- [ ] **Step 3: Implementar `Dinero`**

Archivo `apps/api/internal/domain/dinero.go`:

```go
package domain

import (
	"strconv"
	"strings"
)

// Dinero representa un monto en centavos.
//
// Nunca usar float para dinero: float64 no puede representar 0,10 de forma
// exacta, y este sistema va a cobrar consultas y liquidar honorarios. El tipo
// propio además impide sumar un precio con una cantidad por accidente.
type Dinero int64

// String formatea el monto con la convención argentina: punto para miles,
// coma para decimales. Dinero(1200000) → "$12.000,00"
func (m Dinero) String() string {
	negativo := m < 0
	if negativo {
		m = -m
	}

	pesos := int64(m) / 100
	centavos := int64(m) % 100

	digitos := strconv.FormatInt(pesos, 10)
	var b strings.Builder
	for i := range digitos {
		if i > 0 && (len(digitos)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(digitos[i])
	}

	centavosStr := strconv.FormatInt(centavos, 10)
	if centavos < 10 {
		centavosStr = "0" + centavosStr
	}

	salida := "$" + b.String() + "," + centavosStr
	if negativo {
		return "-" + salida
	}
	return salida
}
```

- [ ] **Step 4: Correr el test y verificar que pasa**

Run: `cd apps/api && go test ./internal/domain/ -run TestDineroString -v`
Expected: PASS, los ocho subtests en verde.

- [ ] **Step 5: Escribir los tests de normalización**

Archivo `apps/api/internal/domain/texto_test.go`:

```go
package domain

import "testing"

func TestNormalizar(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado string
	}{
		{"minusculas", "GONZÁLEZ", "gonzalez"},
		{"acentos", "Martín González", "martin gonzalez"},
		{"enie", "Muñoz", "munoz"},
		{"dieresis", "Agüero", "aguero"},
		{"todas las vocales", "áéíóú", "aeiou"},
		{"recorta espacios", "  Ana  ", "ana"},
		{"vacio", "", ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := Normalizar(caso.entrada); obtenido != caso.esperado {
				t.Errorf("Normalizar(%q) = %q, se esperaba %q", caso.entrada, obtenido, caso.esperado)
			}
		})
	}
}

func TestGenerarSlug(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado string
	}{
		{"nombre simple", "Martín González", "martin-gonzalez"},
		{"acentos y enies", "Íñigo Muñoz Ríos", "inigo-munoz-rios"},
		{"espacios repetidos", "José  de  la  Cruz", "jose-de-la-cruz"},
		{"puntuacion", "Dr. Juan Pérez", "dr-juan-perez"},
		{"guiones existentes", "Ana-María López", "ana-maria-lopez"},
		{"numeros", "Clínica 24hs", "clinica-24hs"},
		{"solo simbolos", "...", ""},
		{"vacio", "", ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := GenerarSlug(caso.entrada); obtenido != caso.esperado {
				t.Errorf("GenerarSlug(%q) = %q, se esperaba %q", caso.entrada, obtenido, caso.esperado)
			}
		})
	}
}
```

- [ ] **Step 6: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestNormalizar|TestGenerarSlug' -v`
Expected: FAIL con `undefined: Normalizar` y `undefined: GenerarSlug`

- [ ] **Step 7: Implementar la normalización**

Archivo `apps/api/internal/domain/texto.go`:

```go
package domain

import (
	"strings"
	"unicode"
)

// Normalizar baja a minúsculas y saca acentos y eñes, para poder comparar
// "González" con "gonzalez".
//
// Lo usan dos cosas: la generación del slug y el filtro de búsqueda del
// listado. Una sola función, un solo lugar donde arreglarla. En un producto
// argentino, una búsqueda que distingue acentos es una búsqueda rota.
func Normalizar(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch r {
		case 'á', 'à', 'ä', 'â', 'ã':
			b.WriteRune('a')
		case 'é', 'è', 'ë', 'ê':
			b.WriteRune('e')
		case 'í', 'ì', 'ï', 'î':
			b.WriteRune('i')
		case 'ó', 'ò', 'ö', 'ô', 'õ':
			b.WriteRune('o')
		case 'ú', 'ù', 'ü', 'û':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		case 'ç':
			b.WriteRune('c')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GenerarSlug genera la parte legible de la URL pública del profesional.
// "Íñigo Muñoz Ríos" → "inigo-munoz-rios"
//
// No se ocupa de la unicidad: eso necesita mirar los demás profesionales y
// por lo tanto vive en el servicio.
func GenerarSlug(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	guionPendiente := false
	for _, r := range Normalizar(s) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			if guionPendiente && b.Len() > 0 {
				b.WriteByte('-')
			}
			guionPendiente = false
			b.WriteRune(r)
		case unicode.IsSpace(r), r == '-', r == '_':
			guionPendiente = true
		}
		// el resto (puntos, comas, símbolos) se descarta sin separar
	}
	return b.String()
}
```

- [ ] **Step 8: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -v`
Expected: PASS en `TestDineroString`, `TestNormalizar` y `TestGenerarSlug`.

Verificar el caso `"Dr. Juan Pérez"`: el punto se descarta sin marcar separación, pero el espacio que le sigue sí la marca. Resultado `dr-juan-perez`, no `dr--juan-perez`.

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): Dinero en centavos y normalización de texto

Dinero es int64 de centavos con formato argentino. Normalizar y GenerarSlug
comparten la misma normalización de acentos: la usa el slug público y
el filtro de búsqueda del listado.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 3: `Matricula` y los enums del dominio

**Files:**
- Create: `apps/api/internal/domain/matricula.go`
- Create: `apps/api/internal/domain/enums.go`
- Test: `apps/api/internal/domain/matricula_test.go`
- Test: `apps/api/internal/domain/enums_test.go`

**Interfaces:**
- Consumes: nada
- Produces:
  - `domain.MatriculaTipo` con `MatriculaNacional` (`"MN"`) y `MatriculaProvincial` (`"MP"`)
  - `domain.Matricula{Tipo MatriculaTipo; Numero string}`
  - `func domain.ParsearMatricula(string) (Matricula, error)`
  - `func (Matricula) String() string`, `func (Matricula) EsCero() bool`
  - `domain.Especialidad` con `EspecialidadPsicologia`, `EspecialidadKinesiologia`, `EspecialidadOdontologia`
  - `domain.Modalidad` con `ModalidadTelemedicina`, `ModalidadPresencial`, `ModalidadDomicilio`
  - `domain.Estado` con `EstadoActivo`, `EstadoInactivo`
  - `domain.EstadoVerificacion` con `VerificacionPendiente`, `VerificacionVerificada`, `VerificacionRechazada`
  - Cada uno con método `EsValida() bool` (`EsValido() bool` en `Estado` y `EstadoVerificacion`)

- [ ] **Step 1: Escribir los tests de `ParsearMatricula`**

Archivo `apps/api/internal/domain/matricula_test.go`:

```go
package domain

import "testing"

func TestParsearMatriculaValida(t *testing.T) {
	casos := []struct {
		nombre         string
		entrada        string
		tipoEsperado   MatriculaTipo
		numeroEsperado string
	}{
		{"formato canonico", "MN 98234", MatriculaNacional, "98234"},
		{"con puntos de miles", "MN 98.234", MatriculaNacional, "98234"},
		{"con puntos en el tipo", "M.N. 45321", MatriculaNacional, "45321"},
		{"sin espacios ni puntos", "mn98234", MatriculaNacional, "98234"},
		{"provincial", "MP 12345", MatriculaProvincial, "12345"},
		{"minusculas", "mp 12345", MatriculaProvincial, "12345"},
		{"con guion", "MN-98234", MatriculaNacional, "98234"},
		{"un solo digito", "MN 7", MatriculaNacional, "7"},
		{"diez digitos", "MN 1234567890", MatriculaNacional, "1234567890"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			m, err := ParsearMatricula(caso.entrada)
			if err != nil {
				t.Fatalf("ParsearMatricula(%q) devolvió error: %v", caso.entrada, err)
			}
			if m.Tipo != caso.tipoEsperado {
				t.Errorf("tipo = %q, se esperaba %q", m.Tipo, caso.tipoEsperado)
			}
			if m.Numero != caso.numeroEsperado {
				t.Errorf("numero = %q, se esperaba %q", m.Numero, caso.numeroEsperado)
			}
		})
	}
}

func TestParsearMatriculaInvalida(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
	}{
		{"vacia", ""},
		{"solo el tipo", "MN"},
		{"sin tipo", "98234"},
		{"tipo desconocido", "XX 98234"},
		{"numero con letras", "MN 98A34"},
		{"mas de diez digitos", "MN 12345678901"},
		{"solo espacios", "   "},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if _, err := ParsearMatricula(caso.entrada); err == nil {
				t.Errorf("ParsearMatricula(%q) debía fallar y no falló", caso.entrada)
			}
		})
	}
}

func TestMatriculaString(t *testing.T) {
	m, err := ParsearMatricula("m.n. 98.234")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// distintas formas de escribir la misma matrícula tienen que converger
	// a una sola representación, o la unicidad no sirve de nada
	if obtenido := m.String(); obtenido != "MN 98234" {
		t.Errorf("String() = %q, se esperaba %q", obtenido, "MN 98234")
	}
}

func TestMatriculaEsCero(t *testing.T) {
	var cero Matricula
	if !cero.EsCero() {
		t.Error("la matrícula vacía debía ser EsCero")
	}

	m, _ := ParsearMatricula("MN 1")
	if m.EsCero() {
		t.Error("una matrícula parseada no debía ser EsCero")
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run TestMatricula -v`
Expected: FAIL con `undefined: ParsearMatricula`

- [ ] **Step 3: Implementar `Matricula`**

Archivo `apps/api/internal/domain/matricula.go`:

```go
package domain

import (
	"errors"
	"strings"
)

type MatriculaTipo string

const (
	MatriculaNacional   MatriculaTipo = "MN"
	MatriculaProvincial MatriculaTipo = "MP"
)

const maxDigitosMatricula = 10

// Matricula es la identidad profesional de una persona: es el único dato que
// la ata a una habilitación real y es sobre lo que se apoya toda la confianza
// del producto.
type Matricula struct {
	Tipo   MatriculaTipo
	Numero string
}

var limpiadorMatricula = strings.NewReplacer(".", "", " ", "", "-", "", "/", "")

// ParsearMatricula acepta las formas que se usan en la práctica —"MN 98.234",
// "M.N. 45321", "mn98234", "MP 12345"— y las normaliza a "MN 98234".
//
// La validación es deliberadamente laxa. Las matrículas argentinas varían por
// jurisdicción y por profesión, y rechazar a un profesional real es peor error
// que aceptar un número raro: el que queda afuera no vuelve. La verificación
// seria llega cuando exista la integración con REFEPS.
func ParsearMatricula(s string) (Matricula, error) {
	limpia := limpiadorMatricula.Replace(strings.ToUpper(s))

	if len(limpia) < 3 {
		return Matricula{}, errors.New("debe tener tipo (MN o MP) y número")
	}

	tipo := MatriculaTipo(limpia[:2])
	if tipo != MatriculaNacional && tipo != MatriculaProvincial {
		return Matricula{}, errors.New("el tipo debe ser MN (nacional) o MP (provincial)")
	}

	numero := limpia[2:]
	if len(numero) > maxDigitosMatricula {
		return Matricula{}, errors.New("el número no puede tener más de 10 dígitos")
	}
	for i := range numero {
		if numero[i] < '0' || numero[i] > '9' {
			return Matricula{}, errors.New("el número solo puede contener dígitos")
		}
	}

	return Matricula{Tipo: tipo, Numero: numero}, nil
}

// String devuelve la forma canónica. Dos matrículas escritas distinto pero
// iguales en el fondo comparan iguales porque el parser las converge acá.
func (m Matricula) String() string {
	return string(m.Tipo) + " " + m.Numero
}

func (m Matricula) EsCero() bool {
	return m.Tipo == "" && m.Numero == ""
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -run TestMatricula -v`
Expected: PASS. Verificar que `"MN"` (solo el tipo) falla por longitud y que `"98234"` falla por tipo desconocido (`"98"`).

- [ ] **Step 5: Escribir los tests de los enums**

Archivo `apps/api/internal/domain/enums_test.go`:

```go
package domain

import "testing"

func TestEspecialidadEsValida(t *testing.T) {
	validas := []Especialidad{
		EspecialidadPsicologia,
		EspecialidadKinesiologia,
		EspecialidadOdontologia,
	}
	for _, e := range validas {
		if !e.EsValida() {
			t.Errorf("Especialidad(%q) debía ser válida", e)
		}
	}

	invalidas := []Especialidad{"", "cardiologia", "Psicologia", "PSICOLOGIA"}
	for _, e := range invalidas {
		if e.EsValida() {
			t.Errorf("Especialidad(%q) no debía ser válida", e)
		}
	}
}

func TestModalidadEsValida(t *testing.T) {
	validas := []Modalidad{ModalidadTelemedicina, ModalidadPresencial, ModalidadDomicilio}
	for _, m := range validas {
		if !m.EsValida() {
			t.Errorf("Modalidad(%q) debía ser válida", m)
		}
	}

	invalidas := []Modalidad{"", "online", "Presencial"}
	for _, m := range invalidas {
		if m.EsValida() {
			t.Errorf("Modalidad(%q) no debía ser válida", m)
		}
	}
}

func TestEstadoEsValido(t *testing.T) {
	if !EstadoActivo.EsValido() || !EstadoInactivo.EsValido() {
		t.Error("activo e inactivo debían ser válidos")
	}
	if Estado("suspendido").EsValido() {
		t.Error("suspendido todavía no existe: no debía ser válido")
	}
}

func TestEstadoVerificacionEsValido(t *testing.T) {
	validos := []EstadoVerificacion{VerificacionPendiente, VerificacionVerificada, VerificacionRechazada}
	for _, v := range validos {
		if !v.EsValido() {
			t.Errorf("EstadoVerificacion(%q) debía ser válido", v)
		}
	}
	if EstadoVerificacion("desconocido").EsValido() {
		t.Error("desconocido no debía ser válido")
	}
}
```

- [ ] **Step 6: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestEspecialidad|TestModalidad|TestEstado' -v`
Expected: FAIL con `undefined: Especialidad`

- [ ] **Step 7: Implementar los enums**

Archivo `apps/api/internal/domain/enums.go`:

```go
package domain

// Especialidad son los tres verticales de lanzamiento definidos en
// research/data/vertical_scores.csv. Es un enum cerrado a propósito: con texto
// libre terminás con "Psicología", "psicologia" y "Psicólogo clínico" como tres
// especialidades distintas, y los filtros dejan de servir.
//
// Agregar una cuarta es una constante y un caso más en EsValida().
type Especialidad string

const (
	EspecialidadPsicologia   Especialidad = "psicologia"
	EspecialidadKinesiologia Especialidad = "kinesiologia"
	EspecialidadOdontologia  Especialidad = "odontologia"
)

func (e Especialidad) EsValida() bool {
	switch e {
	case EspecialidadPsicologia, EspecialidadKinesiologia, EspecialidadOdontologia:
		return true
	}
	return false
}

type Modalidad string

const (
	ModalidadTelemedicina Modalidad = "telemedicina"
	ModalidadPresencial   Modalidad = "presencial"
	ModalidadDomicilio    Modalidad = "domicilio"
)

func (m Modalidad) EsValida() bool {
	switch m {
	case ModalidadTelemedicina, ModalidadPresencial, ModalidadDomicilio:
		return true
	}
	return false
}

// Estado dice si el profesional opera hoy en la plataforma.
//
// No confundir con EstadoVerificacion: son dos ejes distintos. Un profesional
// puede estar verificado y de licencia, o recién anotado y sin verificar.
type Estado string

const (
	EstadoActivo   Estado = "activo"
	EstadoInactivo Estado = "inactivo"
)

func (s Estado) EsValido() bool {
	switch s {
	case EstadoActivo, EstadoInactivo:
		return true
	}
	return false
}

// EstadoVerificacion dice si la matrícula fue verificada contra el mundo real.
// Por ahora todos nacen en pendiente: la integración con REFEPS es una etapa
// posterior y nada la mueve automáticamente todavía.
type EstadoVerificacion string

const (
	VerificacionPendiente  EstadoVerificacion = "pendiente"
	VerificacionVerificada EstadoVerificacion = "verificada"
	VerificacionRechazada  EstadoVerificacion = "rechazada"
)

func (v EstadoVerificacion) EsValido() bool {
	switch v {
	case VerificacionPendiente, VerificacionVerificada, VerificacionRechazada:
		return true
	}
	return false
}
```

- [ ] **Step 8: Correr toda la suite del dominio**

Run: `cd apps/api && go test ./internal/domain/ -v`
Expected: PASS en todos los tests de Task 2 y Task 3.

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): value object Matricula y enums del dominio

ParsearMatricula acepta los formatos reales del mercado argentino y los
normaliza a una forma canónica. La validación es laxa a propósito:
rechazar a un profesional real es peor que aceptar un número raro.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 4: Errores del dominio y la entidad `Profesional`

El corazón del dominio: la invariante de que no se puede construir un profesional inválido.

**Files:**
- Create: `apps/api/internal/domain/errores.go`
- Create: `apps/api/internal/domain/profesional.go`
- Test: `apps/api/internal/domain/profesional_test.go`

**Interfaces:**
- Consumes: `Matricula`, `Especialidad`, `Modalidad`, `Estado`, `EstadoVerificacion`, `Dinero`, `GenerarSlug`, `Normalizar` (Tasks 2 y 3)
- Produces:
  - `domain.ErrNoEncontrado`, `domain.ErrMatriculaEnUso`
  - `domain.ErrorCampo{Campo, Mensaje string}`
  - `domain.ErrorValidacion{Campos []ErrorCampo}` con método `Error() string`
  - `domain.EntradaProfesional` — struct de entrada con todo en tipos primitivos
  - `domain.Profesional` — la entidad
  - `func domain.NuevoProfesional(EntradaProfesional, time.Time) (Profesional, error)`
  - `func (Profesional) AplicarCambios(EntradaProfesional, time.Time) (Profesional, error)`
  - `func (Profesional) DarDeBaja(time.Time) Profesional`
  - `func (Profesional) Reactivar(time.Time) Profesional`
  - `func (Profesional) Clonar() Profesional`
  - `func (Profesional) NombreCompleto() string`

- [ ] **Step 1: Escribir `errores.go`**

Este archivo no lleva test propio: sus dos funciones se ejercitan enteras desde los tests de la entidad del paso 3.

Archivo `apps/api/internal/domain/errores.go`:

```go
package domain

import (
	"errors"
	"strings"
)

var (
	// ErrNoEncontrado lo devuelve el repositorio cuando no existe el registro.
	ErrNoEncontrado = errors.New("profesional no encontrado")

	// ErrMatriculaEnUso lo devuelve el servicio: la matrícula es la única
	// identidad real de una persona en este sistema y no puede repetirse.
	ErrMatriculaEnUso = errors.New("matricula ya registrada")
)

// ErrorCampo señala un campo puntual. Las etiquetas JSON coinciden con el
// formato problem+json que espera el cliente.
type ErrorCampo struct {
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}

// ErrorValidacion junta todos los campos inválidos de una sola pasada.
// Devolver solo el primero obliga al cliente a corregir de a uno, que es una
// experiencia horrible en un formulario de alta con nueve campos.
type ErrorValidacion struct {
	Campos []ErrorCampo
}

func (e ErrorValidacion) Error() string {
	partes := make([]string, 0, len(e.Campos))
	for _, f := range e.Campos {
		partes = append(partes, f.Campo+": "+f.Mensaje)
	}
	return "validación fallida — " + strings.Join(partes, "; ")
}

func (e *ErrorValidacion) agregar(campo, mensaje string) {
	e.Campos = append(e.Campos, ErrorCampo{Campo: campo, Mensaje: mensaje})
}

func (e ErrorValidacion) tieneErrores() bool {
	return len(e.Campos) > 0
}
```

- [ ] **Step 2: Escribir los tests de la entidad**

Archivo `apps/api/internal/domain/profesional_test.go`:

```go
package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var ahoraDePrueba = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

// entradaValida devuelve una entrada correcta. Cada test la copia y rompe un
// solo campo, así el que falla es siempre el campo bajo prueba.
func entradaValida() EntradaProfesional {
	return EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico con orientación cognitivo-conductual.",
		PrecioConsulta: 1200000,
		Modalidades:    []string{"telemedicina", "presencial"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE", "Swiss Medical"},
	}
}

func TestNuevoProfesionalValido(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("el ID debía generarse")
	}
	if p.Slug != "martin-gonzalez" {
		t.Errorf("Slug = %q, se esperaba %q", p.Slug, "martin-gonzalez")
	}
	if p.Matricula.String() != "MN 98234" {
		t.Errorf("Matricula = %q, se esperaba %q", p.Matricula, "MN 98234")
	}
	if p.PrecioConsulta != Dinero(1200000) {
		t.Errorf("PrecioConsulta = %d, se esperaba 1200000", p.PrecioConsulta)
	}
	if p.Estado != EstadoActivo {
		t.Errorf("Estado = %q, se esperaba activo", p.Estado)
	}
	// nadie nace verificado: la verificación es un acto contra REFEPS
	if p.Verificacion != VerificacionPendiente {
		t.Errorf("Verificacion = %q, se esperaba pendiente", p.Verificacion)
	}
	if !p.CreadoEn.Equal(ahoraDePrueba) || !p.ActualizadoEn.Equal(ahoraDePrueba) {
		t.Error("las marcas de tiempo debían ser el ahora recibido")
	}
	if p.DadoDeBajaEn != nil {
		t.Error("DadoDeBajaEn debía ser nil")
	}
}

func TestNuevoProfesionalCamposInvalidos(t *testing.T) {
	casos := []struct {
		nombre        string
		mutar         func(*EntradaProfesional)
		campoEsperado string
	}{
		{"nombre vacio", func(entrada *EntradaProfesional) { entrada.Nombre = "   " }, "nombre"},
		{"nombre muy largo", func(entrada *EntradaProfesional) { entrada.Nombre = strings.Repeat("a", 101) }, "nombre"},
		{"apellido vacio", func(entrada *EntradaProfesional) { entrada.Apellido = "" }, "apellido"},
		{"matricula invalida", func(entrada *EntradaProfesional) { entrada.Matricula = "XX 123" }, "matricula"},
		{"especialidad desconocida", func(entrada *EntradaProfesional) { entrada.Especialidad = "cardiologia" }, "especialidad"},
		{"bio muy larga", func(entrada *EntradaProfesional) { entrada.Bio = strings.Repeat("a", 2001) }, "bio"},
		{"precio negativo", func(entrada *EntradaProfesional) { entrada.PrecioConsulta = -1 }, "precioConsultaCentavos"},
		{"sin modalidades", func(entrada *EntradaProfesional) { entrada.Modalidades = nil }, "modalidades"},
		{"modalidad desconocida", func(entrada *EntradaProfesional) { entrada.Modalidades = []string{"online"} }, "modalidades"},
		{"modalidad repetida", func(entrada *EntradaProfesional) { entrada.Modalidades = []string{"presencial", "presencial"} }, "modalidades"},
		{"zona vacia", func(entrada *EntradaProfesional) { entrada.Zona = "" }, "zona"},
		{"obra social repetida", func(entrada *EntradaProfesional) { entrada.ObrasSociales = []string{"OSDE", "osde"} }, "obrasSociales"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := entradaValida()
			caso.mutar(&entrada)

			_, err := NuevoProfesional(entrada, ahoraDePrueba)
			if err == nil {
				t.Fatal("se esperaba un error de validación")
			}

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
			}

			encontrado := false
			for _, f := range verr.Campos {
				if f.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestNuevoProfesionalAcumulaErrores(t *testing.T) {
	entrada := entradaValida()
	entrada.Nombre = ""
	entrada.Matricula = "roto"
	entrada.Zona = ""

	_, err := NuevoProfesional(entrada, ahoraDePrueba)

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
	}
	// el punto de acumular: el cliente corrige los tres de una
	if len(verr.Campos) != 3 {
		t.Errorf("se esperaban 3 campos con error, se obtuvieron %d: %+v", len(verr.Campos), verr.Campos)
	}
}

func TestNuevoProfesionalNormalizaEntrada(t *testing.T) {
	entrada := entradaValida()
	entrada.Nombre = "  Martín  "
	entrada.Especialidad = "  PSICOLOGIA  "
	entrada.Modalidades = []string{" Telemedicina "}

	p, err := NuevoProfesional(entrada, ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Nombre != "Martín" {
		t.Errorf("Nombre = %q, se esperaba sin espacios", p.Nombre)
	}
	if p.Especialidad != EspecialidadPsicologia {
		t.Errorf("Especialidad = %q, se esperaba psicologia", p.Especialidad)
	}
	if len(p.Modalidades) != 1 || p.Modalidades[0] != ModalidadTelemedicina {
		t.Errorf("Modalidades = %v, se esperaba [telemedicina]", p.Modalidades)
	}
}

func TestAplicarCambiosResetaVerificacion(t *testing.T) {
	base, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	base.Verificacion = VerificacionVerificada

	masTarde := ahoraDePrueba.Add(time.Hour)

	t.Run("cambiar la matricula resetea", func(t *testing.T) {
		entrada := entradaValida()
		entrada.Matricula = "MN 11111"

		actualizado, err := base.AplicarCambios(entrada, masTarde)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if actualizado.Verificacion != VerificacionPendiente {
			t.Error("cambiar la matrícula tenía que volver la verificación a pendiente")
		}
	})

	t.Run("cambiar la especialidad resetea", func(t *testing.T) {
		entrada := entradaValida()
		entrada.Especialidad = "odontologia"

		actualizado, err := base.AplicarCambios(entrada, masTarde)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if actualizado.Verificacion != VerificacionPendiente {
			t.Error("cambiar la especialidad tenía que volver la verificación a pendiente")
		}
	})

	t.Run("cambiar la bio no resetea", func(t *testing.T) {
		entrada := entradaValida()
		entrada.Bio = "Otra bio."

		actualizado, err := base.AplicarCambios(entrada, masTarde)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if actualizado.Verificacion != VerificacionVerificada {
			t.Error("editar la bio no tenía por qué tocar la verificación")
		}
	})
}

func TestAplicarCambiosPreservaCamposNoEditables(t *testing.T) {
	base, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	entrada := entradaValida()
	entrada.Nombre = "Otro"
	entrada.Apellido = "Nombre"

	masTarde := ahoraDePrueba.Add(time.Hour)
	actualizado, err := base.AplicarCambios(entrada, masTarde)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if actualizado.ID != base.ID {
		t.Error("el ID no es editable")
	}
	// el slug es una URL pública: regenerarlo al cambiar el nombre rompe
	// enlaces y posicionamiento
	if actualizado.Slug != base.Slug {
		t.Errorf("el slug no debía cambiar: %q → %q", base.Slug, actualizado.Slug)
	}
	if !actualizado.CreadoEn.Equal(base.CreadoEn) {
		t.Error("CreadoEn no es editable")
	}
	if !actualizado.ActualizadoEn.Equal(masTarde) {
		t.Error("ActualizadoEn debía avanzar")
	}
}

func TestDarDeBajaReactivar(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	masTarde := ahoraDePrueba.Add(time.Hour)
	baja := p.DarDeBaja(masTarde)

	if baja.Estado != EstadoInactivo {
		t.Errorf("Estado = %q, se esperaba inactivo", baja.Estado)
	}
	if baja.DadoDeBajaEn == nil || !baja.DadoDeBajaEn.Equal(masTarde) {
		t.Error("DadoDeBajaEn debía sellarse con el momento de la baja")
	}
	// valor receiver: el original no se toca
	if p.Estado != EstadoActivo {
		t.Error("DarDeBaja no debía mutar el receptor")
	}

	// idempotente: dar de baja algo ya dado de baja no es un error ni
	// corre la fecha original
	muchoMasTarde := masTarde.Add(time.Hour)
	otraVez := baja.DarDeBaja(muchoMasTarde)
	if !otraVez.DadoDeBajaEn.Equal(masTarde) {
		t.Error("una segunda baja no debía correr la fecha de la primera")
	}

	reactivado := baja.Reactivar(muchoMasTarde)
	if reactivado.Estado != EstadoActivo {
		t.Errorf("Estado = %q, se esperaba activo", reactivado.Estado)
	}
	if reactivado.DadoDeBajaEn != nil {
		t.Error("reactivar debía limpiar DadoDeBajaEn")
	}
}

func TestClonarEsCopiaProfunda(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	c := p.Clonar()
	c.Modalidades[0] = ModalidadDomicilio
	c.ObrasSociales[0] = "MUTADA"

	if p.Modalidades[0] == ModalidadDomicilio {
		t.Error("mutar el clon alteró las modalidades del original")
	}
	if p.ObrasSociales[0] == "MUTADA" {
		t.Error("mutar el clon alteró las obras sociales del original")
	}

	baja := p.DarDeBaja(ahoraDePrueba)
	clonBaja := baja.Clonar()
	*clonBaja.DadoDeBajaEn = ahoraDePrueba.Add(time.Hour)
	if baja.DadoDeBajaEn.Equal(ahoraDePrueba.Add(time.Hour)) {
		t.Error("mutar el clon alteró el DadoDeBajaEn del original")
	}
}

func TestNombreCompleto(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if obtenido := p.NombreCompleto(); obtenido != "Martín González" {
		t.Errorf("NombreCompleto() = %q, se esperaba %q", obtenido, "Martín González")
	}
}
```

- [ ] **Step 3: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run TestNuevoProfesional -v`
Expected: FAIL con `undefined: EntradaProfesional` y `undefined: NuevoProfesional`

- [ ] **Step 4: Implementar la entidad**

Archivo `apps/api/internal/domain/profesional.go`:

```go
package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxLargoNombre = 100
	maxLargoBio    = 2000
	maxLargoZona   = 100
)

// Profesional es un profesional de la salud dado de alta en la plataforma.
//
// Invariante del paquete: no se puede construir uno inválido desde afuera.
// No hay setters públicos; la única puerta de entrada es NuevoProfesional, y la
// única forma de modificarlo es AplicarCambios, que revalida todo.
type Profesional struct {
	ID             uuid.UUID
	Slug           string
	Nombre         string
	Apellido       string
	Matricula      Matricula
	Especialidad   Especialidad
	Bio            string
	PrecioConsulta Dinero
	Modalidades    []Modalidad
	Zona           string
	ObrasSociales  []string
	Estado         Estado
	Verificacion   EstadoVerificacion
	CreadoEn       time.Time
	ActualizadoEn  time.Time
	DadoDeBajaEn   *time.Time
}

// EntradaProfesional es la entrada cruda, en tipos primitivos. Que sea primitiva
// no es descuido: obliga a que todo el parseo y toda la validación ocurran acá
// adentro, y no repartidos por los handlers.
type EntradaProfesional struct {
	Nombre         string
	Apellido       string
	Matricula      string
	Especialidad   string
	Bio            string
	PrecioConsulta int64
	Modalidades    []string
	Zona           string
	ObrasSociales  []string
}

// NuevoProfesional valida la entrada y devuelve un profesional consistente o un
// ErrorValidacion con todos los campos que fallaron.
func NuevoProfesional(entrada EntradaProfesional, ahora time.Time) (Profesional, error) {
	p, verr := construir(entrada)
	if verr.tieneErrores() {
		return Profesional{}, verr
	}

	p.ID = uuid.New()
	p.Slug = GenerarSlug(p.NombreCompleto())
	if p.Slug == "" {
		// el nombre pasó la validación pero no dejó ningún carácter usable
		// (por ejemplo "..."). Sin esto quedaría un slug vacío y la URL
		// pública del profesional colisionaría con la de cualquier otro.
		p.Slug = p.ID.String()
	}
	p.Estado = EstadoActivo
	p.Verificacion = VerificacionPendiente
	p.CreadoEn = ahora
	p.ActualizadoEn = ahora

	return p, nil
}

// AplicarCambios reemplaza los campos editables y devuelve el resultado sin tocar
// el receptor. ID, Slug, Estado, CreadoEn y DadoDeBajaEn no son editables.
func (p Profesional) AplicarCambios(entrada EntradaProfesional, ahora time.Time) (Profesional, error) {
	actualizado, verr := construir(entrada)
	if verr.tieneErrores() {
		return Profesional{}, verr
	}

	actualizado.ID = p.ID
	actualizado.Slug = p.Slug
	actualizado.Estado = p.Estado
	actualizado.CreadoEn = p.CreadoEn
	actualizado.DadoDeBajaEn = p.DadoDeBajaEn
	actualizado.ActualizadoEn = ahora

	// La verificación se hizo sobre una matrícula y una especialidad
	// concretas. Si cambian, deja de valer: toda orientación, agenda o cobro
	// depende de que el profesional esté verificado.
	if actualizado.Matricula != p.Matricula || actualizado.Especialidad != p.Especialidad {
		actualizado.Verificacion = VerificacionPendiente
	} else {
		actualizado.Verificacion = p.Verificacion
	}

	return actualizado, nil
}

// DarDeBaja da de baja al profesional. No es un borrado: los turnos y
// comprobantes históricos siguen apuntando a este registro. Es idempotente y
// no corre la fecha de la primera baja.
func (p Profesional) DarDeBaja(ahora time.Time) Profesional {
	if p.Estado == EstadoInactivo {
		return p
	}
	p.Estado = EstadoInactivo
	p.DadoDeBajaEn = &ahora
	p.ActualizadoEn = ahora
	return p
}

// Reactivar revierte la baja. Idempotente.
func (p Profesional) Reactivar(ahora time.Time) Profesional {
	if p.Estado == EstadoActivo {
		return p
	}
	p.Estado = EstadoActivo
	p.DadoDeBajaEn = nil
	p.ActualizadoEn = ahora
	return p
}

// Clonar devuelve una copia profunda.
//
// Una copia superficial comparte el array que hay debajo de los slices, y deja
// que quien la reciba mute el original desde afuera sin enterarse. Es el bug
// número uno de un repositorio en memoria.
func (p Profesional) Clonar() Profesional {
	c := p
	c.Modalidades = slices.Clone(p.Modalidades)
	c.ObrasSociales = slices.Clone(p.ObrasSociales)
	if p.DadoDeBajaEn != nil {
		t := *p.DadoDeBajaEn
		c.DadoDeBajaEn = &t
	}
	return c
}

func (p Profesional) NombreCompleto() string {
	return p.Nombre + " " + p.Apellido
}

// construir parsea y valida la entrada, acumulando todos los errores. Es la única
// implementación de las reglas: la comparten el alta y la edición.
func construir(entrada EntradaProfesional) (Profesional, ErrorValidacion) {
	var p Profesional
	var verr ErrorValidacion

	p.Nombre = validarNombre(entrada.Nombre, "nombre", &verr)
	p.Apellido = validarNombre(entrada.Apellido, "apellido", &verr)

	if m, err := ParsearMatricula(entrada.Matricula); err != nil {
		verr.agregar("matricula", err.Error())
	} else {
		p.Matricula = m
	}

	esp := Especialidad(strings.ToLower(strings.TrimSpace(entrada.Especialidad)))
	if !esp.EsValida() {
		verr.agregar("especialidad", "debe ser psicologia, kinesiologia u odontologia")
	} else {
		p.Especialidad = esp
	}

	p.Bio = strings.TrimSpace(entrada.Bio)
	if utf8.RuneCountInString(p.Bio) > maxLargoBio {
		verr.agregar("bio", fmt.Sprintf("no puede superar los %d caracteres", maxLargoBio))
	}

	if entrada.PrecioConsulta < 0 {
		verr.agregar("precioConsultaCentavos", "no puede ser negativo")
	} else {
		p.PrecioConsulta = Dinero(entrada.PrecioConsulta)
	}

	p.Modalidades = construirModalidades(entrada.Modalidades, &verr)

	p.Zona = strings.TrimSpace(entrada.Zona)
	switch {
	case p.Zona == "":
		verr.agregar("zona", "es obligatoria")
	case utf8.RuneCountInString(p.Zona) > maxLargoZona:
		verr.agregar("zona", fmt.Sprintf("no puede superar los %d caracteres", maxLargoZona))
	}

	p.ObrasSociales = construirObrasSociales(entrada.ObrasSociales, &verr)

	return p, verr
}

func validarNombre(crudo, campo string, verr *ErrorValidacion) string {
	nombre := strings.TrimSpace(crudo)
	switch {
	case nombre == "":
		verr.agregar(campo, "es obligatorio")
	case utf8.RuneCountInString(nombre) > maxLargoNombre:
		verr.agregar(campo, fmt.Sprintf("no puede superar los %d caracteres", maxLargoNombre))
	}
	return nombre
}

func construirModalidades(crudo []string, verr *ErrorValidacion) []Modalidad {
	if len(crudo) == 0 {
		verr.agregar("modalidades", "se requiere al menos una")
		return nil
	}

	visto := make(map[Modalidad]bool, len(crudo))
	salida := make([]Modalidad, 0, len(crudo))

	for _, r := range crudo {
		m := Modalidad(strings.ToLower(strings.TrimSpace(r)))
		switch {
		case !m.EsValida():
			verr.agregar("modalidades", fmt.Sprintf("%q no es una modalidad válida", r))
		case visto[m]:
			verr.agregar("modalidades", fmt.Sprintf("%q está repetida", r))
		default:
			visto[m] = true
			salida = append(salida, m)
		}
	}
	return salida
}

func construirObrasSociales(crudo []string, verr *ErrorValidacion) []string {
	// puede estar vacía: un profesional que solo atiende privado es válido
	visto := make(map[string]bool, len(crudo))
	salida := make([]string, 0, len(crudo))

	for _, r := range crudo {
		v := strings.TrimSpace(r)
		if v == "" {
			continue
		}
		// "OSDE" y "osde" son la misma obra social
		clave := Normalizar(v)
		if visto[clave] {
			verr.agregar("obrasSociales", fmt.Sprintf("%q está repetida", v))
			continue
		}
		visto[clave] = true
		salida = append(salida, v)
	}
	return salida
}
```

- [ ] **Step 5: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -v`
Expected: PASS en todos. Prestar atención a `TestClonarEsCopiaProfunda`: si falla, `Clonar` está haciendo copia superficial y todo el repositorio en memoria de la Task 5 va a estar roto.

- [ ] **Step 6: Verificar que el dominio no importa nada del proyecto**

Run: `cd apps/api && go list -deps ./internal/domain | grep "joaquinfochoa"`
Expected: una sola línea, `github.com/joaquinfochoa/Salud/apps/api/internal/domain`. Cualquier otra línea significa que el dominio depende de otra capa y la arquitectura se rompió.

- [ ] **Step 7: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): entidad Profesional con validación acumulativa

No se puede construir un Profesional inválido desde fuera del paquete.
ErrorValidacion junta todos los campos que fallan de una pasada.

Cambiar la matrícula o la especialidad devuelve la verificación a
pendiente: la verificación se hizo sobre esos datos y deja de valer si
cambian.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 5: Interfaz del repositorio e implementación en memoria

**Files:**
- Create: `apps/api/internal/repository/profesional.go`
- Create: `apps/api/internal/repository/memory/profesional.go`
- Test: `apps/api/internal/repository/memory/profesional_test.go`

**Interfaces:**
- Consumes: todo el paquete `domain` (Tasks 2-4)
- Produces:
  - `repository.Filtro{Especialidad *domain.Especialidad; Zona *string; Estado *domain.Estado; Busqueda *string; Limite int; Desplazamiento int}`
  - `repository.Profesional` — la interfaz con `Crear`, `ObtenerPorID`, `ObtenerPorSlug`, `ObtenerPorMatricula`, `Listar`, `Actualizar`
  - `func memory.NuevoProfesional() *memory.Profesional` — implementa `repository.Profesional`

- [ ] **Step 1: Escribir la interfaz**

No lleva test: es una declaración de tipos. El test de que la implementación la cumple es la aserción de compilación del paso 4.

Archivo `apps/api/internal/repository/profesional.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Filtro son los criterios del listado. Los punteros distinguen "no filtrar
// por este campo" de "filtrar por el valor cero", que con valores planos no se
// puede: un Estado("") sería indistinguible de "sin filtro".
type Filtro struct {
	Especialidad   *domain.Especialidad
	Zona           *string
	Estado         *domain.Estado
	Busqueda       *string // busca en nombre y apellido, sin distinguir acentos
	Limite         int
	Desplazamiento int
}

// Profesional es el punto de cambio a PostgreSQL. Cuando exista la
// implementación con base de datos, migrar es cambiar una línea de main.go y
// nada más.
//
// Todos los métodos reciben context.Context aunque la implementación en
// memoria lo ignore: agregarlo después obligaría a tocar todas las firmas.
//
// No hay borrado físico: la baja es lógica y se hace con Actualizar.
type Profesional interface {
	Crear(ctx context.Context, p domain.Profesional) error
	ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Profesional, error)
	ObtenerPorSlug(ctx context.Context, slug string) (domain.Profesional, error)
	ObtenerPorMatricula(ctx context.Context, m domain.Matricula) (domain.Profesional, error)
	Listar(ctx context.Context, f Filtro) ([]domain.Profesional, int, error)
	Actualizar(ctx context.Context, p domain.Profesional) error
}
```

- [ ] **Step 2: Escribir los tests del repositorio en memoria**

Archivo `apps/api/internal/repository/memory/profesional_test.go`:

```go
package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var ahoraDePrueba = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

func hacerProfesional(t *testing.T, nombre, apellido, matricula string, esp domain.Especialidad, zona string) domain.Profesional {
	t.Helper()
	p, err := domain.NuevoProfesional(domain.EntradaProfesional{
		Nombre:         nombre,
		Apellido:       apellido,
		Matricula:      matricula,
		Especialidad:   string(esp),
		Bio:            "bio",
		PrecioConsulta: 1000000,
		Modalidades:    []string{"telemedicina"},
		Zona:           zona,
		ObrasSociales:  []string{"OSDE"},
	}, ahoraDePrueba)
	if err != nil {
		t.Fatalf("no se pudo construir el profesional de prueba: %v", err)
	}
	return p
}

func TestCrearYObtenerPorID(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")

	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenido, err := repo.ObtenerPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID devolvió error: %v", err)
	}
	if obtenido.ID != p.ID || obtenido.Slug != p.Slug {
		t.Errorf("el profesional recuperado no coincide: %+v", obtenido)
	}
}

func TestObtenerPorIDNoExiste(t *testing.T) {
	_, err := NuevoProfesional().ObtenerPorID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestObtenerPorSlugYPorMatricula(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	porSlug, err := repo.ObtenerPorSlug(ctx, "martin-gonzalez")
	if err != nil {
		t.Fatalf("ObtenerPorSlug devolvió error: %v", err)
	}
	if porSlug.ID != p.ID {
		t.Error("ObtenerPorSlug devolvió otro profesional")
	}

	porMatricula, err := repo.ObtenerPorMatricula(ctx, p.Matricula)
	if err != nil {
		t.Fatalf("ObtenerPorMatricula devolvió error: %v", err)
	}
	if porMatricula.ID != p.ID {
		t.Error("ObtenerPorMatricula devolvió otro profesional")
	}

	if _, err := repo.ObtenerPorSlug(ctx, "no-existe"); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El test más importante de este paquete. Si el repositorio devuelve algo que
// comparte memoria con lo que guardó, quien lo reciba puede mutar el store
// desde afuera sin enterarse.
func TestElStoreDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	// mutar lo que devolvió el repositorio
	obtenido, _ := repo.ObtenerPorID(ctx, p.ID)
	obtenido.Nombre = "MUTADO"
	obtenido.Modalidades[0] = domain.ModalidadDomicilio
	obtenido.ObrasSociales[0] = "MUTADA"

	recargado, _ := repo.ObtenerPorID(ctx, p.ID)
	if recargado.Nombre == "MUTADO" {
		t.Error("mutar el resultado alteró el store")
	}
	if recargado.Modalidades[0] == domain.ModalidadDomicilio {
		t.Error("las modalidades comparten memoria con el store")
	}
	if recargado.ObrasSociales[0] == "MUTADA" {
		t.Error("las obras sociales comparten memoria con el store")
	}

	// y al revés: mutar lo que se pasó a Crear tampoco debe afectar
	p.Modalidades[0] = domain.ModalidadPresencial
	recargado2, _ := repo.ObtenerPorID(ctx, p.ID)
	if recargado2.Modalidades[0] == domain.ModalidadPresencial {
		t.Error("Crear guardó una referencia en vez de una copia")
	}
}

func TestActualizar(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	p.Zona = "GBA Norte"
	if err := repo.Actualizar(ctx, p); err != nil {
		t.Fatalf("Actualizar devolvió error: %v", err)
	}

	obtenido, _ := repo.ObtenerPorID(ctx, p.ID)
	if obtenido.Zona != "GBA Norte" {
		t.Errorf("Zona = %q, se esperaba %q", obtenido.Zona, "GBA Norte")
	}

	desconocido := hacerProfesional(t, "Otro", "Nombre", "MN 11111", domain.EspecialidadOdontologia, "CABA")
	if err := repo.Actualizar(ctx, desconocido); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado al actualizar algo inexistente, se obtuvo %v", err)
	}
}

func TestListarFiltros(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	psico := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	kine := hacerProfesional(t, "Pablo", "Moreno", "MN 45321", domain.EspecialidadKinesiologia, "CABA")
	odonto := hacerProfesional(t, "Gabriela", "Ríos", "MN 67890", domain.EspecialidadOdontologia, "GBA Norte")
	// las fechas distintas fuerzan un orden determinista
	kine.CreadoEn = ahoraDePrueba.Add(time.Minute)
	odonto.CreadoEn = ahoraDePrueba.Add(2 * time.Minute)

	for _, p := range []domain.Profesional{psico, kine, odonto} {
		if err := repo.Crear(ctx, p); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	t.Run("sin filtros devuelve todo", func(t *testing.T) {
		obtenido, total, err := repo.Listar(ctx, repository.Filtro{Limite: 10})
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if total != 3 || len(obtenido) != 3 {
			t.Errorf("total=%d len=%d, se esperaba 3 y 3", total, len(obtenido))
		}
	})

	t.Run("por especialidad", func(t *testing.T) {
		esp := domain.EspecialidadPsicologia
		obtenido, total, _ := repo.Listar(ctx, repository.Filtro{Especialidad: &esp, Limite: 10})
		if total != 1 || obtenido[0].ID != psico.ID {
			t.Errorf("se esperaba solo el psicólogo, se obtuvo total=%d", total)
		}
	})

	t.Run("por zona sin distinguir acentos", func(t *testing.T) {
		zona := "caba"
		_, total, _ := repo.Listar(ctx, repository.Filtro{Zona: &zona, Limite: 10})
		if total != 2 {
			t.Errorf("total = %d, se esperaban 2 en CABA", total)
		}
	})

	t.Run("busqueda sin acentos", func(t *testing.T) {
		// buscar "gonzalez" tiene que encontrar a "González"
		q := "gonzalez"
		obtenido, total, _ := repo.Listar(ctx, repository.Filtro{Busqueda: &q, Limite: 10})
		if total != 1 || obtenido[0].ID != psico.ID {
			t.Errorf("la búsqueda sin acentos no encontró a González: total=%d", total)
		}
	})

	t.Run("busqueda por nombre parcial", func(t *testing.T) {
		q := "pab"
		_, total, _ := repo.Listar(ctx, repository.Filtro{Busqueda: &q, Limite: 10})
		if total != 1 {
			t.Errorf("total = %d, se esperaba 1", total)
		}
	})

	t.Run("por estado", func(t *testing.T) {
		baja := psico.DarDeBaja(ahoraDePrueba.Add(time.Hour))
		if err := repo.Actualizar(ctx, baja); err != nil {
			t.Fatalf("Actualizar devolvió error: %v", err)
		}

		inactivo := domain.EstadoInactivo
		_, total, _ := repo.Listar(ctx, repository.Filtro{Estado: &inactivo, Limite: 10})
		if total != 1 {
			t.Errorf("total = %d, se esperaba 1 inactivo", total)
		}

		activo := domain.EstadoActivo
		_, total, _ = repo.Listar(ctx, repository.Filtro{Estado: &activo, Limite: 10})
		if total != 2 {
			t.Errorf("total = %d, se esperaban 2 activos", total)
		}
	})
}

func TestListarPaginacionEsEstable(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	for i := range 10 {
		p := hacerProfesional(t, fmt.Sprintf("Nombre%d", i), "Apellido",
			fmt.Sprintf("MN %d", 10000+i), domain.EspecialidadPsicologia, "CABA")
		p.CreadoEn = ahoraDePrueba.Add(time.Duration(i) * time.Minute)
		if err := repo.Crear(ctx, p); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	pagina1, total, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 0})
	if total != 10 || len(pagina1) != 3 {
		t.Fatalf("total=%d len=%d, se esperaba 10 y 3", total, len(pagina1))
	}

	pagina2, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 3})
	if len(pagina2) != 3 {
		t.Fatalf("len(pagina2) = %d, se esperaba 3", len(pagina2))
	}

	// el mapa de Go itera en orden aleatorio: sin ordenar, dos llamadas
	// idénticas devolverían páginas distintas y la paginación no serviría
	for range 5 {
		otraVez, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 0})
		for i := range pagina1 {
			if otraVez[i].ID != pagina1[i].ID {
				t.Fatal("dos llamadas idénticas devolvieron órdenes distintos")
			}
		}
	}

	// las páginas no se solapan
	for _, a := range pagina1 {
		for _, b := range pagina2 {
			if a.ID == b.ID {
				t.Error("la página 1 y la 2 comparten un elemento")
			}
		}
	}

	ultima, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 9})
	if len(ultima) != 1 {
		t.Errorf("la última página tenía %d elementos, se esperaba 1", len(ultima))
	}

	vacio, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 50})
	if len(vacio) != 0 {
		t.Errorf("un desplazamiento más allá del total debía devolver vacío, devolvió %d", len(vacio))
	}
}

func TestAccesoConcurrente(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := hacerProfesional(t, fmt.Sprintf("N%d", i), "A",
				fmt.Sprintf("MN %d", 20000+i), domain.EspecialidadPsicologia, "CABA")
			_ = repo.Crear(ctx, p)
			_, _, _ = repo.Listar(ctx, repository.Filtro{Limite: 10})
			_, _ = repo.ObtenerPorID(ctx, p.ID)
		}(i)
	}
	wg.Wait()

	_, total, _ := repo.Listar(ctx, repository.Filtro{Limite: 100})
	if total != 50 {
		t.Errorf("total = %d, se esperaban 50", total)
	}
}
```

- [ ] **Step 3: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/repository/... -v`
Expected: FAIL con `undefined: NuevoProfesional`

- [ ] **Step 4: Implementar el repositorio en memoria**

Archivo `apps/api/internal/repository/memory/profesional.go`:

```go
package memory

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// aserción de compilación: si la implementación deja de cumplir la interfaz,
// el error aparece acá y no en main.go
var _ repository.Profesional = (*Profesional)(nil)

// Profesional guarda los profesionales en memoria. Se pierde todo al
// reiniciar, y está bien: sirve para definir el dominio antes de comprometerse
// con un esquema de base de datos, y para correr los casos sin infraestructura.
type Profesional struct {
	mu    sync.RWMutex
	datos map[uuid.UUID]domain.Profesional
}

func NuevoProfesional() *Profesional {
	return &Profesional{datos: make(map[uuid.UUID]domain.Profesional)}
}

func (r *Profesional) Crear(_ context.Context, p domain.Profesional) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[p.ID]; existe {
		return fmt.Errorf("ya existe un profesional con id %s", p.ID)
	}
	r.datos[p.ID] = p.Clonar()
	return nil
}

func (r *Profesional) ObtenerPorID(_ context.Context, id uuid.UUID) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.datos[id]
	if !ok {
		return domain.Profesional{}, domain.ErrNoEncontrado
	}
	return p.Clonar(), nil
}

func (r *Profesional) ObtenerPorSlug(_ context.Context, slug string) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.datos {
		if p.Slug == slug {
			return p.Clonar(), nil
		}
	}
	return domain.Profesional{}, domain.ErrNoEncontrado
}

func (r *Profesional) ObtenerPorMatricula(_ context.Context, m domain.Matricula) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.datos {
		if p.Matricula == m {
			return p.Clonar(), nil
		}
	}
	return domain.Profesional{}, domain.ErrNoEncontrado
}

func (r *Profesional) Listar(_ context.Context, f repository.Filtro) ([]domain.Profesional, int, error) {
	// El slice de Go entra en pánico con un índice negativo, a diferencia de un
	// OFFSET de SQL. El servicio ya normaliza estos valores, pero el que corta
	// el slice es este método, así que la guarda va acá.
	if f.Desplazamiento < 0 {
		f.Desplazamiento = 0
	}
	if f.Limite < 0 {
		f.Limite = 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// ponytail: scan O(n), correcto para un store en memoria. La
	// implementación Postgres resuelve esto con índices sobre especialidad,
	// zona y estado.
	coincidentes := make([]domain.Profesional, 0, len(r.datos))
	for _, p := range r.datos {
		if coincide(p, f) {
			coincidentes = append(coincidentes, p.Clonar())
		}
	}

	// El mapa de Go itera en orden aleatorio. Sin este orden, dos llamadas
	// idénticas devolverían páginas distintas y la paginación sería inútil.
	slices.SortFunc(coincidentes, func(a, b domain.Profesional) int {
		if c := a.CreadoEn.Compare(b.CreadoEn); c != 0 {
			return c
		}
		return strings.Compare(a.ID.String(), b.ID.String())
	})

	total := len(coincidentes)
	if f.Desplazamiento >= total {
		return []domain.Profesional{}, total, nil
	}

	fin := total
	if f.Limite > 0 && f.Desplazamiento+f.Limite < total {
		fin = f.Desplazamiento + f.Limite
	}
	return coincidentes[f.Desplazamiento:fin], total, nil
}

func (r *Profesional) Actualizar(_ context.Context, p domain.Profesional) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[p.ID]; !existe {
		return domain.ErrNoEncontrado
	}
	r.datos[p.ID] = p.Clonar()
	return nil
}

func coincide(p domain.Profesional, f repository.Filtro) bool {
	if f.Especialidad != nil && p.Especialidad != *f.Especialidad {
		return false
	}
	if f.Estado != nil && p.Estado != *f.Estado {
		return false
	}
	if f.Zona != nil && domain.Normalizar(p.Zona) != domain.Normalizar(*f.Zona) {
		return false
	}
	if f.Busqueda != nil {
		q := domain.Normalizar(*f.Busqueda)
		if q != "" && !strings.Contains(domain.Normalizar(p.NombreCompleto()), q) {
			return false
		}
	}
	return true
}
```

- [ ] **Step 5: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/repository/... -v`
Expected: PASS en todos.

- [ ] **Step 6: Correr el detector de carreras**

Run: `cd apps/api && go test ./internal/repository/... -race`
Expected: PASS sin `WARNING: DATA RACE`. Si aparece una carrera, falta un `Lock` o hay un `RLock` donde debería haber `Lock`.

- [ ] **Step 7: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/repository/
git commit -m "feat(repository): interfaz y repositorio en memoria

La interfaz es el punto de cambio a PostgreSQL: migrar es cambiar una
línea de main.go. Recibe context.Context desde el día 1 aunque la
implementación en memoria lo ignore.

El store guarda y devuelve copias profundas: una copia superficial deja
que quien la recibe mute el store desde afuera. El listado ordena antes
de paginar porque el mapa de Go itera en orden aleatorio.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 6: El servicio — alta y lecturas

**Files:**
- Create: `apps/api/internal/service/profesional.go`
- Test: `apps/api/internal/service/profesional_test.go`

**Interfaces:**
- Consumes: `domain` (Tasks 2-4), `repository.Profesional` y `repository.Filtro` (Task 5), `memory.NuevoProfesional` (Task 5, solo en tests)
- Produces:
  - `func service.NuevoProfesional(repository.Profesional) *service.Profesional`
  - `func (*service.Profesional) Crear(context.Context, domain.EntradaProfesional) (domain.Profesional, error)`
  - `func (*service.Profesional) ObtenerPorID(context.Context, uuid.UUID) (domain.Profesional, error)`
  - `func (*service.Profesional) ObtenerPorSlug(context.Context, string) (domain.Profesional, error)`
  - `func (*service.Profesional) Listar(context.Context, repository.Filtro) ([]domain.Profesional, int, error)`
  - Constantes `service.LimitePorDefecto = 20`, `service.LimiteMaximo = 100`

- [ ] **Step 1: Escribir los tests del alta y las lecturas**

Archivo `apps/api/internal/service/profesional_test.go`:

```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

// No hay mocks en este proyecto. El repositorio en memoria es rápido y
// determinista, así que es el doble de test: se prueba contra la
// implementación de verdad. Si un test pareciera necesitar un mock, la
// frontera está mal dibujada.
func nuevoServicioDePrueba() *Profesional {
	return NuevoProfesional(memory.NuevoProfesional())
}

func entradaValida() domain.EntradaProfesional {
	return domain.EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico.",
		PrecioConsulta: 1200000,
		Modalidades:    []string{"telemedicina"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE"},
	}
}

func TestCrear(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}
	if p.Slug != "martin-gonzalez" {
		t.Errorf("Slug = %q, se esperaba %q", p.Slug, "martin-gonzalez")
	}
	if p.Estado != domain.EstadoActivo || p.Verificacion != domain.VerificacionPendiente {
		t.Error("un profesional nuevo nace activo y sin verificar")
	}

	// tiene que quedar realmente guardado
	obtenido, err := svc.ObtenerPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID devolvió error: %v", err)
	}
	if obtenido.ID != p.ID {
		t.Error("el profesional no quedó persistido")
	}
}

func TestCrearValidacion(t *testing.T) {
	entrada := entradaValida()
	entrada.Nombre = ""

	_, err := nuevoServicioDePrueba().Crear(context.Background(), entrada)

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
}

func TestCrearMatriculaDuplicada(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	if _, err := svc.Crear(ctx, entradaValida()); err != nil {
		t.Fatalf("el primer alta falló: %v", err)
	}

	// la misma matrícula escrita distinto sigue siendo la misma matrícula
	otro := entradaValida()
	otro.Nombre = "Otro"
	otro.Apellido = "Profesional"
	otro.Matricula = "m.n. 98234"

	_, err := svc.Crear(ctx, otro)
	if !errors.Is(err, domain.ErrMatriculaEnUso) {
		t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
	}
}

func TestCrearSlugUnico(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	// tres homónimos: dos "Martín González" son perfectamente posibles y no
	// pueden ser un error para el cliente
	primero, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	entrada2 := entradaValida()
	entrada2.Matricula = "MN 11111"
	segundo, err := svc.Crear(ctx, entrada2)
	if err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	entrada3 := entradaValida()
	entrada3.Matricula = "MN 22222"
	tercero, err := svc.Crear(ctx, entrada3)
	if err != nil {
		t.Fatalf("alta 3 falló: %v", err)
	}

	if primero.Slug != "martin-gonzalez" {
		t.Errorf("slug 1 = %q", primero.Slug)
	}
	if segundo.Slug != "martin-gonzalez-2" {
		t.Errorf("slug 2 = %q, se esperaba martin-gonzalez-2", segundo.Slug)
	}
	if tercero.Slug != "martin-gonzalez-3" {
		t.Errorf("slug 3 = %q, se esperaba martin-gonzalez-3", tercero.Slug)
	}
}

func TestObtenerPorSlug(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenido, err := svc.ObtenerPorSlug(ctx, p.Slug)
	if err != nil {
		t.Fatalf("ObtenerPorSlug devolvió error: %v", err)
	}
	if obtenido.ID != p.ID {
		t.Error("ObtenerPorSlug devolvió otro profesional")
	}

	if _, err := svc.ObtenerPorSlug(ctx, "no-existe"); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestObtenerPorIDNoExiste(t *testing.T) {
	_, err := nuevoServicioDePrueba().ObtenerPorID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestListarDefaults(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	for i := range 3 {
		entrada := entradaValida()
		entrada.Matricula = "MN 3000" + string(rune('0'+i))
		if _, err := svc.Crear(ctx, entrada); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	t.Run("limite por defecto", func(t *testing.T) {
		f := repository.Filtro{}
		_, total, err := svc.Listar(ctx, f)
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if total != 3 {
			t.Errorf("total = %d, se esperaba 3", total)
		}
	})

	t.Run("limite recortado al maximo", func(t *testing.T) {
		obtenido, _, err := svc.Listar(ctx, repository.Filtro{Limite: 5000})
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if len(obtenido) > LimiteMaximo {
			t.Errorf("devolvió %d elementos, el máximo es %d", len(obtenido), LimiteMaximo)
		}
	})

	t.Run("desplazamiento negativo se normaliza", func(t *testing.T) {
		obtenido, _, err := svc.Listar(ctx, repository.Filtro{Desplazamiento: -10})
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if len(obtenido) != 3 {
			t.Errorf("len = %d, se esperaba 3", len(obtenido))
		}
	})
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/service/ -v`
Expected: FAIL con `undefined: NuevoProfesional`

- [ ] **Step 3: Implementar el servicio con alta y lecturas**

Archivo `apps/api/internal/service/profesional.go`:

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

const (
	// LimitePorDefecto es cuántos profesionales devuelve el listado si el cliente
	// no pide un tamaño.
	LimitePorDefecto = 20

	// LimiteMaximo es el techo. Sin techo, un cliente puede pedir el padrón
	// entero en una llamada.
	LimiteMaximo = 100
)

// Profesional resuelve los casos de uso que necesitan mirar más de un
// profesional a la vez. Las reglas que se deciden con una sola entidad viven
// en el dominio, no acá.
type Profesional struct {
	repo repository.Profesional

	// ahora es inyectable para que los casos no dependan del reloj.
	ahora func() time.Time
}

func NuevoProfesional(repo repository.Profesional) *Profesional {
	return &Profesional{
		repo:  repo,
		ahora: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Profesional) Crear(ctx context.Context, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	p, err := domain.NuevoProfesional(entrada, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}

	// La matrícula es la única identidad real de una persona en este sistema.
	// El parser ya normalizó "M.N. 98.234" y "MN 98234" a lo mismo, así que
	// esta comparación atrapa los duplicados escritos distinto.
	if err := s.verificarMatriculaLibre(ctx, p.Matricula, uuid.Nil); err != nil {
		return domain.Profesional{}, err
	}

	slug, err := s.slugUnico(ctx, p.Slug)
	if err != nil {
		return domain.Profesional{}, err
	}
	p.Slug = slug

	if err := s.repo.Crear(ctx, p); err != nil {
		return domain.Profesional{}, err
	}
	return p, nil
}

func (s *Profesional) ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Profesional, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *Profesional) ObtenerPorSlug(ctx context.Context, slug string) (domain.Profesional, error) {
	return s.repo.ObtenerPorSlug(ctx, slug)
}

func (s *Profesional) Listar(ctx context.Context, f repository.Filtro) ([]domain.Profesional, int, error) {
	if f.Limite <= 0 {
		f.Limite = LimitePorDefecto
	}
	if f.Limite > LimiteMaximo {
		f.Limite = LimiteMaximo
	}
	if f.Desplazamiento < 0 {
		f.Desplazamiento = 0
	}
	return s.repo.Listar(ctx, f)
}

// verificarMatriculaLibre falla si otro profesional ya tiene esa matrícula.
// excluir permite ignorar al propio profesional durante una edición.
func (s *Profesional) verificarMatriculaLibre(ctx context.Context, m domain.Matricula, excluir uuid.UUID) error {
	existente, err := s.repo.ObtenerPorMatricula(ctx, m)
	switch {
	case errors.Is(err, domain.ErrNoEncontrado):
		return nil
	case err != nil:
		return err
	case existente.ID == excluir:
		return nil
	default:
		return domain.ErrMatriculaEnUso
	}
}

// slugUnico resuelve las colisiones agregando un sufijo numérico.
//
// Nunca es un error para el cliente: dos "Martín González" son perfectamente
// posibles y no hay razón para rechazar al segundo.
func (s *Profesional) slugUnico(ctx context.Context, base string) (string, error) {
	candidato := base
	for i := 2; ; i++ {
		_, err := s.repo.ObtenerPorSlug(ctx, candidato)
		if errors.Is(err, domain.ErrNoEncontrado) {
			return candidato, nil
		}
		if err != nil {
			return "", err
		}
		candidato = fmt.Sprintf("%s-%d", base, i)
	}
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/service/ -v`
Expected: PASS. `TestCrearSlugUnico` es el que confirma la cadena `martin-gonzalez`, `-2`, `-3`.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/service/
git commit -m "feat(service): alta y lecturas de Profesional

La matrícula es única: el parser normaliza antes de comparar, así que
'M.N. 98.234' y 'MN 98234' colisionan como corresponde.

Las colisiones de slug se resuelven con sufijo y nunca son un error
para el cliente: dos homónimos son perfectamente posibles.

Sin mocks: los tests corren contra el repositorio en memoria real.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 7: El servicio — edición, baja y reactivación

**Files:**
- Modify: `apps/api/internal/service/profesional.go` (agregar tres métodos)
- Modify: `apps/api/internal/service/profesional_test.go` (agregar los tests)

**Interfaces:**
- Consumes: todo lo de la Task 6
- Produces:
  - `func (*service.Profesional) Actualizar(context.Context, uuid.UUID, domain.EntradaProfesional) (domain.Profesional, error)`
  - `func (*service.Profesional) DarDeBaja(context.Context, uuid.UUID) error`
  - `func (*service.Profesional) Reactivar(context.Context, uuid.UUID) (domain.Profesional, error)`

- [ ] **Step 1: Escribir los tests**

Agregar al final de `apps/api/internal/service/profesional_test.go`:

```go
func TestActualizar(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	entrada := entradaValida()
	entrada.Bio = "Bio actualizada."
	entrada.Zona = "GBA Norte"

	actualizado, err := svc.Actualizar(ctx, p.ID, entrada)
	if err != nil {
		t.Fatalf("Actualizar devolvió error: %v", err)
	}
	if actualizado.Bio != "Bio actualizada." || actualizado.Zona != "GBA Norte" {
		t.Error("los campos editables no se aplicaron")
	}
	if actualizado.Slug != p.Slug {
		t.Error("el slug es una URL pública y no debía cambiar")
	}

	// tiene que haber quedado persistido
	obtenido, _ := svc.ObtenerPorID(ctx, p.ID)
	if obtenido.Zona != "GBA Norte" {
		t.Error("el cambio no quedó guardado")
	}
}

func TestActualizarNoExiste(t *testing.T) {
	_, err := nuevoServicioDePrueba().Actualizar(context.Background(), uuid.New(), entradaValida())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestActualizarMatriculaDeOtro(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	primero, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	entrada2 := entradaValida()
	entrada2.Nombre = "Carolina"
	entrada2.Apellido = "Vega"
	entrada2.Matricula = "MN 11111"
	if _, err := svc.Crear(ctx, entrada2); err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	// el primero intenta quedarse con la matrícula del segundo
	robo := entradaValida()
	robo.Matricula = "MN 11111"
	if _, err := svc.Actualizar(ctx, primero.ID, robo); !errors.Is(err, domain.ErrMatriculaEnUso) {
		t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
	}
}

func TestActualizarConservaLaPropiaMatricula(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	// editar sin cambiar la matrícula no puede chocar consigo mismo
	entrada := entradaValida()
	entrada.Bio = "Otra bio."
	if _, err := svc.Actualizar(ctx, p.ID, entrada); err != nil {
		t.Errorf("editar conservando la matrícula propia falló: %v", err)
	}
}

func TestActualizarResetaVerificacion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProfesional()
	svc := NuevoProfesional(repo)

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	// simular que ya fue verificado
	p.Verificacion = domain.VerificacionVerificada
	if err := repo.Actualizar(ctx, p); err != nil {
		t.Fatalf("no se pudo preparar el estado: %v", err)
	}

	entrada := entradaValida()
	entrada.Matricula = "MN 55555"
	actualizado, err := svc.Actualizar(ctx, p.ID, entrada)
	if err != nil {
		t.Fatalf("Actualizar devolvió error: %v", err)
	}
	if actualizado.Verificacion != domain.VerificacionPendiente {
		t.Error("cambiar la matrícula tenía que invalidar la verificación")
	}
}

func TestDarDeBajaYReactivar(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	if err := svc.DarDeBaja(ctx, p.ID); err != nil {
		t.Fatalf("DarDeBaja devolvió error: %v", err)
	}

	// el registro sigue existiendo: un turno pasado apunta acá
	obtenido, err := svc.ObtenerPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("el profesional dado de baja tenía que seguir existiendo: %v", err)
	}
	if obtenido.Estado != domain.EstadoInactivo {
		t.Errorf("Estado = %q, se esperaba inactivo", obtenido.Estado)
	}
	if obtenido.DadoDeBajaEn == nil {
		t.Error("DadoDeBajaEn debía sellarse")
	}

	// pero no aparece en el listado por defecto
	_, total, _ := svc.Listar(ctx, repository.Filtro{})
	if total != 0 {
		t.Errorf("el listado por defecto devolvió %d, se esperaba 0", total)
	}

	// filtrando explícitamente sí aparece
	inactivo := domain.EstadoInactivo
	_, total, _ = svc.Listar(ctx, repository.Filtro{Estado: &inactivo})
	if total != 1 {
		t.Errorf("el listado de inactivos devolvió %d, se esperaba 1", total)
	}

	// dar de baja dos veces es idempotente, no un error
	if err := svc.DarDeBaja(ctx, p.ID); err != nil {
		t.Errorf("la segunda baja debía ser idempotente, devolvió %v", err)
	}

	reactivado, err := svc.Reactivar(ctx, p.ID)
	if err != nil {
		t.Fatalf("Reactivar devolvió error: %v", err)
	}
	if reactivado.Estado != domain.EstadoActivo || reactivado.DadoDeBajaEn != nil {
		t.Error("reactivar debía dejarlo activo y limpiar DadoDeBajaEn")
	}

	_, total, _ = svc.Listar(ctx, repository.Filtro{})
	if total != 1 {
		t.Errorf("después de reactivar el listado devolvió %d, se esperaba 1", total)
	}
}

func TestDarDeBajaNoExiste(t *testing.T) {
	if err := nuevoServicioDePrueba().DarDeBaja(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}
```

Nota: el `Listar` por defecto todavía no filtra por activos. Ese comportamiento se agrega en el paso 3 y es lo que hace pasar `TestDarDeBajaYReactivar`.

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/service/ -run 'TestActualizar|TestDarDeBaja' -v`
Expected: FAIL con `svc.Actualizar undefined` y `svc.DarDeBaja undefined`

- [ ] **Step 3: Implementar los tres métodos y el filtro por defecto**

En `apps/api/internal/service/profesional.go`, reemplazar el método `Listar` por esta versión:

```go
func (s *Profesional) Listar(ctx context.Context, f repository.Filtro) ([]domain.Profesional, int, error) {
	if f.Limite <= 0 {
		f.Limite = LimitePorDefecto
	}
	if f.Limite > LimiteMaximo {
		f.Limite = LimiteMaximo
	}
	if f.Desplazamiento < 0 {
		f.Desplazamiento = 0
	}
	// Por defecto solo los activos: un profesional dado de baja no tiene que
	// aparecer en una búsqueda de pacientes. Para verlos hay que pedirlos.
	if f.Estado == nil {
		activo := domain.EstadoActivo
		f.Estado = &activo
	}
	return s.repo.Listar(ctx, f)
}
```

Y agregar al final del archivo:

```go
// Actualizar reemplaza los campos editables. Funciona también sobre profesionales
// dados de baja: editar los datos de alguien inactivo no tiene por qué
// bloquearse, y no cambia su estado.
func (s *Profesional) Actualizar(ctx context.Context, id uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}

	actualizado, err := actual.AplicarCambios(entrada, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}

	if actualizado.Matricula != actual.Matricula {
		if err := s.verificarMatriculaLibre(ctx, actualizado.Matricula, id); err != nil {
			return domain.Profesional{}, err
		}
	}

	if err := s.repo.Actualizar(ctx, actualizado); err != nil {
		return domain.Profesional{}, err
	}
	return actualizado, nil
}

// DarDeBaja da de baja. No borra: los turnos, comprobantes y pagos históricos
// siguen apuntando a este registro, y sin él el comprobante que un paciente
// presentó para un reintegro queda huérfano.
func (s *Profesional) DarDeBaja(ctx context.Context, id uuid.UUID) error {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, actual.DarDeBaja(s.ahora()))
}

func (s *Profesional) Reactivar(ctx context.Context, id uuid.UUID) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}

	reactivado := actual.Reactivar(s.ahora())
	if err := s.repo.Actualizar(ctx, reactivado); err != nil {
		return domain.Profesional{}, err
	}
	return reactivado, nil
}
```

- [ ] **Step 4: Correr toda la suite del servicio**

Run: `cd apps/api && go test ./internal/service/ -v`
Expected: PASS en todos, incluidos los de la Task 6.

- [ ] **Step 5: Correr toda la suite con detector de carreras**

Run: `cd apps/api && go test ./... -race`
Expected: PASS en `domain`, `repository/memory` y `service`.

- [ ] **Step 6: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/service/
git commit -m "feat(service): edición, baja lógica y reactivación

La baja no borra: pone Estado en inactivo y sella DadoDeBajaEn. Un
turno pasado apunta al profesional, y sin el registro el comprobante
que el paciente presentó para un reintegro queda huérfano.

El listado devuelve solo activos por defecto. Editar la matrícula
verifica que no sea de otro, ignorando al propio profesional.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 8: El contrato OpenAPI

Se escribe **antes** que los handlers. El YAML es la fuente de verdad: los handlers de las Tasks 10 y 11 lo implementan, no al revés.

**Files:**
- Create: `apps/api/api/openapi.yaml`

**Interfaces:**
- Consumes: el modelo de dominio de la Task 4 (nombres de campos y enums)
- Produces: el contrato que implementan las Tasks 10 y 11

- [ ] **Step 1: Escribir el contrato**

Archivo `apps/api/api/openapi.yaml`:

```yaml
openapi: 3.1.0

info:
  title: Salud API
  version: 0.1.0
  description: |
    API de la plataforma de salud digital. Esta versión cubre únicamente el
    CRUD de profesionales, sin autenticación y sin persistencia real.

    Los montos viajan como enteros en centavos. Nunca como decimales: un
    número de JSON se convierte en float64 del otro lado.

servers:
  - url: http://localhost:8080
    description: Desarrollo local

tags:
  - name: profesionales
  - name: salud

paths:
  /healthz:
    get:
      tags: [salud]
      summary: Verifica que el servidor responde
      responses:
        '200':
          description: El servidor está vivo
          content:
            application/json:
              schema:
                type: object
                properties:
                  estado: { type: string, example: ok }

  /api/v1/profesionales:
    get:
      tags: [profesionales]
      summary: Lista profesionales
      description: |
        Por defecto devuelve solo los activos. Para ver los dados de baja hay
        que pedir `estado=inactivo` explícitamente.
      parameters:
        - name: especialidad
          in: query
          schema: { $ref: '#/components/schemas/Especialidad' }
        - name: zona
          in: query
          description: Compara sin distinguir mayúsculas ni acentos
          schema: { type: string }
        - name: estado
          in: query
          schema: { $ref: '#/components/schemas/Estado' }
        - name: busqueda
          in: query
          description: |
            Busca en nombre y apellido, sin distinguir mayúsculas ni acentos.
            Buscar `gonzalez` encuentra a `González`.
          schema: { type: string }
        - name: limite
          in: query
          schema: { type: integer, minimum: 1, maximum: 100, default: 20 }
        - name: desplazamiento
          in: query
          schema: { type: integer, minimum: 0, default: 0 }
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ListaProfesionales' }

    post:
      tags: [profesionales]
      summary: Da de alta un profesional
      description: |
        El profesional nace con `estado: activo` y `verificacion: pendiente`.
        Nadie nace verificado: la verificación es un acto contra REFEPS que
        todavía no está implementado.
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PeticionProfesional' }
      responses:
        '201':
          description: Creado
          headers:
            Location:
              description: URL del recurso creado
              schema: { type: string }
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Profesional' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '409': { $ref: '#/components/responses/Conflicto' }
        '422': { $ref: '#/components/responses/ValidacionFallida' }

  /api/v1/profesionales/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }

    get:
      tags: [profesionales]
      summary: Trae un profesional por ID
      description: |
        Un profesional dado de baja devuelve 200 con `estado: inactivo`. El
        recurso existe, simplemente no opera.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Profesional' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

    put:
      tags: [profesionales]
      summary: Reemplaza los campos editables
      description: |
        Reemplazo total, no parcial. `id`, `slug`, `estado`, `verificacion` y
        las marcas de tiempo no son editables.

        El slug no se regenera al cambiar el nombre: es una URL pública y
        romperla rompe enlaces y posicionamiento.

        Cambiar la matrícula o la especialidad devuelve `verificacion` a
        `pendiente`: la verificación se hizo sobre esos datos.

        Funciona sobre profesionales inactivos y no cambia su estado.
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PeticionProfesional' }
      responses:
        '200':
          description: Actualizado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Profesional' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }
        '409': { $ref: '#/components/responses/Conflicto' }
        '422': { $ref: '#/components/responses/ValidacionFallida' }

    delete:
      tags: [profesionales]
      summary: Da de baja al profesional
      description: |
        Baja lógica, no borrado. Pone `estado` en `inactivo` y sella
        `dadoDeBajaEn`. El registro se queda: los turnos y comprobantes
        históricos apuntan a él.

        Es idempotente: dar de baja algo ya dado de baja devuelve 204.
      responses:
        '204': { description: Dado de baja }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

  /api/v1/profesionales/{id}/reactivar:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }
    post:
      tags: [profesionales]
      summary: Revierte la baja
      description: Idempotente. Sobre alguien ya activo devuelve 200 sin cambios.
      responses:
        '200':
          description: Reactivado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Profesional' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

  /api/v1/profesionales/por-slug/{slug}:
    parameters:
      - name: slug
        in: path
        required: true
        schema: { type: string }
    get:
      tags: [profesionales]
      summary: Trae un profesional por su slug público
      description: Es lo que consume la página `/p/{slug}` del frontend.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Profesional' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

components:
  schemas:
    Especialidad:
      type: string
      enum: [psicologia, kinesiologia, odontologia]

    Modalidad:
      type: string
      enum: [telemedicina, presencial, domicilio]

    Estado:
      type: string
      enum: [activo, inactivo]
      description: Si el profesional opera hoy en la plataforma

    Verificacion:
      type: string
      enum: [pendiente, verificada, rechazada]
      description: Si su matrícula fue verificada contra el mundo real

    Profesional:
      type: object
      required:
        - id
        - slug
        - nombre
        - apellido
        - matricula
        - especialidad
        - bio
        - precioConsultaCentavos
        - modalidades
        - zona
        - obrasSociales
        - estado
        - verificacion
        - creadoEn
        - actualizadoEn
      properties:
        id:
          type: string
          format: uuid
        slug:
          type: string
          example: martin-gonzalez
        nombre:
          type: string
          example: Martín
        apellido:
          type: string
          example: González
        matricula:
          type: string
          description: Forma canónica. Distintas escrituras convergen acá.
          example: MN 98234
        especialidad:
          $ref: '#/components/schemas/Especialidad'
        bio:
          type: string
        precioConsultaCentavos:
          type: integer
          format: int64
          minimum: 0
          description: Centavos. 1200000 son $12.000,00.
          example: 1200000
        modalidades:
          type: array
          minItems: 1
          items: { $ref: '#/components/schemas/Modalidad' }
        zona:
          type: string
          example: CABA
        obrasSociales:
          type: array
          items: { type: string }
          example: [OSDE, Swiss Medical]
        estado:
          $ref: '#/components/schemas/Estado'
        verificacion:
          $ref: '#/components/schemas/Verificacion'
        creadoEn:
          type: string
          format: date-time
        actualizadoEn:
          type: string
          format: date-time
        dadoDeBajaEn:
          type: [string, 'null']
          format: date-time

    PeticionProfesional:
      type: object
      required:
        - nombre
        - apellido
        - matricula
        - especialidad
        - precioConsultaCentavos
        - modalidades
        - zona
      properties:
        nombre:
          type: string
          maxLength: 100
        apellido:
          type: string
          maxLength: 100
        matricula:
          type: string
          description: |
            Acepta las formas reales del mercado: "MN 98.234", "M.N. 45321",
            "mn98234", "MP 12345". Se normaliza al guardar.
          example: MN 98.234
        especialidad:
          $ref: '#/components/schemas/Especialidad'
        bio:
          type: string
          maxLength: 2000
        precioConsultaCentavos:
          type: integer
          format: int64
          minimum: 0
        modalidades:
          type: array
          minItems: 1
          items: { $ref: '#/components/schemas/Modalidad' }
        zona:
          type: string
          maxLength: 100
        obrasSociales:
          type: array
          items: { type: string }
      additionalProperties: false

    ListaProfesionales:
      type: object
      required: [datos, paginacion]
      properties:
        datos:
          type: array
          items: { $ref: '#/components/schemas/Profesional' }
        paginacion:
          type: object
          required: [total, limite, desplazamiento]
          properties:
            total:
              type: integer
              description: Cuántos hay en total con esos filtros, no en la página
            limite: { type: integer }
            desplazamiento: { type: integer }

    Problema:
      type: object
      description: Detalles del problema según RFC 7807
      required: [type, title, status]
      properties:
        type: { type: string, format: uri }
        title: { type: string }
        status: { type: integer }
        detail: { type: string }
        errores:
          type: array
          description: Presente solo en errores de validación
          items:
            type: object
            required: [campo, mensaje]
            properties:
              campo: { type: string }
              mensaje: { type: string }

  responses:
    PeticionInvalida:
      description: JSON malformado o parámetro con formato inválido
      content:
        application/problem+json:
          schema: { $ref: '#/components/schemas/Problema' }

    NoEncontrado:
      description: No existe
      content:
        application/problem+json:
          schema: { $ref: '#/components/schemas/Problema' }

    Conflicto:
      description: La matrícula ya está registrada por otro profesional
      content:
        application/problem+json:
          schema: { $ref: '#/components/schemas/Problema' }

    ValidacionFallida:
      description: |
        JSON válido pero datos inválidos. Distinto de 400: "entendí perfecto
        y está mal" es otro problema que "no entiendo lo que mandaste".
      content:
        application/problem+json:
          schema:
            allOf:
              - $ref: '#/components/schemas/Problema'
              - type: object
                required: [errores]
```

- [ ] **Step 2: Validar que el YAML es sintácticamente correcto**

Run: `cd apps/api && python -c "import yaml,sys; yaml.safe_load(open('api/openapi.yaml',encoding='utf-8')); print('YAML OK')"`
Expected: `YAML OK`

Si Python no está disponible, cualquier validador de YAML sirve. La validación semántica contra el estándar OpenAPI se agrega en la Task 15 junto con el CI.

- [ ] **Step 3: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/api/openapi.yaml
git commit -m "docs(api): contrato OpenAPI del CRUD de Profesional

Escrito a mano y antes que los handlers: el YAML es la fuente de verdad.
Sirve para probar la API con Swagger UI mientras no exista el front, y
después para generar el cliente TypeScript.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 9: Errores HTTP y DTOs

La traducción entre el dominio y el mundo HTTP. Es la única capa que conoce códigos de estado.

**Files:**
- Create: `apps/api/internal/handler/problema.go`
- Create: `apps/api/internal/handler/dto.go`
- Test: `apps/api/internal/handler/problema_test.go`
- Test: `apps/api/internal/handler/dto_test.go`

**Interfaces:**
- Consumes: `domain` (Tasks 2-4)
- Produces:
  - `handler.Problema` con etiquetas JSON `type`, `title`, `status`, `detail`, `errores`
  - `func escribirProblema(http.ResponseWriter, Problema)`
  - `func escribirError(http.ResponseWriter, *http.Request, error)` — mapea errores de dominio a HTTP
  - `func escribirJSON(http.ResponseWriter, int, any)`
  - `func decodificarJSON(http.ResponseWriter, *http.Request, any) error`
  - `respuestaProfesional`, `peticionProfesional`, `respuestaListado` con sus conversores

- [ ] **Step 1: Escribir los tests del mapeo de errores**

Archivo `apps/api/internal/handler/problema_test.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func TestEscribirErrorMapeaLosErroresDelDominio(t *testing.T) {
	casos := []struct {
		nombre         string
		err            error
		statusEsperado int
	}{
		{"no encontrado", domain.ErrNoEncontrado, http.StatusNotFound},
		{"matricula tomada", domain.ErrMatriculaEnUso, http.StatusConflict},
		{
			"validacion",
			domain.ErrorValidacion{Campos: []domain.ErrorCampo{{Campo: "zona", Mensaje: "es obligatoria"}}},
			http.StatusUnprocessableEntity,
		},
		{"desconocido", errors.New("algo explotó"), http.StatusInternalServerError},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			escribirError(rec, req, caso.err)

			if rec.Code != caso.statusEsperado {
				t.Errorf("status = %d, se esperaba %d", rec.Code, caso.statusEsperado)
			}
			if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
				t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
			}

			var p Problema
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
				t.Fatalf("el cuerpo no es JSON válido: %v", err)
			}
			if p.Estado != caso.statusEsperado {
				t.Errorf("problem.status = %d, se esperaba %d", p.Estado, caso.statusEsperado)
			}
			if p.Titulo == "" || p.Tipo == "" {
				t.Error("title y type son obligatorios en RFC 7807")
			}
		})
	}
}

func TestEscribirErrorNoFiltraDetallesInternos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// un error interno no puede llegarle al cliente: puede tener nombres de
	// tablas, rutas del servidor o datos de otro usuario
	escribirError(rec, req, errors.New("pq: relation \"profesionales\" does not exist"))

	if strings.Contains(rec.Body.String(), "profesionales") {
		t.Error("el error interno se filtró al cliente")
	}
}

func TestEscribirErrorDeValidacionIncluyeLosCampos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	verr := domain.ErrorValidacion{Campos: []domain.ErrorCampo{
		{Campo: "matricula", Mensaje: "formato inválido"},
		{Campo: "modalidades", Mensaje: "se requiere al menos una"},
	}}
	escribirError(rec, req, verr)

	var p Problema
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	if len(p.Errores) != 2 {
		t.Fatalf("errores tenía %d elementos, se esperaban 2", len(p.Errores))
	}
	if p.Errores[0].Campo != "matricula" {
		t.Errorf("errores[0].campo = %q", p.Errores[0].Campo)
	}
}

func TestDecodificarJSONRechazaCamposDesconocidos(t *testing.T) {
	rec := httptest.NewRecorder()
	// "precioConsulta" en vez de "precioConsultaCentavos" es exactamente el typo
	// que este modo estricto tiene que atrapar
	cuerpo := strings.NewReader(`{"nombre":"Ana","precioConsulta":100}`)
	req := httptest.NewRequest(http.MethodPost, "/", cuerpo)

	var destino peticionProfesional
	if err := decodificarJSON(rec, req, &destino); err == nil {
		t.Error("un campo desconocido debía ser rechazado")
	}
}

func TestDecodificarJSONRechazaBasuraDespuesDelObjeto(t *testing.T) {
	rec := httptest.NewRecorder()
	cuerpo := strings.NewReader(`{"nombre":"Ana"} {"nombre":"Otro"}`)
	req := httptest.NewRequest(http.MethodPost, "/", cuerpo)

	var destino peticionProfesional
	if err := decodificarJSON(rec, req, &destino); err == nil {
		t.Error("dos objetos JSON seguidos debían ser rechazados")
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: FAIL con `undefined: escribirError`

- [ ] **Step 3: Implementar `problema.go`**

Archivo `apps/api/internal/handler/problema.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const tipoContenidoProblema = "application/problem+json"

// Los tipos de error del contrato. Son URIs por convención de RFC 7807: no
// hace falta que resuelvan a una página, alcanza con que identifiquen la clase
// de problema de forma estable.
const (
	tipoPeticionInvalida = "https://salud.app/errors/bad-request"
	tipoValidacion       = "https://salud.app/errors/validation"
	tipoNoEncontrado     = "https://salud.app/errors/not-found"
	tipoConflicto        = "https://salud.app/errors/conflict"
	tipoInterno          = "https://salud.app/errors/internal"
)

// Problema es la representación de un error según RFC 7807.
type Problema struct {
	Tipo    string              `json:"type"`
	Titulo  string              `json:"title"`
	Estado  int                 `json:"status"`
	Detalle string              `json:"detail,omitempty"`
	Errores []domain.ErrorCampo `json:"errores,omitempty"`
}

func escribirProblema(w http.ResponseWriter, p Problema) {
	w.Header().Set("Content-Type", tipoContenidoProblema)
	w.WriteHeader(p.Estado)
	_ = json.NewEncoder(w).Encode(p)
}

func escribirJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// escribirError traduce un error del dominio al problema HTTP que le corresponde.
//
// Es el único lugar del proyecto donde el dominio se convierte en códigos de
// estado. Las capas de abajo no saben que existe HTTP.
func escribirError(w http.ResponseWriter, r *http.Request, err error) {
	var verr domain.ErrorValidacion

	switch {
	case errors.As(err, &verr):
		escribirProblema(w, Problema{
			Tipo:    tipoValidacion,
			Titulo:  "Datos inválidos",
			Estado:  http.StatusUnprocessableEntity,
			Detalle: "Uno o más campos no cumplen las reglas del sistema",
			Errores: verr.Campos,
		})

	case errors.Is(err, domain.ErrNoEncontrado):
		escribirProblema(w, Problema{
			Tipo:    tipoNoEncontrado,
			Titulo:  "No encontrado",
			Estado:  http.StatusNotFound,
			Detalle: "El profesional solicitado no existe",
		})

	case errors.Is(err, domain.ErrMatriculaEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Matrícula ya registrada",
			Estado:  http.StatusConflict,
			Detalle: "Otro profesional ya tiene registrada esa matrícula",
		})

	default:
		// El error real va al log, nunca al cliente: puede contener nombres
		// de tablas, rutas del servidor o datos de otro usuario.
		slog.ErrorContext(r.Context(), "error no manejado",
			"error", err,
			"metodo", r.Method,
			"ruta", r.URL.Path,
		)
		escribirProblema(w, Problema{
			Tipo:    tipoInterno,
			Titulo:  "Error interno",
			Estado:  http.StatusInternalServerError,
			Detalle: "Ocurrió un error inesperado. Volvé a intentar.",
		})
	}
}

func escribirPeticionInvalida(w http.ResponseWriter, detail string) {
	escribirProblema(w, Problema{
		Tipo:    tipoPeticionInvalida,
		Titulo:  "Petición inválida",
		Estado:  http.StatusBadRequest,
		Detalle: detail,
	})
}
```

- [ ] **Step 4: Implementar `dto.go`**

Archivo `apps/api/internal/handler/dto.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const maxBytesCuerpo = 1 << 20 // 1 MB

// peticionProfesional es lo que entra. Deliberadamente no incluye id, slug,
// estado, verificacion ni marcas de tiempo: son campos que el servidor decide,
// y aceptarlos sería dejar que el cliente se autoverifique.
type peticionProfesional struct {
	Nombre                 string   `json:"nombre"`
	Apellido               string   `json:"apellido"`
	Matricula              string   `json:"matricula"`
	Especialidad           string   `json:"especialidad"`
	Bio                    string   `json:"bio"`
	PrecioConsultaCentavos int64    `json:"precioConsultaCentavos"`
	Modalidades            []string `json:"modalidades"`
	Zona                   string   `json:"zona"`
	ObrasSociales          []string `json:"obrasSociales"`
}

func (r peticionProfesional) aEntrada() domain.EntradaProfesional {
	return domain.EntradaProfesional{
		Nombre:         r.Nombre,
		Apellido:       r.Apellido,
		Matricula:      r.Matricula,
		Especialidad:   r.Especialidad,
		Bio:            r.Bio,
		PrecioConsulta: r.PrecioConsultaCentavos,
		Modalidades:    r.Modalidades,
		Zona:           r.Zona,
		ObrasSociales:  r.ObrasSociales,
	}
}

type respuestaProfesional struct {
	ID                     string     `json:"id"`
	Slug                   string     `json:"slug"`
	Nombre                 string     `json:"nombre"`
	Apellido               string     `json:"apellido"`
	Matricula              string     `json:"matricula"`
	Especialidad           string     `json:"especialidad"`
	Bio                    string     `json:"bio"`
	PrecioConsultaCentavos int64      `json:"precioConsultaCentavos"`
	Modalidades            []string   `json:"modalidades"`
	Zona                   string     `json:"zona"`
	ObrasSociales          []string   `json:"obrasSociales"`
	Estado                 string     `json:"estado"`
	Verificacion           string     `json:"verificacion"`
	CreadoEn               time.Time  `json:"creadoEn"`
	ActualizadoEn          time.Time  `json:"actualizadoEn"`
	DadoDeBajaEn           *time.Time `json:"dadoDeBajaEn"`
}

func aRespuesta(p domain.Profesional) respuestaProfesional {
	// make con len 0 en vez de nil: un slice nil se serializa como null y el
	// cliente TypeScript tendría que chequearlo en cada uso
	mods := make([]string, 0, len(p.Modalidades))
	for _, m := range p.Modalidades {
		mods = append(mods, string(m))
	}

	obras := make([]string, 0, len(p.ObrasSociales))
	obras = append(obras, p.ObrasSociales...)

	return respuestaProfesional{
		ID:                     p.ID.String(),
		Slug:                   p.Slug,
		Nombre:                 p.Nombre,
		Apellido:               p.Apellido,
		Matricula:              p.Matricula.String(),
		Especialidad:           string(p.Especialidad),
		Bio:                    p.Bio,
		PrecioConsultaCentavos: int64(p.PrecioConsulta),
		Modalidades:            mods,
		Zona:                   p.Zona,
		ObrasSociales:          obras,
		Estado:                 string(p.Estado),
		Verificacion:           string(p.Verificacion),
		CreadoEn:               p.CreadoEn,
		ActualizadoEn:          p.ActualizadoEn,
		DadoDeBajaEn:           p.DadoDeBajaEn,
	}
}

type respuestaPaginacion struct {
	Total          int `json:"total"`
	Limite         int `json:"limite"`
	Desplazamiento int `json:"desplazamiento"`
}

type respuestaListado struct {
	Datos      []respuestaProfesional `json:"datos"`
	Paginacion respuestaPaginacion    `json:"paginacion"`
}

func aRespuestaListado(ps []domain.Profesional, total, limite, desplazamiento int) respuestaListado {
	datos := make([]respuestaProfesional, 0, len(ps))
	for _, p := range ps {
		datos = append(datos, aRespuesta(p))
	}
	return respuestaListado{
		Datos:      datos,
		Paginacion: respuestaPaginacion{Total: total, Limite: limite, Desplazamiento: desplazamiento},
	}
}

// decodificarJSON lee el cuerpo en modo estricto.
//
// DisallowUnknownFields atrapa el typo más probable de esta API: mandar
// "precioConsulta" en vez de "precioConsultaCentavos" y que el precio quede en cero
// sin que nadie se entere.
func decodificarJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytesCuerpo)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(destino); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("el cuerpo debe contener un único objeto JSON")
	}
	return nil
}
```

- [ ] **Step 5: Escribir el test de serialización**

Archivo `apps/api/internal/handler/dto_test.go`:

```go
package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func TestARespuestaSerializaLosCamposDelContrato(t *testing.T) {
	ahora := time.Date(2026, 8, 21, 14, 2, 11, 0, time.UTC)
	p, err := domain.NuevoProfesional(domain.EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico.",
		PrecioConsulta: 1200000,
		Modalidades:    []string{"telemedicina", "presencial"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE"},
	}, ahora)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	crudo, err := json.Marshal(aRespuesta(p))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	cuerpo := string(crudo)

	// el precio viaja como entero, sin punto decimal
	if !strings.Contains(cuerpo, `"precioConsultaCentavos":1200000`) {
		t.Errorf("el precio no viaja como entero de centavos: %s", cuerpo)
	}
	// la matrícula viaja en forma canónica
	if !strings.Contains(cuerpo, `"matricula":"MN 98234"`) {
		t.Errorf("la matrícula no viaja canónica: %s", cuerpo)
	}
	if !strings.Contains(cuerpo, `"dadoDeBajaEn":null`) {
		t.Errorf("dadoDeBajaEn debía estar presente como null: %s", cuerpo)
	}
}

func TestARespuestaNuncaSerializaSlicesComoNull(t *testing.T) {
	// un slice nil se serializa como null y obliga al cliente TypeScript a
	// chequearlo en cada uso
	var p domain.Profesional
	crudo, err := json.Marshal(aRespuesta(p))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	cuerpo := string(crudo)

	if strings.Contains(cuerpo, `"modalidades":null`) {
		t.Error("modalidades se serializó como null en vez de []")
	}
	if strings.Contains(cuerpo, `"obrasSociales":null`) {
		t.Error("obrasSociales se serializó como null en vez de []")
	}
}

func TestARespuestaListadoConListaVacia(t *testing.T) {
	crudo, err := json.Marshal(aRespuestaListado(nil, 0, 20, 0))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	if strings.Contains(string(crudo), `"datos":null`) {
		t.Errorf("datos se serializó como null en vez de []: %s", crudo)
	}
}
```

- [ ] **Step 6: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: PASS en los tests de `problema_test.go` y `dto_test.go`.

- [ ] **Step 7: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/handler/
git commit -m "feat(handler): errores RFC 7807 y DTOs

escribirError es el único lugar donde el dominio se convierte en códigos
HTTP. Los errores internos van al log, nunca al cliente.

decodificarJSON rechaza campos desconocidos: atrapa el typo de mandar
precioConsulta en vez de precioConsultaCentavos, que dejaría el precio en
cero sin que nadie se entere.

Los slices se serializan como [] y nunca como null.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 10: Middleware

**Files:**
- Create: `apps/api/internal/handler/middleware.go`
- Test: `apps/api/internal/handler/middleware_test.go`

**Interfaces:**
- Consumes: `problema.go` (Task 9). No depende de los handlers: por eso va antes.
- Produces:
  - `func handler.Encadenar(http.Handler, ...func(http.Handler) http.Handler) http.Handler`
  - `func handler.IDPeticion(http.Handler) http.Handler`
  - `func handler.RegistrarPeticiones(http.Handler) http.Handler`
  - `func handler.RecuperarPanic(http.Handler) http.Handler`

- [ ] **Step 1: Escribir los tests**

Archivo `apps/api/internal/handler/middleware_test.go`:

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIDPeticionGeneraUnoSiNoViene(t *testing.T) {
	var visto string
	h := IDPeticion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visto = IDPeticionDe(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if visto == "" {
		t.Error("no se generó un request id")
	}
	if rec.Header().Get("X-Request-ID") != visto {
		t.Error("el request id tenía que volver en el header de la respuesta")
	}
}

func TestIDPeticionRespetaElQueViene(t *testing.T) {
	var visto string
	h := IDPeticion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visto = IDPeticionDe(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "trae-el-suyo")

	h.ServeHTTP(httptest.NewRecorder(), req)

	// permite seguir un request a través de varios servicios
	if visto != "trae-el-suyo" {
		t.Errorf("request id = %q, se esperaba el que vino en el header", visto)
	}
}

func TestRecuperarPanicDevuelve500(t *testing.T) {
	h := RecuperarPanic(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("algo explotó")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, se esperaba 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}
	// el mensaje del panic no puede llegarle al cliente
	if strings.Contains(rec.Body.String(), "algo explotó") {
		t.Error("el mensaje del panic se filtró al cliente")
	}
}

func TestEncadenarAplicaEnOrden(t *testing.T) {
	var orden []string

	marcar := func(nombre string) func(http.Handler) http.Handler {
		return func(siguiente http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				orden = append(orden, nombre)
				siguiente.ServeHTTP(w, r)
			})
		}
	}

	final := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		orden = append(orden, "handler")
	})

	Encadenar(final, marcar("primero"), marcar("segundo")).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	esperado := []string{"primero", "segundo", "handler"}
	if len(orden) != len(esperado) {
		t.Fatalf("orden = %v, se esperaba %v", orden, esperado)
	}
	for i := range esperado {
		if orden[i] != esperado[i] {
			t.Fatalf("orden = %v, se esperaba %v", orden, esperado)
		}
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/handler/ -run 'TestIDPeticion|TestRecuperar|TestEncadenar' -v`
Expected: FAIL con `undefined: IDPeticion`

- [ ] **Step 3: Implementar el middleware**

Archivo `apps/api/internal/handler/middleware.go`:

```go
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const headerIDPeticion = "X-Request-ID"

type claveContexto string

const claveIDPeticion claveContexto = "requestID"

// Encadenar envuelve el handler. El primer middleware de la lista es el más
// externo: el primero en ver el request y el último en ver la respuesta.
func Encadenar(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// IDPeticion asegura que todo request tenga un identificador. Si el cliente
// manda uno, se respeta: permite seguir una operación a través de varios
// servicios en los logs.
func IDPeticion(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerIDPeticion)
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set(headerIDPeticion, id)
		siguiente.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claveIDPeticion, id)))
	})
}

// IDPeticionDe devuelve el identificador del request, o cadena vacía si el
// middleware no corrió.
func IDPeticionDe(ctx context.Context) string {
	id, _ := ctx.Value(claveIDPeticion).(string)
	return id
}

// grabadorEstado captura el código de estado para poder loguearlo: el
// http.ResponseWriter no lo expone.
type grabadorEstado struct {
	http.ResponseWriter
	estado int
	bytes  int
}

func (w *grabadorEstado) WriteHeader(codigo int) {
	w.estado = codigo
	w.ResponseWriter.WriteHeader(codigo)
}

func (w *grabadorEstado) Write(b []byte) (int, error) {
	if w.estado == 0 {
		w.estado = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func RegistrarPeticiones(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rec := &grabadorEstado{ResponseWriter: w, estado: http.StatusOK}

		siguiente.ServeHTTP(rec, r)

		slog.InfoContext(r.Context(), "peticion",
			"idPeticion", IDPeticionDe(r.Context()),
			"metodo", r.Method,
			"ruta", r.URL.Path,
			"estado", rec.estado,
			"bytes", rec.bytes,
			"duracionMs", time.Since(inicio).Milliseconds(),
		)
	})
}

// RecuperarPanic evita que un panic en un handler tire el proceso entero.
//
// Va por dentro de RegistrarPeticiones para que el log registre el 500 resultante.
func RecuperarPanic(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// el detalle del panic va al log, nunca al cliente
			slog.ErrorContext(r.Context(), "panic recuperado",
				"idPeticion", IDPeticionDe(r.Context()),
				"panic", rec,
				"metodo", r.Method,
				"ruta", r.URL.Path,
			)

			escribirProblema(w, Problema{
				Tipo:    tipoInterno,
				Titulo:  "Error interno",
				Estado:  http.StatusInternalServerError,
				Detalle: "Ocurrió un error inesperado. Volvé a intentar.",
			})
		}()

		siguiente.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4: Correr toda la suite del paquete handler**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: PASS en los tests del middleware y en los de `problema.go` y `dto.go` de la Task 9.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/handler/
git commit -m "feat(handler): middleware de request id, logging y recover

IDPeticion respeta el que manda el cliente: permite seguir una operación
a través de varios servicios en los logs.

RecuperarPanic evita que un handler tire el proceso, y el detalle del
panic va al log y nunca al cliente.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 11: Handlers y router

**Files:**
- Create: `apps/api/internal/handler/profesional.go`
- Create: `apps/api/internal/handler/router.go`
- Test: `apps/api/internal/handler/profesional_test.go`

**Interfaces:**
- Consumes: `service.Profesional` (Tasks 6-7), `problema.go` y `dto.go` (Task 9), el middleware (Task 10)
- Produces:
  - `func handler.NuevoProfesional(*service.Profesional) *handler.ManejadorProfesional`
  - Métodos `Crear`, `Listar`, `ObtenerPorID`, `ObtenerPorSlug`, `Actualizar`, `DarDeBaja`, `Reactivar`
  - `func handler.NuevoRouter(*ManejadorProfesional) http.Handler`

- [ ] **Step 1: Escribir los tests de la capa HTTP**

Estos tests corren contra el stack completo cableado: router, handler, servicio y repositorio en memoria. No hay mocks.

Archivo `apps/api/internal/handler/profesional_test.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// nuevoServidorDePrueba cablea el stack real de punta a punta.
func nuevoServidorDePrueba(t *testing.T) *httptest.Server {
	t.Helper()
	repo := memory.NuevoProfesional()
	svc := service.NuevoProfesional(repo)
	srv := httptest.NewServer(NuevoRouter(NuevoProfesional(svc)))
	t.Cleanup(srv.Close)
	return srv
}

const cuerpoValido = `{
  "nombre": "Martín",
  "apellido": "González",
  "matricula": "MN 98.234",
  "especialidad": "psicologia",
  "bio": "Psicólogo clínico.",
  "precioConsultaCentavos": 1200000,
  "modalidades": ["telemedicina", "presencial"],
  "zona": "CABA",
  "obrasSociales": ["OSDE"]
}`

func postear(t *testing.T, srv *httptest.Server, ruta, cuerpo string) *http.Response {
	t.Helper()
	resp, err := srv.Client().Post(srv.URL+ruta, "application/json", strings.NewReader(cuerpo))
	if err != nil {
		t.Fatalf("POST %s falló: %v", ruta, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func obtener(t *testing.T, srv *httptest.Server, ruta string) *http.Response {
	t.Helper()
	resp, err := srv.Client().Get(srv.URL + ruta)
	if err != nil {
		t.Fatalf("GET %s falló: %v", ruta, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func ejecutar(t *testing.T, srv *httptest.Server, metodo, ruta, cuerpo string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(metodo, srv.URL+ruta, strings.NewReader(cuerpo))
	if err != nil {
		t.Fatalf("no se pudo armar el request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s falló: %v", metodo, ruta, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decodificarProfesional(t *testing.T, resp *http.Response) respuestaProfesional {
	t.Helper()
	var p respuestaProfesional
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	return p
}

func TestHealthz(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := obtener(t, srv, "/healthz")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, se esperaba 200", resp.StatusCode)
	}
}

func TestCrearDevuelve201YLocation(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201", resp.StatusCode)
	}

	p := decodificarProfesional(t, resp)
	if p.Slug != "martin-gonzalez" {
		t.Errorf("slug = %q, se esperaba martin-gonzalez", p.Slug)
	}
	if p.Verificacion != "pendiente" {
		t.Errorf("verificacion = %q, se esperaba pendiente", p.Verificacion)
	}

	ubicacion := resp.Header.Get("Location")
	if ubicacion != "/api/v1/profesionales/"+p.ID {
		t.Errorf("Location = %q, se esperaba la URL del recurso creado", ubicacion)
	}
}

func TestCrearJSONMalformadoDevuelve400(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := postear(t, srv, "/api/v1/profesionales", `{"nombre":`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}
}

func TestCrearDatosInvalidosDevuelve422(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	// JSON perfectamente válido, datos mal: es otro problema que un 400
	cuerpo := `{
	  "nombre": "",
	  "apellido": "González",
	  "matricula": "roto",
	  "especialidad": "cardiologia",
	  "precioConsultaCentavos": -5,
	  "modalidades": [],
	  "zona": ""
	}`
	resp := postear(t, srv, "/api/v1/profesionales", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, se esperaba 422", resp.StatusCode)
	}

	var p Problema
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("no se pudo decodificar el problem: %v", err)
	}
	if len(p.Errores) < 5 {
		t.Errorf("se esperaban al menos 5 campos con error, se obtuvieron %d: %+v", len(p.Errores), p.Errores)
	}
}

func TestCrearMatriculaDuplicadaDevuelve409(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, se esperaba 409", resp.StatusCode)
	}
}

func TestObtenerPorIDYPorSlug(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))

	porID := obtener(t, srv, "/api/v1/profesionales/"+creado.ID)
	if porID.StatusCode != http.StatusOK {
		t.Errorf("GET por id: status = %d, se esperaba 200", porID.StatusCode)
	}

	porSlug := obtener(t, srv, "/api/v1/profesionales/por-slug/"+creado.Slug)
	if porSlug.StatusCode != http.StatusOK {
		t.Errorf("GET por slug: status = %d, se esperaba 200", porSlug.StatusCode)
	}
	if decodificarProfesional(t, porSlug).ID != creado.ID {
		t.Error("GET por slug devolvió otro profesional")
	}
}

func TestObtenerIDInexistenteDevuelve404(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := obtener(t, srv, "/api/v1/profesionales/6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, se esperaba 404", resp.StatusCode)
	}
}

func TestObtenerIDMalFormadoDevuelve400(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	// no es un UUID: es un problema del cliente, no un recurso que falta
	resp := obtener(t, srv, "/api/v1/profesionales/no-soy-un-uuid")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestListarPaginaYFiltra(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	segundo := strings.Replace(cuerpoValido, `"MN 98.234"`, `"MN 11111"`, 1)
	segundo = strings.Replace(segundo, `"psicologia"`, `"odontologia"`, 1)
	postear(t, srv, "/api/v1/profesionales", segundo)

	t.Run("sin filtros", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales")
		var listado respuestaListado
		if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if listado.Paginacion.Total != 2 {
			t.Errorf("total = %d, se esperaba 2", listado.Paginacion.Total)
		}
		if listado.Paginacion.Limite != service.LimitePorDefecto {
			t.Errorf("limite = %d, se esperaba el default %d", listado.Paginacion.Limite, service.LimitePorDefecto)
		}
	})

	t.Run("por especialidad", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?especialidad=odontologia")
		var listado respuestaListado
		if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if listado.Paginacion.Total != 1 {
			t.Errorf("total = %d, se esperaba 1", listado.Paginacion.Total)
		}
	})

	t.Run("busqueda sin acentos", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?busqueda=gonzalez")
		var listado respuestaListado
		if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if listado.Paginacion.Total != 2 {
			t.Errorf("total = %d, se esperaban 2: ambos apellidan González", listado.Paginacion.Total)
		}
	})

	t.Run("limite invalido devuelve 400", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?limite=abc")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
		}
	})

	t.Run("especialidad invalida devuelve 400", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?especialidad=cardiologia")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
		}
	})
}

func TestActualizar(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))

	cuerpo := strings.Replace(cuerpoValido, `"CABA"`, `"GBA Norte"`, 1)
	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+creado.ID, cuerpo)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}
	actualizado := decodificarProfesional(t, resp)
	if actualizado.Zona != "GBA Norte" {
		t.Errorf("zona = %q, se esperaba GBA Norte", actualizado.Zona)
	}
	if actualizado.Slug != creado.Slug {
		t.Error("el slug es una URL pública y no debía cambiar")
	}
}

func TestDeleteEsBajaLogicaYIdempotente(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))

	resp := ejecutar(t, srv, http.MethodDelete, "/api/v1/profesionales/"+creado.ID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, se esperaba 204", resp.StatusCode)
	}

	// el recurso sigue existiendo: no fue un borrado
	despues := obtener(t, srv, "/api/v1/profesionales/"+creado.ID)
	if despues.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: el profesional dado de baja tenía que seguir existiendo", despues.StatusCode)
	}
	if p := decodificarProfesional(t, despues); p.Estado != "inactivo" || p.DadoDeBajaEn == nil {
		t.Error("debía quedar inactivo con dadoDeBajaEn sellado")
	}

	// pero no aparece en el listado por defecto
	respListado := obtener(t, srv, "/api/v1/profesionales")
	var listado respuestaListado
	if err := json.NewDecoder(respListado.Body).Decode(&listado); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if listado.Paginacion.Total != 0 {
		t.Errorf("total = %d, se esperaba 0", listado.Paginacion.Total)
	}

	// y una segunda baja no es un error
	otraVez := ejecutar(t, srv, http.MethodDelete, "/api/v1/profesionales/"+creado.ID, "")
	if otraVez.StatusCode != http.StatusNoContent {
		t.Errorf("la segunda baja devolvió %d, se esperaba 204", otraVez.StatusCode)
	}
}

func TestReactivar(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))
	ejecutar(t, srv, http.MethodDelete, "/api/v1/profesionales/"+creado.ID, "")

	resp := postear(t, srv, "/api/v1/profesionales/"+creado.ID+"/reactivar", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	p := decodificarProfesional(t, resp)
	if p.Estado != "activo" || p.DadoDeBajaEn != nil {
		t.Error("debía quedar activo y con dadoDeBajaEn en null")
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/handler/ -run 'TestHealthz|TestCrear|TestObtener|TestListar|TestActualizar|TestDelete|TestReactivar' -v`
Expected: FAIL con `undefined: NuevoRouter`

- [ ] **Step 3: Implementar los handlers**

Archivo `apps/api/internal/handler/profesional.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// ManejadorProfesional traduce entre HTTP y el servicio. No toma decisiones de
// negocio: decodifica, delega y serializa.
type ManejadorProfesional struct {
	svc *service.Profesional
}

func NuevoProfesional(svc *service.Profesional) *ManejadorProfesional {
	return &ManejadorProfesional{svc: svc}
}

func (h *ManejadorProfesional) Crear(w http.ResponseWriter, r *http.Request) {
	var req peticionProfesional
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirPeticionInvalida(w, "el cuerpo no es un JSON válido: "+err.Error())
		return
	}

	p, err := h.svc.Crear(r.Context(), req.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/profesionales/"+p.ID.String())
	escribirJSON(w, http.StatusCreated, aRespuesta(p))
}

func (h *ManejadorProfesional) ObtenerPorID(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

func (h *ManejadorProfesional) ObtenerPorSlug(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.ObtenerPorSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

func (h *ManejadorProfesional) Listar(w http.ResponseWriter, r *http.Request) {
	f, err := parsearFiltro(r)
	if err != nil {
		escribirPeticionInvalida(w, err.Error())
		return
	}

	ps, total, err := h.svc.Listar(r.Context(), f)
	if err != nil {
		escribirError(w, r, err)
		return
	}

	limite := f.Limite
	if limite <= 0 {
		limite = service.LimitePorDefecto
	}
	escribirJSON(w, http.StatusOK, aRespuestaListado(ps, total, limite, f.Desplazamiento))
}

func (h *ManejadorProfesional) Actualizar(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	var req peticionProfesional
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirPeticionInvalida(w, "el cuerpo no es un JSON válido: "+err.Error())
		return
	}

	p, err := h.svc.Actualizar(r.Context(), id, req.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

// DarDeBaja implementa el DELETE. Se llama DarDeBaja y no Delete porque eso
// es lo que hace: baja lógica. El verbo HTTP conserva el nombre que espera
// cualquiera que lea un CRUD.
func (h *ManejadorProfesional) DarDeBaja(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	if err := h.svc.DarDeBaja(r.Context(), id); err != nil {
		escribirError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorProfesional) Reactivar(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.Reactivar(r.Context(), id)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

// parsearID devuelve false y ya escribió la respuesta si el ID no es un UUID.
// Un ID mal formado es un error del cliente (400), no un recurso que falta.
func parsearID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	crudo := r.PathValue("id")
	id, err := uuid.Parse(crudo)
	if err != nil {
		escribirPeticionInvalida(w, "el id debe ser un UUID válido")
		return uuid.Nil, false
	}
	return id, true
}

func parsearFiltro(r *http.Request) (repository.Filtro, error) {
	q := r.URL.Query()
	var f repository.Filtro

	if crudo := q.Get("especialidad"); crudo != "" {
		esp := domain.Especialidad(crudo)
		if !esp.EsValida() {
			return f, errParametroInvalido("especialidad", "debe ser psicologia, kinesiologia u odontologia")
		}
		f.Especialidad = &esp
	}

	if crudo := q.Get("estado"); crudo != "" {
		st := domain.Estado(crudo)
		if !st.EsValido() {
			return f, errParametroInvalido("estado", "debe ser activo o inactivo")
		}
		f.Estado = &st
	}

	if crudo := q.Get("zona"); crudo != "" {
		f.Zona = &crudo
	}
	if crudo := q.Get("busqueda"); crudo != "" {
		f.Busqueda = &crudo
	}

	if crudo := q.Get("limite"); crudo != "" {
		v, err := strconv.Atoi(crudo)
		if err != nil || v < 1 {
			return f, errParametroInvalido("limite", "debe ser un entero mayor a cero")
		}
		f.Limite = v
	}

	if crudo := q.Get("desplazamiento"); crudo != "" {
		v, err := strconv.Atoi(crudo)
		if err != nil || v < 0 {
			return f, errParametroInvalido("desplazamiento", "debe ser un entero mayor o igual a cero")
		}
		f.Desplazamiento = v
	}

	return f, nil
}

type errorParametro struct {
	parametro string
	mensaje   string
}

func (e errorParametro) Error() string {
	return "parámetro " + e.parametro + ": " + e.mensaje
}

func errParametroInvalido(parametro, mensaje string) error {
	return errorParametro{parametro: parametro, mensaje: mensaje}
}
```

- [ ] **Step 4: Implementar el router**

Archivo `apps/api/internal/handler/router.go`:

```go
package handler

import "net/http"

// NuevoRouter arma la tabla de rutas.
//
// Usa el ServeMux de la stdlib: desde Go 1.22 entiende método y parámetros de
// ruta, así que no hace falta chi, gin ni echo para esto.
func NuevoRouter(ph *ManejadorProfesional) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)

	mux.HandleFunc("GET /api/v1/profesionales", ph.Listar)
	mux.HandleFunc("POST /api/v1/profesionales", ph.Crear)
	mux.HandleFunc("GET /api/v1/profesionales/{id}", ph.ObtenerPorID)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}", ph.Actualizar)
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}", ph.DarDeBaja)
	mux.HandleFunc("POST /api/v1/profesionales/{id}/reactivar", ph.Reactivar)

	// No colisiona con /{id}: tiene un segmento más y el ServeMux resuelve
	// por especificidad.
	mux.HandleFunc("GET /api/v1/profesionales/por-slug/{slug}", ph.ObtenerPorSlug)

	// El orden es de afuera hacia adentro. IDPeticion va primero para que el
	// log lo tenga; RegistrarPeticiones envuelve a RecuperarPanic para que un panic
	// quede registrado con su 500.
	return Encadenar(mux, IDPeticion, RegistrarPeticiones, RecuperarPanic)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}
```

- [ ] **Step 5: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: PASS en todos los tests del paquete, incluidos los del middleware de la Task 10.

- [ ] **Step 6: Correr toda la suite con detector de carreras**

Run: `cd apps/api && go test ./... -race`
Expected: PASS en `domain`, `repository/memory`, `service` y `handler`.

- [ ] **Step 7: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/handler/
git commit -m "feat(handler): controllers y tabla de rutas

Los handlers no toman decisiones de negocio: decodifican, delegan y
serializan. Un id mal formado es 400, no 404: es un error del cliente,
no un recurso que falta.

El ServeMux de la stdlib alcanza: desde Go 1.22 entiende método y
parámetros de ruta.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 12: Sembrar de desarrollo

**Files:**
- Create: `apps/api/internal/repository/memory/semilla.go`
- Test: `apps/api/internal/repository/memory/semilla_test.go`

**Interfaces:**
- Consumes: `domain` (Task 4), `memory.Profesional` (Task 5)
- Produces: `func memory.Sembrar(context.Context, *Profesional) error`

- [ ] **Step 1: Escribir el test**

Archivo `apps/api/internal/repository/memory/semilla_test.go`:

```go
package memory

import (
	"context"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

func TestSembrar(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	if err := Sembrar(ctx, repo); err != nil {
		t.Fatalf("Sembrar devolvió error: %v", err)
	}

	_, total, err := repo.Listar(ctx, repository.Filtro{Limite: 100})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if total != 4 {
		t.Errorf("total = %d, se esperaban 4 profesionales de prueba", total)
	}
}

func TestSembrarGeneraSlugsUnicos(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	if err := Sembrar(ctx, repo); err != nil {
		t.Fatalf("Sembrar devolvió error: %v", err)
	}

	ps, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 100})

	slugs := make(map[string]bool, len(ps))
	matriculas := make(map[string]bool, len(ps))
	for _, p := range ps {
		if slugs[p.Slug] {
			t.Errorf("slug repetido en el seed: %q", p.Slug)
		}
		slugs[p.Slug] = true

		if matriculas[p.Matricula.String()] {
			t.Errorf("matrícula repetida en el seed: %q", p.Matricula)
		}
		matriculas[p.Matricula.String()] = true
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/repository/memory/ -run TestSembrar -v`
Expected: FAIL con `undefined: Sembrar`

- [ ] **Step 3: Implementar el seed**

Los datos salen de `legacy/prototype/src/data/profesionales.js`, adaptados al modelo nuevo. Los precios del prototipo estaban en pesos; acá van en centavos.

Archivo `apps/api/internal/repository/memory/semilla.go`:

```go
package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Sembrar carga profesionales de prueba. Solo se llama en desarrollo: main.go lo
// invoca únicamente cuando APP_ENV=development.
//
// Los datos vienen del prototipo React, en legacy/prototype/src/data. Los
// precios estaban en pesos y acá van en centavos.
func Sembrar(ctx context.Context, repo *Profesional) error {
	ahora := time.Now().UTC()

	entradas := []domain.EntradaProfesional{
		{
			Nombre:         "Martín",
			Apellido:       "González",
			Matricula:      "MN 98.234",
			Especialidad:   string(domain.EspecialidadPsicologia),
			Bio:            "Psicólogo clínico con orientación cognitivo-conductual. Atiendo adultos y adolescentes con ansiedad, depresión y crisis vitales. Más de 8 años de experiencia.",
			PrecioConsulta: 1_200_000,
			Modalidades:    []string{"telemedicina", "presencial"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Swiss Medical", "Galeno", "Medifé"},
		},
		{
			Nombre:         "Carolina",
			Apellido:       "Vega",
			Matricula:      "MN 112.087",
			Especialidad:   string(domain.EspecialidadPsicologia),
			Bio:            "Especializada en terapia sistémica y de parejas. Trabajo con adultos en procesos de cambio, duelos y conflictos relacionales.",
			PrecioConsulta: 1_400_000,
			Modalidades:    []string{"telemedicina"},
			Zona:           "GBA Norte",
			ObrasSociales:  []string{"OSDE", "OMINT", "Swiss Medical", "Sanitas"},
		},
		{
			Nombre:         "Pablo",
			Apellido:       "Moreno",
			Matricula:      "MN 45.321",
			Especialidad:   string(domain.EspecialidadKinesiologia),
			Bio:            "Kinesiólogo especializado en traumatología deportiva y rehabilitación postquirúrgica. Atiendo a domicilio y en consultorio.",
			PrecioConsulta: 950_000,
			Modalidades:    []string{"presencial", "domicilio"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Galeno", "IOMA", "PAMI"},
		},
		{
			Nombre:         "Gabriela",
			Apellido:       "Ríos",
			Matricula:      "MN 67.890",
			Especialidad:   string(domain.EspecialidadOdontologia),
			Bio:            "Odontóloga general con especialización en estética dental. Trabajo con materiales de primera calidad en un consultorio moderno en Palermo.",
			PrecioConsulta: 1_500_000,
			Modalidades:    []string{"presencial"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Swiss Medical", "Galeno", "Medifé", "OMINT"},
		},
	}

	for i, entrada := range entradas {
		p, err := domain.NuevoProfesional(entrada, ahora.Add(time.Duration(i)*time.Second))
		if err != nil {
			return fmt.Errorf("seed: profesional %d (%s %s): %w", i, entrada.Nombre, entrada.Apellido, err)
		}
		if err := repo.Crear(ctx, p); err != nil {
			return fmt.Errorf("seed: guardando %s %s: %w", entrada.Nombre, entrada.Apellido, err)
		}
	}
	return nil
}
```

Nota: el seed usa `domain.NuevoProfesional` directamente en vez del servicio, así que no resuelve colisiones de slug. Los cuatro nombres son distintos, y `TestSembrarGeneraSlugsUnicos` lo verifica: si alguien agrega un homónimo, el test lo atrapa.

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/repository/memory/ -v`
Expected: PASS en todos.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/repository/memory/semilla.go apps/api/internal/repository/memory/semilla_test.go
git commit -m "feat(repository): seed de desarrollo con los datos del prototipo

Cuatro profesionales adaptados de legacy/prototype. Los precios pasan
de pesos a centavos. Solo se carga con APP_ENV=development.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 13: Configuración y composition root

El binario que corre de verdad.

**Files:**
- Create: `apps/api/internal/config/config.go`
- Test: `apps/api/internal/config/config_test.go`
- Modify: `apps/api/cmd/api/main.go` (reemplaza el placeholder de la Task 1)
- Create: `apps/api/.env.example`

**Interfaces:**
- Consumes: todo lo anterior
- Produces:
  - `config.Config{Puerto string; Entorno string; NivelLog slog.Level; TimeoutApagado time.Duration}`
  - `func config.Cargar() (Config, error)`
  - `func (Config) EsDesarrollo() bool`
  - El binario `apps/api/cmd/api`

- [ ] **Step 1: Escribir los tests de configuración**

Archivo `apps/api/internal/config/config_test.go`:

```go
package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestCargarDefaults(t *testing.T) {
	// t.Setenv restaura el entorno al terminar el test
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar devolvió error: %v", err)
	}

	if cfg.Puerto != "8080" {
		t.Errorf("Puerto = %q, se esperaba 8080", cfg.Puerto)
	}
	if cfg.Entorno != "development" {
		t.Errorf("Entorno = %q, se esperaba development", cfg.Entorno)
	}
	if cfg.NivelLog != slog.LevelInfo {
		t.Errorf("NivelLog = %v, se esperaba info", cfg.NivelLog)
	}
	if cfg.TimeoutApagado != 10*time.Second {
		t.Errorf("TimeoutApagado = %v, se esperaba 10s", cfg.TimeoutApagado)
	}
	if !cfg.EsDesarrollo() {
		t.Error("con APP_ENV vacío tenía que ser desarrollo")
	}
}

func TestCargarDesdeElEntorno(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar devolvió error: %v", err)
	}

	if cfg.Puerto != "9000" {
		t.Errorf("Puerto = %q", cfg.Puerto)
	}
	if cfg.NivelLog != slog.LevelDebug {
		t.Errorf("NivelLog = %v, se esperaba debug", cfg.NivelLog)
	}
	if cfg.TimeoutApagado != 30*time.Second {
		t.Errorf("TimeoutApagado = %v", cfg.TimeoutApagado)
	}
	if cfg.EsDesarrollo() {
		t.Error("con APP_ENV=production no tenía que ser desarrollo")
	}
}

func TestCargarFallaRapidoConValoresInvalidos(t *testing.T) {
	casos := []struct {
		nombre string
		clave  string
		valor  string
	}{
		{"puerto no numerico", "PORT", "ocho-mil"},
		{"puerto fuera de rango", "PORT", "99999"},
		{"nivel de log desconocido", "LOG_LEVEL", "verbose"},
		{"timeout mal formado", "SHUTDOWN_TIMEOUT", "diez segundos"},
		{"entorno desconocido", "APP_ENV", "staging-raro"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(caso.clave, caso.valor)
			// mejor no arrancar que arrancar mal configurado
			if _, err := Cargar(); err == nil {
				t.Errorf("%s=%q debía fallar y no falló", caso.clave, caso.valor)
			}
		})
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/config/ -v`
Expected: FAIL con `undefined: Cargar`

- [ ] **Step 3: Implementar la configuración**

Archivo `apps/api/internal/config/config.go`:

```go
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	EntornoDesarrollo = "development"
	EntornoProduccion = "production"
)

// Config es todo lo que el binario necesita saber del entorno. Sin librería:
// son cuatro variables y un struct.
type Config struct {
	Puerto         string
	Entorno        string
	NivelLog       slog.Level
	TimeoutApagado time.Duration
}

func (c Config) EsDesarrollo() bool {
	return c.Entorno == EntornoDesarrollo
}

// Cargar lee el entorno y falla si algo está mal.
//
// Falla rápido a propósito: un servidor que arranca con una configuración
// inválida es peor que uno que no arranca, porque el problema aparece más
// tarde y en otro lado.
func Cargar() (Config, error) {
	cfg := Config{
		Puerto:         leerEntorno("PORT", "8080"),
		Entorno:        leerEntorno("APP_ENV", EntornoDesarrollo),
		TimeoutApagado: 10 * time.Second,
	}

	puerto, err := strconv.Atoi(cfg.Puerto)
	if err != nil || puerto < 1 || puerto > 65535 {
		return Config{}, fmt.Errorf("PORT inválido: %q", cfg.Puerto)
	}

	if cfg.Entorno != EntornoDesarrollo && cfg.Entorno != EntornoProduccion {
		return Config{}, fmt.Errorf("APP_ENV inválido: %q (debe ser %s o %s)", cfg.Entorno, EntornoDesarrollo, EntornoProduccion)
	}

	nivel, err := parsearNivelLog(leerEntorno("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.NivelLog = nivel

	if crudo := os.Getenv("SHUTDOWN_TIMEOUT"); crudo != "" {
		d, err := time.ParseDuration(crudo)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT inválido: %q (ejemplo válido: 30s)", crudo)
		}
		cfg.TimeoutApagado = d
	}

	return cfg, nil
}

func leerEntorno(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}

func parsearNivelLog(crudo string) (slog.Level, error) {
	switch crudo {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL inválido: %q (debe ser debug, info, warn o error)", crudo)
	}
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/config/ -v`
Expected: PASS en los tres tests.

- [ ] **Step 5: Escribir el composition root**

Archivo `apps/api/cmd/api/main.go` (reemplaza completo el placeholder de la Task 1):

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/config"
	"github.com/joaquinfochoa/Salud/apps/api/internal/handler"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

func main() {
	if err := ejecutar(); err != nil {
		fmt.Fprintln(os.Stderr, "error fatal:", err)
		os.Exit(1)
	}
}

func ejecutar() error {
	cfg, err := config.Cargar()
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.NivelLog,
	})))

	// El cableado de dependencias, explícito y de arriba abajo. Sin
	// anotaciones ni contenedor: no hay magia que debuggear a las 3 de la
	// mañana.
	//
	// Migrar a PostgreSQL es cambiar esta línea por
	// postgres.NuevoProfesional(db). Nada más.
	repo := memory.NuevoProfesional()

	if cfg.EsDesarrollo() {
		if err := memory.Sembrar(context.Background(), repo); err != nil {
			return fmt.Errorf("cargando el seed: %w", err)
		}
		slog.Info("seed de desarrollo cargado")
	}

	svc := service.NuevoProfesional(repo)
	router := handler.NuevoRouter(handler.NuevoProfesional(svc))

	srv := &http.Server{
		Addr:    ":" + cfg.Puerto,
		Handler: router,
		// Los valores por defecto de http.Server son cero, o sea sin límite:
		// una conexión lenta puede quedarse tomada para siempre.
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	errServidor := make(chan error, 1)
	go func() {
		slog.Info("servidor escuchando", "direccion", srv.Addr, "entorno", cfg.Entorno)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errServidor <- err
		}
	}()

	select {
	case err := <-errServidor:
		return fmt.Errorf("el servidor falló: %w", err)

	case <-ctx.Done():
		// Apagado gracioso: sin esto, cada deploy corta los requests que
		// están a mitad de camino.
		slog.Info("apagando", "timeout", cfg.TimeoutApagado)

		ctxApagado, cancelar := context.WithTimeout(context.Background(), cfg.TimeoutApagado)
		defer cancelar()

		if err := srv.Shutdown(ctxApagado); err != nil {
			return fmt.Errorf("apagado forzado: %w", err)
		}
		slog.Info("apagado limpio")
		return nil
	}
}
```

- [ ] **Step 6: Escribir el `.env.example`**

Archivo `apps/api/.env.example`:

```dotenv
# Puerto del servidor HTTP
PORT=8080

# development | production
# En development se carga el seed de profesionales de prueba.
APP_ENV=development

# debug | info | warn | error
LOG_LEVEL=info

# Cuánto espera a que terminen los requests en curso antes de apagar
SHUTDOWN_TIMEOUT=10s
```

- [ ] **Step 7: Verificar que el binario arranca y responde**

Run:
```bash
cd apps/api
go build ./...
APP_ENV=development go run ./cmd/api &
sleep 2
curl -s http://localhost:8080/healthz
curl -s "http://localhost:8080/api/v1/profesionales?busqueda=gonzalez" | head -c 400
kill %1
```

Expected:
- `{"estado":"ok"}`
- Un JSON con `"datos":[...]` conteniendo a Martín González y `"total":1`.
- En los logs, líneas JSON con `"msg":"peticion"` y `"idPeticion"`.

En PowerShell, el equivalente:
```powershell
cd apps/api
go build ./...
$env:APP_ENV="development"; Start-Job { go run ./cmd/api }
Start-Sleep 3
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod "http://localhost:8080/api/v1/profesionales?busqueda=gonzalez"
Get-Job | Stop-Job; Get-Job | Remove-Job
```

- [ ] **Step 8: Verificar el apagado gracioso**

Arrancar el servidor en una terminal, mandarle Ctrl+C y confirmar que en los logs aparecen `"msg":"apagando"` y después `"msg":"apagado limpio"`, y que el proceso termina con código 0.

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/config/ apps/api/cmd/ apps/api/.env.example
git commit -m "feat(api): configuración por entorno y composition root

El cableado de dependencias es explícito y vive entero en main.go.
Migrar a PostgreSQL es cambiar la línea de memory.NuevoProfesional().

La configuración falla rápido: arrancar mal configurado es peor que no
arrancar. El servidor tiene timeouts explícitos porque los de
http.Server son cero, o sea sin límite, y apaga graciosamente para no
cortar requests en cada deploy.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 14: Herramientas de build, contenedor e integración continua

**Files:**
- Create: `apps/api/Makefile`
- Create: `apps/api/Dockerfile`
- Create: `apps/api/.dockerignore`
- Create: `apps/api/.golangci.yml`
- Create: `apps/api/README.md`
- Create: `.github/workflows/api.yml`

**Interfaces:**
- Consumes: el binario de la Task 13
- Produces: `make test`, `make lint`, `make docker-build` y el pipeline de CI

- [ ] **Step 1: Escribir el Makefile**

Archivo `apps/api/Makefile`:

```makefile
.PHONY: run test test-race lint fmt build docker-build tidy check

run:
	APP_ENV=development go run ./cmd/api

test:
	go test ./...

test-race:
	go test ./... -race

# Cobertura para mirar, no para poner un umbral: se prueban las reglas,
# no las líneas.
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	go vet ./...

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

docker-build:
	docker build -t salud-api:local .

# Lo que tiene que pasar antes de un commit
check: fmt test-race lint
```

- [ ] **Step 2: Verificar el Makefile**

Run: `cd apps/api && make test-race`
Expected: PASS en todos los paquetes.

Si `make` no está disponible en Windows, los comandos se corren directo: `go test ./... -race`.

- [ ] **Step 3: Escribir la configuración del linter**

Archivo `apps/api/.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - errcheck      # errores ignorados
    - govet
    - ineffassign
    - staticcheck
    - unused
    - bodyclose     # response bodies sin cerrar
    - errorlint     # comparaciones de error sin errors.Is/As
    - gosec
    - misspell
    - revive
    - unconvert

  settings:
    gosec:
      excludes:
        # G104 se solapa con errcheck
        - G104

  exclusions:
    rules:
      # En los tests se ignoran errores a propósito para armar estados
      - path: _test\.go
        linters: [errcheck, gosec]

formatters:
  enable:
    - gofmt
    - goimports
```

- [ ] **Step 4: Correr el linter y arreglar lo que aparezca**

Run: `cd apps/api && golangci-lint run ./...`
Expected: sin hallazgos.

Si `golangci-lint` no está instalado: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

Hallazgos probables y cómo se resuelven:
- `errcheck` sobre `json.NewEncoder(w).Encode(...)`: ya está silenciado con `_ =` en `problema.go`.
- `bodyclose` en los tests HTTP: ya se cierran con `t.Cleanup`.
- `errorlint` sobre comparaciones con `==`: usar `errors.Is`.

- [ ] **Step 5: Escribir el `.dockerignore`**

Archivo `apps/api/.dockerignore`:

```
bin/
coverage.out
.env
*_test.go
Makefile
README.md
```

- [ ] **Step 6: Escribir el Dockerfile**

Archivo `apps/api/Dockerfile`:

```dockerfile
# Compilación
FROM golang:1.24-alpine AS build

WORKDIR /src

# Las dependencias se copian primero para que Docker cachee esta capa: el
# go.mod cambia mucho menos que el código.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produce un binario estático, requisito para correr sobre
# distroless. -w -s saca la información de debug y baja el tamaño.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# Imagen final
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /bin/api /api

# nonroot ya viene definido en la imagen base. Un contenedor que corre como
# root es una escalada de privilegios servida.
USER nonroot:nonroot

EXPOSE 8080
ENV APP_ENV=production

ENTRYPOINT ["/api"]
```

- [ ] **Step 7: Construir la imagen y verificar que corre**

Run:
```bash
cd apps/api
docker build -t salud-api:local .
docker images salud-api:local --format "{{.Size}}"
docker run --rm -d -p 8080:8080 --name salud-api-test salud-api:local
sleep 2
curl -s http://localhost:8080/healthz
docker stop salud-api-test
```

Expected:
- La imagen pesa menos de 25 MB.
- `{"estado":"ok"}`.
- El listado viene vacío: en `production` no se carga el seed.

- [ ] **Step 8: Escribir el README del backend**

Archivo `apps/api/README.md`:

```markdown
# Salud API

Backend en Go. CRUD de profesionales, sin base de datos.

## Correr

```bash
make run                    # con seed de desarrollo
curl localhost:8080/healthz
```

Configuración en `.env.example`.

## Arquitectura

Cuatro capas. `domain` no importa nada del proyecto.

```
handler ──▶ service ──▶ repository (interfaz)
   │           │              ▲
   └───────────┴──────────────┤
               ▼              │
            domain      repository/memory
```

| Carpeta | Equivalente en Spring / ASP.NET |
|---|---|
| `internal/handler` | `@RestController` |
| `internal/service` | `@Service` |
| `internal/repository/memory` | `@Repository` |
| `internal/domain` | Entidades y value objects |
| `cmd/api/main.go` | El contenedor de DI, pero explícito |

## Migrar a PostgreSQL

Implementar `repository.Profesional` en `internal/repository/postgres/` y
cambiar una línea de `cmd/api/main.go`. Nada más.

## Contrato

`api/openapi.yaml` es la fuente de verdad. Se escribe antes que los handlers.

Para navegarlo con Swagger UI:

```bash
docker run --rm -p 8081:8080 \
  -e SWAGGER_JSON=/spec/openapi.yaml \
  -v "$(pwd)/api:/spec" swaggerapi/swagger-ui
```

## Comandos

| Comando | Qué hace |
|---|---|
| `make run` | Levanta el servidor con seed |
| `make test` | Corre los tests |
| `make test-race` | Tests con detector de carreras |
| `make lint` | golangci-lint |
| `make check` | fmt + test-race + lint. Correr antes de commitear. |
| `make docker-build` | Imagen local |

## Convenciones

- Sin mocks. El repositorio en memoria es el doble de test.
- Dinero en `int64` de centavos, nunca `float64`.
- Todo lo que escribimos nosotros va en español: tipos, funciones, campos,
  constantes, comentarios, mensajes y los nombres de campo del JSON. Quedan
  en inglés los paquetes (`domain`, `service`, ...), `String()` y `Error()`,
  las variables de entorno y las claves `type`/`title`/`status`/`detail`
  del RFC 7807.
- Una sola dependencia externa: `github.com/google/uuid`.
```

- [ ] **Step 9: Escribir el pipeline de CI**

Archivo `.github/workflows/api.yml` (en la raíz del repo, no en `apps/api`):

```yaml
name: api

on:
  push:
    branches: [main, refactor-gian]
    paths: ['apps/api/**', '.github/workflows/api.yml']
  pull_request:
    paths: ['apps/api/**', '.github/workflows/api.yml']

defaults:
  run:
    working-directory: apps/api

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: apps/api/go.sum

      - name: Verificar que go.mod está ordenado
        run: |
          go mod tidy
          git diff --exit-code go.mod go.sum

      - name: Tests con detector de carreras
        run: go test ./... -race

      - name: Verificar que el dominio no depende de otras capas
        run: |
          deps=$(go list -deps ./internal/domain | grep 'joaquinfochoa' | grep -v 'internal/domain$' || true)
          if [ -n "$deps" ]; then
            echo "internal/domain importa otras capas del proyecto:"
            echo "$deps"
            exit 1
          fi

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: apps/api/go.sum
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          working-directory: apps/api

  openapi:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - name: Validar el contrato OpenAPI
        run: npx --yes @redocly/cli@latest lint apps/api/api/openapi.yaml
        working-directory: .

  docker:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v4
      - name: Construir la imagen
        run: docker build -t salud-api:ci .
```

- [ ] **Step 10: Verificar el pipeline localmente**

Run:
```bash
cd apps/api
go mod tidy && git diff --exit-code go.mod go.sum
go test ./... -race
golangci-lint run ./...
go list -deps ./internal/domain | grep joaquinfochoa
npx --yes @redocly/cli@latest lint api/openapi.yaml
```

Expected:
- `go mod tidy` no deja cambios.
- Todos los tests en verde.
- El linter sin hallazgos.
- `go list -deps` devuelve una sola línea: el propio paquete `domain`.
- Redocly valida el contrato sin errores.

- [ ] **Step 11: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/Makefile apps/api/Dockerfile apps/api/.dockerignore \
        apps/api/.golangci.yml apps/api/README.md .github/
git commit -m "chore(api): build, contenedor e integración continua

Imagen distroless de menos de 25 MB, sin CGO y corriendo como nonroot:
es la razón por la que se eligió Go para auto-hospedar en Argentina.

El CI verifica que internal/domain no importe otras capas: la regla
arquitectónica es ejecutable, no un acuerdo verbal.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 15: Verificación final contra los criterios de aceptación

**Files:** ninguno nuevo. Es la pasada de cierre.

- [ ] **Step 1: Toda la suite con detector de carreras**

Run: `cd apps/api && go test ./... -race -count=1`
Expected: `ok` en `internal/domain`, `internal/repository/memory`, `internal/service`, `internal/handler` e `internal/config`. Sin `WARNING: DATA RACE`.

El `-count=1` evita que Go devuelva resultados cacheados.

- [ ] **Step 2: El linter**

Run: `cd apps/api && golangci-lint run ./...`
Expected: sin hallazgos.

- [ ] **Step 3: El dominio no depende de nadie**

Run: `cd apps/api && go list -deps ./internal/domain | grep joaquinfochoa`
Expected: exactamente una línea, `github.com/joaquinfochoa/Salud/apps/api/internal/domain`.

- [ ] **Step 4: Los siete endpoints, contra el servidor real**

Levantar el servidor con `make run` y correr:

```bash
BASE=http://localhost:8080/api/v1/profesionales

# healthz
curl -s http://localhost:8080/healthz

# listado con seed
curl -s "$BASE" | head -c 300

# búsqueda sin acentos: tiene que encontrar a González
curl -s "$BASE?busqueda=gonzalez" | head -c 300

# alta
ID=$(curl -s -X POST "$BASE" -H 'Content-Type: application/json' -d '{
  "nombre":"Ana","apellido":"Pérez","matricula":"MP 55.123",
  "especialidad":"odontologia","bio":"Odontóloga general.",
  "precioConsultaCentavos":1800000,"modalidades":["presencial"],
  "zona":"CABA","obrasSociales":["OSDE"]
}' | python -c "import sys,json; print(json.load(sys.stdin)['id'])")

# lectura por id y por slug
curl -s "$BASE/$ID" | head -c 200
curl -s "$BASE/por-slug/ana-perez" | head -c 200

# edición
curl -s -X PUT "$BASE/$ID" -H 'Content-Type: application/json' -d '{
  "nombre":"Ana","apellido":"Pérez","matricula":"MP 55.123",
  "especialidad":"odontologia","bio":"Bio editada.",
  "precioConsultaCentavos":1800000,"modalidades":["presencial","telemedicina"],
  "zona":"GBA Norte","obrasSociales":["OSDE"]
}' | head -c 200

# baja y reactivación
curl -s -o /dev/null -w "DELETE: %{http_code}\n" -X DELETE "$BASE/$ID"
curl -s -o /dev/null -w "DELETE otra vez: %{http_code}\n" -X DELETE "$BASE/$ID"
curl -s -o /dev/null -w "reactivar: %{http_code}\n" -X POST "$BASE/$ID/reactivar"

# los códigos de error
curl -s -o /dev/null -w "JSON roto: %{http_code}\n" -X POST "$BASE" -H 'Content-Type: application/json' -d '{'
curl -s -o /dev/null -w "datos inválidos: %{http_code}\n" -X POST "$BASE" -H 'Content-Type: application/json' -d '{"nombre":"","apellido":"","matricula":"x","especialidad":"y","precioConsultaCentavos":0,"modalidades":[],"zona":""}'
curl -s -o /dev/null -w "id inexistente: %{http_code}\n" "$BASE/6ba7b810-9dad-11d1-80b4-00c04fd430c8"
curl -s -o /dev/null -w "id mal formado: %{http_code}\n" "$BASE/no-es-uuid"
```

Expected: `DELETE: 204`, `DELETE otra vez: 204`, `reactivar: 200`, `JSON roto: 400`, `datos inválidos: 422`, `id inexistente: 404`, `id mal formado: 400`.

- [ ] **Step 5: La imagen de Docker arranca y responde**

Run:
```bash
cd apps/api
docker build -t salud-api:local .
docker run --rm -d -p 8080:8080 --name salud-check salud-api:local
sleep 2
curl -s http://localhost:8080/healthz
docker stop salud-check
```

Expected: `{"estado":"ok"}`.

- [ ] **Step 6: El punto de cambio a PostgreSQL es una sola línea**

Verificación por lectura, sin escribir código: buscar todas las menciones a `memory.` fuera de su propio paquete y de los tests.

Run: `cd apps/api && grep -rn "repository/memory" --include="*.go" . | grep -v "_test.go" | grep -v "internal/repository/memory/"`
Expected: exactamente dos líneas, ambas en `cmd/api/main.go` — el import y la llamada a `memory.NuevoProfesional()`. La de `memory.Sembrar` es la tercera y también es esperada. Si aparece cualquier otra, alguna capa se acopló a la implementación y hay que revisarla.

- [ ] **Step 7: Actualizar el spec con lo que se aprendió**

Si durante la implementación algo se decidió distinto de lo que dice `docs/superpowers/specs/2026-08-21-professional-crud-go-design.md`, corregir el spec. Un spec que quedó desactualizado engaña a quien lo lea después.

- [ ] **Step 8: Commit final**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add -A
git commit -m "chore(api): verificación final de los criterios de aceptación

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Lo que queda fuera, y cuándo entra

Registrado también en la sección 10 del spec.

| Pendiente | Disparador |
|---|---|
| Autenticación y autorización | Spec propia. Obligatorio antes de exponer la API a internet. |
| CORS | Cuando exista `apps/web` |
| `Patient` y `Appointment` | Después de que este esqueleto esté cerrado |
| PostgreSQL | Cuando el modelo de dominio deje de moverse |
| Integración con REFEPS | Spec propia. Hoy `verificacion` queda en `pendiente` para siempre. |
| `EstadoSuspendido` | Cuando REFEPS pueda disparar una suspensión |
| Endpoint de purga (Ley 25.326 art. 16) | Cuando el abogado defina qué se está obligado a conservar |
| Rating, reseñas, horarios, coseguros | Son entidades propias, no campos |
| Cliente TypeScript generado del OpenAPI | Etapa del frontend |
