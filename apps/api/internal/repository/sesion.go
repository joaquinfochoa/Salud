package repository

import (
	"context"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Sesion es el almacenamiento de sesiones vigentes.
//
// La clave es el hash del token y no un ID propio: la única forma de llegar a
// una sesión es presentando el token, y guardar además un ID sería una segunda
// llave que nadie usa.
//
// No hay listado ni borrado masivo. "Cerrar todas mis sesiones" y la limpieza
// de vencidas son casos que todavía no existen; con PostgreSQL la limpieza es
// un DELETE por expira_en y no necesita pasar por esta interfaz.
type Sesion interface {
	Crear(ctx context.Context, s domain.Sesion) error

	// ObtenerPorTokenHash devuelve la sesión sin mirar si venció. Decidir eso
	// necesita un reloj, y el reloj vive en el servicio: un repositorio que
	// filtra por tiempo es un repositorio imposible de testear con fechas
	// fijas.
	ObtenerPorTokenHash(ctx context.Context, hash [32]byte) (domain.Sesion, error)

	// Eliminar es idempotente: borrar una sesión que no existe no es un error.
	Eliminar(ctx context.Context, hash [32]byte) error
}
