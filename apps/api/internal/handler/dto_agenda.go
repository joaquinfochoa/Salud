package handler

import (
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

const formatoFecha = "2006-01-02"

type peticionHorario struct {
	DiaSemana   string `json:"diaSemana"`
	Desde       string `json:"desde"`
	Hasta       string `json:"hasta"`
	DuracionMin int    `json:"duracionMin"`
	Modalidad   string `json:"modalidad"`
}

type peticionHorarios struct {
	Horarios []peticionHorario `json:"horarios"`
}

func (p peticionHorarios) aEntradas() []domain.EntradaHorarioSemanal {
	entradas := make([]domain.EntradaHorarioSemanal, 0, len(p.Horarios))
	for _, h := range p.Horarios {
		entradas = append(entradas, domain.EntradaHorarioSemanal{
			DiaSemana:   h.DiaSemana,
			Desde:       h.Desde,
			Hasta:       h.Hasta,
			DuracionMin: h.DuracionMin,
			Modalidad:   h.Modalidad,
		})
	}
	return entradas
}

type respuestaHorario struct {
	DiaSemana   string `json:"diaSemana"`
	Desde       string `json:"desde"`
	Hasta       string `json:"hasta"`
	DuracionMin int    `json:"duracionMin"`
	Modalidad   string `json:"modalidad"`
}

type respuestaHorarios struct {
	Horarios []respuestaHorario `json:"horarios"`

	// TurnosCancelados solo lo llena el PUT, que es el que puede dejar turnos
	// huérfanos. El GET lo manda en cero: la clave está siempre presente para
	// que el cliente no tenga que distinguir "no vino" de "ninguno".
	TurnosCancelados int `json:"turnosCancelados"`
}

func aRespuestaHorarios(semana []domain.HorarioSemanal) respuestaHorarios {
	horarios := make([]respuestaHorario, 0, len(semana))
	for _, h := range semana {
		horarios = append(horarios, respuestaHorario{
			DiaSemana:   string(h.DiaSemana),
			Desde:       h.Desde.String(),
			Hasta:       h.Hasta.String(),
			DuracionMin: h.DuracionMin,
			Modalidad:   string(h.Modalidad),
		})
	}
	return respuestaHorarios{Horarios: horarios}
}

type peticionBloqueo struct {
	Desde  time.Time `json:"desde"`
	Hasta  time.Time `json:"hasta"`
	Motivo string    `json:"motivo"`
}

func (p peticionBloqueo) aEntrada() domain.EntradaBloqueo {
	return domain.EntradaBloqueo{
		Desde:  p.Desde.In(domain.ZonaHoraria),
		Hasta:  p.Hasta.In(domain.ZonaHoraria),
		Motivo: p.Motivo,
	}
}

// respuestaBloqueoCreado agrega al bloqueo cuántos turnos canceló al crearse.
// Es su propio tipo y no un campo de respuestaBloqueo porque el listado
// devuelve muchos bloqueos y el contador es de la operación, no del recurso.
type respuestaBloqueoCreado struct {
	respuestaBloqueo
	TurnosCancelados int `json:"turnosCancelados"`
}

type respuestaBloqueo struct {
	ID       string    `json:"id"`
	Desde    time.Time `json:"desde"`
	Hasta    time.Time `json:"hasta"`
	Motivo   string    `json:"motivo"`
	CreadoEn time.Time `json:"creadoEn"`
}

func aRespuestaBloqueo(b domain.Bloqueo) respuestaBloqueo {
	return respuestaBloqueo{
		ID:       b.ID.String(),
		Desde:    b.Desde.In(domain.ZonaHoraria),
		Hasta:    b.Hasta.In(domain.ZonaHoraria),
		Motivo:   b.Motivo,
		CreadoEn: b.CreadoEn.In(domain.ZonaHoraria),
	}
}

type respuestaBloqueos struct {
	Datos []respuestaBloqueo `json:"datos"`
}

func aRespuestaBloqueos(bloqueos []domain.Bloqueo) respuestaBloqueos {
	datos := make([]respuestaBloqueo, 0, len(bloqueos))
	for _, b := range bloqueos {
		datos = append(datos, aRespuestaBloqueo(b))
	}
	return respuestaBloqueos{Datos: datos}
}

type respuestaHueco struct {
	Inicio    time.Time `json:"inicio"`
	Fin       time.Time `json:"fin"`
	Modalidad string    `json:"modalidad"`
}

type rangoRespuesta struct {
	Desde string `json:"desde"`
	Hasta string `json:"hasta"`
}

type respuestaHuecos struct {
	Datos []respuestaHueco `json:"datos"`
	Rango rangoRespuesta   `json:"rango"`
}

func aRespuestaHuecos(resultado service.ResultadoHuecos) respuestaHuecos {
	datos := make([]respuestaHueco, 0, len(resultado.Huecos))
	for _, h := range resultado.Huecos {
		datos = append(datos, respuestaHueco{
			Inicio:    h.Inicio,
			Fin:       h.Fin,
			Modalidad: string(h.Modalidad),
		})
	}
	return respuestaHuecos{
		Datos: datos,
		Rango: rangoRespuesta{
			Desde: resultado.Desde.Format(formatoFecha),
			Hasta: resultado.Hasta.Format(formatoFecha),
		},
	}
}
