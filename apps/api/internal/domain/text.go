package domain

import (
	"strings"
	"unicode"
)

// Normalizar baja a minúsculas y saca acentos y eñes, para poder comparar
// "González" con "gonzalez".
//
// Lo usan dos cosas: la generación del slug y el filtro de búsqueda del
// listado. Una sola función, un solo lugar donde arreglarla. En un producto
// argentino, una búsqueda que distingue acentos es una búsqueda rota.
func Normalizar(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch r {
		case 'á', 'à', 'ä', 'â', 'ã':
			b.WriteRune('a')
		case 'é', 'è', 'ë', 'ê':
			b.WriteRune('e')
		case 'í', 'ì', 'ï', 'î':
			b.WriteRune('i')
		case 'ó', 'ò', 'ö', 'ô', 'õ':
			b.WriteRune('o')
		case 'ú', 'ù', 'ü', 'û':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		case 'ç':
			b.WriteRune('c')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GenerarSlug genera la parte legible de la URL pública del profesional.
// "Íñigo Muñoz Ríos" → "inigo-munoz-rios"
//
// No se ocupa de la unicidad: eso necesita mirar los demás profesionales y
// por lo tanto vive en el servicio.
func GenerarSlug(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	guionPendiente := false
	for _, r := range Normalizar(s) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			if guionPendiente && b.Len() > 0 {
				b.WriteByte('-')
			}
			guionPendiente = false
			b.WriteRune(r)
		case unicode.IsSpace(r), r == '-', r == '_':
			guionPendiente = true
		}
		// el resto (puntos, comas, símbolos) se descarta sin separar
	}
	return b.String()
}
