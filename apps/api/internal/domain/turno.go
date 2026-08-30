package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// maxLargoMotivoTurno es más corto que el de Bio a propósito: un motivo de
// consulta es una línea, no una historia clínica. Y no debe serlo: cuanto menos
// dato de salud viaje y se guarde, menos superficie tiene la Ley 25.326.
const maxLargoMotivoTurno = 500

// EstadoTurno tiene dos valores y no cuatro.
//
// "atendido" y "ausente" se difirieron: hoy nada consume esa información —no
// hay cobro, ni reputación, ni política de ausencias— y agregarlas después es
// una constante acá, un método de transición y un endpoint, sin migración.
//
// Un turno pasado no necesita estado: eso lo dice el reloj.
type EstadoTurno string

const (
	TurnoReservado EstadoTurno = "reservado"
	TurnoCancelado EstadoTurno = "cancelado"
)

func (e EstadoTurno) EsValido() bool {
	switch e {
	case TurnoReservado, TurnoCancelado:
		return true
	}
	return false
}

// Turno es un acuerdo entre un paciente y un profesional para un intervalo
// concreto.
//
// No hay entidad Paciente: PacienteID referencia a un Usuario. Una entidad que
// solo tuviera un UsuarioID y nada más es ceremonia; aparece el día que el
// paciente necesite datos que un usuario no tiene.
type Turno struct {
	ID            uuid.UUID
	ProfesionalID uuid.UUID
	PacienteID    uuid.UUID

	// Inicio, Fin y Modalidad son una copia del hueco al momento de reservar,
	// no una lectura del horario actual. Si el profesional convierte los
	// martes de presencial a telemedicina, este turno sigue diciendo lo que se
	// pactó: leerlo del horario lo reescribiría hacia atrás.
	Inicio    time.Time
	Fin       time.Time
	Modalidad Modalidad

	Estado   EstadoTurno
	Motivo   string
	CreadoEn time.Time

	// CanceladoEn y CanceladoPor existen desde el día uno aunque nada los lea
	// todavía. El día que haya una política de ausencias —"cancelar con menos
	// de 24 horas cuenta como ausencia"— este dato no se puede reconstruir si
	// no se guardó. Mismo argumento que DadoDeBajaEn en Profesional.
	//
	// CanceladoPor es un UsuarioID y no un enum paciente|profesional: guarda
	// más por el mismo precio, y quién es cada uno se deduce comparando contra
	// PacienteID.
	CanceladoEn  *time.Time
	CanceladoPor *uuid.UUID
}

// NuevoTurno recibe el hueco entero y no un inicio suelto. Es deliberado: Fin y
// Modalidad salen de ahí, así que no hay forma de construir un turno de tres
// horas en una agenda de cincuenta minutos ni de inventarse una modalidad que
// el profesional no ofrece.
func NuevoTurno(profesionalID, pacienteID uuid.UUID, hueco Hueco, motivo string, ahora time.Time) (Turno, error) {
	var verr ErrorValidacion

	limpio := strings.TrimSpace(motivo)
	if utf8.RuneCountInString(limpio) > maxLargoMotivoTurno {
		verr.agregar("motivo", fmt.Sprintf("no puede superar los %d caracteres", maxLargoMotivoTurno))
	}

	if verr.tieneErrores() {
		return Turno{}, verr
	}

	return Turno{
		ID:            uuid.New(),
		ProfesionalID: profesionalID,
		PacienteID:    pacienteID,
		Inicio:        hueco.Inicio,
		Fin:           hueco.Fin,
		Modalidad:     hueco.Modalidad,
		Estado:        TurnoReservado,
		Motivo:        limpio,
		CreadoEn:      ahora,
	}, nil
}

// Cancelar devuelve el turno cancelado sin tocar el receptor.
//
// No es idempotente, a diferencia del logout: cancelar dos veces probablemente
// significa que el cliente cree algo distinto de lo que pasó, y decírselo es
// más útil que devolverle un éxito falso.
//
// Un turno que ya empezó no se cancela. Cancelarlo hacia atrás borraría el
// registro de una consulta que efectivamente ocurrió.
func (t Turno) Cancelar(porUsuarioID uuid.UUID, ahora time.Time) (Turno, error) {
	switch {
	case t.Estado == TurnoCancelado:
		return Turno{}, ErrorValidacion{Campos: []ErrorCampo{
			{Campo: "estado", Mensaje: "el turno ya está cancelado"},
		}}
	case !t.Inicio.After(ahora):
		return Turno{}, ErrorValidacion{Campos: []ErrorCampo{
			{Campo: "inicio", Mensaje: "el turno ya empezó y no se puede cancelar"},
		}}
	}

	t.Estado = TurnoCancelado
	t.CanceladoEn = &ahora
	t.CanceladoPor = &porUsuarioID
	return t, nil
}

// EstaActivo dice si el turno sigue ocupando su hueco. Un turno cancelado lo
// libera, que es el punto de cancelar.
func (t Turno) EstaActivo() bool {
	return t.Estado == TurnoReservado
}

// SeSolapaCon dice si el turno pisa el intervalo [inicio, fin).
//
// Semiabierto, igual que Bloqueo.SeSolapaCon: un turno que termina 10:50 no
// pisa uno que empieza 10:50. Usar <= en cualquiera de las dos comparaciones
// haría desaparecer un turno por día.
func (t Turno) SeSolapaCon(inicio, fin time.Time) bool {
	return inicio.Before(t.Fin) && fin.After(t.Inicio)
}

// Clonar devuelve una copia profunda. Los dos punteros importan: sin esto,
// quien reciba la copia puede correrle la fecha de cancelación al original.
func (t Turno) Clonar() Turno {
	c := t
	if t.CanceladoEn != nil {
		v := *t.CanceladoEn
		c.CanceladoEn = &v
	}
	if t.CanceladoPor != nil {
		v := *t.CanceladoPor
		c.CanceladoPor = &v
	}
	return c
}
