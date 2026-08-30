package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// maxLargoEmail es el máximo del RFC 5321: 64 de parte local más @ más 253 de
// dominio daría 318, pero la ruta de retorno completa está acotada a 254. Es
// el número que usan los sistemas reales.
const maxLargoEmail = 254

// Email es una dirección normalizada. Existe como tipo y no como string para
// que "Juan@Ejemplo.COM" y "juan@ejemplo.com" no puedan ser dos usuarios
// distintos: el parser las converge antes de que nada las compare.
type Email string

// ParsearEmail normaliza y valida.
//
// La validación se apoya en net/mail y no en una expresión regular propia. Una
// regex de email escrita a mano rechaza direcciones válidas —y las direcciones
// válidas son mucho más raras de lo que parece— y ese usuario no vuelve. Lo
// que sí se agrega sobre net/mail es rechazar la forma con nombre visible:
// "Juan <juan@ej.com>" es un encabezado de correo legítimo, pero como
// credencial de login registraría al usuario con una dirección distinta a la
// que tipeó.
func ParsearEmail(s string) (Email, error) {
	limpio := strings.ToLower(strings.TrimSpace(s))

	if limpio == "" {
		return "", errors.New("es obligatorio")
	}
	if len(limpio) > maxLargoEmail {
		return "", errors.New("no puede superar los 254 caracteres")
	}

	dir, err := mail.ParseAddress(limpio)
	if err != nil || dir.Address != limpio {
		return "", errors.New("no tiene un formato válido")
	}
	return Email(limpio), nil
}

func (e Email) String() string {
	return string(e)
}
