package domain

import (
	"errors"
	"strings"
)

type MatriculaTipo string

const (
	MatriculaNacional   MatriculaTipo = "MN"
	MatriculaProvincial MatriculaTipo = "MP"
)

const maxDigitosMatricula = 10

// Matricula es la identidad profesional de una persona: es el único dato que
// la ata a una habilitación real y es sobre lo que se apoya toda la confianza
// del producto.
type Matricula struct {
	Tipo   MatriculaTipo
	Numero string
}

var limpiadorMatricula = strings.NewReplacer(".", "", " ", "", "-", "", "/", "")

// ParsearMatricula acepta las formas que se usan en la práctica —"MN 98.234",
// "M.N. 45321", "mn98234", "MP 12345"— y las normaliza a "MN 98234".
//
// La validación es deliberadamente laxa. Las matrículas argentinas varían por
// jurisdicción y por profesión, y rechazar a un profesional real es peor error
// que aceptar un número raro: el que queda afuera no vuelve. La verificación
// seria llega cuando exista la integración con REFEPS.
func ParsearMatricula(s string) (Matricula, error) {
	limpia := limpiadorMatricula.Replace(strings.ToUpper(s))

	if len(limpia) < 3 {
		return Matricula{}, errors.New("debe tener tipo (MN o MP) y número")
	}

	tipo := MatriculaTipo(limpia[:2])
	if tipo != MatriculaNacional && tipo != MatriculaProvincial {
		return Matricula{}, errors.New("el tipo debe ser MN (nacional) o MP (provincial)")
	}

	numero := limpia[2:]
	if len(numero) > maxDigitosMatricula {
		return Matricula{}, errors.New("el número no puede tener más de 10 dígitos")
	}
	for i := range numero {
		if numero[i] < '0' || numero[i] > '9' {
			return Matricula{}, errors.New("el número solo puede contener dígitos")
		}
	}

	return Matricula{Tipo: tipo, Numero: numero}, nil
}

// String devuelve la forma canónica. Dos matrículas escritas distinto pero
// iguales en el fondo comparan iguales porque el parser las converge acá.
func (m Matricula) String() string {
	return string(m.Tipo) + " " + m.Numero
}

func (m Matricula) EsCero() bool {
	return m.Tipo == "" && m.Numero == ""
}
