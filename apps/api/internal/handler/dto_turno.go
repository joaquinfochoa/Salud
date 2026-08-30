package handler

import (
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
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

// Las dos partes de un turno. Cada listado devuelve la OTRA: el paciente ve con
// quién es su turno, el profesional ve quién viene. Son dos formas y no una con
// los dos campos opcionales porque así el consumidor no tiene que acordarse de
// qué endpoint lo sacó para saber cuál viene lleno.
type parteProfesional struct {
	ID           string `json:"id"`
	Nombre       string `json:"nombre"`
	Apellido     string `json:"apellido"`
	Slug         string `json:"slug"`
	Especialidad string `json:"especialidad"`
}

// Sin email: el profesional necesita saber a quién atiende, no cómo contactarlo
// por fuera de la plataforma.
type partePaciente struct {
	ID       string `json:"id"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
}

type respuestaTurnoConProfesional struct {
	respuestaTurno
	Profesional parteProfesional `json:"profesional"`
}

type respuestaTurnoConPaciente struct {
	respuestaTurno
	Paciente partePaciente `json:"paciente"`
}

// Los dos listados usan la clave "datos", igual que el de profesionales. Sin
// paginación: no son colecciones paginadas sino ventanas temporales, como
// ListaBloqueos.
type respuestaListaTurnosConProfesional struct {
	Datos []respuestaTurnoConProfesional `json:"datos"`
}

type respuestaListaTurnosConPaciente struct {
	Datos []respuestaTurnoConPaciente `json:"datos"`
}

func aRespuestaListaDePaciente(turnos []service.TurnoConProfesional) respuestaListaTurnosConProfesional {
	datos := make([]respuestaTurnoConProfesional, 0, len(turnos))
	for _, t := range turnos {
		datos = append(datos, respuestaTurnoConProfesional{
			respuestaTurno: aRespuestaTurno(t.Turno),
			Profesional: parteProfesional{
				ID:           t.Profesional.ID.String(),
				Nombre:       t.Profesional.Nombre,
				Apellido:     t.Profesional.Apellido,
				Slug:         t.Profesional.Slug,
				Especialidad: string(t.Profesional.Especialidad),
			},
		})
	}
	return respuestaListaTurnosConProfesional{Datos: datos}
}

func aRespuestaListaDeProfesional(turnos []service.TurnoConPaciente) respuestaListaTurnosConPaciente {
	datos := make([]respuestaTurnoConPaciente, 0, len(turnos))
	for _, t := range turnos {
		datos = append(datos, respuestaTurnoConPaciente{
			respuestaTurno: aRespuestaTurno(t.Turno),
			Paciente: partePaciente{
				ID:       t.Paciente.ID.String(),
				Nombre:   t.Paciente.Nombre,
				Apellido: t.Paciente.Apellido,
			},
		})
	}
	return respuestaListaTurnosConPaciente{Datos: datos}
}
