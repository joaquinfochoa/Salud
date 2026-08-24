package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// ManejadorAgenda traduce entre HTTP y el servicio de agenda. Como el resto de
// los handlers, no decide nada: decodifica, delega y serializa.
type ManejadorAgenda struct {
	svc *service.Agenda
}

func NuevaAgenda(svc *service.Agenda) *ManejadorAgenda {
	return &ManejadorAgenda{svc: svc}
}

func (h *ManejadorAgenda) ListarHorarios(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	semana, err := h.svc.ListarHorarios(r.Context(), id)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaHorarios(semana))
}

func (h *ManejadorAgenda) ReemplazarHorarios(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	var peticion peticionHorarios
	if err := decodificarJSON(w, r, &peticion); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	semana, err := h.svc.ReemplazarHorarios(r.Context(), id, peticion.aEntradas())
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaHorarios(semana))
}

func (h *ManejadorAgenda) ListarBloqueos(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	consulta := r.URL.Query()

	// Sin rango se devuelven los vigentes y futuros: el servicio resuelve esa
	// ventana con su propio reloj, así que acá solo se traduce "no vino" a
	// nil, sin decidir nada de negocio.
	var desde, hasta *time.Time

	if crudo := consulta.Get("desde"); crudo != "" {
		fecha, err := parsearFecha(crudo)
		if err != nil {
			escribirPeticionInvalida(w, "el parámetro desde tiene que ser una fecha AAAA-MM-DD")
			return
		}
		desde = &fecha
	}
	if crudo := consulta.Get("hasta"); crudo != "" {
		fecha, err := parsearFecha(crudo)
		if err != nil {
			escribirPeticionInvalida(w, "el parámetro hasta tiene que ser una fecha AAAA-MM-DD")
			return
		}
		fecha = fecha.AddDate(0, 0, 1) // el día pedido entra entero
		hasta = &fecha
	}

	bloqueos, err := h.svc.ListarBloqueos(r.Context(), id, desde, hasta)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaBloqueos(bloqueos))
}

func (h *ManejadorAgenda) CrearBloqueo(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	var peticion peticionBloqueo
	if err := decodificarJSON(w, r, &peticion); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	bloqueo, err := h.svc.CrearBloqueo(r.Context(), id, peticion.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/profesionales/"+id.String()+"/bloqueos/"+bloqueo.ID.String())
	escribirJSON(w, http.StatusCreated, aRespuestaBloqueo(bloqueo))
}

func (h *ManejadorAgenda) EliminarBloqueo(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	bloqueoID, err := uuid.Parse(r.PathValue("bloqueoId"))
	if err != nil {
		escribirPeticionInvalida(w, "el id del bloqueo tiene que ser un UUID válido")
		return
	}

	if err := h.svc.EliminarBloqueo(r.Context(), id, bloqueoID); err != nil {
		escribirError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorAgenda) HuecosLibres(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	consulta := r.URL.Query()

	desde, err := parsearFecha(consulta.Get("desde"))
	if err != nil {
		escribirPeticionInvalida(w, "el parámetro desde es obligatorio y tiene que ser una fecha AAAA-MM-DD")
		return
	}

	hasta, err := parsearFecha(consulta.Get("hasta"))
	if err != nil {
		escribirPeticionInvalida(w, "el parámetro hasta es obligatorio y tiene que ser una fecha AAAA-MM-DD")
		return
	}

	// Un rango invertido es un query param mal formado, no una entidad
	// semánticamente inválida: el contrato pide 400, no 422.
	if hasta.Before(desde) {
		escribirPeticionInvalida(w, "el parámetro hasta tiene que ser posterior o igual a desde")
		return
	}

	resultado, err := h.svc.HuecosLibres(r.Context(), id, desde, hasta)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaHuecos(resultado))
}

// parsearFecha lee una fecha AAAA-MM-DD como el arranque de ese día en la zona
// del sistema.
func parsearFecha(crudo string) (time.Time, error) {
	return time.ParseInLocation(formatoFecha, crudo, domain.ZonaHoraria)
}
