# CRUD de Professional en Go — Plan de Implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Construir el backend en Go con un CRUD completo de `Professional`, sin base de datos, con arquitectura en capas y el punto de cambio a PostgreSQL aislado en una interfaz.

**Architecture:** Cuatro capas — `handler → service → repository → domain` — donde `domain` no importa nada del proyecto. El repositorio en memoria implementa una interfaz que después implementará PostgreSQL sin tocar el resto del código. El cableado de dependencias es explícito en `cmd/api/main.go`, sin contenedor de inyección.

**Tech Stack:** Go 1.24+, `net/http` de la stdlib (ServeMux con patrones de Go 1.22+), `log/slog`, `testing` + `net/http/httptest`. Una única dependencia externa: `github.com/google/uuid`.

## Global Constraints

- **Module path:** `github.com/joaquinfochoa/Salud/apps/api`
- **Go mínimo:** 1.24 (el `ServeMux` con `"GET /path/{id}"` requiere 1.22+)
- **Dependencias externas permitidas:** únicamente `github.com/google/uuid`. Cualquier otro `go get` es un error del plan — preguntar antes.
- **Idioma del código:** identificadores, funciones y comentarios técnicos en inglés. En español solo los términos del dominio sin traducción fiel: `Matricula`, `Especialidad`, `Modalidad`, `ObraSocial`, `Zona`, `Coseguro`. Los mensajes de error que ve el usuario, siempre en español.
- **Dinero:** `int64` en centavos, tipo `domain.Money`. Nunca `float64`, nunca en ningún lugar.
- **JSON:** camelCase. El precio viaja como `consultaPriceCents` (entero).
- **Errores HTTP:** `application/problem+json` (RFC 7807).
- **Sin mocks.** El repositorio en memoria es el doble de test. Si un test parece necesitar un mock, la frontera está mal dibujada — parar y preguntar.
- **`internal/domain` no importa ningún otro paquete del proyecto.** Es verificable y es criterio de aceptación.
- **Comentarios:** en español, explicando el *por qué*, no el *qué*. No comentar lo obvio.

---

## Estructura de archivos

| Archivo | Responsabilidad |
|---|---|
| `apps/api/go.mod` | Módulo y dependencias |
| `apps/api/internal/domain/money.go` | Tipo `Money` y su formato |
| `apps/api/internal/domain/text.go` | `Normalize` y `Slugify` — compartidos por slug y búsqueda |
| `apps/api/internal/domain/enums.go` | `Especialidad`, `Modalidad`, `Status`, `VerificationStatus` |
| `apps/api/internal/domain/matricula.go` | Value object `Matricula` y su parser |
| `apps/api/internal/domain/errors.go` | Errores centinela y `ValidationError` |
| `apps/api/internal/domain/professional.go` | La entidad, su constructor y sus transiciones |
| `apps/api/internal/repository/professional.go` | La interfaz y el `Filter` |
| `apps/api/internal/repository/memory/professional.go` | Implementación en memoria |
| `apps/api/internal/repository/memory/seed.go` | Datos de desarrollo |
| `apps/api/internal/service/professional.go` | Casos de uso |
| `apps/api/internal/handler/problem.go` | Errores de dominio → HTTP |
| `apps/api/internal/handler/dto.go` | Structs de request y response |
| `apps/api/internal/handler/professional.go` | Controllers |
| `apps/api/internal/handler/middleware.go` | Request ID, logging, recover |
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

## Task 2: `Money` y normalización de texto

Dos piezas puras, sin dependencias, que el resto del dominio usa.

**Files:**
- Create: `apps/api/internal/domain/money.go`
- Create: `apps/api/internal/domain/text.go`
- Test: `apps/api/internal/domain/money_test.go`
- Test: `apps/api/internal/domain/text_test.go`

**Interfaces:**
- Consumes: nada
- Produces:
  - `domain.Money` (`int64`), `func (Money) String() string`
  - `func domain.Normalize(string) string`
  - `func domain.Slugify(string) string`

- [ ] **Step 1: Escribir los tests de `Money`**

Archivo `apps/api/internal/domain/money_test.go`:

```go
package domain

import "testing"

func TestMoneyString(t *testing.T) {
	tests := []struct {
		name string
		in   Money
		want string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("Money(%d).String() = %q, se esperaba %q", int64(tt.in), got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run TestMoneyString -v`
Expected: FAIL con `undefined: Money`

- [ ] **Step 3: Implementar `Money`**

Archivo `apps/api/internal/domain/money.go`:

```go
package domain

import (
	"strconv"
	"strings"
)

// Money representa un monto en centavos.
//
// Nunca usar float para dinero: float64 no puede representar 0,10 de forma
// exacta, y este sistema va a cobrar consultas y liquidar honorarios. El tipo
// propio además impide sumar un precio con una cantidad por accidente.
type Money int64

// String formatea el monto con la convención argentina: punto para miles,
// coma para decimales. Money(1200000) → "$12.000,00"
func (m Money) String() string {
	negative := m < 0
	if negative {
		m = -m
	}

	pesos := int64(m) / 100
	cents := int64(m) % 100

	digits := strconv.FormatInt(pesos, 10)
	var b strings.Builder
	for i := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(digits[i])
	}

	centsStr := strconv.FormatInt(cents, 10)
	if cents < 10 {
		centsStr = "0" + centsStr
	}

	out := "$" + b.String() + "," + centsStr
	if negative {
		return "-" + out
	}
	return out
}
```

- [ ] **Step 4: Correr el test y verificar que pasa**

Run: `cd apps/api && go test ./internal/domain/ -run TestMoneyString -v`
Expected: PASS, los ocho subtests en verde.

- [ ] **Step 5: Escribir los tests de normalización**

Archivo `apps/api/internal/domain/text_test.go`:

```go
package domain

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"minusculas", "GONZÁLEZ", "gonzalez"},
		{"acentos", "Martín González", "martin gonzalez"},
		{"enie", "Muñoz", "munoz"},
		{"dieresis", "Agüero", "aguero"},
		{"todas las vocales", "áéíóú", "aeiou"},
		{"recorta espacios", "  Ana  ", "ana"},
		{"vacio", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.in); got != tt.want {
				t.Errorf("Normalize(%q) = %q, se esperaba %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.in); got != tt.want {
				t.Errorf("Slugify(%q) = %q, se esperaba %q", tt.in, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 6: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestNormalize|TestSlugify' -v`
Expected: FAIL con `undefined: Normalize` y `undefined: Slugify`

- [ ] **Step 7: Implementar la normalización**

Archivo `apps/api/internal/domain/text.go`:

```go
package domain

import (
	"strings"
	"unicode"
)

// Normalize baja a minúsculas y saca acentos y eñes, para poder comparar
// "González" con "gonzalez".
//
// Lo usan dos cosas: la generación del slug y el filtro de búsqueda del
// listado. Una sola función, un solo lugar donde arreglarla. En un producto
// argentino, una búsqueda que distingue acentos es una búsqueda rota.
func Normalize(s string) string {
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

// Slugify genera la parte legible de la URL pública del profesional.
// "Íñigo Muñoz Ríos" → "inigo-munoz-rios"
//
// No se ocupa de la unicidad: eso necesita mirar los demás profesionales y
// por lo tanto vive en el servicio.
func Slugify(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	pendingHyphen := false
	for _, r := range Normalize(s) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			if pendingHyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingHyphen = false
			b.WriteRune(r)
		case unicode.IsSpace(r), r == '-', r == '_':
			pendingHyphen = true
		}
		// el resto (puntos, comas, símbolos) se descarta sin separar
	}
	return b.String()
}
```

- [ ] **Step 8: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -v`
Expected: PASS en `TestMoneyString`, `TestNormalize` y `TestSlugify`.

Verificar el caso `"Dr. Juan Pérez"`: el punto se descarta sin marcar separación, pero el espacio que le sigue sí la marca. Resultado `dr-juan-perez`, no `dr--juan-perez`.

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): Money en centavos y normalización de texto

Money es int64 de centavos con formato argentino. Normalize y Slugify
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
  - `func domain.ParseMatricula(string) (Matricula, error)`
  - `func (Matricula) String() string`, `func (Matricula) IsZero() bool`
  - `domain.Especialidad` con `EspecialidadPsicologia`, `EspecialidadKinesiologia`, `EspecialidadOdontologia`
  - `domain.Modalidad` con `ModalidadTelemedicina`, `ModalidadPresencial`, `ModalidadDomicilio`
  - `domain.Status` con `StatusActive`, `StatusInactive`
  - `domain.VerificationStatus` con `VerificationPending`, `VerificationVerified`, `VerificationRejected`
  - Cada uno con método `Valid() bool`

- [ ] **Step 1: Escribir los tests de `ParseMatricula`**

Archivo `apps/api/internal/domain/matricula_test.go`:

```go
package domain

import "testing"

func TestParseMatriculaValida(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantTipo   MatriculaTipo
		wantNumero string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := ParseMatricula(tt.in)
			if err != nil {
				t.Fatalf("ParseMatricula(%q) devolvió error: %v", tt.in, err)
			}
			if m.Tipo != tt.wantTipo {
				t.Errorf("tipo = %q, se esperaba %q", m.Tipo, tt.wantTipo)
			}
			if m.Numero != tt.wantNumero {
				t.Errorf("numero = %q, se esperaba %q", m.Numero, tt.wantNumero)
			}
		})
	}
}

func TestParseMatriculaInvalida(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"vacia", ""},
		{"solo el tipo", "MN"},
		{"sin tipo", "98234"},
		{"tipo desconocido", "XX 98234"},
		{"numero con letras", "MN 98A34"},
		{"mas de diez digitos", "MN 12345678901"},
		{"solo espacios", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseMatricula(tt.in); err == nil {
				t.Errorf("ParseMatricula(%q) debía fallar y no falló", tt.in)
			}
		})
	}
}

func TestMatriculaString(t *testing.T) {
	m, err := ParseMatricula("m.n. 98.234")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// distintas formas de escribir la misma matrícula tienen que converger
	// a una sola representación, o la unicidad no sirve de nada
	if got := m.String(); got != "MN 98234" {
		t.Errorf("String() = %q, se esperaba %q", got, "MN 98234")
	}
}

func TestMatriculaIsZero(t *testing.T) {
	var zero Matricula
	if !zero.IsZero() {
		t.Error("la matrícula vacía debía ser IsZero")
	}

	m, _ := ParseMatricula("MN 1")
	if m.IsZero() {
		t.Error("una matrícula parseada no debía ser IsZero")
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run TestMatricula -v`
Expected: FAIL con `undefined: ParseMatricula`

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

const maxMatriculaDigits = 10

// Matricula es la identidad profesional de una persona: es el único dato que
// la ata a una habilitación real y es sobre lo que se apoya toda la confianza
// del producto.
type Matricula struct {
	Tipo   MatriculaTipo
	Numero string
}

var matriculaCleaner = strings.NewReplacer(".", "", " ", "", "-", "", "/", "")

// ParseMatricula acepta las formas que se usan en la práctica —"MN 98.234",
// "M.N. 45321", "mn98234", "MP 12345"— y las normaliza a "MN 98234".
//
// La validación es deliberadamente laxa. Las matrículas argentinas varían por
// jurisdicción y por profesión, y rechazar a un profesional real es peor error
// que aceptar un número raro: el que queda afuera no vuelve. La verificación
// seria llega cuando exista la integración con REFEPS.
func ParseMatricula(s string) (Matricula, error) {
	clean := matriculaCleaner.Replace(strings.ToUpper(s))

	if len(clean) < 3 {
		return Matricula{}, errors.New("debe tener tipo (MN o MP) y número")
	}

	tipo := MatriculaTipo(clean[:2])
	if tipo != MatriculaNacional && tipo != MatriculaProvincial {
		return Matricula{}, errors.New("el tipo debe ser MN (nacional) o MP (provincial)")
	}

	numero := clean[2:]
	if len(numero) > maxMatriculaDigits {
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

func (m Matricula) IsZero() bool {
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

func TestEspecialidadValid(t *testing.T) {
	validas := []Especialidad{
		EspecialidadPsicologia,
		EspecialidadKinesiologia,
		EspecialidadOdontologia,
	}
	for _, e := range validas {
		if !e.Valid() {
			t.Errorf("Especialidad(%q) debía ser válida", e)
		}
	}

	invalidas := []Especialidad{"", "cardiologia", "Psicologia", "PSICOLOGIA"}
	for _, e := range invalidas {
		if e.Valid() {
			t.Errorf("Especialidad(%q) no debía ser válida", e)
		}
	}
}

func TestModalidadValid(t *testing.T) {
	validas := []Modalidad{ModalidadTelemedicina, ModalidadPresencial, ModalidadDomicilio}
	for _, m := range validas {
		if !m.Valid() {
			t.Errorf("Modalidad(%q) debía ser válida", m)
		}
	}

	invalidas := []Modalidad{"", "online", "Presencial"}
	for _, m := range invalidas {
		if m.Valid() {
			t.Errorf("Modalidad(%q) no debía ser válida", m)
		}
	}
}

func TestStatusValid(t *testing.T) {
	if !StatusActive.Valid() || !StatusInactive.Valid() {
		t.Error("active e inactive debían ser válidos")
	}
	if Status("suspended").Valid() {
		t.Error("suspended todavía no existe: no debía ser válido")
	}
}

func TestVerificationStatusValid(t *testing.T) {
	validos := []VerificationStatus{VerificationPending, VerificationVerified, VerificationRejected}
	for _, v := range validos {
		if !v.Valid() {
			t.Errorf("VerificationStatus(%q) debía ser válido", v)
		}
	}
	if VerificationStatus("unknown").Valid() {
		t.Error("unknown no debía ser válido")
	}
}
```

- [ ] **Step 6: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestEspecialidad|TestModalidad|TestStatus|TestVerification' -v`
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
// Agregar una cuarta es una constante y un caso más en Valid().
type Especialidad string

const (
	EspecialidadPsicologia   Especialidad = "psicologia"
	EspecialidadKinesiologia Especialidad = "kinesiologia"
	EspecialidadOdontologia  Especialidad = "odontologia"
)

func (e Especialidad) Valid() bool {
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

func (m Modalidad) Valid() bool {
	switch m {
	case ModalidadTelemedicina, ModalidadPresencial, ModalidadDomicilio:
		return true
	}
	return false
}

// Status dice si el profesional opera hoy en la plataforma.
//
// No confundir con VerificationStatus: son dos ejes distintos. Un profesional
// puede estar verificado y de licencia, o recién anotado y sin verificar.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive:
		return true
	}
	return false
}

// VerificationStatus dice si la matrícula fue verificada contra el mundo real.
// Por ahora todos nacen en pending: la integración con REFEPS es una etapa
// posterior y nada la mueve automáticamente todavía.
type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationVerified VerificationStatus = "verified"
	VerificationRejected VerificationStatus = "rejected"
)

func (v VerificationStatus) Valid() bool {
	switch v {
	case VerificationPending, VerificationVerified, VerificationRejected:
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

ParseMatricula acepta los formatos reales del mercado argentino y los
normaliza a una forma canónica. La validación es laxa a propósito:
rechazar a un profesional real es peor que aceptar un número raro.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 4: Errores del dominio y la entidad `Professional`

El corazón del dominio: la invariante de que no se puede construir un profesional inválido.

**Files:**
- Create: `apps/api/internal/domain/errors.go`
- Create: `apps/api/internal/domain/professional.go`
- Test: `apps/api/internal/domain/professional_test.go`

**Interfaces:**
- Consumes: `Matricula`, `Especialidad`, `Modalidad`, `Status`, `VerificationStatus`, `Money`, `Slugify`, `Normalize` (Tasks 2 y 3)
- Produces:
  - `domain.ErrNotFound`, `domain.ErrMatriculaTaken`
  - `domain.FieldError{Field, Message string}`
  - `domain.ValidationError{Fields []FieldError}` con método `Error() string`
  - `domain.ProfessionalInput` — struct de entrada con todo en tipos primitivos
  - `domain.Professional` — la entidad
  - `func domain.NewProfessional(ProfessionalInput, time.Time) (Professional, error)`
  - `func (Professional) ApplyUpdate(ProfessionalInput, time.Time) (Professional, error)`
  - `func (Professional) Deactivate(time.Time) Professional`
  - `func (Professional) Reactivate(time.Time) Professional`
  - `func (Professional) Clone() Professional`
  - `func (Professional) FullName() string`

- [ ] **Step 1: Escribir `errors.go`**

Este archivo no lleva test propio: sus dos funciones se ejercitan enteras desde los tests de la entidad del paso 3.

Archivo `apps/api/internal/domain/errors.go`:

```go
package domain

import (
	"errors"
	"strings"
)

var (
	// ErrNotFound lo devuelve el repositorio cuando no existe el registro.
	ErrNotFound = errors.New("professional not found")

	// ErrMatriculaTaken lo devuelve el servicio: la matrícula es la única
	// identidad real de una persona en este sistema y no puede repetirse.
	ErrMatriculaTaken = errors.New("matricula already registered")
)

// FieldError señala un campo puntual. Las etiquetas JSON coinciden con el
// formato problem+json que espera el cliente.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError junta todos los campos inválidos de una sola pasada.
// Devolver solo el primero obliga al cliente a corregir de a uno, que es una
// experiencia horrible en un formulario de alta con nueve campos.
type ValidationError struct {
	Fields []FieldError
}

func (e ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		parts = append(parts, f.Field+": "+f.Message)
	}
	return "validación fallida — " + strings.Join(parts, "; ")
}

func (e *ValidationError) add(field, message string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Message: message})
}

func (e ValidationError) hasErrors() bool {
	return len(e.Fields) > 0
}
```

- [ ] **Step 2: Escribir los tests de la entidad**

Archivo `apps/api/internal/domain/professional_test.go`:

```go
package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

// validInput devuelve una entrada correcta. Cada test la copia y rompe un
// solo campo, así el que falla es siempre el campo bajo prueba.
func validInput() ProfessionalInput {
	return ProfessionalInput{
		FirstName:     "Martín",
		LastName:      "González",
		Matricula:     "MN 98.234",
		Especialidad:  "psicologia",
		Bio:           "Psicólogo clínico con orientación cognitivo-conductual.",
		ConsultaPrice: 1200000,
		Modalidades:   []string{"telemedicina", "presencial"},
		Zona:          "CABA",
		ObrasSociales: []string{"OSDE", "Swiss Medical"},
	}
}

func TestNewProfessionalValido(t *testing.T) {
	p, err := NewProfessional(validInput(), testNow)
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
	if p.ConsultaPrice != Money(1200000) {
		t.Errorf("ConsultaPrice = %d, se esperaba 1200000", p.ConsultaPrice)
	}
	if p.Status != StatusActive {
		t.Errorf("Status = %q, se esperaba active", p.Status)
	}
	// nadie nace verificado: la verificación es un acto contra REFEPS
	if p.Verification != VerificationPending {
		t.Errorf("Verification = %q, se esperaba pending", p.Verification)
	}
	if !p.CreatedAt.Equal(testNow) || !p.UpdatedAt.Equal(testNow) {
		t.Error("las marcas de tiempo debían ser el now recibido")
	}
	if p.DeactivatedAt != nil {
		t.Error("DeactivatedAt debía ser nil")
	}
}

func TestNewProfessionalCamposInvalidos(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*ProfessionalInput)
		wantField string
	}{
		{"nombre vacio", func(in *ProfessionalInput) { in.FirstName = "   " }, "firstName"},
		{"nombre muy largo", func(in *ProfessionalInput) { in.FirstName = strings.Repeat("a", 101) }, "firstName"},
		{"apellido vacio", func(in *ProfessionalInput) { in.LastName = "" }, "lastName"},
		{"matricula invalida", func(in *ProfessionalInput) { in.Matricula = "XX 123" }, "matricula"},
		{"especialidad desconocida", func(in *ProfessionalInput) { in.Especialidad = "cardiologia" }, "especialidad"},
		{"bio muy larga", func(in *ProfessionalInput) { in.Bio = strings.Repeat("a", 2001) }, "bio"},
		{"precio negativo", func(in *ProfessionalInput) { in.ConsultaPrice = -1 }, "consultaPriceCents"},
		{"sin modalidades", func(in *ProfessionalInput) { in.Modalidades = nil }, "modalidades"},
		{"modalidad desconocida", func(in *ProfessionalInput) { in.Modalidades = []string{"online"} }, "modalidades"},
		{"modalidad repetida", func(in *ProfessionalInput) { in.Modalidades = []string{"presencial", "presencial"} }, "modalidades"},
		{"zona vacia", func(in *ProfessionalInput) { in.Zona = "" }, "zona"},
		{"obra social repetida", func(in *ProfessionalInput) { in.ObrasSociales = []string{"OSDE", "osde"} }, "obrasSociales"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.mutate(&in)

			_, err := NewProfessional(in, testNow)
			if err == nil {
				t.Fatal("se esperaba un error de validación")
			}

			var verr ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ValidationError, se obtuvo %T", err)
			}

			found := false
			for _, f := range verr.Fields {
				if f.Field == tt.wantField {
					found = true
				}
			}
			if !found {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", tt.wantField, verr.Fields)
			}
		})
	}
}

func TestNewProfessionalAcumulaErrores(t *testing.T) {
	in := validInput()
	in.FirstName = ""
	in.Matricula = "roto"
	in.Zona = ""

	_, err := NewProfessional(in, testNow)

	var verr ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ValidationError, se obtuvo %T", err)
	}
	// el punto de acumular: el cliente corrige los tres de una
	if len(verr.Fields) != 3 {
		t.Errorf("se esperaban 3 campos con error, se obtuvieron %d: %+v", len(verr.Fields), verr.Fields)
	}
}

func TestNewProfessionalNormalizaEntrada(t *testing.T) {
	in := validInput()
	in.FirstName = "  Martín  "
	in.Especialidad = "  PSICOLOGIA  "
	in.Modalidades = []string{" Telemedicina "}

	p, err := NewProfessional(in, testNow)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.FirstName != "Martín" {
		t.Errorf("FirstName = %q, se esperaba sin espacios", p.FirstName)
	}
	if p.Especialidad != EspecialidadPsicologia {
		t.Errorf("Especialidad = %q, se esperaba psicologia", p.Especialidad)
	}
	if len(p.Modalidades) != 1 || p.Modalidades[0] != ModalidadTelemedicina {
		t.Errorf("Modalidades = %v, se esperaba [telemedicina]", p.Modalidades)
	}
}

func TestApplyUpdateResetaVerificacion(t *testing.T) {
	base, err := NewProfessional(validInput(), testNow)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	base.Verification = VerificationVerified

	later := testNow.Add(time.Hour)

	t.Run("cambiar la matricula resetea", func(t *testing.T) {
		in := validInput()
		in.Matricula = "MN 11111"

		updated, err := base.ApplyUpdate(in, later)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if updated.Verification != VerificationPending {
			t.Error("cambiar la matrícula tenía que volver la verificación a pending")
		}
	})

	t.Run("cambiar la especialidad resetea", func(t *testing.T) {
		in := validInput()
		in.Especialidad = "odontologia"

		updated, err := base.ApplyUpdate(in, later)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if updated.Verification != VerificationPending {
			t.Error("cambiar la especialidad tenía que volver la verificación a pending")
		}
	})

	t.Run("cambiar la bio no resetea", func(t *testing.T) {
		in := validInput()
		in.Bio = "Otra bio."

		updated, err := base.ApplyUpdate(in, later)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if updated.Verification != VerificationVerified {
			t.Error("editar la bio no tenía por qué tocar la verificación")
		}
	})
}

func TestApplyUpdatePreservaCamposNoEditables(t *testing.T) {
	base, err := NewProfessional(validInput(), testNow)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	in := validInput()
	in.FirstName = "Otro"
	in.LastName = "Nombre"

	later := testNow.Add(time.Hour)
	updated, err := base.ApplyUpdate(in, later)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if updated.ID != base.ID {
		t.Error("el ID no es editable")
	}
	// el slug es una URL pública: regenerarlo al cambiar el nombre rompe
	// enlaces y posicionamiento
	if updated.Slug != base.Slug {
		t.Errorf("el slug no debía cambiar: %q → %q", base.Slug, updated.Slug)
	}
	if !updated.CreatedAt.Equal(base.CreatedAt) {
		t.Error("CreatedAt no es editable")
	}
	if !updated.UpdatedAt.Equal(later) {
		t.Error("UpdatedAt debía avanzar")
	}
}

func TestDeactivateReactivate(t *testing.T) {
	p, err := NewProfessional(validInput(), testNow)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	later := testNow.Add(time.Hour)
	off := p.Deactivate(later)

	if off.Status != StatusInactive {
		t.Errorf("Status = %q, se esperaba inactive", off.Status)
	}
	if off.DeactivatedAt == nil || !off.DeactivatedAt.Equal(later) {
		t.Error("DeactivatedAt debía sellarse con el momento de la baja")
	}
	// value receiver: el original no se toca
	if p.Status != StatusActive {
		t.Error("Deactivate no debía mutar el receptor")
	}

	// idempotente: dar de baja algo ya dado de baja no es un error ni
	// corre la fecha original
	evenLater := later.Add(time.Hour)
	again := off.Deactivate(evenLater)
	if !again.DeactivatedAt.Equal(later) {
		t.Error("una segunda baja no debía correr la fecha de la primera")
	}

	back := off.Reactivate(evenLater)
	if back.Status != StatusActive {
		t.Errorf("Status = %q, se esperaba active", back.Status)
	}
	if back.DeactivatedAt != nil {
		t.Error("reactivar debía limpiar DeactivatedAt")
	}
}

func TestCloneEsCopiaProfunda(t *testing.T) {
	p, err := NewProfessional(validInput(), testNow)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	c := p.Clone()
	c.Modalidades[0] = ModalidadDomicilio
	c.ObrasSociales[0] = "MUTADA"

	if p.Modalidades[0] == ModalidadDomicilio {
		t.Error("mutar el clon alteró las modalidades del original")
	}
	if p.ObrasSociales[0] == "MUTADA" {
		t.Error("mutar el clon alteró las obras sociales del original")
	}

	off := p.Deactivate(testNow)
	offClone := off.Clone()
	*offClone.DeactivatedAt = testNow.Add(time.Hour)
	if off.DeactivatedAt.Equal(testNow.Add(time.Hour)) {
		t.Error("mutar el clon alteró el DeactivatedAt del original")
	}
}

func TestFullName(t *testing.T) {
	p, err := NewProfessional(validInput(), testNow)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got := p.FullName(); got != "Martín González" {
		t.Errorf("FullName() = %q, se esperaba %q", got, "Martín González")
	}
}
```

- [ ] **Step 3: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/domain/ -run TestNewProfessional -v`
Expected: FAIL con `undefined: ProfessionalInput` y `undefined: NewProfessional`

- [ ] **Step 4: Implementar la entidad**

Archivo `apps/api/internal/domain/professional.go`:

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
	maxNameLen = 100
	maxBioLen  = 2000
	maxZonaLen = 100
)

// Professional es un profesional de la salud dado de alta en la plataforma.
//
// Invariante del paquete: no se puede construir uno inválido desde afuera.
// No hay setters públicos; la única puerta de entrada es NewProfessional, y la
// única forma de modificarlo es ApplyUpdate, que revalida todo.
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

// ProfessionalInput es la entrada cruda, en tipos primitivos. Que sea primitiva
// no es descuido: obliga a que todo el parseo y toda la validación ocurran acá
// adentro, y no repartidos por los handlers.
type ProfessionalInput struct {
	FirstName     string
	LastName      string
	Matricula     string
	Especialidad  string
	Bio           string
	ConsultaPrice int64
	Modalidades   []string
	Zona          string
	ObrasSociales []string
}

// NewProfessional valida la entrada y devuelve un profesional consistente o un
// ValidationError con todos los campos que fallaron.
func NewProfessional(in ProfessionalInput, now time.Time) (Professional, error) {
	p, verr := build(in)
	if verr.hasErrors() {
		return Professional{}, verr
	}

	p.ID = uuid.New()
	p.Slug = Slugify(p.FullName())
	if p.Slug == "" {
		// el nombre pasó la validación pero no dejó ningún carácter usable
		// (por ejemplo "..."). Sin esto quedaría un slug vacío y la URL
		// pública del profesional colisionaría con la de cualquier otro.
		p.Slug = p.ID.String()
	}
	p.Status = StatusActive
	p.Verification = VerificationPending
	p.CreatedAt = now
	p.UpdatedAt = now

	return p, nil
}

// ApplyUpdate reemplaza los campos editables y devuelve el resultado sin tocar
// el receptor. ID, Slug, Status, CreatedAt y DeactivatedAt no son editables.
func (p Professional) ApplyUpdate(in ProfessionalInput, now time.Time) (Professional, error) {
	updated, verr := build(in)
	if verr.hasErrors() {
		return Professional{}, verr
	}

	updated.ID = p.ID
	updated.Slug = p.Slug
	updated.Status = p.Status
	updated.CreatedAt = p.CreatedAt
	updated.DeactivatedAt = p.DeactivatedAt
	updated.UpdatedAt = now

	// La verificación se hizo sobre una matrícula y una especialidad
	// concretas. Si cambian, deja de valer: toda orientación, agenda o cobro
	// depende de que el profesional esté verificado.
	if updated.Matricula != p.Matricula || updated.Especialidad != p.Especialidad {
		updated.Verification = VerificationPending
	} else {
		updated.Verification = p.Verification
	}

	return updated, nil
}

// Deactivate da de baja al profesional. No es un borrado: los turnos y
// comprobantes históricos siguen apuntando a este registro. Es idempotente y
// no corre la fecha de la primera baja.
func (p Professional) Deactivate(now time.Time) Professional {
	if p.Status == StatusInactive {
		return p
	}
	p.Status = StatusInactive
	p.DeactivatedAt = &now
	p.UpdatedAt = now
	return p
}

// Reactivate revierte la baja. Idempotente.
func (p Professional) Reactivate(now time.Time) Professional {
	if p.Status == StatusActive {
		return p
	}
	p.Status = StatusActive
	p.DeactivatedAt = nil
	p.UpdatedAt = now
	return p
}

// Clone devuelve una copia profunda.
//
// Una copia superficial comparte el array que hay debajo de los slices, y deja
// que quien la reciba mute el original desde afuera sin enterarse. Es el bug
// número uno de un repositorio en memoria.
func (p Professional) Clone() Professional {
	c := p
	c.Modalidades = slices.Clone(p.Modalidades)
	c.ObrasSociales = slices.Clone(p.ObrasSociales)
	if p.DeactivatedAt != nil {
		t := *p.DeactivatedAt
		c.DeactivatedAt = &t
	}
	return c
}

func (p Professional) FullName() string {
	return p.FirstName + " " + p.LastName
}

// build parsea y valida la entrada, acumulando todos los errores. Es la única
// implementación de las reglas: la comparten el alta y la edición.
func build(in ProfessionalInput) (Professional, ValidationError) {
	var p Professional
	var verr ValidationError

	p.FirstName = validateName(in.FirstName, "firstName", &verr)
	p.LastName = validateName(in.LastName, "lastName", &verr)

	if m, err := ParseMatricula(in.Matricula); err != nil {
		verr.add("matricula", err.Error())
	} else {
		p.Matricula = m
	}

	esp := Especialidad(strings.ToLower(strings.TrimSpace(in.Especialidad)))
	if !esp.Valid() {
		verr.add("especialidad", "debe ser psicologia, kinesiologia u odontologia")
	} else {
		p.Especialidad = esp
	}

	p.Bio = strings.TrimSpace(in.Bio)
	if utf8.RuneCountInString(p.Bio) > maxBioLen {
		verr.add("bio", fmt.Sprintf("no puede superar los %d caracteres", maxBioLen))
	}

	if in.ConsultaPrice < 0 {
		verr.add("consultaPriceCents", "no puede ser negativo")
	} else {
		p.ConsultaPrice = Money(in.ConsultaPrice)
	}

	p.Modalidades = buildModalidades(in.Modalidades, &verr)

	p.Zona = strings.TrimSpace(in.Zona)
	switch {
	case p.Zona == "":
		verr.add("zona", "es obligatoria")
	case utf8.RuneCountInString(p.Zona) > maxZonaLen:
		verr.add("zona", fmt.Sprintf("no puede superar los %d caracteres", maxZonaLen))
	}

	p.ObrasSociales = buildObrasSociales(in.ObrasSociales, &verr)

	return p, verr
}

func validateName(raw, field string, verr *ValidationError) string {
	name := strings.TrimSpace(raw)
	switch {
	case name == "":
		verr.add(field, "es obligatorio")
	case utf8.RuneCountInString(name) > maxNameLen:
		verr.add(field, fmt.Sprintf("no puede superar los %d caracteres", maxNameLen))
	}
	return name
}

func buildModalidades(raw []string, verr *ValidationError) []Modalidad {
	if len(raw) == 0 {
		verr.add("modalidades", "se requiere al menos una")
		return nil
	}

	seen := make(map[Modalidad]bool, len(raw))
	out := make([]Modalidad, 0, len(raw))

	for _, r := range raw {
		m := Modalidad(strings.ToLower(strings.TrimSpace(r)))
		switch {
		case !m.Valid():
			verr.add("modalidades", fmt.Sprintf("%q no es una modalidad válida", r))
		case seen[m]:
			verr.add("modalidades", fmt.Sprintf("%q está repetida", r))
		default:
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

func buildObrasSociales(raw []string, verr *ValidationError) []string {
	// puede estar vacía: un profesional que solo atiende privado es válido
	seen := make(map[string]bool, len(raw))
	out := make([]string, 0, len(raw))

	for _, r := range raw {
		v := strings.TrimSpace(r)
		if v == "" {
			continue
		}
		// "OSDE" y "osde" son la misma obra social
		key := Normalize(v)
		if seen[key] {
			verr.add("obrasSociales", fmt.Sprintf("%q está repetida", v))
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	return out
}
```

- [ ] **Step 5: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -v`
Expected: PASS en todos. Prestar atención a `TestCloneEsCopiaProfunda`: si falla, `Clone` está haciendo copia superficial y todo el repositorio en memoria de la Task 5 va a estar roto.

- [ ] **Step 6: Verificar que el dominio no importa nada del proyecto**

Run: `cd apps/api && go list -deps ./internal/domain | grep "joaquinfochoa"`
Expected: una sola línea, `github.com/joaquinfochoa/Salud/apps/api/internal/domain`. Cualquier otra línea significa que el dominio depende de otra capa y la arquitectura se rompió.

- [ ] **Step 7: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): entidad Professional con validación acumulativa

No se puede construir un Professional inválido desde fuera del paquete.
ValidationError junta todos los campos que fallan de una pasada.

Cambiar la matrícula o la especialidad devuelve la verificación a
pending: la verificación se hizo sobre esos datos y deja de valer si
cambian.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 5: Interfaz del repositorio e implementación en memoria

**Files:**
- Create: `apps/api/internal/repository/professional.go`
- Create: `apps/api/internal/repository/memory/professional.go`
- Test: `apps/api/internal/repository/memory/professional_test.go`

**Interfaces:**
- Consumes: todo el paquete `domain` (Tasks 2-4)
- Produces:
  - `repository.Filter{Especialidad *domain.Especialidad; Zona *string; Status *domain.Status; Query *string; Limit int; Offset int}`
  - `repository.Professional` — la interfaz con `Create`, `GetByID`, `GetBySlug`, `GetByMatricula`, `List`, `Update`
  - `func memory.NewProfessional() *memory.Professional` — implementa `repository.Professional`

- [ ] **Step 1: Escribir la interfaz**

No lleva test: es una declaración de tipos. El test de que la implementación la cumple es la aserción de compilación del paso 4.

Archivo `apps/api/internal/repository/professional.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Filter son los criterios del listado. Los punteros distinguen "no filtrar
// por este campo" de "filtrar por el valor cero", que con valores planos no se
// puede: un Status("") sería indistinguible de "sin filtro".
type Filter struct {
	Especialidad *domain.Especialidad
	Zona         *string
	Status       *domain.Status
	Query        *string // busca en nombre y apellido, sin distinguir acentos
	Limit        int
	Offset       int
}

// Professional es el punto de cambio a PostgreSQL. Cuando exista la
// implementación con base de datos, migrar es cambiar una línea de main.go y
// nada más.
//
// Todos los métodos reciben context.Context aunque la implementación en
// memoria lo ignore: agregarlo después obligaría a tocar todas las firmas.
//
// No hay Delete: la baja es lógica y se hace con Update.
type Professional interface {
	Create(ctx context.Context, p domain.Professional) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Professional, error)
	GetBySlug(ctx context.Context, slug string) (domain.Professional, error)
	GetByMatricula(ctx context.Context, m domain.Matricula) (domain.Professional, error)
	List(ctx context.Context, f Filter) ([]domain.Professional, int, error)
	Update(ctx context.Context, p domain.Professional) error
}
```

- [ ] **Step 2: Escribir los tests del repositorio en memoria**

Archivo `apps/api/internal/repository/memory/professional_test.go`:

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

var testNow = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

func makeProfessional(t *testing.T, first, last, matricula string, esp domain.Especialidad, zona string) domain.Professional {
	t.Helper()
	p, err := domain.NewProfessional(domain.ProfessionalInput{
		FirstName:     first,
		LastName:      last,
		Matricula:     matricula,
		Especialidad:  string(esp),
		Bio:           "bio",
		ConsultaPrice: 1000000,
		Modalidades:   []string{"telemedicina"},
		Zona:          zona,
		ObrasSociales: []string{"OSDE"},
	}, testNow)
	if err != nil {
		t.Fatalf("no se pudo construir el profesional de prueba: %v", err)
	}
	return p
}

func TestCreateYGetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()
	p := makeProfessional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID devolvió error: %v", err)
	}
	if got.ID != p.ID || got.Slug != p.Slug {
		t.Errorf("el profesional recuperado no coincide: %+v", got)
	}
}

func TestGetByIDNoExiste(t *testing.T) {
	_, err := NewProfessional().GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound, se obtuvo %v", err)
	}
}

func TestGetBySlugYPorMatricula(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()
	p := makeProfessional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	bySlug, err := repo.GetBySlug(ctx, "martin-gonzalez")
	if err != nil {
		t.Fatalf("GetBySlug devolvió error: %v", err)
	}
	if bySlug.ID != p.ID {
		t.Error("GetBySlug devolvió otro profesional")
	}

	byMat, err := repo.GetByMatricula(ctx, p.Matricula)
	if err != nil {
		t.Fatalf("GetByMatricula devolvió error: %v", err)
	}
	if byMat.ID != p.ID {
		t.Error("GetByMatricula devolvió otro profesional")
	}

	if _, err := repo.GetBySlug(ctx, "no-existe"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound, se obtuvo %v", err)
	}
}

// El test más importante de este paquete. Si el repositorio devuelve algo que
// comparte memoria con lo que guardó, quien lo reciba puede mutar el store
// desde afuera sin enterarse.
func TestElStoreDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()
	p := makeProfessional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	// mutar lo que devolvió el repositorio
	got, _ := repo.GetByID(ctx, p.ID)
	got.FirstName = "MUTADO"
	got.Modalidades[0] = domain.ModalidadDomicilio
	got.ObrasSociales[0] = "MUTADA"

	fresh, _ := repo.GetByID(ctx, p.ID)
	if fresh.FirstName == "MUTADO" {
		t.Error("mutar el resultado alteró el store")
	}
	if fresh.Modalidades[0] == domain.ModalidadDomicilio {
		t.Error("las modalidades comparten memoria con el store")
	}
	if fresh.ObrasSociales[0] == "MUTADA" {
		t.Error("las obras sociales comparten memoria con el store")
	}

	// y al revés: mutar lo que se pasó a Create tampoco debe afectar
	p.Modalidades[0] = domain.ModalidadPresencial
	fresh2, _ := repo.GetByID(ctx, p.ID)
	if fresh2.Modalidades[0] == domain.ModalidadPresencial {
		t.Error("Create guardó una referencia en vez de una copia")
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()
	p := makeProfessional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	p.Zona = "GBA Norte"
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("Update devolvió error: %v", err)
	}

	got, _ := repo.GetByID(ctx, p.ID)
	if got.Zona != "GBA Norte" {
		t.Errorf("Zona = %q, se esperaba %q", got.Zona, "GBA Norte")
	}

	desconocido := makeProfessional(t, "Otro", "Nombre", "MN 11111", domain.EspecialidadOdontologia, "CABA")
	if err := repo.Update(ctx, desconocido); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound al actualizar algo inexistente, se obtuvo %v", err)
	}
}

func TestListFiltros(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()

	psico := makeProfessional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	kine := makeProfessional(t, "Pablo", "Moreno", "MN 45321", domain.EspecialidadKinesiologia, "CABA")
	odonto := makeProfessional(t, "Gabriela", "Ríos", "MN 67890", domain.EspecialidadOdontologia, "GBA Norte")
	// las fechas distintas fuerzan un orden determinista
	kine.CreatedAt = testNow.Add(time.Minute)
	odonto.CreatedAt = testNow.Add(2 * time.Minute)

	for _, p := range []domain.Professional{psico, kine, odonto} {
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create devolvió error: %v", err)
		}
	}

	t.Run("sin filtros devuelve todo", func(t *testing.T) {
		got, total, err := repo.List(ctx, repository.Filter{Limit: 10})
		if err != nil {
			t.Fatalf("List devolvió error: %v", err)
		}
		if total != 3 || len(got) != 3 {
			t.Errorf("total=%d len=%d, se esperaba 3 y 3", total, len(got))
		}
	})

	t.Run("por especialidad", func(t *testing.T) {
		esp := domain.EspecialidadPsicologia
		got, total, _ := repo.List(ctx, repository.Filter{Especialidad: &esp, Limit: 10})
		if total != 1 || got[0].ID != psico.ID {
			t.Errorf("se esperaba solo el psicólogo, se obtuvo total=%d", total)
		}
	})

	t.Run("por zona sin distinguir acentos", func(t *testing.T) {
		zona := "caba"
		_, total, _ := repo.List(ctx, repository.Filter{Zona: &zona, Limit: 10})
		if total != 2 {
			t.Errorf("total = %d, se esperaban 2 en CABA", total)
		}
	})

	t.Run("busqueda sin acentos", func(t *testing.T) {
		// buscar "gonzalez" tiene que encontrar a "González"
		q := "gonzalez"
		got, total, _ := repo.List(ctx, repository.Filter{Query: &q, Limit: 10})
		if total != 1 || got[0].ID != psico.ID {
			t.Errorf("la búsqueda sin acentos no encontró a González: total=%d", total)
		}
	})

	t.Run("busqueda por nombre parcial", func(t *testing.T) {
		q := "pab"
		_, total, _ := repo.List(ctx, repository.Filter{Query: &q, Limit: 10})
		if total != 1 {
			t.Errorf("total = %d, se esperaba 1", total)
		}
	})

	t.Run("por status", func(t *testing.T) {
		off := psico.Deactivate(testNow.Add(time.Hour))
		if err := repo.Update(ctx, off); err != nil {
			t.Fatalf("Update devolvió error: %v", err)
		}

		inactive := domain.StatusInactive
		_, total, _ := repo.List(ctx, repository.Filter{Status: &inactive, Limit: 10})
		if total != 1 {
			t.Errorf("total = %d, se esperaba 1 inactivo", total)
		}

		active := domain.StatusActive
		_, total, _ = repo.List(ctx, repository.Filter{Status: &active, Limit: 10})
		if total != 2 {
			t.Errorf("total = %d, se esperaban 2 activos", total)
		}
	})
}

func TestListPaginacionEsEstable(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()

	for i := range 10 {
		p := makeProfessional(t, fmt.Sprintf("Nombre%d", i), "Apellido",
			fmt.Sprintf("MN %d", 10000+i), domain.EspecialidadPsicologia, "CABA")
		p.CreatedAt = testNow.Add(time.Duration(i) * time.Minute)
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create devolvió error: %v", err)
		}
	}

	page1, total, _ := repo.List(ctx, repository.Filter{Limit: 3, Offset: 0})
	if total != 10 || len(page1) != 3 {
		t.Fatalf("total=%d len=%d, se esperaba 10 y 3", total, len(page1))
	}

	page2, _, _ := repo.List(ctx, repository.Filter{Limit: 3, Offset: 3})
	if len(page2) != 3 {
		t.Fatalf("len(page2) = %d, se esperaba 3", len(page2))
	}

	// el mapa de Go itera en orden aleatorio: sin ordenar, dos llamadas
	// idénticas devolverían páginas distintas y la paginación no serviría
	for range 5 {
		again, _, _ := repo.List(ctx, repository.Filter{Limit: 3, Offset: 0})
		for i := range page1 {
			if again[i].ID != page1[i].ID {
				t.Fatal("dos llamadas idénticas devolvieron órdenes distintos")
			}
		}
	}

	// las páginas no se solapan
	for _, a := range page1 {
		for _, b := range page2 {
			if a.ID == b.ID {
				t.Error("la página 1 y la 2 comparten un elemento")
			}
		}
	}

	last, _, _ := repo.List(ctx, repository.Filter{Limit: 3, Offset: 9})
	if len(last) != 1 {
		t.Errorf("la última página tenía %d elementos, se esperaba 1", len(last))
	}

	empty, _, _ := repo.List(ctx, repository.Filter{Limit: 3, Offset: 50})
	if len(empty) != 0 {
		t.Errorf("un offset más allá del total debía devolver vacío, devolvió %d", len(empty))
	}
}

func TestAccesoConcurrente(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := makeProfessional(t, fmt.Sprintf("N%d", i), "A",
				fmt.Sprintf("MN %d", 20000+i), domain.EspecialidadPsicologia, "CABA")
			_ = repo.Create(ctx, p)
			_, _, _ = repo.List(ctx, repository.Filter{Limit: 10})
			_, _ = repo.GetByID(ctx, p.ID)
		}(i)
	}
	wg.Wait()

	_, total, _ := repo.List(ctx, repository.Filter{Limit: 100})
	if total != 50 {
		t.Errorf("total = %d, se esperaban 50", total)
	}
}
```

- [ ] **Step 3: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/repository/... -v`
Expected: FAIL con `undefined: NewProfessional`

- [ ] **Step 4: Implementar el repositorio en memoria**

Archivo `apps/api/internal/repository/memory/professional.go`:

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
var _ repository.Professional = (*Professional)(nil)

// Professional guarda los profesionales en memoria. Se pierde todo al
// reiniciar, y está bien: sirve para definir el dominio antes de comprometerse
// con un esquema de base de datos, y para correr los tests sin infraestructura.
type Professional struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Professional
}

func NewProfessional() *Professional {
	return &Professional{data: make(map[uuid.UUID]domain.Professional)}
}

func (r *Professional) Create(_ context.Context, p domain.Professional) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[p.ID]; exists {
		return fmt.Errorf("ya existe un profesional con id %s", p.ID)
	}
	r.data[p.ID] = p.Clone()
	return nil
}

func (r *Professional) GetByID(_ context.Context, id uuid.UUID) (domain.Professional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.data[id]
	if !ok {
		return domain.Professional{}, domain.ErrNotFound
	}
	return p.Clone(), nil
}

func (r *Professional) GetBySlug(_ context.Context, slug string) (domain.Professional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.data {
		if p.Slug == slug {
			return p.Clone(), nil
		}
	}
	return domain.Professional{}, domain.ErrNotFound
}

func (r *Professional) GetByMatricula(_ context.Context, m domain.Matricula) (domain.Professional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.data {
		if p.Matricula == m {
			return p.Clone(), nil
		}
	}
	return domain.Professional{}, domain.ErrNotFound
}

func (r *Professional) List(_ context.Context, f repository.Filter) ([]domain.Professional, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// ponytail: scan O(n), correcto para un store en memoria. La
	// implementación Postgres resuelve esto con índices sobre especialidad,
	// zona y status.
	matched := make([]domain.Professional, 0, len(r.data))
	for _, p := range r.data {
		if matches(p, f) {
			matched = append(matched, p.Clone())
		}
	}

	// El mapa de Go itera en orden aleatorio. Sin este orden, dos llamadas
	// idénticas devolverían páginas distintas y la paginación sería inútil.
	slices.SortFunc(matched, func(a, b domain.Professional) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID.String(), b.ID.String())
	})

	total := len(matched)
	if f.Offset >= total {
		return []domain.Professional{}, total, nil
	}

	end := total
	if f.Limit > 0 && f.Offset+f.Limit < total {
		end = f.Offset + f.Limit
	}
	return matched[f.Offset:end], total, nil
}

func (r *Professional) Update(_ context.Context, p domain.Professional) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[p.ID]; !exists {
		return domain.ErrNotFound
	}
	r.data[p.ID] = p.Clone()
	return nil
}

func matches(p domain.Professional, f repository.Filter) bool {
	if f.Especialidad != nil && p.Especialidad != *f.Especialidad {
		return false
	}
	if f.Status != nil && p.Status != *f.Status {
		return false
	}
	if f.Zona != nil && domain.Normalize(p.Zona) != domain.Normalize(*f.Zona) {
		return false
	}
	if f.Query != nil {
		q := domain.Normalize(*f.Query)
		if q != "" && !strings.Contains(domain.Normalize(p.FullName()), q) {
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
- Create: `apps/api/internal/service/professional.go`
- Test: `apps/api/internal/service/professional_test.go`

**Interfaces:**
- Consumes: `domain` (Tasks 2-4), `repository.Professional` y `repository.Filter` (Task 5), `memory.NewProfessional` (Task 5, solo en tests)
- Produces:
  - `func service.NewProfessional(repository.Professional) *service.Professional`
  - `func (*service.Professional) Create(context.Context, domain.ProfessionalInput) (domain.Professional, error)`
  - `func (*service.Professional) GetByID(context.Context, uuid.UUID) (domain.Professional, error)`
  - `func (*service.Professional) GetBySlug(context.Context, string) (domain.Professional, error)`
  - `func (*service.Professional) List(context.Context, repository.Filter) ([]domain.Professional, int, error)`
  - Constantes `service.DefaultLimit = 20`, `service.MaxLimit = 100`

- [ ] **Step 1: Escribir los tests del alta y las lecturas**

Archivo `apps/api/internal/service/professional_test.go`:

```go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

// No hay mocks en este proyecto. El repositorio en memoria es rápido y
// determinista, así que es el doble de test: se prueba contra la
// implementación de verdad. Si un test pareciera necesitar un mock, la
// frontera está mal dibujada.
func newTestService() *Professional {
	return NewProfessional(memory.NewProfessional())
}

func validInput() domain.ProfessionalInput {
	return domain.ProfessionalInput{
		FirstName:     "Martín",
		LastName:      "González",
		Matricula:     "MN 98.234",
		Especialidad:  "psicologia",
		Bio:           "Psicólogo clínico.",
		ConsultaPrice: 1200000,
		Modalidades:   []string{"telemedicina"},
		Zona:          "CABA",
		ObrasSociales: []string{"OSDE"},
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	p, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}
	if p.Slug != "martin-gonzalez" {
		t.Errorf("Slug = %q, se esperaba %q", p.Slug, "martin-gonzalez")
	}
	if p.Status != domain.StatusActive || p.Verification != domain.VerificationPending {
		t.Error("un profesional nuevo nace activo y sin verificar")
	}

	// tiene que quedar realmente guardado
	got, err := svc.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID devolvió error: %v", err)
	}
	if got.ID != p.ID {
		t.Error("el profesional no quedó persistido")
	}
}

func TestCreateValidacion(t *testing.T) {
	in := validInput()
	in.FirstName = ""

	_, err := newTestService().Create(context.Background(), in)

	var verr domain.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ValidationError, se obtuvo %T: %v", err, err)
	}
}

func TestCreateMatriculaDuplicada(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.Create(ctx, validInput()); err != nil {
		t.Fatalf("el primer alta falló: %v", err)
	}

	// la misma matrícula escrita distinto sigue siendo la misma matrícula
	otro := validInput()
	otro.FirstName = "Otro"
	otro.LastName = "Profesional"
	otro.Matricula = "m.n. 98234"

	_, err := svc.Create(ctx, otro)
	if !errors.Is(err, domain.ErrMatriculaTaken) {
		t.Errorf("se esperaba ErrMatriculaTaken, se obtuvo %v", err)
	}
}

func TestCreateSlugUnico(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	// tres homónimos: dos "Martín González" son perfectamente posibles y no
	// pueden ser un error para el cliente
	first, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	in2 := validInput()
	in2.Matricula = "MN 11111"
	second, err := svc.Create(ctx, in2)
	if err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	in3 := validInput()
	in3.Matricula = "MN 22222"
	third, err := svc.Create(ctx, in3)
	if err != nil {
		t.Fatalf("alta 3 falló: %v", err)
	}

	if first.Slug != "martin-gonzalez" {
		t.Errorf("slug 1 = %q", first.Slug)
	}
	if second.Slug != "martin-gonzalez-2" {
		t.Errorf("slug 2 = %q, se esperaba martin-gonzalez-2", second.Slug)
	}
	if third.Slug != "martin-gonzalez-3" {
		t.Errorf("slug 3 = %q, se esperaba martin-gonzalez-3", third.Slug)
	}
}

func TestGetBySlug(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	p, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	got, err := svc.GetBySlug(ctx, p.Slug)
	if err != nil {
		t.Fatalf("GetBySlug devolvió error: %v", err)
	}
	if got.ID != p.ID {
		t.Error("GetBySlug devolvió otro profesional")
	}

	if _, err := svc.GetBySlug(ctx, "no-existe"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound, se obtuvo %v", err)
	}
}

func TestGetByIDNoExiste(t *testing.T) {
	_, err := newTestService().GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound, se obtuvo %v", err)
	}
}

func TestListDefaults(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	for i := range 3 {
		in := validInput()
		in.Matricula = "MN 3000" + string(rune('0'+i))
		if _, err := svc.Create(ctx, in); err != nil {
			t.Fatalf("Create devolvió error: %v", err)
		}
	}

	t.Run("limite por defecto", func(t *testing.T) {
		f := repository.Filter{}
		_, total, err := svc.List(ctx, f)
		if err != nil {
			t.Fatalf("List devolvió error: %v", err)
		}
		if total != 3 {
			t.Errorf("total = %d, se esperaba 3", total)
		}
	})

	t.Run("limite recortado al maximo", func(t *testing.T) {
		got, _, err := svc.List(ctx, repository.Filter{Limit: 5000})
		if err != nil {
			t.Fatalf("List devolvió error: %v", err)
		}
		if len(got) > MaxLimit {
			t.Errorf("devolvió %d elementos, el máximo es %d", len(got), MaxLimit)
		}
	})

	t.Run("offset negativo se normaliza", func(t *testing.T) {
		got, _, err := svc.List(ctx, repository.Filter{Offset: -10})
		if err != nil {
			t.Fatalf("List devolvió error: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("len = %d, se esperaba 3", len(got))
		}
	})
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/service/ -v`
Expected: FAIL con `undefined: NewProfessional`

- [ ] **Step 3: Implementar el servicio con alta y lecturas**

Archivo `apps/api/internal/service/professional.go`:

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
	// DefaultLimit es cuántos profesionales devuelve el listado si el cliente
	// no pide un tamaño.
	DefaultLimit = 20

	// MaxLimit es el techo. Sin techo, un cliente puede pedir el padrón
	// entero en una llamada.
	MaxLimit = 100
)

// Professional resuelve los casos de uso que necesitan mirar más de un
// profesional a la vez. Las reglas que se deciden con una sola entidad viven
// en el dominio, no acá.
type Professional struct {
	repo repository.Professional

	// now es inyectable para que los tests no dependan del reloj.
	now func() time.Time
}

func NewProfessional(repo repository.Professional) *Professional {
	return &Professional{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Professional) Create(ctx context.Context, in domain.ProfessionalInput) (domain.Professional, error) {
	p, err := domain.NewProfessional(in, s.now())
	if err != nil {
		return domain.Professional{}, err
	}

	// La matrícula es la única identidad real de una persona en este sistema.
	// El parser ya normalizó "M.N. 98.234" y "MN 98234" a lo mismo, así que
	// esta comparación atrapa los duplicados escritos distinto.
	if err := s.assertMatriculaLibre(ctx, p.Matricula, uuid.Nil); err != nil {
		return domain.Professional{}, err
	}

	slug, err := s.uniqueSlug(ctx, p.Slug)
	if err != nil {
		return domain.Professional{}, err
	}
	p.Slug = slug

	if err := s.repo.Create(ctx, p); err != nil {
		return domain.Professional{}, err
	}
	return p, nil
}

func (s *Professional) GetByID(ctx context.Context, id uuid.UUID) (domain.Professional, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Professional) GetBySlug(ctx context.Context, slug string) (domain.Professional, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *Professional) List(ctx context.Context, f repository.Filter) ([]domain.Professional, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.List(ctx, f)
}

// assertMatriculaLibre falla si otro profesional ya tiene esa matrícula.
// exclude permite ignorar al propio profesional durante una edición.
func (s *Professional) assertMatriculaLibre(ctx context.Context, m domain.Matricula, exclude uuid.UUID) error {
	existing, err := s.repo.GetByMatricula(ctx, m)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil
	case err != nil:
		return err
	case existing.ID == exclude:
		return nil
	default:
		return domain.ErrMatriculaTaken
	}
}

// uniqueSlug resuelve las colisiones agregando un sufijo numérico.
//
// Nunca es un error para el cliente: dos "Martín González" son perfectamente
// posibles y no hay razón para rechazar al segundo.
func (s *Professional) uniqueSlug(ctx context.Context, base string) (string, error) {
	candidate := base
	for i := 2; ; i++ {
		_, err := s.repo.GetBySlug(ctx, candidate)
		if errors.Is(err, domain.ErrNotFound) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/service/ -v`
Expected: PASS. `TestCreateSlugUnico` es el que confirma la cadena `martin-gonzalez`, `-2`, `-3`.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/service/
git commit -m "feat(service): alta y lecturas de Professional

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
- Modify: `apps/api/internal/service/professional.go` (agregar tres métodos)
- Modify: `apps/api/internal/service/professional_test.go` (agregar los tests)

**Interfaces:**
- Consumes: todo lo de la Task 6
- Produces:
  - `func (*service.Professional) Update(context.Context, uuid.UUID, domain.ProfessionalInput) (domain.Professional, error)`
  - `func (*service.Professional) Deactivate(context.Context, uuid.UUID) error`
  - `func (*service.Professional) Reactivate(context.Context, uuid.UUID) (domain.Professional, error)`

- [ ] **Step 1: Escribir los tests**

Agregar al final de `apps/api/internal/service/professional_test.go`:

```go
func TestUpdate(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	p, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	in := validInput()
	in.Bio = "Bio actualizada."
	in.Zona = "GBA Norte"

	updated, err := svc.Update(ctx, p.ID, in)
	if err != nil {
		t.Fatalf("Update devolvió error: %v", err)
	}
	if updated.Bio != "Bio actualizada." || updated.Zona != "GBA Norte" {
		t.Error("los campos editables no se aplicaron")
	}
	if updated.Slug != p.Slug {
		t.Error("el slug es una URL pública y no debía cambiar")
	}

	// tiene que haber quedado persistido
	got, _ := svc.GetByID(ctx, p.ID)
	if got.Zona != "GBA Norte" {
		t.Error("el cambio no quedó guardado")
	}
}

func TestUpdateNoExiste(t *testing.T) {
	_, err := newTestService().Update(context.Background(), uuid.New(), validInput())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound, se obtuvo %v", err)
	}
}

func TestUpdateMatriculaDeOtro(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	primero, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	in2 := validInput()
	in2.FirstName = "Carolina"
	in2.LastName = "Vega"
	in2.Matricula = "MN 11111"
	if _, err := svc.Create(ctx, in2); err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	// el primero intenta quedarse con la matrícula del segundo
	robo := validInput()
	robo.Matricula = "MN 11111"
	if _, err := svc.Update(ctx, primero.ID, robo); !errors.Is(err, domain.ErrMatriculaTaken) {
		t.Errorf("se esperaba ErrMatriculaTaken, se obtuvo %v", err)
	}
}

func TestUpdateConservaLaPropiaMatricula(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	p, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	// editar sin cambiar la matrícula no puede chocar consigo mismo
	in := validInput()
	in.Bio = "Otra bio."
	if _, err := svc.Update(ctx, p.ID, in); err != nil {
		t.Errorf("editar conservando la matrícula propia falló: %v", err)
	}
}

func TestUpdateResetaVerificacion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewProfessional()
	svc := NewProfessional(repo)

	p, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	// simular que ya fue verificado
	p.Verification = domain.VerificationVerified
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("no se pudo preparar el estado: %v", err)
	}

	in := validInput()
	in.Matricula = "MN 55555"
	updated, err := svc.Update(ctx, p.ID, in)
	if err != nil {
		t.Fatalf("Update devolvió error: %v", err)
	}
	if updated.Verification != domain.VerificationPending {
		t.Error("cambiar la matrícula tenía que invalidar la verificación")
	}
}

func TestDeactivateYReactivate(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	p, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create devolvió error: %v", err)
	}

	if err := svc.Deactivate(ctx, p.ID); err != nil {
		t.Fatalf("Deactivate devolvió error: %v", err)
	}

	// el registro sigue existiendo: un turno pasado apunta acá
	got, err := svc.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("el profesional dado de baja tenía que seguir existiendo: %v", err)
	}
	if got.Status != domain.StatusInactive {
		t.Errorf("Status = %q, se esperaba inactive", got.Status)
	}
	if got.DeactivatedAt == nil {
		t.Error("DeactivatedAt debía sellarse")
	}

	// pero no aparece en el listado por defecto
	_, total, _ := svc.List(ctx, repository.Filter{})
	if total != 0 {
		t.Errorf("el listado por defecto devolvió %d, se esperaba 0", total)
	}

	// filtrando explícitamente sí aparece
	inactive := domain.StatusInactive
	_, total, _ = svc.List(ctx, repository.Filter{Status: &inactive})
	if total != 1 {
		t.Errorf("el listado de inactivos devolvió %d, se esperaba 1", total)
	}

	// dar de baja dos veces es idempotente, no un error
	if err := svc.Deactivate(ctx, p.ID); err != nil {
		t.Errorf("la segunda baja debía ser idempotente, devolvió %v", err)
	}

	back, err := svc.Reactivate(ctx, p.ID)
	if err != nil {
		t.Fatalf("Reactivate devolvió error: %v", err)
	}
	if back.Status != domain.StatusActive || back.DeactivatedAt != nil {
		t.Error("reactivar debía dejarlo activo y limpiar DeactivatedAt")
	}

	_, total, _ = svc.List(ctx, repository.Filter{})
	if total != 1 {
		t.Errorf("después de reactivar el listado devolvió %d, se esperaba 1", total)
	}
}

func TestDeactivateNoExiste(t *testing.T) {
	if err := newTestService().Deactivate(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("se esperaba ErrNotFound, se obtuvo %v", err)
	}
}
```

Nota: el `List` por defecto todavía no filtra por activos. Ese comportamiento se agrega en el paso 3 y es lo que hace pasar `TestDeactivateYReactivate`.

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/service/ -run 'TestUpdate|TestDeactivate' -v`
Expected: FAIL con `svc.Update undefined` y `svc.Deactivate undefined`

- [ ] **Step 3: Implementar los tres métodos y el filtro por defecto**

En `apps/api/internal/service/professional.go`, reemplazar el método `List` por esta versión:

```go
func (s *Professional) List(ctx context.Context, f repository.Filter) ([]domain.Professional, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	// Por defecto solo los activos: un profesional dado de baja no tiene que
	// aparecer en una búsqueda de pacientes. Para verlos hay que pedirlos.
	if f.Status == nil {
		active := domain.StatusActive
		f.Status = &active
	}
	return s.repo.List(ctx, f)
}
```

Y agregar al final del archivo:

```go
// Update reemplaza los campos editables. Funciona también sobre profesionales
// dados de baja: editar los datos de alguien inactivo no tiene por qué
// bloquearse, y no cambia su estado.
func (s *Professional) Update(ctx context.Context, id uuid.UUID, in domain.ProfessionalInput) (domain.Professional, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Professional{}, err
	}

	updated, err := current.ApplyUpdate(in, s.now())
	if err != nil {
		return domain.Professional{}, err
	}

	if updated.Matricula != current.Matricula {
		if err := s.assertMatriculaLibre(ctx, updated.Matricula, id); err != nil {
			return domain.Professional{}, err
		}
	}

	if err := s.repo.Update(ctx, updated); err != nil {
		return domain.Professional{}, err
	}
	return updated, nil
}

// Deactivate da de baja. No borra: los turnos, comprobantes y pagos históricos
// siguen apuntando a este registro, y sin él el comprobante que un paciente
// presentó para un reintegro queda huérfano.
func (s *Professional) Deactivate(ctx context.Context, id uuid.UUID) error {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, current.Deactivate(s.now()))
}

func (s *Professional) Reactivate(ctx context.Context, id uuid.UUID) (domain.Professional, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Professional{}, err
	}

	reactivated := current.Reactivate(s.now())
	if err := s.repo.Update(ctx, reactivated); err != nil {
		return domain.Professional{}, err
	}
	return reactivated, nil
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

La baja no borra: pone Status en inactive y sella DeactivatedAt. Un
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
  - name: professionals
  - name: health

paths:
  /healthz:
    get:
      tags: [health]
      summary: Verifica que el servidor responde
      responses:
        '200':
          description: El servidor está vivo
          content:
            application/json:
              schema:
                type: object
                properties:
                  status: { type: string, example: ok }

  /api/v1/professionals:
    get:
      tags: [professionals]
      summary: Lista profesionales
      description: |
        Por defecto devuelve solo los activos. Para ver los dados de baja hay
        que pedir `status=inactive` explícitamente.
      parameters:
        - name: especialidad
          in: query
          schema: { $ref: '#/components/schemas/Especialidad' }
        - name: zona
          in: query
          description: Compara sin distinguir mayúsculas ni acentos
          schema: { type: string }
        - name: status
          in: query
          schema: { $ref: '#/components/schemas/Status' }
        - name: q
          in: query
          description: |
            Busca en nombre y apellido, sin distinguir mayúsculas ni acentos.
            Buscar `gonzalez` encuentra a `González`.
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, minimum: 1, maximum: 100, default: 20 }
        - name: offset
          in: query
          schema: { type: integer, minimum: 0, default: 0 }
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ProfessionalList' }

    post:
      tags: [professionals]
      summary: Da de alta un profesional
      description: |
        El profesional nace con `status: active` y `verification: pending`.
        Nadie nace verificado: la verificación es un acto contra REFEPS que
        todavía no está implementado.
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/ProfessionalRequest' }
      responses:
        '201':
          description: Creado
          headers:
            Location:
              description: URL del recurso creado
              schema: { type: string }
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Professional' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '409': { $ref: '#/components/responses/Conflict' }
        '422': { $ref: '#/components/responses/ValidationFailed' }

  /api/v1/professionals/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }

    get:
      tags: [professionals]
      summary: Trae un profesional por ID
      description: |
        Un profesional dado de baja devuelve 200 con `status: inactive`. El
        recurso existe, simplemente no opera.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Professional' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '404': { $ref: '#/components/responses/NotFound' }

    put:
      tags: [professionals]
      summary: Reemplaza los campos editables
      description: |
        Reemplazo total, no parcial. `id`, `slug`, `status`, `verification` y
        las marcas de tiempo no son editables.

        El slug no se regenera al cambiar el nombre: es una URL pública y
        romperla rompe enlaces y posicionamiento.

        Cambiar la matrícula o la especialidad devuelve `verification` a
        `pending`: la verificación se hizo sobre esos datos.

        Funciona sobre profesionales inactivos y no cambia su estado.
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/ProfessionalRequest' }
      responses:
        '200':
          description: Actualizado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Professional' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '404': { $ref: '#/components/responses/NotFound' }
        '409': { $ref: '#/components/responses/Conflict' }
        '422': { $ref: '#/components/responses/ValidationFailed' }

    delete:
      tags: [professionals]
      summary: Da de baja al profesional
      description: |
        Baja lógica, no borrado. Pone `status` en `inactive` y sella
        `deactivatedAt`. El registro se queda: los turnos y comprobantes
        históricos apuntan a él.

        Es idempotente: dar de baja algo ya dado de baja devuelve 204.
      responses:
        '204': { description: Dado de baja }
        '400': { $ref: '#/components/responses/BadRequest' }
        '404': { $ref: '#/components/responses/NotFound' }

  /api/v1/professionals/{id}/reactivate:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }
    post:
      tags: [professionals]
      summary: Revierte la baja
      description: Idempotente. Sobre alguien ya activo devuelve 200 sin cambios.
      responses:
        '200':
          description: Reactivado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Professional' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '404': { $ref: '#/components/responses/NotFound' }

  /api/v1/professionals/by-slug/{slug}:
    parameters:
      - name: slug
        in: path
        required: true
        schema: { type: string }
    get:
      tags: [professionals]
      summary: Trae un profesional por su slug público
      description: Es lo que consume la página `/p/{slug}` del frontend.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Professional' }
        '404': { $ref: '#/components/responses/NotFound' }

components:
  schemas:
    Especialidad:
      type: string
      enum: [psicologia, kinesiologia, odontologia]

    Modalidad:
      type: string
      enum: [telemedicina, presencial, domicilio]

    Status:
      type: string
      enum: [active, inactive]
      description: Si el profesional opera hoy en la plataforma

    Verification:
      type: string
      enum: [pending, verified, rejected]
      description: Si su matrícula fue verificada contra el mundo real

    Professional:
      type: object
      required:
        - id
        - slug
        - firstName
        - lastName
        - matricula
        - especialidad
        - bio
        - consultaPriceCents
        - modalidades
        - zona
        - obrasSociales
        - status
        - verification
        - createdAt
        - updatedAt
      properties:
        id:
          type: string
          format: uuid
        slug:
          type: string
          example: martin-gonzalez
        firstName:
          type: string
          example: Martín
        lastName:
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
        consultaPriceCents:
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
        status:
          $ref: '#/components/schemas/Status'
        verification:
          $ref: '#/components/schemas/Verification'
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
        deactivatedAt:
          type: [string, 'null']
          format: date-time

    ProfessionalRequest:
      type: object
      required:
        - firstName
        - lastName
        - matricula
        - especialidad
        - consultaPriceCents
        - modalidades
        - zona
      properties:
        firstName:
          type: string
          maxLength: 100
        lastName:
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
        consultaPriceCents:
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

    ProfessionalList:
      type: object
      required: [data, pagination]
      properties:
        data:
          type: array
          items: { $ref: '#/components/schemas/Professional' }
        pagination:
          type: object
          required: [total, limit, offset]
          properties:
            total:
              type: integer
              description: Cuántos hay en total con esos filtros, no en la página
            limit: { type: integer }
            offset: { type: integer }

    Problem:
      type: object
      description: RFC 7807 Problem Details
      required: [type, title, status]
      properties:
        type: { type: string, format: uri }
        title: { type: string }
        status: { type: integer }
        detail: { type: string }
        errors:
          type: array
          description: Presente solo en errores de validación
          items:
            type: object
            required: [field, message]
            properties:
              field: { type: string }
              message: { type: string }

  responses:
    BadRequest:
      description: JSON malformado o parámetro con formato inválido
      content:
        application/problem+json:
          schema: { $ref: '#/components/schemas/Problem' }

    NotFound:
      description: No existe
      content:
        application/problem+json:
          schema: { $ref: '#/components/schemas/Problem' }

    Conflict:
      description: La matrícula ya está registrada por otro profesional
      content:
        application/problem+json:
          schema: { $ref: '#/components/schemas/Problem' }

    ValidationFailed:
      description: |
        JSON válido pero datos inválidos. Distinto de 400: "entendí perfecto
        y está mal" es otro problema que "no entiendo lo que mandaste".
      content:
        application/problem+json:
          schema:
            allOf:
              - $ref: '#/components/schemas/Problem'
              - type: object
                required: [errors]
```

- [ ] **Step 2: Validar que el YAML es sintácticamente correcto**

Run: `cd apps/api && python -c "import yaml,sys; yaml.safe_load(open('api/openapi.yaml',encoding='utf-8')); print('YAML OK')"`
Expected: `YAML OK`

Si Python no está disponible, cualquier validador de YAML sirve. La validación semántica contra el estándar OpenAPI se agrega en la Task 15 junto con el CI.

- [ ] **Step 3: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/api/openapi.yaml
git commit -m "docs(api): contrato OpenAPI del CRUD de Professional

Escrito a mano y antes que los handlers: el YAML es la fuente de verdad.
Sirve para probar la API con Swagger UI mientras no exista el front, y
después para generar el cliente TypeScript.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 9: Errores HTTP y DTOs

La traducción entre el dominio y el mundo HTTP. Es la única capa que conoce códigos de estado.

**Files:**
- Create: `apps/api/internal/handler/problem.go`
- Create: `apps/api/internal/handler/dto.go`
- Test: `apps/api/internal/handler/problem_test.go`
- Test: `apps/api/internal/handler/dto_test.go`

**Interfaces:**
- Consumes: `domain` (Tasks 2-4)
- Produces:
  - `handler.Problem` con etiquetas JSON `type`, `title`, `status`, `detail`, `errors`
  - `func writeProblem(http.ResponseWriter, Problem)`
  - `func writeError(http.ResponseWriter, *http.Request, error)` — mapea errores de dominio a HTTP
  - `func writeJSON(http.ResponseWriter, int, any)`
  - `func decodeJSON(http.ResponseWriter, *http.Request, any) error`
  - `professionalResponse`, `professionalRequest`, `listResponse` con sus conversores

- [ ] **Step 1: Escribir los tests del mapeo de errores**

Archivo `apps/api/internal/handler/problem_test.go`:

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

func TestWriteErrorMapeaLosErroresDelDominio(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"no encontrado", domain.ErrNotFound, http.StatusNotFound},
		{"matricula tomada", domain.ErrMatriculaTaken, http.StatusConflict},
		{
			"validacion",
			domain.ValidationError{Fields: []domain.FieldError{{Field: "zona", Message: "es obligatoria"}}},
			http.StatusUnprocessableEntity,
		},
		{"desconocido", errors.New("algo explotó"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			writeError(rec, req, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, se esperaba %d", rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); ct != problemContentType {
				t.Errorf("Content-Type = %q, se esperaba %q", ct, problemContentType)
			}

			var p Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
				t.Fatalf("el cuerpo no es JSON válido: %v", err)
			}
			if p.Status != tt.wantStatus {
				t.Errorf("problem.status = %d, se esperaba %d", p.Status, tt.wantStatus)
			}
			if p.Title == "" || p.Type == "" {
				t.Error("title y type son obligatorios en RFC 7807")
			}
		})
	}
}

func TestWriteErrorNoFiltraDetallesInternos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// un error interno no puede llegarle al cliente: puede tener nombres de
	// tablas, rutas del servidor o datos de otro usuario
	writeError(rec, req, errors.New("pq: relation \"professionals\" does not exist"))

	if strings.Contains(rec.Body.String(), "professionals") {
		t.Error("el error interno se filtró al cliente")
	}
}

func TestWriteErrorDeValidacionIncluyeLosCampos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	verr := domain.ValidationError{Fields: []domain.FieldError{
		{Field: "matricula", Message: "formato inválido"},
		{Field: "modalidades", Message: "se requiere al menos una"},
	}}
	writeError(rec, req, verr)

	var p Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	if len(p.Errors) != 2 {
		t.Fatalf("errors tenía %d elementos, se esperaban 2", len(p.Errors))
	}
	if p.Errors[0].Field != "matricula" {
		t.Errorf("errors[0].field = %q", p.Errors[0].Field)
	}
}

func TestDecodeJSONRechazaCamposDesconocidos(t *testing.T) {
	rec := httptest.NewRecorder()
	// "consultaPrice" en vez de "consultaPriceCents" es exactamente el typo
	// que este modo estricto tiene que atrapar
	body := strings.NewReader(`{"firstName":"Ana","consultaPrice":100}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var dst professionalRequest
	if err := decodeJSON(rec, req, &dst); err == nil {
		t.Error("un campo desconocido debía ser rechazado")
	}
}

func TestDecodeJSONRechazaBasuraDespuesDelObjeto(t *testing.T) {
	rec := httptest.NewRecorder()
	body := strings.NewReader(`{"firstName":"Ana"} {"firstName":"Otro"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var dst professionalRequest
	if err := decodeJSON(rec, req, &dst); err == nil {
		t.Error("dos objetos JSON seguidos debían ser rechazados")
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: FAIL con `undefined: writeError`

- [ ] **Step 3: Implementar `problem.go`**

Archivo `apps/api/internal/handler/problem.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const problemContentType = "application/problem+json"

// Los tipos de error del contrato. Son URIs por convención de RFC 7807: no
// hace falta que resuelvan a una página, alcanza con que identifiquen la clase
// de problema de forma estable.
const (
	typeBadRequest = "https://salud.app/errors/bad-request"
	typeValidation = "https://salud.app/errors/validation"
	typeNotFound   = "https://salud.app/errors/not-found"
	typeConflict   = "https://salud.app/errors/conflict"
	typeInternal   = "https://salud.app/errors/internal"
)

// Problem es la representación de un error según RFC 7807.
type Problem struct {
	Type   string              `json:"type"`
	Title  string              `json:"title"`
	Status int                 `json:"status"`
	Detail string              `json:"detail,omitempty"`
	Errors []domain.FieldError `json:"errors,omitempty"`
}

func writeProblem(w http.ResponseWriter, p Problem) {
	w.Header().Set("Content-Type", problemContentType)
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError traduce un error del dominio al problema HTTP que le corresponde.
//
// Es el único lugar del proyecto donde el dominio se convierte en códigos de
// estado. Las capas de abajo no saben que existe HTTP.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var verr domain.ValidationError

	switch {
	case errors.As(err, &verr):
		writeProblem(w, Problem{
			Type:   typeValidation,
			Title:  "Datos inválidos",
			Status: http.StatusUnprocessableEntity,
			Detail: "Uno o más campos no cumplen las reglas del sistema",
			Errors: verr.Fields,
		})

	case errors.Is(err, domain.ErrNotFound):
		writeProblem(w, Problem{
			Type:   typeNotFound,
			Title:  "No encontrado",
			Status: http.StatusNotFound,
			Detail: "El profesional solicitado no existe",
		})

	case errors.Is(err, domain.ErrMatriculaTaken):
		writeProblem(w, Problem{
			Type:   typeConflict,
			Title:  "Matrícula ya registrada",
			Status: http.StatusConflict,
			Detail: "Otro profesional ya tiene registrada esa matrícula",
		})

	default:
		// El error real va al log, nunca al cliente: puede contener nombres
		// de tablas, rutas del servidor o datos de otro usuario.
		slog.ErrorContext(r.Context(), "error no manejado",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
		)
		writeProblem(w, Problem{
			Type:   typeInternal,
			Title:  "Error interno",
			Status: http.StatusInternalServerError,
			Detail: "Ocurrió un error inesperado. Volvé a intentar.",
		})
	}
}

func writeBadRequest(w http.ResponseWriter, detail string) {
	writeProblem(w, Problem{
		Type:   typeBadRequest,
		Title:  "Petición inválida",
		Status: http.StatusBadRequest,
		Detail: detail,
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

const maxBodyBytes = 1 << 20 // 1 MB

// professionalRequest es lo que entra. Deliberadamente no incluye id, slug,
// status, verification ni marcas de tiempo: son campos que el servidor decide,
// y aceptarlos sería dejar que el cliente se autoverifique.
type professionalRequest struct {
	FirstName          string   `json:"firstName"`
	LastName           string   `json:"lastName"`
	Matricula          string   `json:"matricula"`
	Especialidad       string   `json:"especialidad"`
	Bio                string   `json:"bio"`
	ConsultaPriceCents int64    `json:"consultaPriceCents"`
	Modalidades        []string `json:"modalidades"`
	Zona               string   `json:"zona"`
	ObrasSociales      []string `json:"obrasSociales"`
}

func (r professionalRequest) toInput() domain.ProfessionalInput {
	return domain.ProfessionalInput{
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Matricula:     r.Matricula,
		Especialidad:  r.Especialidad,
		Bio:           r.Bio,
		ConsultaPrice: r.ConsultaPriceCents,
		Modalidades:   r.Modalidades,
		Zona:          r.Zona,
		ObrasSociales: r.ObrasSociales,
	}
}

type professionalResponse struct {
	ID                 string     `json:"id"`
	Slug               string     `json:"slug"`
	FirstName          string     `json:"firstName"`
	LastName           string     `json:"lastName"`
	Matricula          string     `json:"matricula"`
	Especialidad       string     `json:"especialidad"`
	Bio                string     `json:"bio"`
	ConsultaPriceCents int64      `json:"consultaPriceCents"`
	Modalidades        []string   `json:"modalidades"`
	Zona               string     `json:"zona"`
	ObrasSociales      []string   `json:"obrasSociales"`
	Status             string     `json:"status"`
	Verification       string     `json:"verification"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DeactivatedAt      *time.Time `json:"deactivatedAt"`
}

func toResponse(p domain.Professional) professionalResponse {
	// make con len 0 en vez de nil: un slice nil se serializa como null y el
	// cliente TypeScript tendría que chequearlo en cada uso
	mods := make([]string, 0, len(p.Modalidades))
	for _, m := range p.Modalidades {
		mods = append(mods, string(m))
	}

	obras := make([]string, 0, len(p.ObrasSociales))
	obras = append(obras, p.ObrasSociales...)

	return professionalResponse{
		ID:                 p.ID.String(),
		Slug:               p.Slug,
		FirstName:          p.FirstName,
		LastName:           p.LastName,
		Matricula:          p.Matricula.String(),
		Especialidad:       string(p.Especialidad),
		Bio:                p.Bio,
		ConsultaPriceCents: int64(p.ConsultaPrice),
		Modalidades:        mods,
		Zona:               p.Zona,
		ObrasSociales:      obras,
		Status:             string(p.Status),
		Verification:       string(p.Verification),
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		DeactivatedAt:      p.DeactivatedAt,
	}
}

type paginationResponse struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type listResponse struct {
	Data       []professionalResponse `json:"data"`
	Pagination paginationResponse     `json:"pagination"`
}

func toListResponse(ps []domain.Professional, total, limit, offset int) listResponse {
	data := make([]professionalResponse, 0, len(ps))
	for _, p := range ps {
		data = append(data, toResponse(p))
	}
	return listResponse{
		Data:       data,
		Pagination: paginationResponse{Total: total, Limit: limit, Offset: offset},
	}
}

// decodeJSON lee el cuerpo en modo estricto.
//
// DisallowUnknownFields atrapa el typo más probable de esta API: mandar
// "consultaPrice" en vez de "consultaPriceCents" y que el precio quede en cero
// sin que nadie se entere.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
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

func TestToResponseSerializaLosCamposDelContrato(t *testing.T) {
	now := time.Date(2026, 8, 21, 14, 2, 11, 0, time.UTC)
	p, err := domain.NewProfessional(domain.ProfessionalInput{
		FirstName:     "Martín",
		LastName:      "González",
		Matricula:     "MN 98.234",
		Especialidad:  "psicologia",
		Bio:           "Psicólogo clínico.",
		ConsultaPrice: 1200000,
		Modalidades:   []string{"telemedicina", "presencial"},
		Zona:          "CABA",
		ObrasSociales: []string{"OSDE"},
	}, now)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	raw, err := json.Marshal(toResponse(p))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	body := string(raw)

	// el precio viaja como entero, sin punto decimal
	if !strings.Contains(body, `"consultaPriceCents":1200000`) {
		t.Errorf("el precio no viaja como entero de centavos: %s", body)
	}
	// la matrícula viaja en forma canónica
	if !strings.Contains(body, `"matricula":"MN 98234"`) {
		t.Errorf("la matrícula no viaja canónica: %s", body)
	}
	if !strings.Contains(body, `"deactivatedAt":null`) {
		t.Errorf("deactivatedAt debía estar presente como null: %s", body)
	}
}

func TestToResponseNuncaSerializaSlicesComoNull(t *testing.T) {
	// un slice nil se serializa como null y obliga al cliente TypeScript a
	// chequearlo en cada uso
	var p domain.Professional
	raw, err := json.Marshal(toResponse(p))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	body := string(raw)

	if strings.Contains(body, `"modalidades":null`) {
		t.Error("modalidades se serializó como null en vez de []")
	}
	if strings.Contains(body, `"obrasSociales":null`) {
		t.Error("obrasSociales se serializó como null en vez de []")
	}
}

func TestToListResponseConListaVacia(t *testing.T) {
	raw, err := json.Marshal(toListResponse(nil, 0, 20, 0))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	if strings.Contains(string(raw), `"data":null`) {
		t.Errorf("data se serializó como null en vez de []: %s", raw)
	}
}
```

- [ ] **Step 6: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: PASS en los tests de `problem_test.go` y `dto_test.go`.

- [ ] **Step 7: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/handler/
git commit -m "feat(handler): errores RFC 7807 y DTOs

writeError es el único lugar donde el dominio se convierte en códigos
HTTP. Los errores internos van al log, nunca al cliente.

decodeJSON rechaza campos desconocidos: atrapa el typo de mandar
consultaPrice en vez de consultaPriceCents, que dejaría el precio en
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
- Consumes: `problem.go` (Task 9). No depende de los handlers: por eso va antes.
- Produces:
  - `func handler.Chain(http.Handler, ...func(http.Handler) http.Handler) http.Handler`
  - `func handler.RequestID(http.Handler) http.Handler`
  - `func handler.LogRequests(http.Handler) http.Handler`
  - `func handler.RecoverPanic(http.Handler) http.Handler`

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

func TestRequestIDGeneraUnoSiNoViene(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if seen == "" {
		t.Error("no se generó un request id")
	}
	if rec.Header().Get("X-Request-ID") != seen {
		t.Error("el request id tenía que volver en el header de la respuesta")
	}
}

func TestRequestIDRespetaElQueViene(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "trae-el-suyo")

	h.ServeHTTP(httptest.NewRecorder(), req)

	// permite seguir un request a través de varios servicios
	if seen != "trae-el-suyo" {
		t.Errorf("request id = %q, se esperaba el que vino en el header", seen)
	}
}

func TestRecoverPanicDevuelve500(t *testing.T) {
	h := RecoverPanic(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("algo explotó")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, se esperaba 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != problemContentType {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, problemContentType)
	}
	// el mensaje del panic no puede llegarle al cliente
	if strings.Contains(rec.Body.String(), "algo explotó") {
		t.Error("el mensaje del panic se filtró al cliente")
	}
}

func TestChainAplicaEnOrden(t *testing.T) {
	var order []string

	mark := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	final := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		order = append(order, "handler")
	})

	Chain(final, mark("primero"), mark("segundo")).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"primero", "segundo", "handler"}
	if len(order) != len(want) {
		t.Fatalf("orden = %v, se esperaba %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("orden = %v, se esperaba %v", order, want)
		}
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/handler/ -run 'TestRequestID|TestRecover|TestChain' -v`
Expected: FAIL con `undefined: RequestID`

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

const requestIDHeader = "X-Request-ID"

type contextKey string

const requestIDKey contextKey = "requestID"

// Chain envuelve el handler. El primer middleware de la lista es el más
// externo: el primero en ver el request y el último en ver la respuesta.
func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// RequestID asegura que todo request tenga un identificador. Si el cliente
// manda uno, se respeta: permite seguir una operación a través de varios
// servicios en los logs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

// RequestIDFrom devuelve el identificador del request, o cadena vacía si el
// middleware no corrió.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// statusRecorder captura el código de estado para poder loguearlo: el
// http.ResponseWriter no lo expone.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		slog.InfoContext(r.Context(), "request",
			"requestId", RequestIDFrom(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"durationMs", time.Since(start).Milliseconds(),
		)
	})
}

// RecoverPanic evita que un panic en un handler tire el proceso entero.
//
// Va por dentro de LogRequests para que el log registre el 500 resultante.
func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// el detalle del panic va al log, nunca al cliente
			slog.ErrorContext(r.Context(), "panic recuperado",
				"requestId", RequestIDFrom(r.Context()),
				"panic", rec,
				"method", r.Method,
				"path", r.URL.Path,
			)

			writeProblem(w, Problem{
				Type:   typeInternal,
				Title:  "Error interno",
				Status: http.StatusInternalServerError,
				Detail: "Ocurrió un error inesperado. Volvé a intentar.",
			})
		}()

		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4: Correr toda la suite del paquete handler**

Run: `cd apps/api && go test ./internal/handler/ -v`
Expected: PASS en los tests del middleware y en los de `problem.go` y `dto.go` de la Task 9.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/handler/
git commit -m "feat(handler): middleware de request id, logging y recover

RequestID respeta el que manda el cliente: permite seguir una operación
a través de varios servicios en los logs.

RecoverPanic evita que un handler tire el proceso, y el detalle del
panic va al log y nunca al cliente.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 11: Handlers y router

**Files:**
- Create: `apps/api/internal/handler/professional.go`
- Create: `apps/api/internal/handler/router.go`
- Test: `apps/api/internal/handler/professional_test.go`

**Interfaces:**
- Consumes: `service.Professional` (Tasks 6-7), `problem.go` y `dto.go` (Task 9), el middleware (Task 10)
- Produces:
  - `func handler.NewProfessional(*service.Professional) *handler.ProfessionalHandler`
  - Métodos `Create`, `List`, `GetByID`, `GetBySlug`, `Update`, `Deactivate`, `Reactivate`
  - `func handler.NewRouter(*ProfessionalHandler) http.Handler`

- [ ] **Step 1: Escribir los tests de la capa HTTP**

Estos tests corren contra el stack completo cableado: router, handler, servicio y repositorio en memoria. No hay mocks.

Archivo `apps/api/internal/handler/professional_test.go`:

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

// newTestServer cablea el stack real de punta a punta.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	repo := memory.NewProfessional()
	svc := service.NewProfessional(repo)
	srv := httptest.NewServer(NewRouter(NewProfessional(svc)))
	t.Cleanup(srv.Close)
	return srv
}

const validBody = `{
  "firstName": "Martín",
  "lastName": "González",
  "matricula": "MN 98.234",
  "especialidad": "psicologia",
  "bio": "Psicólogo clínico.",
  "consultaPriceCents": 1200000,
  "modalidades": ["telemedicina", "presencial"],
  "zona": "CABA",
  "obrasSociales": ["OSDE"]
}`

func post(t *testing.T, srv *httptest.Server, path, body string) *http.Response {
	t.Helper()
	resp, err := srv.Client().Post(srv.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s falló: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func get(t *testing.T, srv *httptest.Server, path string) *http.Response {
	t.Helper()
	resp, err := srv.Client().Get(srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s falló: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func do(t *testing.T, srv *httptest.Server, method, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("no se pudo armar el request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s falló: %v", method, path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decodeProfessional(t *testing.T, resp *http.Response) professionalResponse {
	t.Helper()
	var p professionalResponse
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	return p
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t)
	resp := get(t, srv, "/healthz")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, se esperaba 200", resp.StatusCode)
	}
}

func TestCreateDevuelve201YLocation(t *testing.T) {
	srv := newTestServer(t)
	resp := post(t, srv, "/api/v1/professionals", validBody)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201", resp.StatusCode)
	}

	p := decodeProfessional(t, resp)
	if p.Slug != "martin-gonzalez" {
		t.Errorf("slug = %q, se esperaba martin-gonzalez", p.Slug)
	}
	if p.Verification != "pending" {
		t.Errorf("verification = %q, se esperaba pending", p.Verification)
	}

	loc := resp.Header.Get("Location")
	if loc != "/api/v1/professionals/"+p.ID {
		t.Errorf("Location = %q, se esperaba la URL del recurso creado", loc)
	}
}

func TestCreateJSONMalformadoDevuelve400(t *testing.T) {
	srv := newTestServer(t)
	resp := post(t, srv, "/api/v1/professionals", `{"firstName":`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != problemContentType {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, problemContentType)
	}
}

func TestCreateDatosInvalidosDevuelve422(t *testing.T) {
	srv := newTestServer(t)
	// JSON perfectamente válido, datos mal: es otro problema que un 400
	body := `{
	  "firstName": "",
	  "lastName": "González",
	  "matricula": "roto",
	  "especialidad": "cardiologia",
	  "consultaPriceCents": -5,
	  "modalidades": [],
	  "zona": ""
	}`
	resp := post(t, srv, "/api/v1/professionals", body)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, se esperaba 422", resp.StatusCode)
	}

	var p Problem
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("no se pudo decodificar el problem: %v", err)
	}
	if len(p.Errors) < 5 {
		t.Errorf("se esperaban al menos 5 campos con error, se obtuvieron %d: %+v", len(p.Errors), p.Errors)
	}
}

func TestCreateMatriculaDuplicadaDevuelve409(t *testing.T) {
	srv := newTestServer(t)
	post(t, srv, "/api/v1/professionals", validBody)
	resp := post(t, srv, "/api/v1/professionals", validBody)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, se esperaba 409", resp.StatusCode)
	}
}

func TestGetByIDYPorSlug(t *testing.T) {
	srv := newTestServer(t)
	created := decodeProfessional(t, post(t, srv, "/api/v1/professionals", validBody))

	byID := get(t, srv, "/api/v1/professionals/"+created.ID)
	if byID.StatusCode != http.StatusOK {
		t.Errorf("GET por id: status = %d, se esperaba 200", byID.StatusCode)
	}

	bySlug := get(t, srv, "/api/v1/professionals/by-slug/"+created.Slug)
	if bySlug.StatusCode != http.StatusOK {
		t.Errorf("GET por slug: status = %d, se esperaba 200", bySlug.StatusCode)
	}
	if decodeProfessional(t, bySlug).ID != created.ID {
		t.Error("GET por slug devolvió otro profesional")
	}
}

func TestGetIDInexistenteDevuelve404(t *testing.T) {
	srv := newTestServer(t)
	resp := get(t, srv, "/api/v1/professionals/6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, se esperaba 404", resp.StatusCode)
	}
}

func TestGetIDMalFormadoDevuelve400(t *testing.T) {
	srv := newTestServer(t)
	// no es un UUID: es un problema del cliente, no un recurso que falta
	resp := get(t, srv, "/api/v1/professionals/no-soy-un-uuid")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestListPaginaYFiltra(t *testing.T) {
	srv := newTestServer(t)
	post(t, srv, "/api/v1/professionals", validBody)

	segundo := strings.Replace(validBody, `"MN 98.234"`, `"MN 11111"`, 1)
	segundo = strings.Replace(segundo, `"psicologia"`, `"odontologia"`, 1)
	post(t, srv, "/api/v1/professionals", segundo)

	t.Run("sin filtros", func(t *testing.T) {
		resp := get(t, srv, "/api/v1/professionals")
		var list listResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if list.Pagination.Total != 2 {
			t.Errorf("total = %d, se esperaba 2", list.Pagination.Total)
		}
		if list.Pagination.Limit != service.DefaultLimit {
			t.Errorf("limit = %d, se esperaba el default %d", list.Pagination.Limit, service.DefaultLimit)
		}
	})

	t.Run("por especialidad", func(t *testing.T) {
		resp := get(t, srv, "/api/v1/professionals?especialidad=odontologia")
		var list listResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if list.Pagination.Total != 1 {
			t.Errorf("total = %d, se esperaba 1", list.Pagination.Total)
		}
	})

	t.Run("busqueda sin acentos", func(t *testing.T) {
		resp := get(t, srv, "/api/v1/professionals?q=gonzalez")
		var list listResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if list.Pagination.Total != 2 {
			t.Errorf("total = %d, se esperaban 2: ambos apellidan González", list.Pagination.Total)
		}
	})

	t.Run("limit invalido devuelve 400", func(t *testing.T) {
		resp := get(t, srv, "/api/v1/professionals?limit=abc")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
		}
	})

	t.Run("especialidad invalida devuelve 400", func(t *testing.T) {
		resp := get(t, srv, "/api/v1/professionals?especialidad=cardiologia")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
		}
	})
}

func TestUpdate(t *testing.T) {
	srv := newTestServer(t)
	created := decodeProfessional(t, post(t, srv, "/api/v1/professionals", validBody))

	body := strings.Replace(validBody, `"CABA"`, `"GBA Norte"`, 1)
	resp := do(t, srv, http.MethodPut, "/api/v1/professionals/"+created.ID, body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}
	updated := decodeProfessional(t, resp)
	if updated.Zona != "GBA Norte" {
		t.Errorf("zona = %q, se esperaba GBA Norte", updated.Zona)
	}
	if updated.Slug != created.Slug {
		t.Error("el slug es una URL pública y no debía cambiar")
	}
}

func TestDeleteEsBajaLogicaYIdempotente(t *testing.T) {
	srv := newTestServer(t)
	created := decodeProfessional(t, post(t, srv, "/api/v1/professionals", validBody))

	resp := do(t, srv, http.MethodDelete, "/api/v1/professionals/"+created.ID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, se esperaba 204", resp.StatusCode)
	}

	// el recurso sigue existiendo: no fue un borrado
	after := get(t, srv, "/api/v1/professionals/"+created.ID)
	if after.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: el profesional dado de baja tenía que seguir existiendo", after.StatusCode)
	}
	if p := decodeProfessional(t, after); p.Status != "inactive" || p.DeactivatedAt == nil {
		t.Error("debía quedar inactive con deactivatedAt sellado")
	}

	// pero no aparece en el listado por defecto
	list := get(t, srv, "/api/v1/professionals")
	var lr listResponse
	if err := json.NewDecoder(list.Body).Decode(&lr); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if lr.Pagination.Total != 0 {
		t.Errorf("total = %d, se esperaba 0", lr.Pagination.Total)
	}

	// y una segunda baja no es un error
	again := do(t, srv, http.MethodDelete, "/api/v1/professionals/"+created.ID, "")
	if again.StatusCode != http.StatusNoContent {
		t.Errorf("la segunda baja devolvió %d, se esperaba 204", again.StatusCode)
	}
}

func TestReactivate(t *testing.T) {
	srv := newTestServer(t)
	created := decodeProfessional(t, post(t, srv, "/api/v1/professionals", validBody))
	do(t, srv, http.MethodDelete, "/api/v1/professionals/"+created.ID, "")

	resp := post(t, srv, "/api/v1/professionals/"+created.ID+"/reactivate", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	p := decodeProfessional(t, resp)
	if p.Status != "active" || p.DeactivatedAt != nil {
		t.Error("debía quedar activo y con deactivatedAt en null")
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/handler/ -run 'TestHealthz|TestCreate|TestGet|TestList|TestUpdate|TestDelete|TestReactivate' -v`
Expected: FAIL con `undefined: NewRouter`

- [ ] **Step 3: Implementar los handlers**

Archivo `apps/api/internal/handler/professional.go`:

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

// ProfessionalHandler traduce entre HTTP y el servicio. No toma decisiones de
// negocio: decodifica, delega y serializa.
type ProfessionalHandler struct {
	svc *service.Professional
}

func NewProfessional(svc *service.Professional) *ProfessionalHandler {
	return &ProfessionalHandler{svc: svc}
}

func (h *ProfessionalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req professionalRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeBadRequest(w, "el cuerpo no es un JSON válido: "+err.Error())
		return
	}

	p, err := h.svc.Create(r.Context(), req.toInput())
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/professionals/"+p.ID.String())
	writeJSON(w, http.StatusCreated, toResponse(p))
}

func (h *ProfessionalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(p))
}

func (h *ProfessionalHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(p))
}

func (h *ProfessionalHandler) List(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilter(r)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	ps, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, r, err)
		return
	}

	limit := f.Limit
	if limit <= 0 {
		limit = service.DefaultLimit
	}
	writeJSON(w, http.StatusOK, toListResponse(ps, total, limit, f.Offset))
}

func (h *ProfessionalHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req professionalRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeBadRequest(w, "el cuerpo no es un JSON válido: "+err.Error())
		return
	}

	p, err := h.svc.Update(r.Context(), id, req.toInput())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(p))
}

// Deactivate implementa el DELETE. Se llama Deactivate y no Delete porque eso
// es lo que hace: baja lógica. El verbo HTTP conserva el nombre que espera
// cualquiera que lea un CRUD.
func (h *ProfessionalHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Deactivate(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProfessionalHandler) Reactivate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.Reactivate(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(p))
}

// parseID devuelve false y ya escribió la respuesta si el ID no es un UUID.
// Un ID mal formado es un error del cliente (400), no un recurso que falta.
func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := r.PathValue("id")
	id, err := uuid.Parse(raw)
	if err != nil {
		writeBadRequest(w, "el id debe ser un UUID válido")
		return uuid.Nil, false
	}
	return id, true
}

func parseFilter(r *http.Request) (repository.Filter, error) {
	q := r.URL.Query()
	var f repository.Filter

	if raw := q.Get("especialidad"); raw != "" {
		esp := domain.Especialidad(raw)
		if !esp.Valid() {
			return f, errInvalidQuery("especialidad", "debe ser psicologia, kinesiologia u odontologia")
		}
		f.Especialidad = &esp
	}

	if raw := q.Get("status"); raw != "" {
		st := domain.Status(raw)
		if !st.Valid() {
			return f, errInvalidQuery("status", "debe ser active o inactive")
		}
		f.Status = &st
	}

	if raw := q.Get("zona"); raw != "" {
		f.Zona = &raw
	}
	if raw := q.Get("q"); raw != "" {
		f.Query = &raw
	}

	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			return f, errInvalidQuery("limit", "debe ser un entero mayor a cero")
		}
		f.Limit = v
	}

	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return f, errInvalidQuery("offset", "debe ser un entero mayor o igual a cero")
		}
		f.Offset = v
	}

	return f, nil
}

type queryError struct {
	param   string
	message string
}

func (e queryError) Error() string {
	return "parámetro " + e.param + ": " + e.message
}

func errInvalidQuery(param, message string) error {
	return queryError{param: param, message: message}
}
```

- [ ] **Step 4: Implementar el router**

Archivo `apps/api/internal/handler/router.go`:

```go
package handler

import "net/http"

// NewRouter arma la tabla de rutas.
//
// Usa el ServeMux de la stdlib: desde Go 1.22 entiende método y parámetros de
// ruta, así que no hace falta chi, gin ni echo para esto.
func NewRouter(ph *ProfessionalHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)

	mux.HandleFunc("GET /api/v1/professionals", ph.List)
	mux.HandleFunc("POST /api/v1/professionals", ph.Create)
	mux.HandleFunc("GET /api/v1/professionals/{id}", ph.GetByID)
	mux.HandleFunc("PUT /api/v1/professionals/{id}", ph.Update)
	mux.HandleFunc("DELETE /api/v1/professionals/{id}", ph.Deactivate)
	mux.HandleFunc("POST /api/v1/professionals/{id}/reactivate", ph.Reactivate)

	// No colisiona con /{id}: tiene un segmento más y el ServeMux resuelve
	// por especificidad.
	mux.HandleFunc("GET /api/v1/professionals/by-slug/{slug}", ph.GetBySlug)

	// El orden es de afuera hacia adentro. RequestID va primero para que el
	// log lo tenga; LogRequests envuelve a RecoverPanic para que un panic
	// quede registrado con su 500.
	return Chain(mux, RequestID, LogRequests, RecoverPanic)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

## Task 12: Seed de desarrollo

**Files:**
- Create: `apps/api/internal/repository/memory/seed.go`
- Test: `apps/api/internal/repository/memory/seed_test.go`

**Interfaces:**
- Consumes: `domain` (Task 4), `memory.Professional` (Task 5)
- Produces: `func memory.Seed(context.Context, *Professional) error`

- [ ] **Step 1: Escribir el test**

Archivo `apps/api/internal/repository/memory/seed_test.go`:

```go
package memory

import (
	"context"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

func TestSeed(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()

	if err := Seed(ctx, repo); err != nil {
		t.Fatalf("Seed devolvió error: %v", err)
	}

	_, total, err := repo.List(ctx, repository.Filter{Limit: 100})
	if err != nil {
		t.Fatalf("List devolvió error: %v", err)
	}
	if total != 4 {
		t.Errorf("total = %d, se esperaban 4 profesionales de prueba", total)
	}
}

func TestSeedGeneraSlugsUnicos(t *testing.T) {
	ctx := context.Background()
	repo := NewProfessional()
	if err := Seed(ctx, repo); err != nil {
		t.Fatalf("Seed devolvió error: %v", err)
	}

	ps, _, _ := repo.List(ctx, repository.Filter{Limit: 100})

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

Run: `cd apps/api && go test ./internal/repository/memory/ -run TestSeed -v`
Expected: FAIL con `undefined: Seed`

- [ ] **Step 3: Implementar el seed**

Los datos salen de `legacy/prototype/src/data/profesionales.js`, adaptados al modelo nuevo. Los precios del prototipo estaban en pesos; acá van en centavos.

Archivo `apps/api/internal/repository/memory/seed.go`:

```go
package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Seed carga profesionales de prueba. Solo se llama en desarrollo: main.go lo
// invoca únicamente cuando APP_ENV=development.
//
// Los datos vienen del prototipo React, en legacy/prototype/src/data. Los
// precios estaban en pesos y acá van en centavos.
func Seed(ctx context.Context, repo *Professional) error {
	now := time.Now().UTC()

	inputs := []domain.ProfessionalInput{
		{
			FirstName:     "Martín",
			LastName:      "González",
			Matricula:     "MN 98.234",
			Especialidad:  string(domain.EspecialidadPsicologia),
			Bio:           "Psicólogo clínico con orientación cognitivo-conductual. Atiendo adultos y adolescentes con ansiedad, depresión y crisis vitales. Más de 8 años de experiencia.",
			ConsultaPrice: 1_200_000,
			Modalidades:   []string{"telemedicina", "presencial"},
			Zona:          "CABA",
			ObrasSociales: []string{"OSDE", "Swiss Medical", "Galeno", "Medifé"},
		},
		{
			FirstName:     "Carolina",
			LastName:      "Vega",
			Matricula:     "MN 112.087",
			Especialidad:  string(domain.EspecialidadPsicologia),
			Bio:           "Especializada en terapia sistémica y de parejas. Trabajo con adultos en procesos de cambio, duelos y conflictos relacionales.",
			ConsultaPrice: 1_400_000,
			Modalidades:   []string{"telemedicina"},
			Zona:          "GBA Norte",
			ObrasSociales: []string{"OSDE", "OMINT", "Swiss Medical", "Sanitas"},
		},
		{
			FirstName:     "Pablo",
			LastName:      "Moreno",
			Matricula:     "MN 45.321",
			Especialidad:  string(domain.EspecialidadKinesiologia),
			Bio:           "Kinesiólogo especializado en traumatología deportiva y rehabilitación postquirúrgica. Atiendo a domicilio y en consultorio.",
			ConsultaPrice: 950_000,
			Modalidades:   []string{"presencial", "domicilio"},
			Zona:          "CABA",
			ObrasSociales: []string{"OSDE", "Galeno", "IOMA", "PAMI"},
		},
		{
			FirstName:     "Gabriela",
			LastName:      "Ríos",
			Matricula:     "MN 67.890",
			Especialidad:  string(domain.EspecialidadOdontologia),
			Bio:           "Odontóloga general con especialización en estética dental. Trabajo con materiales de primera calidad en un consultorio moderno en Palermo.",
			ConsultaPrice: 1_500_000,
			Modalidades:   []string{"presencial"},
			Zona:          "CABA",
			ObrasSociales: []string{"OSDE", "Swiss Medical", "Galeno", "Medifé", "OMINT"},
		},
	}

	for i, in := range inputs {
		p, err := domain.NewProfessional(in, now.Add(time.Duration(i)*time.Second))
		if err != nil {
			return fmt.Errorf("seed: profesional %d (%s %s): %w", i, in.FirstName, in.LastName, err)
		}
		if err := repo.Create(ctx, p); err != nil {
			return fmt.Errorf("seed: guardando %s %s: %w", in.FirstName, in.LastName, err)
		}
	}
	return nil
}
```

Nota: el seed usa `domain.NewProfessional` directamente en vez del servicio, así que no resuelve colisiones de slug. Los cuatro nombres son distintos, y `TestSeedGeneraSlugsUnicos` lo verifica: si alguien agrega un homónimo, el test lo atrapa.

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/repository/memory/ -v`
Expected: PASS en todos.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/repository/memory/seed.go apps/api/internal/repository/memory/seed_test.go
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
  - `config.Config{Port string; Env string; LogLevel slog.Level; ShutdownTimeout time.Duration}`
  - `func config.Load() (Config, error)`
  - `func (Config) IsDevelopment() bool`
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

func TestLoadDefaults(t *testing.T) {
	// t.Setenv restaura el entorno al terminar el test
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load devolvió error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, se esperaba 8080", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, se esperaba development", cfg.Env)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, se esperaba info", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, se esperaba 10s", cfg.ShutdownTimeout)
	}
	if !cfg.IsDevelopment() {
		t.Error("con APP_ENV vacío tenía que ser desarrollo")
	}
}

func TestLoadDesdeElEntorno(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load devolvió error: %v", err)
	}

	if cfg.Port != "9000" {
		t.Errorf("Port = %q", cfg.Port)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, se esperaba debug", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout = %v", cfg.ShutdownTimeout)
	}
	if cfg.IsDevelopment() {
		t.Error("con APP_ENV=production no tenía que ser desarrollo")
	}
}

func TestLoadFallaRapidoConValoresInvalidos(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"puerto no numerico", "PORT", "ocho-mil"},
		{"puerto fuera de rango", "PORT", "99999"},
		{"nivel de log desconocido", "LOG_LEVEL", "verbose"},
		{"timeout mal formado", "SHUTDOWN_TIMEOUT", "diez segundos"},
		{"entorno desconocido", "APP_ENV", "staging-raro"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			// mejor no arrancar que arrancar mal configurado
			if _, err := Load(); err == nil {
				t.Errorf("%s=%q debía fallar y no falló", tt.key, tt.value)
			}
		})
	}
}
```

- [ ] **Step 2: Correr los tests y verificar que fallan**

Run: `cd apps/api && go test ./internal/config/ -v`
Expected: FAIL con `undefined: Load`

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
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Config es todo lo que el binario necesita saber del entorno. Sin librería:
// son cuatro variables y un struct.
type Config struct {
	Port            string
	Env             string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

func (c Config) IsDevelopment() bool {
	return c.Env == EnvDevelopment
}

// Load lee el entorno y falla si algo está mal.
//
// Falla rápido a propósito: un servidor que arranca con una configuración
// inválida es peor que uno que no arranca, porque el problema aparece más
// tarde y en otro lado.
func Load() (Config, error) {
	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("APP_ENV", EnvDevelopment),
		ShutdownTimeout: 10 * time.Second,
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT inválido: %q", cfg.Port)
	}

	if cfg.Env != EnvDevelopment && cfg.Env != EnvProduction {
		return Config{}, fmt.Errorf("APP_ENV inválido: %q (debe ser %s o %s)", cfg.Env, EnvDevelopment, EnvProduction)
	}

	level, err := parseLogLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	if raw := os.Getenv("SHUTDOWN_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT inválido: %q (ejemplo válido: 30s)", raw)
		}
		cfg.ShutdownTimeout = d
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(raw string) (slog.Level, error) {
	switch raw {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL inválido: %q (debe ser debug, info, warn o error)", raw)
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
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	// El cableado de dependencias, explícito y de arriba abajo. Sin
	// anotaciones ni contenedor: no hay magia que debuggear a las 3 de la
	// mañana.
	//
	// Migrar a PostgreSQL es cambiar esta línea por
	// postgres.NewProfessional(db). Nada más.
	repo := memory.NewProfessional()

	if cfg.IsDevelopment() {
		if err := memory.Seed(context.Background(), repo); err != nil {
			return fmt.Errorf("cargando el seed: %w", err)
		}
		slog.Info("seed de desarrollo cargado")
	}

	svc := service.NewProfessional(repo)
	router := handler.NewRouter(handler.NewProfessional(svc))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
		// Los valores por defecto de http.Server son cero, o sea sin límite:
		// una conexión lenta puede quedarse tomada para siempre.
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("servidor escuchando", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("el servidor falló: %w", err)

	case <-ctx.Done():
		// Apagado gracioso: sin esto, cada deploy corta los requests que
		// están a mitad de camino.
		slog.Info("apagando", "timeout", cfg.ShutdownTimeout)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
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
curl -s "http://localhost:8080/api/v1/professionals?q=gonzalez" | head -c 400
kill %1
```

Expected:
- `{"status":"ok"}`
- Un JSON con `"data":[...]` conteniendo a Martín González y `"total":1`.
- En los logs, líneas JSON con `"msg":"request"` y `"requestId"`.

En PowerShell, el equivalente:
```powershell
cd apps/api
go build ./...
$env:APP_ENV="development"; Start-Job { go run ./cmd/api }
Start-Sleep 3
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod "http://localhost:8080/api/v1/professionals?q=gonzalez"
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
Migrar a PostgreSQL es cambiar la línea de memory.NewProfessional().

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
- `errcheck` sobre `json.NewEncoder(w).Encode(...)`: ya está silenciado con `_ =` en `problem.go`.
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
- `{"status":"ok"}`.
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

Implementar `repository.Professional` en `internal/repository/postgres/` y
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
- Identificadores en inglés; en español solo los términos del dominio que
  no traducen: `Matricula`, `Especialidad`, `ObraSocial`, `Zona`.
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
BASE=http://localhost:8080/api/v1/professionals

# healthz
curl -s http://localhost:8080/healthz

# listado con seed
curl -s "$BASE" | head -c 300

# búsqueda sin acentos: tiene que encontrar a González
curl -s "$BASE?q=gonzalez" | head -c 300

# alta
ID=$(curl -s -X POST "$BASE" -H 'Content-Type: application/json' -d '{
  "firstName":"Ana","lastName":"Pérez","matricula":"MP 55.123",
  "especialidad":"odontologia","bio":"Odontóloga general.",
  "consultaPriceCents":1800000,"modalidades":["presencial"],
  "zona":"CABA","obrasSociales":["OSDE"]
}' | python -c "import sys,json; print(json.load(sys.stdin)['id'])")

# lectura por id y por slug
curl -s "$BASE/$ID" | head -c 200
curl -s "$BASE/by-slug/ana-perez" | head -c 200

# edición
curl -s -X PUT "$BASE/$ID" -H 'Content-Type: application/json' -d '{
  "firstName":"Ana","lastName":"Pérez","matricula":"MP 55.123",
  "especialidad":"odontologia","bio":"Bio editada.",
  "consultaPriceCents":1800000,"modalidades":["presencial","telemedicina"],
  "zona":"GBA Norte","obrasSociales":["OSDE"]
}' | head -c 200

# baja y reactivación
curl -s -o /dev/null -w "DELETE: %{http_code}\n" -X DELETE "$BASE/$ID"
curl -s -o /dev/null -w "DELETE otra vez: %{http_code}\n" -X DELETE "$BASE/$ID"
curl -s -o /dev/null -w "reactivate: %{http_code}\n" -X POST "$BASE/$ID/reactivate"

# los códigos de error
curl -s -o /dev/null -w "JSON roto: %{http_code}\n" -X POST "$BASE" -H 'Content-Type: application/json' -d '{'
curl -s -o /dev/null -w "datos inválidos: %{http_code}\n" -X POST "$BASE" -H 'Content-Type: application/json' -d '{"firstName":"","lastName":"","matricula":"x","especialidad":"y","consultaPriceCents":0,"modalidades":[],"zona":""}'
curl -s -o /dev/null -w "id inexistente: %{http_code}\n" "$BASE/6ba7b810-9dad-11d1-80b4-00c04fd430c8"
curl -s -o /dev/null -w "id mal formado: %{http_code}\n" "$BASE/no-es-uuid"
```

Expected: `DELETE: 204`, `DELETE otra vez: 204`, `reactivate: 200`, `JSON roto: 400`, `datos inválidos: 422`, `id inexistente: 404`, `id mal formado: 400`.

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

Expected: `{"status":"ok"}`.

- [ ] **Step 6: El punto de cambio a PostgreSQL es una sola línea**

Verificación por lectura, sin escribir código: buscar todas las menciones a `memory.` fuera de su propio paquete y de los tests.

Run: `cd apps/api && grep -rn "repository/memory" --include="*.go" . | grep -v "_test.go" | grep -v "internal/repository/memory/"`
Expected: exactamente dos líneas, ambas en `cmd/api/main.go` — el import y la llamada a `memory.NewProfessional()`. La de `memory.Seed` es la tercera y también es esperada. Si aparece cualquier otra, alguna capa se acopló a la implementación y hay que revisarla.

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
| Integración con REFEPS | Spec propia. Hoy `verification` queda en `pending` para siempre. |
| `StatusSuspended` | Cuando REFEPS pueda disparar una suspensión |
| Endpoint de purga (Ley 25.326 art. 16) | Cuando el abogado defina qué se está obligado a conservar |
| Rating, reseñas, horarios, coseguros | Son entidades propias, no campos |
| Cliente TypeScript generado del OpenAPI | Etapa del frontend |
