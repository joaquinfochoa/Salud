package handler

import (
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// peticionRegistro es lo que entra al alta de usuario. Contrasena viaja en
// claro sobre TLS, se hashea en el dominio y no vuelve a salir del servidor.
type peticionRegistro struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
	Nombre     string `json:"nombre"`
	Apellido   string `json:"apellido"`
	Telefono   string `json:"telefono"`
}

func (p peticionRegistro) aEntrada() domain.EntradaUsuario {
	return domain.EntradaUsuario{
		Email:      p.Email,
		Contrasena: p.Contrasena,
		Nombre:     p.Nombre,
		Apellido:   p.Apellido,
		Telefono:   p.Telefono,
	}
}

// peticionPerfil son los datos personales que alguien puede cambiar de su
// cuenta. Sin contraseña: cambiarla es otra operación, con otra regla —pedir la
// actual— y meterla acá haría que un formulario de datos personales pueda
// reemplazar una credencial.
type peticionPerfil struct {
	Email    string `json:"email"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Telefono string `json:"telefono"`
}

func (p peticionPerfil) aEntrada() domain.EntradaPerfil {
	return domain.EntradaPerfil{
		Email:    p.Email,
		Nombre:   p.Nombre,
		Apellido: p.Apellido,
		Telefono: p.Telefono,
	}
}

type peticionLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

// respuestaUsuario no tiene hash, contraseña ni token. No es una omisión: es
// la regla. El token viaja solo en el Set-Cookie, y el hash no sale nunca.
type respuestaUsuario struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	// Telefono sale acá porque este DTO responde SOBRE VOS: el registro y
	// /usuarios/yo. Tu propio teléfono lo podés ver; el de otra persona sale
	// por otro camino y con otras reglas.
	Telefono string    `json:"telefono"`
	Nombre   string    `json:"nombre"`
	Apellido string    `json:"apellido"`
	CreadoEn time.Time `json:"creadoEn"`
}

func aRespuestaUsuario(u domain.Usuario) respuestaUsuario {
	return respuestaUsuario{
		ID:       u.ID.String(),
		Email:    u.Email.String(),
		Telefono: u.Telefono.String(),
		Nombre:   u.Nombre,
		Apellido: u.Apellido,
		CreadoEn: u.CreadoEn,
	}
}

// respuestaUsuarioActual es lo que devuelve GET /usuarios/yo.
type respuestaUsuarioActual struct {
	respuestaUsuario

	// PerfilProfesionalID es null si el usuario no tiene perfil. Es lo único
	// que necesita el front para saber qué vista mostrar, y es derivado: no hay
	// campo Rol en ningún lado, justamente para que no pueda desincronizarse
	// de esto.
	PerfilProfesionalID *string `json:"perfilProfesionalId"`
}
