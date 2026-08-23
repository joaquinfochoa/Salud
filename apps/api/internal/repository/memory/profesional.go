package memory

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// aserción de compilación: si la implementación deja de cumplir la interfaz,
// el error aparece acá y no en main.go
var _ repository.Profesional = (*Profesional)(nil)

// Profesional guarda los profesionales en memoria. Se pierde todo al
// reiniciar, y está bien: sirve para definir el dominio antes de comprometerse
// con un esquema de base de datos, y para correr los casos sin infraestructura.
type Profesional struct {
	mu    sync.RWMutex
	datos map[uuid.UUID]domain.Profesional
}

func NuevoProfesional() *Profesional {
	return &Profesional{datos: make(map[uuid.UUID]domain.Profesional)}
}

func (r *Profesional) Crear(_ context.Context, p domain.Profesional) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[p.ID]; existe {
		return domain.ErrIDEnUso
	}
	if err := r.conflicto(p); err != nil {
		return err
	}
	r.datos[p.ID] = p.Clonar()
	return nil
}

// conflicto busca un registro ajeno que ya ocupe la matrícula o el slug de p.
// Solo se llama con el lock de escritura tomado, y ahí está todo el punto:
// chequear la unicidad en una llamada y escribir en otra son dos operaciones,
// y entre las dos entra otro request. Bajo un mismo lock son una sola.
//
// Es el equivalente en memoria de las dos constraints UNIQUE que la
// implementación PostgreSQL va a declarar sobre la tabla, y por eso el servicio
// no cambia cuando se migre: allá el motor rechaza el INSERT y el repositorio
// traduce la violación a estos mismos sentinelas.
//
// Excluye al propio registro para que Actualizar no lo haga chocar consigo
// mismo: editar la bio conservando la matrícula propia no es un conflicto.
//
// ponytail: scan O(n), igual que Listar. Con índices por matrícula y slug se
// evita, pero para un store en memoria de desarrollo no vale el mapa extra que
// hay que mantener sincronizado en cada escritura.
func (r *Profesional) conflicto(p domain.Profesional) error {
	for _, otro := range r.datos {
		if otro.ID == p.ID {
			continue
		}
		if otro.Matricula == p.Matricula {
			return domain.ErrMatriculaEnUso
		}
		if otro.Slug == p.Slug {
			return domain.ErrSlugEnUso
		}
	}
	return nil
}

func (r *Profesional) ObtenerPorID(_ context.Context, id uuid.UUID) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.datos[id]
	if !ok {
		return domain.Profesional{}, domain.ErrNoEncontrado
	}
	return p.Clonar(), nil
}

func (r *Profesional) ObtenerPorSlug(_ context.Context, slug string) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.datos {
		if p.Slug == slug {
			return p.Clonar(), nil
		}
	}
	return domain.Profesional{}, domain.ErrNoEncontrado
}

func (r *Profesional) ObtenerPorMatricula(_ context.Context, m domain.Matricula) (domain.Profesional, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.datos {
		if p.Matricula == m {
			return p.Clonar(), nil
		}
	}
	return domain.Profesional{}, domain.ErrNoEncontrado
}

func (r *Profesional) Listar(_ context.Context, f repository.Filtro) ([]domain.Profesional, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// El slice de Go entra en pánico con un índice negativo, a diferencia de un
	// OFFSET de SQL. El servicio ya normaliza estos valores, pero el que corta el
	// slice es este método, así que la guarda va acá.
	if f.Desplazamiento < 0 {
		f.Desplazamiento = 0
	}
	if f.Limite < 0 {
		f.Limite = 0
	}

	// ponytail: scan O(n), correcto para un store en memoria. La
	// implementación Postgres resuelve esto con índices sobre especialidad,
	// zona y estado.
	coincidentes := make([]domain.Profesional, 0, len(r.datos))
	for _, p := range r.datos {
		if coincide(p, f) {
			coincidentes = append(coincidentes, p.Clonar())
		}
	}

	// El mapa de Go itera en orden aleatorio. Sin este orden, dos llamadas
	// idénticas devolverían páginas distintas y la paginación sería inútil.
	slices.SortFunc(coincidentes, func(a, b domain.Profesional) int {
		if c := a.CreadoEn.Compare(b.CreadoEn); c != 0 {
			return c
		}
		return strings.Compare(a.ID.String(), b.ID.String())
	})

	total := len(coincidentes)
	if f.Desplazamiento >= total {
		return []domain.Profesional{}, total, nil
	}

	fin := total
	if f.Limite > 0 && f.Desplazamiento+f.Limite < total {
		fin = f.Desplazamiento + f.Limite
	}
	return coincidentes[f.Desplazamiento:fin], total, nil
}

func (r *Profesional) Actualizar(_ context.Context, p domain.Profesional) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[p.ID]; !existe {
		return domain.ErrNoEncontrado
	}
	if err := r.conflicto(p); err != nil {
		return err
	}
	r.datos[p.ID] = p.Clonar()
	return nil
}

func coincide(p domain.Profesional, f repository.Filtro) bool {
	if f.Especialidad != nil && p.Especialidad != *f.Especialidad {
		return false
	}
	if f.Estado != nil && p.Estado != *f.Estado {
		return false
	}
	if f.Zona != nil && domain.Normalizar(p.Zona) != domain.Normalizar(*f.Zona) {
		return false
	}
	if f.Busqueda != nil {
		q := domain.Normalizar(*f.Busqueda)
		if q != "" && !strings.Contains(domain.Normalizar(p.NombreCompleto()), q) {
			return false
		}
	}
	return true
}
