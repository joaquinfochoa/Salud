package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

func nuevaAuthDePrueba() *Autenticacion {
	return NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion())
}

func entradaAuthValida() domain.EntradaUsuario {
	return domain.EntradaUsuario{
		Email:      "juan@ejemplo.com",
		Contrasena: "unaclave8",
		Nombre:     "Juan",
		Apellido:   "Pérez",
		Telefono:   "11 1234-5678",
	}
}

func TestRegistrarDevuelveSesion(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	u, token, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	if token == "" {
		t.Fatal("Registrar no devolvió token: registrarse tiene que loguear")
	}

	resuelto, err := auth.ResolverSesion(ctx, token)
	if err != nil {
		t.Fatalf("ResolverSesion: %v", err)
	}
	if resuelto.ID != u.ID {
		t.Errorf("ID = %v, se esperaba %v", resuelto.ID, u.ID)
	}
}

func TestRegistrarEmailDuplicado(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, _, err := auth.Registrar(ctx, entradaAuthValida()); err != nil {
		t.Fatalf("primer Registrar: %v", err)
	}

	// mismo email escrito distinto: el VO Email lo normaliza, así que tiene
	// que chocar igual
	entrada := entradaAuthValida()
	entrada.Email = "JUAN@Ejemplo.com"

	_, _, err := auth.Registrar(ctx, entrada)
	if !errors.Is(err, domain.ErrEmailEnUso) {
		t.Errorf("se esperaba ErrEmailEnUso, se obtuvo %v", err)
	}
}

func TestIniciarSesion(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	u, _, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	mismo, token, err := auth.IniciarSesion(ctx, "juan@ejemplo.com", "unaclave8")
	if err != nil {
		t.Fatalf("IniciarSesion: %v", err)
	}
	if mismo.ID != u.ID {
		t.Errorf("ID = %v, se esperaba %v", mismo.ID, u.ID)
	}
	if token == "" {
		t.Error("IniciarSesion no devolvió token")
	}
}

// Un solo error para las dos formas de fallar. Si alguna vez se separan, el
// login se convierte en un oráculo de qué emails están registrados.
func TestIniciarSesionFallaIgualPorAmbosLados(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, _, err := auth.Registrar(ctx, entradaAuthValida()); err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	casos := []struct {
		nombre     string
		email      string
		contrasena string
	}{
		{"email inexistente", "otro@ejemplo.com", "unaclave8"},
		{"contrasena incorrecta", "juan@ejemplo.com", "otraclave"},
		{"email mal formado", "no-es-un-email", "unaclave8"},
		{"contrasena vacia", "juan@ejemplo.com", ""},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, _, err := auth.IniciarSesion(ctx, c.email, c.contrasena)
			if !errors.Is(err, domain.ErrCredencialesInvalidas) {
				t.Errorf("se esperaba ErrCredencialesInvalidas, se obtuvo %v", err)
			}
		})
	}
}

func TestIniciarSesionNormalizaElEmail(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, _, err := auth.Registrar(ctx, entradaAuthValida()); err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	if _, _, err := auth.IniciarSesion(ctx, "  JUAN@Ejemplo.COM ", "unaclave8"); err != nil {
		t.Errorf("el login debería aceptar el email escrito distinto: %v", err)
	}
}

func TestCerrarSesion(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	_, token, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	if err := auth.CerrarSesion(ctx, token); err != nil {
		t.Fatalf("CerrarSesion: %v", err)
	}
	if _, err := auth.ResolverSesion(ctx, token); err == nil {
		t.Error("la sesión sigue viva después del logout")
	}
}

func TestCerrarSesionEsIdempotente(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if err := auth.CerrarSesion(ctx, "token-que-no-existe"); err != nil {
		t.Errorf("cerrar una sesión inexistente devolvió %v, se esperaba nil", err)
	}
}

// Criterio de aceptación 6: una sesión vencida deja de valer sola, sin que
// nadie la haya borrado.
func TestResolverSesionVencida(t *testing.T) {
	ctx := context.Background()
	reloj := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	auth := NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion(),
		ConRelojAuth(func() time.Time { return reloj }))

	_, token, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	// justo antes de vencer sigue viva
	reloj = reloj.Add(domain.DuracionSesion - time.Second)
	if _, err := auth.ResolverSesion(ctx, token); err != nil {
		t.Fatalf("la sesión no debería haber vencido todavía: %v", err)
	}

	// pasado el vencimiento, no
	reloj = reloj.Add(2 * time.Second)
	if _, err := auth.ResolverSesion(ctx, token); err == nil {
		t.Error("una sesión vencida siguió resolviendo")
	}
}

func TestResolverSesionConTokenInventado(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	if _, err := auth.ResolverSesion(ctx, "inventado"); err == nil {
		t.Error("un token inventado resolvió una sesión")
	}
	if _, err := auth.ResolverSesion(ctx, ""); err == nil {
		t.Error("un token vacío resolvió una sesión")
	}
}

func TestActualizarPerfil(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	u, _, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	nuevo, err := auth.ActualizarPerfil(ctx, u.ID, domain.EntradaPerfil{
		Email:    "juan.nuevo@ejemplo.com",
		Nombre:   "Juan Carlos",
		Apellido: "Pérez",
		Telefono: "011 15 5555-4444",
	})
	if err != nil {
		t.Fatalf("ActualizarPerfil: %v", err)
	}
	if nuevo.Email != "juan.nuevo@ejemplo.com" || nuevo.Telefono != "+5491155554444" {
		t.Errorf("no se guardaron los cambios: %+v", nuevo)
	}

	// Y se puede entrar con el email nuevo, que es lo que hace que el cambio
	// sea real y no solo un campo distinto en la respuesta.
	if _, _, err := auth.IniciarSesion(ctx, "juan.nuevo@ejemplo.com", "unaclave8"); err != nil {
		t.Errorf("no se puede entrar con el email nuevo: %v", err)
	}
	if _, _, err := auth.IniciarSesion(ctx, "juan@ejemplo.com", "unaclave8"); err == nil {
		t.Error("todavía se puede entrar con el email viejo")
	}
}

// El email es con lo que se entra: dos cuentas con el mismo dejarían una
// inaccesible. Es el mismo 409 que devuelve el registro.
func TestActualizarPerfilRechazaUnEmailAjeno(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	uno, _, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	otra := entradaAuthValida()
	otra.Email = "otra@ejemplo.com"
	if _, _, err := auth.Registrar(ctx, otra); err != nil {
		t.Fatalf("Registrar: %v", err)
	}

	_, err = auth.ActualizarPerfil(ctx, uno.ID, domain.EntradaPerfil{
		Email:    "otra@ejemplo.com",
		Nombre:   "Juan",
		Apellido: "Pérez",
		Telefono: "11 1234-5678",
	})
	if !errors.Is(err, domain.ErrEmailEnUso) {
		t.Fatalf("se esperaba ErrEmailEnUso, vino %v", err)
	}

	// Y el usuario quedó como estaba: un rechazo no deja la cuenta a medias.
	sigue, err := auth.usuarios.ObtenerPorID(ctx, uno.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if sigue.Email != "juan@ejemplo.com" {
		t.Errorf("el email quedó en %q", sigue.Email)
	}
}

// Guardar sin cambiar el email no puede fallar por "email en uso": el propio no
// choca consigo mismo.
func TestActualizarPerfilConElMismoEmail(t *testing.T) {
	ctx := context.Background()
	auth := nuevaAuthDePrueba()

	u, _, err := auth.Registrar(ctx, entradaAuthValida())
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	if _, err := auth.ActualizarPerfil(ctx, u.ID, domain.EntradaPerfil{
		Email:    "juan@ejemplo.com",
		Nombre:   "Juan",
		Apellido: "Pérez",
		Telefono: "11 1234-5678",
	}); err != nil {
		t.Fatalf("ActualizarPerfil con el mismo email: %v", err)
	}
}
