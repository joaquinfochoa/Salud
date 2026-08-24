# Disponibilidad del Profesional — Plan de Implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Que un profesional pueda cargar su horario semanal y sus bloqueos, y que cualquiera pueda consultar los huecos libres resultantes.

**Architecture:** Se extiende el backend existente sin cambiar su forma. Toda la matemática de calendario vive en `internal/domain` como una función pura sobre entradas explícitas; el servicio carga los datos y arma esas entradas. Los horarios son hora de reloj, no instantes: el `HorarioSemanal` guarda día de la semana y `HoraDelDia`, y solo los huecos calculados son instantes.

**Tech Stack:** Go 1.24+, `net/http` de la stdlib, `log/slog`, `testing` + `net/http/httptest`. Sigue habiendo una sola dependencia externa: `github.com/google/uuid`.

## Global Constraints

Valen para todas las tareas. No se repiten en cada una.

- **Module path:** `github.com/joaquinfochoa/Salud/apps/api`. Correr `go` desde `apps/api`.
- **Go mínimo:** 1.24. No tocar la directiva del `go.mod`.
- **Dependencias externas:** únicamente `github.com/google/uuid`. Cualquier otro `go get` es un error del plan: parar y preguntar.
- **Idioma:** todo lo que escribimos va en español — tipos, funciones, campos, constantes, comentarios, mensajes y claves del JSON. Los paquetes quedan en inglés (`domain`, `repository`, `memory`, `service`, `handler`). Los nombres de archivo van en español. Quedan en inglés por contrato externo: `String()` y `Error()` (interfaces de Go), el prefijo `Test`, las variables de entorno, y las claves `type`/`title`/`status`/`detail` de RFC 7807.
- **Concordancia de género en los enums**, ya establecida: `EsValida()` en femenino (`Especialidad`, `Modalidad`), `EsValido()` en masculino (`Estado`, `EstadoVerificacion`, y ahora `DiaSemana`).
- **`internal/domain` no importa ningún paquete del proyecto.** El CI lo verifica y rompe el build si se viola.
- **Dinero:** `int64` en centavos. Nunca `float64`.
- **Sin mocks.** Los repositorios en memoria son los dobles de test. Si un test parece necesitar un mock, la frontera está mal dibujada: parar y reportar.
- **Comentarios:** en español, explicando el *por qué*, no el *qué*.
- **TDD obligatorio** salvo donde la tarea diga lo contrario: test que falla, correrlo, implementar, correrlo.
- **Los intervalos son semiabiertos** en todo el sistema: `[inicio, fin)`. Un hueco que termina 09:50 y un bloqueo que empieza 09:50 no se tocan.

### Cómo correr las verificaciones en esta máquina

```bash
# El detector de carreras necesita cgo y el compilador de C está fuera del PATH
export PATH="/c/Users/gianl/AppData/Local/Microsoft/WinGet/Packages/BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe/mingw64/bin:$(go env GOPATH)/bin:$PATH"
export CGO_ENABLED=1
```

`golangci-lint` está instalado en `$(go env GOPATH)/bin`.

---

## Estructura de archivos

| Archivo | Responsabilidad | Tarea |
|---|---|---|
| `internal/domain/hora.go` | `HoraDelDia` y su parser | 1 |
| `internal/domain/zona.go` | La zona horaria del sistema | 1 |
| `internal/domain/enums.go` | *Modificar:* sumar `DiaSemana` | 1 |
| `internal/domain/horario.go` | `HorarioSemanal`, su entrada y la validación de la semana | 2 |
| `internal/domain/profesional.go` | *Modificar:* anticipación y horizonte | 2 |
| `internal/domain/bloqueo.go` | `Bloqueo`, su entrada y su validación | 3 |
| `internal/domain/huecos.go` | `Hueco`, `CalculoHuecos` y el cálculo | 4 |
| `internal/repository/horario.go` | La interfaz de horarios | 5 |
| `internal/repository/bloqueo.go` | La interfaz de bloqueos | 5 |
| `internal/repository/memory/horario.go` | Implementación en memoria | 5 |
| `internal/repository/memory/bloqueo.go` | Implementación en memoria | 5 |
| `internal/service/agenda.go` | Los casos de uso de la agenda | 6 |
| `apps/api/api/openapi.yaml` | *Modificar:* las seis rutas nuevas | 7 |
| `internal/handler/dto_agenda.go` | Los DTOs de la agenda | 8 |
| `internal/handler/agenda.go` | Los controllers | 8 |
| `internal/handler/router.go` | *Modificar:* registrar las rutas | 8 |
| `cmd/api/main.go` | *Modificar:* cablear el servicio y embeber tzdata | 8 |

### Modelo sugerido por tarea

Para ahorrar sin bajar calidad: las tareas de transcripción van en el modelo más barato, y el cálculo —lo único con lógica de verdad— en el más capaz.

| Tarea | Modelo | Por qué |
|---|---|---|
| 1, 3, 7 | barato | Transcripción con el código completo en el brief |
| 2, 5, 6, 8, 9 | estándar | Validación cruzada, concurrencia, superficie amplia |
| 4 | el más capaz | Matemática de calendario con bordes que se rompen en silencio |

---

## Task 1: Vocabulario de calendario

Tres piezas puras sin dependencias entre sí, que todo lo demás usa.

**Files:**
- Create: `apps/api/internal/domain/hora.go`
- Create: `apps/api/internal/domain/hora_test.go`
- Create: `apps/api/internal/domain/zona.go`
- Modify: `apps/api/internal/domain/enums.go` (agregar al final)
- Modify: `apps/api/internal/domain/enums_test.go` (agregar al final)

**Interfaces:**
- Consumes: nada
- Produces:
  - `domain.HoraDelDia{Minutos int}`, `ParsearHoraDelDia(string) (HoraDelDia, error)`, `(HoraDelDia) String() string`, `(HoraDelDia) Antes(HoraDelDia) bool`
  - `domain.ZonaHoraria *time.Location`
  - `domain.DiaSemana` con `DiaLunes`…`DiaDomingo`, `(DiaSemana) EsValido() bool`, `(DiaSemana) AWeekday() time.Weekday`, `(DiaSemana) Orden() int`, `DiaSemanaDe(time.Weekday) DiaSemana`

- [ ] **Step 1: Escribir el test de `HoraDelDia`**

Archivo `apps/api/internal/domain/hora_test.go`:

```go
package domain

import "testing"

func TestParsearHoraDelDiaValida(t *testing.T) {
	casos := []struct {
		entrada string
		minutos int
	}{
		{"00:00", 0},
		{"09:00", 540},
		{"13:30", 810},
		{"23:59", 1439},
		{" 09:00 ", 540},
	}

	for _, caso := range casos {
		t.Run(caso.entrada, func(t *testing.T) {
			h, err := ParsearHoraDelDia(caso.entrada)
			if err != nil {
				t.Fatalf("ParsearHoraDelDia(%q) devolvió error: %v", caso.entrada, err)
			}
			if h.Minutos != caso.minutos {
				t.Errorf("minutos = %d, se esperaba %d", h.Minutos, caso.minutos)
			}
		})
	}
}

func TestParsearHoraDelDiaInvalida(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
	}{
		{"vacia", ""},
		{"sin separador", "0900"},
		{"una sola cifra en la hora", "9:00"},
		{"una sola cifra en los minutos", "09:0"},
		{"hora fuera de rango", "24:00"},
		{"minutos fuera de rango", "09:60"},
		{"con segundos", "09:00:00"},
		{"letras", "ab:cd"},
		{"negativa", "-1:00"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if _, err := ParsearHoraDelDia(caso.entrada); err == nil {
				t.Errorf("ParsearHoraDelDia(%q) debía fallar y no falló", caso.entrada)
			}
		})
	}
}

func TestHoraDelDiaString(t *testing.T) {
	casos := []struct {
		minutos int
		esperado string
	}{
		{0, "00:00"},
		{540, "09:00"},
		{810, "13:30"},
		{1439, "23:59"},
	}

	for _, caso := range casos {
		t.Run(caso.esperado, func(t *testing.T) {
			if obtenido := (HoraDelDia{Minutos: caso.minutos}).String(); obtenido != caso.esperado {
				t.Errorf("String() = %q, se esperaba %q", obtenido, caso.esperado)
			}
		})
	}
}

func TestHoraDelDiaIdaYVuelta(t *testing.T) {
	// lo que sale de String() tiene que volver a entrar por el parser
	for minutos := 0; minutos < 24*60; minutos++ {
		original := HoraDelDia{Minutos: minutos}
		vuelta, err := ParsearHoraDelDia(original.String())
		if err != nil {
			t.Fatalf("no se pudo reparsear %q: %v", original, err)
		}
		if vuelta != original {
			t.Fatalf("ida y vuelta de %q dio %q", original, vuelta)
		}
	}
}

func TestHoraDelDiaAntes(t *testing.T) {
	nueve := HoraDelDia{Minutos: 540}
	trece := HoraDelDia{Minutos: 780}

	if !nueve.Antes(trece) {
		t.Error("09:00 tenía que ser antes de 13:00")
	}
	if trece.Antes(nueve) {
		t.Error("13:00 no tenía que ser antes de 09:00")
	}
	if nueve.Antes(nueve) {
		t.Error("una hora no es anterior a sí misma")
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run TestParsearHoraDelDia -v`
Expected: FAIL con `undefined: ParsearHoraDelDia`

- [ ] **Step 3: Implementar `HoraDelDia`**

Archivo `apps/api/internal/domain/hora.go`:

```go
package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// HoraDelDia es una hora de reloj sin fecha: "las nueve de la mañana".
//
// time.Time no sirve para esto porque siempre carga una fecha, y guardar una
// fecha arbitraria para después ignorarla es la clase de convención que alguien
// termina rompiendo. El horario de un profesional se repite todas las semanas:
// no es un instante, es una hora que vuelve.
type HoraDelDia struct {
	Minutos int // desde medianoche. 09:00 son 540.
}

// ParsearHoraDelDia acepta "HH:MM" en formato de 24 horas, con las dos cifras.
//
// Es estricto a propósito: es lo que manda un <input type="time"> del
// navegador, y aceptar "9:5" obligaría a decidir si son las 9:05 o las 9:50.
func ParsearHoraDelDia(s string) (HoraDelDia, error) {
	partes := strings.Split(strings.TrimSpace(s), ":")
	if len(partes) != 2 || len(partes[0]) != 2 || len(partes[1]) != 2 {
		return HoraDelDia{}, errors.New("el formato es HH:MM")
	}

	horas, err := strconv.Atoi(partes[0])
	if err != nil || horas < 0 || horas > 23 {
		return HoraDelDia{}, errors.New("la hora tiene que estar entre 00 y 23")
	}

	minutos, err := strconv.Atoi(partes[1])
	if err != nil || minutos < 0 || minutos > 59 {
		return HoraDelDia{}, errors.New("los minutos tienen que estar entre 00 y 59")
	}

	return HoraDelDia{Minutos: horas*60 + minutos}, nil
}

func (h HoraDelDia) String() string {
	return fmt.Sprintf("%02d:%02d", h.Minutos/60, h.Minutos%60)
}

// Antes ordena dos horas del mismo día.
func (h HoraDelDia) Antes(otra HoraDelDia) bool {
	return h.Minutos < otra.Minutos
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestParsearHoraDelDia|TestHoraDelDia' -v`
Expected: PASS. `TestHoraDelDiaIdaYVuelta` recorre los 1440 minutos del día.

- [ ] **Step 5: Escribir el test de `DiaSemana`**

Agregar al final de `apps/api/internal/domain/enums_test.go`:

```go
func TestDiaSemanaEsValido(t *testing.T) {
	validos := []DiaSemana{DiaLunes, DiaMartes, DiaMiercoles, DiaJueves, DiaViernes, DiaSabado, DiaDomingo}
	for _, d := range validos {
		if !d.EsValido() {
			t.Errorf("DiaSemana(%q) debía ser válido", d)
		}
	}

	invalidos := []DiaSemana{"", "Lunes", "LUNES", "miércoles", "monday"}
	for _, d := range invalidos {
		if d.EsValido() {
			t.Errorf("DiaSemana(%q) no debía ser válido", d)
		}
	}
}

func TestDiaSemanaIdaYVuelta(t *testing.T) {
	// la conversión contra time.Weekday tiene que cerrar en los dos sentidos,
	// o el cálculo de huecos va a mirar el día equivocado
	casos := map[DiaSemana]time.Weekday{
		DiaDomingo:   time.Sunday,
		DiaLunes:     time.Monday,
		DiaMartes:    time.Tuesday,
		DiaMiercoles: time.Wednesday,
		DiaJueves:    time.Thursday,
		DiaViernes:   time.Friday,
		DiaSabado:    time.Saturday,
	}

	for dia, weekday := range casos {
		if obtenido := dia.AWeekday(); obtenido != weekday {
			t.Errorf("%q.AWeekday() = %v, se esperaba %v", dia, obtenido, weekday)
		}
		if obtenido := DiaSemanaDe(weekday); obtenido != dia {
			t.Errorf("DiaSemanaDe(%v) = %q, se esperaba %q", weekday, obtenido, dia)
		}
	}
}

func TestDiaSemanaOrdenArrancaEnLunes(t *testing.T) {
	// time.Weekday pone el domingo en cero, que no es cómo se lee una agenda
	// en Argentina
	esperado := []DiaSemana{DiaLunes, DiaMartes, DiaMiercoles, DiaJueves, DiaViernes, DiaSabado, DiaDomingo}
	for i, dia := range esperado {
		if dia.Orden() != i {
			t.Errorf("%q.Orden() = %d, se esperaba %d", dia, dia.Orden(), i)
		}
	}
}
```

Nota: agregar `"time"` a los imports de `enums_test.go`.

- [ ] **Step 6: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run TestDiaSemana -v`
Expected: FAIL con `undefined: DiaSemana`

- [ ] **Step 7: Implementar `DiaSemana`**

Agregar al final de `apps/api/internal/domain/enums.go`, y sumar `"time"` a sus imports:

```go
// DiaSemana usa el mismo vocabulario que el resto de los enums: español y sin
// acentos.
//
// Existe en vez de usar time.Weekday directo para mantener la frontera del
// dominio y para no exponer una numeración donde el domingo vale cero, que es
// una convención heredada de C que nadie recuerda de memoria.
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

var diasPorWeekday = map[time.Weekday]DiaSemana{
	time.Sunday:    DiaDomingo,
	time.Monday:    DiaLunes,
	time.Tuesday:   DiaMartes,
	time.Wednesday: DiaMiercoles,
	time.Thursday:  DiaJueves,
	time.Friday:    DiaViernes,
	time.Saturday:  DiaSabado,
}

// El mapa inverso se deriva del primero en vez de escribirse a mano: dos
// literales pueden desincronizarse y nadie lo nota hasta que la agenda muestra
// el día equivocado.
var weekdaysPorDia = func() map[DiaSemana]time.Weekday {
	m := make(map[DiaSemana]time.Weekday, len(diasPorWeekday))
	for weekday, dia := range diasPorWeekday {
		m[dia] = weekday
	}
	return m
}()

func (d DiaSemana) EsValido() bool {
	_, existe := weekdaysPorDia[d]
	return existe
}

func (d DiaSemana) AWeekday() time.Weekday {
	return weekdaysPorDia[d]
}

// Orden pone el lunes en cero, que es como se lee una agenda acá. time.Weekday
// arranca en domingo y ordenar por ese número dejaría el domingo primero.
func (d DiaSemana) Orden() int {
	return (int(d.AWeekday()) + 6) % 7
}

func DiaSemanaDe(w time.Weekday) DiaSemana {
	return diasPorWeekday[w]
}
```

- [ ] **Step 8: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -v`
Expected: PASS en todo el paquete, incluidos los tests que ya existían.

- [ ] **Step 9: Escribir la zona horaria**

Archivo `apps/api/internal/domain/zona.go`:

```go
package domain

import (
	"time"

	// La imagen del contenedor es distroless y no trae /usr/share/zoneinfo,
	// así que sin esto LoadLocation falla y el servidor no arranca en
	// producción — aunque funcione perfecto en la máquina de desarrollo y en
	// el CI, que sí tienen la base de zonas del sistema.
	//
	// Se importa acá y no en main porque ZonaHoraria se inicializa al cargar
	// este paquete, y el orden de init entre paquetes hermanos no está
	// garantizado: importándolo donde se usa, la dependencia queda explícita.
	_ "time/tzdata"
)

// ZonaHoraria es la zona en la que se interpretan los horarios de la agenda.
//
// ponytail: zona fija. El producto es Argentina, que tiene un huso único. El
// día que haya profesionales fuera del país esto pasa a ser un campo de
// Profesional, y el cambio queda local porque el resto del modelo ya trata las
// horas como hora de reloj.
//
// Que la zona sea constante no es lo mismo que guardar UTC: las horas siguen
// siendo de reloj, lo único fijo es a qué reloj se refieren.
var ZonaHoraria = cargarZonaHoraria()

// InicioDelDia devuelve la medianoche del día al que pertenece ese instante,
// en la zona del sistema.
//
// Vive acá y no en el cálculo porque la usan los dos: el dominio para recorrer
// los días del rango y el servicio para normalizar las fechas de la consulta.
// Definirla dos veces es la clase de duplicación que se desincroniza justo
// cuando alguien toca la zona horaria.
func InicioDelDia(t time.Time) time.Time {
	local := t.In(ZonaHoraria)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, ZonaHoraria)
}

func cargarZonaHoraria() *time.Location {
	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		// Sin zona horaria la agenda no significa nada, así que es mejor no
		// arrancar que arrancar calculando mal.
		panic("no se pudo cargar la zona horaria: " + err.Error())
	}
	return loc
}
```

- [ ] **Step 10: Verificar que la zona carga**

Run: `cd apps/api && go run ./cmd/api 2>&1 | head -3`
Expected: el servidor arranca normalmente. Si la zona no cargara, entraría en pánico en el arranque. Cortar con Ctrl+C.

Run: `cd apps/api && go test ./internal/domain/ -count=1`
Expected: `ok`.

- [ ] **Step 11: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): vocabulario de calendario

HoraDelDia es una hora de reloj sin fecha: el horario de un profesional
se repite todas las semanas, no es un instante. Guardarlo como time.Time
obligaría a inventar una fecha y después acordarse de ignorarla.

DiaSemana existe en vez de usar time.Weekday para no exponer una
numeración donde el domingo vale cero, y Orden() arranca en lunes
porque es como se lee una agenda acá.

La zona horaria importa time/tzdata: la imagen distroless no trae la
base de zonas del sistema, así que sin eso el servidor arranca bien en
desarrollo y muere en producción.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 2: `HorarioSemanal` y la configuración de agenda del profesional

**Files:**
- Create: `apps/api/internal/domain/horario.go`
- Create: `apps/api/internal/domain/horario_test.go`
- Modify: `apps/api/internal/domain/profesional.go`
- Modify: `apps/api/internal/domain/profesional_test.go`

**Interfaces:**
- Consumes: `HoraDelDia`, `DiaSemana`, `Modalidad`, `ErrorValidacion`, `ErrorCampo` (Tasks 1 y anteriores)
- Produces:
  - `domain.HorarioSemanal{ProfesionalID uuid.UUID; DiaSemana DiaSemana; Desde, Hasta HoraDelDia; DuracionMin int; Modalidad Modalidad}`
  - `domain.EntradaHorarioSemanal{DiaSemana, Desde, Hasta string; DuracionMin int; Modalidad string}`
  - `func domain.NuevaSemana(profesionalID uuid.UUID, entradas []EntradaHorarioSemanal) ([]HorarioSemanal, error)`
  - En `Profesional`: los campos `AnticipacionMinimaMin int` y `HorizonteDias int`
  - En `EntradaProfesional`: `AnticipacionMinimaMin *int` y `HorizonteDias *int`

- [ ] **Step 1: Escribir el test de la semana**

Archivo `apps/api/internal/domain/horario_test.go`:

```go
package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func entradaHorarioValida() EntradaHorarioSemanal {
	return EntradaHorarioSemanal{
		DiaSemana:   "lunes",
		Desde:       "09:00",
		Hasta:       "13:00",
		DuracionMin: 50,
		Modalidad:   "telemedicina",
	}
}

func TestNuevaSemanaValida(t *testing.T) {
	profesionalID := uuid.New()
	entradas := []EntradaHorarioSemanal{
		{DiaSemana: "miercoles", Desde: "15:00", Hasta: "19:00", DuracionMin: 60, Modalidad: "presencial"},
		{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
	}

	semana, err := NuevaSemana(profesionalID, entradas)
	if err != nil {
		t.Fatalf("NuevaSemana devolvió error: %v", err)
	}
	if len(semana) != 2 {
		t.Fatalf("len(semana) = %d, se esperaba 2", len(semana))
	}

	// se devuelve ordenada, y el lunes va primero aunque haya llegado segundo
	if semana[0].DiaSemana != DiaLunes {
		t.Errorf("el primer bloque es %q, se esperaba lunes", semana[0].DiaSemana)
	}
	if semana[0].ProfesionalID != profesionalID {
		t.Error("el bloque no quedó atado a su profesional")
	}
	if semana[0].Desde.String() != "09:00" || semana[0].Hasta.String() != "13:00" {
		t.Errorf("horas = %q a %q", semana[0].Desde, semana[0].Hasta)
	}
}

func TestNuevaSemanaVacia(t *testing.T) {
	// vaciar la agenda es legítimo: un profesional puede dejar de atender
	semana, err := NuevaSemana(uuid.New(), nil)
	if err != nil {
		t.Fatalf("una semana vacía debía ser válida, devolvió: %v", err)
	}
	if len(semana) != 0 {
		t.Errorf("len(semana) = %d, se esperaba 0", len(semana))
	}
}

func TestNuevaSemanaCamposInvalidos(t *testing.T) {
	casos := []struct {
		nombre       string
		mutar        func(*EntradaHorarioSemanal)
		campoEsperado string
	}{
		{"dia desconocido", func(e *EntradaHorarioSemanal) { e.DiaSemana = "lunez" }, "horarios[0].diaSemana"},
		{"desde mal formado", func(e *EntradaHorarioSemanal) { e.Desde = "9am" }, "horarios[0].desde"},
		{"hasta mal formado", func(e *EntradaHorarioSemanal) { e.Hasta = "25:00" }, "horarios[0].hasta"},
		{"hasta antes que desde", func(e *EntradaHorarioSemanal) { e.Hasta = "08:00" }, "horarios[0].hasta"},
		{"hasta igual a desde", func(e *EntradaHorarioSemanal) { e.Hasta = "09:00" }, "horarios[0].hasta"},
		{"duracion muy corta", func(e *EntradaHorarioSemanal) { e.DuracionMin = 5 }, "horarios[0].duracionMin"},
		{"duracion muy larga", func(e *EntradaHorarioSemanal) { e.DuracionMin = 500 }, "horarios[0].duracionMin"},
		{"modalidad desconocida", func(e *EntradaHorarioSemanal) { e.Modalidad = "carta" }, "horarios[0].modalidad"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := entradaHorarioValida()
			caso.mutar(&entrada)

			_, err := NuevaSemana(uuid.New(), []EntradaHorarioSemanal{entrada})
			if err == nil {
				t.Fatal("se esperaba un error de validación")
			}

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
			}

			encontrado := false
			for _, campo := range verr.Campos {
				if campo.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestNuevaSemanaBloqueSinHuecos(t *testing.T) {
	// "lunes de 9:00 a 9:30" con sesiones de 50 minutos no genera un solo
	// turno. El profesional tiene que enterarse al cargarlo y no dos semanas
	// después, al no recibir nada.
	entrada := entradaHorarioValida()
	entrada.Hasta = "09:30"

	_, err := NuevaSemana(uuid.New(), []EntradaHorarioSemanal{entrada})

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
	encontrado := false
	for _, campo := range verr.Campos {
		if campo.Campo == "horarios[0].duracionMin" {
			encontrado = true
		}
	}
	if !encontrado {
		t.Errorf("se esperaba un error en duracionMin, se obtuvo %+v", verr.Campos)
	}
}

func TestNuevaSemanaBloqueJusto(t *testing.T) {
	// el borde del caso anterior: un bloque donde entra exactamente una sesión
	entrada := entradaHorarioValida()
	entrada.Hasta = "09:50"

	semana, err := NuevaSemana(uuid.New(), []EntradaHorarioSemanal{entrada})
	if err != nil {
		t.Fatalf("un bloque donde entra justo una sesión es válido, devolvió: %v", err)
	}
	if len(semana) != 1 {
		t.Errorf("len(semana) = %d, se esperaba 1", len(semana))
	}
}

func TestNuevaSemanaSolapamiento(t *testing.T) {
	t.Run("dos bloques del mismo dia que se pisan", func(t *testing.T) {
		entradas := []EntradaHorarioSemanal{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
			{DiaSemana: "lunes", Desde: "12:00", Hasta: "15:00", DuracionMin: 50, Modalidad: "telemedicina"},
		}

		_, err := NuevaSemana(uuid.New(), entradas)
		var verr ErrorValidacion
		if !errors.As(err, &verr) {
			t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
		}
	})

	t.Run("dos bloques pegados no se pisan", func(t *testing.T) {
		// los intervalos son semiabiertos: uno que termina 13:00 y otro que
		// empieza 13:00 son contiguos, no solapados
		entradas := []EntradaHorarioSemanal{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
			{DiaSemana: "lunes", Desde: "13:00", Hasta: "17:00", DuracionMin: 50, Modalidad: "telemedicina"},
		}

		if _, err := NuevaSemana(uuid.New(), entradas); err != nil {
			t.Errorf("dos bloques contiguos son válidos, devolvió: %v", err)
		}
	})

	t.Run("mismo horario en dias distintos no se pisa", func(t *testing.T) {
		entradas := []EntradaHorarioSemanal{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
			{DiaSemana: "martes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
		}

		if _, err := NuevaSemana(uuid.New(), entradas); err != nil {
			t.Errorf("el mismo horario en días distintos es válido, devolvió: %v", err)
		}
	})
}

func TestNuevaSemanaAcumulaErroresDeVariosBloques(t *testing.T) {
	entradas := []EntradaHorarioSemanal{
		{DiaSemana: "lunez", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
		{DiaSemana: "martes", Desde: "nueve", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
	}

	_, err := NuevaSemana(uuid.New(), entradas)

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
	}
	// el índice del bloque tiene que estar en el nombre del campo, o el cliente
	// no sabe cuál de los dos corregir
	if len(verr.Campos) != 2 {
		t.Fatalf("se esperaban 2 campos con error, se obtuvieron %d: %+v", len(verr.Campos), verr.Campos)
	}
	if verr.Campos[0].Campo != "horarios[0].diaSemana" || verr.Campos[1].Campo != "horarios[1].desde" {
		t.Errorf("los campos no identifican el bloque: %+v", verr.Campos)
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run TestNuevaSemana -v`
Expected: FAIL con `undefined: EntradaHorarioSemanal` y `undefined: NuevaSemana`

- [ ] **Step 3: Implementar `HorarioSemanal`**

Archivo `apps/api/internal/domain/horario.go`:

```go
package domain

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

const (
	minDuracionMin = 10
	maxDuracionMin = 480
)

// HorarioSemanal es un bloque del horario habitual de un profesional. El que
// atiende mañana y tarde tiene dos bloques para el mismo día.
//
// No tiene ID a propósito: la semana se reemplaza entera, así que ningún
// endpoint direcciona un bloque suelto y nada lo referencia. Un ID sería peso
// muerto que además cambiaría en cada guardado. Un horario es un valor, no una
// entidad con identidad.
type HorarioSemanal struct {
	ProfesionalID uuid.UUID
	DiaSemana     DiaSemana
	Desde         HoraDelDia
	Hasta         HoraDelDia
	DuracionMin   int
	Modalidad     Modalidad
}

type EntradaHorarioSemanal struct {
	DiaSemana   string
	Desde       string
	Hasta       string
	DuracionMin int
	Modalidad   string
}

// NuevaSemana valida el conjunto entero de bloques de un profesional.
//
// Valida el conjunto y no bloque por bloque porque el solapamiento solo se ve
// mirando todos juntos, y porque la semana se guarda de una sola vez: no existe
// un estado donde la mitad de los bloques estén cargados.
//
// Una semana vacía es válida: un profesional puede dejar de atender.
func NuevaSemana(profesionalID uuid.UUID, entradas []EntradaHorarioSemanal) ([]HorarioSemanal, error) {
	var verr ErrorValidacion
	bloques := make([]HorarioSemanal, 0, len(entradas))

	for i, entrada := range entradas {
		bloque, errores := construirBloque(profesionalID, entrada)
		for _, e := range errores {
			// el índice ubica al cliente en cuál de los bloques está el problema
			verr.agregar(fmt.Sprintf("horarios[%d].%s", i, e.Campo), e.Mensaje)
		}
		bloques = append(bloques, bloque)
	}

	if verr.tieneErrores() {
		return nil, verr
	}

	if err := verificarSolapamiento(bloques); err != nil {
		return nil, err
	}

	ordenarSemana(bloques)
	return bloques, nil
}

func construirBloque(profesionalID uuid.UUID, entrada EntradaHorarioSemanal) (HorarioSemanal, []ErrorCampo) {
	var errores []ErrorCampo
	bloque := HorarioSemanal{ProfesionalID: profesionalID}

	dia := DiaSemana(strings.ToLower(strings.TrimSpace(entrada.DiaSemana)))
	if !dia.EsValido() {
		errores = append(errores, ErrorCampo{
			Campo:   "diaSemana",
			Mensaje: "debe ser lunes, martes, miercoles, jueves, viernes, sabado o domingo",
		})
	} else {
		bloque.DiaSemana = dia
	}

	modalidad := Modalidad(strings.ToLower(strings.TrimSpace(entrada.Modalidad)))
	if !modalidad.EsValida() {
		errores = append(errores, ErrorCampo{
			Campo:   "modalidad",
			Mensaje: "debe ser telemedicina, presencial o domicilio",
		})
	} else {
		bloque.Modalidad = modalidad
	}

	desde, errDesde := ParsearHoraDelDia(entrada.Desde)
	if errDesde != nil {
		errores = append(errores, ErrorCampo{Campo: "desde", Mensaje: errDesde.Error()})
	} else {
		bloque.Desde = desde
	}

	hasta, errHasta := ParsearHoraDelDia(entrada.Hasta)
	if errHasta != nil {
		errores = append(errores, ErrorCampo{Campo: "hasta", Mensaje: errHasta.Error()})
	} else {
		bloque.Hasta = hasta
	}

	duracionValida := entrada.DuracionMin >= minDuracionMin && entrada.DuracionMin <= maxDuracionMin
	switch {
	case entrada.DuracionMin < minDuracionMin:
		errores = append(errores, ErrorCampo{
			Campo:   "duracionMin",
			Mensaje: fmt.Sprintf("no puede ser menor a %d minutos", minDuracionMin),
		})
	case entrada.DuracionMin > maxDuracionMin:
		errores = append(errores, ErrorCampo{
			Campo:   "duracionMin",
			Mensaje: fmt.Sprintf("no puede superar los %d minutos", maxDuracionMin),
		})
	default:
		bloque.DuracionMin = entrada.DuracionMin
	}

	// Las dos reglas que siguen comparan campos entre sí, así que solo tienen
	// sentido si los campos individuales parsearon bien.
	if errDesde != nil || errHasta != nil || !duracionValida {
		return bloque, errores
	}

	if !bloque.Desde.Antes(bloque.Hasta) {
		errores = append(errores, ErrorCampo{Campo: "hasta", Mensaje: "tiene que ser posterior a desde"})
		return bloque, errores
	}

	// Un bloque donde no entra ni una sesión no genera un solo turno. Es mejor
	// que el profesional se entere ahora, y no dos semanas después al no
	// recibir nada.
	if bloque.Hasta.Minutos-bloque.Desde.Minutos < bloque.DuracionMin {
		errores = append(errores, ErrorCampo{
			Campo:   "duracionMin",
			Mensaje: "no entra ninguna sesión en ese bloque",
		})
	}

	return bloque, errores
}

func verificarSolapamiento(bloques []HorarioSemanal) error {
	var verr ErrorValidacion

	for i := range bloques {
		for j := i + 1; j < len(bloques); j++ {
			a, b := bloques[i], bloques[j]
			if a.DiaSemana != b.DiaSemana {
				continue
			}
			// intervalos semiabiertos: uno que termina 13:00 y otro que empieza
			// 13:00 son contiguos, no solapados
			if a.Desde.Minutos < b.Hasta.Minutos && b.Desde.Minutos < a.Hasta.Minutos {
				verr.agregar(
					fmt.Sprintf("horarios[%d]", j),
					fmt.Sprintf("se solapa con el bloque %d del mismo día", i),
				)
			}
		}
	}

	if verr.tieneErrores() {
		return verr
	}
	return nil
}

// ordenarSemana deja la semana en un orden estable para que la respuesta no
// dependa de en qué orden la mandó el cliente.
func ordenarSemana(bloques []HorarioSemanal) {
	slices.SortFunc(bloques, func(a, b HorarioSemanal) int {
		if c := a.DiaSemana.Orden() - b.DiaSemana.Orden(); c != 0 {
			return c
		}
		return a.Desde.Minutos - b.Desde.Minutos
	})
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -run TestNuevaSemana -v`
Expected: PASS en todos los subtests.

- [ ] **Step 5: Escribir el test de la configuración de agenda**

Agregar al final de `apps/api/internal/domain/profesional_test.go`:

```go
func TestNuevoProfesionalConfiguracionDeAgendaPorDefecto(t *testing.T) {
	// no mandar los campos es lo normal: el profesional no debería tener que
	// decidir esto al registrarse
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.AnticipacionMinimaMin != anticipacionMinimaPorDefecto {
		t.Errorf("AnticipacionMinimaMin = %d, se esperaba %d", p.AnticipacionMinimaMin, anticipacionMinimaPorDefecto)
	}
	if p.HorizonteDias != horizonteDiasPorDefecto {
		t.Errorf("HorizonteDias = %d, se esperaba %d", p.HorizonteDias, horizonteDiasPorDefecto)
	}
}

func TestNuevoProfesionalConfiguracionDeAgendaExplicita(t *testing.T) {
	entrada := entradaValida()
	anticipacion := 30
	horizonte := 90
	entrada.AnticipacionMinimaMin = &anticipacion
	entrada.HorizonteDias = &horizonte

	p, err := NuevoProfesional(entrada, ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.AnticipacionMinimaMin != 30 {
		t.Errorf("AnticipacionMinimaMin = %d, se esperaba 30", p.AnticipacionMinimaMin)
	}
	if p.HorizonteDias != 90 {
		t.Errorf("HorizonteDias = %d, se esperaba 90", p.HorizonteDias)
	}
}

func TestNuevoProfesionalConfiguracionDeAgendaInvalida(t *testing.T) {
	casos := []struct {
		nombre        string
		mutar         func(*EntradaProfesional)
		campoEsperado string
	}{
		{"anticipacion negativa", func(e *EntradaProfesional) { v := -1; e.AnticipacionMinimaMin = &v }, "anticipacionMinimaMin"},
		{"anticipacion mayor a una semana", func(e *EntradaProfesional) { v := 10081; e.AnticipacionMinimaMin = &v }, "anticipacionMinimaMin"},
		{"horizonte en cero", func(e *EntradaProfesional) { v := 0; e.HorizonteDias = &v }, "horizonteDias"},
		{"horizonte sobre el tope", func(e *EntradaProfesional) { v := 181; e.HorizonteDias = &v }, "horizonteDias"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := entradaValida()
			caso.mutar(&entrada)

			_, err := NuevoProfesional(entrada, ahoraDePrueba)

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
			}
			encontrado := false
			for _, campo := range verr.Campos {
				if campo.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestAplicarCambiosVuelveAlDefaultSiNoMandanLaConfiguracion(t *testing.T) {
	// PUT es reemplazo total: omitir un campo lo devuelve a su valor por
	// defecto, igual que pasa con el resto
	entrada := entradaValida()
	anticipacion := 30
	entrada.AnticipacionMinimaMin = &anticipacion

	base, err := NuevoProfesional(entrada, ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if base.AnticipacionMinimaMin != 30 {
		t.Fatalf("no se pudo preparar el estado: %d", base.AnticipacionMinimaMin)
	}

	actualizado, err := base.AplicarCambios(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if actualizado.AnticipacionMinimaMin != anticipacionMinimaPorDefecto {
		t.Errorf("AnticipacionMinimaMin = %d, se esperaba el default %d",
			actualizado.AnticipacionMinimaMin, anticipacionMinimaPorDefecto)
	}
}
```

- [ ] **Step 6: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestNuevoProfesionalConfiguracion|TestAplicarCambiosVuelve' -v`
Expected: FAIL con `undefined: anticipacionMinimaPorDefecto` y campos inexistentes.

- [ ] **Step 7: Agregar los campos a `Profesional`**

En `apps/api/internal/domain/profesional.go`:

1. Agregar las constantes junto a las que ya existen (`maxLargoNombre` y compañía):

```go
const (
	anticipacionMinimaPorDefecto = 120       // dos horas
	maxAnticipacionMinimaMin     = 7 * 24 * 60 // una semana
	horizonteDiasPorDefecto      = 60
	maxHorizonteDias             = 180
)
```

2. Agregar los dos campos al struct `Profesional`, después de `ObrasSociales`:

```go
	// Configuración de la agenda. Vive acá y no en una entidad aparte por la
	// misma razón que PrecioConsulta: no es identidad, es cómo esa persona
	// ejerce, y una entidad para dos campos es ceremonia.
	AnticipacionMinimaMin int
	HorizonteDias         int
```

3. Agregar los dos campos a `EntradaProfesional`, después de `ObrasSociales`:

```go
	// Punteros para distinguir "no lo mandaron" de "mandaron cero". A
	// diferencia de PrecioConsulta, acá la ausencia no es un error: significa
	// usar el valor por defecto, porque nadie debería tener que decidir esto
	// al registrarse.
	AnticipacionMinimaMin *int
	HorizonteDias         *int
```

4. Agregar al final de `construir`, justo antes del `return`:

```go
	p.AnticipacionMinimaMin = anticipacionMinimaPorDefecto
	if entrada.AnticipacionMinimaMin != nil {
		switch valor := *entrada.AnticipacionMinimaMin; {
		case valor < 0:
			verr.agregar("anticipacionMinimaMin", "no puede ser negativa")
		case valor > maxAnticipacionMinimaMin:
			verr.agregar("anticipacionMinimaMin", "no puede superar una semana")
		default:
			p.AnticipacionMinimaMin = valor
		}
	}

	p.HorizonteDias = horizonteDiasPorDefecto
	if entrada.HorizonteDias != nil {
		switch valor := *entrada.HorizonteDias; {
		case valor < 1:
			verr.agregar("horizonteDias", "tiene que ser al menos 1")
		case valor > maxHorizonteDias:
			verr.agregar("horizonteDias", fmt.Sprintf("no puede superar los %d días", maxHorizonteDias))
		default:
			p.HorizonteDias = valor
		}
	}
```

- [ ] **Step 8: Correr toda la suite del dominio**

Run: `cd apps/api && go test ./internal/domain/ -count=1 -v 2>&1 | tail -20`
Expected: PASS en todo, incluidos los tests de `Profesional` que ya existían.

Run: `cd apps/api && go test ./... -count=1`
Expected: `ok` en todos los paquetes. Los tests del servicio y del handler siguen pasando porque los dos campos nuevos tienen valor por defecto.

- [ ] **Step 9: Verificar que el dominio sigue aislado**

Run: `cd apps/api && go list -deps ./internal/domain | grep "^$(go list -m)"`
Expected: exactamente una línea, el propio paquete `domain`.

- [ ] **Step 10: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/
git commit -m "feat(domain): horario semanal y configuración de agenda

NuevaSemana valida el conjunto entero y no bloque por bloque, porque el
solapamiento solo se ve mirando todos juntos y porque la semana se
guarda de una sola vez.

Un bloque donde no entra ninguna sesión es un error de validación: el
profesional se entera al cargarlo y no dos semanas después, al no
recibir ningún turno.

La anticipación mínima y el horizonte van en Profesional por la misma
razón que PrecioConsulta: no son identidad, son cómo esa persona ejerce.
Ambos con valor por defecto, para que nadie tenga que decidirlos al
registrarse.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 3: `Bloqueo`

**Files:**
- Create: `apps/api/internal/domain/bloqueo.go`
- Create: `apps/api/internal/domain/bloqueo_test.go`

**Interfaces:**
- Consumes: `ErrorValidacion`, `ZonaHoraria` (Tasks 1 y anteriores)
- Produces:
  - `domain.Bloqueo{ID, ProfesionalID uuid.UUID; Desde, Hasta time.Time; Motivo string; CreadoEn time.Time}`
  - `domain.EntradaBloqueo{Desde, Hasta time.Time; Motivo string}`
  - `func domain.NuevoBloqueo(profesionalID uuid.UUID, entrada EntradaBloqueo, ahora time.Time) (Bloqueo, error)`
  - `func (Bloqueo) SeSolapaCon(inicio, fin time.Time) bool`

- [ ] **Step 1: Escribir el test**

Archivo `apps/api/internal/domain/bloqueo_test.go`:

```go
package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func momento(t *testing.T, s string) time.Time {
	t.Helper()
	instante, err := time.ParseInLocation("2006-01-02 15:04", s, ZonaHoraria)
	if err != nil {
		t.Fatalf("no se pudo parsear %q: %v", s, err)
	}
	return instante
}

func TestNuevoBloqueoValido(t *testing.T) {
	ahora := momento(t, "2026-08-22 10:00")
	profesionalID := uuid.New()

	b, err := NuevoBloqueo(profesionalID, EntradaBloqueo{
		Desde:  momento(t, "2026-09-10 00:00"),
		Hasta:  momento(t, "2026-09-20 23:59"),
		Motivo: "  Vacaciones  ",
	}, ahora)
	if err != nil {
		t.Fatalf("NuevoBloqueo devolvió error: %v", err)
	}

	if b.ProfesionalID != profesionalID {
		t.Error("el bloqueo no quedó atado a su profesional")
	}
	if b.Motivo != "Vacaciones" {
		t.Errorf("Motivo = %q, se esperaba sin espacios alrededor", b.Motivo)
	}
	if !b.CreadoEn.Equal(ahora) {
		t.Error("CreadoEn tenía que ser el ahora recibido")
	}
	if b.ID == uuid.Nil {
		t.Error("el ID tenía que generarse")
	}
}

func TestNuevoBloqueoSinMotivo(t *testing.T) {
	// el motivo es opcional: bloquear sin explicar por qué es legítimo
	_, err := NuevoBloqueo(uuid.New(), EntradaBloqueo{
		Desde: momento(t, "2026-09-10 00:00"),
		Hasta: momento(t, "2026-09-11 00:00"),
	}, momento(t, "2026-08-22 10:00"))
	if err != nil {
		t.Errorf("un bloqueo sin motivo es válido, devolvió: %v", err)
	}
}

func TestNuevoBloqueoInvalido(t *testing.T) {
	ahora := momento(t, "2026-08-22 10:00")

	casos := []struct {
		nombre        string
		entrada       EntradaBloqueo
		campoEsperado string
	}{
		{
			"desde vacio",
			EntradaBloqueo{Hasta: momento(t, "2026-09-11 00:00")},
			"desde",
		},
		{
			"hasta vacio",
			EntradaBloqueo{Desde: momento(t, "2026-09-10 00:00")},
			"hasta",
		},
		{
			"hasta antes que desde",
			EntradaBloqueo{Desde: momento(t, "2026-09-20 00:00"), Hasta: momento(t, "2026-09-10 00:00")},
			"hasta",
		},
		{
			"desde igual a hasta",
			EntradaBloqueo{Desde: momento(t, "2026-09-10 00:00"), Hasta: momento(t, "2026-09-10 00:00")},
			"hasta",
		},
		{
			"enteramente en el pasado",
			EntradaBloqueo{Desde: momento(t, "2026-07-01 00:00"), Hasta: momento(t, "2026-07-10 00:00")},
			"hasta",
		},
		{
			"motivo demasiado largo",
			EntradaBloqueo{
				Desde:  momento(t, "2026-09-10 00:00"),
				Hasta:  momento(t, "2026-09-11 00:00"),
				Motivo: strings.Repeat("a", 201),
			},
			"motivo",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := NuevoBloqueo(uuid.New(), caso.entrada, ahora)

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
			}
			encontrado := false
			for _, campo := range verr.Campos {
				if campo.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestNuevoBloqueoQueEmpezoPeroNoTermino(t *testing.T) {
	// unas vacaciones que arrancaron ayer y siguen hasta la semana que viene
	// son perfectamente válidas: lo que no sirve es bloquear algo ya terminado
	ahora := momento(t, "2026-08-22 10:00")

	_, err := NuevoBloqueo(uuid.New(), EntradaBloqueo{
		Desde: momento(t, "2026-08-20 00:00"),
		Hasta: momento(t, "2026-08-30 00:00"),
	}, ahora)
	if err != nil {
		t.Errorf("un bloqueo en curso es válido, devolvió: %v", err)
	}
}

func TestBloqueoSeSolapaCon(t *testing.T) {
	b := Bloqueo{
		Desde: momento(t, "2026-08-24 09:50"),
		Hasta: momento(t, "2026-08-24 10:00"),
	}

	casos := []struct {
		nombre   string
		inicio   string
		fin      string
		esperado bool
	}{
		{"adentro", "2026-08-24 09:52", "2026-08-24 09:55", true},
		{"lo contiene", "2026-08-24 09:00", "2026-08-24 11:00", true},
		{"pisa el arranque", "2026-08-24 09:45", "2026-08-24 09:55", true},
		{"pisa el final", "2026-08-24 09:55", "2026-08-24 10:05", true},
		// los dos bordes: intervalos semiabiertos [inicio, fin)
		{"termina justo cuando el bloqueo empieza", "2026-08-24 09:00", "2026-08-24 09:50", false},
		{"empieza justo cuando el bloqueo termina", "2026-08-24 10:00", "2026-08-24 10:50", false},
		{"muy antes", "2026-08-24 07:00", "2026-08-24 08:00", false},
		{"muy despues", "2026-08-24 15:00", "2026-08-24 16:00", false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			obtenido := b.SeSolapaCon(momento(t, caso.inicio), momento(t, caso.fin))
			if obtenido != caso.esperado {
				t.Errorf("SeSolapaCon(%s, %s) = %v, se esperaba %v", caso.inicio, caso.fin, obtenido, caso.esperado)
			}
		})
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run 'TestNuevoBloqueo|TestBloqueo' -v`
Expected: FAIL con `undefined: EntradaBloqueo` y `undefined: NuevoBloqueo`

- [ ] **Step 3: Implementar `Bloqueo`**

Archivo `apps/api/internal/domain/bloqueo.go`:

```go
package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxLargoMotivo = 200

// Bloqueo es lo que rompe el horario habitual: vacaciones, un feriado, cerrar
// temprano un martes.
//
// Guarda fechas y horas locales, no instantes UTC, por la misma razón que
// HorarioSemanal guarda hora de reloj: "me voy del 10 al 20" es una fecha del
// calendario del profesional.
//
// A diferencia de HorarioSemanal sí tiene ID, porque se borra de a uno.
type Bloqueo struct {
	ID            uuid.UUID
	ProfesionalID uuid.UUID
	Desde         time.Time
	Hasta         time.Time
	Motivo        string
	CreadoEn      time.Time
}

type EntradaBloqueo struct {
	Desde  time.Time
	Hasta  time.Time
	Motivo string
}

// NuevoBloqueo valida la entrada y devuelve un bloqueo consistente.
//
// El ahora llega por parámetro y no se lee del reloj del sistema para que la
// regla de "no bloquear el pasado" sea determinista en los tests.
func NuevoBloqueo(profesionalID uuid.UUID, entrada EntradaBloqueo, ahora time.Time) (Bloqueo, error) {
	var verr ErrorValidacion

	if entrada.Desde.IsZero() {
		verr.agregar("desde", "es obligatorio")
	}
	if entrada.Hasta.IsZero() {
		verr.agregar("hasta", "es obligatorio")
	}

	if !entrada.Desde.IsZero() && !entrada.Hasta.IsZero() {
		if !entrada.Desde.Before(entrada.Hasta) {
			verr.agregar("hasta", "tiene que ser posterior a desde")
		} else if !entrada.Hasta.After(ahora) {
			// Un bloqueo ya terminado no cambia nada: los huecos que tapaba
			// pasaron y de todos modos no se pueden reservar. Uno que empezó y
			// todavía no terminó sí es válido.
			verr.agregar("hasta", "tiene que estar en el futuro")
		}
	}

	motivo := strings.TrimSpace(entrada.Motivo)
	if utf8.RuneCountInString(motivo) > maxLargoMotivo {
		verr.agregar("motivo", fmt.Sprintf("no puede superar los %d caracteres", maxLargoMotivo))
	}

	if verr.tieneErrores() {
		return Bloqueo{}, verr
	}

	return Bloqueo{
		ID:            uuid.New(),
		ProfesionalID: profesionalID,
		Desde:         entrada.Desde,
		Hasta:         entrada.Hasta,
		Motivo:        motivo,
		CreadoEn:      ahora,
	}, nil
}

// SeSolapaCon dice si el bloqueo pisa el intervalo [inicio, fin).
//
// Los intervalos son semiabiertos en todo el sistema: un bloqueo que empieza
// 09:50 no pisa un hueco que termina 09:50. Es el off-by-one clásico de las
// agendas, y usar <= en cualquiera de las dos comparaciones haría desaparecer
// un turno por día sin que nadie entienda por qué.
func (b Bloqueo) SeSolapaCon(inicio, fin time.Time) bool {
	return inicio.Before(b.Hasta) && fin.After(b.Desde)
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -count=1`
Expected: `ok`. Los ocho subtests de `TestBloqueoSeSolapaCon` son los que importan, en particular los dos bordes.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/bloqueo.go apps/api/internal/domain/bloqueo_test.go
git commit -m "feat(domain): bloqueos de agenda

Un bloqueo tapa parte del horario habitual: vacaciones, un feriado,
cerrar temprano. Guarda fechas locales por la misma razón que el horario
guarda hora de reloj.

SeSolapaCon usa intervalos semiabiertos, y sus dos casos borde tienen
test propio: usar <= en cualquiera de las comparaciones haría
desaparecer un turno por día sin que nadie entienda por qué.

Bloquear algo ya terminado es un error de validación; un bloqueo en
curso, que empezó ayer y sigue, es válido.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 4: El cálculo de huecos

**La tarea con la lógica de verdad.** Todo lo demás es transporte alrededor de esto.

**Files:**
- Create: `apps/api/internal/domain/huecos.go`
- Create: `apps/api/internal/domain/huecos_test.go`

**Interfaces:**
- Consumes: `HorarioSemanal`, `Bloqueo`, `DiaSemanaDe`, `ZonaHoraria`, `Modalidad` (Tasks 1-3)
- Produces:
  - `domain.Hueco{Inicio, Fin time.Time; Modalidad Modalidad}`
  - `domain.CalculoHuecos{Horarios []HorarioSemanal; Bloqueos []Bloqueo; Desde, Hasta time.Time; AnticipacionMinimaMin int; Ahora time.Time}`
  - `func (CalculoHuecos) Generar() []Hueco`

- [ ] **Step 1: Escribir el test**

Archivo `apps/api/internal/domain/huecos_test.go`:

```go
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// 2026-08-24 es lunes. Todas las fechas de este archivo se apoyan en eso.
const (
	lunes     = "2026-08-24"
	martes    = "2026-08-25"
	lunesQueViene = "2026-08-31"
)

func bloqueLunes() HorarioSemanal {
	return HorarioSemanal{
		ProfesionalID: uuid.New(),
		DiaSemana:     DiaLunes,
		Desde:         HoraDelDia{Minutos: 9 * 60},
		Hasta:         HoraDelDia{Minutos: 13 * 60},
		DuracionMin:   50,
		Modalidad:     ModalidadTelemedicina,
	}
}

// rango arma el intervalo semiabierto que espera CalculoHuecos a partir de dos
// fechas inclusivas, igual que hace el servicio.
func rango(t *testing.T, desdeFecha, hastaFecha string) (time.Time, time.Time) {
	t.Helper()
	desde := momento(t, desdeFecha+" 00:00")
	hasta := momento(t, hastaFecha+" 00:00").AddDate(0, 0, 1)
	return desde, hasta
}

func inicios(huecos []Hueco) []string {
	salida := make([]string, 0, len(huecos))
	for _, h := range huecos {
		salida = append(salida, h.Inicio.Format("2006-01-02 15:04"))
	}
	return salida
}

func compararInicios(t *testing.T, huecos []Hueco, esperado []string) {
	t.Helper()
	obtenido := inicios(huecos)
	if len(obtenido) != len(esperado) {
		t.Fatalf("se obtuvieron %d huecos %v, se esperaban %d %v", len(obtenido), obtenido, len(esperado), esperado)
	}
	for i := range esperado {
		if obtenido[i] != esperado[i] {
			t.Errorf("hueco %d = %q, se esperaba %q", i, obtenido[i], esperado[i])
		}
	}
}

func TestGenerarUnBloque(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	c := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}

	// de 09:00 a 13:00 con sesiones de 50 minutos entran cuatro: la quinta
	// arrancaría 12:20 y terminaría 13:10, pasada del final del bloque
	compararInicios(t, c.Generar(), []string{
		lunes + " 09:00",
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarLlevaLaModalidadDelBloque(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)
	bloque := bloqueLunes()
	bloque.Modalidad = ModalidadDomicilio

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloque},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) == 0 {
		t.Fatal("no se generó ningún hueco")
	}
	for _, h := range huecos {
		if h.Modalidad != ModalidadDomicilio {
			t.Errorf("modalidad = %q, se esperaba domicilio", h.Modalidad)
		}
	}
}

func TestGenerarBloqueDondeNoEntraNinguno(t *testing.T) {
	// la validación rechaza esto al cargarlo, pero el cálculo tampoco tiene
	// que inventar un hueco que se pasa del bloque
	desde, hasta := rango(t, lunes, lunes)
	bloque := bloqueLunes()
	bloque.Hasta = HoraDelDia{Minutos: 9*60 + 30}

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloque},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("se generaron %d huecos %v, se esperaban 0", len(huecos), inicios(huecos))
	}
}

func TestGenerarDosBloquesElMismoDia(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	tarde := bloqueLunes()
	tarde.Desde = HoraDelDia{Minutos: 15 * 60}
	tarde.Hasta = HoraDelDia{Minutos: 17 * 60}
	tarde.DuracionMin = 60

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{tarde, bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	// salen ordenados aunque los bloques hayan llegado al revés
	compararInicios(t, huecos, []string{
		lunes + " 09:00",
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
		lunes + " 15:00",
		lunes + " 16:00",
	})
}

func TestGenerarIgnoraLosBloquesDeOtroDia(t *testing.T) {
	desde, hasta := rango(t, martes, martes)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("un bloque de lunes no debía producir nada un martes: %v", inicios(huecos))
	}
}

func TestGenerarRepiteElBloqueCadaSemana(t *testing.T) {
	desde, hasta := rango(t, lunes, lunesQueViene)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	// dos lunes en el rango, cuatro huecos cada uno
	if len(huecos) != 8 {
		t.Fatalf("se obtuvieron %d huecos %v, se esperaban 8", len(huecos), inicios(huecos))
	}
	if inicios(huecos)[0] != lunes+" 09:00" || inicios(huecos)[4] != lunesQueViene+" 09:00" {
		t.Errorf("el bloque no se repitió como corresponde: %v", inicios(huecos))
	}
}

func TestGenerarBloqueoQueTapaTodoElDia(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Bloqueos: []Bloqueo{{
			Desde: momento(t, lunes+" 00:00"),
			Hasta: momento(t, martes+" 00:00"),
		}},
		Desde: desde,
		Hasta: hasta,
		Ahora: momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("se generaron %d huecos %v con el día entero bloqueado", len(huecos), inicios(huecos))
	}
}

func TestGenerarBloqueoParcial(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Bloqueos: []Bloqueo{{
			Desde: momento(t, lunes+" 09:30"),
			Hasta: momento(t, lunes+" 10:30"),
		}},
		Desde: desde,
		Hasta: hasta,
		Ahora: momento(t, "2026-08-01 00:00"),
	}.Generar()

	// 09:00-09:50 y 09:50-10:40 pisan el bloqueo; 10:40 y 11:30 sobreviven
	compararInicios(t, huecos, []string{
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarBordeDelBloqueo(t *testing.T) {
	// El test que más importa de este archivo. Un bloqueo que arranca
	// exactamente cuando termina un hueco no lo pisa, y uno que termina
	// exactamente cuando el hueco arranca tampoco.
	desde, hasta := rango(t, lunes, lunes)

	t.Run("empieza cuando el hueco termina", func(t *testing.T) {
		huecos := CalculoHuecos{
			Horarios: []HorarioSemanal{bloqueLunes()},
			Bloqueos: []Bloqueo{{
				Desde: momento(t, lunes+" 09:50"),
				Hasta: momento(t, lunes+" 23:00"),
			}},
			Desde: desde,
			Hasta: hasta,
			Ahora: momento(t, "2026-08-01 00:00"),
		}.Generar()

		// solo sobrevive el de 09:00, que termina justo a las 09:50
		compararInicios(t, huecos, []string{lunes + " 09:00"})
	})

	t.Run("termina cuando el hueco empieza", func(t *testing.T) {
		huecos := CalculoHuecos{
			Horarios: []HorarioSemanal{bloqueLunes()},
			Bloqueos: []Bloqueo{{
				Desde: momento(t, lunes+" 00:00"),
				Hasta: momento(t, lunes+" 09:00"),
			}},
			Desde: desde,
			Hasta: hasta,
			Ahora: momento(t, "2026-08-01 00:00"),
		}.Generar()

		// no se pierde ninguno: el bloqueo termina justo cuando arranca el primero
		compararInicios(t, huecos, []string{
			lunes + " 09:00",
			lunes + " 09:50",
			lunes + " 10:40",
			lunes + " 11:30",
		})
	})
}

func TestGenerarAnticipacionMinima(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios:              []HorarioSemanal{bloqueLunes()},
		Desde:                 desde,
		Hasta:                 hasta,
		AnticipacionMinimaMin: 120,
		// son las 08:00 del mismo lunes: con dos horas de anticipación, lo
		// más temprano reservable son las 10:00
		Ahora: momento(t, lunes+" 08:00"),
	}.Generar()

	compararInicios(t, huecos, []string{
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarAnticipacionMinimaEnElBordeExacto(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios:              []HorarioSemanal{bloqueLunes()},
		Desde:                 desde,
		Hasta:                 hasta,
		AnticipacionMinimaMin: 60,
		// a las 08:50 con una hora de anticipación, el mínimo cae exactamente
		// en 09:50: ese hueco entra
		Ahora: momento(t, lunes+" 08:50"),
	}.Generar()

	compararInicios(t, huecos, []string{
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarRangoInvertido(t *testing.T) {
	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    momento(t, lunesQueViene+" 00:00"),
		Hasta:    momento(t, lunes+" 00:00"),
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("un rango invertido no produce nada, dio %v", inicios(huecos))
	}
}

func TestGenerarSinHorarios(t *testing.T) {
	desde, hasta := rango(t, lunes, lunesQueViene)

	huecos := CalculoHuecos{
		Desde: desde,
		Hasta: hasta,
		Ahora: momento(t, "2026-08-01 00:00"),
	}.Generar()

	// devuelve una lista vacía, no nil: el handler la serializa como [] y el
	// cliente no tiene que chequear null
	if huecos == nil {
		t.Error("se esperaba una lista vacía, no nil")
	}
	if len(huecos) != 0 {
		t.Errorf("se esperaban 0 huecos, se obtuvieron %d", len(huecos))
	}
}
```

- [ ] **Step 2: Correr el test y verificar que falla**

Run: `cd apps/api && go test ./internal/domain/ -run TestGenerar -v`
Expected: FAIL con `undefined: CalculoHuecos`

- [ ] **Step 3: Implementar el cálculo**

Archivo `apps/api/internal/domain/huecos.go`:

```go
package domain

import (
	"slices"
	"time"
)

// Hueco es un turno reservable: un intervalo concreto en el que un profesional
// atiende y nadie tomó todavía.
type Hueco struct {
	Inicio    time.Time
	Fin       time.Time
	Modalidad Modalidad
}

// CalculoHuecos son todas las entradas del cálculo, explícitas.
//
// Es un struct y no una función con seis parámetros para que nadie confunda el
// orden de dos time.Time, y porque así los tests se leen.
//
// No recibe repositorios ni el reloj del sistema: el servicio carga los datos y
// arma esto. Esa separación es la que permite probar los bordes del calendario
// sin levantar nada.
//
// Cuando exista Turno, se le suma un campo con los turnos ya tomados y se
// restan igual que los bloqueos. La firma pública no cambia.
type CalculoHuecos struct {
	Horarios []HorarioSemanal
	Bloqueos []Bloqueo

	// Desde y Hasta acotan el cálculo como intervalo semiabierto [Desde,
	// Hasta). El servicio traduce las fechas de la consulta —que sí incluyen
	// los dos días— a este formato: del 25 al 27 llega acá como [25 00:00, 28
	// 00:00).
	Desde time.Time
	Hasta time.Time

	AnticipacionMinimaMin int
	Ahora                 time.Time
}

func (c CalculoHuecos) Generar() []Hueco {
	huecos := make([]Hueco, 0)
	if !c.Desde.Before(c.Hasta) {
		return huecos
	}

	minimo := c.Ahora.Add(time.Duration(c.AnticipacionMinimaMin) * time.Minute)

	for dia := InicioDelDia(c.Desde); dia.Before(c.Hasta); dia = dia.AddDate(0, 0, 1) {
		diaSemana := DiaSemanaDe(dia.Weekday())

		for _, bloque := range c.Horarios {
			if bloque.DiaSemana != diaSemana {
				continue
			}
			huecos = append(huecos, c.huecosDelBloque(dia, bloque, minimo)...)
		}
	}

	slices.SortFunc(huecos, func(a, b Hueco) int { return a.Inicio.Compare(b.Inicio) })
	return huecos
}

func (c CalculoHuecos) huecosDelBloque(dia time.Time, bloque HorarioSemanal, minimo time.Time) []Hueco {
	var huecos []Hueco
	duracion := time.Duration(bloque.DuracionMin) * time.Minute

	for minuto := bloque.Desde.Minutos; minuto+bloque.DuracionMin <= bloque.Hasta.Minutos; minuto += bloque.DuracionMin {
		// Se construye con time.Date y no sumándole una duración al inicio del
		// día: sumar es aritmética de instantes, y si el país vuelve a tener
		// horario de verano eso correría la hora de reloj. time.Date respeta
		// que "las nueve" son las nueve.
		inicio := time.Date(dia.Year(), dia.Month(), dia.Day(), minuto/60, minuto%60, 0, 0, ZonaHoraria)
		fin := inicio.Add(duracion)

		switch {
		case inicio.Before(c.Desde) || fin.After(c.Hasta):
			continue // se sale del rango pedido
		case inicio.Before(minimo):
			continue // no llega a la anticipación mínima
		case c.bloqueado(inicio, fin):
			continue
		}

		huecos = append(huecos, Hueco{Inicio: inicio, Fin: fin, Modalidad: bloque.Modalidad})
	}

	return huecos
}

func (c CalculoHuecos) bloqueado(inicio, fin time.Time) bool {
	for _, bloqueo := range c.Bloqueos {
		if bloqueo.SeSolapaCon(inicio, fin) {
			return true
		}
	}
	return false
}

```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/domain/ -run TestGenerar -v`
Expected: PASS en todos. Si `TestGenerarBordeDelBloqueo` falla, la comparación de solapamiento está usando `<=` en vez de `<` — arreglar `SeSolapaCon` en `bloqueo.go`, no el test.

- [ ] **Step 5: Correr toda la suite**

Run: `cd apps/api && go test ./... -count=1` y `go vet ./...`
Expected: `ok` en todos los paquetes, vet limpio.

- [ ] **Step 6: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/domain/huecos.go apps/api/internal/domain/huecos_test.go
git commit -m "feat(domain): cálculo de huecos libres

CalculoHuecos es una función pura sobre entradas explícitas: sin
repositorios, sin reloj del sistema, sin contexto. El servicio carga los
datos y arma el struct. Esa separación es la que permite probar los
bordes del calendario sin levantar nada.

Cada hueco se construye con time.Date y no sumándole una duración al
inicio del día: sumar es aritmética de instantes y correría la hora de
reloj si el país vuelve a tener horario de verano.

El test del borde es el que importa: un bloqueo que arranca exactamente
cuando termina un hueco no lo pisa.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 5: Los repositorios

**Files:**
- Create: `apps/api/internal/repository/horario.go`
- Create: `apps/api/internal/repository/bloqueo.go`
- Create: `apps/api/internal/repository/memory/horario.go`
- Create: `apps/api/internal/repository/memory/horario_test.go`
- Create: `apps/api/internal/repository/memory/bloqueo.go`
- Create: `apps/api/internal/repository/memory/bloqueo_test.go`

**Interfaces:**
- Consumes: `domain.HorarioSemanal`, `domain.Bloqueo`, `domain.ErrNoEncontrado` (Tasks 2-3)
- Produces:
  - `repository.HorarioSemanal` con `ReemplazarDeProfesional`, `ListarDeProfesional`
  - `repository.Bloqueo` con `Crear`, `ObtenerPorID`, `Eliminar`, `ListarDeProfesional`
  - `memory.NuevoHorarioSemanal() *memory.HorarioSemanal`
  - `memory.NuevoBloqueo() *memory.Bloqueo`

- [ ] **Step 1: Escribir las dos interfaces**

Sin test propio: la aserción de compilación en cada implementación es lo que prueba que se cumplen.

Archivo `apps/api/internal/repository/horario.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// HorarioSemanal guarda el horario habitual de cada profesional.
//
// No hay alta ni baja de bloques sueltos: la semana se reemplaza entera. Es lo
// que hace trivial la validación de solapamiento —se valida el conjunto y se
// guarda— y evita que exista un estado con la mitad de los bloques cargados.
type HorarioSemanal interface {
	// ReemplazarDeProfesional deja al profesional exactamente con los bloques
	// recibidos. Una lista vacía lo deja sin horarios, que es legítimo.
	//
	// Es una sola operación a propósito: borra y escribe bajo el mismo lock,
	// igual que un DELETE más INSERT dentro de una transacción cuando llegue
	// PostgreSQL. Si fueran dos llamadas, entre una y otra el profesional queda
	// sin horarios y alguien puede leer ese estado.
	ReemplazarDeProfesional(ctx context.Context, profesionalID uuid.UUID, horarios []domain.HorarioSemanal) error

	ListarDeProfesional(ctx context.Context, profesionalID uuid.UUID) ([]domain.HorarioSemanal, error)
}
```

Archivo `apps/api/internal/repository/bloqueo.go`:

```go
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Bloqueo guarda los bloqueos de agenda. A diferencia de los horarios, se
// manejan de a uno: se agrega unas vacaciones, después se borran.
type Bloqueo interface {
	Crear(ctx context.Context, b domain.Bloqueo) error
	ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Bloqueo, error)
	Eliminar(ctx context.Context, id uuid.UUID) error

	// ListarDeProfesional devuelve los bloqueos que pisan el intervalo
	// semiabierto [desde, hasta), incluidos los que solo lo pisan en parte.
	ListarDeProfesional(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Bloqueo, error)
}
```

- [ ] **Step 2: Escribir el test del repositorio de horarios**

Archivo `apps/api/internal/repository/memory/horario_test.go`:

```go
package memory

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func semanaDePrueba(profesionalID uuid.UUID) []domain.HorarioSemanal {
	return []domain.HorarioSemanal{
		{
			ProfesionalID: profesionalID,
			DiaSemana:     domain.DiaLunes,
			Desde:         domain.HoraDelDia{Minutos: 9 * 60},
			Hasta:         domain.HoraDelDia{Minutos: 13 * 60},
			DuracionMin:   50,
			Modalidad:     domain.ModalidadTelemedicina,
		},
	}
}

func TestReemplazarYListarHorarios(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("ReemplazarDeProfesional devolvió error: %v", err)
	}

	obtenida, err := repo.ListarDeProfesional(ctx, profesionalID)
	if err != nil {
		t.Fatalf("ListarDeProfesional devolvió error: %v", err)
	}
	if len(obtenida) != 1 || obtenida[0].DiaSemana != domain.DiaLunes {
		t.Errorf("la semana recuperada no coincide: %+v", obtenida)
	}
}

func TestReemplazarPisaLaSemanaAnterior(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("primer reemplazo falló: %v", err)
	}

	nueva := semanaDePrueba(profesionalID)
	nueva[0].DiaSemana = domain.DiaMartes

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, nueva); err != nil {
		t.Fatalf("segundo reemplazo falló: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, profesionalID)
	if len(obtenida) != 1 {
		t.Fatalf("len = %d, se esperaba 1: reemplazar no debe acumular", len(obtenida))
	}
	if obtenida[0].DiaSemana != domain.DiaMartes {
		t.Errorf("día = %q, se esperaba martes", obtenida[0].DiaSemana)
	}
}

func TestReemplazarConListaVaciaDejaSinHorarios(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("ReemplazarDeProfesional devolvió error: %v", err)
	}
	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, nil); err != nil {
		t.Fatalf("vaciar la semana devolvió error: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, profesionalID)
	if len(obtenida) != 0 {
		t.Errorf("len = %d, se esperaba 0", len(obtenida))
	}
}

func TestListarHorariosDeProfesionalSinCargar(t *testing.T) {
	obtenida, err := NuevoHorarioSemanal().ListarDeProfesional(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("devolvió error: %v", err)
	}
	// lista vacía, no nil: el handler la serializa como [] y el cliente no
	// tiene que chequear null
	if obtenida == nil {
		t.Error("se esperaba una lista vacía, no nil")
	}
}

func TestHorariosNoSeMezclanEntreProfesionales(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	uno, otro := uuid.New(), uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, uno, semanaDePrueba(uno)); err != nil {
		t.Fatalf("devolvió error: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, otro)
	if len(obtenida) != 0 {
		t.Errorf("el otro profesional no tenía horarios y devolvió %d", len(obtenida))
	}
}

func TestElStoreDeHorariosDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("devolvió error: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, profesionalID)
	obtenida[0].DuracionMin = 999

	fresca, _ := repo.ListarDeProfesional(ctx, profesionalID)
	if fresca[0].DuracionMin == 999 {
		t.Error("mutar el resultado alteró el store")
	}
}
```

- [ ] **Step 3: Correr y verificar que falla**

Run: `cd apps/api && go test ./internal/repository/memory/ -run 'Horario' -v`
Expected: FAIL con `undefined: NuevoHorarioSemanal`

- [ ] **Step 4: Implementar el repositorio de horarios**

Archivo `apps/api/internal/repository/memory/horario.go`:

```go
package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var _ repository.HorarioSemanal = (*HorarioSemanal)(nil)

// HorarioSemanal guarda la semana de cada profesional en memoria.
//
// A diferencia de Profesional, domain.HorarioSemanal no tiene slices ni
// punteros mutables adentro, así que copiar el struct ya es una copia profunda:
// alcanza con clonar el slice que los contiene.
type HorarioSemanal struct {
	mu    sync.RWMutex
	datos map[uuid.UUID][]domain.HorarioSemanal
}

func NuevoHorarioSemanal() *HorarioSemanal {
	return &HorarioSemanal{datos: make(map[uuid.UUID][]domain.HorarioSemanal)}
}

func (r *HorarioSemanal) ReemplazarDeProfesional(_ context.Context, profesionalID uuid.UUID, horarios []domain.HorarioSemanal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(horarios) == 0 {
		delete(r.datos, profesionalID)
		return nil
	}
	r.datos[profesionalID] = slices.Clone(horarios)
	return nil
}

func (r *HorarioSemanal) ListarDeProfesional(_ context.Context, profesionalID uuid.UUID) ([]domain.HorarioSemanal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	horarios, existe := r.datos[profesionalID]
	if !existe {
		return []domain.HorarioSemanal{}, nil
	}
	return slices.Clone(horarios), nil
}
```

- [ ] **Step 5: Escribir el test del repositorio de bloqueos**

Archivo `apps/api/internal/repository/memory/bloqueo_test.go`:

```go
package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func instante(t *testing.T, s string) time.Time {
	t.Helper()
	momento, err := time.ParseInLocation("2006-01-02 15:04", s, domain.ZonaHoraria)
	if err != nil {
		t.Fatalf("no se pudo parsear %q: %v", s, err)
	}
	return momento
}

func bloqueoDePrueba(t *testing.T, profesionalID uuid.UUID, desde, hasta string) domain.Bloqueo {
	t.Helper()
	return domain.Bloqueo{
		ID:            uuid.New(),
		ProfesionalID: profesionalID,
		Desde:         instante(t, desde),
		Hasta:         instante(t, hasta),
		Motivo:        "Vacaciones",
		CreadoEn:      instante(t, "2026-08-22 10:00"),
	}
}

func TestCrearYObtenerBloqueo(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	b := bloqueoDePrueba(t, uuid.New(), "2026-09-10 00:00", "2026-09-20 00:00")

	if err := repo.Crear(ctx, b); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenido, err := repo.ObtenerPorID(ctx, b.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID devolvió error: %v", err)
	}
	if obtenido.ID != b.ID || obtenido.Motivo != "Vacaciones" {
		t.Errorf("el bloqueo recuperado no coincide: %+v", obtenido)
	}
}

func TestObtenerBloqueoInexistente(t *testing.T) {
	_, err := NuevoBloqueo().ObtenerPorID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestEliminarBloqueo(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	b := bloqueoDePrueba(t, uuid.New(), "2026-09-10 00:00", "2026-09-20 00:00")

	if err := repo.Crear(ctx, b); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}
	if err := repo.Eliminar(ctx, b.ID); err != nil {
		t.Fatalf("Eliminar devolvió error: %v", err)
	}
	if _, err := repo.ObtenerPorID(ctx, b.ID); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("el bloqueo seguía existiendo tras eliminarlo")
	}
	if err := repo.Eliminar(ctx, b.ID); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("eliminar algo inexistente debía dar ErrNoEncontrado, dio %v", err)
	}
}

func TestListarBloqueosPorRango(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	profesionalID := uuid.New()

	dentro := bloqueoDePrueba(t, profesionalID, "2026-09-12 00:00", "2026-09-14 00:00")
	pisaElArranque := bloqueoDePrueba(t, profesionalID, "2026-09-05 00:00", "2026-09-11 00:00")
	pisaElFinal := bloqueoDePrueba(t, profesionalID, "2026-09-19 00:00", "2026-09-25 00:00")
	loContiene := bloqueoDePrueba(t, profesionalID, "2026-08-01 00:00", "2026-10-01 00:00")
	muyAntes := bloqueoDePrueba(t, profesionalID, "2026-07-01 00:00", "2026-07-10 00:00")
	muyDespues := bloqueoDePrueba(t, profesionalID, "2026-11-01 00:00", "2026-11-10 00:00")

	for _, b := range []domain.Bloqueo{dentro, pisaElArranque, pisaElFinal, loContiene, muyAntes, muyDespues} {
		if err := repo.Crear(ctx, b); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	obtenidos, err := repo.ListarDeProfesional(ctx, profesionalID,
		instante(t, "2026-09-10 00:00"), instante(t, "2026-09-20 00:00"))
	if err != nil {
		t.Fatalf("ListarDeProfesional devolvió error: %v", err)
	}

	// los cuatro que pisan el rango, aunque sea en parte; los dos de afuera no
	if len(obtenidos) != 4 {
		t.Fatalf("se obtuvieron %d bloqueos, se esperaban 4", len(obtenidos))
	}

	// y salen ordenados por fecha de inicio
	for i := 1; i < len(obtenidos); i++ {
		if obtenidos[i].Desde.Before(obtenidos[i-1].Desde) {
			t.Errorf("los bloqueos no salieron ordenados: %v antes que %v",
				obtenidos[i-1].Desde, obtenidos[i].Desde)
		}
	}
}

func TestBloqueosNoSeMezclanEntreProfesionales(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	uno, otro := uuid.New(), uuid.New()

	if err := repo.Crear(ctx, bloqueoDePrueba(t, uno, "2026-09-10 00:00", "2026-09-20 00:00")); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenidos, _ := repo.ListarDeProfesional(ctx, otro,
		instante(t, "2026-09-01 00:00"), instante(t, "2026-10-01 00:00"))
	if len(obtenidos) != 0 {
		t.Errorf("el otro profesional no tenía bloqueos y devolvió %d", len(obtenidos))
	}
}
```

- [ ] **Step 6: Correr y verificar que falla**

Run: `cd apps/api && go test ./internal/repository/memory/ -run Bloqueo -v`
Expected: FAIL con `undefined: NuevoBloqueo`

- [ ] **Step 7: Implementar el repositorio de bloqueos**

Archivo `apps/api/internal/repository/memory/bloqueo.go`:

```go
package memory

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var _ repository.Bloqueo = (*Bloqueo)(nil)

// Bloqueo guarda los bloqueos de agenda en memoria.
//
// domain.Bloqueo no tiene slices ni punteros mutables, así que copiar el struct
// alcanza: los time.Time comparten un *Location, pero las zonas horarias son
// inmutables y globales.
type Bloqueo struct {
	mu    sync.RWMutex
	datos map[uuid.UUID]domain.Bloqueo
}

func NuevoBloqueo() *Bloqueo {
	return &Bloqueo{datos: make(map[uuid.UUID]domain.Bloqueo)}
}

func (r *Bloqueo) Crear(_ context.Context, b domain.Bloqueo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.datos[b.ID] = b
	return nil
}

func (r *Bloqueo) ObtenerPorID(_ context.Context, id uuid.UUID) (domain.Bloqueo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, existe := r.datos[id]
	if !existe {
		return domain.Bloqueo{}, domain.ErrNoEncontrado
	}
	return b, nil
}

func (r *Bloqueo) Eliminar(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[id]; !existe {
		return domain.ErrNoEncontrado
	}
	delete(r.datos, id)
	return nil
}

func (r *Bloqueo) ListarDeProfesional(_ context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Bloqueo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// ponytail: scan O(n), correcto para un store en memoria. La
	// implementación Postgres lo resuelve con un índice sobre profesional y
	// fecha.
	coincidentes := make([]domain.Bloqueo, 0)
	for _, b := range r.datos {
		if b.ProfesionalID != profesionalID {
			continue
		}
		// SeSolapaCon es la misma comparación que usa el cálculo de huecos, así
		// que el criterio de "pisa el rango" no puede divergir entre las dos.
		if b.SeSolapaCon(desde, hasta) {
			coincidentes = append(coincidentes, b)
		}
	}

	// El mapa de Go itera en orden aleatorio: sin esto, dos llamadas idénticas
	// devolverían los bloqueos en distinto orden.
	slices.SortFunc(coincidentes, func(a, b domain.Bloqueo) int {
		if c := a.Desde.Compare(b.Desde); c != 0 {
			return c
		}
		return strings.Compare(a.ID.String(), b.ID.String())
	})

	return coincidentes, nil
}
```

- [ ] **Step 8: Correr todo y verificar**

Run: `cd apps/api && go test ./... -count=1` y luego, con cgo habilitado, `go test ./internal/repository/... -race`
Expected: `ok` en todo, sin carreras.

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/repository/
git commit -m "feat(repository): horarios y bloqueos

El horario se reemplaza entero bajo un solo lock: si fueran dos
llamadas, entre borrar y escribir el profesional queda sin horarios y
alguien puede leer ese estado.

El listado de bloqueos por rango reutiliza SeSolapaCon del dominio, así
el criterio de 'pisa el rango' no puede divergir entre el repositorio y
el cálculo de huecos. Su test cubre los cuatro casos que sí entran
—adentro, pisando cada borde, y conteniéndolo— y los dos que no.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 6: El servicio `Agenda`

**Files:**
- Create: `apps/api/internal/service/agenda.go`
- Create: `apps/api/internal/service/agenda_test.go`

**Interfaces:**
- Consumes: todo lo anterior, más `repository.Profesional` y `service.Profesional` (del CRUD existente)
- Produces:
  - `service.ResultadoHuecos{Huecos []domain.Hueco; Desde, Hasta time.Time}`
  - `func service.NuevaAgenda(repository.Profesional, repository.HorarioSemanal, repository.Bloqueo) *service.Agenda`
  - `(*Agenda) ReemplazarHorarios(ctx, profesionalID, []domain.EntradaHorarioSemanal) ([]domain.HorarioSemanal, error)`
  - `(*Agenda) ListarHorarios(ctx, profesionalID) ([]domain.HorarioSemanal, error)`
  - `(*Agenda) CrearBloqueo(ctx, profesionalID, domain.EntradaBloqueo) (domain.Bloqueo, error)`
  - `(*Agenda) ListarBloqueos(ctx, profesionalID, desde, hasta time.Time) ([]domain.Bloqueo, error)`
  - `(*Agenda) EliminarBloqueo(ctx, profesionalID, bloqueoID uuid.UUID) error`
  - `(*Agenda) HuecosLibres(ctx, profesionalID, desde, hasta time.Time) (ResultadoHuecos, error)`

- [ ] **Step 1: Escribir el test**

Archivo `apps/api/internal/service/agenda_test.go`:

```go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

const lunesDePrueba = "2026-08-24" // es lunes

func instante(t *testing.T, s string) time.Time {
	t.Helper()
	momento, err := time.ParseInLocation("2006-01-02 15:04", s, domain.ZonaHoraria)
	if err != nil {
		t.Fatalf("no se pudo parsear %q: %v", s, err)
	}
	return momento
}

func dia(t *testing.T, fecha string) time.Time {
	t.Helper()
	return instante(t, fecha+" 00:00")
}

// bancoDePrueba arma el stack real: repositorios en memoria, sin mocks.
type bancoDePrueba struct {
	profesionales *memory.Profesional
	agenda        *Agenda
	svcProf       *Profesional
}

func nuevoBancoDePrueba() *bancoDePrueba {
	profesionales := memory.NuevoProfesional()
	return &bancoDePrueba{
		profesionales: profesionales,
		svcProf:       NuevoProfesional(profesionales),
		agenda:        NuevaAgenda(profesionales, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo()),
	}
}

func (b *bancoDePrueba) crearProfesional(t *testing.T) domain.Profesional {
	t.Helper()
	p, err := b.svcProf.Crear(context.Background(), entradaValida())
	if err != nil {
		t.Fatalf("no se pudo crear el profesional de prueba: %v", err)
	}
	return p
}

func entradaHorarioLunes() domain.EntradaHorarioSemanal {
	return domain.EntradaHorarioSemanal{
		DiaSemana:   "lunes",
		Desde:       "09:00",
		Hasta:       "13:00",
		DuracionMin: 50,
		Modalidad:   "telemedicina",
	}
}

func TestReemplazarHorarios(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	semana, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()})
	if err != nil {
		t.Fatalf("ReemplazarHorarios devolvió error: %v", err)
	}
	if len(semana) != 1 {
		t.Fatalf("len = %d, se esperaba 1", len(semana))
	}

	guardada, err := banco.agenda.ListarHorarios(ctx, p.ID)
	if err != nil {
		t.Fatalf("ListarHorarios devolvió error: %v", err)
	}
	if len(guardada) != 1 {
		t.Error("la semana no quedó persistida")
	}
}

func TestReemplazarHorariosDeProfesionalInexistente(t *testing.T) {
	_, err := nuevoBancoDePrueba().agenda.ReemplazarHorarios(
		context.Background(), uuid.New(), []domain.EntradaHorarioSemanal{entradaHorarioLunes()})
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestReemplazarHorariosConModalidadQueElProfesionalNoOfrece(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()

	// entradaValida() declara solo telemedicina
	p := banco.crearProfesional(t)

	entrada := entradaHorarioLunes()
	entrada.Modalidad = "domicilio"

	_, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entrada})

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
	if len(verr.Campos) == 0 {
		t.Fatal("el error no nombra ningún campo")
	}
}

func TestCrearYListarBloqueos(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	b, err := banco.agenda.CrearBloqueo(ctx, p.ID, domain.EntradaBloqueo{
		Desde:  dia(t, "2099-09-10"),
		Hasta:  dia(t, "2099-09-20"),
		Motivo: "Vacaciones",
	})
	if err != nil {
		t.Fatalf("CrearBloqueo devolvió error: %v", err)
	}

	obtenidos, err := banco.agenda.ListarBloqueos(ctx, p.ID, dia(t, "2099-09-01"), dia(t, "2099-10-01"))
	if err != nil {
		t.Fatalf("ListarBloqueos devolvió error: %v", err)
	}
	if len(obtenidos) != 1 || obtenidos[0].ID != b.ID {
		t.Errorf("no se recuperó el bloqueo creado: %+v", obtenidos)
	}
}

func TestEliminarBloqueoDeOtroProfesional(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	uno := banco.crearProfesional(t)

	otraEntrada := entradaValida()
	otraEntrada.Matricula = "MN 77777"
	otro, err := banco.svcProf.Crear(ctx, otraEntrada)
	if err != nil {
		t.Fatalf("no se pudo crear el segundo profesional: %v", err)
	}

	b, err := banco.agenda.CrearBloqueo(ctx, uno.ID, domain.EntradaBloqueo{
		Desde: dia(t, "2099-09-10"),
		Hasta: dia(t, "2099-09-20"),
	})
	if err != nil {
		t.Fatalf("CrearBloqueo devolvió error: %v", err)
	}

	// borrar el bloqueo de otro desde la ruta de este es un 404, no un éxito
	if err := banco.agenda.EliminarBloqueo(ctx, otro.ID, b.ID); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}

	// y el bloqueo sigue existiendo
	obtenidos, _ := banco.agenda.ListarBloqueos(ctx, uno.ID, dia(t, "2099-09-01"), dia(t, "2099-10-01"))
	if len(obtenidos) != 1 {
		t.Error("el bloqueo se borró igual")
	}
}

func TestHuecosLibres(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	if _, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()}); err != nil {
		t.Fatalf("ReemplazarHorarios devolvió error: %v", err)
	}

	// el reloj del servicio se fija para que la anticipación mínima no borre
	// los huecos del lunes de prueba
	banco.agenda.ahora = func() time.Time { return instante(t, "2026-08-01 00:00") }

	resultado, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, lunesDePrueba), dia(t, lunesDePrueba))
	if err != nil {
		t.Fatalf("HuecosLibres devolvió error: %v", err)
	}
	if len(resultado.Huecos) != 4 {
		t.Errorf("se obtuvieron %d huecos, se esperaban 4", len(resultado.Huecos))
	}
}

func TestHuecosLibresDeProfesionalInactivo(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	if _, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()}); err != nil {
		t.Fatalf("ReemplazarHorarios devolvió error: %v", err)
	}
	if err := banco.svcProf.DarDeBaja(ctx, p.ID); err != nil {
		t.Fatalf("DarDeBaja devolvió error: %v", err)
	}

	banco.agenda.ahora = func() time.Time { return instante(t, "2026-08-01 00:00") }

	resultado, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, lunesDePrueba), dia(t, lunesDePrueba))
	// un profesional dado de baja no tiene disponibilidad, pero el recurso
	// existe: es una lista vacía, no un error
	if err != nil {
		t.Fatalf("se esperaba una lista vacía, se obtuvo error: %v", err)
	}
	if len(resultado.Huecos) != 0 {
		t.Errorf("se obtuvieron %d huecos de un profesional inactivo", len(resultado.Huecos))
	}
}

func TestHuecosLibresRecortaAlHorizonte(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()

	entrada := entradaValida()
	horizonte := 7
	entrada.HorizonteDias = &horizonte
	p, err := banco.svcProf.Crear(ctx, entrada)
	if err != nil {
		t.Fatalf("no se pudo crear el profesional: %v", err)
	}

	// el reloj se fija en el lunes de prueba: el horizonte se cuenta desde hoy
	banco.agenda.ahora = func() time.Time { return instante(t, lunesDePrueba+" 08:00") }

	// se piden 90 días a alguien que expone 7
	resultado, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, lunesDePrueba), dia(t, "2026-11-24"))
	if err != nil {
		t.Fatalf("HuecosLibres devolvió error: %v", err)
	}

	// recorta, no rechaza, y lo informa
	ultimoEsperado := dia(t, "2026-08-30") // hoy + 6 días
	if !resultado.Hasta.Equal(ultimoEsperado) {
		t.Errorf("Hasta = %v, se esperaba %v", resultado.Hasta, ultimoEsperado)
	}
	if !resultado.Desde.Equal(dia(t, lunesDePrueba)) {
		t.Errorf("Desde = %v, se esperaba el pedido", resultado.Desde)
	}
}

func TestHuecosLibresRangoInvertido(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	_, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, "2026-09-10"), dia(t, "2026-09-01"))

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Errorf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
}
```

- [ ] **Step 2: Correr y verificar que falla**

Run: `cd apps/api && go test ./internal/service/ -run 'Agenda|Horarios|Bloqueo|Huecos' -v`
Expected: FAIL con `undefined: NuevaAgenda`

- [ ] **Step 3: Implementar el servicio**

Archivo `apps/api/internal/service/agenda.go`:

```go
package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// Agenda resuelve los casos de uso que necesitan mirar más de una entidad: el
// horario, los bloqueos y el profesional dueño de los dos.
type Agenda struct {
	profesionales repository.Profesional
	horarios      repository.HorarioSemanal
	bloqueos      repository.Bloqueo

	ahora func() time.Time
}

func NuevaAgenda(profesionales repository.Profesional, horarios repository.HorarioSemanal, bloqueos repository.Bloqueo) *Agenda {
	return &Agenda{
		profesionales: profesionales,
		horarios:      horarios,
		bloqueos:      bloqueos,
		ahora:         func() time.Time { return time.Now().In(domain.ZonaHoraria) },
	}
}

// ResultadoHuecos lleva los huecos y el rango que de verdad se usó, que puede
// ser menor al pedido si el horizonte del profesional lo recortó.
type ResultadoHuecos struct {
	Huecos []domain.Hueco
	Desde  time.Time
	Hasta  time.Time
}

func (s *Agenda) ReemplazarHorarios(ctx context.Context, profesionalID uuid.UUID, entradas []domain.EntradaHorarioSemanal) ([]domain.HorarioSemanal, error) {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return nil, err
	}

	semana, err := domain.NuevaSemana(profesionalID, entradas)
	if err != nil {
		return nil, err
	}

	if err := verificarModalidades(semana, profesional); err != nil {
		return nil, err
	}

	if err := s.horarios.ReemplazarDeProfesional(ctx, profesionalID, semana); err != nil {
		return nil, err
	}
	return semana, nil
}

func (s *Agenda) ListarHorarios(ctx context.Context, profesionalID uuid.UUID) ([]domain.HorarioSemanal, error) {
	if _, err := s.profesionales.ObtenerPorID(ctx, profesionalID); err != nil {
		return nil, err
	}
	return s.horarios.ListarDeProfesional(ctx, profesionalID)
}

func (s *Agenda) CrearBloqueo(ctx context.Context, profesionalID uuid.UUID, entrada domain.EntradaBloqueo) (domain.Bloqueo, error) {
	if _, err := s.profesionales.ObtenerPorID(ctx, profesionalID); err != nil {
		return domain.Bloqueo{}, err
	}

	bloqueo, err := domain.NuevoBloqueo(profesionalID, entrada, s.ahora())
	if err != nil {
		return domain.Bloqueo{}, err
	}

	if err := s.bloqueos.Crear(ctx, bloqueo); err != nil {
		return domain.Bloqueo{}, err
	}
	return bloqueo, nil
}

func (s *Agenda) ListarBloqueos(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Bloqueo, error) {
	if _, err := s.profesionales.ObtenerPorID(ctx, profesionalID); err != nil {
		return nil, err
	}
	return s.bloqueos.ListarDeProfesional(ctx, profesionalID, desde, hasta)
}

// EliminarBloqueo exige que el bloqueo sea de ese profesional.
//
// Sin autenticación cualquiera puede llamar a esto, pero al menos la ruta y el
// recurso tienen que ser coherentes: borrar el bloqueo de otro desde la ruta de
// este es un 404, no un éxito silencioso.
func (s *Agenda) EliminarBloqueo(ctx context.Context, profesionalID, bloqueoID uuid.UUID) error {
	if _, err := s.profesionales.ObtenerPorID(ctx, profesionalID); err != nil {
		return err
	}

	bloqueo, err := s.bloqueos.ObtenerPorID(ctx, bloqueoID)
	if err != nil {
		return err
	}
	if bloqueo.ProfesionalID != profesionalID {
		return domain.ErrNoEncontrado
	}

	return s.bloqueos.Eliminar(ctx, bloqueoID)
}

// HuecosLibres calcula los turnos reservables de un profesional en un rango.
//
// desde y hasta son fechas y las dos entran: pedir del 25 al 27 incluye los
// tres días.
func (s *Agenda) HuecosLibres(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) (ResultadoHuecos, error) {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return ResultadoHuecos{}, err
	}

	desde = domain.InicioDelDia(desde)
	hasta = domain.InicioDelDia(hasta)

	if hasta.Before(desde) {
		return ResultadoHuecos{}, domain.ErrorValidacion{Campos: []domain.ErrorCampo{
			{Campo: "hasta", Mensaje: "tiene que ser posterior o igual a desde"},
		}}
	}

	// El horizonte se cuenta desde hoy, no desde la fecha pedida: es cuánto de
	// su agenda el profesional expone hacia adelante. Contarlo desde `desde`
	// dejaría que un cliente pidiera septiembre de 2099 y reservara turnos a
	// tres años vista.
	//
	// Recorta, no rechaza, igual que paginacion.limite en el listado de
	// profesionales, y el resultado informa el rango que de verdad se usó.
	ultimoDia := domain.InicioDelDia(s.ahora()).AddDate(0, 0, profesional.HorizonteDias-1)
	if hasta.After(ultimoDia) {
		hasta = ultimoDia
	}

	// Si el rango entero cae más allá del horizonte no queda nada que
	// calcular, pero el recurso existe: es una lista vacía, no un error.
	if desde.After(ultimoDia) {
		return ResultadoHuecos{Huecos: []domain.Hueco{}, Desde: desde, Hasta: desde}, nil
	}

	resultado := ResultadoHuecos{Huecos: []domain.Hueco{}, Desde: desde, Hasta: hasta}

	// Un profesional dado de baja no tiene disponibilidad, sin importar lo que
	// digan sus reglas. No es un error: el recurso existe, simplemente no opera.
	if profesional.Estado != domain.EstadoActivo {
		return resultado, nil
	}

	horarios, err := s.horarios.ListarDeProfesional(ctx, profesionalID)
	if err != nil {
		return ResultadoHuecos{}, err
	}

	// el cálculo trabaja con el intervalo semiabierto [desde, finExclusivo)
	finExclusivo := hasta.AddDate(0, 0, 1)

	bloqueos, err := s.bloqueos.ListarDeProfesional(ctx, profesionalID, desde, finExclusivo)
	if err != nil {
		return ResultadoHuecos{}, err
	}

	resultado.Huecos = domain.CalculoHuecos{
		Horarios:              horarios,
		Bloqueos:              bloqueos,
		Desde:                 desde,
		Hasta:                 finExclusivo,
		AnticipacionMinimaMin: profesional.AnticipacionMinimaMin,
		Ahora:                 s.ahora(),
	}.Generar()

	return resultado, nil
}

// verificarModalidades comprueba que cada bloque use una modalidad que el
// profesional declara ofrecer.
//
// Es una regla entre dos entidades y por eso no puede vivir en el dominio: un
// bloque solo no sabe qué ofrece su profesional. Cargar un bloque presencial en
// un perfil que solo hace telemedicina produce huecos que el paciente ve y no
// puede usar.
//
// Identifica el bloque por su día y hora y no por su índice: NuevaSemana
// devuelve la semana ordenada, así que el índice ya no coincide con el orden en
// que el cliente los mandó.
func verificarModalidades(semana []domain.HorarioSemanal, profesional domain.Profesional) error {
	var campos []domain.ErrorCampo

	for _, bloque := range semana {
		if slices.Contains(profesional.Modalidades, bloque.Modalidad) {
			continue
		}
		campos = append(campos, domain.ErrorCampo{
			Campo: "horarios",
			Mensaje: fmt.Sprintf("el bloque de %s %s usa la modalidad %q, que el profesional no ofrece (ofrece: %s)",
				bloque.DiaSemana, bloque.Desde, bloque.Modalidad, listarModalidades(profesional.Modalidades)),
		})
	}

	if len(campos) > 0 {
		return domain.ErrorValidacion{Campos: campos}
	}
	return nil
}

func listarModalidades(modalidades []domain.Modalidad) string {
	nombres := make([]string, 0, len(modalidades))
	for _, m := range modalidades {
		nombres = append(nombres, string(m))
	}
	return strings.Join(nombres, ", ")
}
```

- [ ] **Step 4: Correr los tests y verificar que pasan**

Run: `cd apps/api && go test ./internal/service/ -count=1 -v 2>&1 | tail -30`
Expected: PASS en los tests nuevos y en los del CRUD que ya existían.

- [ ] **Step 5: Correr toda la suite con detector de carreras**

Run (con cgo habilitado): `cd apps/api && go test ./... -race -count=1`
Expected: `ok` en todos los paquetes.

- [ ] **Step 6: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/service/agenda.go apps/api/internal/service/agenda_test.go
git commit -m "feat(service): casos de uso de la agenda

La modalidad de un bloque se valida contra las que el profesional
declara: es una regla entre dos entidades y por eso no puede vivir en el
dominio. Identifica el bloque por día y hora y no por índice, porque
NuevaSemana devuelve la semana ordenada y el índice ya no coincide con
el orden en que llegó.

El horizonte recorta y no rechaza, e informa el rango que se usó de
verdad. Un profesional dado de baja devuelve una lista vacía y no un
error: el recurso existe, simplemente no opera.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 7: El contrato OpenAPI

Se escribe antes que los handlers, como en la etapa anterior: el YAML es la fuente de verdad y la Task 8 lo implementa.

**Files:**
- Modify: `apps/api/api/openapi.yaml`

**Interfaces:**
- Consumes: los nombres de campo de las Tasks 2-6
- Produces: el contrato que implementa la Task 8

- [ ] **Step 1: Sumar los dos campos nuevos a los schemas de `Profesional`**

En `components.schemas.Profesional`, agregar a `properties` (y a `required`, porque el servidor siempre los manda):

```yaml
        anticipacionMinimaMin:
          type: integer
          minimum: 0
          maximum: 10080
          default: 120
          description: |
            Con cuánta anticipación mínima, en minutos, este profesional acepta
            reservas. El tope es una semana.
        horizonteDias:
          type: integer
          minimum: 1
          maximum: 180
          default: 60
          description: Cuántos días de agenda expone hacia adelante.
```

En `components.schemas.PeticionProfesional`, agregar las mismas dos propiedades **sin** sumarlas a `required`: omitirlas significa usar el valor por defecto.

- [ ] **Step 2: Sumar los schemas de la agenda**

En `components.schemas`:

```yaml
    DiaSemana:
      type: string
      enum: [lunes, martes, miercoles, jueves, viernes, sabado, domingo]

    HorarioSemanal:
      type: object
      required: [diaSemana, desde, hasta, duracionMin, modalidad]
      properties:
        diaSemana:
          $ref: '#/components/schemas/DiaSemana'
        desde:
          type: string
          pattern: '^([01][0-9]|2[0-3]):[0-5][0-9]$'
          description: |
            Hora de reloj, no un instante. "Atiendo de 9 a 13" significa las 9
            del reloj del profesional, siempre.
          example: "09:00"
        hasta:
          type: string
          pattern: '^([01][0-9]|2[0-3]):[0-5][0-9]$'
          example: "13:00"
        duracionMin:
          type: integer
          minimum: 10
          maximum: 480
          description: |
            Duración de cada sesión. Va en el bloque y no en el profesional
            porque una teleconsulta y una presencial no duran lo mismo.
          example: 50
        modalidad:
          $ref: '#/components/schemas/Modalidad'

    ListaHorarios:
      type: object
      required: [horarios]
      properties:
        horarios:
          type: array
          items: { $ref: '#/components/schemas/HorarioSemanal' }

    Bloqueo:
      type: object
      required: [id, desde, hasta, motivo, creadoEn]
      properties:
        id: { type: string, format: uuid }
        desde: { type: string, format: date-time }
        hasta: { type: string, format: date-time }
        motivo:
          type: string
          maxLength: 200
          description: Opcional. Puede ser una cadena vacía.
        creadoEn: { type: string, format: date-time }

    PeticionBloqueo:
      type: object
      required: [desde, hasta]
      properties:
        desde: { type: string, format: date-time }
        hasta: { type: string, format: date-time }
        motivo: { type: string, maxLength: 200 }
      additionalProperties: false

    ListaBloqueos:
      type: object
      required: [datos]
      properties:
        datos:
          type: array
          items: { $ref: '#/components/schemas/Bloqueo' }

    Hueco:
      type: object
      required: [inicio, fin, modalidad]
      properties:
        inicio: { type: string, format: date-time }
        fin: { type: string, format: date-time }
        modalidad: { $ref: '#/components/schemas/Modalidad' }

    ListaHuecos:
      type: object
      required: [datos, rango]
      properties:
        datos:
          type: array
          items: { $ref: '#/components/schemas/Hueco' }
        rango:
          type: object
          required: [desde, hasta]
          description: |
            El rango que de verdad se usó, que puede ser menor al pedido si el
            horizonte del profesional lo recortó.
          properties:
            desde: { type: string, format: date }
            hasta: { type: string, format: date }
```

Y en `components.schemas.PeticionProfesional`, nada más. El cuerpo del `PUT` de horarios usa este schema nuevo:

```yaml
    PeticionHorarios:
      type: object
      required: [horarios]
      properties:
        horarios:
          type: array
          items:
            type: object
            required: [diaSemana, desde, hasta, duracionMin, modalidad]
            properties:
              diaSemana: { $ref: '#/components/schemas/DiaSemana' }
              # Las restricciones repiten las de HorarioSemanal a propósito: si
              # el pedido admitiera lo que la respuesta prohíbe, un cliente
              # generado desde este contrato mandaría horarios que su propio
              # schema de respuesta declara inválidos.
              desde:
                type: string
                pattern: '^([01][0-9]|2[0-3]):[0-5][0-9]$'
                example: "09:00"
              hasta:
                type: string
                pattern: '^([01][0-9]|2[0-3]):[0-5][0-9]$'
                example: "13:00"
              duracionMin:
                type: integer
                minimum: 10
                maximum: 480
                example: 50
              modalidad: { $ref: '#/components/schemas/Modalidad' }
            additionalProperties: false
      additionalProperties: false
```

- [ ] **Step 3: Sumar las seis rutas**

En `paths`, después de las que ya existen:

```yaml
  /api/v1/profesionales/{id}/horarios:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }

    get:
      tags: [profesionales]
      summary: El horario semanal del profesional
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ListaHorarios' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

    put:
      tags: [profesionales]
      summary: Reemplaza el horario semanal completo
      description: |
        Reemplazo total: la semana se edita como una unidad. Una lista vacía
        deja al profesional sin horarios, que es legítimo.

        Se rechaza con 422 si dos bloques del mismo día se solapan, si en algún
        bloque no entra ninguna sesión, o si algún bloque usa una modalidad que
        el profesional no declara ofrecer.
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PeticionHorarios' }
      responses:
        '200':
          description: La semana resultante, ordenada
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ListaHorarios' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }
        '422': { $ref: '#/components/responses/ValidacionFallida' }

  /api/v1/profesionales/{id}/bloqueos:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }

    get:
      tags: [profesionales]
      summary: Los bloqueos del profesional
      parameters:
        - name: desde
          in: query
          schema: { type: string, format: date }
        - name: hasta
          in: query
          schema: { type: string, format: date }
      description: |
        Sin `desde` ni `hasta` devuelve los vigentes y futuros. Con rango,
        devuelve los que lo pisan, aunque sea en parte.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ListaBloqueos' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

    post:
      tags: [profesionales]
      summary: Bloquea un rango de la agenda
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PeticionBloqueo' }
      responses:
        '201':
          description: Creado
          headers:
            Location:
              schema: { type: string }
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Bloqueo' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }
        '422': { $ref: '#/components/responses/ValidacionFallida' }

  /api/v1/profesionales/{id}/bloqueos/{bloqueoId}:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }
      - name: bloqueoId
        in: path
        required: true
        schema: { type: string, format: uuid }
    delete:
      tags: [profesionales]
      summary: Elimina un bloqueo
      description: |
        El bloqueo tiene que ser de ese profesional. Borrar el de otro desde
        esta ruta es un 404.
      responses:
        '204': { description: Eliminado }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }

  /api/v1/profesionales/{id}/huecos:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string, format: uuid }
      - name: desde
        in: query
        required: true
        schema: { type: string, format: date }
      - name: hasta
        in: query
        required: true
        schema: { type: string, format: date }
    get:
      tags: [profesionales]
      summary: Los huecos reservables del profesional
      description: |
        `desde` y `hasta` son fechas y las dos entran: del 25 al 27 incluye los
        tres días.

        El rango se recorta al horizonte del profesional, no se rechaza, y
        `rango` en la respuesta informa el que de verdad se usó.

        Un profesional dado de baja devuelve una lista vacía, no un error.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ListaHuecos' }
        '400': { $ref: '#/components/responses/PeticionInvalida' }
        '404': { $ref: '#/components/responses/NoEncontrado' }
        '422': { $ref: '#/components/responses/ValidacionFallida' }
```

- [ ] **Step 4: Validar el contrato**

Run desde la raíz del repo:
```bash
python -c "import yaml; yaml.safe_load(open('apps/api/api/openapi.yaml',encoding='utf-8')); print('YAML OK')"
npx --yes @redocly/cli@2.47.0 lint apps/api/api/openapi.yaml
```
Expected: `YAML OK`, y Redocly con **cero errores**. Las advertencias de estilo son conocidas y están fuera de alcance. Si aparece un error de `$ref` que no resuelve, es un schema mal escrito: arreglarlo.

- [ ] **Step 5: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/api/openapi.yaml
git commit -m "docs(api): contrato de la agenda

Seis rutas nuevas colgando del profesional, porque una disponibilidad no
existe sin dueño. El horario se reemplaza entero: una semana se edita
como una unidad.

Los horarios viajan como hora de reloj (\"09:00\") y los huecos como
instantes con offset, que es la distinción que sostiene todo el modelo.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 8: DTOs, handlers y rutas

**Files:**
- Create: `apps/api/internal/handler/dto_agenda.go`
- Create: `apps/api/internal/handler/agenda.go`
- Create: `apps/api/internal/handler/agenda_test.go`
- Modify: `apps/api/internal/handler/dto.go` (los dos campos nuevos de `Profesional`)
- Modify: `apps/api/internal/handler/router.go`
- Modify: `apps/api/cmd/api/main.go`

**Interfaces:**
- Consumes: `service.Agenda`, `service.ResultadoHuecos`, los DTOs y helpers que ya existen (`escribirJSON`, `escribirError`, `escribirPeticionInvalida`, `decodificarJSON`, `parsearID`)
- Produces:
  - `handler.ManejadorAgenda` con `ListarHorarios`, `ReemplazarHorarios`, `ListarBloqueos`, `CrearBloqueo`, `EliminarBloqueo`, `HuecosLibres`
  - `func handler.NuevaAgenda(*service.Agenda) *ManejadorAgenda`
  - `NuevoRouter` pasa a recibir también el manejador de agenda

- [ ] **Step 1: Escribir el test**

Archivo `apps/api/internal/handler/agenda_test.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

const cuerpoHorarios = `{
  "horarios": [
    {"diaSemana": "lunes", "desde": "09:00", "hasta": "13:00", "duracionMin": 50, "modalidad": "telemedicina"}
  ]
}`

// servidorConAgenda cablea el stack real completo, sin mocks.
func servidorConAgenda(t *testing.T) *httptest.Server {
	t.Helper()

	profesionales := memory.NuevoProfesional()
	svcProf := service.NuevoProfesional(profesionales)
	svcAgenda := service.NuevaAgenda(profesionales, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo())

	router := NuevoRouter(NuevoProfesional(svcProf), NuevaAgenda(svcAgenda))
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

// proximoLunes devuelve el lunes siguiente a dentro de una semana, en formato
// AAAA-MM-DD.
//
// Los tests de handler corren contra el reloj real —el servicio no está
// intervenido— y el horizonte por defecto son 60 días desde hoy, así que una
// fecha clavada como "2099-09-07" quedaría fuera de rango y además dependería
// de que ese día caiga lunes. Calcularla lo vuelve robusto al paso del tiempo.
func proximoLunes(t *testing.T) string {
	t.Helper()
	fecha := time.Now().In(domain.ZonaHoraria).AddDate(0, 0, 7)
	for fecha.Weekday() != time.Monday {
		fecha = fecha.AddDate(0, 0, 1)
	}
	return fecha.Format("2006-01-02")
}

func crearProfesionalPorHTTP(t *testing.T, srv *httptest.Server) respuestaProfesional {
	t.Helper()
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo crear el profesional: status %d", resp.StatusCode)
	}
	return decodificarProfesional(t, resp)
}

func TestPutHorarios(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpoHorarios)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	var lista respuestaHorarios
	if err := json.NewDecoder(resp.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if len(lista.Horarios) != 1 {
		t.Fatalf("len = %d, se esperaba 1", len(lista.Horarios))
	}
	if lista.Horarios[0].Desde != "09:00" {
		t.Errorf("desde = %q, se esperaba 09:00", lista.Horarios[0].Desde)
	}
}

func TestPutHorariosConBloqueSinHuecos(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	cuerpo := strings.Replace(cuerpoHorarios, `"hasta": "13:00"`, `"hasta": "09:30"`, 1)
	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, se esperaba 422", resp.StatusCode)
	}

	var problema Problema
	if err := json.NewDecoder(resp.Body).Decode(&problema); err != nil {
		t.Fatalf("no se pudo decodificar el problema: %v", err)
	}
	if len(problema.Errores) == 0 {
		t.Error("el 422 no nombra ningún campo")
	}
}

func TestPutHorariosConModalidadNoOfrecida(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	cuerpo := strings.Replace(cuerpoHorarios, `"telemedicina"`, `"domicilio"`, 1)
	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, se esperaba 422", resp.StatusCode)
	}
}

func TestGetHorariosDeProfesionalSinAgenda(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/horarios")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	cuerpo := new(strings.Builder)
	if _, err := cuerpo.ReadFrom(resp.Body); err != nil {
		t.Fatalf("no se pudo leer el cuerpo: %v", err)
	}
	// lista vacía, nunca null
	if strings.Contains(cuerpo.String(), `"horarios":null`) {
		t.Errorf("horarios llegó como null: %s", cuerpo.String())
	}
}

func TestHorariosDeProfesionalInexistente(t *testing.T) {
	srv := servidorConAgenda(t)
	resp := obtener(t, srv, "/api/v1/profesionales/6ba7b810-9dad-11d1-80b4-00c04fd430c8/horarios")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, se esperaba 404", resp.StatusCode)
	}
}

func TestCicloDeBloqueo(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)
	base := "/api/v1/profesionales/" + p.ID + "/bloqueos"

	cuerpo := `{"desde": "2099-09-10T00:00:00-03:00", "hasta": "2099-09-20T00:00:00-03:00", "motivo": "Vacaciones"}`
	resp := postear(t, srv, base, cuerpo)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201", resp.StatusCode)
	}

	var bloqueo respuestaBloqueo
	if err := json.NewDecoder(resp.Body).Decode(&bloqueo); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if bloqueo.Motivo != "Vacaciones" {
		t.Errorf("motivo = %q", bloqueo.Motivo)
	}
	if loc := resp.Header.Get("Location"); loc != base+"/"+bloqueo.ID {
		t.Errorf("Location = %q, se esperaba %q", loc, base+"/"+bloqueo.ID)
	}

	listado := obtener(t, srv, base)
	var lista respuestaBloqueos
	if err := json.NewDecoder(listado.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar el listado: %v", err)
	}
	if len(lista.Datos) != 1 {
		t.Errorf("el listado devolvió %d bloqueos, se esperaba 1", len(lista.Datos))
	}

	borrado := ejecutar(t, srv, http.MethodDelete, base+"/"+bloqueo.ID, "")
	if borrado.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, se esperaba 204", borrado.StatusCode)
	}

	deNuevo := ejecutar(t, srv, http.MethodDelete, base+"/"+bloqueo.ID, "")
	if deNuevo.StatusCode != http.StatusNotFound {
		t.Errorf("borrar dos veces debía dar 404, dio %d", deNuevo.StatusCode)
	}
}

func TestCrearBloqueoEnElPasado(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	cuerpo := `{"desde": "2020-01-01T00:00:00-03:00", "hasta": "2020-01-10T00:00:00-03:00"}`
	resp := postear(t, srv, "/api/v1/profesionales/"+p.ID+"/bloqueos", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, se esperaba 422", resp.StatusCode)
	}
}

func TestGetHuecos(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	if resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpoHorarios); resp.StatusCode != http.StatusOK {
		t.Fatalf("no se pudo cargar el horario: status %d", resp.StatusCode)
	}

	lunes := proximoLunes(t)
	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde="+lunes+"&hasta="+lunes)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	var lista respuestaHuecos
	if err := json.NewDecoder(resp.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if lista.Rango.Desde != lunes || lista.Rango.Hasta != lunes {
		t.Errorf("rango = %+v, se esperaba el pedido", lista.Rango)
	}
	if len(lista.Datos) != 4 {
		t.Errorf("se obtuvieron %d huecos, se esperaban 4", len(lista.Datos))
	}
}

func TestGetHuecosSinParametros(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestGetHuecosConFechaMalFormada(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde=ayer&hasta=2099-09-07")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestGetHuecosSeRecortaAlHorizonte(t *testing.T) {
	srv := servidorConAgenda(t)

	// cuerpo propio en vez de parchear cuentaValido con un replace: el replace
	// depende de cómo esté formateada esa constante y se rompe en silencio si
	// alguien la reindenta
	cuerpo := `{
	  "nombre": "Ana", "apellido": "Pérez",
	  "matricula": "MP 55.123", "especialidad": "odontologia",
	  "bio": "Odontóloga general.", "precioConsultaCentavos": 1800000,
	  "modalidades": ["presencial"], "zona": "CABA", "obrasSociales": [],
	  "horizonteDias": 7
	}`
	resp := postear(t, srv, "/api/v1/profesionales", cuerpo)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo crear el profesional: status %d", resp.StatusCode)
	}
	p := decodificarProfesional(t, resp)

	// El horizonte se cuenta desde hoy, así que las fechas se calculan
	// relativas al reloj real: el servicio no está intervenido en este test.
	hoy := time.Now().In(domain.ZonaHoraria)
	desde := hoy.Format("2006-01-02")
	hasta := hoy.AddDate(0, 3, 0).Format("2006-01-02")
	ultimoEsperado := domain.InicioDelDia(hoy).AddDate(0, 0, 6).Format("2006-01-02")

	huecos := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde="+desde+"&hasta="+hasta)
	if huecos.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200: el horizonte recorta, no rechaza", huecos.StatusCode)
	}

	var lista respuestaHuecos
	if err := json.NewDecoder(huecos.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if lista.Rango.Hasta != ultimoEsperado {
		t.Errorf("rango.hasta = %q, se esperaba %q (siete días contados desde hoy)", lista.Rango.Hasta, ultimoEsperado)
	}
}
```

- [ ] **Step 2: Correr y verificar que falla**

Run: `cd apps/api && go test ./internal/handler/ -run 'Horarios|Bloqueo|Huecos' -v`
Expected: FAIL con `undefined: NuevaAgenda` y `undefined: respuestaHorarios`

- [ ] **Step 3: Escribir los DTOs de la agenda**

Archivo `apps/api/internal/handler/dto_agenda.go`:

```go
package handler

import (
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

const formatoFecha = "2006-01-02"

type peticionHorario struct {
	DiaSemana   string `json:"diaSemana"`
	Desde       string `json:"desde"`
	Hasta       string `json:"hasta"`
	DuracionMin int    `json:"duracionMin"`
	Modalidad   string `json:"modalidad"`
}

type peticionHorarios struct {
	Horarios []peticionHorario `json:"horarios"`
}

func (p peticionHorarios) aEntradas() []domain.EntradaHorarioSemanal {
	entradas := make([]domain.EntradaHorarioSemanal, 0, len(p.Horarios))
	for _, h := range p.Horarios {
		entradas = append(entradas, domain.EntradaHorarioSemanal{
			DiaSemana:   h.DiaSemana,
			Desde:       h.Desde,
			Hasta:       h.Hasta,
			DuracionMin: h.DuracionMin,
			Modalidad:   h.Modalidad,
		})
	}
	return entradas
}

type respuestaHorario struct {
	DiaSemana   string `json:"diaSemana"`
	Desde       string `json:"desde"`
	Hasta       string `json:"hasta"`
	DuracionMin int    `json:"duracionMin"`
	Modalidad   string `json:"modalidad"`
}

type respuestaHorarios struct {
	Horarios []respuestaHorario `json:"horarios"`
}

func aRespuestaHorarios(semana []domain.HorarioSemanal) respuestaHorarios {
	horarios := make([]respuestaHorario, 0, len(semana))
	for _, h := range semana {
		horarios = append(horarios, respuestaHorario{
			DiaSemana:   string(h.DiaSemana),
			Desde:       h.Desde.String(),
			Hasta:       h.Hasta.String(),
			DuracionMin: h.DuracionMin,
			Modalidad:   string(h.Modalidad),
		})
	}
	return respuestaHorarios{Horarios: horarios}
}

type peticionBloqueo struct {
	Desde  time.Time `json:"desde"`
	Hasta  time.Time `json:"hasta"`
	Motivo string    `json:"motivo"`
}

func (p peticionBloqueo) aEntrada() domain.EntradaBloqueo {
	return domain.EntradaBloqueo{
		Desde:  p.Desde.In(domain.ZonaHoraria),
		Hasta:  p.Hasta.In(domain.ZonaHoraria),
		Motivo: p.Motivo,
	}
}

type respuestaBloqueo struct {
	ID       string    `json:"id"`
	Desde    time.Time `json:"desde"`
	Hasta    time.Time `json:"hasta"`
	Motivo   string    `json:"motivo"`
	CreadoEn time.Time `json:"creadoEn"`
}

func aRespuestaBloqueo(b domain.Bloqueo) respuestaBloqueo {
	return respuestaBloqueo{
		ID:       b.ID.String(),
		Desde:    b.Desde.In(domain.ZonaHoraria),
		Hasta:    b.Hasta.In(domain.ZonaHoraria),
		Motivo:   b.Motivo,
		CreadoEn: b.CreadoEn.In(domain.ZonaHoraria),
	}
}

type respuestaBloqueos struct {
	Datos []respuestaBloqueo `json:"datos"`
}

func aRespuestaBloqueos(bloqueos []domain.Bloqueo) respuestaBloqueos {
	datos := make([]respuestaBloqueo, 0, len(bloqueos))
	for _, b := range bloqueos {
		datos = append(datos, aRespuestaBloqueo(b))
	}
	return respuestaBloqueos{Datos: datos}
}

type respuestaHueco struct {
	Inicio    time.Time `json:"inicio"`
	Fin       time.Time `json:"fin"`
	Modalidad string    `json:"modalidad"`
}

type rangoRespuesta struct {
	Desde string `json:"desde"`
	Hasta string `json:"hasta"`
}

type respuestaHuecos struct {
	Datos []respuestaHueco `json:"datos"`
	Rango rangoRespuesta   `json:"rango"`
}

func aRespuestaHuecos(resultado service.ResultadoHuecos) respuestaHuecos {
	datos := make([]respuestaHueco, 0, len(resultado.Huecos))
	for _, h := range resultado.Huecos {
		datos = append(datos, respuestaHueco{
			Inicio:    h.Inicio,
			Fin:       h.Fin,
			Modalidad: string(h.Modalidad),
		})
	}
	return respuestaHuecos{
		Datos: datos,
		Rango: rangoRespuesta{
			Desde: resultado.Desde.Format(formatoFecha),
			Hasta: resultado.Hasta.Format(formatoFecha),
		},
	}
}
```

- [ ] **Step 4: Sumar los dos campos nuevos a los DTOs de `Profesional`**

En `apps/api/internal/handler/dto.go`:

En `peticionProfesional`, agregar:
```go
	AnticipacionMinimaMin *int `json:"anticipacionMinimaMin"`
	HorizonteDias         *int `json:"horizonteDias"`
```

En `aEntrada()`, pasarlos:
```go
		AnticipacionMinimaMin: r.AnticipacionMinimaMin,
		HorizonteDias:         r.HorizonteDias,
```

En `respuestaProfesional`, agregar:
```go
	AnticipacionMinimaMin int `json:"anticipacionMinimaMin"`
	HorizonteDias         int `json:"horizonteDias"`
```

En `aRespuesta()`, poblarlos desde `p.AnticipacionMinimaMin` y `p.HorizonteDias`.

- [ ] **Step 5: Escribir los handlers**

Archivo `apps/api/internal/handler/agenda.go`:

```go
package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// ManejadorAgenda traduce entre HTTP y el servicio de agenda. Como el resto de
// los handlers, no decide nada: decodifica, delega y serializa.
type ManejadorAgenda struct {
	svc *service.Agenda
}

func NuevaAgenda(svc *service.Agenda) *ManejadorAgenda {
	return &ManejadorAgenda{svc: svc}
}

func (h *ManejadorAgenda) ListarHorarios(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	semana, err := h.svc.ListarHorarios(r.Context(), id)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaHorarios(semana))
}

func (h *ManejadorAgenda) ReemplazarHorarios(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	var peticion peticionHorarios
	if err := decodificarJSON(w, r, &peticion); err != nil {
		escribirPeticionInvalida(w, "el cuerpo no es un JSON válido")
		return
	}

	semana, err := h.svc.ReemplazarHorarios(r.Context(), id, peticion.aEntradas())
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaHorarios(semana))
}

func (h *ManejadorAgenda) ListarBloqueos(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	consulta := r.URL.Query()
	ahora := time.Now().In(domain.ZonaHoraria)

	// Sin rango se devuelven los vigentes y futuros. La traducción se hace acá
	// para que la interfaz del repositorio no necesite punteros ni centinelas.
	desde := ahora
	hasta := ahora.AddDate(10, 0, 0)

	if crudo := consulta.Get("desde"); crudo != "" {
		fecha, err := parsearFecha(crudo)
		if err != nil {
			escribirPeticionInvalida(w, "el parámetro desde tiene que ser una fecha AAAA-MM-DD")
			return
		}
		desde = fecha
	}
	if crudo := consulta.Get("hasta"); crudo != "" {
		fecha, err := parsearFecha(crudo)
		if err != nil {
			escribirPeticionInvalida(w, "el parámetro hasta tiene que ser una fecha AAAA-MM-DD")
			return
		}
		hasta = fecha.AddDate(0, 0, 1) // el día pedido entra entero
	}

	bloqueos, err := h.svc.ListarBloqueos(r.Context(), id, desde, hasta)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaBloqueos(bloqueos))
}

func (h *ManejadorAgenda) CrearBloqueo(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	var peticion peticionBloqueo
	if err := decodificarJSON(w, r, &peticion); err != nil {
		escribirPeticionInvalida(w, "el cuerpo no es un JSON válido")
		return
	}

	bloqueo, err := h.svc.CrearBloqueo(r.Context(), id, peticion.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/profesionales/"+id.String()+"/bloqueos/"+bloqueo.ID.String())
	escribirJSON(w, http.StatusCreated, aRespuestaBloqueo(bloqueo))
}

func (h *ManejadorAgenda) EliminarBloqueo(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	bloqueoID, err := uuid.Parse(r.PathValue("bloqueoId"))
	if err != nil {
		escribirPeticionInvalida(w, "el id del bloqueo tiene que ser un UUID válido")
		return
	}

	if err := h.svc.EliminarBloqueo(r.Context(), id, bloqueoID); err != nil {
		escribirError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorAgenda) HuecosLibres(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	consulta := r.URL.Query()

	desde, err := parsearFecha(consulta.Get("desde"))
	if err != nil {
		escribirPeticionInvalida(w, "el parámetro desde es obligatorio y tiene que ser una fecha AAAA-MM-DD")
		return
	}

	hasta, err := parsearFecha(consulta.Get("hasta"))
	if err != nil {
		escribirPeticionInvalida(w, "el parámetro hasta es obligatorio y tiene que ser una fecha AAAA-MM-DD")
		return
	}

	resultado, err := h.svc.HuecosLibres(r.Context(), id, desde, hasta)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaHuecos(resultado))
}

// parsearFecha lee una fecha AAAA-MM-DD como el arranque de ese día en la zona
// del sistema.
func parsearFecha(crudo string) (time.Time, error) {
	return time.ParseInLocation(formatoFecha, crudo, domain.ZonaHoraria)
}
```

- [ ] **Step 6: Registrar las rutas**

En `apps/api/internal/handler/router.go`, cambiar la firma y agregar las seis rutas:

```go
func NuevoRouter(ph *ManejadorProfesional, ah *ManejadorAgenda) http.Handler {
	mux := http.NewServeMux()

	// ... las rutas que ya existen, sin cambios ...

	mux.HandleFunc("GET /api/v1/profesionales/{id}/horarios", ah.ListarHorarios)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}/horarios", ah.ReemplazarHorarios)
	mux.HandleFunc("GET /api/v1/profesionales/{id}/bloqueos", ah.ListarBloqueos)
	mux.HandleFunc("POST /api/v1/profesionales/{id}/bloqueos", ah.CrearBloqueo)
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}/bloqueos/{bloqueoId}", ah.EliminarBloqueo)
	mux.HandleFunc("GET /api/v1/profesionales/{id}/huecos", ah.HuecosLibres)

	// ... el resto igual ...
}
```

Nota sobre las colisiones: `POST /profesionales/{id}/reactivar` y las rutas nuevas comparten forma con otras del mismo largo y se separan por el método. Las rutas nuevas tienen literales distintos en la última posición (`horarios`, `bloqueos`, `huecos`), así que el `ServeMux` las prefiere sobre `{id}` por especificidad. Si alguna se registra sin su verbo adelante, el `ServeMux` entra en pánico al arrancar.

(Revisión posterior: la ruta vieja del perfil público por slug bajo `/profesionales` chocaba de forma irresoluble con estas tres rutas nuevas —mismo largo de cinco segmentos, literal nuevo en una posición distinta en cada par— y se movió a `GET /api/v1/perfiles/{slug}`, su propio recurso. Ver `apps/api/internal/handler/router.go`.)

- [ ] **Step 7: Cablear en `main.go`**

En `apps/api/cmd/api/main.go`, después de crear `repo`:

```go
	repoHorarios := memory.NuevoHorarioSemanal()
	repoBloqueos := memory.NuevoBloqueo()

	svc := service.NuevoProfesional(repo)
	svcAgenda := service.NuevaAgenda(repo, repoHorarios, repoBloqueos)

	router := handler.NuevoRouter(
		handler.NuevoProfesional(svc),
		handler.NuevaAgenda(svcAgenda),
	)
```

Los repositorios de agenda no se siembran: un profesional sembrado arranca sin horarios, y cargarlos es parte de probar la API.

- [ ] **Step 8: Correr los tests**

Run: `cd apps/api && go test ./... -count=1`
Expected: `ok` en todos los paquetes. Los tests del CRUD que ya existían tienen que seguir pasando: si `NuevoRouter` cambió de firma, hay que actualizar el helper `nuevoServidorDePrueba` de `profesional_test.go` para que le pase también el manejador de agenda.

Run (con cgo): `go test ./... -race -count=1`
Expected: `ok`, sin carreras.

Run: `golangci-lint run ./...`
Expected: `0 issues`.

- [ ] **Step 9: Probar contra el servidor de verdad**

```bash
cd apps/api
APP_ENV=development go run ./cmd/api &
sleep 3
BASE=http://localhost:8080/api/v1/profesionales
ID=$(curl -s "$BASE?busqueda=gonzalez" | python -c "import sys,json; print(json.load(sys.stdin)['datos'][0]['id'])")

curl -s -X PUT "$BASE/$ID/horarios" -H 'Content-Type: application/json' -d '{
  "horarios": [
    {"diaSemana":"lunes","desde":"09:00","hasta":"13:00","duracionMin":50,"modalidad":"telemedicina"}
  ]
}'
echo
curl -s "$BASE/$ID/huecos?desde=2099-09-07&hasta=2099-09-07"
echo
kill %1
```

Expected: el PUT devuelve la semana, y los huecos devuelven cuatro entradas del 7 de septiembre de 2099 —que es lunes— a las 09:00, 09:50, 10:40 y 11:30 con offset `-03:00`.

- [ ] **Step 10: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add apps/api/internal/handler/ apps/api/cmd/api/main.go
git commit -m "feat(handler): rutas de la agenda

Seis rutas colgando del profesional. Los horarios viajan como hora de
reloj y los huecos como instantes con offset, que es la distinción que
sostiene todo el modelo.

El listado de bloqueos sin rango traduce la ausencia de parámetros a
'vigentes y futuros' acá, para que la interfaz del repositorio no
necesite punteros ni centinelas.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Task 9: Verificación final

**Files:** ninguno nuevo. Es la pasada de cierre contra los criterios de aceptación de la sección 11 del spec.

Reportar la salida **real** de cada paso. Si algo falla, decirlo: una verificación que informa un éxito que no observó convierte una incógnita en una certeza falsa.

- [ ] **Step 1: Toda la suite con detector de carreras**

Run (con cgo habilitado): `cd apps/api && go test ./... -race -count=1`
Expected: `ok` en los seis paquetes, sin carreras. El `-count=1` evita resultados cacheados.

- [ ] **Step 2: El linter**

Run: `cd apps/api && golangci-lint run ./...`
Expected: `0 issues.`

- [ ] **Step 3: El dominio sigue sin importar nada del proyecto**

Run: `cd apps/api && go list -deps ./internal/domain | grep "^$(go list -m)"`
Expected: exactamente una línea, el propio paquete `domain`.

- [ ] **Step 4: El contrato**

Run desde la raíz: `npx --yes @redocly/cli@2.47.0 lint apps/api/api/openapi.yaml`
Expected: cero errores.

- [ ] **Step 5: Las seis rutas contra el servidor**

Levantar con `make run` y ejercitar cada una, reportando el código de estado observado:

| Petición | Esperado |
|---|---|
| `GET /profesionales/{id}/horarios` de un profesional sin agenda | `200`, `horarios: []` |
| `PUT /profesionales/{id}/horarios` con una semana válida | `200` con la semana ordenada |
| `PUT` con un bloque de 09:00 a 09:30 y sesiones de 50 min | `422` nombrando `duracionMin` |
| `PUT` con dos bloques del mismo día que se solapan | `422` |
| `PUT` con modalidad que el profesional no ofrece | `422` |
| `POST /profesionales/{id}/bloqueos` a futuro | `201` + `Location` |
| `POST` con un rango en el pasado | `422` |
| `GET /profesionales/{id}/bloqueos` | `200` con el bloqueo |
| `DELETE /profesionales/{id}/bloqueos/{bloqueoId}` | `204` |
| `DELETE` el mismo otra vez | `404` |
| `GET /profesionales/{id}/huecos?desde=&hasta=` | `200` con los huecos y el rango |
| `GET /huecos` sin parámetros | `400` |
| `GET /huecos` de un profesional inexistente | `404` |

- [ ] **Step 6: El hueco del borde, contra el servidor**

Cargar un horario de lunes 09:00 a 13:00 con sesiones de 50 minutos, crear un bloqueo que arranque exactamente a las 09:50 de ese lunes, y pedir los huecos.

Expected: sobrevive el de 09:00 y desaparecen los demás. Es el criterio 4 del spec y el que un `<=` mal puesto rompe en silencio.

- [ ] **Step 7: La imagen arranca con la zona horaria embebida**

Run:
```bash
cd apps/api
docker build -t salud-api:agenda .
docker run --rm -d -p 8080:8080 --name salud-agenda salud-api:agenda
sleep 2
curl -s http://localhost:8080/healthz
docker logs salud-agenda
docker stop salud-agenda
```

Expected: `{"estado":"ok"}` y **ningún pánico en los logs**. Este paso es el que verifica que `time/tzdata` está haciendo su trabajo: la imagen distroless no trae la base de zonas del sistema, así que sin ese import el contenedor muere al arrancar aunque todo funcione en desarrollo.

- [ ] **Step 8: Actualizar el spec si algo derivó**

Leer `docs/superpowers/specs/2026-08-22-disponibilidad-design.md` y contrastar sus afirmaciones con lo observado. Corregir lo que quedó desactualizado. Un spec que le miente a su próximo lector es peor que no tener spec.

- [ ] **Step 9: Commit**

```bash
cd "c:/Users/gianl/Desktop/Salud"
git add -A
git commit -m "chore(agenda): verificación final de los criterios de aceptación

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Lo que queda afuera

Registrado también en la sección 9 del spec.

| Pendiente | Disparador |
|---|---|
| Autenticación | **Antes de `Turno`, y antes de exponer la API.** Esta etapa agranda la deuda: ahora un desconocido no solo puede dar de baja a un profesional, también puede vaciarle la agenda. |
| `Turno` | Después de auth |
| Restar turnos tomados al calcular huecos | Cuando exista `Turno`. `CalculoHuecos` recibe un campo más; la firma pública no cambia. |
| Horarios extra puntuales | Cuando alguien lo pida. Es un concepto distinto de `Bloqueo`. |
| Zona horaria por profesional | Si alguna vez hay profesionales fuera de Argentina |
| Feriados nacionales automáticos | Necesita una fuente de feriados y decidir si se aplican por defecto |
| Duración variable dentro de un bloque | Hoy un profesional con primeras consultas de 80 minutos y seguimientos de 50 necesita dos bloques |
| PostgreSQL | Cuando el modelo deje de moverse |
