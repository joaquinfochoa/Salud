package handler

import (
	"net/http"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// ManejadorTurno traduce entre HTTP y el servicio de turnos. Los cuatro
// endpoints son privados: no hay ninguna vista pública de un turno.
type ManejadorTurno struct {
	svc *service.Turno
}

func NuevoTurno(svc *service.Turno) *ManejadorTurno {
	return &ManejadorTurno{svc: svc}
}

func (h *ManejadorTurno) Reservar(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}
	profesionalID, ok := parsearID(w, r)
	if !ok {
		return
	}

	var req peticionTurno
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	// El paciente sale de la sesión, nunca del cuerpo: aceptarlo sería dejar
	// que cualquiera reserve a nombre de otro.
	turno, err := h.svc.Reservar(r.Context(), usuarioID, profesionalID, req.Inicio.In(domain.ZonaHoraria), req.Motivo)
	if err != nil {
		escribirError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/turnos/"+turno.ID.String())
	escribirJSON(w, http.StatusCreated, aRespuestaTurno(turno))
}

func (h *ManejadorTurno) Cancelar(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}
	turnoID, ok := parsearID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Cancelar(r.Context(), usuarioID, turnoID); err != nil {
		escribirError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MisTurnos no acepta ningún filtro por paciente: el paciente es el de la
// sesión. Aceptarlo convertiría este endpoint en una forma de leer la agenda
// de cualquiera.
func (h *ManejadorTurno) MisTurnos(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	desde, hasta, ok := parsearVentana(w, r)
	if !ok {
		return
	}

	turnos, err := h.svc.ListarDePaciente(r.Context(), usuarioID, desde, hasta)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaListaTurnos(turnos))
}

func (h *ManejadorTurno) DeProfesional(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}
	profesionalID, ok := parsearID(w, r)
	if !ok {
		return
	}

	desde, hasta, ok := parsearVentana(w, r)
	if !ok {
		return
	}

	turnos, err := h.svc.ListarDeProfesional(r.Context(), usuarioID, profesionalID, desde, hasta)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaListaTurnos(turnos))
}

// parsearVentana traduce los desde/hasta opcionales de la query, con el mismo
// formato AAAA-MM-DD que usa ListarBloqueos. Devuelve false si ya escribió el
// 400.
//
// "No vino" se traduce a nil y nada más: qué significa un rango ausente lo
// decide el servicio con su propio reloj, porque es negocio y no transporte.
func parsearVentana(w http.ResponseWriter, r *http.Request) (*time.Time, *time.Time, bool) {
	consulta := r.URL.Query()
	var desde, hasta *time.Time

	if crudo := consulta.Get("desde"); crudo != "" {
		fecha, err := parsearFecha(crudo)
		if err != nil {
			escribirPeticionInvalida(w, "el parámetro desde tiene que ser una fecha AAAA-MM-DD")
			return nil, nil, false
		}
		desde = &fecha
	}
	if crudo := consulta.Get("hasta"); crudo != "" {
		fecha, err := parsearFecha(crudo)
		if err != nil {
			escribirPeticionInvalida(w, "el parámetro hasta tiene que ser una fecha AAAA-MM-DD")
			return nil, nil, false
		}
		fecha = fecha.AddDate(0, 0, 1) // el día pedido entra entero
		hasta = &fecha
	}
	return desde, hasta, true
}
