package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func huecoDePrueba() Hueco {
	inicio := time.Date(2026, 9, 15, 10, 0, 0, 0, ZonaHoraria)
	return Hueco{Inicio: inicio, Fin: inicio.Add(50 * time.Minute), Modalidad: ModalidadTelemedicina}
}

func TestNuevoTurno(t *testing.T) {
	profesionalID, pacienteID := uuid.New(), uuid.New()
	hueco := huecoDePrueba()
	ahora := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	turno, err := NuevoTurno(profesionalID, pacienteID, hueco, "  dolor lumbar  ", ahora)
	if err != nil {
		t.Fatalf("NuevoTurno devolvió error: %v", err)
	}

	if turno.ProfesionalID != profesionalID || turno.PacienteID != pacienteID {
		t.Error("las dos partes del turno no quedaron guardadas")
	}
	if !turno.Inicio.Equal(hueco.Inicio) || !turno.Fin.Equal(hueco.Fin) {
		t.Errorf("intervalo = [%v, %v), se esperaba [%v, %v)", turno.Inicio, turno.Fin, hueco.Inicio, hueco.Fin)
	}
	// La modalidad se copia del hueco. Si el profesional después cambia el
	// bloque de presencial a telemedicina, este turno tiene que seguir
	// diciendo lo que se pactó.
	if turno.Modalidad != hueco.Modalidad {
		t.Errorf("Modalidad = %q, se esperaba %q", turno.Modalidad, hueco.Modalidad)
	}
	if turno.Estado != TurnoReservado {
		t.Errorf("Estado = %q, se esperaba reservado", turno.Estado)
	}
	if turno.Motivo != "dolor lumbar" {
		t.Errorf("Motivo = %q, se esperaba recortado", turno.Motivo)
	}
	if !turno.CreadoEn.Equal(ahora) {
		t.Errorf("CreadoEn = %v", turno.CreadoEn)
	}
	if turno.CanceladoEn != nil || turno.CanceladoPor != nil {
		t.Error("un turno nuevo no puede nacer cancelado")
	}
}

func TestNuevoTurnoSinMotivo(t *testing.T) {
	// El motivo es opcional: nadie tiene por qué justificar una consulta para
	// poder pedirla.
	if _, err := NuevoTurno(uuid.New(), uuid.New(), huecoDePrueba(), "", time.Now()); err != nil {
		t.Errorf("un turno sin motivo debería ser válido: %v", err)
	}
}

func TestNuevoTurnoMotivoDemasiadoLargo(t *testing.T) {
	_, err := NuevoTurno(uuid.New(), uuid.New(), huecoDePrueba(), strings.Repeat("a", 501), time.Now())

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %v", err)
	}
	if verr.Campos[0].Campo != "motivo" {
		t.Errorf("campo = %q, se esperaba motivo", verr.Campos[0].Campo)
	}
}

func TestCancelar(t *testing.T) {
	ahora := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	quienCancela := uuid.New()

	turno, err := NuevoTurno(uuid.New(), uuid.New(), huecoDePrueba(), "", ahora)
	if err != nil {
		t.Fatalf("NuevoTurno: %v", err)
	}

	cancelado, err := turno.Cancelar(quienCancela, ahora)
	if err != nil {
		t.Fatalf("Cancelar devolvió error: %v", err)
	}

	if cancelado.Estado != TurnoCancelado {
		t.Errorf("Estado = %q, se esperaba cancelado", cancelado.Estado)
	}
	if cancelado.CanceladoEn == nil || !cancelado.CanceladoEn.Equal(ahora) {
		t.Error("CanceladoEn no quedó registrado")
	}
	if cancelado.CanceladoPor == nil || *cancelado.CanceladoPor != quienCancela {
		t.Error("CanceladoPor no quedó registrado")
	}
	// Receptor por valor: el original no se toca.
	if turno.Estado != TurnoReservado {
		t.Error("Cancelar mutó el receptor")
	}
}

func TestCancelarDosVecesFalla(t *testing.T) {
	ahora := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	turno, err := NuevoTurno(uuid.New(), uuid.New(), huecoDePrueba(), "", ahora)
	if err != nil {
		t.Fatalf("NuevoTurno: %v", err)
	}

	cancelado, err := turno.Cancelar(uuid.New(), ahora)
	if err != nil {
		t.Fatalf("primer Cancelar: %v", err)
	}

	// No es idempotente a propósito: repetir la operación probablemente
	// significa que el cliente cree algo distinto de lo que pasó.
	if _, err := cancelado.Cancelar(uuid.New(), ahora); err == nil {
		t.Error("cancelar dos veces debería fallar")
	}
}

// El borde: un turno se puede cancelar hasta el instante en que empieza, no
// después. Los dos casos, porque un > mal escrito pasa desapercibido con uno
// solo.
func TestCancelarEnElBorde(t *testing.T) {
	hueco := huecoDePrueba()
	turno, err := NuevoTurno(uuid.New(), uuid.New(), hueco, "", hueco.Inicio.Add(-time.Hour))
	if err != nil {
		t.Fatalf("NuevoTurno: %v", err)
	}

	if _, err := turno.Cancelar(uuid.New(), hueco.Inicio.Add(-time.Nanosecond)); err != nil {
		t.Errorf("un instante antes de empezar todavía se cancela: %v", err)
	}
	if _, err := turno.Cancelar(uuid.New(), hueco.Inicio); err == nil {
		t.Error("en el instante exacto de inicio ya no se cancela")
	}
	if _, err := turno.Cancelar(uuid.New(), hueco.Fin.Add(time.Hour)); err == nil {
		t.Error("un turno pasado no se cancela")
	}
}

func TestTurnoEstaActivo(t *testing.T) {
	ahora := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	turno, err := NuevoTurno(uuid.New(), uuid.New(), huecoDePrueba(), "", ahora)
	if err != nil {
		t.Fatalf("NuevoTurno: %v", err)
	}
	if !turno.EstaActivo() {
		t.Error("un turno recién reservado está activo")
	}

	cancelado, err := turno.Cancelar(uuid.New(), ahora)
	if err != nil {
		t.Fatalf("Cancelar: %v", err)
	}
	if cancelado.EstaActivo() {
		t.Error("un turno cancelado no está activo")
	}
}

// Semiabierto, igual que Bloqueo.SeSolapaCon: un turno que termina 10:50 no
// pisa uno que empieza 10:50.
func TestTurnoSeSolapaCon(t *testing.T) {
	hueco := huecoDePrueba() // 10:00 a 10:50
	turno, err := NuevoTurno(uuid.New(), uuid.New(), hueco, "", time.Now())
	if err != nil {
		t.Fatalf("NuevoTurno: %v", err)
	}

	casos := []struct {
		nombre         string
		desplazamiento time.Duration
		duracion       time.Duration
		esperado       bool
	}{
		{"identico", 0, 50 * time.Minute, true},
		{"empieza adentro", 25 * time.Minute, 50 * time.Minute, true},
		{"termina adentro", -25 * time.Minute, 50 * time.Minute, true},
		{"lo contiene", -time.Hour, 3 * time.Hour, true},
		{"pegado despues", 50 * time.Minute, 50 * time.Minute, false},
		{"pegado antes", -50 * time.Minute, 50 * time.Minute, false},
		{"muy despues", 5 * time.Hour, 50 * time.Minute, false},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			inicio := hueco.Inicio.Add(c.desplazamiento)
			if got := turno.SeSolapaCon(inicio, inicio.Add(c.duracion)); got != c.esperado {
				t.Errorf("SeSolapaCon = %v, se esperaba %v", got, c.esperado)
			}
		})
	}
}

func TestTurnoClonar(t *testing.T) {
	ahora := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	turno, err := NuevoTurno(uuid.New(), uuid.New(), huecoDePrueba(), "", ahora)
	if err != nil {
		t.Fatalf("NuevoTurno: %v", err)
	}
	cancelado, err := turno.Cancelar(uuid.New(), ahora)
	if err != nil {
		t.Fatalf("Cancelar: %v", err)
	}

	clon := cancelado.Clonar()
	*clon.CanceladoEn = clon.CanceladoEn.Add(time.Hour)
	*clon.CanceladoPor = uuid.New()

	if cancelado.CanceladoEn.Equal(*clon.CanceladoEn) {
		t.Error("mutar CanceladoEn del clon alteró el original")
	}
	if *cancelado.CanceladoPor == *clon.CanceladoPor {
		t.Error("mutar CanceladoPor del clon alteró el original")
	}
}

func TestEstadoTurnoEsValido(t *testing.T) {
	for _, e := range []EstadoTurno{TurnoReservado, TurnoCancelado} {
		if !e.EsValido() {
			t.Errorf("%q debería ser válido", e)
		}
	}
	for _, e := range []EstadoTurno{"", "atendido", "RESERVADO"} {
		if e.EsValido() {
			t.Errorf("%q no debería ser válido", e)
		}
	}
}
