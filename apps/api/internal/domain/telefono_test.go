package domain

import "testing"

func TestParsearTelefono(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado string
		falla    bool
	}{
		// Las formas que la gente escribe de verdad. Todas son el mismo
		// teléfono de CABA y tienen que converger al mismo valor.
		{"con +54 9 y espacios", "+54 9 11 1234-5678", "+5491112345678", false},
		{"con +54 9 pegado", "+5491112345678", "+5491112345678", false},
		{"con 0 y 15", "011 15 1234-5678", "+5491112345678", false},
		{"con 0 y 15 pegados", "0111512345678", "+5491112345678", false},
		{"sin 0 ni 15", "11 1234 5678", "+5491112345678", false},
		{"con parentesis", "(011) 15 1234-5678", "+5491112345678", false},
		{"con puntos", "11.1234.5678", "+5491112345678", false},

		// Un interior: Rosario es 341 y el abonado tiene 7 dígitos.
		{"interior con 0 y 15", "0341 15 456-7890", "+5493414567890", false},
		{"interior sin prefijos", "341 456 7890", "+5493414567890", false},

		// Un fijo. No se le inventa el 9: el 9 dice "es un móvil", y mandarle
		// un SMS a un fijo no llega. En formato internacional lo marca la
		// ausencia del 9; en nacional, la del 15.
		{"fijo en internacional", "+54 11 4321-1000", "+541143211000", false},
		{"fijo en nacional", "(011) 4321-1000", "+541143211000", false},
		{"fijo del interior", "0341 456-7890", "+543414567890", false},

		{"vacio falla", "", "", true},
		{"solo espacios falla", "   ", "", true},
		{"con letras falla", "11 1234 ABCD", "", true},
		{"demasiado corto falla", "1234", "", true},
		{"demasiado largo falla", "+54 9 11 1234 5678 9012", "", true},
		// Un número de otro país no se rechaza por ser de otro país: se
		// rechaza porque hoy no sabemos validarlo, y guardar algo que no
		// podemos usar es peor que pedirlo de nuevo.
		{"otro pais falla", "+1 415 555 2671", "", true},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			tel, err := ParsearTelefono(c.entrada)
			if c.falla {
				if err == nil {
					t.Fatalf("ParsearTelefono(%q) = %q, se esperaba error", c.entrada, tel)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsearTelefono(%q) devolvió error: %v", c.entrada, err)
			}
			if tel.String() != c.esperado {
				t.Errorf("ParsearTelefono(%q) = %q, se esperaba %q", c.entrada, tel, c.esperado)
			}
		})
	}
}

// Dos personas que escriben el mismo teléfono distinto tienen que quedar con
// el mismo valor guardado. Es la misma razón por la que Matricula converge.
func TestTelefonosEquivalentesConvergen(t *testing.T) {
	formas := []string{
		"+54 9 11 1234-5678",
		"011 15 1234-5678",
		"11 1234 5678",
		"(011) 15 1234 5678",
		"11.1234.5678",
	}

	var primero Telefono
	for i, f := range formas {
		tel, err := ParsearTelefono(f)
		if err != nil {
			t.Fatalf("ParsearTelefono(%q) devolvió error: %v", f, err)
		}
		if i == 0 {
			primero = tel
			continue
		}
		if tel != primero {
			t.Errorf("%q dio %q, pero %q dio %q", f, tel, formas[0], primero)
		}
	}
}

func TestTelefonoParaMostrar(t *testing.T) {
	casos := []struct{ guardado, esperado string }{
		{"+5491112345678", "+54 9 11 1234-5678"},
		{"+5493414567890", "+54 9 341 456-7890"},
		{"+541143211000", "+54 11 4321-1000"},
	}

	for _, c := range casos {
		tel, err := ParsearTelefono(c.guardado)
		if err != nil {
			t.Fatalf("ParsearTelefono(%q) devolvió error: %v", c.guardado, err)
		}
		if got := tel.ParaMostrar(); got != c.esperado {
			t.Errorf("ParaMostrar(%q) = %q, se esperaba %q", c.guardado, got, c.esperado)
		}
	}
}
