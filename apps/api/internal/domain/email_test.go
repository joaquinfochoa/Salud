package domain

import "testing"

func TestParsearEmail(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado Email
		falla    bool
	}{
		{"normaliza a minusculas", "Juan@Ejemplo.COM", "juan@ejemplo.com", false},
		{"recorta espacios", "  juan@ejemplo.com  ", "juan@ejemplo.com", false},
		{"acepta subdominio", "juan@mail.ejemplo.com.ar", "juan@mail.ejemplo.com.ar", false},
		{"acepta mas y punto", "juan.perez+turnos@ejemplo.com", "juan.perez+turnos@ejemplo.com", false},
		{"vacio falla", "", "", true},
		{"solo espacios falla", "   ", "", true},
		{"sin arroba falla", "juanejemplo.com", "", true},
		{"sin dominio falla", "juan@", "", true},
		// mail.ParseAddress acepta la forma con nombre visible. Un email de
		// login no es un encabezado de correo: si entra así, el usuario se
		// registra con una dirección distinta a la que escribió.
		{"con nombre visible falla", "Juan Perez <juan@ejemplo.com>", "", true},
		// Un salto al final es basura de formulario y se recorta como
		// cualquier espacio. Uno en el medio es otra cosa: es el intento de
		// meter un encabezado de correo dentro de la dirección, y tiene que
		// morir acá.
		{"salto al final se recorta", "juan@ejemplo.com\n", "juan@ejemplo.com", false},
		{"salto en el medio falla", "juan@ejemplo.com\nBcc: otro@mal.com", "", true},
		{"retorno de carro en el medio falla", "juan\r\n@ejemplo.com", "", true},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e, err := ParsearEmail(c.entrada)
			if c.falla {
				if err == nil {
					t.Fatalf("ParsearEmail(%q) = %q, se esperaba error", c.entrada, e)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsearEmail(%q) devolvió error: %v", c.entrada, err)
			}
			if e != c.esperado {
				t.Errorf("ParsearEmail(%q) = %q, se esperaba %q", c.entrada, e, c.esperado)
			}
		})
	}
}

func TestEmailDemasiadoLargo(t *testing.T) {
	largo := ""
	for range 250 {
		largo += "a"
	}
	if _, err := ParsearEmail(largo + "@ejemplo.com"); err == nil {
		t.Error("se esperaba error por longitud")
	}
}
