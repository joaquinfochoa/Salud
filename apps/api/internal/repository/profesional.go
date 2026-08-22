package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Filtro son los criterios del listado. Los punteros distinguen "no filtrar
// por este campo" de "filtrar por el valor cero", que con valores planos no se
// puede: un Estado("") sería indistinguible de "sin filtro".
type Filtro struct {
	Especialidad   *domain.Especialidad
	Zona           *string
	Estado         *domain.Estado
	Busqueda       *string // busca en nombre y apellido, sin distinguir acentos
	Limite         int     // cero o menos significa sin límite; el servicio aplica el valor por defecto y el techo
	Desplazamiento int
}

// Profesional es el punto de cambio a PostgreSQL. Cuando exista la
// implementación con base de datos, migrar es cambiar una línea de main.go y
// nada más.
//
// Todos los métodos reciben context.Context aunque la implementación en
// memoria lo ignore: agregarlo después obligaría a tocar todas las firmas.
//
// No hay borrado físico: la baja es lógica y se hace con Actualizar.
type Profesional interface {
	Crear(ctx context.Context, p domain.Profesional) error
	ObtenerPorID(ctx context.Context, id uuid.UUID) (domain.Profesional, error)
	ObtenerPorSlug(ctx context.Context, slug string) (domain.Profesional, error)
	ObtenerPorMatricula(ctx context.Context, m domain.Matricula) (domain.Profesional, error)
	Listar(ctx context.Context, f Filtro) ([]domain.Profesional, int, error)
	Actualizar(ctx context.Context, p domain.Profesional) error
}
