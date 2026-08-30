package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

// Los métodos que mutan rechazan a un usuario que no es el dueño. Una tabla y
// no tests sueltos: si mañana se agrega otro método que muta, agregar la fila
// es más difícil de olvidar que escribir el test.
func TestSoloElDuenoPuedeMutar(t *testing.T) {
	ctx := context.Background()
	svc := NuevoProfesional(memory.NuevoProfesional())

	dueno := uuid.New()
	intruso := uuid.New()

	p, err := svc.Crear(ctx, dueno, entradaValida())
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	casos := []struct {
		nombre string
		correr func(usuarioID uuid.UUID) error
	}{
		{"Actualizar", func(u uuid.UUID) error {
			_, err := svc.Actualizar(ctx, u, p.ID, entradaValida())
			return err
		}},
		{"DarDeBaja", func(u uuid.UUID) error {
			return svc.DarDeBaja(ctx, u, p.ID)
		}},
		{"Reactivar", func(u uuid.UUID) error {
			_, err := svc.Reactivar(ctx, u, p.ID)
			return err
		}},
	}

	for _, c := range casos {
		t.Run(c.nombre+" con intruso", func(t *testing.T) {
			if err := c.correr(intruso); !errors.Is(err, domain.ErrNoAutorizado) {
				t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
			}
		})
		t.Run(c.nombre+" con dueno", func(t *testing.T) {
			if err := c.correr(dueno); err != nil {
				t.Errorf("el dueño no debería recibir error: %v", err)
			}
		})
	}
}

// Un perfil inexistente da ErrNoEncontrado y no ErrNoAutorizado, incluso para
// un intruso. Al revés sería un oráculo: probando IDs, un 403 confirmaría que
// ese perfil existe.
func TestPerfilInexistenteDaNoEncontrado(t *testing.T) {
	ctx := context.Background()
	svc := NuevoProfesional(memory.NuevoProfesional())

	if err := svc.DarDeBaja(ctx, uuid.New(), uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestUnUsuarioNoPuedeTenerDosPerfiles(t *testing.T) {
	ctx := context.Background()
	svc := NuevoProfesional(memory.NuevoProfesional())
	usuarioID := uuid.New()

	if _, err := svc.Crear(ctx, usuarioID, entradaValida()); err != nil {
		t.Fatalf("primer Crear: %v", err)
	}

	otra := entradaValida()
	otra.Matricula = "MN 777888"

	if _, err := svc.Crear(ctx, usuarioID, otra); !errors.Is(err, domain.ErrYaTienePerfil) {
		t.Errorf("se esperaba ErrYaTienePerfil, se obtuvo %v", err)
	}
}

func TestSoloElDuenoPuedeTocarLaAgenda(t *testing.T) {
	ctx := context.Background()
	repoProf := memory.NuevoProfesional()
	svcProf := NuevoProfesional(repoProf)
	svc := NuevaAgenda(repoProf, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo())

	dueno := uuid.New()
	intruso := uuid.New()

	p, err := svcProf.Crear(ctx, dueno, entradaValida())
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	t.Run("ReemplazarHorarios", func(t *testing.T) {
		if _, err := svc.ReemplazarHorarios(ctx, intruso, p.ID, nil); !errors.Is(err, domain.ErrNoAutorizado) {
			t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
		}
	})

	t.Run("CrearBloqueo", func(t *testing.T) {
		if _, err := svc.CrearBloqueo(ctx, intruso, p.ID, domain.EntradaBloqueo{}); !errors.Is(err, domain.ErrNoAutorizado) {
			t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
		}
	})

	t.Run("EliminarBloqueo", func(t *testing.T) {
		if err := svc.EliminarBloqueo(ctx, intruso, p.ID, uuid.New()); !errors.Is(err, domain.ErrNoAutorizado) {
			t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
		}
	})

	// La autorización va antes que la validación. Un intruso que manda datos
	// inválidos tiene que recibir 403, no un 422 que le confirme que el perfil
	// existe y de paso le enseñe el esquema del cuerpo.
	t.Run("el 403 gana al 422", func(t *testing.T) {
		_, err := svc.CrearBloqueo(ctx, intruso, p.ID, domain.EntradaBloqueo{})
		var verr domain.ErrorValidacion
		if errors.As(err, &verr) {
			t.Error("un intruso recibió un error de validación en vez de 403")
		}
	})
}

// La agenda pública se sigue leyendo sin dueño. Es el criterio de aceptación 4
// medido en la capa de servicio.
func TestLasLecturasDeAgendaNoPidenDueno(t *testing.T) {
	ctx := context.Background()
	repoProf := memory.NuevoProfesional()
	svcProf := NuevoProfesional(repoProf)
	svc := NuevaAgenda(repoProf, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo())

	p, err := svcProf.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	if _, err := svc.ListarHorarios(ctx, p.ID); err != nil {
		t.Errorf("ListarHorarios sin sesión falló: %v", err)
	}
	if _, err := svc.ListarBloqueos(ctx, p.ID, nil, nil); err != nil {
		t.Errorf("ListarBloqueos sin sesión falló: %v", err)
	}
}
