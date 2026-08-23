package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestCargarDefaults(t *testing.T) {
	// t.Setenv restaura el entorno al terminar el test
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar devolvió error: %v", err)
	}

	if cfg.Puerto != "8080" {
		t.Errorf("Puerto = %q, se esperaba 8080", cfg.Puerto)
	}
	if cfg.Entorno != "development" {
		t.Errorf("Entorno = %q, se esperaba development", cfg.Entorno)
	}
	if cfg.NivelLog != slog.LevelInfo {
		t.Errorf("NivelLog = %v, se esperaba info", cfg.NivelLog)
	}
	if cfg.TimeoutApagado != 10*time.Second {
		t.Errorf("TimeoutApagado = %v, se esperaba 10s", cfg.TimeoutApagado)
	}
	if !cfg.EsDesarrollo() {
		t.Error("con APP_ENV vacío tenía que ser desarrollo")
	}
}

func TestCargarDesdeElEntorno(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar devolvió error: %v", err)
	}

	if cfg.Puerto != "9000" {
		t.Errorf("Puerto = %q", cfg.Puerto)
	}
	if cfg.NivelLog != slog.LevelDebug {
		t.Errorf("NivelLog = %v, se esperaba debug", cfg.NivelLog)
	}
	if cfg.TimeoutApagado != 30*time.Second {
		t.Errorf("TimeoutApagado = %v", cfg.TimeoutApagado)
	}
	if cfg.EsDesarrollo() {
		t.Error("con APP_ENV=production no tenía que ser desarrollo")
	}
}

func TestCargarFallaRapidoConValoresInvalidos(t *testing.T) {
	casos := []struct {
		nombre string
		clave  string
		valor  string
	}{
		{"puerto no numerico", "PORT", "ocho-mil"},
		{"puerto fuera de rango", "PORT", "99999"},
		{"nivel de log desconocido", "LOG_LEVEL", "verbose"},
		{"timeout mal formado", "SHUTDOWN_TIMEOUT", "diez segundos"},
		{"entorno desconocido", "APP_ENV", "staging-raro"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(caso.clave, caso.valor)
			// mejor no arrancar que arrancar mal configurado
			if _, err := Cargar(); err == nil {
				t.Errorf("%s=%q debía fallar y no falló", caso.clave, caso.valor)
			}
		})
	}
}
