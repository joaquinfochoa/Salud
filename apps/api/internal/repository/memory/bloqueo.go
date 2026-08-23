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
