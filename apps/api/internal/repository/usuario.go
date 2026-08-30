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
// existen, y una interfaz con métodos que nadie implementa contra PostgreSQL
// es trabajo por adelantado.
//
// Contrato de unicidad, igual que en Profesional: lo garantiza la escritura, no
// el que llama. Chequear antes y escribir después son dos operaciones y entre
// las dos entra otro request. En PostgreSQL esto es una constraint UNIQUE sobre
// email, y traducir su violación a ErrEmailEnUso es todo lo que tiene que hacer
// esa implementación.
type Usuario interface {
	Crear(ctx context.Context, u domain.Usuario) error
	ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Usuario, error)
	ObtenerPorEmail(ctx context.Context, e domain.Email) (domain.Usuario, error)
}
