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

var _ repository.Turno = (*Turno)(nil)

// Turno guarda los turnos en memoria. Se pierde todo al reiniciar, igual que
// el resto de los repositorios.
type Turno struct {
	mu    sync.RWMutex
	datos map[uuid.UUID]domain.Turno
}

func NuevoTurno() *Turno {
	return &Turno{datos: make(map[uuid.UUID]domain.Turno)}
}

func (r *Turno) Crear(_ context.Context, t domain.Turno) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[t.ID]; existe {
		return domain.ErrIDEnUso
	}
	if err := r.conflicto(t); err != nil {
		return err
	}

	r.datos[t.ID] = t.Clonar()
	return nil
}

// conflicto busca un turno ajeno y activo que pise a t. Solo se llama con el
// lock de escritura tomado, y ahí está todo el punto: chequear la
// disponibilidad en una llamada y escribir en otra son dos operaciones, y entre
// las dos entra otro paciente. Bajo un mismo lock son una sola.
//
// Un turno cancelado no conflictúa con nada: libera su hueco y libera a su
// paciente. Es la misma regla que aplica CalculoHuecos, escrita en el otro
// lugar donde tiene que valer.
//
// ponytail: scan O(n), igual que profesional.go. Un índice por profesional y
// día lo evita, pero hay que mantenerlo sincronizado en cada escritura y para
// un store de desarrollo no vale el mapa extra.
func (r *Turno) conflicto(t domain.Turno) error {
	if !t.EstaActivo() {
		return nil // un turno que nace cancelado no puede chocar con nadie
	}
	for _, otro := range r.datos {
		if otro.ID == t.ID || !otro.EstaActivo() || !otro.SeSolapaCon(t.Inicio, t.Fin) {
			continue
		}
		if otro.ProfesionalID == t.ProfesionalID {
			return domain.ErrHuecoTomado
		}
		if otro.PacienteID == t.PacienteID {
			return domain.ErrPacienteSolapado
		}
	}
	return nil
}

func (r *Turno) ObtenerPorID(_ context.Context, id uuid.UUID) (domain.Turno, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.datos[id]
	if !ok {
		return domain.Turno{}, domain.ErrNoEncontrado
	}
	return t.Clonar(), nil
}

func (r *Turno) Actualizar(_ context.Context, t domain.Turno) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[t.ID]; !existe {
		return domain.ErrNoEncontrado
	}
	if err := r.conflicto(t); err != nil {
		return err
	}

	r.datos[t.ID] = t.Clonar()
	return nil
}

func (r *Turno) ListarDeProfesional(_ context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Turno, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.listar(func(t domain.Turno) bool { return t.ProfesionalID == profesionalID }, desde, hasta), nil
}

func (r *Turno) ListarDePaciente(_ context.Context, pacienteID uuid.UUID, desde, hasta time.Time) ([]domain.Turno, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.listar(func(t domain.Turno) bool { return t.PacienteID == pacienteID }, desde, hasta), nil
}

// listar es la mitad común de los dos listados: lo único que los distingue es a
// quién le preguntan. Solo se llama con el lock de lectura tomado.
func (r *Turno) listar(coincide func(domain.Turno) bool, desde, hasta time.Time) []domain.Turno {
	salida := make([]domain.Turno, 0, len(r.datos))
	for _, t := range r.datos {
		if coincide(t) && t.SeSolapaCon(desde, hasta) {
			salida = append(salida, t.Clonar())
		}
	}

	// El mapa de Go itera en orden aleatorio. Sin este orden, dos llamadas
	// idénticas devolverían la agenda del día en un orden distinto cada vez.
	slices.SortFunc(salida, func(a, b domain.Turno) int {
		if c := a.Inicio.Compare(b.Inicio); c != 0 {
			return c
		}
		// Desempate por ID: dos turnos pueden empezar a la misma hora si son
		// de profesionales distintos, y sin esto el orden entre ellos queda a
		// merced del mapa.
		return strings.Compare(a.ID.String(), b.ID.String())
	})
	return salida
}
