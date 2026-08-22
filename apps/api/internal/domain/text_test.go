package domain

import "testing"

func TestNormalizar(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado string
	}{
		{"minusculas", "GONZÁLEZ", "gonzalez"},
		{"acentos", "Martín González", "martin gonzalez"},
		{"enie", "Muñoz", "munoz"},
		{"dieresis", "Agüero", "aguero"},
		{"todas las vocales", "áéíóú", "aeiou"},
		{"recorta espacios", "  Ana  ", "ana"},
		{"vacio", "", ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := Normalizar(caso.entrada); obtenido != caso.esperado {
				t.Errorf("Normalizar(%q) = %q, se esperaba %q", caso.entrada, obtenido, caso.esperado)
			}
		})
	}
}

func TestGenerarSlug(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado string
	}{
		{"nombre simple", "Martín González", "martin-gonzalez"},
		{"acentos y enies", "Íñigo Muñoz Ríos", "inigo-munoz-rios"},
		{"espacios repetidos", "José  de  la  Cruz", "jose-de-la-cruz"},
		{"puntuacion", "Dr. Juan Pérez", "dr-juan-perez"},
		{"guiones existentes", "Ana-María López", "ana-maria-lopez"},
		{"numeros", "Clínica 24hs", "clinica-24hs"},
		{"solo simbolos", "...", ""},
		{"vacio", "", ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := GenerarSlug(caso.entrada); obtenido != caso.esperado {
				t.Errorf("GenerarSlug(%q) = %q, se esperaba %q", caso.entrada, obtenido, caso.esperado)
			}
		})
	}
}
