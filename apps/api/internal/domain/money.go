package domain

import (
	"strconv"
	"strings"
)

// Dinero representa un monto en centavos.
//
// Nunca usar float para dinero: float64 no puede representar 0,10 de forma
// exacta, y este sistema va a cobrar consultas y liquidar honorarios. El tipo
// propio además impide sumar un precio con una cantidad por accidente.
type Dinero int64

// String formatea el monto con la convención argentina: punto para miles,
// coma para decimales. Dinero(1200000) → "$12.000,00"
func (m Dinero) String() string {
	negativo := m < 0
	if negativo {
		m = -m
	}

	pesos := int64(m) / 100
	centavos := int64(m) % 100

	digitos := strconv.FormatInt(pesos, 10)
	var b strings.Builder
	for i := range digitos {
		if i > 0 && (len(digitos)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(digitos[i])
	}

	centavosStr := strconv.FormatInt(centavos, 10)
	if centavos < 10 {
		centavosStr = "0" + centavosStr
	}

	salida := "$" + b.String() + "," + centavosStr
	if negativo {
		return "-" + salida
	}
	return salida
}
