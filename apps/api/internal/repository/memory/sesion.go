package memory

import (
	"context"
	"sync"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var _ repository.Sesion = (*Sesion)(nil)

// Sesion guarda las sesiones vigentes en memoria. Reiniciar el proceso
// desloguea a todo el mundo, que en desarrollo es lo esperable.
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
