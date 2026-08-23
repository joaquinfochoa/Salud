package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func momento(t *testing.T, s string) time.Time {
	t.Helper()
	instante, err := time.ParseInLocation("2006-01-02 15:04", s, ZonaHoraria)
	if err != nil {
		t.Fatalf("no se pudo parsear %q: %v", s, err)
	}
	return instante
}

func TestNuevoBloqueoValido(t *testing.T) {
	ahora := momento(t, "2026-08-22 10:00")
	profesionalID := uuid.New()

	b, err := NuevoBloqueo(profesionalID, EntradaBloqueo{
		Desde:  momento(t, "2026-09-10 00:00"),
		Hasta:  momento(t, "2026-09-20 23:59"),
		Motivo: "  Vacaciones  ",
	}, ahora)
	if err != nil {
		t.Fatalf("NuevoBloqueo devolvió error: %v", err)
	}

	if b.ProfesionalID != profesionalID {
		t.Error("el bloqueo no quedó atado a su profesional")
	}
	if b.Motivo != "Vacaciones" {
		t.Errorf("Motivo = %q, se esperaba sin espacios alrededor", b.Motivo)
	}
	if !b.CreadoEn.Equal(ahora) {
		t.Error("CreadoEn tenía que ser el ahora recibido")
	}
	if b.ID == uuid.Nil {
		t.Error("el ID tenía que generarse")
	}
}

func TestNuevoBloqueoSinMotivo(t *testing.T) {
	// el motivo es opcional: bloquear sin explicar por qué es legítimo
	_, err := NuevoBloqueo(uuid.New(), EntradaBloqueo{
		Desde: momento(t, "2026-09-10 00:00"),
		Hasta: momento(t, "2026-09-11 00:00"),
	}, momento(t, "2026-08-22 10:00"))
	if err != nil {
		t.Errorf("un bloqueo sin motivo es válido, devolvió: %v", err)
	}
}

func TestNuevoBloqueoInvalido(t *testing.T) {
	ahora := momento(t, "2026-08-22 10:00")

	casos := []struct {
		nombre        string
		entrada       EntradaBloqueo
		campoEsperado string
	}{
		{
			"desde vacio",
			EntradaBloqueo{Hasta: momento(t, "2026-09-11 00:00")},
			"desde",
		},
		{
			"hasta vacio",
			EntradaBloqueo{Desde: momento(t, "2026-09-10 00:00")},
			"hasta",
		},
		{
			"hasta antes que desde",
			EntradaBloqueo{Desde: momento(t, "2026-09-20 00:00"), Hasta: momento(t, "2026-09-10 00:00")},
			"hasta",
		},
		{
			"desde igual a hasta",
			EntradaBloqueo{Desde: momento(t, "2026-09-10 00:00"), Hasta: momento(t, "2026-09-10 00:00")},
			"hasta",
		},
		{
			"enteramente en el pasado",
			EntradaBloqueo{Desde: momento(t, "2026-07-01 00:00"), Hasta: momento(t, "2026-07-10 00:00")},
			"hasta",
		},
		{
			"motivo demasiado largo",
			EntradaBloqueo{
				Desde:  momento(t, "2026-09-10 00:00"),
				Hasta:  momento(t, "2026-09-11 00:00"),
				Motivo: strings.Repeat("a", 201),
			},
			"motivo",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := NuevoBloqueo(uuid.New(), caso.entrada, ahora)

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
			}
			encontrado := false
			for _, campo := range verr.Campos {
				if campo.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestNuevoBloqueoQueEmpezoPeroNoTermino(t *testing.T) {
	// unas vacaciones que arrancaron ayer y siguen hasta la semana que viene
	// son perfectamente válidas: lo que no sirve es bloquear algo ya terminado
	ahora := momento(t, "2026-08-22 10:00")

	_, err := NuevoBloqueo(uuid.New(), EntradaBloqueo{
		Desde: momento(t, "2026-08-20 00:00"),
		Hasta: momento(t, "2026-08-30 00:00"),
	}, ahora)
	if err != nil {
		t.Errorf("un bloqueo en curso es válido, devolvió: %v", err)
	}
}

func TestBloqueoSeSolapaCon(t *testing.T) {
	b := Bloqueo{
		Desde: momento(t, "2026-08-24 09:50"),
		Hasta: momento(t, "2026-08-24 10:00"),
	}

	casos := []struct {
		nombre   string
		inicio   string
		fin      string
		esperado bool
	}{
		{"adentro", "2026-08-24 09:52", "2026-08-24 09:55", true},
		{"lo contiene", "2026-08-24 09:00", "2026-08-24 11:00", true},
		{"pisa el arranque", "2026-08-24 09:45", "2026-08-24 09:55", true},
		{"pisa el final", "2026-08-24 09:55", "2026-08-24 10:05", true},
		// los dos bordes: intervalos semiabiertos [inicio, fin)
		{"termina justo cuando el bloqueo empieza", "2026-08-24 09:00", "2026-08-24 09:50", false},
		{"empieza justo cuando el bloqueo termina", "2026-08-24 10:00", "2026-08-24 10:50", false},
		{"muy antes", "2026-08-24 07:00", "2026-08-24 08:00", false},
		{"muy despues", "2026-08-24 15:00", "2026-08-24 16:00", false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			obtenido := b.SeSolapaCon(momento(t, caso.inicio), momento(t, caso.fin))
			if obtenido != caso.esperado {
				t.Errorf("SeSolapaCon(%s, %s) = %v, se esperaba %v", caso.inicio, caso.fin, obtenido, caso.esperado)
			}
		})
	}
}
