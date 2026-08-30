package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EntornoDesarrollo = "development"
	EntornoProduccion = "production"
)

// Config es todo lo que el binario necesita saber del entorno. Sin librería:
// son cuatro variables y un struct.
type Config struct {
	Puerto         string
	Entorno        string
	NivelLog       slog.Level
	TimeoutApagado time.Duration

	// CORSOrigenes son los orígenes que pueden llamar a la API desde un
	// browser. Vacío significa apagado, que es lo correcto por defecto: una
	// lista vacía no puede permitir a nadie por accidente.
	//
	// En desarrollo se llena solo con el front local; en producción hay que
	// declararla explícitamente.
	CORSOrigenes []string
}

func (c Config) EsDesarrollo() bool {
	return c.Entorno == EntornoDesarrollo
}

// Cargar lee el entorno y falla si algo está mal.
//
// Falla rápido a propósito: un servidor que arranca con una configuración
// inválida es peor que uno que no arranca, porque el problema aparece más
// tarde y en otro lado.
func Cargar() (Config, error) {
	cfg := Config{
		Puerto:         leerEntorno("PORT", "8080"),
		Entorno:        leerEntorno("APP_ENV", EntornoDesarrollo),
		TimeoutApagado: 10 * time.Second,
	}

	puerto, err := strconv.Atoi(cfg.Puerto)
	if err != nil || puerto < 1 || puerto > 65535 {
		return Config{}, fmt.Errorf("PORT inválido: %q", cfg.Puerto)
	}

	if cfg.Entorno != EntornoDesarrollo && cfg.Entorno != EntornoProduccion {
		return Config{}, fmt.Errorf("APP_ENV inválido: %q (debe ser %s o %s)", cfg.Entorno, EntornoDesarrollo, EntornoProduccion)
	}

	nivel, err := parsearNivelLog(leerEntorno("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.NivelLog = nivel

	cfg.CORSOrigenes = separarPorComas(os.Getenv("CORS_ORIGENES"))
	if len(cfg.CORSOrigenes) == 0 && cfg.EsDesarrollo() {
		// En desarrollo el front vive en el 3000. Es un default cómodo y sin
		// riesgo: solo aplica con APP_ENV=development.
		cfg.CORSOrigenes = []string{"http://localhost:3000"}
	}
	for _, o := range cfg.CORSOrigenes {
		if !strings.HasPrefix(o, "http://") && !strings.HasPrefix(o, "https://") {
			return Config{}, fmt.Errorf("CORS_ORIGENES inválido: %q (tiene que incluir el esquema, por ejemplo https://salud.app)", o)
		}
		if strings.HasSuffix(o, "/") {
			// El header Origin que manda el browser nunca trae barra final, así
			// que un valor con barra no matchearía nunca y el front fallaría
			// sin ningún error visible del lado del servidor.
			return Config{}, fmt.Errorf("CORS_ORIGENES inválido: %q (sin barra final)", o)
		}
	}

	if crudo := os.Getenv("SHUTDOWN_TIMEOUT"); crudo != "" {
		d, err := time.ParseDuration(crudo)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT inválido: %q (ejemplo válido: 30s)", crudo)
		}
		cfg.TimeoutApagado = d
	}

	return cfg, nil
}

// separarPorComas parte una lista del entorno. "a, b" es una forma
// perfectamente razonable de escribirla, así que se recortan los espacios.
func separarPorComas(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	partes := strings.Split(s, ",")
	salida := make([]string, 0, len(partes))
	for _, p := range partes {
		if v := strings.TrimSpace(p); v != "" {
			salida = append(salida, v)
		}
	}
	return salida
}

func leerEntorno(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}

func parsearNivelLog(crudo string) (slog.Level, error) {
	switch crudo {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL inválido: %q (debe ser debug, info, warn o error)", crudo)
	}
}
