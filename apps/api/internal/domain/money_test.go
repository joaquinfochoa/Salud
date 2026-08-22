package domain

import "testing"

func TestDineroString(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  Dinero
		esperado string
	}{
		{"cero", 0, "$0,00"},
		{"un peso", 100, "$1,00"},
		{"centavos sueltos", 5, "$0,05"},
		{"menos de un peso", 999, "$9,99"},
		{"miles", 1200000, "$12.000,00"},
		{"millones", 123456789, "$1.234.567,89"},
		{"tres digitos", 99900, "$999,00"},
		{"negativo", -50, "-$0,50"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := caso.entrada.String(); obtenido != caso.esperado {
				t.Errorf("Dinero(%d).String() = %q, se esperaba %q", int64(caso.entrada), obtenido, caso.esperado)
			}
		})
	}
}
