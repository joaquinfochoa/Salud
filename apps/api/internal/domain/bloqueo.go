package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxLargoMotivo = 200

// Bloqueo es lo que rompe el horario habitual: vacaciones, un feriado, cerrar
// temprano un martes.
//
// Guarda fechas y horas locales, no instantes UTC, por la misma razón que
// HorarioSemanal guarda hora de reloj: "me voy del 10 al 20" es una fecha del
// calendario del profesional.
//
// A diferencia de HorarioSemanal sí tiene ID, porque se borra de a uno.
type Bloqueo struct {
	ID            uuid.UUID
	ProfesionalID uuid.UUID
	Desde         time.Time
	Hasta         time.Time
	Motivo        string
	CreadoEn      time.Time
}

type EntradaBloqueo struct {
	Desde  time.Time
	Hasta  time.Time
	Motivo string
}

// NuevoBloqueo valida la entrada y devuelve un bloqueo consistente.
//
// El ahora llega por parámetro y no se lee del reloj del sistema para que la
// regla de "no bloquear el pasado" sea determinista en los tests.
func NuevoBloqueo(profesionalID uuid.UUID, entrada EntradaBloqueo, ahora time.Time) (Bloqueo, error) {
	var verr ErrorValidacion

	if entrada.Desde.IsZero() {
		verr.agregar("desde", "es obligatorio")
	}
	if entrada.Hasta.IsZero() {
		verr.agregar("hasta", "es obligatorio")
	}

	if !entrada.Desde.IsZero() && !entrada.Hasta.IsZero() {
		if !entrada.Desde.Before(entrada.Hasta) {
			verr.agregar("hasta", "tiene que ser posterior a desde")
		} else if !entrada.Hasta.After(ahora) {
			// Un bloqueo ya terminado no cambia nada: los huecos que tapaba
			// pasaron y de todos modos no se pueden reservar. Uno que empezó y
			// todavía no terminó sí es válido.
			verr.agregar("hasta", "tiene que estar en el futuro")
		}
	}

	motivo := strings.TrimSpace(entrada.Motivo)
	if utf8.RuneCountInString(motivo) > maxLargoMotivo {
		verr.agregar("motivo", fmt.Sprintf("no puede superar los %d caracteres", maxLargoMotivo))
	}

	if verr.tieneErrores() {
		return Bloqueo{}, verr
	}

	return Bloqueo{
		ID:            uuid.New(),
		ProfesionalID: profesionalID,
		Desde:         entrada.Desde,
		Hasta:         entrada.Hasta,
		Motivo:        motivo,
		CreadoEn:      ahora,
	}, nil
}

// SeSolapaCon dice si el bloqueo pisa el intervalo [inicio, fin).
//
// Los intervalos son semiabiertos en todo el sistema: un bloqueo que empieza
// 09:50 no pisa un hueco que termina 09:50. Es el off-by-one clásico de las
// agendas, y usar <= en cualquiera de las dos comparaciones haría desaparecer
// un turno por día sin que nadie entienda por qué.
func (b Bloqueo) SeSolapaCon(inicio, fin time.Time) bool {
	return inicio.Before(b.Hasta) && fin.After(b.Desde)
}
