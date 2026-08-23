package domain

import "testing"

func TestParsearHoraDelDiaValida(t *testing.T) {
	casos := []struct {
		entrada string
		minutos int
	}{
		{"00:00", 0},
		{"09:00", 540},
		{"13:30", 810},
		{"23:59", 1439},
		{" 09:00 ", 540},
	}

	for _, caso := range casos {
		t.Run(caso.entrada, func(t *testing.T) {
			h, err := ParsearHoraDelDia(caso.entrada)
			if err != nil {
				t.Fatalf("ParsearHoraDelDia(%q) devolvió error: %v", caso.entrada, err)
			}
			if h.Minutos != caso.minutos {
				t.Errorf("minutos = %d, se esperaba %d", h.Minutos, caso.minutos)
			}
		})
	}
}

func TestParsearHoraDelDiaInvalida(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
	}{
		{"vacia", ""},
		{"sin separador", "0900"},
		{"una sola cifra en la hora", "9:00"},
		{"una sola cifra en los minutos", "09:0"},
		{"hora fuera de rango", "24:00"},
		{"minutos fuera de rango", "09:60"},
		{"con segundos", "09:00:00"},
		{"letras", "ab:cd"},
		{"negativa", "-1:00"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if _, err := ParsearHoraDelDia(caso.entrada); err == nil {
				t.Errorf("ParsearHoraDelDia(%q) debía fallar y no falló", caso.entrada)
			}
		})
	}
}

func TestHoraDelDiaString(t *testing.T) {
	casos := []struct {
		minutos int
		esperado string
	}{
		{0, "00:00"},
		{540, "09:00"},
		{810, "13:30"},
		{1439, "23:59"},
	}

	for _, caso := range casos {
		t.Run(caso.esperado, func(t *testing.T) {
			if obtenido := (HoraDelDia{Minutos: caso.minutos}).String(); obtenido != caso.esperado {
				t.Errorf("String() = %q, se esperaba %q", obtenido, caso.esperado)
			}
		})
	}
}

func TestHoraDelDiaIdaYVuelta(t *testing.T) {
	// lo que sale de String() tiene que volver a entrar por el parser
	for minutos := 0; minutos < 24*60; minutos++ {
		original := HoraDelDia{Minutos: minutos}
		vuelta, err := ParsearHoraDelDia(original.String())
		if err != nil {
			t.Fatalf("no se pudo reparsear %q: %v", original, err)
		}
		if vuelta != original {
			t.Fatalf("ida y vuelta de %q dio %q", original, vuelta)
		}
	}
}

func TestHoraDelDiaAntes(t *testing.T) {
	nueve := HoraDelDia{Minutos: 540}
	trece := HoraDelDia{Minutos: 780}

	if !nueve.Antes(trece) {
		t.Error("09:00 tenía que ser antes de 13:00")
	}
	if trece.Antes(nueve) {
		t.Error("13:00 no tenía que ser antes de 09:00")
	}
	if nueve.Antes(nueve) {
		t.Error("una hora no es anterior a sí misma")
	}
}
