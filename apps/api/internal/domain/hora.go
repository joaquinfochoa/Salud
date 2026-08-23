package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// HoraDelDia es una hora de reloj sin fecha: "las nueve de la mañana".
//
// time.Time no sirve para esto porque siempre carga una fecha, y guardar una
// fecha arbitraria para después ignorarla es la clase de convención que alguien
// termina rompiendo. El horario de un profesional se repite todas las semanas:
// no es un instante, es una hora que vuelve.
type HoraDelDia struct {
	Minutos int // desde medianoche. 09:00 son 540.
}

// ParsearHoraDelDia acepta "HH:MM" en formato de 24 horas, con las dos cifras.
//
// Es estricto a propósito: es lo que manda un <input type="time"> del
// navegador, y aceptar "9:5" obligaría a decidir si son las 9:05 o las 9:50.
func ParsearHoraDelDia(s string) (HoraDelDia, error) {
	partes := strings.Split(strings.TrimSpace(s), ":")
	if len(partes) != 2 || len(partes[0]) != 2 || len(partes[1]) != 2 {
		return HoraDelDia{}, errors.New("el formato es HH:MM")
	}

	horas, err := strconv.Atoi(partes[0])
	if err != nil || horas < 0 || horas > 23 {
		return HoraDelDia{}, errors.New("la hora tiene que estar entre 00 y 23")
	}

	minutos, err := strconv.Atoi(partes[1])
	if err != nil || minutos < 0 || minutos > 59 {
		return HoraDelDia{}, errors.New("los minutos tienen que estar entre 00 y 59")
	}

	return HoraDelDia{Minutos: horas*60 + minutos}, nil
}

func (h HoraDelDia) String() string {
	return fmt.Sprintf("%02d:%02d", h.Minutos/60, h.Minutos%60)
}

// Antes ordena dos horas del mismo día.
func (h HoraDelDia) Antes(otra HoraDelDia) bool {
	return h.Minutos < otra.Minutos
}
