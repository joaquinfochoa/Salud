package domain

import (
	"errors"
	"strings"
)

var (
	// ErrNoEncontrado lo devuelve el repositorio cuando no existe el registro.
	ErrNoEncontrado = errors.New("profesional no encontrado")

	// ErrMatriculaEnUso lo devuelve el repositorio al escribir: la matrícula es
	// la única identidad real de una persona en este sistema y no puede
	// repetirse. El servicio también lo devuelve desde su chequeo previo, pero
	// la garantía está en la escritura.
	ErrMatriculaEnUso = errors.New("matricula ya registrada")

	// ErrSlugEnUso lo devuelve el repositorio cuando otro registro ya ocupa ese
	// slug. Casi nunca llega al cliente: dos homónimos son perfectamente
	// posibles, así que el servicio reintenta con el sufijo siguiente en vez de
	// rechazar el alta.
	ErrSlugEnUso = errors.New("slug ya registrado")

	// ErrIDEnUso lo devuelve el repositorio ante un alta con un ID que ya
	// existe. Es un conflicto como los otros dos y no un error interno: si
	// cayera en el default del handler, el cliente vería un 500.
	ErrIDEnUso = errors.New("id ya registrado")
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
