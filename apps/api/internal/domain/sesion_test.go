package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNuevaSesion(t *testing.T) {
	usuarioID := uuid.New()
	ahora := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	s, token, err := NuevaSesion(usuarioID, ahora)
	if err != nil {
		t.Fatalf("NuevaSesion devolvió error: %v", err)
	}
	if s.UsuarioID != usuarioID {
		t.Errorf("UsuarioID = %v, se esperaba %v", s.UsuarioID, usuarioID)
	}
	if !s.CreadaEn.Equal(ahora) {
		t.Errorf("CreadaEn = %v", s.CreadaEn)
	}
	if !s.ExpiraEn.Equal(ahora.Add(DuracionSesion)) {
		t.Errorf("ExpiraEn = %v, se esperaba %v", s.ExpiraEn, ahora.Add(DuracionSesion))
	}
	if s.TokenHash != HashearToken(token) {
		t.Error("el hash guardado no corresponde al token devuelto")
	}
	// 32 bytes en base64 sin relleno son 43 caracteres.
	if len(token) != 43 {
		t.Errorf("largo del token = %d, se esperaban 43", len(token))
	}
}

// El token es lo único que prueba la identidad. Dos sesiones con el mismo
// token significan que rand.Read no está haciendo lo que creemos.
func TestTokensDeSesionSonDistintos(t *testing.T) {
	vistos := make(map[string]bool, 100)
	for range 100 {
		_, token, err := NuevaSesion(uuid.New(), time.Now())
		if err != nil {
			t.Fatalf("NuevaSesion: %v", err)
		}
		if vistos[token] {
			t.Fatalf("token repetido: %q", token)
		}
		vistos[token] = true
	}
}

// Se guarda el hash, no el token: quien lea el almacenamiento no puede
// suplantar a nadie. Este test es el que se rompe si alguien "simplifica"
// guardando el token en claro.
func TestElTokenEnClaroNoQuedaEnLaSesion(t *testing.T) {
	s, token, err := NuevaSesion(uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}
	if string(s.TokenHash[:]) == token {
		t.Error("la sesión guarda el token en claro")
	}
}

func TestSesionEstaVencida(t *testing.T) {
	ahora := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	s, _, err := NuevaSesion(uuid.New(), ahora)
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}

	casos := []struct {
		nombre   string
		momento  time.Time
		esperado bool
	}{
		{"recien creada", ahora, false},
		{"un dia despues", ahora.Add(24 * time.Hour), false},
		{"un instante antes de expirar", s.ExpiraEn.Add(-time.Nanosecond), false},
		// El borde es cerrado hacia afuera: en el instante exacto de
		// vencimiento la sesión ya no sirve. La alternativa deja viva una
		// sesión vencida durante un tick del reloj.
		{"en el instante exacto", s.ExpiraEn, true},
		{"un instante despues", s.ExpiraEn.Add(time.Nanosecond), true},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := s.EstaVencida(c.momento); got != c.esperado {
				t.Errorf("EstaVencida(%v) = %v, se esperaba %v", c.momento, got, c.esperado)
			}
		})
	}
}
