package domain

import "testing"

func TestParsearMatriculaValida(t *testing.T) {
	casos := []struct {
		nombre         string
		entrada        string
		tipoEsperado   MatriculaTipo
		numeroEsperado string
	}{
		{"formato canonico", "MN 98234", MatriculaNacional, "98234"},
		{"con puntos de miles", "MN 98.234", MatriculaNacional, "98234"},
		{"con puntos en el tipo", "M.N. 45321", MatriculaNacional, "45321"},
		{"sin espacios ni puntos", "mn98234", MatriculaNacional, "98234"},
		{"provincial", "MP 12345", MatriculaProvincial, "12345"},
		{"minusculas", "mp 12345", MatriculaProvincial, "12345"},
		{"con guion", "MN-98234", MatriculaNacional, "98234"},
		{"un solo digito", "MN 7", MatriculaNacional, "7"},
		{"diez digitos", "MN 1234567890", MatriculaNacional, "1234567890"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			m, err := ParsearMatricula(caso.entrada)
			if err != nil {
				t.Fatalf("ParsearMatricula(%q) devolvió error: %v", caso.entrada, err)
			}
			if m.Tipo != caso.tipoEsperado {
				t.Errorf("tipo = %q, se esperaba %q", m.Tipo, caso.tipoEsperado)
			}
			if m.Numero != caso.numeroEsperado {
				t.Errorf("numero = %q, se esperaba %q", m.Numero, caso.numeroEsperado)
			}
		})
	}
}

func TestParsearMatriculaInvalida(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
	}{
		{"vacia", ""},
		{"solo el tipo", "MN"},
		{"sin tipo", "98234"},
		{"tipo desconocido", "XX 98234"},
		{"numero con letras", "MN 98A34"},
		{"mas de diez digitos", "MN 12345678901"},
		{"solo espacios", "   "},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if _, err := ParsearMatricula(caso.entrada); err == nil {
				t.Errorf("ParsearMatricula(%q) debía fallar y no falló", caso.entrada)
			}
		})
	}
}

func TestMatriculaString(t *testing.T) {
	m, err := ParsearMatricula("m.n. 98.234")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// distintas formas de escribir la misma matrícula tienen que converger
	// a una sola representación, o la unicidad no sirve de nada
	if obtenido := m.String(); obtenido != "MN 98234" {
		t.Errorf("String() = %q, se esperaba %q", obtenido, "MN 98234")
	}
}
