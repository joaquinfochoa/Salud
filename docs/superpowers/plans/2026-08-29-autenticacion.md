# Plan de implementación — Identidad y autenticación

> **Para agentes:** SUB-SKILL REQUERIDA: usar `superpowers:subagent-driven-development`
> (recomendado) o `superpowers:executing-plans` para implementar tarea por tarea.
> Los pasos usan checkbox (`- [ ]`) para el seguimiento.

**Objetivo:** que cada `Profesional` tenga un dueño, y que solo ese dueño pueda
editar su perfil, sus horarios y sus bloqueos.

**Arquitectura:** dos entidades nuevas (`Usuario`, `Sesion`) con sus
repositorios en memoria, un servicio de autenticación, y un middleware que
resuelve la cookie de sesión y deja el `UsuarioID` en el `context`. La
autorización se decide **en el servicio**: los handlers le pasan el `UsuarioID`
como parámetro explícito y el servicio compara contra el dueño. Ninguna capa
nueva; se usan las cuatro que ya existen.

**Stack:** Go 1.24, stdlib (`net/mail`, `crypto/rand`, `crypto/sha256`,
`net/http`), `github.com/google/uuid`, y **una dependencia nueva**:
`golang.org/x/crypto/bcrypt`.

**Spec:** [2026-08-29-autenticacion-design.md](../specs/2026-08-29-autenticacion-design.md)
**Base:** rama `refactor-gian`, commit `52c551f`

---

## Restricciones globales

Aplican a **todas** las tareas. No se repiten en cada una.

- **Idioma:** tipos, funciones, campos, constantes, comentarios, mensajes y
  claves JSON en **español**. Paquetes en inglés. `String()` y `Error()` en
  inglés por las interfaces de Go. Las claves `type`/`title`/`status`/`detail`
  de RFC 7807 en inglés.
- **Sin acentos ni eñes en identificadores ni en claves JSON.** Precedente:
  `domain/enums.go`, `DiaSemana` — *"español y sin acentos"*. Por eso
  `Contrasena` y `"contrasena"`, no `Contraseña`. Los **mensajes** de error sí
  llevan acentos.
- **Sin mocks.** Los repositorios en memoria son el doble de test.
- **`internal/domain` no importa nada del proyecto.** El CI lo verifica con
  `go list -deps`. Dependencias externas sí están permitidas (`uuid` ya está
  ahí, `bcrypt` se suma).
- **Dinero en `int64` de centavos.** No aplica a esta etapa pero sigue vigente.
- **Antes de cada commit:** `make check` (fmt + test-race + lint).
- **`-race` en Windows** necesita `CGO_ENABLED=1` y gcc en el PATH. Ver
  `apps/api/README.md`, sección "El detector de carreras en Windows".
- **Módulo:** `github.com/joaquinfochoa/Salud/apps/api`
- Todos los comandos se corren desde `apps/api/`.

---

## Estructura de archivos

**Crear:**

| Archivo | Responsabilidad |
|---|---|
| `internal/domain/email.go` | VO `Email`: parseo, normalización, validación |
| `internal/domain/usuario.go` | Entidad `Usuario`, reglas de contraseña, hashing |
| `internal/domain/sesion.go` | Entidad `Sesion`, generación y hasheo del token |
| `internal/repository/usuario.go` | Interfaz `repository.Usuario` |
| `internal/repository/sesion.go` | Interfaz `repository.Sesion` |
| `internal/repository/memory/usuario.go` | Implementación en memoria |
| `internal/repository/memory/sesion.go` | Implementación en memoria |
| `internal/service/autenticacion.go` | Registro, login, logout, resolución de sesión |
| `internal/handler/autenticacion.go` | Los 4 endpoints nuevos y la cookie |
| `internal/handler/dto_autenticacion.go` | DTOs de entrada y salida |

**Modificar:**

| Archivo | Qué cambia |
|---|---|
| `internal/domain/errores.go` | 4 centinelas nuevos |
| `internal/domain/profesional.go` | Campo `UsuarioID` + `Clonar` |
| `internal/repository/profesional.go` | `ObtenerPorUsuarioID` en la interfaz |
| `internal/repository/memory/profesional.go` | Unicidad de `UsuarioID` en `conflicto` |
| `internal/service/profesional.go` | `usuarioID` en los 4 métodos que mutan |
| `internal/service/agenda.go` | `usuarioID` en los 3 métodos que mutan |
| `internal/handler/middleware.go` | `Autenticar`, `RequerirSesion`, `UsuarioIDDe` |
| `internal/handler/dto.go` | 415 si el `Content-Type` no es JSON |
| `internal/handler/problema.go` | Mapeo de los centinelas nuevos |
| `internal/handler/profesional.go` | Leer el `usuarioID` del contexto |
| `internal/handler/agenda.go` | Idem |
| `internal/handler/router.go` | Rutas nuevas y encadenado de `Autenticar` |
| `api/openapi.yaml` | Contrato: 4 operaciones nuevas y los códigos nuevos |
| `cmd/api/semilla.go` | Sembrar usuarios antes que profesionales |
| `cmd/api/main.go` | Cablear los repos y servicios nuevos |
| `apps/api/README.md` | La convención de dependencias pasa de una a dos |
| `apps/api/.env.example` | Nada nuevo por ahora; verificar que siga correcto |

---

## Task 1: VO `Email` y los centinelas nuevos

**Archivos:**
- Crear: `internal/domain/email.go`
- Crear: `internal/domain/email_test.go`
- Modificar: `internal/domain/errores.go`

**Interfaces:**
- Consume: `ErrorValidacion` de `internal/domain/errores.go`
- Produce: `domain.Email` (tipo `string`), `domain.ParsearEmail(string) (Email, error)`,
  `Email.String() string`, `Email.EsCero() bool`, y los centinelas
  `ErrEmailEnUso`, `ErrNoAutorizado`, `ErrYaTienePerfil`,
  `ErrCredencialesInvalidas`.

- [ ] **Paso 1: escribir el test que falla**

`internal/domain/email_test.go`:

```go
package domain

import "testing"

func TestParsearEmail(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado Email
		falla    bool
	}{
		{"normaliza a minusculas", "Juan@Ejemplo.COM", "juan@ejemplo.com", false},
		{"recorta espacios", "  juan@ejemplo.com  ", "juan@ejemplo.com", false},
		{"acepta subdominio", "juan@mail.ejemplo.com.ar", "juan@mail.ejemplo.com.ar", false},
		{"acepta mas y punto", "juan.perez+turnos@ejemplo.com", "juan.perez+turnos@ejemplo.com", false},
		{"vacio falla", "", "", true},
		{"solo espacios falla", "   ", "", true},
		{"sin arroba falla", "juanejemplo.com", "", true},
		{"sin dominio falla", "juan@", "", true},
		// mail.ParseAddress acepta la forma con nombre visible. Un email de
		// login no es un encabezado de correo: si entra así, el usuario se
		// registra con una dirección distinta a la que escribió.
		{"con nombre visible falla", "Juan Perez <juan@ejemplo.com>", "", true},
		{"con salto de linea falla", "juan@ejemplo.com\n", "", true},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e, err := ParsearEmail(c.entrada)
			if c.falla {
				if err == nil {
					t.Fatalf("ParsearEmail(%q) = %q, se esperaba error", c.entrada, e)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsearEmail(%q) devolvió error: %v", c.entrada, err)
			}
			if e != c.esperado {
				t.Errorf("ParsearEmail(%q) = %q, se esperaba %q", c.entrada, e, c.esperado)
			}
		})
	}
}

func TestEmailDemasiadoLargo(t *testing.T) {
	largo := ""
	for range 250 {
		largo += "a"
	}
	if _, err := ParsearEmail(largo + "@ejemplo.com"); err == nil {
		t.Error("se esperaba error por longitud")
	}
}

func TestEmailEsCero(t *testing.T) {
	var e Email
	if !e.EsCero() {
		t.Error("el Email cero debería ser cero")
	}
	if Email("juan@ejemplo.com").EsCero() {
		t.Error("un email con valor no debería ser cero")
	}
}
```

- [ ] **Paso 2: correr el test y verificar que falla**

```bash
go test ./internal/domain/ -run TestParsearEmail -v
```

Esperado: FAIL — `undefined: ParsearEmail`.

- [ ] **Paso 3: escribir la implementación**

`internal/domain/email.go`:

```go
package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// maxLargoEmail es el máximo del RFC 5321: 64 de parte local + @ + 253 de
// dominio da 318, pero la ruta de retorno completa está acotada a 254. Es el
// número que usan los sistemas reales.
const maxLargoEmail = 254

// Email es una dirección normalizada. Existe como tipo y no como string para
// que "Juan@Ejemplo.COM" y "juan@ejemplo.com" no puedan ser dos usuarios
// distintos: el parser las converge antes de que nada las compare.
type Email string

// ParsearEmail normaliza y valida.
//
// La validación se apoya en net/mail y no en una expresión regular propia. Una
// regex de email escrita a mano rechaza direcciones válidas —y las
// direcciones válidas son mucho más raras de lo que parece— y ese usuario no
// vuelve. Lo que sí se agrega sobre net/mail es rechazar la forma con nombre
// visible: "Juan <juan@ej.com>" es un encabezado de correo legítimo, pero como
// credencial de login registraría al usuario con una dirección distinta a la
// que tipeó.
func ParsearEmail(s string) (Email, error) {
	limpio := strings.ToLower(strings.TrimSpace(s))

	if limpio == "" {
		return "", errors.New("es obligatorio")
	}
	if len(limpio) > maxLargoEmail {
		return "", errors.New("no puede superar los 254 caracteres")
	}

	dir, err := mail.ParseAddress(limpio)
	if err != nil || dir.Address != limpio {
		return "", errors.New("no tiene un formato válido")
	}
	return Email(limpio), nil
}

func (e Email) String() string {
	return string(e)
}

func (e Email) EsCero() bool {
	return e == ""
}
```

- [ ] **Paso 4: agregar los centinelas**

En `internal/domain/errores.go`, dentro del bloque `var (...)` existente,
después de `ErrIDEnUso`:

```go
	// ErrEmailEnUso lo devuelve el repositorio de usuarios al escribir. El
	// email es la identidad de login y no puede repetirse.
	ErrEmailEnUso = errors.New("email ya registrado")

	// ErrNoAutorizado lo devuelven los servicios cuando el usuario de la
	// sesión no es el dueño del recurso. Es distinto de "no hay sesión": eso
	// lo resuelve el middleware con un 401 antes de llegar al servicio.
	ErrNoAutorizado = errors.New("no autorizado")

	// ErrYaTienePerfil lo devuelve el servicio ante un segundo alta de perfil
	// profesional del mismo usuario. Un usuario tiene como máximo uno.
	ErrYaTienePerfil = errors.New("el usuario ya tiene un perfil profesional")

	// ErrCredencialesInvalidas es uno solo para "ese email no existe" y "esa
	// contraseña está mal", a propósito. Distinguirlos convierte al login en
	// un oráculo de qué direcciones están registradas: probando emails contra
	// el endpoint se arma el padrón de usuarios sin adivinar una sola
	// contraseña.
	ErrCredencialesInvalidas = errors.New("credenciales invalidas")
```

- [ ] **Paso 5: correr los tests y verificar que pasan**

```bash
go test ./internal/domain/ -run 'TestParsearEmail|TestEmail' -v
```

Esperado: PASS, 13 subtests.

- [ ] **Paso 6: commit**

```bash
make check
git add internal/domain/email.go internal/domain/email_test.go internal/domain/errores.go
git commit -m "feat(domain): value object Email y centinelas de autenticacion"
```

---

## Task 2: entidad `Usuario`

**Archivos:**
- Crear: `internal/domain/usuario.go`
- Crear: `internal/domain/usuario_test.go`
- Modificar: `go.mod`, `go.sum`

**Interfaces:**
- Consume: `Email`, `ParsearEmail`, `ErrorValidacion`, `validarNombre`
  (ya existe en `internal/domain/profesional.go:236`)
- Produce: `domain.Usuario{ID uuid.UUID, Email Email, Hash []byte, Nombre,
  Apellido string, CreadoEn time.Time}`, `domain.EntradaUsuario{Email,
  Contrasena, Nombre, Apellido string}`,
  `domain.NuevoUsuario(EntradaUsuario, time.Time) (Usuario, error)`,
  `Usuario.VerificarContrasena(string) bool`, `Usuario.Clonar() Usuario`,
  y la variable de paquete `costoBcrypt`.

- [ ] **Paso 1: agregar la dependencia**

```bash
go get golang.org/x/crypto@latest
```

Esperado: `go.mod` gana `require golang.org/x/crypto vX.Y.Z`.

- [ ] **Paso 2: escribir el test que falla**

`internal/domain/usuario_test.go`:

```go
package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// TestMain baja el costo de bcrypt para toda la suite del paquete. Con
// DefaultCost cada hash tarda unos 60 ms y estos tests hacen más de veinte.
func TestMain(m *testing.M) {
	costoBcrypt = bcrypt.MinCost
	m.Run()
}

func entradaUsuarioValida() EntradaUsuario {
	return EntradaUsuario{
		Email:      "juan@ejemplo.com",
		Contrasena: "unaclave8",
		Nombre:     "Juan",
		Apellido:   "Pérez",
	}
}

func TestNuevoUsuario(t *testing.T) {
	ahora := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	u, err := NuevoUsuario(entradaUsuarioValida(), ahora)
	if err != nil {
		t.Fatalf("NuevoUsuario devolvió error: %v", err)
	}
	if u.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("el ID quedó en cero")
	}
	if u.Email != "juan@ejemplo.com" {
		t.Errorf("Email = %q", u.Email)
	}
	if !u.CreadoEn.Equal(ahora) {
		t.Errorf("CreadoEn = %v, se esperaba %v", u.CreadoEn, ahora)
	}
	if len(u.Hash) == 0 {
		t.Fatal("el hash quedó vacío")
	}
	if strings.Contains(string(u.Hash), "unaclave8") {
		t.Error("la contraseña en claro aparece dentro del hash")
	}
}

func TestVerificarContrasena(t *testing.T) {
	u, err := NuevoUsuario(entradaUsuarioValida(), time.Now())
	if err != nil {
		t.Fatalf("NuevoUsuario: %v", err)
	}

	if !u.VerificarContrasena("unaclave8") {
		t.Error("la contraseña correcta no verificó")
	}
	if u.VerificarContrasena("otraclave") {
		t.Error("una contraseña incorrecta verificó")
	}
	if u.VerificarContrasena("") {
		t.Error("la contraseña vacía verificó")
	}
}

// Un usuario de SSO no tiene contraseña. Sin esta guarda,
// bcrypt.CompareHashAndPassword contra un hash vacío devuelve error y el
// resultado sería el correcto por casualidad, no por diseño.
func TestVerificarContrasenaSinHash(t *testing.T) {
	u := Usuario{Email: "juan@ejemplo.com"}
	if u.VerificarContrasena("") {
		t.Error("un usuario sin hash no puede verificar ninguna contraseña")
	}
	if u.VerificarContrasena("loquesea") {
		t.Error("un usuario sin hash no puede verificar ninguna contraseña")
	}
}

func TestNuevoUsuarioValidaciones(t *testing.T) {
	casos := []struct {
		nombre string
		ajuste func(*EntradaUsuario)
		campo  string
	}{
		{"email invalido", func(e *EntradaUsuario) { e.Email = "no-es-un-email" }, "email"},
		{"email vacio", func(e *EntradaUsuario) { e.Email = "" }, "email"},
		{"contrasena corta", func(e *EntradaUsuario) { e.Contrasena = "corta7c" }, "contrasena"},
		{"contrasena vacia", func(e *EntradaUsuario) { e.Contrasena = "" }, "contrasena"},
		{"nombre vacio", func(e *EntradaUsuario) { e.Nombre = "  " }, "nombre"},
		{"apellido vacio", func(e *EntradaUsuario) { e.Apellido = "" }, "apellido"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			entrada := entradaUsuarioValida()
			c.ajuste(&entrada)

			_, err := NuevoUsuario(entrada, time.Now())
			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %v", err)
			}
			for _, campo := range verr.Campos {
				if campo.Campo == c.campo {
					return
				}
			}
			t.Errorf("no se reportó el campo %q; campos = %+v", c.campo, verr.Campos)
		})
	}
}

// Ocho caracteres es el piso exacto. Un test en el borde de abajo y otro justo
// afuera: sin los dos, un >= mal escrito pasa desapercibido.
func TestContrasenaEnElBorde(t *testing.T) {
	entrada := entradaUsuarioValida()

	entrada.Contrasena = "12345678"
	if _, err := NuevoUsuario(entrada, time.Now()); err != nil {
		t.Errorf("8 caracteres debería ser válido: %v", err)
	}

	entrada.Contrasena = "1234567"
	if _, err := NuevoUsuario(entrada, time.Now()); err == nil {
		t.Error("7 caracteres debería fallar")
	}
}

// bcrypt trunca en silencio a partir del byte 72. Truncar significa que
// "<72 bytes>abc" y "<72 bytes>xyz" son la misma contraseña, y el usuario no
// se entera de que la mitad de su clave no cuenta. Se rechaza en bytes, no en
// caracteres: con acentos el byte 72 llega antes.
func TestContrasenaDemasiadoLarga(t *testing.T) {
	entrada := entradaUsuarioValida()

	entrada.Contrasena = strings.Repeat("a", 72)
	if _, err := NuevoUsuario(entrada, time.Now()); err != nil {
		t.Errorf("72 bytes debería ser válido: %v", err)
	}

	entrada.Contrasena = strings.Repeat("a", 73)
	if _, err := NuevoUsuario(entrada, time.Now()); err == nil {
		t.Error("73 bytes debería fallar")
	}

	// 37 eñes son 74 bytes pero solo 37 caracteres: si la guarda contara
	// caracteres, esto pasaría y bcrypt truncaría sin avisar.
	entrada.Contrasena = strings.Repeat("ñ", 37)
	if _, err := NuevoUsuario(entrada, time.Now()); err == nil {
		t.Error("74 bytes en 37 caracteres debería fallar")
	}
}

// Acumular todos los campos inválidos de una pasada, igual que Profesional.
func TestNuevoUsuarioAcumulaErrores(t *testing.T) {
	_, err := NuevoUsuario(EntradaUsuario{}, time.Now())

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %v", err)
	}
	if len(verr.Campos) < 4 {
		t.Errorf("se esperaban 4 campos inválidos, se obtuvieron %d: %+v", len(verr.Campos), verr.Campos)
	}
}

func TestUsuarioClonar(t *testing.T) {
	u, err := NuevoUsuario(entradaUsuarioValida(), time.Now())
	if err != nil {
		t.Fatalf("NuevoUsuario: %v", err)
	}

	c := u.Clonar()
	c.Hash[0] = 'X'

	if u.Hash[0] == 'X' {
		t.Error("mutar el hash del clon alteró el original")
	}
}
```

- [ ] **Paso 3: correr el test y verificar que falla**

```bash
go test ./internal/domain/ -run TestNuevoUsuario -v
```

Esperado: FAIL — `undefined: NuevoUsuario`.

- [ ] **Paso 4: escribir la implementación**

`internal/domain/usuario.go`:

```go
package domain

import (
	"errors"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	minLargoContrasena = 8

	// maxBytesContrasena es el límite duro de bcrypt: descarta todo lo que
	// pase del byte 72 sin devolver error. Sin esta guarda, dos contraseñas
	// largas que comparten los primeros 72 bytes abren la misma cuenta.
	maxBytesContrasena = 72
)

// costoBcrypt es variable y no constante para que los tests puedan bajarlo a
// bcrypt.MinCost. Con DefaultCost cada hash tarda unos 60 ms —que es
// exactamente el punto de bcrypt— y una suite con veinte altas se va a más de
// un segundo sin verificar nada distinto.
//
// ponytail: cost por defecto de la librería. Recalibrar contra el hardware de
// producción el día que haya producción: la recomendación es subirlo hasta que
// un hash tarde ~250 ms en la máquina real.
var costoBcrypt = bcrypt.DefaultCost

// Usuario es la identidad de login. No es el perfil profesional: Profesional
// lo referencia con UsuarioID, y un mismo Usuario puede reservar turnos como
// paciente y atender como profesional.
//
// No hay campo Rol a propósito. Ver la sección 3.1 de la spec: sos profesional
// si existe un Profesional con tu UsuarioID, y un enum almacenado solo agrega
// un estado que puede quedar desincronizado del perfil real.
type Usuario struct {
	ID       uuid.UUID
	Email    Email
	Nombre   string
	Apellido string
	CreadoEn time.Time

	// Hash puede ser nil: un usuario que entra con SSO no tiene contraseña.
	// Hoy ningún constructor produce ese estado —NuevoUsuario exige
	// contraseña— pero el campo ya lo admite para que agregar
	// NuevoUsuarioConGoogle no obligue a tocar el tipo, su validación y sus
	// tests. Ver la sección 3.5 de la spec.
	//
	// El invariante "todo Usuario tiene al menos una forma de autenticarse" lo
	// sostienen los constructores, que son los únicos que arman la entidad.
	Hash []byte
}

// EntradaUsuario es la entrada cruda. Contrasena viaja en claro hasta acá y no
// sale del paquete: NuevoUsuario la hashea y nunca la guarda.
type EntradaUsuario struct {
	Email      string
	Contrasena string
	Nombre     string
	Apellido   string
}

// NuevoUsuario valida y devuelve un usuario consistente, o un ErrorValidacion
// con todos los campos que fallaron.
func NuevoUsuario(entrada EntradaUsuario, ahora time.Time) (Usuario, error) {
	var u Usuario
	var verr ErrorValidacion

	if e, err := ParsearEmail(entrada.Email); err != nil {
		verr.agregar("email", err.Error())
	} else {
		u.Email = e
	}

	if hash, err := hashearContrasena(entrada.Contrasena); err != nil {
		verr.agregar("contrasena", err.Error())
	} else {
		u.Hash = hash
	}

	u.Nombre = validarNombre(entrada.Nombre, "nombre", &verr)
	u.Apellido = validarNombre(entrada.Apellido, "apellido", &verr)

	if verr.tieneErrores() {
		return Usuario{}, verr
	}

	u.ID = uuid.New()
	u.CreadoEn = ahora
	return u, nil
}

// VerificarContrasena compara la contraseña en claro contra el hash.
//
// bcrypt.CompareHashAndPassword compara en tiempo constante, así que no filtra
// por cuánto tarda cuántos bytes del hash coincidían.
func (u Usuario) VerificarContrasena(plana string) bool {
	if len(u.Hash) == 0 {
		return false // usuario de SSO: no hay contraseña contra la cual comparar
	}
	return bcrypt.CompareHashAndPassword(u.Hash, []byte(plana)) == nil
}

// Clonar devuelve una copia profunda. Hash es un slice: sin esto, quien
// reciba la copia puede mutar el hash del original desde afuera.
func (u Usuario) Clonar() Usuario {
	c := u
	c.Hash = slices.Clone(u.Hash)
	return c
}

// hashearContrasena valida las reglas y devuelve el hash.
//
// Sin reglas de composición —mayúscula, número, símbolo—: alargan el
// formulario sin agregar entropía real y empujan a la gente hacia "Password1!".
// El piso de 8 caracteres es el de NIST SP 800-63B.
func hashearContrasena(plana string) ([]byte, error) {
	switch {
	case utf8.RuneCountInString(plana) < minLargoContrasena:
		return nil, errors.New("tiene que tener al menos 8 caracteres")
	case len(plana) > maxBytesContrasena:
		return nil, errors.New("no puede superar los 72 bytes")
	}
	return bcrypt.GenerateFromPassword([]byte(plana), costoBcrypt)
}
```

- [ ] **Paso 5: correr los tests y verificar que pasan**

```bash
go test ./internal/domain/ -v
```

Esperado: PASS. Toda la suite del paquete, no solo la nueva: `TestMain` es
nuevo en el paquete y hay que confirmar que no rompió los tests existentes.

- [ ] **Paso 6: verificar que el dominio sigue sin importar el proyecto**

```bash
go list -deps ./internal/domain/ | grep 'joaquinfochoa' | grep -v 'internal/domain'
```

Esperado: sin salida. Es el mismo chequeo que corre el CI.

- [ ] **Paso 7: commit**

```bash
make check
git add go.mod go.sum internal/domain/usuario.go internal/domain/usuario_test.go
git commit -m "feat(domain): entidad Usuario con hashing bcrypt"
```

---

## Task 3: entidad `Sesion`

**Archivos:**
- Crear: `internal/domain/sesion.go`
- Crear: `internal/domain/sesion_test.go`

**Interfaces:**
- Produce: `domain.Sesion{TokenHash [32]byte, UsuarioID uuid.UUID, CreadaEn,
  ExpiraEn time.Time}`, `domain.DuracionSesion` (constante),
  `domain.NuevaSesion(uuid.UUID, time.Time) (Sesion, string, error)` —el string
  es el token en claro, que solo se devuelve acá y nunca se guarda—,
  `domain.HashearToken(string) [32]byte`,
  `Sesion.EstaVencida(time.Time) bool`.

- [ ] **Paso 1: escribir el test que falla**

`internal/domain/sesion_test.go`:

```go
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNuevaSesion(t *testing.T) {
	usuarioID := uuid.New()
	ahora := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	s, token, err := NuevaSesion(usuarioID, ahora)
	if err != nil {
		t.Fatalf("NuevaSesion devolvió error: %v", err)
	}
	if s.UsuarioID != usuarioID {
		t.Errorf("UsuarioID = %v, se esperaba %v", s.UsuarioID, usuarioID)
	}
	if !s.CreadaEn.Equal(ahora) {
		t.Errorf("CreadaEn = %v", s.CreadaEn)
	}
	if !s.ExpiraEn.Equal(ahora.Add(DuracionSesion)) {
		t.Errorf("ExpiraEn = %v, se esperaba %v", s.ExpiraEn, ahora.Add(DuracionSesion))
	}
	if s.TokenHash != HashearToken(token) {
		t.Error("el hash guardado no corresponde al token devuelto")
	}
	// 32 bytes en base64 sin relleno son 43 caracteres.
	if len(token) != 43 {
		t.Errorf("largo del token = %d, se esperaban 43", len(token))
	}
}

// El token es lo único que prueba la identidad. Dos sesiones con el mismo
// token significan que rand.Read no está haciendo lo que creemos.
func TestTokensDeSesionSonDistintos(t *testing.T) {
	vistos := make(map[string]bool, 100)
	for range 100 {
		_, token, err := NuevaSesion(uuid.New(), time.Now())
		if err != nil {
			t.Fatalf("NuevaSesion: %v", err)
		}
		if vistos[token] {
			t.Fatalf("token repetido: %q", token)
		}
		vistos[token] = true
	}
}

// Se guarda el hash, no el token: quien lea el almacenamiento no puede
// suplantar a nadie. Este test es el que se rompe si alguien "simplifica"
// guardando el token en claro.
func TestElTokenEnClaroNoQuedaEnLaSesion(t *testing.T) {
	s, token, err := NuevaSesion(uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}
	if string(s.TokenHash[:]) == token {
		t.Error("la sesión guarda el token en claro")
	}
}

func TestSesionEstaVencida(t *testing.T) {
	ahora := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	s, _, err := NuevaSesion(uuid.New(), ahora)
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}

	casos := []struct {
		nombre   string
		momento  time.Time
		esperado bool
	}{
		{"recien creada", ahora, false},
		{"un dia despues", ahora.Add(24 * time.Hour), false},
		{"un instante antes de expirar", s.ExpiraEn.Add(-time.Nanosecond), false},
		// El borde es cerrado hacia afuera: en el instante exacto de
		// vencimiento la sesión ya no sirve. La alternativa deja viva una
		// sesión vencida durante un tick del reloj.
		{"en el instante exacto", s.ExpiraEn, true},
		{"un instante despues", s.ExpiraEn.Add(time.Nanosecond), true},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := s.EstaVencida(c.momento); got != c.esperado {
				t.Errorf("EstaVencida(%v) = %v, se esperaba %v", c.momento, got, c.esperado)
			}
		})
	}
}
```

- [ ] **Paso 2: correr el test y verificar que falla**

```bash
go test ./internal/domain/ -run TestNuevaSesion -v
```

Esperado: FAIL — `undefined: NuevaSesion`.

- [ ] **Paso 3: escribir la implementación**

`internal/domain/sesion.go`:

```go
package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// DuracionSesion es cuánto vale una sesión desde que se crea. Es absoluta: no
// se renueva con el uso. La renovación deslizante es más cómoda y se agrega
// después sin cambiar ninguna firma; empezar sin ella es menos código y un
// techo de exposición conocido.
const DuracionSesion = 7 * 24 * time.Hour

// bytesToken son 256 bits de entropía. Un token de sesión es la credencial
// completa: quien lo tiene es el usuario, sin nada más que presentar.
const bytesToken = 32

// Sesion es una autenticación vigente.
//
// Guarda el hash del token, nunca el token. Si alguien lee el almacenamiento
// —un dump, un log, una réplica— se lleva hashes que no sirven para
// autenticarse. Es la misma razón por la que no se guardan contraseñas en
// claro, y cuesta una línea.
type Sesion struct {
	TokenHash [32]byte
	UsuarioID uuid.UUID
	CreadaEn  time.Time
	ExpiraEn  time.Time
}

// NuevaSesion devuelve la sesión a guardar y el token en claro, que es lo
// único que se le manda al cliente. El token no se puede recuperar después:
// esta es la única vez que existe.
func NuevaSesion(usuarioID uuid.UUID, ahora time.Time) (Sesion, string, error) {
	b := make([]byte, bytesToken)
	if _, err := rand.Read(b); err != nil {
		return Sesion{}, "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)

	return Sesion{
		TokenHash: HashearToken(token),
		UsuarioID: usuarioID,
		CreadaEn:  ahora,
		ExpiraEn:  ahora.Add(DuracionSesion),
	}, token, nil
}

// HashearToken convierte un token en la clave con la que se lo busca.
//
// SHA-256 pelado y no bcrypt: acá no hace falta un hash lento. bcrypt existe
// para que una contraseña de baja entropía no se pueda romper por fuerza
// bruta; un token de 256 bits aleatorios no tiene ese problema, y un hash
// lento en cada request sería un costo por nada.
func HashearToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}

// EstaVencida cierra el intervalo hacia afuera: en el instante exacto del
// vencimiento la sesión ya no vale.
func (s Sesion) EstaVencida(ahora time.Time) bool {
	return !ahora.Before(s.ExpiraEn)
}
```

- [ ] **Paso 4: correr los tests y verificar que pasan**

```bash
go test ./internal/domain/ -run 'TestNuevaSesion|TestToken|TestElToken|TestSesion' -v
```

Esperado: PASS, 9 subtests entre todos.

- [ ] **Paso 5: commit**

```bash
make check
git add internal/domain/sesion.go internal/domain/sesion_test.go
git commit -m "feat(domain): entidad Sesion con token opaco hasheado"
```

---

## Task 4: repositorio de usuarios

**Archivos:**
- Crear: `internal/repository/usuario.go`
- Crear: `internal/repository/memory/usuario.go`
- Crear: `internal/repository/memory/usuario_test.go`
- Modificar: `internal/domain/errores.go`

**Interfaces:**
- Consume: `domain.Usuario`, `domain.Email`, `domain.ErrEmailEnUso`,
  `domain.ErrIDEnUso`, `domain.ErrNoEncontrado`
- Produce: interfaz `repository.Usuario` con `Crear`, `ObtenerPorID`,
  `ObtenerPorEmail`; y `memory.NuevoUsuario() *memory.Usuario`

- [ ] **Paso 1: arreglar el mensaje de `ErrNoEncontrado`**

Hoy dice `"profesional no encontrado"` y ya lo comparten los bloqueos —está
registrado como minor en `.superpowers/sdd/progress.md`, Etapa 2 Task 9—.
Con usuarios y sesiones encima, el mensaje pasa a mentir en cuatro lugares.

En `internal/domain/errores.go`:

```go
	// ErrNoEncontrado lo devuelve el repositorio cuando no existe el registro.
	// El mensaje es genérico porque el centinela lo comparten profesionales,
	// bloqueos, usuarios y sesiones: el handler ya arma el texto que ve el
	// cliente según la ruta.
	ErrNoEncontrado = errors.New("recurso no encontrado")
```

```bash
grep -rn "profesional no encontrado" --include='*.go' .
```

Esperado: sin salida. Si aparece alguna aserción en un test, corregirla.

- [ ] **Paso 2: escribir el test que falla**

`internal/repository/memory/usuario_test.go`:

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

func usuarioDePrueba(t *testing.T, email string) domain.Usuario {
	t.Helper()
	u, err := domain.NuevoUsuario(domain.EntradaUsuario{
		Email:      email,
		Contrasena: "unaclave8",
		Nombre:     "Juan",
		Apellido:   "Pérez",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoUsuario: %v", err)
	}
	return u
}

func TestUsuarioCrearYObtener(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()
	u := usuarioDePrueba(t, "juan@ejemplo.com")

	if err := r.Crear(ctx, u); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	porID, err := r.ObtenerPorID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if porID.Email != u.Email {
		t.Errorf("Email = %q, se esperaba %q", porID.Email, u.Email)
	}

	porEmail, err := r.ObtenerPorEmail(ctx, u.Email)
	if err != nil {
		t.Fatalf("ObtenerPorEmail: %v", err)
	}
	if porEmail.ID != u.ID {
		t.Errorf("ID = %v, se esperaba %v", porEmail.ID, u.ID)
	}
}

func TestUsuarioEmailDuplicado(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()

	if err := r.Crear(ctx, usuarioDePrueba(t, "juan@ejemplo.com")); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// Otro usuario, mismo email: es otro ID, así que el mapa no lo detecta.
	// Lo tiene que atrapar el chequeo de unicidad bajo el lock de escritura.
	err := r.Crear(ctx, usuarioDePrueba(t, "juan@ejemplo.com"))
	if !errors.Is(err, domain.ErrEmailEnUso) {
		t.Errorf("se esperaba ErrEmailEnUso, se obtuvo %v", err)
	}
}

func TestUsuarioIDDuplicado(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()
	u := usuarioDePrueba(t, "juan@ejemplo.com")

	if err := r.Crear(ctx, u); err != nil {
		t.Fatalf("Crear: %v", err)
	}
	if err := r.Crear(ctx, u); !errors.Is(err, domain.ErrIDEnUso) {
		t.Errorf("se esperaba ErrIDEnUso, se obtuvo %v", err)
	}
}

func TestUsuarioNoEncontrado(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()

	if _, err := r.ObtenerPorID(ctx, uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("ObtenerPorID: se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
	if _, err := r.ObtenerPorEmail(ctx, "nadie@ejemplo.com"); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("ObtenerPorEmail: se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El repositorio guarda y devuelve copias. Sin esto, quien recibe un usuario
// puede mutarle el hash al que está guardado.
func TestUsuarioDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()
	u := usuarioDePrueba(t, "juan@ejemplo.com")

	if err := r.Crear(ctx, u); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// mutar lo que se le pasó a Crear no tiene que afectar lo guardado
	u.Hash[0] = 'X'

	guardado, err := r.ObtenerPorID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if guardado.Hash[0] == 'X' {
		t.Fatal("Crear guardó el slice del llamador en vez de una copia")
	}

	// y mutar lo devuelto tampoco
	guardado.Hash[0] = 'Y'
	otra, err := r.ObtenerPorID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if otra.Hash[0] == 'Y' {
		t.Error("ObtenerPorID devolvió el slice guardado en vez de una copia")
	}
}
```

- [ ] **Paso 3: correr el test y verificar que falla**

```bash
go test ./internal/repository/memory/ -run TestUsuario -v
```

Esperado: FAIL — `undefined: NuevoUsuario`.

- [ ] **Paso 4: escribir la interfaz**

`internal/repository/usuario.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Usuario es el almacenamiento de identidades de login.
//
// No tiene Actualizar ni Eliminar: nada en esta etapa cambia un usuario ya
// creado. Cambiar el email o la contraseña son casos de uso que todavía no
// existen, y una interfaz con métodos que nadie implementa contra Postgres es
// trabajo por adelantado.
//
// Contrato de unicidad, igual que en Profesional: lo garantiza la escritura,
// no el que llama. En PostgreSQL esto es una constraint UNIQUE sobre email, y
// traducir su violación a ErrEmailEnUso es todo lo que tiene que hacer esa
// implementación.
type Usuario interface {
	Crear(ctx context.Context, u domain.Usuario) error
	ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Usuario, error)
	ObtenerPorEmail(ctx context.Context, e domain.Email) (domain.Usuario, error)
}
```

- [ ] **Paso 5: escribir la implementación en memoria**

`internal/repository/memory/usuario.go`. Sigue el mismo patrón que
`internal/repository/memory/profesional.go:19-70`: aserción de compilación,
`sync.RWMutex`, `conflicto` bajo el lock de escritura, y `Clonar()` en ambos
sentidos.

```go
package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var _ repository.Usuario = (*Usuario)(nil)

type Usuario struct {
	mu    sync.RWMutex
	datos map[uuid.UUID]domain.Usuario
}

func NuevoUsuario() *Usuario {
	return &Usuario{datos: make(map[uuid.UUID]domain.Usuario)}
}

func (r *Usuario) Crear(_ context.Context, u domain.Usuario) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[u.ID]; existe {
		return domain.ErrIDEnUso
	}
	// Chequear la unicidad en una llamada y escribir en otra son dos
	// operaciones, y entre las dos entra otro registro con el mismo email.
	// Bajo un mismo lock son una sola.
	//
	// ponytail: scan O(n) igual que en profesional.go. Un índice por email lo
	// evita, pero hay que mantenerlo sincronizado en cada escritura y para un
	// store de desarrollo no vale.
	for _, otro := range r.datos {
		if otro.Email == u.Email {
			return domain.ErrEmailEnUso
		}
	}

	r.datos[u.ID] = u.Clonar()
	return nil
}

func (r *Usuario) ObtenerPorID(_ context.Context, id uuid.UUID) (domain.Usuario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.datos[id]
	if !ok {
		return domain.Usuario{}, domain.ErrNoEncontrado
	}
	return u.Clonar(), nil
}

func (r *Usuario) ObtenerPorEmail(_ context.Context, e domain.Email) (domain.Usuario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.datos {
		if u.Email == e {
			return u.Clonar(), nil
		}
	}
	return domain.Usuario{}, domain.ErrNoEncontrado
}
```

- [ ] **Paso 6: correr los tests y verificar que pasan**

```bash
go test ./internal/repository/memory/ -run TestUsuario -race -v
go test ./... 
```

Esperado: PASS en los dos. El segundo confirma que cambiar el mensaje de
`ErrNoEncontrado` no rompió nada.

- [ ] **Paso 7: commit**

```bash
make check
git add internal/repository/usuario.go internal/repository/memory/usuario.go \
        internal/repository/memory/usuario_test.go internal/domain/errores.go
git commit -m "feat(repository): repositorio de usuarios en memoria"
```

---

## Task 5: repositorio de sesiones

**Archivos:**
- Crear: `internal/repository/sesion.go`
- Crear: `internal/repository/memory/sesion.go`
- Crear: `internal/repository/memory/sesion_test.go`

**Interfaces:**
- Consume: `domain.Sesion`, `domain.NuevaSesion`, `domain.ErrNoEncontrado`
- Produce: interfaz `repository.Sesion` con
  `Crear(ctx, domain.Sesion) error`,
  `ObtenerPorTokenHash(ctx, [32]byte) (domain.Sesion, error)`,
  `Eliminar(ctx, [32]byte) error`; y `memory.NuevaSesion() *memory.Sesion`

- [ ] **Paso 1: escribir el test que falla**

`internal/repository/memory/sesion_test.go`:

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

func TestSesionCrearYObtener(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	s, token, err := domain.NuevaSesion(uuid.New(), time.Now().UTC())
	if err != nil {
		t.Fatalf("domain.NuevaSesion: %v", err)
	}
	if err := r.Crear(ctx, s); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	obtenida, err := r.ObtenerPorTokenHash(ctx, domain.HashearToken(token))
	if err != nil {
		t.Fatalf("ObtenerPorTokenHash: %v", err)
	}
	if obtenida.UsuarioID != s.UsuarioID {
		t.Errorf("UsuarioID = %v, se esperaba %v", obtenida.UsuarioID, s.UsuarioID)
	}
}

func TestSesionEliminar(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	s, token, err := domain.NuevaSesion(uuid.New(), time.Now().UTC())
	if err != nil {
		t.Fatalf("domain.NuevaSesion: %v", err)
	}
	if err := r.Crear(ctx, s); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	if err := r.Eliminar(ctx, domain.HashearToken(token)); err != nil {
		t.Fatalf("Eliminar: %v", err)
	}
	if _, err := r.ObtenerPorTokenHash(ctx, domain.HashearToken(token)); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("después de eliminar se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El logout tiene que poder repetirse sin explotar: un cliente que reintenta
// un DELETE no está haciendo nada malo.
func TestSesionEliminarEsIdempotente(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	if err := r.Eliminar(ctx, domain.HashearToken("no-existe")); err != nil {
		t.Errorf("eliminar una sesión inexistente devolvió %v, se esperaba nil", err)
	}
}

func TestSesionTokenDesconocido(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	if _, err := r.ObtenerPorTokenHash(ctx, domain.HashearToken("inventado")); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}
```

- [ ] **Paso 2: correr el test y verificar que falla**

```bash
go test ./internal/repository/memory/ -run TestSesion -v
```

Esperado: FAIL — `undefined: NuevaSesion` en el paquete `memory`.

- [ ] **Paso 3: escribir la interfaz**

`internal/repository/sesion.go`:

```go
package repository

import (
	"context"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Sesion es el almacenamiento de sesiones vigentes.
//
// La clave es el hash del token, no un ID propio: la única forma de llegar a
// una sesión es presentando el token, y guardar además un ID sería una
// segunda llave que nadie usa.
//
// No hay listado ni borrado masivo. "Cerrar todas mis sesiones" y la limpieza
// de vencidas son casos que todavía no existen; con PostgreSQL la limpieza es
// un DELETE por expira_en y no necesita pasar por esta interfaz.
type Sesion interface {
	Crear(ctx context.Context, s domain.Sesion) error

	// ObtenerPorTokenHash devuelve la sesión sin mirar si venció. Decidir eso
	// necesita un reloj, y el reloj vive en el servicio: un repositorio que
	// filtra por tiempo es un repositorio imposible de testear con fechas
	// fijas.
	ObtenerPorTokenHash(ctx context.Context, hash [32]byte) (domain.Sesion, error)

	// Eliminar es idempotente: borrar una sesión que no existe no es un error.
	Eliminar(ctx context.Context, hash [32]byte) error
}
```

- [ ] **Paso 4: escribir la implementación en memoria**

`internal/repository/memory/sesion.go`:

```go
package memory

import (
	"context"
	"sync"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var _ repository.Sesion = (*Sesion)(nil)

type Sesion struct {
	mu    sync.RWMutex
	datos map[[32]byte]domain.Sesion
}

func NuevaSesion() *Sesion {
	return &Sesion{datos: make(map[[32]byte]domain.Sesion)}
}

// Crear no chequea duplicados: la clave es un hash de 256 bits aleatorios, así
// que una colisión es tan improbable como adivinar el token. Si el mismo hash
// llegara dos veces, sobrescribir es lo correcto.
//
// domain.Sesion no tiene slices ni punteros, así que la asignación ya es una
// copia completa y no hace falta Clonar().
func (r *Sesion) Crear(_ context.Context, s domain.Sesion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.datos[s.TokenHash] = s
	return nil
}

func (r *Sesion) ObtenerPorTokenHash(_ context.Context, hash [32]byte) (domain.Sesion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.datos[hash]
	if !ok {
		return domain.Sesion{}, domain.ErrNoEncontrado
	}
	return s, nil
}

func (r *Sesion) Eliminar(_ context.Context, hash [32]byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.datos, hash)
	return nil
}
```

- [ ] **Paso 5: correr los tests y verificar que pasan**

```bash
go test ./internal/repository/memory/ -run TestSesion -race -v
```

Esperado: PASS, 4 tests.

- [ ] **Paso 6: commit**

```bash
make check
git add internal/repository/sesion.go internal/repository/memory/sesion.go \
        internal/repository/memory/sesion_test.go
git commit -m "feat(repository): repositorio de sesiones en memoria"
```

---

## Task 6: servicio de autenticación

**Archivos:**
- Crear: `internal/service/autenticacion.go`
- Crear: `internal/service/autenticacion_test.go`

**Interfaces:**
- Consume: `repository.Usuario`, `repository.Sesion`, `domain.NuevoUsuario`,
  `domain.NuevaSesion`, `domain.HashearToken`, `domain.ErrCredencialesInvalidas`
- Produce:
  - `service.NuevaAutenticacion(repository.Usuario, repository.Sesion, ...func(*Autenticacion)) *Autenticacion`
  - `service.ConRelojAuth(func() time.Time) func(*Autenticacion)`
  - `(*Autenticacion).Registrar(ctx, domain.EntradaUsuario) (domain.Usuario, string, error)`
  - `(*Autenticacion).IniciarSesion(ctx, email, contrasena string) (domain.Usuario, string, error)`
  - `(*Autenticacion).CerrarSesion(ctx, token string) error`
  - `(*Autenticacion).ResolverSesion(ctx, token string) (domain.Usuario, error)`

  En las dos primeras, el `string` devuelto es el **token en claro**: es la
  única vez que existe y el handler lo pone en la cookie.

- [ ] **Paso 1: escribir el test que falla**

`internal/service/autenticacion_test.go`:

```go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

func nuevaAuthDePrueba() *Autenticacion {
	return NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion())
}

func entradaAuthValida() domain.EntradaUsuario {
	return domain.EntradaUsuario{
		Email:      "juan@ejemplo.com",
		Contrasena: "unaclave8",
		Nombre:     "Juan",
		Apellido:   "Pérez",
	}
}

func TestRegistrarDevuelveSesion(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	u, token, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	if token == "" {
		t.Fatal("Registrar no devolvió token: registrarse tiene que loguear")
	}

	resuelto, err := auth.ResolverSesion(ctx, token)
	if err != nil {
		t.Fatalf("ResolverSesion: %v", err)
	}
	if resuelto.ID != u.ID {
		t.Errorf("ID = %v, se esperaba %v", resuelto.ID, u.ID)
	}
}

func TestRegistrarEmailDuplicado(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, _, err := auth.Registrar(ctx, entradaAuthValida()); err != nil {
		t.Fatalf("primer Registrar: %v", err)
	}

	// mismo email escrito distinto: el VO Email lo normaliza, así que tiene
	// que chocar igual
	entrada := entradaAuthValida()
	entrada.Email = "JUAN@Ejemplo.com"

	_, _, err := auth.Registrar(ctx, entrada)
	if !errors.Is(err, domain.ErrEmailEnUso) {
		t.Errorf("se esperaba ErrEmailEnUso, se obtuvo %v", err)
	}
}

func TestIniciarSesion(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	u, _, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	mismo, token, err := auth.IniciarSesion(ctx, "juan@ejemplo.com", "unaclave8")
	if err != nil {
		t.Fatalf("IniciarSesion: %v", err)
	}
	if mismo.ID != u.ID {
		t.Errorf("ID = %v, se esperaba %v", mismo.ID, u.ID)
	}
	if token == "" {
		t.Error("IniciarSesion no devolvió token")
	}
}

// Un solo error para las dos formas de fallar. Si alguna vez se separan, el
// login se convierte en un oráculo de qué emails están registrados.
func TestIniciarSesionFallaIgualPorAmbosLados(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, _, err := auth.Registrar(ctx, entradaAuthValida()); err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	casos := []struct {
		nombre     string
		email      string
		contrasena string
	}{
		{"email inexistente", "otro@ejemplo.com", "unaclave8"},
		{"contrasena incorrecta", "juan@ejemplo.com", "otraclave"},
		{"email mal formado", "no-es-un-email", "unaclave8"},
		{"contrasena vacia", "juan@ejemplo.com", ""},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, _, err := auth.IniciarSesion(ctx, c.email, c.contrasena)
			if !errors.Is(err, domain.ErrCredencialesInvalidas) {
				t.Errorf("se esperaba ErrCredencialesInvalidas, se obtuvo %v", err)
			}
		})
	}
}

func TestIniciarSesionNormalizaElEmail(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, _, err := auth.Registrar(ctx, entradaAuthValida()); err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	if _, _, err := auth.IniciarSesion(ctx, "  JUAN@Ejemplo.COM ", "unaclave8"); err != nil {
		t.Errorf("el login debería aceptar el email escrito distinto: %v", err)
	}
}

func TestCerrarSesion(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	_, token, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	if err := auth.CerrarSesion(ctx, token); err != nil {
		t.Fatalf("CerrarSesion: %v", err)
	}
	if _, err := auth.ResolverSesion(ctx, token); err == nil {
		t.Error("la sesión sigue viva después del logout")
	}
}

func TestCerrarSesionEsIdempotente(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if err := auth.CerrarSesion(ctx, "token-que-no-existe"); err != nil {
		t.Errorf("cerrar una sesión inexistente devolvió %v, se esperaba nil", err)
	}
}

// El criterio de aceptación 6: una sesión vencida deja de valer sola, sin que
// nadie la haya borrado.
func TestResolverSesionVencida(t *testing.T) {
	ctx := context.Background()
	reloj := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	auth := NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion(),
		ConRelojAuth(func() time.Time { return reloj }))

	_, token, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	// justo antes de vencer sigue viva
	reloj = reloj.Add(domain.DuracionSesion - time.Second)
	if _, err := auth.ResolverSesion(ctx, token); err != nil {
		t.Fatalf("la sesión no debería haber vencido todavía: %v", err)
	}

	// pasado el vencimiento, no
	reloj = reloj.Add(2 * time.Second)
	if _, err := auth.ResolverSesion(ctx, token); err == nil {
		t.Error("una sesión vencida siguió resolviendo")
	}
}

func TestResolverSesionConTokenInventado(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, err := auth.ResolverSesion(ctx, "inventado"); err == nil {
		t.Error("un token inventado resolvió una sesión")
	}
	if _, err := auth.ResolverSesion(ctx, ""); err == nil {
		t.Error("un token vacío resolvió una sesión")
	}
}
```

- [ ] **Paso 2: correr el test y verificar que falla**

```bash
go test ./internal/service/ -run TestRegistrar -v
```

Esperado: FAIL — `undefined: NuevaAutenticacion`.

- [ ] **Paso 3: escribir la implementación**

`internal/service/autenticacion.go`:

```go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// Autenticacion resuelve el registro, el login y la vida de las sesiones. No
// sabe nada de HTTP ni de cookies: devuelve el token en claro y el handler
// decide cómo transportarlo.
type Autenticacion struct {
	usuarios repository.Usuario
	sesiones repository.Sesion
	ahora    func() time.Time
}

// ConRelojAuth inyecta el reloj. Se llama así y no ConReloj porque ese nombre
// ya lo usa el servicio de agenda en este mismo paquete.
func ConRelojAuth(ahora func() time.Time) func(*Autenticacion) {
	return func(a *Autenticacion) { a.ahora = ahora }
}

func NuevaAutenticacion(usuarios repository.Usuario, sesiones repository.Sesion, opciones ...func(*Autenticacion)) *Autenticacion {
	a := &Autenticacion{
		usuarios: usuarios,
		sesiones: sesiones,
		ahora:    func() time.Time { return time.Now().UTC() },
	}
	for _, o := range opciones {
		o(a)
	}
	return a
}

// Registrar crea el usuario y le abre una sesión de una. Pedirle que se
// loguee inmediatamente después de registrarse es un paso que no informa nada.
func (a *Autenticacion) Registrar(ctx context.Context, entrada domain.EntradaUsuario) (domain.Usuario, string, error) {
	u, err := domain.NuevoUsuario(entrada, a.ahora())
	if err != nil {
		return domain.Usuario{}, "", err
	}
	if err := a.usuarios.Crear(ctx, u); err != nil {
		return domain.Usuario{}, "", err
	}

	token, err := a.abrirSesion(ctx, u)
	if err != nil {
		return domain.Usuario{}, "", err
	}
	return u, token, nil
}

// IniciarSesion valida las credenciales y abre una sesión.
//
// Todos los caminos de fallo devuelven ErrCredencialesInvalidas. Distinguir
// "ese email no existe" de "esa contraseña está mal" convierte al login en un
// oráculo: probando direcciones se arma el padrón de usuarios sin adivinar una
// sola contraseña.
//
// ponytail: queda un canal lateral por tiempo. Cuando el email no existe se
// vuelve sin llamar a bcrypt, así que la respuesta llega bastante antes que
// cuando sí existe. Cerrarlo es hashear contra un hash de relleno; no se hace
// todavía porque explotarlo requiere muchos intentos medidos, que es
// exactamente lo que frena el rate limiting del login —ya anotado en la spec
// como requisito previo a exponer la API—. Si el rate limiting se posterga,
// esto se construye.
func (a *Autenticacion) IniciarSesion(ctx context.Context, email, contrasena string) (domain.Usuario, string, error) {
	e, err := domain.ParsearEmail(email)
	if err != nil {
		return domain.Usuario{}, "", domain.ErrCredencialesInvalidas
	}

	u, err := a.usuarios.ObtenerPorEmail(ctx, e)
	switch {
	case errors.Is(err, domain.ErrNoEncontrado):
		return domain.Usuario{}, "", domain.ErrCredencialesInvalidas
	case err != nil:
		return domain.Usuario{}, "", err
	}

	if !u.VerificarContrasena(contrasena) {
		return domain.Usuario{}, "", domain.ErrCredencialesInvalidas
	}

	token, err := a.abrirSesion(ctx, u)
	if err != nil {
		return domain.Usuario{}, "", err
	}
	return u, token, nil
}

// CerrarSesion borra la sesión. Es idempotente: reintentar un logout no es un
// error, y un token inventado tampoco.
func (a *Autenticacion) CerrarSesion(ctx context.Context, token string) error {
	return a.sesiones.Eliminar(ctx, domain.HashearToken(token))
}

// ResolverSesion devuelve el usuario detrás de un token, o ErrNoEncontrado si
// el token no existe o la sesión venció.
//
// La sesión vencida se borra al detectarla: es la única limpieza que hay, y
// alcanza porque una sesión a la que nadie vuelve tampoco molesta a nadie.
func (a *Autenticacion) ResolverSesion(ctx context.Context, token string) (domain.Usuario, error) {
	if token == "" {
		return domain.Usuario{}, domain.ErrNoEncontrado
	}

	hash := domain.HashearToken(token)
	s, err := a.sesiones.ObtenerPorTokenHash(ctx, hash)
	if err != nil {
		return domain.Usuario{}, err
	}

	if s.EstaVencida(a.ahora()) {
		if err := a.sesiones.Eliminar(ctx, hash); err != nil {
			return domain.Usuario{}, err
		}
		return domain.Usuario{}, domain.ErrNoEncontrado
	}

	return a.usuarios.ObtenerPorID(ctx, s.UsuarioID)
}

func (a *Autenticacion) abrirSesion(ctx context.Context, u domain.Usuario) (string, error) {
	s, token, err := domain.NuevaSesion(u.ID, a.ahora())
	if err != nil {
		return "", err
	}
	if err := a.sesiones.Crear(ctx, s); err != nil {
		return "", err
	}
	return token, nil
}
```

- [ ] **Paso 4: correr los tests y verificar que pasan**

```bash
go test ./internal/service/ -race -v
```

Esperado: PASS. Toda la suite del paquete, no solo la nueva.

- [ ] **Paso 5: commit**

```bash
make check
git add internal/service/autenticacion.go internal/service/autenticacion_test.go
git commit -m "feat(service): registro, login, logout y resolucion de sesion"
```

---

## Task 7: `Profesional` tiene dueño

Es la tarea que rompe más código existente. `NuevoProfesional` cambia de firma,
así que **todos** los tests del dominio, del repositorio y del servicio que la
llaman dejan de compilar. Es esperado y forma parte de la tarea.

**Archivos:**
- Modificar: `internal/domain/profesional.go`, `internal/domain/profesional_test.go`
- Modificar: `internal/repository/profesional.go`
- Modificar: `internal/repository/memory/profesional.go`, `internal/repository/memory/profesional_test.go`

**Interfaces:**
- Produce: campo `Profesional.UsuarioID uuid.UUID`; firma nueva
  `domain.NuevoProfesional(entrada EntradaProfesional, usuarioID uuid.UUID, ahora time.Time) (Profesional, error)`;
  método nuevo en la interfaz
  `repository.Profesional.ObtenerPorUsuarioID(ctx, uuid.UUID) (domain.Profesional, error)`;
  y `domain.ErrUsuarioEnUso` para la unicidad.

- [ ] **Paso 1: escribir los tests que fallan**

Agregar a `internal/domain/profesional_test.go`:

```go
func TestNuevoProfesionalGuardaElDueno(t *testing.T) {
	usuarioID := uuid.New()

	p, err := NuevoProfesional(entradaValida(), usuarioID, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoProfesional: %v", err)
	}
	if p.UsuarioID != usuarioID {
		t.Errorf("UsuarioID = %v, se esperaba %v", p.UsuarioID, usuarioID)
	}
}

// El dueño no se edita: viene de la sesión al crear y no vuelve a mirarse.
// Sin esta preservación, cualquier PUT dejaría el perfil sin dueño —el cero de
// uuid.UUID— y a partir de ahí nadie podría volver a editarlo.
func TestAplicarCambiosPreservaElDueno(t *testing.T) {
	usuarioID := uuid.New()
	p, err := NuevoProfesional(entradaValida(), usuarioID, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoProfesional: %v", err)
	}

	entrada := entradaValida()
	entrada.Bio = "otra bio"

	actualizado, err := p.AplicarCambios(entrada, time.Now().UTC())
	if err != nil {
		t.Fatalf("AplicarCambios: %v", err)
	}
	if actualizado.UsuarioID != usuarioID {
		t.Errorf("UsuarioID = %v, se esperaba %v", actualizado.UsuarioID, usuarioID)
	}
}
```

> `entradaValida()` ya existe en `internal/domain/profesional_test.go`. Si el
> nombre del helper difiere, usar el que esté.

Agregar a `internal/repository/memory/profesional_test.go`:

```go
// Un usuario tiene como máximo un perfil profesional. Lo garantiza la
// escritura, igual que la matrícula y el slug: el chequeo del servicio lee y
// suelta el lock, y entre las dos operaciones entra otro alta.
func TestProfesionalUsuarioDuplicado(t *testing.T) {
	ctx := context.Background()
	r := NuevoProfesional()
	usuarioID := uuid.New()

	primero, err := domain.NuevoProfesional(entradaValidaRepo(), usuarioID, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoProfesional: %v", err)
	}
	if err := r.Crear(ctx, primero); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// otra matrícula, otro nombre, mismo dueño
	entrada := entradaValidaRepo()
	entrada.Matricula = "MN 777888"
	entrada.Nombre = "Otra"
	segundo, err := domain.NuevoProfesional(entrada, usuarioID, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoProfesional: %v", err)
	}

	if err := r.Crear(ctx, segundo); !errors.Is(err, domain.ErrUsuarioEnUso) {
		t.Errorf("se esperaba ErrUsuarioEnUso, se obtuvo %v", err)
	}
}

func TestObtenerPorUsuarioID(t *testing.T) {
	ctx := context.Background()
	r := NuevoProfesional()
	usuarioID := uuid.New()

	p, err := domain.NuevoProfesional(entradaValidaRepo(), usuarioID, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoProfesional: %v", err)
	}
	if err := r.Crear(ctx, p); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	encontrado, err := r.ObtenerPorUsuarioID(ctx, usuarioID)
	if err != nil {
		t.Fatalf("ObtenerPorUsuarioID: %v", err)
	}
	if encontrado.ID != p.ID {
		t.Errorf("ID = %v, se esperaba %v", encontrado.ID, p.ID)
	}

	if _, err := r.ObtenerPorUsuarioID(ctx, uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}
```

- [ ] **Paso 2: correr y verificar que no compila**

```bash
go build ./...
```

Esperado: errores de compilación por la cantidad de argumentos de
`NuevoProfesional`. Es la señal de que la firma cambió de verdad.

- [ ] **Paso 3: agregar el campo y el centinela**

En `internal/domain/profesional.go`, dentro del struct `Profesional`, después
de `ID`:

```go
	// UsuarioID es el dueño. Obligatorio: un Profesional sin dueño es un
	// perfil que cualquiera puede editar, que es exactamente lo que esta
	// etapa vino a arreglar.
	//
	// No está en EntradaProfesional a propósito. Sale de la sesión, nunca del
	// body: aceptarlo del cliente sería dejar que cualquiera se declare dueño
	// del perfil de otro.
	UsuarioID uuid.UUID
```

En `internal/domain/errores.go`, junto a los otros:

```go
	// ErrUsuarioEnUso lo devuelve el repositorio al escribir: un usuario tiene
	// como máximo un perfil profesional.
	ErrUsuarioEnUso = errors.New("usuario ya asociado a otro perfil")
```

> `ErrYaTienePerfil` (Task 1) es el del servicio y `ErrUsuarioEnUso` el del
> repositorio, igual que la matrícula tiene el chequeo previo del servicio y la
> garantía de la escritura. El handler los mapea a la misma respuesta.

En `NuevoProfesional`, cambiar la firma y asignar el campo:

```go
func NuevoProfesional(entrada EntradaProfesional, usuarioID uuid.UUID, ahora time.Time) (Profesional, error) {
	p, verr := construir(entrada)
	if verr.tieneErrores() {
		return Profesional{}, verr
	}

	p.ID = uuid.New()
	p.UsuarioID = usuarioID
	// ... el resto queda igual
```

En `AplicarCambios`, junto a los otros campos no editables:

```go
	actualizado.UsuarioID = p.UsuarioID
```

- [ ] **Paso 4: agregar `ObtenerPorUsuarioID` a la interfaz**

En `internal/repository/profesional.go`, dentro de la interfaz:

```go
	ObtenerPorUsuarioID(ctx context.Context, usuarioID uuid.UUID) (domain.Profesional, error)
```

Y en el comentario de contrato de unicidad, agregar `UsuarioID` a la lista:
`Crear` ahora devuelve también `ErrUsuarioEnUso`.

- [ ] **Paso 5: implementar en memoria**

En `internal/repository/memory/profesional.go`, dentro de `conflicto`, después
del chequeo de slug:

```go
		if otro.UsuarioID == p.UsuarioID {
			return domain.ErrUsuarioEnUso
		}
```

Y el método nuevo, con el mismo patrón que `ObtenerPorMatricula`:

```go
func (r *Profesional) ObtenerPorUsuarioID(_ context.Context, usuarioID uuid.UUID) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.datos {
		if p.UsuarioID == usuarioID {
			return p.Clonar(), nil
		}
	}
	return domain.Profesional{}, domain.ErrNoEncontrado
}
```

- [ ] **Paso 6: arreglar todas las llamadas rotas**

```bash
go build ./... 2>&1 | grep NuevoProfesional
```

Cada llamada a `domain.NuevoProfesional` en tests necesita un `uuid.New()` como
segundo argumento. En los tests que no verifican propiedad, un UUID nuevo por
llamada es lo correcto: mantiene la unicidad de dueño y no acopla los casos.

- [ ] **Paso 7: correr toda la suite**

```bash
go test ./... -race
```

Esperado: PASS. `internal/service` puede seguir roto — lo arregla la Task 8. Si
es así, correr por paquete y dejar `service` para la tarea siguiente.

- [ ] **Paso 8: commit**

```bash
make check
git add internal/domain/ internal/repository/
git commit -m "feat(domain): Profesional tiene dueño obligatorio"
```

---

## Task 8: autorización en el servicio de profesionales

**Archivos:**
- Modificar: `internal/service/profesional.go`, `internal/service/profesional_test.go`

**Interfaces:**
- Firmas nuevas, todas con el `usuarioID` **primero** después del contexto,
  para que quede visible que es una operación autenticada:
  - `Crear(ctx, usuarioID uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error)`
  - `Actualizar(ctx, usuarioID, id uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error)`
  - `DarDeBaja(ctx, usuarioID, id uuid.UUID) error`
  - `Reactivar(ctx, usuarioID, id uuid.UUID) (domain.Profesional, error)`
- Las lecturas (`ObtenerPorID`, `ObtenerPorSlug`, `Listar`) **no cambian**: son
  públicas.

- [ ] **Paso 1: escribir los tests que fallan**

Agregar a `internal/service/profesional_test.go`:

```go
// Los cuatro métodos que mutan rechazan a un usuario que no es el dueño. Una
// tabla y no cuatro tests sueltos: si mañana se agrega un quinto método que
// muta, agregar la fila es más difícil de olvidar que escribir el test.
func TestSoloElDuenoPuedeMutar(t *testing.T) {
	ctx := context.Background()
	svc := NuevoProfesional(memory.NuevoProfesional())

	dueno := uuid.New()
	intruso := uuid.New()

	p, err := svc.Crear(ctx, dueno, entradaValidaSvc())
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	casos := []struct {
		nombre string
		correr func(usuarioID uuid.UUID) error
	}{
		{"Actualizar", func(u uuid.UUID) error {
			_, err := svc.Actualizar(ctx, u, p.ID, entradaValidaSvc())
			return err
		}},
		{"DarDeBaja", func(u uuid.UUID) error {
			return svc.DarDeBaja(ctx, u, p.ID)
		}},
		{"Reactivar", func(u uuid.UUID) error {
			_, err := svc.Reactivar(ctx, u, p.ID)
			return err
		}},
	}

	for _, c := range casos {
		t.Run(c.nombre+" con intruso", func(t *testing.T) {
			if err := c.correr(intruso); !errors.Is(err, domain.ErrNoAutorizado) {
				t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
			}
		})
		t.Run(c.nombre+" con dueño", func(t *testing.T) {
			if err := c.correr(dueno); err != nil {
				t.Errorf("el dueño no debería recibir error: %v", err)
			}
		})
	}
}

// Un perfil inexistente da ErrNoEncontrado y no ErrNoAutorizado, incluso para
// un intruso. Al revés sería un oráculo: probando IDs, un 403 confirmaría que
// ese perfil existe.
func TestPerfilInexistenteDaNoEncontrado(t *testing.T) {
	ctx := context.Background()
	svc := NuevoProfesional(memory.NuevoProfesional())

	err := svc.DarDeBaja(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestUnUsuarioNoPuedeTenerDosPerfiles(t *testing.T) {
	ctx := context.Background()
	svc := NuevoProfesional(memory.NuevoProfesional())
	usuarioID := uuid.New()

	if _, err := svc.Crear(ctx, usuarioID, entradaValidaSvc()); err != nil {
		t.Fatalf("primer Crear: %v", err)
	}

	otra := entradaValidaSvc()
	otra.Matricula = "MN 777888"

	_, err := svc.Crear(ctx, usuarioID, otra)
	if !errors.Is(err, domain.ErrYaTienePerfil) {
		t.Errorf("se esperaba ErrYaTienePerfil, se obtuvo %v", err)
	}
}
```

> `entradaValidaSvc()` es el helper que ya usa este archivo. Si tiene otro
> nombre, usar el existente.

- [ ] **Paso 2: correr y verificar que falla**

```bash
go test ./internal/service/ -run TestSoloElDueno -v
```

Esperado: FAIL de compilación por la cantidad de argumentos.

- [ ] **Paso 3: implementar**

En `internal/service/profesional.go`, agregar el helper:

```go
// verificarPropiedad es la única implementación de "este perfil es tuyo".
//
// Vive en el servicio y no en el handler a propósito: acá la cubren los tests
// sin levantar un servidor HTTP, y cualquier consumidor futuro —un comando de
// consola, una cola— pasa por la misma regla en vez de reimplementarla.
func verificarPropiedad(p domain.Profesional, usuarioID uuid.UUID) error {
	if p.UsuarioID != usuarioID {
		return domain.ErrNoAutorizado
	}
	return nil
}
```

`Crear` gana el `usuarioID` y el chequeo de perfil único:

```go
func (s *Profesional) Crear(ctx context.Context, usuarioID uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	// Camino rápido, con la misma salvedad que el de la matrícula: lee y
	// suelta el lock, así que la garantía la da repo.Crear con ErrUsuarioEnUso.
	switch _, err := s.repo.ObtenerPorUsuarioID(ctx, usuarioID); {
	case err == nil:
		return domain.Profesional{}, domain.ErrYaTienePerfil
	case !errors.Is(err, domain.ErrNoEncontrado):
		return domain.Profesional{}, err
	}

	p, err := domain.NuevoProfesional(entrada, usuarioID, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}
	// ... el resto del método queda igual
```

Los otros tres siguen el mismo molde: obtener, verificar propiedad, seguir.

```go
func (s *Profesional) Actualizar(ctx context.Context, usuarioID, id uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}
	if err := verificarPropiedad(actual, usuarioID); err != nil {
		return domain.Profesional{}, err
	}
	// ... el resto queda igual
```

```go
func (s *Profesional) DarDeBaja(ctx context.Context, usuarioID, id uuid.UUID) error {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}
	if err := verificarPropiedad(actual, usuarioID); err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, actual.DarDeBaja(s.ahora()))
}
```

```go
func (s *Profesional) Reactivar(ctx context.Context, usuarioID, id uuid.UUID) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}
	if err := verificarPropiedad(actual, usuarioID); err != nil {
		return domain.Profesional{}, err
	}
	// ... el resto queda igual
```

- [ ] **Paso 4: correr los tests y verificar que pasan**

```bash
go test ./internal/service/ -race -v
```

Esperado: PASS, incluidos los 6 subtests de `TestSoloElDuenoPuedeMutar`.

- [ ] **Paso 5: mutation testing del chequeo**

Cambiar temporalmente `verificarPropiedad` para que devuelva siempre `nil`.

```bash
go test ./internal/service/ -run TestSoloElDueno
```

Esperado: FAIL en los tres subtests "con intruso". Revertir el cambio.

Sin este paso no hay forma de saber si los tests prueban la autorización o
simplemente pasan.

- [ ] **Paso 6: commit**

```bash
make check
git add internal/service/profesional.go internal/service/profesional_test.go
git commit -m "feat(service): solo el dueño puede editar su perfil"
```

---

## Task 9: autorización en el servicio de agenda

Mismo molde que la Task 8, sobre los tres métodos que mutan. Las tres lecturas
—`ListarHorarios`, `ListarBloqueos`, `HuecosLibres`— **no cambian**: la agenda
disponible es pública y el paciente la mira sin estar logueado.

**Archivos:**
- Modificar: `internal/service/agenda.go`, `internal/service/agenda_test.go`

**Interfaces:**
- `ReemplazarHorarios(ctx, usuarioID, profesionalID uuid.UUID, entradas []domain.EntradaHorarioSemanal) ([]domain.HorarioSemanal, error)`
- `CrearBloqueo(ctx, usuarioID, profesionalID uuid.UUID, entrada domain.EntradaBloqueo) (domain.Bloqueo, error)`
- `EliminarBloqueo(ctx, usuarioID, profesionalID, bloqueoID uuid.UUID) error`

- [ ] **Paso 1: escribir el test que falla**

Agregar a `internal/service/agenda_test.go`:

```go
func TestSoloElDuenoPuedeTocarLaAgenda(t *testing.T) {
	ctx := context.Background()
	repoProf := memory.NuevoProfesional()
	svcProf := NuevoProfesional(repoProf)
	svc := NuevaAgenda(repoProf, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo())

	dueno := uuid.New()
	intruso := uuid.New()

	p, err := svcProf.Crear(ctx, dueno, entradaValidaSvc())
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	t.Run("ReemplazarHorarios", func(t *testing.T) {
		_, err := svc.ReemplazarHorarios(ctx, intruso, p.ID, nil)
		if !errors.Is(err, domain.ErrNoAutorizado) {
			t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
		}
	})

	t.Run("CrearBloqueo", func(t *testing.T) {
		_, err := svc.CrearBloqueo(ctx, intruso, p.ID, domain.EntradaBloqueo{})
		if !errors.Is(err, domain.ErrNoAutorizado) {
			t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
		}
	})

	t.Run("EliminarBloqueo", func(t *testing.T) {
		err := svc.EliminarBloqueo(ctx, intruso, p.ID, uuid.New())
		if !errors.Is(err, domain.ErrNoAutorizado) {
			t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
		}
	})

	// La autorización va antes que la validación. Un intruso que manda datos
	// inválidos tiene que recibir 403, no un 422 que le confirme que el
	// perfil existe y le enseñe el esquema del cuerpo.
	t.Run("el 403 gana al 422", func(t *testing.T) {
		_, err := svc.CrearBloqueo(ctx, intruso, p.ID, domain.EntradaBloqueo{})
		var verr domain.ErrorValidacion
		if errors.As(err, &verr) {
			t.Error("un intruso recibió un error de validación en vez de 403")
		}
	})
}

// La agenda pública se sigue leyendo sin sesión. Es el criterio de
// aceptación 4 medido en la capa de servicio.
func TestLasLecturasDeAgendaNoPidenDueno(t *testing.T) {
	ctx := context.Background()
	repoProf := memory.NuevoProfesional()
	svcProf := NuevoProfesional(repoProf)
	svc := NuevaAgenda(repoProf, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo())

	p, err := svcProf.Crear(ctx, uuid.New(), entradaValidaSvc())
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	if _, err := svc.ListarHorarios(ctx, p.ID); err != nil {
		t.Errorf("ListarHorarios sin sesión falló: %v", err)
	}
	if _, err := svc.ListarBloqueos(ctx, p.ID, nil, nil); err != nil {
		t.Errorf("ListarBloqueos sin sesión falló: %v", err)
	}
}
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
go test ./internal/service/ -run TestSoloElDuenoPuedeTocarLaAgenda -v
```

Esperado: FAIL de compilación.

- [ ] **Paso 3: implementar**

Los tres métodos ya obtienen el profesional del repositorio para validar
—`ReemplazarHorarios` lo necesita para `verificarModalidades`—. Donde no lo
hagan, agregar el `ObtenerPorID` y usar el mismo `verificarPropiedad` de la
Task 8, que ya está en el paquete.

**El chequeo va inmediatamente después de obtener el profesional y antes de
cualquier validación de la entrada.** El orden importa: validar primero le
devuelve a un intruso un 422 que confirma que el perfil existe y le describe el
esquema del cuerpo.

```go
func (s *Agenda) CrearBloqueo(ctx context.Context, usuarioID, profesionalID uuid.UUID, entrada domain.EntradaBloqueo) (domain.Bloqueo, error) {
	p, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return domain.Bloqueo{}, err
	}
	if err := verificarPropiedad(p, usuarioID); err != nil {
		return domain.Bloqueo{}, err
	}
	// ... la validación y el resto quedan igual
```

- [ ] **Paso 4: correr los tests y verificar que pasan**

```bash
go test ./internal/service/ -race -v
```

Esperado: PASS.

- [ ] **Paso 5: commit**

```bash
make check
git add internal/service/agenda.go internal/service/agenda_test.go
git commit -m "feat(service): solo el dueño puede tocar su agenda"
```

---

## Task 10: contrato OpenAPI

El contrato se escribe **antes** que los handlers, igual que en las dos etapas
anteriores.

**Archivos:**
- Modificar: `api/openapi.yaml`

- [ ] **Paso 1: agregar los esquemas nuevos**

En `components/schemas`:

- `PeticionRegistro`: `email` (string, format email), `contrasena` (string,
  minLength 8, maxLength 72, `format: password`), `nombre`, `apellido`.
  Los cuatro `required`.
- `PeticionLogin`: `email`, `contrasena`. Los dos `required`.
- `Usuario`: `id` (uuid), `email`, `nombre`, `apellido`, `creadoEn`
  (date-time). **Sin `hash` y sin token.**
- `UsuarioActual`: `Usuario` más `perfilProfesionalId` (uuid, **nullable**),
  que es lo que le dice al front si mostrar la vista de profesional o la de
  paciente. `null` significa "no tiene perfil".

`contrasena` lleva `format: password` para que las herramientas que generan
clientes o UIs no la muestren en claro, y **no** aparece en ninguna respuesta.

- [ ] **Paso 2: agregar las cuatro operaciones**

```yaml
  /api/v1/usuarios:
    post:      # 201 + Set-Cookie; 409 email en uso; 422 validación; 415
  /api/v1/usuarios/yo:
    get:       # 200 UsuarioActual; 401 sin sesión
  /api/v1/sesiones:
    post:      # 201 + Set-Cookie; 401 credenciales inválidas; 415
  /api/v1/sesiones/actual:
    delete:    # 204 siempre, incluso sin sesión: el logout es idempotente
```

El `Set-Cookie` se documenta como header de respuesta en las dos operaciones
que abren sesión.

- [ ] **Paso 3: declarar el esquema de seguridad**

```yaml
components:
  securitySchemes:
    sesionCookie:
      type: apiKey
      in: cookie
      name: sesion
```

Aplicarlo con `security: [{ sesionCookie: [] }]` **solo** en las operaciones
privadas: `POST /profesionales`, `PUT`/`DELETE /profesionales/{id}`,
`POST .../reactivar`, `PUT .../horarios`, `POST`/`DELETE .../bloqueos`, y
`GET /usuarios/yo`. **No** en los `GET` públicos.

- [ ] **Paso 4: agregar las respuestas de error nuevas**

- **401** en todas las operaciones privadas y en `POST /sesiones`.
- **403** en todas las privadas menos `POST /profesionales` (ahí no hay un
  dueño previo contra el cual fallar).
- **409** en `POST /usuarios` (email en uso) y en `POST /profesionales`
  (el usuario ya tiene perfil).
- **415** en todas las operaciones con cuerpo.

Todas con el schema `Problema` que ya existe.

- [ ] **Paso 5: documentar que `POST /profesionales` cambió**

En su `description`, dejar escrito que crea **el perfil del usuario
autenticado**, que el dueño sale de la sesión y no del cuerpo, y que un
segundo intento del mismo usuario da 409.

- [ ] **Paso 6: validar el contrato**

```bash
npx @redocly/cli lint api/openapi.yaml
```

Esperado: cero errores. El `redocly.yaml` de la raíz (`extends: minimal`) ya
está configurado — **sin punto adelante**, con `.redocly.yaml` el CLI ignora la
config en silencio.

- [ ] **Paso 7: caminar los `$ref` a mano**

```bash
grep -o '\$ref: .*' api/openapi.yaml | sort -u
```

Verificar que cada uno resuelve a un schema declarado. El linter no atrapa un
`$ref` a un componente inexistente en todos los casos.

- [ ] **Paso 8: commit**

```bash
git add api/openapi.yaml
git commit -m "feat(api): contrato de registro, login y logout"
```

---

## Task 11: middleware, 415 y mapeo de errores

**Archivos:**
- Modificar: `internal/handler/middleware.go`, `internal/handler/middleware_test.go`
- Modificar: `internal/handler/dto.go`, `internal/handler/dto_test.go`
- Modificar: `internal/handler/problema.go`, `internal/handler/problema_test.go`

**Interfaces:**
- Produce: `handler.UsuarioIDDe(context.Context) (uuid.UUID, bool)`,
  `handler.RequerirSesion(http.HandlerFunc) http.HandlerFunc`, y los tipos de
  problema `tipoNoAutenticado` y `tipoNoAutorizado`.
- El middleware `Autenticar` lo produce la Task 12 como método de
  `ManejadorAutenticacion`, porque necesita el servicio.

- [ ] **Paso 1: escribir los tests que fallan**

Agregar a `internal/handler/middleware_test.go`:

```go
func TestRequerirSesionSinUsuario(t *testing.T) {
	llamado := false
	h := RequerirSesion(func(w http.ResponseWriter, r *http.Request) {
		llamado = true
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/api/v1/profesionales", nil))

	if llamado {
		t.Error("el handler corrió sin sesión")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("estado = %d, se esperaba 401", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}
}

func TestRequerirSesionConUsuario(t *testing.T) {
	usuarioID := uuid.New()
	var visto uuid.UUID

	h := RequerirSesion(func(w http.ResponseWriter, r *http.Request) {
		visto, _ = UsuarioIDDe(r.Context())
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profesionales", nil)
	req = req.WithContext(context.WithValue(req.Context(), claveUsuarioID, usuarioID))

	h(httptest.NewRecorder(), req)

	if visto != usuarioID {
		t.Errorf("UsuarioIDDe = %v, se esperaba %v", visto, usuarioID)
	}
}

func TestUsuarioIDDeSinValor(t *testing.T) {
	if _, ok := UsuarioIDDe(context.Background()); ok {
		t.Error("UsuarioIDDe devolvió ok sobre un contexto vacío")
	}
}
```

Agregar a `internal/handler/dto_test.go`:

```go
// Sin este chequeo, un formulario HTML de otro sitio puede mandar un POST con
// la cookie de sesión adjunta. Con él, cualquier cuerpo cross-origin muere en
// 415 antes de tocar el servicio: un form solo puede mandar
// form-urlencoded, multipart o text/plain, nunca application/json.
func TestDecodificarJSONExigeContentType(t *testing.T) {
	casos := []struct {
		nombre      string
		contentType string
		esperado    int
	}{
		{"json", "application/json", 0},
		{"json con charset", "application/json; charset=utf-8", 0},
		{"form urlencoded", "application/x-www-form-urlencoded", http.StatusUnsupportedMediaType},
		{"texto plano", "text/plain", http.StatusUnsupportedMediaType},
		{"multipart", "multipart/form-data; boundary=x", http.StatusUnsupportedMediaType},
		{"ausente", "", http.StatusUnsupportedMediaType},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
			if c.contentType != "" {
				req.Header.Set("Content-Type", c.contentType)
			}
			rec := httptest.NewRecorder()

			var destino struct{}
			err := decodificarJSON(rec, req, &destino)

			if c.esperado == 0 {
				if err != nil {
					t.Fatalf("se esperaba éxito, se obtuvo %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("se esperaba error por Content-Type")
			}
			escribirErrorDeDecodificacion(rec, req, err)
			if rec.Code != c.esperado {
				t.Errorf("estado = %d, se esperaba %d", rec.Code, c.esperado)
			}
		})
	}
}
```

Agregar a `internal/handler/problema_test.go`, en la tabla de casos existente:

```go
		{"no autorizado", domain.ErrNoAutorizado, http.StatusForbidden},
		{"credenciales invalidas", domain.ErrCredencialesInvalidas, http.StatusUnauthorized},
		{"email en uso", domain.ErrEmailEnUso, http.StatusConflict},
		{"ya tiene perfil", domain.ErrYaTienePerfil, http.StatusConflict},
		{"usuario en uso", domain.ErrUsuarioEnUso, http.StatusConflict},
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
go test ./internal/handler/ -run 'TestRequerirSesion|TestDecodificarJSONExige' -v
```

Esperado: FAIL — `undefined: RequerirSesion`, `undefined: claveUsuarioID`.

- [ ] **Paso 3: agregar el contexto y `RequerirSesion`**

En `internal/handler/middleware.go`, junto a `claveIDPeticion`:

```go
const claveUsuarioID claveContexto = "usuarioID"

// UsuarioIDDe devuelve el usuario autenticado. El segundo valor dice si había
// sesión: distinguirlo del uuid.Nil importa, porque uuid.Nil comparado contra
// un UsuarioID real da "no autorizado" y no "no autenticado", y esos son dos
// códigos HTTP distintos.
func UsuarioIDDe(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(claveUsuarioID).(uuid.UUID)
	return id, ok
}

// RequerirSesion corta con 401 si el middleware Autenticar no dejó un usuario.
//
// Se aplica por ruta y no globalmente: la mayoría del contrato es pública, y
// una lista de excepciones se desactualiza en silencio la primera vez que
// alguien agrega un endpoint. Envolver el handler es visible en la tabla de
// rutas.
func RequerirSesion(siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UsuarioIDDe(r.Context()); !ok {
			escribirProblema(w, Problema{
				Tipo:    tipoNoAutenticado,
				Titulo:  "No autenticado",
				Estado:  http.StatusUnauthorized,
				Detalle: "Necesitás iniciar sesión para hacer esto",
			})
			return
		}
		siguiente(w, r)
	}
}
```

- [ ] **Paso 4: agregar el chequeo de `Content-Type`**

En `internal/handler/dto.go`, al principio de `decodificarJSON`:

```go
	if err := verificarContentTypeJSON(r); err != nil {
		return err
	}
```

Y el helper más el error centinela, en el mismo archivo:

```go
// errTipoDeContenido lo distingue escribirErrorDeDecodificacion para responder
// 415 en vez de 400.
var errTipoDeContenido = errors.New("el Content-Type debe ser application/json")

// verificarContentTypeJSON es media defensa contra CSRF, y la otra media es
// SameSite=Lax en la cookie. Un formulario de otro sitio solo puede mandar
// form-urlencoded, multipart o text/plain: ninguno pasa de acá, así que no
// puede forjar una escritura aunque el browser adjunte la cookie.
//
// Se parsea en vez de comparar la cadena entera porque
// "application/json; charset=utf-8" es legítimo y un cliente real lo manda.
func verificarContentTypeJSON(r *http.Request) error {
	tipo, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || tipo != "application/json" {
		return errTipoDeContenido
	}
	return nil
}
```

Y en `escribirErrorDeDecodificacion`, antes del chequeo de `MaxBytesError`:

```go
	if errors.Is(err, errTipoDeContenido) {
		escribirProblema(w, Problema{
			Tipo:    tipoTipoDeContenido,
			Titulo:  "Tipo de contenido no soportado",
			Estado:  http.StatusUnsupportedMediaType,
			Detalle: "el cuerpo tiene que ser application/json",
		})
		return
	}
```

- [ ] **Paso 5: mapear los errores nuevos**

En `internal/handler/problema.go`, agregar a la lista de tipos:

```go
	tipoNoAutenticado   = "https://salud.app/errors/unauthorized"
	tipoNoAutorizado    = "https://salud.app/errors/forbidden"
	tipoTipoDeContenido = "https://salud.app/errors/unsupported-media-type"
```

Y en `escribirError`, **antes del `default`**:

```go
	case errors.Is(err, domain.ErrCredencialesInvalidas):
		escribirProblema(w, Problema{
			Tipo:    tipoNoAutenticado,
			Titulo:  "Credenciales inválidas",
			Estado:  http.StatusUnauthorized,
			Detalle: "El email o la contraseña no son correctos",
		})

	case errors.Is(err, domain.ErrNoAutorizado):
		escribirProblema(w, Problema{
			Tipo:    tipoNoAutorizado,
			Titulo:  "No autorizado",
			Estado:  http.StatusForbidden,
			Detalle: "Solo el dueño de este perfil puede modificarlo",
		})

	case errors.Is(err, domain.ErrEmailEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Email ya registrado",
			Estado:  http.StatusConflict,
			Detalle: "Ya existe una cuenta con ese email",
		})

	case errors.Is(err, domain.ErrYaTienePerfil), errors.Is(err, domain.ErrUsuarioEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Ya tenés un perfil profesional",
			Estado:  http.StatusConflict,
			Detalle: "Cada cuenta puede tener un solo perfil profesional",
		})
```

Agregar `"mime"` a los imports de `dto.go`.

- [ ] **Paso 6: correr los tests y verificar que pasan**

```bash
go test ./internal/handler/ -race -v
```

Esperado: PASS. **Van a fallar tests existentes** que hacen POST/PUT sin
`Content-Type`. Es correcto: hay que agregarles el header, no aflojar la
regla. Los tests que ya lo mandan están en
`internal/handler/profesional_test.go:63`.

- [ ] **Paso 7: commit**

```bash
make check
git add internal/handler/
git commit -m "feat(handler): RequerirSesion, 415 por Content-Type y errores de auth"
```

---

## Task 12: endpoints de autenticación

**Archivos:**
- Crear: `internal/handler/autenticacion.go`, `internal/handler/dto_autenticacion.go`
- Crear: `internal/handler/autenticacion_test.go`
- Modificar: `internal/handler/router.go`, `internal/handler/router_test.go`

**Interfaces:**
- Produce: `handler.NuevaAutenticacion(svc *service.Autenticacion, profesionales *service.Profesional, cookieSegura bool) *ManejadorAutenticacion`
  con los métodos `Registrar`, `IniciarSesion`, `CerrarSesion`, `Yo`, y el
  middleware `Autenticar(http.Handler) http.Handler`.
- `NuevoRouter` pasa a recibir un tercer parámetro: `*ManejadorAutenticacion`.

- [ ] **Paso 1: escribir los tests que fallan**

`internal/handler/autenticacion_test.go` cubre, contra un servidor real
(`httptest.NewServer`) y con un `http.Client` que tenga `Jar` para que las
cookies viajen solas:

| Test | Verifica |
|---|---|
| `TestRegistroDevuelveCookie` | 201, `Set-Cookie` con `HttpOnly`, y que el cuerpo **no** trae el token ni el hash |
| `TestRegistroEmailDuplicado` | 409 |
| `TestRegistroInvalido` | 422 nombrando el campo |
| `TestLoginCorrecto` | 201 + cookie |
| `TestLoginIncorrecto` | 401, y el **mismo** cuerpo para email inexistente y contraseña mala |
| `TestYoConSesion` | 200 con el email y `perfilProfesionalId: null` |
| `TestYoSinSesion` | 401 |
| `TestYoConPerfil` | tras crear un perfil, `perfilProfesionalId` trae el UUID |
| `TestLogoutInvalidaLaSesion` | 204, y el `GET /usuarios/yo` siguiente da 401 |
| `TestLogoutSinSesion` | 204 igual: es idempotente |

El de la cookie es el que más importa que sea explícito:

```go
func TestRegistroDevuelveCookie(t *testing.T) {
	srv := servidorDePrueba(t)
	defer srv.Close()

	resp := postJSON(t, srv, "/api/v1/usuarios", `{
		"email": "juan@ejemplo.com",
		"contrasena": "unaclave8",
		"nombre": "Juan",
		"apellido": "Pérez"
	}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("estado = %d, se esperaba 201", resp.StatusCode)
	}

	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == nombreCookieSesion {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("no vino la cookie de sesión")
	}
	if !cookie.HttpOnly {
		t.Error("la cookie no es HttpOnly: cualquier XSS puede leerla")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Error("la cookie no es SameSite=Lax")
	}
	if cookie.Value == "" {
		t.Error("la cookie vino vacía")
	}

	cuerpo, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("leyendo el cuerpo: %v", err)
	}
	// El token viaja SOLO en el Set-Cookie. Si aparece en el JSON, cualquier
	// script de la página puede leerlo y el HttpOnly no sirvió de nada.
	if strings.Contains(string(cuerpo), cookie.Value) {
		t.Error("el token de sesión salió en el cuerpo de la respuesta")
	}
	if strings.Contains(string(cuerpo), "unaclave8") ||
		strings.Contains(string(cuerpo), "hash") {
		t.Error("la respuesta filtró la contraseña o el hash")
	}
}
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
go test ./internal/handler/ -run TestRegistro -v
```

Esperado: FAIL — `undefined: NuevaAutenticacion`.

- [ ] **Paso 3: escribir los DTOs**

`internal/handler/dto_autenticacion.go`: `peticionRegistro`, `peticionLogin`,
`respuestaUsuario` y `respuestaUsuarioActual`, con las claves JSON del
contrato. La respuesta **no** tiene `hash`, `contrasena` ni `token` —no es una
omisión, es la regla: nada de eso sale nunca del servidor—.

```go
type respuestaUsuarioActual struct {
	respuestaUsuario
	// PerfilProfesionalID es null si el usuario no tiene perfil. Es lo único
	// que necesita el front para saber qué vista mostrar, y es derivado: no
	// hay campo Rol en ningún lado. Ver la sección 3.1 de la spec.
	PerfilProfesionalID *string `json:"perfilProfesionalId"`
}
```

- [ ] **Paso 4: escribir el manejador**

`internal/handler/autenticacion.go`:

```go
const nombreCookieSesion = "sesion"

type ManejadorAutenticacion struct {
	svc           *service.Autenticacion
	profesionales *service.Profesional
	cookieSegura  bool
}

// NuevaAutenticacion recibe cookieSegura en vez de decidirlo solo. El flag
// Secure impide que la cookie viaje por HTTP sin TLS, que es lo que se quiere
// en producción; en desarrollo contra http://localhost los browsers la
// aceptan igual —localhost es un origen de confianza— pero cualquier otra
// dirección de desarrollo, una IP de la LAN para probar desde el teléfono,
// la descarta en silencio y el login "no anda" sin ningún error visible.
func NuevaAutenticacion(svc *service.Autenticacion, profesionales *service.Profesional, cookieSegura bool) *ManejadorAutenticacion {
	return &ManejadorAutenticacion{svc: svc, profesionales: profesionales, cookieSegura: cookieSegura}
}

// Autenticar resuelve la cookie y deja el UsuarioID en el contexto.
//
// No rechaza nada: un request sin cookie o con una cookie vencida sigue de
// largo sin usuario. Rechazar es trabajo de RequerirSesion, ruta por ruta,
// porque la mayor parte del contrato es pública y un middleware global que
// corte tendría que llevar una lista de excepciones que se desactualiza sola.
func (h *ManejadorAutenticacion) Autenticar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(nombreCookieSesion)
		if err != nil {
			siguiente.ServeHTTP(w, r)
			return
		}

		u, err := h.svc.ResolverSesion(r.Context(), cookie.Value)
		if err != nil {
			// Token inválido o sesión vencida: se sigue como anónimo y se
			// borra la cookie muerta para que el browser deje de mandarla.
			h.borrarCookie(w)
			siguiente.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), claveUsuarioID, u.ID)
		siguiente.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *ManejadorAutenticacion) ponerCookie(w http.ResponseWriter, token string, expira time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookieSesion,
		Value:    token,
		Path:     "/",
		Expires:  expira,
		HttpOnly: true, // invisible para JS: un XSS no se lleva la sesión
		Secure:   h.cookieSegura,
		SameSite: http.SameSiteLaxMode, // media defensa contra CSRF; la otra media es el 415
	})
}

func (h *ManejadorAutenticacion) borrarCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookieSesion,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSegura,
		SameSite: http.SameSiteLaxMode,
	})
}
```

Los cuatro handlers: `Registrar` (201), `IniciarSesion` (201),
`CerrarSesion` (204 siempre, borra la cookie aunque no hubiera sesión), `Yo`
(200; consulta `profesionales.ObtenerPorUsuarioID` y traduce `ErrNoEncontrado`
a `null`, que **no** es un error).

- [ ] **Paso 5: cablear el router**

En `internal/handler/router.go`, cambiar la firma y agregar las rutas:

```go
func NuevoRouter(ph *ManejadorProfesional, ah *ManejadorAgenda, mh *ManejadorAutenticacion) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)

	mux.HandleFunc("POST /api/v1/usuarios", mh.Registrar)
	mux.HandleFunc("GET /api/v1/usuarios/yo", RequerirSesion(mh.Yo))
	mux.HandleFunc("POST /api/v1/sesiones", mh.IniciarSesion)
	mux.HandleFunc("DELETE /api/v1/sesiones/actual", mh.CerrarSesion)
```

Y envolver las privadas. La tabla queda diciendo de un vistazo qué es público
y qué no:

```go
	mux.HandleFunc("GET /api/v1/profesionales", ph.Listar)
	mux.HandleFunc("POST /api/v1/profesionales", RequerirSesion(ph.Crear))
	mux.HandleFunc("GET /api/v1/profesionales/{id}", ph.ObtenerPorID)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}", RequerirSesion(ph.Actualizar))
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}", RequerirSesion(ph.DarDeBaja))
	mux.HandleFunc("POST /api/v1/profesionales/{id}/reactivar", RequerirSesion(ph.Reactivar))

	mux.HandleFunc("GET /api/v1/perfiles/{slug}", ph.ObtenerPorSlug)

	mux.HandleFunc("GET /api/v1/profesionales/{id}/horarios", ah.ListarHorarios)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}/horarios", RequerirSesion(ah.ReemplazarHorarios))
	mux.HandleFunc("GET /api/v1/profesionales/{id}/bloqueos", ah.ListarBloqueos)
	mux.HandleFunc("POST /api/v1/profesionales/{id}/bloqueos", RequerirSesion(ah.CrearBloqueo))
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}/bloqueos/{bloqueoId}", RequerirSesion(ah.EliminarBloqueo))
	mux.HandleFunc("GET /api/v1/profesionales/{id}/huecos", ah.HuecosLibres)
```

**`Autenticar` va después de `IDPeticion` y antes de `RegistrarPeticiones`**,
para que el log de cada request pueda incluir el usuario:

```go
	return Encadenar(envolverErroresDeRuteo(mux),
		IDPeticion, mh.Autenticar, RegistrarPeticiones, RecuperarPanic)
```

Las rutas nuevas no colisionan: `/api/v1/usuarios` y `/api/v1/sesiones` son
prefijos que no existían, y `usuarios/yo` es un literal de cuatro segmentos sin
ninguna plantilla `{...}` que compita.

- [ ] **Paso 6: correr los tests y verificar que pasan**

```bash
go test ./internal/handler/ -race -v
```

Esperado: PASS, incluidos los 10 tests nuevos.

- [ ] **Paso 7: commit**

```bash
make check
git add internal/handler/
git commit -m "feat(handler): endpoints de registro, login, logout y usuario actual"
```

---

## Task 13: los handlers existentes usan la sesión

**Archivos:**
- Modificar: `internal/handler/profesional.go`, `internal/handler/agenda.go`
- Modificar: `internal/handler/profesional_test.go`, `internal/handler/agenda_test.go`

- [ ] **Paso 1: escribir el test que falla**

Un test por cada operación privada verificando el 403 con sesión ajena, y uno
de humo verificando que las públicas siguen andando sin cookie:

```go
// Criterio de aceptación 3, medido en HTTP.
func TestUnIntrusoNoPuedeEditarUnPerfilAjeno(t *testing.T) {
	srv := servidorDePrueba(t)
	defer srv.Close()

	dueno := registrarYLoguear(t, srv, "dueno@ejemplo.com")
	perfil := crearPerfil(t, srv, dueno)
	intruso := registrarYLoguear(t, srv, "intruso@ejemplo.com")

	casos := []struct {
		nombre string
		metodo string
		ruta   string
		cuerpo string
	}{
		{"editar perfil", http.MethodPut, "/api/v1/profesionales/" + perfil, cuerpoPerfilValido},
		{"dar de baja", http.MethodDelete, "/api/v1/profesionales/" + perfil, ""},
		{"reactivar", http.MethodPost, "/api/v1/profesionales/" + perfil + "/reactivar", ""},
		{"cargar horarios", http.MethodPut, "/api/v1/profesionales/" + perfil + "/horarios", `{"horarios":[]}`},
		{"crear bloqueo", http.MethodPost, "/api/v1/profesionales/" + perfil + "/bloqueos", cuerpoBloqueoValido},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resp := pedir(t, intruso, c.metodo, srv.URL+c.ruta, c.cuerpo)
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("estado = %d, se esperaba 403", resp.StatusCode)
			}
		})
	}
}

// Criterio de aceptación 4: sin cookie, todo lo público sigue igual.
func TestLasRutasPublicasNoPidenSesion(t *testing.T) {
	srv := servidorDePrueba(t)
	defer srv.Close()

	dueno := registrarYLoguear(t, srv, "dueno@ejemplo.com")
	perfil := crearPerfil(t, srv, dueno)

	anonimo := &http.Client{} // sin cookie jar: nunca manda cookies

	rutas := []string{
		"/api/v1/profesionales",
		"/api/v1/profesionales/" + perfil,
		"/api/v1/profesionales/" + perfil + "/horarios",
		"/api/v1/profesionales/" + perfil + "/bloqueos",
	}

	for _, ruta := range rutas {
		t.Run(ruta, func(t *testing.T) {
			resp, err := anonimo.Get(srv.URL + ruta)
			if err != nil {
				t.Fatalf("GET %s: %v", ruta, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("estado = %d, se esperaba 200", resp.StatusCode)
			}
		})
	}
}
```

- [ ] **Paso 2: correr y verificar que falla**

```bash
go test ./internal/handler/ -run TestUnIntruso -v
```

- [ ] **Paso 3: implementar**

Cada handler privado empieza leyendo el usuario. Como la ruta está envuelta en
`RequerirSesion`, `ok` siempre es `true` — pero el `if` queda igual: si alguien
saca el `RequerirSesion` de la tabla de rutas, esto responde 401 en vez de
tratar `uuid.Nil` como un usuario legítimo.

```go
func (h *ManejadorProfesional) Crear(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDDe(r.Context())
	if !ok {
		escribirProblema(w, Problema{
			Tipo:    tipoNoAutenticado,
			Titulo:  "No autenticado",
			Estado:  http.StatusUnauthorized,
			Detalle: "Necesitás iniciar sesión para hacer esto",
		})
		return
	}

	var req peticionProfesional
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	p, err := h.svc.Crear(r.Context(), usuarioID, req.aEntrada())
	// ... el resto queda igual
```

Extraer ese bloque a un helper `usuarioAutenticado(w, r) (uuid.UUID, bool)` en
`middleware.go` y usarlo en los 8 handlers privados: repetirlo ocho veces es
ocho lugares donde escribir mal el código de estado.

- [ ] **Paso 4: correr los tests y verificar que pasan**

```bash
go test ./... -race
```

Esperado: PASS en todos los paquetes.

- [ ] **Paso 5: commit**

```bash
make check
git add internal/handler/
git commit -m "feat(handler): las operaciones privadas usan el usuario de la sesion"
```

---

## Task 14: seed, composition root y documentación

**Archivos:**
- Modificar: `cmd/api/semilla.go`, `cmd/api/semilla_test.go`, `cmd/api/main.go`
- Modificar: `apps/api/README.md`, `apps/api/.env.example`

- [ ] **Paso 1: sembrar usuarios**

`sembrar` pasa a recibir los dos servicios y crea un usuario por profesional
**antes** que su perfil, porque ahora el perfil necesita dueño.

```go
func sembrar(ctx context.Context, auth *service.Autenticacion, svc *service.Profesional) error {
```

Emails derivados del nombre: `martin.gonzalez@ejemplo.com`, etc. Una sola
contraseña para los cuatro, declarada en una constante con su comentario:

```go
// contrasenaSemilla es la misma para los cuatro profesionales de prueba. Es
// seguro porque sembrar() solo corre con APP_ENV=development —main lo gatea
// en cfg.EsDesarrollo()— y en producción el binario arranca vacío.
const contrasenaSemilla = "desarrollo123"
```

- [ ] **Paso 2: actualizar `main.go`**

```go
	repo := memory.NuevoProfesional()
	repoHorarios := memory.NuevoHorarioSemanal()
	repoBloqueos := memory.NuevoBloqueo()
	repoUsuarios := memory.NuevoUsuario()
	repoSesiones := memory.NuevaSesion()

	svc := service.NuevoProfesional(repo)
	svcAgenda := service.NuevaAgenda(repo, repoHorarios, repoBloqueos)
	svcAuth := service.NuevaAutenticacion(repoUsuarios, repoSesiones)

	if cfg.EsDesarrollo() {
		if err := sembrar(context.Background(), svcAuth, svc); err != nil {
			return fmt.Errorf("cargando el seed: %w", err)
		}
		slog.Info("seed de desarrollo cargado")
	}

	router := handler.NuevoRouter(
		handler.NuevoProfesional(svc),
		handler.NuevaAgenda(svcAgenda),
		// La cookie va con Secure salvo en desarrollo: sin TLS el browser la
		// descarta, y en desarrollo se sirve por HTTP.
		handler.NuevaAutenticacion(svcAuth, svc, !cfg.EsDesarrollo()),
	)
```

- [ ] **Paso 3: actualizar el README**

En "Convenciones", la línea de dependencias:

```markdown
- Dos dependencias externas: `github.com/google/uuid` y
  `golang.org/x/crypto` (bcrypt). El hashing de contraseñas es la única parte
  del proyecto que no se escribe a mano.
```

En "Arquitectura", agregar que `Autenticar` deja el `UsuarioID` en el contexto
y que la autorización se decide en el servicio, no en el handler.

Agregar una sección corta con el flujo para probar a mano:

```bash
curl -c galletas.txt -X POST localhost:8080/api/v1/sesiones \
  -H 'Content-Type: application/json' \
  -d '{"email":"martin.gonzalez@ejemplo.com","contrasena":"desarrollo123"}'

curl -b galletas.txt localhost:8080/api/v1/usuarios/yo
```

- [ ] **Paso 4: verificar el arranque en los dos entornos**

```bash
make run
# en otra terminal: el login de arriba, y después
curl localhost:8080/api/v1/profesionales | head -c 300
```

```bash
APP_ENV=production go run ./cmd/api
curl localhost:8080/api/v1/profesionales
```

Esperado: en producción, listado vacío y **sin** usuarios sembrados.

- [ ] **Paso 5: commit**

```bash
make check
git add cmd/api/ README.md .env.example
git commit -m "feat(api): seed con usuarios y cableado de autenticacion"
```

---

## Task 15: verificación final

Sin commit propio salvo que aparezca algo para arreglar.

- [ ] **Paso 1: la suite completa con detector de carreras**

```bash
make check
```

Esperado: verde. **Criterio 1.**

- [ ] **Paso 2: recorrer el flujo completo contra el servidor real**

```bash
make run
```

En otra terminal, con `curl -c/-b` para llevar la cookie:

| # | Qué | Esperado | Criterio |
|---|---|---|---|
| 1 | Registro de un usuario nuevo | 201 + `Set-Cookie` | |
| 2 | Crear su perfil profesional | 201 | **2** |
| 3 | El perfil aparece en `GET /api/v1/profesionales` | 200, ahí está | **2** |
| 4 | Segundo usuario intenta editar ese perfil | **403** | **3** |
| 5 | Idem sobre horarios y bloqueos | **403** | **3** |
| 6 | Todos los `GET` públicos sin cookie | 200 | **4** |
| 7 | `DELETE /api/v1/sesiones/actual` y reintentar | 204, después **401** | **5** |
| 8 | `POST` con `Content-Type: text/plain` | **415** | **7** |
| 9 | El seed: login con `martin.gonzalez@ejemplo.com` | 201 | **8** |

- [ ] **Paso 3: verificar el vencimiento de sesión**

**Criterio 6.** No se puede esperar 7 días: se verifica con el test
`TestResolverSesionVencida` de la Task 6, que mueve el reloj inyectado.

```bash
go test ./internal/service/ -run TestResolverSesionVencida -v
```

Confirmar además, leyendo el código, que `ResolverSesion` chequea
`EstaVencida` **antes** de devolver el usuario.

- [ ] **Paso 4: mutation testing de las dos reglas que sostienen la etapa**

Los tests que verifican una prohibición pasan también cuando la prohibición no
existe, si están mal escritos. Romper la regla a propósito es la única forma de
saberlo.

1. Hacer que `verificarPropiedad` devuelva siempre `nil`.
   → tienen que fallar `TestSoloElDuenoPuedeMutar`,
   `TestSoloElDuenoPuedeTocarLaAgenda` y `TestUnIntrusoNoPuedeEditarUnPerfilAjeno`.
2. Hacer que `Sesion.EstaVencida` devuelva siempre `false`.
   → tiene que fallar `TestResolverSesionVencida`.

Revertir las dos.

- [ ] **Paso 5: verificar el aislamiento del dominio**

```bash
go list -deps ./internal/domain/ | grep 'joaquinfochoa' | grep -v 'internal/domain'
```

Esperado: sin salida.

- [ ] **Paso 6: validar el contrato contra la implementación**

```bash
npx @redocly/cli lint api/openapi.yaml
```

Y a mano: que cada ruta de `router.go` esté en `openapi.yaml` con el mismo
método, y que cada operación marcada con `security: sesionCookie` esté envuelta
en `RequerirSesion` en el router. Las dos listas tienen que coincidir
exactamente — una ruta protegida en el contrato pero abierta en el código es un
agujero que ningún test atrapa.

- [ ] **Paso 7: verificar la imagen**

```bash
make docker-build
docker run --rm -p 8080:8080 -e APP_ENV=production salud-api
curl localhost:8080/healthz
```

Esperado: arranca, responde `{"estado":"ok"}`, sin datos sembrados.

- [ ] **Paso 8: registrar el progreso**

Agregar la sección "Etapa 3 — Identidad y autenticación" a
`.superpowers/sdd/progress.md`, con el estado de cada tarea y los minor que
hayan quedado pendientes, igual que las dos etapas anteriores.

---

## Notas de riesgo

**La tarea más peligrosa es la 7.** Cambia la firma de `NuevoProfesional`, y
todo lo que no compile es visible. Lo que **no** es visible es el
`AplicarCambios` que se olvide de preservar `UsuarioID`: el código compila, los
tests viejos pasan, y cada `PUT` deja el perfil huérfano —`uuid.Nil`— así que
el dueño pierde el acceso a su propio perfil en cuanto edita su bio. Por eso
`TestAplicarCambiosPreservaElDueno` está en la misma tarea y no después.

**La tarea 13 es la que puede dejar un agujero silencioso.** Un handler privado
al que se le olvide leer el `usuarioID` y le pase `uuid.Nil` al servicio no
falla en ningún test que no lo busque específicamente: devuelve 403 siempre, que
parece correcto. El paso 6 de la Task 15 —cruzar las rutas de `router.go`
contra el `security` del contrato— es lo que lo atrapa.

**Lo que esta etapa deja sin cerrar, a propósito:** rate limiting del login (y
con él, el canal lateral por tiempo de `IniciarSesion`), recuperación de
contraseña, y limpieza de sesiones vencidas más allá del borrado al leer. Los
tres están en la sección 7 de la spec con su disparador.
