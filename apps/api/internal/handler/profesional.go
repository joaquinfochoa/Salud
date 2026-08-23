package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// ManejadorProfesional traduce entre HTTP y el servicio. No toma decisiones de
// negocio: decodifica, delega y serializa.
type ManejadorProfesional struct {
	svc *service.Profesional
}

func NuevoProfesional(svc *service.Profesional) *ManejadorProfesional {
	return &ManejadorProfesional{svc: svc}
}

func (h *ManejadorProfesional) Crear(w http.ResponseWriter, r *http.Request) {
	var req peticionProfesional
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	p, err := h.svc.Crear(r.Context(), req.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/profesionales/"+p.ID.String())
	escribirJSON(w, http.StatusCreated, aRespuesta(p))
}

func (h *ManejadorProfesional) ObtenerPorID(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

func (h *ManejadorProfesional) ObtenerPorSlug(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.ObtenerPorSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

func (h *ManejadorProfesional) Listar(w http.ResponseWriter, r *http.Request) {
	f, err := parsearFiltro(r)
	if err != nil {
		escribirPeticionInvalida(w, err.Error())
		return
	}

	f = service.NormalizarFiltro(f)

	ps, total, err := h.svc.Listar(r.Context(), f)
	if err != nil {
		escribirError(w, r, err)
		return
	}

	escribirJSON(w, http.StatusOK, aRespuestaListado(ps, total, f.Limite, f.Desplazamiento))
}

func (h *ManejadorProfesional) Actualizar(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	var req peticionProfesional
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	p, err := h.svc.Actualizar(r.Context(), id, req.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

// DarDeBaja implementa el DELETE. Se llama DarDeBaja y no Delete porque eso
// es lo que hace: baja lógica. El verbo HTTP conserva el nombre que espera
// cualquiera que lea un CRUD.
func (h *ManejadorProfesional) DarDeBaja(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	if err := h.svc.DarDeBaja(r.Context(), id); err != nil {
		escribirError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorProfesional) Reactivar(w http.ResponseWriter, r *http.Request) {
	id, ok := parsearID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.Reactivar(r.Context(), id)
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuesta(p))
}

// parsearID devuelve false y ya escribió la respuesta si el ID no es un UUID.
// Un ID mal formado es un error del cliente (400), no un recurso que falta.
func parsearID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	crudo := r.PathValue("id")
	id, err := uuid.Parse(crudo)
	if err != nil {
		escribirPeticionInvalida(w, "el id debe ser un UUID válido")
		return uuid.Nil, false
	}
	return id, true
}

func parsearFiltro(r *http.Request) (repository.Filtro, error) {
	q := r.URL.Query()
	var f repository.Filtro

	if crudo := q.Get("especialidad"); crudo != "" {
		esp := domain.Especialidad(crudo)
		if !esp.EsValida() {
			return f, errParametroInvalido("especialidad", "debe ser psicologia, kinesiologia u odontologia")
		}
		f.Especialidad = &esp
	}

	if crudo := q.Get("estado"); crudo != "" {
		st := domain.Estado(crudo)
		if !st.EsValido() {
			return f, errParametroInvalido("estado", "debe ser activo o inactivo")
		}
		f.Estado = &st
	}

	if crudo := q.Get("zona"); crudo != "" {
		f.Zona = &crudo
	}
	if crudo := q.Get("busqueda"); crudo != "" {
		f.Busqueda = &crudo
	}

	if crudo := q.Get("limite"); crudo != "" {
		v, err := strconv.Atoi(crudo)
		if err != nil || v < 1 {
			return f, errParametroInvalido("limite", "debe ser un entero mayor a cero")
		}
		f.Limite = v
	}

	if crudo := q.Get("desplazamiento"); crudo != "" {
		v, err := strconv.Atoi(crudo)
		if err != nil || v < 0 {
			return f, errParametroInvalido("desplazamiento", "debe ser un entero mayor o igual a cero")
		}
		f.Desplazamiento = v
	}

	return f, nil
}

type errorParametro struct {
	parametro string
	mensaje   string
}

func (e errorParametro) Error() string {
	return "parámetro " + e.parametro + ": " + e.mensaje
}

func errParametroInvalido(parametro, mensaje string) error {
	return errorParametro{parametro: parametro, mensaje: mensaje}
}
