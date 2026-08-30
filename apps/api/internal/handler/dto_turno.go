package handler

import (
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// peticionTurno es todo lo que el cliente manda para reservar.
//
// Deliberadamente no tiene fin ni modalidad: los dos salen del hueco que se
// está tomando. Aceptarlos sería dejar que el cliente se invente un turno de
// tres horas en una agenda de cincuenta minutos, o una modalidad que el
// profesional no ofrece. Tampoco tiene pacienteId: ese sale de la sesión.
type peticionTurno struct {
	Inicio time.Time `json:"inicio"`
	Motivo string    `json:"motivo"`
}

type respuestaTurno struct {
	ID            string     `json:"id"`
	ProfesionalID string     `json:"profesionalId"`
	PacienteID    string     `json:"pacienteId"`
	Inicio        time.Time  `json:"inicio"`
	Fin           time.Time  `json:"fin"`
	Modalidad     string     `json:"modalidad"`
	Estado        string     `json:"estado"`
	Motivo        string     `json:"motivo"`
	CreadoEn      time.Time  `json:"creadoEn"`
	CanceladoEn   *time.Time `json:"canceladoEn"`
	CanceladoPor  *string    `json:"canceladoPor"`
}

func aRespuestaTurno(t domain.Turno) respuestaTurno {
	r := respuestaTurno{
		ID:            t.ID.String(),
		ProfesionalID: t.ProfesionalID.String(),
		PacienteID:    t.PacienteID.String(),
		Inicio:        t.Inicio.In(domain.ZonaHoraria),
		Fin:           t.Fin.In(domain.ZonaHoraria),
		Modalidad:     string(t.Modalidad),
		Estado:        string(t.Estado),
		Motivo:        t.Motivo,
		CreadoEn:      t.CreadoEn.In(domain.ZonaHoraria),
	}

	// Las dos claves salen siempre, en null cuando no hay cancelación: están
	// en `required` del contrato, así que un cliente estricto las espera.
	if t.CanceladoEn != nil {
		en := t.CanceladoEn.In(domain.ZonaHoraria)
		r.CanceladoEn = &en
	}
	if t.CanceladoPor != nil {
		por := t.CanceladoPor.String()
		r.CanceladoPor = &por
	}
	return r
}

// respuestaListaTurnos usa la clave "datos", igual que el listado de
// profesionales. Sin paginación: no es una colección paginada sino una
// ventana temporal, como ListaBloqueos.
type respuestaListaTurnos struct {
	Datos []respuestaTurno `json:"datos"`
}

func aRespuestaListaTurnos(turnos []domain.Turno) respuestaListaTurnos {
	datos := make([]respuestaTurno, 0, len(turnos))
	for _, t := range turnos {
		datos = append(datos, aRespuestaTurno(t))
	}
	return respuestaListaTurnos{Datos: datos}
}
