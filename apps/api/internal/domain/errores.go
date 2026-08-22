package domain

import (
	"errors"
	"strings"
)

var (
	// ErrNoEncontrado lo devuelve el repositorio cuando no existe el registro.
	ErrNoEncontrado = errors.New("profesional no encontrado")

	// ErrMatriculaEnUso lo devuelve el servicio: la matrícula es la única
	// identidad real de una persona en este sistema y no puede repetirse.
	ErrMatriculaEnUso = errors.New("matricula ya registrada")
)

// ErrorCampo señala un campo puntual. Las etiquetas JSON coinciden con el
// formato problem+json que espera el cliente.
type ErrorCampo struct {
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}

// ErrorValidacion junta todos los campos inválidos de una sola pasada.
// Devolver solo el primero obliga al cliente a corregir de a uno, que es una
// experiencia horrible en un formulario de alta con nueve campos.
type ErrorValidacion struct {
	Campos []ErrorCampo
}

func (e ErrorValidacion) Error() string {
	partes := make([]string, 0, len(e.Campos))
	for _, f := range e.Campos {
		partes = append(partes, f.Campo+": "+f.Mensaje)
	}
	return "validación fallida — " + strings.Join(partes, "; ")
}

func (e *ErrorValidacion) agregar(campo, mensaje string) {
	e.Campos = append(e.Campos, ErrorCampo{Campo: campo, Mensaje: mensaje})
}

func (e ErrorValidacion) tieneErrores() bool {
	return len(e.Campos) > 0
}
