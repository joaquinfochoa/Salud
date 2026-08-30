package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const tipoContenidoProblema = "application/problem+json"

// Los tipos de error del contrato. Son URIs por convención de RFC 7807: no
// hace falta que resuelvan a una página, alcanza con que identifiquen la clase
// de problema de forma estable.
const (
	tipoPeticionInvalida      = "https://salud.app/errors/bad-request"
	tipoValidacion            = "https://salud.app/errors/validation"
	tipoNoEncontrado          = "https://salud.app/errors/not-found"
	tipoConflicto             = "https://salud.app/errors/conflict"
	tipoInterno               = "https://salud.app/errors/internal"
	tipoCuerpoDemasiadoGrande = "https://salud.app/errors/payload-too-large"
	tipoMetodoNoPermitido     = "https://salud.app/errors/method-not-allowed"
	tipoNoAutenticado         = "https://salud.app/errors/unauthorized"
	tipoNoAutorizado          = "https://salud.app/errors/forbidden"
	tipoTipoDeContenido       = "https://salud.app/errors/unsupported-media-type"
)

// Problema es la representación de un error según RFC 7807.
type Problema struct {
	Tipo    string              `json:"type"`
	Titulo  string              `json:"title"`
	Estado  int                 `json:"status"`
	Detalle string              `json:"detail,omitempty"`
	Errores []domain.ErrorCampo `json:"errores,omitempty"`
}

func escribirProblema(w http.ResponseWriter, p Problema) {
	w.Header().Set("Content-Type", tipoContenidoProblema)
	w.WriteHeader(p.Estado)
	_ = json.NewEncoder(w).Encode(p)
}

func escribirJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// escribirError traduce un error del dominio al problema HTTP que le corresponde.
//
// Es el único lugar del proyecto donde el dominio se convierte en códigos de
// estado. Las capas de abajo no saben que existe HTTP.
func escribirError(w http.ResponseWriter, r *http.Request, err error) {
	var verr domain.ErrorValidacion

	switch {
	case errors.As(err, &verr):
		escribirProblema(w, Problema{
			Tipo:    tipoValidacion,
			Titulo:  "Datos inválidos",
			Estado:  http.StatusUnprocessableEntity,
			Detalle: "Uno o más campos no cumplen las reglas del sistema",
			Errores: verr.Campos,
		})

	case errors.Is(err, domain.ErrNoEncontrado):
		escribirProblema(w, Problema{
			Tipo:    tipoNoEncontrado,
			Titulo:  "No encontrado",
			Estado:  http.StatusNotFound,
			Detalle: "El profesional solicitado no existe",
		})

	case errors.Is(err, domain.ErrMatriculaEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Matrícula ya registrada",
			Estado:  http.StatusConflict,
			Detalle: "Otro profesional ya tiene registrada esa matrícula",
		})

	// El repositorio rechaza estos dos al escribir, así que pueden llegar hasta
	// acá. Son choques con otro registro, no fallas del servidor: si cayeran en
	// el default, el cliente vería un 500 por un conflicto que puede resolver
	// reintentando.
	case errors.Is(err, domain.ErrSlugEnUso), errors.Is(err, domain.ErrIDEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Conflicto con otro registro",
			Estado:  http.StatusConflict,
			Detalle: "El alta chocó con un profesional existente. Volvé a intentar.",
		})

	case errors.Is(err, domain.ErrCredencialesInvalidas):
		escribirProblema(w, Problema{
			Tipo:    tipoNoAutenticado,
			Titulo:  "Credenciales inválidas",
			Estado:  http.StatusUnauthorized,
			Detalle: "El email o la contraseña no son correctos",
		})

	case errors.Is(err, domain.ErrNoAutorizado):
		escribirProblema(w, Problema{
			Tipo:    tipoNoAutorizado,
			Titulo:  "No autorizado",
			Estado:  http.StatusForbidden,
			Detalle: "Solo el dueño de este perfil puede modificarlo",
		})

	case errors.Is(err, domain.ErrEmailEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Email ya registrado",
			Estado:  http.StatusConflict,
			Detalle: "Ya existe una cuenta con ese email",
		})

	case errors.Is(err, domain.ErrYaTienePerfil), errors.Is(err, domain.ErrUsuarioEnUso):
		escribirProblema(w, Problema{
			Tipo:    tipoConflicto,
			Titulo:  "Ya tenés un perfil profesional",
			Estado:  http.StatusConflict,
			Detalle: "Cada cuenta puede tener un solo perfil profesional",
		})

	default:
		// El error real va al log, nunca al cliente: puede contener nombres
		// de tablas, rutas del servidor o datos de otro usuario.
		slog.ErrorContext(r.Context(), "error no manejado",
			"error", err,
			"metodo", r.Method,
			"ruta", r.URL.Path,
		)
		escribirProblema(w, Problema{
			Tipo:    tipoInterno,
			Titulo:  "Error interno",
			Estado:  http.StatusInternalServerError,
			Detalle: "Ocurrió un error inesperado. Volvé a intentar.",
		})
	}
}

// escribirNoAutenticado es el 401 de "no sé quién sos". No confundir con el
// 403 de ErrNoAutorizado, que es "sé quién sos y no te alcanza".
func escribirNoAutenticado(w http.ResponseWriter) {
	escribirProblema(w, Problema{
		Tipo:    tipoNoAutenticado,
		Titulo:  "No autenticado",
		Estado:  http.StatusUnauthorized,
		Detalle: "Necesitás iniciar sesión para hacer esto",
	})
}

func escribirPeticionInvalida(w http.ResponseWriter, detail string) {
	escribirProblema(w, Problema{
		Tipo:    tipoPeticionInvalida,
		Titulo:  "Petición inválida",
		Estado:  http.StatusBadRequest,
		Detalle: detail,
	})
}
