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
