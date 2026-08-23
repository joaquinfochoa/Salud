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
