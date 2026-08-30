package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func TestSesionCrearYObtener(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	s, token, err := domain.NuevaSesion(uuid.New(), time.Now().UTC())
	if err != nil {
		t.Fatalf("domain.NuevaSesion: %v", err)
	}
	if err := r.Crear(ctx, s); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	obtenida, err := r.ObtenerPorTokenHash(ctx, domain.HashearToken(token))
	if err != nil {
		t.Fatalf("ObtenerPorTokenHash: %v", err)
	}
	if obtenida.UsuarioID != s.UsuarioID {
		t.Errorf("UsuarioID = %v, se esperaba %v", obtenida.UsuarioID, s.UsuarioID)
	}
}

func TestSesionEliminar(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	s, token, err := domain.NuevaSesion(uuid.New(), time.Now().UTC())
	if err != nil {
		t.Fatalf("domain.NuevaSesion: %v", err)
	}
	if err := r.Crear(ctx, s); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	if err := r.Eliminar(ctx, domain.HashearToken(token)); err != nil {
		t.Fatalf("Eliminar: %v", err)
	}
	if _, err := r.ObtenerPorTokenHash(ctx, domain.HashearToken(token)); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("después de eliminar se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El logout tiene que poder repetirse sin explotar: un cliente que reintenta
// un DELETE no está haciendo nada malo.
func TestSesionEliminarEsIdempotente(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	if err := r.Eliminar(ctx, domain.HashearToken("no-existe")); err != nil {
		t.Errorf("eliminar una sesión inexistente devolvió %v, se esperaba nil", err)
	}
}

func TestSesionTokenDesconocido(t *testing.T) {
	ctx := context.Background()
	r := NuevaSesion()

	if _, err := r.ObtenerPorTokenHash(ctx, domain.HashearToken("inventado")); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}
