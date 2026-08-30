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

	// maxIntentosSlug acota el reintento del alta. El bucle ya termina solo,
	// porque cada vuelta prueba un sufijo distinto y hay una cantidad finita de
	// homónimos; el techo está para que un repositorio con un bug que devuelva
	// ErrSlugEnUso siempre no deje el request girando para siempre.
	maxIntentosSlug = 100
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

// verificarPropiedad es la única implementación de "este perfil es tuyo".
//
// Vive en el servicio y no en el handler a propósito: acá la cubren los tests
// sin levantar un servidor HTTP, y cualquier consumidor futuro —un comando de
// consola, una cola— pasa por la misma regla en vez de reimplementarla.
func verificarPropiedad(p domain.Profesional, usuarioID uuid.UUID) error {
	if p.UsuarioID != usuarioID {
		return domain.ErrNoAutorizado
	}
	return nil
}

func (s *Profesional) Crear(ctx context.Context, usuarioID uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	// Camino rápido, con la misma salvedad que el de la matrícula: lee y suelta
	// el lock, así que la garantía la da repo.Crear con ErrUsuarioEnUso.
	switch _, err := s.repo.ObtenerPorUsuarioID(ctx, usuarioID); {
	case err == nil:
		return domain.Profesional{}, domain.ErrYaTienePerfil
	case !errors.Is(err, domain.ErrNoEncontrado):
		return domain.Profesional{}, err
	}

	p, err := domain.NuevoProfesional(entrada, usuarioID, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}

	// Camino rápido. La matrícula es la única identidad real de una persona en
	// este sistema; el parser ya normalizó "M.N. 98.234" y "MN 98234" a lo
	// mismo, así que esta comparación atrapa los duplicados escritos distinto y
	// devuelve el 409 sin llegar a intentar la escritura.
	//
	// No es la garantía, y por eso el repositorio vuelve a chequear: entre este
	// Obtener y el Crear de abajo se suelta el lock de lectura y se toma el de
	// escritura, y en el medio entra otra alta con la misma matrícula. Las dos
	// pasarían por acá. Los dos chequeos existen a propósito: este da el error
	// lindo en el caso común, el del repositorio es el que realmente sostiene
	// la invariante.
	if err := s.verificarMatriculaLibre(ctx, p.Matricula, uuid.Nil); err != nil {
		return domain.Profesional{}, err
	}

	// Un choque de slug nunca es un error para el cliente: dos "Martín
	// González" son perfectamente posibles y no hay razón para rechazar al
	// segundo. Se reintenta con el sufijo siguiente hasta que entre, que es la
	// misma forma en que habría que resolverlo contra una constraint UNIQUE de
	// PostgreSQL. La secuencia visible es la de siempre: base, base-2, base-3.
	base := p.Slug
	for intento := 1; intento <= maxIntentosSlug; intento++ {
		if intento > 1 {
			p.Slug = fmt.Sprintf("%s-%d", base, intento)
		}

		err := s.repo.Crear(ctx, p)
		if errors.Is(err, domain.ErrSlugEnUso) {
			continue
		}
		if err != nil {
			return domain.Profesional{}, err
		}
		return p, nil
	}
	return domain.Profesional{}, fmt.Errorf("no se encontró un slug libre para %q en %d intentos", base, maxIntentosSlug)
}

func (s *Profesional) ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Profesional, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *Profesional) ObtenerPorSlug(ctx context.Context, slug string) (domain.Profesional, error) {
	return s.repo.ObtenerPorSlug(ctx, slug)
}

// ObtenerPorUsuarioID devuelve el perfil de un usuario, o ErrNoEncontrado si
// no tiene. No tener perfil no es un error del sistema: es la mitad de los
// usuarios, y el que llama decide qué significa.
func (s *Profesional) ObtenerPorUsuarioID(ctx context.Context, usuarioID uuid.UUID) (domain.Profesional, error) {
	return s.repo.ObtenerPorUsuarioID(ctx, usuarioID)
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
//
// Es un chequeo previo, no una garantía: lee y suelta el lock. Quien sostiene
// la invariante es la escritura del repositorio.
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

// Actualizar reemplaza los campos editables. Funciona también sobre profesionales
// dados de baja: editar los datos de alguien inactivo no tiene por qué
// bloquearse, y no cambia su estado.
func (s *Profesional) Actualizar(ctx context.Context, usuarioID, id uuid.UUID, entrada domain.EntradaProfesional) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}
	if err := verificarPropiedad(actual, usuarioID); err != nil {
		return domain.Profesional{}, err
	}

	actualizado, err := actual.AplicarCambios(entrada, s.ahora())
	if err != nil {
		return domain.Profesional{}, err
	}

	// Mismo camino rápido que en el alta, y con la misma salvedad: el que
	// garantiza que la matrícula no se duplique es repo.Actualizar, que la
	// chequea bajo el lock de escritura excluyendo a este mismo registro.
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
func (s *Profesional) DarDeBaja(ctx context.Context, usuarioID, id uuid.UUID) error {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}
	if err := verificarPropiedad(actual, usuarioID); err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, actual.DarDeBaja(s.ahora()))
}

func (s *Profesional) Reactivar(ctx context.Context, usuarioID, id uuid.UUID) (domain.Profesional, error) {
	actual, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return domain.Profesional{}, err
	}
	if err := verificarPropiedad(actual, usuarioID); err != nil {
		return domain.Profesional{}, err
	}

	reactivado := actual.Reactivar(s.ahora())
	if err := s.repo.Actualizar(ctx, reactivado); err != nil {
		return domain.Profesional{}, err
	}
	return reactivado, nil
}
