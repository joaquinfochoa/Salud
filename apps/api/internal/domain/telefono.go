package domain

import (
	"errors"
	"strings"
)

// Telefono es un número argentino en forma canónica E.164: "+5491112345678".
//
// Es un tipo propio y no un string por la misma razón que Matricula: el mismo
// teléfono se escribe de seis maneras distintas —"11 1234-5678", "011 15
// 1234 5678", "+54 9 11 1234 5678"— y si se guardan tal cual, dos personas con
// el mismo número quedan con dos valores distintos y nada los puede comparar.
type Telefono string

const (
	// Largos del número nacional significativo, sin el 54 ni el 9. En
	// Argentina siempre es 10: área (2 a 4 dígitos) + abonado.
	digitosNacional = 10

	// El código de área más largo del país tiene 4 dígitos.
	maxDigitosArea = 4
)

var limpiadorTelefono = strings.NewReplacer(
	" ", "", "-", "", ".", "", "(", "", ")", "", "/", "",
)

/*
ParsearTelefono acepta las formas que se usan de verdad y las normaliza.

Argentina tiene tres prefijos que la gente escribe y que no son parte del
número:

  - el 0 de larga distancia, antes del código de área ("011", "0341")
  - el 15 de móvil, entre el área y el abonado ("011 15 1234-5678")
  - el 9 que va después del 54 en formato internacional, y que significa
    exactamente lo mismo que el 15

Los tres dicen "esto es un celular" o "estoy llamando de otra ciudad". Acá se
sacan y se reconstruye la forma canónica: +54, un 9 si es móvil, y los diez
dígitos.

La validación es laxa a propósito, igual que con Matricula: rechazar a alguien
con un número real es peor error que aceptar uno raro. Lo único que se exige es
que sea argentino y tenga diez dígitos — un número que no podemos usar es peor
que pedirlo de nuevo.
*/
func ParsearTelefono(s string) (Telefono, error) {
	d := limpiadorTelefono.Replace(strings.TrimSpace(s))
	if d == "" {
		return "", errors.New("no puede estar vacío")
	}

	for i := range d {
		if (d[i] < '0' || d[i] > '9') && !(i == 0 && d[i] == '+') {
			return "", errors.New("solo puede tener números, espacios y guiones")
		}
	}
	d = strings.TrimPrefix(d, "+")

	// Cada formato marca "es un móvil" de una manera distinta: el
	// internacional con el 9 después del 54, el nacional con el 15 después del
	// código de área. Un número pelado no lo marca de ninguna, y ahí hay que
	// decidir.
	movil, explicito := false, false
	switch {
	case strings.HasPrefix(d, "549"):
		d, movil, explicito = d[3:], true, true
	case strings.HasPrefix(d, "54"):
		d, explicito = d[2:], true
	case strings.HasPrefix(d, "0"):
		// Larga distancia nacional: 011…, 0341…
		d, explicito = d[1:], true
	}

	// El 15 va entre el código de área y el abonado, así que hay que saber
	// dónde termina el área para encontrarlo. Se prueba de más corto a más
	// largo: "11" es área de CABA, "341" de Rosario, "2966" de Río Gallegos.
	if len(d) == digitosNacional+2 {
		for largo := 2; largo <= maxDigitosArea; largo++ {
			if len(d) > largo+2 && d[largo:largo+2] == "15" {
				d, movil = d[:largo]+d[largo+2:], true
				break
			}
		}
	}

	// Diez dígitos pelados, sin 0 ni 15 ni país: es ambiguo, y se asume móvil.
	// El campo pide un celular, y en 2026 alguien que escribe su número sin
	// prefijos está escribiendo el del teléfono que tiene en la mano. Quien
	// quiera dar un fijo lo dice, con el 0 o con el +54: los dos formatos
	// tienen cómo, y esta rama no los toca.
	if !explicito {
		movil = true
	}

	if len(d) != digitosNacional {
		return "", errors.New("tiene que ser un número argentino de 10 dígitos, con código de área")
	}

	if movil {
		return Telefono("+549" + d), nil
	}
	return Telefono("+54" + d), nil
}

func (t Telefono) String() string { return string(t) }

// ParaMostrar devuelve la forma legible: "+54 9 11 1234-5678".
//
// El almacenamiento es E.164 porque es lo que espera cualquier proveedor de
// SMS, pero un número pegado de trece dígitos no se lee. Esto existe para las
// pantallas, no para comparar.
func (t Telefono) ParaMostrar() string {
	d := strings.TrimPrefix(string(t), "+54")

	prefijo := "+54 "
	if strings.HasPrefix(d, "9") {
		prefijo, d = "+54 9 ", d[1:]
	}
	if len(d) != digitosNacional {
		return string(t) // no debería pasar: solo se construye por el parser
	}

	// El corte entre área y abonado depende del área, y adivinarlo mal parte
	// el número en el lugar equivocado. Estos son los largos reales.
	area := 3
	switch {
	case strings.HasPrefix(d, "11"):
		area = 2
	case len(d) > 0 && d[0] == '2' && !strings.HasPrefix(d, "23"):
		// Buena parte de las áreas que empiezan con 2 tienen cuatro dígitos.
		area = 4
	}

	resto := d[area:]
	mitad := len(resto) / 2
	return prefijo + d[:area] + " " + resto[:mitad] + "-" + resto[mitad:]
}
