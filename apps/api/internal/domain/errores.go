package domain

import (
	"errors"
	"strings"
)

var (
	// ErrNoEncontrado lo devuelve el repositorio cuando no existe el registro.
	// El mensaje es genérico porque el centinela lo comparten profesionales,
	// bloqueos, usuarios y sesiones: el handler ya arma el texto que ve el
	// cliente según la ruta.
	ErrNoEncontrado = errors.New("recurso no encontrado")

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

	// ErrEmailEnUso lo devuelve el repositorio de usuarios al escribir. El
	// email es la identidad de login y no puede repetirse.
	ErrEmailEnUso = errors.New("email ya registrado")

	// ErrNoAutorizado lo devuelven los servicios cuando el usuario de la
	// sesión no es el dueño del recurso. Es distinto de "no hay sesión": eso
	// lo resuelve el handler con un 401 antes de llegar al servicio.
	ErrNoAutorizado = errors.New("no autorizado")

	// ErrYaTienePerfil lo devuelve el servicio ante un segundo alta de perfil
	// profesional del mismo usuario. Un usuario tiene como máximo uno.
	ErrYaTienePerfil = errors.New("el usuario ya tiene un perfil profesional")

	// ErrCredencialesInvalidas es uno solo para "ese email no existe" y "esa
	// contraseña está mal", a propósito. Distinguirlos convierte al login en
	// un oráculo de qué direcciones están registradas: probando emails contra
	// el endpoint se arma el padrón de usuarios sin adivinar una sola
	// contraseña.
	ErrCredencialesInvalidas = errors.New("credenciales invalidas")
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
