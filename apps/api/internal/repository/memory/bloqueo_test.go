package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func instante(t *testing.T, s string) time.Time {
	t.Helper()
	momento, err := time.ParseInLocation("2006-01-02 15:04", s, domain.ZonaHoraria)
	if err != nil {
		t.Fatalf("no se pudo parsear %q: %v", s, err)
	}
	return momento
}

func bloqueoDePrueba(t *testing.T, profesionalID uuid.UUID, desde, hasta string) domain.Bloqueo {
	t.Helper()
	return domain.Bloqueo{
		ID:            uuid.New(),
		ProfesionalID: profesionalID,
		Desde:         instante(t, desde),
		Hasta:         instante(t, hasta),
		Motivo:        "Vacaciones",
		CreadoEn:      instante(t, "2026-08-22 10:00"),
	}
}

func TestCrearYObtenerBloqueo(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	b := bloqueoDePrueba(t, uuid.New(), "2026-09-10 00:00", "2026-09-20 00:00")

	if err := repo.Crear(ctx, b); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenido, err := repo.ObtenerPorID(ctx, b.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID devolvió error: %v", err)
	}
	if obtenido.ID != b.ID || obtenido.Motivo != "Vacaciones" {
		t.Errorf("el bloqueo recuperado no coincide: %+v", obtenido)
	}
}

func TestObtenerBloqueoInexistente(t *testing.T) {
	_, err := NuevoBloqueo().ObtenerPorID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestEliminarBloqueo(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	b := bloqueoDePrueba(t, uuid.New(), "2026-09-10 00:00", "2026-09-20 00:00")

	if err := repo.Crear(ctx, b); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}
	if err := repo.Eliminar(ctx, b.ID); err != nil {
		t.Fatalf("Eliminar devolvió error: %v", err)
	}
	if _, err := repo.ObtenerPorID(ctx, b.ID); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("el bloqueo seguía existiendo tras eliminarlo")
	}
	if err := repo.Eliminar(ctx, b.ID); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("eliminar algo inexistente debía dar ErrNoEncontrado, dio %v", err)
	}
}

func TestListarBloqueosPorRango(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	profesionalID := uuid.New()

	dentro := bloqueoDePrueba(t, profesionalID, "2026-09-12 00:00", "2026-09-14 00:00")
	pisaElArranque := bloqueoDePrueba(t, profesionalID, "2026-09-05 00:00", "2026-09-11 00:00")
	pisaElFinal := bloqueoDePrueba(t, profesionalID, "2026-09-19 00:00", "2026-09-25 00:00")
	loContiene := bloqueoDePrueba(t, profesionalID, "2026-08-01 00:00", "2026-10-01 00:00")
	muyAntes := bloqueoDePrueba(t, profesionalID, "2026-07-01 00:00", "2026-07-10 00:00")
	muyDespues := bloqueoDePrueba(t, profesionalID, "2026-11-01 00:00", "2026-11-10 00:00")

	for _, b := range []domain.Bloqueo{dentro, pisaElArranque, pisaElFinal, loContiene, muyAntes, muyDespues} {
		if err := repo.Crear(ctx, b); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	obtenidos, err := repo.ListarDeProfesional(ctx, profesionalID,
		instante(t, "2026-09-10 00:00"), instante(t, "2026-09-20 00:00"))
	if err != nil {
		t.Fatalf("ListarDeProfesional devolvió error: %v", err)
	}

	// los cuatro que pisan el rango, aunque sea en parte; los dos de afuera no
	if len(obtenidos) != 4 {
		t.Fatalf("se obtuvieron %d bloqueos, se esperaban 4", len(obtenidos))
	}

	// y salen ordenados por fecha de inicio
	for i := 1; i < len(obtenidos); i++ {
		if obtenidos[i].Desde.Before(obtenidos[i-1].Desde) {
			t.Errorf("los bloqueos no salieron ordenados: %v antes que %v",
				obtenidos[i-1].Desde, obtenidos[i].Desde)
		}
	}
}

func TestBloqueosNoSeMezclanEntreProfesionales(t *testing.T) {
	ctx := context.Background()
	repo := NuevoBloqueo()
	uno, otro := uuid.New(), uuid.New()

	if err := repo.Crear(ctx, bloqueoDePrueba(t, uno, "2026-09-10 00:00", "2026-09-20 00:00")); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenidos, _ := repo.ListarDeProfesional(ctx, otro,
		instante(t, "2026-09-01 00:00"), instante(t, "2026-10-01 00:00"))
	if len(obtenidos) != 0 {
		t.Errorf("el otro profesional no tenía bloqueos y devolvió %d", len(obtenidos))
	}
}
