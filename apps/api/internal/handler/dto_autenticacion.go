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
}

func (p peticionRegistro) aEntrada() domain.EntradaUsuario {
	return domain.EntradaUsuario{
		Email:      p.Email,
		Contrasena: p.Contrasena,
		Nombre:     p.Nombre,
		Apellido:   p.Apellido,
	}
}

type peticionLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

// respuestaUsuario no tiene hash, contraseña ni token. No es una omisión: es
// la regla. El token viaja solo en el Set-Cookie, y el hash no sale nunca.
type respuestaUsuario struct {
	ID       string    `json:"id"`
	Email    string    `json:"email"`
	Nombre   string    `json:"nombre"`
	Apellido string    `json:"apellido"`
	CreadoEn time.Time `json:"creadoEn"`
}

func aRespuestaUsuario(u domain.Usuario) respuestaUsuario {
	return respuestaUsuario{
		ID:       u.ID.String(),
		Email:    u.Email.String(),
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
