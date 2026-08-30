package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func usuarioDePrueba(t *testing.T, email string) domain.Usuario {
	t.Helper()
	u, err := domain.NuevoUsuario(domain.EntradaUsuario{
		Email:      email,
		Contrasena: "unaclave8",
		Nombre:     "Juan",
		Apellido:   "Pérez",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NuevoUsuario: %v", err)
	}
	return u
}

func TestUsuarioCrearYObtener(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()
	u := usuarioDePrueba(t, "juan@ejemplo.com")

	if err := r.Crear(ctx, u); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	porID, err := r.ObtenerPorID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if porID.Email != u.Email {
		t.Errorf("Email = %q, se esperaba %q", porID.Email, u.Email)
	}

	porEmail, err := r.ObtenerPorEmail(ctx, u.Email)
	if err != nil {
		t.Fatalf("ObtenerPorEmail: %v", err)
	}
	if porEmail.ID != u.ID {
		t.Errorf("ID = %v, se esperaba %v", porEmail.ID, u.ID)
	}
}

func TestUsuarioEmailDuplicado(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()

	if err := r.Crear(ctx, usuarioDePrueba(t, "juan@ejemplo.com")); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// Otro usuario, mismo email: es otro ID, así que el mapa no lo detecta.
	// Lo tiene que atrapar el chequeo de unicidad bajo el lock de escritura.
	err := r.Crear(ctx, usuarioDePrueba(t, "juan@ejemplo.com"))
	if !errors.Is(err, domain.ErrEmailEnUso) {
		t.Errorf("se esperaba ErrEmailEnUso, se obtuvo %v", err)
	}
}

func TestUsuarioIDDuplicado(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()
	u := usuarioDePrueba(t, "juan@ejemplo.com")

	if err := r.Crear(ctx, u); err != nil {
		t.Fatalf("Crear: %v", err)
	}
	if err := r.Crear(ctx, u); !errors.Is(err, domain.ErrIDEnUso) {
		t.Errorf("se esperaba ErrIDEnUso, se obtuvo %v", err)
	}
}

func TestUsuarioNoEncontrado(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()

	if _, err := r.ObtenerPorID(ctx, uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("ObtenerPorID: se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
	if _, err := r.ObtenerPorEmail(ctx, "nadie@ejemplo.com"); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("ObtenerPorEmail: se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El repositorio guarda y devuelve copias. Sin esto, quien recibe un usuario
// puede mutarle el hash al que está guardado.
func TestUsuarioDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	r := NuevoUsuario()
	u := usuarioDePrueba(t, "juan@ejemplo.com")

	if err := r.Crear(ctx, u); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// mutar lo que se le pasó a Crear no tiene que afectar lo guardado
	u.Hash[0] = 'X'

	guardado, err := r.ObtenerPorID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if guardado.Hash[0] == 'X' {
		t.Fatal("Crear guardó el slice del llamador en vez de una copia")
	}

	// y mutar lo devuelto tampoco
	guardado.Hash[0] = 'Y'
	otra, err := r.ObtenerPorID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if otra.Hash[0] == 'Y' {
		t.Error("ObtenerPorID devolvió el slice guardado en vez de una copia")
	}
}
