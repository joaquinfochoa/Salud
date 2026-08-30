package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// aserción de compilación: si la implementación deja de cumplir la interfaz,
// el error aparece acá y no en main.go
var _ repository.Usuario = (*Usuario)(nil)

// Usuario guarda las identidades de login en memoria. Se pierde todo al
// reiniciar, igual que el resto de los repositorios de esta etapa.
type Usuario struct {
	mu    sync.RWMutex
	datos map[uuid.UUID]domain.Usuario
}

func NuevoUsuario() *Usuario {
	return &Usuario{datos: make(map[uuid.UUID]domain.Usuario)}
}

func (r *Usuario) Crear(_ context.Context, u domain.Usuario) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[u.ID]; existe {
		return domain.ErrIDEnUso
	}

	// Chequear la unicidad en una llamada y escribir en otra son dos
	// operaciones, y entre las dos entra otro registro con el mismo email.
	// Bajo un mismo lock son una sola.
	//
	// ponytail: scan O(n), igual que en profesional.go. Un índice por email lo
	// evita, pero hay que mantenerlo sincronizado en cada escritura y para un
	// store de desarrollo no vale el mapa extra.
	for _, otro := range r.datos {
		if otro.Email == u.Email {
			return domain.ErrEmailEnUso
		}
	}

	r.datos[u.ID] = u.Clonar()
	return nil
}

func (r *Usuario) ObtenerPorID(_ context.Context, id uuid.UUID) (domain.Usuario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.datos[id]
	if !ok {
		return domain.Usuario{}, domain.ErrNoEncontrado
	}
	return u.Clonar(), nil
}

func (r *Usuario) ObtenerPorEmail(_ context.Context, e domain.Email) (domain.Usuario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.datos {
		if u.Email == e {
			return u.Clonar(), nil
		}
	}
	return domain.Usuario{}, domain.ErrNoEncontrado
}
