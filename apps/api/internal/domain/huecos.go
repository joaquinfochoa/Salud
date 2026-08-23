package domain

import (
	"slices"
	"time"
)

// Hueco es un turno reservable: un intervalo concreto en el que un profesional
// atiende y nadie tomó todavía.
type Hueco struct {
	Inicio    time.Time
	Fin       time.Time
	Modalidad Modalidad
}

// CalculoHuecos son todas las entradas del cálculo, explícitas.
//
// Es un struct y no una función con seis parámetros para que nadie confunda el
// orden de dos time.Time, y porque así los tests se leen.
//
// No recibe repositorios ni el reloj del sistema: el servicio carga los datos y
// arma esto. Esa separación es la que permite probar los bordes del calendario
// sin levantar nada.
//
// Cuando exista Turno, se le suma un campo con los turnos ya tomados y se
// restan igual que los bloqueos. La firma pública no cambia.
type CalculoHuecos struct {
	Horarios []HorarioSemanal
	Bloqueos []Bloqueo

	// Desde y Hasta acotan el cálculo como intervalo semiabierto [Desde,
	// Hasta). El servicio traduce las fechas de la consulta —que sí incluyen
	// los dos días— a este formato: del 25 al 27 llega acá como [25 00:00, 28
	// 00:00).
	Desde time.Time
	Hasta time.Time

	AnticipacionMinimaMin int
	Ahora                 time.Time
}

func (c CalculoHuecos) Generar() []Hueco {
	huecos := make([]Hueco, 0)
	if !c.Desde.Before(c.Hasta) {
		return huecos
	}

	minimo := c.Ahora.Add(time.Duration(c.AnticipacionMinimaMin) * time.Minute)

	for dia := InicioDelDia(c.Desde); dia.Before(c.Hasta); dia = dia.AddDate(0, 0, 1) {
		diaSemana := DiaSemanaDe(dia.Weekday())

		for _, bloque := range c.Horarios {
			if bloque.DiaSemana != diaSemana {
				continue
			}
			huecos = append(huecos, c.huecosDelBloque(dia, bloque, minimo)...)
		}
	}

	slices.SortFunc(huecos, func(a, b Hueco) int { return a.Inicio.Compare(b.Inicio) })
	return huecos
}

func (c CalculoHuecos) huecosDelBloque(dia time.Time, bloque HorarioSemanal, minimo time.Time) []Hueco {
	var huecos []Hueco
	duracion := time.Duration(bloque.DuracionMin) * time.Minute

	for minuto := bloque.Desde.Minutos; minuto+bloque.DuracionMin <= bloque.Hasta.Minutos; minuto += bloque.DuracionMin {
		// Se construye con time.Date y no sumándole una duración al inicio del
		// día: sumar es aritmética de instantes, y si el país vuelve a tener
		// horario de verano eso correría la hora de reloj. time.Date respeta
		// que "las nueve" son las nueve.
		inicio := time.Date(dia.Year(), dia.Month(), dia.Day(), minuto/60, minuto%60, 0, 0, ZonaHoraria)
		fin := inicio.Add(duracion)

		switch {
		case inicio.Before(c.Desde) || fin.After(c.Hasta):
			continue // se sale del rango pedido
		case inicio.Before(minimo):
			continue // no llega a la anticipación mínima
		case c.bloqueado(inicio, fin):
			continue
		}

		huecos = append(huecos, Hueco{Inicio: inicio, Fin: fin, Modalidad: bloque.Modalidad})
	}

	return huecos
}

func (c CalculoHuecos) bloqueado(inicio, fin time.Time) bool {
	for _, bloqueo := range c.Bloqueos {
		if bloqueo.SeSolapaCon(inicio, fin) {
			return true
		}
	}
	return false
}
