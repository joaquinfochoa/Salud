package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Turno guarda los turnos reservados.
//
// No hay Eliminar: cancelar es un cambio de estado, no un borrado. El turno
// cancelado sigue siendo parte del historial de las dos partes, y borrarlo
// dejaría al paciente sin forma de ver que le cancelaron algo.
//
// Contrato de unicidad, igual que la matrícula: lo garantiza la escritura, no
// el que llama. Chequear la disponibilidad en una llamada y escribir en otra
// son dos operaciones, y entre las dos entra otro paciente. En PostgreSQL esto
// son dos constraints de exclusión sobre el rango temporal —una por
// profesional, otra por paciente, ambas filtradas por estado='reservado'— y
// traducir su violación a estos centinelas es todo lo que tiene que hacer esa
// implementación:
//
//   - Crear devuelve ErrHuecoTomado si otro turno activo del mismo profesional
//     pisa el intervalo, ErrPacienteSolapado si el que pisa es del mismo
//     paciente, o ErrIDEnUso.
//   - Actualizar devuelve ErrNoEncontrado.
type Turno interface {
	Crear(ctx context.Context, t domain.Turno) error
	ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Turno, error)
	Actualizar(ctx context.Context, t domain.Turno) error

	// Los dos listados devuelven los turnos que pisan el intervalo semiabierto
	// [desde, hasta), cancelados incluidos. Quien necesite solo los activos
	// filtra con EstaActivo(): filtrar acá obligaría a tener dos métodos o un
	// parámetro booleano, y el que llama ya sabe qué quiere.
	ListarDeProfesional(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) ([]domain.Turno, error)
	ListarDePaciente(ctx context.Context, pacienteID uuid.UUID, desde, hasta time.Time) ([]domain.Turno, error)
}
