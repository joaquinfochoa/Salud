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
