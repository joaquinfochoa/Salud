package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

const (
	// LimitePorDefecto es cuántos profesionales devuelve el listado si el cliente
	// no pide un tamaño.
	LimitePorDefecto = 20

	// LimiteMaximo es el techo. Sin techo, un cliente puede pedir el padrón
	// entero en una llamada.
	LimiteMaximo = 100
)

// Profesional resuelve los casos de uso que necesitan mirar más de un
// profesional a la vez. Las reglas que se deciden con una sola entidad viven
// en el dominio, no acá.
type Profesional struct {
	repo repository.Profesional

	// ahora es inyectable para que los casos no dependan del reloj.
	ahora func() time.Time
}

func NuevoProfesional(repo repository.Profesional) *Profesional {
	return &Profesional{
		repo:  repo,
		ahora: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Profesional) Crear(ctx context.Context, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	p, err := domain.NuevoProfesional(entrada, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}

	// La matrícula es la única identidad real de una persona en este sistema.
	// El parser ya normalizó "M.N. 98.234" y "MN 98234" a lo mismo, así que
	// esta comparación atrapa los duplicados escritos distinto.
	if err := s.verificarMatriculaLibre(ctx, p.Matricula, uuid.Nil); err != nil {
		return domain.Profesional{}, err
	}

	slug, err := s.slugUnico(ctx, p.Slug)
	if err != nil {
		return domain.Profesional{}, err
	}
	p.Slug = slug

	if err := s.repo.Crear(ctx, p); err != nil {
		return domain.Profesional{}, err
	}
	return p, nil
}

func (s *Profesional) ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Profesional, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *Profesional) ObtenerPorSlug(ctx context.Context, slug string) (domain.Profesional, error) {
	return s.repo.ObtenerPorSlug(ctx, slug)
}

func (s *Profesional) Listar(ctx context.Context, f repository.Filtro) ([]domain.Profesional, int, error) {
	f = NormalizarFiltro(f)
	return s.repo.Listar(ctx, f)
}

// NormalizarFiltro aplica los límites y valores por defecto de la paginación.
//
// Es exportada a propósito: el handler necesita informar en la respuesta el
// límite que realmente se aplicó, no el que pidió el cliente. Si cada capa
// aplicara su propia versión de la regla, un `limite=5000` devolvería 100
// registros diciendo que devolvió 5000, y el que pagine sumando ese número se
// saltea 4900 sin que falle nada.
//
// Es idempotente: aplicarla dos veces da lo mismo que aplicarla una.
func NormalizarFiltro(f repository.Filtro) repository.Filtro {
	if f.Limite <= 0 {
		f.Limite = LimitePorDefecto
	}
	if f.Limite > LimiteMaximo {
		f.Limite = LimiteMaximo
	}
	if f.Desplazamiento < 0 {
		f.Desplazamiento = 0
	}
	// Por defecto solo los activos: un profesional dado de baja no tiene que
	// aparecer en una búsqueda de pacientes. Para verlos hay que pedirlos.
	if f.Estado == nil {
		activo := domain.EstadoActivo
		f.Estado = &activo
	}
	return f
}

// verificarMatriculaLibre falla si otro profesional ya tiene esa matrícula.
// excluir permite ignorar al propio profesional durante una edición.
func (s *Profesional) verificarMatriculaLibre(ctx context.Context, m domain.Matricula, excluir uuid.UUID) error {
	existente, err := s.repo.ObtenerPorMatricula(ctx, m)
	switch {
	case errors.Is(err, domain.ErrNoEncontrado):
		return nil
	case err != nil:
		return err
	case existente.ID == excluir:
		return nil
	default:
		return domain.ErrMatriculaEnUso
	}
}

// slugUnico resuelve las colisiones agregando un sufijo numérico.
//
// Nunca es un error para el cliente: dos "Martín González" son perfectamente
// posibles y no hay razón para rechazar al segundo.
func (s *Profesional) slugUnico(ctx context.Context, base string) (string, error) {
	candidato := base
	for i := 2; ; i++ {
		_, err := s.repo.ObtenerPorSlug(ctx, candidato)
		if errors.Is(err, domain.ErrNoEncontrado) {
			return candidato, nil
		}
		if err != nil {
			return "", err
		}
		candidato = fmt.Sprintf("%s-%d", base, i)
	}
}

// Actualizar reemplaza los campos editables. Funciona también sobre profesionales
// dados de baja: editar los datos de alguien inactivo no tiene por qué
// bloquearse, y no cambia su estado.
func (s *Profesional) Actualizar(ctx context.Context, id uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}

	actualizado, err := actual.AplicarCambios(entrada, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}

	if actualizado.Matricula != actual.Matricula {
		if err := s.verificarMatriculaLibre(ctx, actualizado.Matricula, id); err != nil {
			return domain.Profesional{}, err
		}
	}

	if err := s.repo.Actualizar(ctx, actualizado); err != nil {
		return domain.Profesional{}, err
	}
	return actualizado, nil
}

// DarDeBaja da de baja. No borra: los turnos, comprobantes y pagos históricos
// siguen apuntando a este registro, y sin él el comprobante que un paciente
// presentó para un reintegro queda huérfano.
func (s *Profesional) DarDeBaja(ctx context.Context, id uuid.UUID) error {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, actual.DarDeBaja(s.ahora()))
}

func (s *Profesional) Reactivar(ctx context.Context, id uuid.UUID) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}

	reactivado := actual.Reactivar(s.ahora())
	if err := s.repo.Actualizar(ctx, reactivado); err != nil {
		return domain.Profesional{}, err
	}
	return reactivado, nil
}
