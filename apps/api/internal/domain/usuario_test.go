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
		Telefono:   "11 1234-5678",
	}
}

// El teléfono es obligatorio y se guarda normalizado: es lo que va a usar el
// día que existan notificaciones, y dos formas de escribir el mismo número
// tienen que quedar iguales.
func TestNuevoUsuarioNormalizaElTelefono(t *testing.T) {
	e := entradaUsuarioValida()
	e.Telefono = "011 15 1234-5678"

	u, err := NuevoUsuario(e, time.Now())
	if err != nil {
		t.Fatalf("NuevoUsuario devolvió error: %v", err)
	}
	if u.Telefono != "+5491112345678" {
		t.Errorf("Telefono = %q, se esperaba %q", u.Telefono, "+5491112345678")
	}
}

func TestNuevoUsuarioExigeTelefono(t *testing.T) {
	for _, caso := range []struct{ nombre, telefono string }{
		{"vacio", ""},
		{"con letras", "no tengo"},
		{"incompleto", "1234"},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			e := entradaUsuarioValida()
			e.Telefono = caso.telefono

			_, err := NuevoUsuario(e, time.Now())
			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, vino %v", err)
			}
			if !strings.Contains(err.Error(), "telefono") {
				t.Errorf("el error no menciona el campo telefono: %v", err)
			}
		})
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

func TestConPerfil(t *testing.T) {
	original, err := NuevoUsuario(entradaUsuarioValida(), time.Now())
	if err != nil {
		t.Fatalf("NuevoUsuario: %v", err)
	}

	cambiado, err := original.ConPerfil(EntradaPerfil{
		Email:    "  Nuevo@Ejemplo.COM ",
		Nombre:   "Juana",
		Apellido: "Pérez",
		Telefono: "011 15 9999-8888",
	})
	if err != nil {
		t.Fatalf("ConPerfil devolvió error: %v", err)
	}

	// Lo que cambia, normalizado igual que en el alta.
	if cambiado.Email != "nuevo@ejemplo.com" {
		t.Errorf("Email = %q", cambiado.Email)
	}
	if cambiado.Telefono != "+5491199998888" {
		t.Errorf("Telefono = %q", cambiado.Telefono)
	}
	if cambiado.Nombre != "Juana" {
		t.Errorf("Nombre = %q", cambiado.Nombre)
	}

	// Lo que NO puede cambiar: la identidad y la credencial.
	if cambiado.ID != original.ID {
		t.Error("cambió el ID")
	}
	if !cambiado.VerificarContrasena("unaclave8") {
		t.Error("se perdió la contraseña")
	}
	if !cambiado.CreadoEn.Equal(original.CreadoEn) {
		t.Error("cambió la fecha de creación")
	}
}

func TestConPerfilInvalidoNoTocaNada(t *testing.T) {
	original, err := NuevoUsuario(entradaUsuarioValida(), time.Now())
	if err != nil {
		t.Fatalf("NuevoUsuario: %v", err)
	}

	for _, caso := range []struct {
		nombre, campo string
		entrada       EntradaPerfil
	}{
		{"email invalido", "email", EntradaPerfil{Email: "no-es-un-email", Nombre: "A", Apellido: "B", Telefono: "11 1234-5678"}},
		{"telefono invalido", "telefono", EntradaPerfil{Email: "a@b.com", Nombre: "A", Apellido: "B", Telefono: "xx"}},
		{"nombre vacio", "nombre", EntradaPerfil{Email: "a@b.com", Nombre: "", Apellido: "B", Telefono: "11 1234-5678"}},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := original.ConPerfil(caso.entrada)
			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, vino %v", err)
			}
			if !strings.Contains(err.Error(), caso.campo) {
				t.Errorf("el error no menciona %q: %v", caso.campo, err)
			}
			// Y el original sigue intacto: ConPerfil devuelve una copia.
			if original.Email != "juan@ejemplo.com" {
				t.Error("se modificó el usuario original")
			}
		})
	}
}
