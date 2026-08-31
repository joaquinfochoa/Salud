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
// No hay campo Rol a propósito. Sos profesional si existe un Profesional con
// tu UsuarioID; un enum almacenado solo agrega un estado que puede quedar
// desincronizado del perfil real.
type Usuario struct {
	ID       uuid.UUID
	Email    Email
	Nombre   string
	Apellido string
	CreadoEn time.Time

	// Telefono es obligatorio. No se muestra en el perfil público ni en
	// ninguna página abierta: es dato personal bajo Ley 25.326, y hoy su
	// única razón de ser es poder avisar de un turno cuando existan
	// notificaciones.
	Telefono Telefono

	// Hash puede ser nil: un usuario que entra con SSO no tiene contraseña.
	// Hoy ningún constructor produce ese estado —NuevoUsuario exige
	// contraseña— pero el campo ya lo admite para que agregar
	// NuevoUsuarioConGoogle no obligue a tocar el tipo, su validación y sus
	// tests.
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
	Telefono   string
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

	if tel, err := ParsearTelefono(entrada.Telefono); err != nil {
		verr.agregar("telefono", err.Error())
	} else {
		u.Telefono = tel
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

// Clonar devuelve una copia profunda. Hash es un slice: sin esto, quien reciba
// la copia puede mutar el hash del original desde afuera.
func (u Usuario) Clonar() Usuario {
	c := u
	c.Hash = slices.Clone(u.Hash)
	return c
}

// hashearContrasena valida las reglas y devuelve el hash.
//
// Sin reglas de composición —mayúscula, número, símbolo—: alargan el
// formulario sin agregar entropía real y empujan a la gente hacia
// "Password1!". El piso de 8 caracteres es el de NIST SP 800-63B.
func hashearContrasena(plana string) ([]byte, error) {
	switch {
	case utf8.RuneCountInString(plana) < minLargoContrasena:
		return nil, errors.New("tiene que tener al menos 8 caracteres")
	case len(plana) > maxBytesContrasena:
		return nil, errors.New("no puede superar los 72 bytes")
	}
	return bcrypt.GenerateFromPassword([]byte(plana), costoBcrypt)
}
