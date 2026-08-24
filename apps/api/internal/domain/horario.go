package domain

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

const (
	minDuracionMin = 10
	maxDuracionMin = 480

	// maxBloquesSemana acota cuántos bloques puede tener la semana de un
	// profesional. Siete días de agenda no necesitan más que esto, y sin un
	// tope el verificador de solapamiento —que es O(n²) y acumula un
	// ErrorCampo por cada par que se pisa— convierte un cuerpo de 1 MB en
	// varios GB de errores.
	maxBloquesSemana = 100
)

// HorarioSemanal es un bloque del horario habitual de un profesional. El que
// atiende mañana y tarde tiene dos bloques para el mismo día.
//
// No tiene ID a propósito: la semana se reemplaza entera, así que ningún
// endpoint direcciona un bloque suelto y nada lo referencia. Un ID sería peso
// muerto que además cambiaría en cada guardado. Un horario es un valor, no una
// entidad con identidad.
type HorarioSemanal struct {
	ProfesionalID uuid.UUID
	DiaSemana     DiaSemana
	Desde         HoraDelDia
	Hasta         HoraDelDia
	DuracionMin   int
	Modalidad     Modalidad
}

type EntradaHorarioSemanal struct {
	DiaSemana   string
	Desde       string
	Hasta       string
	DuracionMin int
	Modalidad   string
}

// NuevaSemana valida el conjunto entero de bloques de un profesional.
//
// Valida el conjunto y no bloque por bloque porque el solapamiento solo se ve
// mirando todos juntos, y porque la semana se guarda de una sola vez: no existe
// un estado donde la mitad de los bloques estén cargados.
//
// Una semana vacía es válida: un profesional puede dejar de atender.
func NuevaSemana(profesionalID uuid.UUID, entradas []EntradaHorarioSemanal) ([]HorarioSemanal, error) {
	if len(entradas) > maxBloquesSemana {
		var verr ErrorValidacion
		verr.agregar("horarios", fmt.Sprintf("no puede tener más de %d bloques", maxBloquesSemana))
		return nil, verr
	}

	var verr ErrorValidacion
	bloques := make([]HorarioSemanal, 0, len(entradas))

	for i, entrada := range entradas {
		bloque, errores := construirBloque(profesionalID, entrada)
		for _, e := range errores {
			// el índice ubica al cliente en cuál de los bloques está el problema
			verr.agregar(fmt.Sprintf("horarios[%d].%s", i, e.Campo), e.Mensaje)
		}
		bloques = append(bloques, bloque)
	}

	if verr.tieneErrores() {
		return nil, verr
	}

	if err := verificarSolapamiento(bloques); err != nil {
		return nil, err
	}

	ordenarSemana(bloques)
	return bloques, nil
}

func construirBloque(profesionalID uuid.UUID, entrada EntradaHorarioSemanal) (HorarioSemanal, []ErrorCampo) {
	var errores []ErrorCampo
	bloque := HorarioSemanal{ProfesionalID: profesionalID}

	dia := DiaSemana(strings.ToLower(strings.TrimSpace(entrada.DiaSemana)))
	if !dia.EsValido() {
		errores = append(errores, ErrorCampo{
			Campo:   "diaSemana",
			Mensaje: "debe ser lunes, martes, miercoles, jueves, viernes, sabado o domingo",
		})
	} else {
		bloque.DiaSemana = dia
	}

	modalidad := Modalidad(strings.ToLower(strings.TrimSpace(entrada.Modalidad)))
	if !modalidad.EsValida() {
		errores = append(errores, ErrorCampo{
			Campo:   "modalidad",
			Mensaje: "debe ser telemedicina, presencial o domicilio",
		})
	} else {
		bloque.Modalidad = modalidad
	}

	desde, errDesde := ParsearHoraDelDia(entrada.Desde)
	if errDesde != nil {
		errores = append(errores, ErrorCampo{Campo: "desde", Mensaje: errDesde.Error()})
	} else {
		bloque.Desde = desde
	}

	hasta, errHasta := ParsearHoraDelDia(entrada.Hasta)
	if errHasta != nil {
		errores = append(errores, ErrorCampo{Campo: "hasta", Mensaje: errHasta.Error()})
	} else {
		bloque.Hasta = hasta
	}

	duracionValida := entrada.DuracionMin >= minDuracionMin && entrada.DuracionMin <= maxDuracionMin
	switch {
	case entrada.DuracionMin < minDuracionMin:
		errores = append(errores, ErrorCampo{
			Campo:   "duracionMin",
			Mensaje: fmt.Sprintf("no puede ser menor a %d minutos", minDuracionMin),
		})
	case entrada.DuracionMin > maxDuracionMin:
		errores = append(errores, ErrorCampo{
			Campo:   "duracionMin",
			Mensaje: fmt.Sprintf("no puede superar los %d minutos", maxDuracionMin),
		})
	default:
		bloque.DuracionMin = entrada.DuracionMin
	}

	// Las dos reglas que siguen comparan campos entre sí, así que solo tienen
	// sentido si los campos individuales parsearon bien.
	if errDesde != nil || errHasta != nil || !duracionValida {
		return bloque, errores
	}

	if !bloque.Desde.Antes(bloque.Hasta) {
		errores = append(errores, ErrorCampo{Campo: "hasta", Mensaje: "tiene que ser posterior a desde"})
		return bloque, errores
	}

	// Un bloque donde no entra ni una sesión no genera un solo turno. Es mejor
	// que el profesional se entere ahora, y no dos semanas después al no
	// recibir nada.
	if bloque.Hasta.Minutos-bloque.Desde.Minutos < bloque.DuracionMin {
		errores = append(errores, ErrorCampo{
			Campo:   "duracionMin",
			Mensaje: "no entra ninguna sesión en ese bloque",
		})
	}

	return bloque, errores
}

func verificarSolapamiento(bloques []HorarioSemanal) error {
	var verr ErrorValidacion

	for i := range bloques {
		for j := i + 1; j < len(bloques); j++ {
			a, b := bloques[i], bloques[j]
			if a.DiaSemana != b.DiaSemana {
				continue
			}
			// intervalos semiabiertos: uno que termina 13:00 y otro que empieza
			// 13:00 son contiguos, no solapados
			if a.Desde.Minutos < b.Hasta.Minutos && b.Desde.Minutos < a.Hasta.Minutos {
				verr.agregar(
					fmt.Sprintf("horarios[%d]", j),
					fmt.Sprintf("se solapa con el bloque %d del mismo día", i),
				)
			}
		}
	}

	if verr.tieneErrores() {
		return verr
	}
	return nil
}

// ordenarSemana deja la semana en un orden estable para que la respuesta no
// dependa de en qué orden la mandó el cliente.
func ordenarSemana(bloques []HorarioSemanal) {
	slices.SortFunc(bloques, func(a, b HorarioSemanal) int {
		if c := a.DiaSemana.Orden() - b.DiaSemana.Orden(); c != 0 {
			return c
		}
		return a.Desde.Minutos - b.Desde.Minutos
	})
}
