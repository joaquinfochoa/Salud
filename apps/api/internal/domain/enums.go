package domain

import (
	"slices"
	"time"
)

// Especialidad es lo que el profesional atiende.
//
// Enum cerrado a propósito: con texto libre terminás con "Psicología",
// "psicologia" y "Psicólogo clínico" como tres especialidades distintas, y los
// filtros de búsqueda —que son cómo un paciente encuentra a alguien— dejan de
// servir.
//
// La lista dejó de ser "los tres verticales de lanzamiento" cuando se decidió
// que la plataforma es multi-área: alguien entra por cualquier motivo de
// consulta y busca a quien lo atienda. Agregar una es una línea en
// Especialidades; la única condición es que sea una profesión con matrícula
// que atiende por turno.
//
// Los valores no llevan acentos ni ñ, como todo identificador del proyecto: el
// nombre para mostrar lo pone el front.
type Especialidad string

// Las tres primeras siguen como constantes porque el seed y los tests las
// nombran. El resto vive solo en la lista: veinticinco constantes que nadie
// referencia serían veinticinco líneas de ruido.
const (
	EspecialidadPsicologia   Especialidad = "psicologia"
	EspecialidadKinesiologia Especialidad = "kinesiologia"
	EspecialidadOdontologia  Especialidad = "odontologia"
)

// Especialidades es la lista completa, en orden alfabético.
//
// Está exportada porque el handler arma su mensaje de error con ella: una
// lista de valores válidos escrita a mano en otro paquete se desactualiza en
// cuanto alguien agrega uno acá.
var Especialidades = []Especialidad{
	"alergia-e-inmunologia",
	"cardiologia",
	"clinica-medica",
	"dermatologia",
	"endocrinologia",
	"enfermeria",
	"fonoaudiologia",
	"gastroenterologia",
	"ginecologia",
	EspecialidadKinesiologia,
	"neumonologia",
	"neurologia",
	"nutricion",
	"obstetricia",
	EspecialidadOdontologia,
	"oftalmologia",
	"otorrinolaringologia",
	"pediatria",
	"podologia",
	EspecialidadPsicologia,
	"psicopedagogia",
	"psiquiatria",
	"reumatologia",
	"terapia-ocupacional",
	"traumatologia",
	"urologia",
}

func (e Especialidad) EsValida() bool {
	return slices.Contains(Especialidades, e)
}

type Modalidad string

const (
	ModalidadTelemedicina Modalidad = "telemedicina"
	ModalidadPresencial   Modalidad = "presencial"
	ModalidadDomicilio    Modalidad = "domicilio"
)

func (m Modalidad) EsValida() bool {
	switch m {
	case ModalidadTelemedicina, ModalidadPresencial, ModalidadDomicilio:
		return true
	}
	return false
}

// Estado dice si el profesional opera hoy en la plataforma.
//
// No confundir con EstadoVerificacion: son dos ejes distintos. Un profesional
// puede estar verificado y de licencia, o recién anotado y sin verificar.
type Estado string

const (
	EstadoActivo   Estado = "activo"
	EstadoInactivo Estado = "inactivo"
)

func (s Estado) EsValido() bool {
	switch s {
	case EstadoActivo, EstadoInactivo:
		return true
	}
	return false
}

// EstadoVerificacion dice si la matrícula fue verificada contra el mundo real.
// Por ahora todos nacen en pendiente: la integración con REFEPS es una etapa
// posterior y nada la mueve automáticamente todavía.
type EstadoVerificacion string

const (
	VerificacionPendiente  EstadoVerificacion = "pendiente"
	VerificacionVerificada EstadoVerificacion = "verificada"
	VerificacionRechazada  EstadoVerificacion = "rechazada"
)

func (v EstadoVerificacion) EsValido() bool {
	switch v {
	case VerificacionPendiente, VerificacionVerificada, VerificacionRechazada:
		return true
	}
	return false
}

// DiaSemana usa el mismo vocabulario que el resto de los enums: español y sin
// acentos.
//
// Existe en vez de usar time.Weekday directo para mantener la frontera del
// dominio y para no exponer una numeración donde el domingo vale cero, que es
// una convención heredada de C que nadie recuerda de memoria.
type DiaSemana string

const (
	DiaLunes     DiaSemana = "lunes"
	DiaMartes    DiaSemana = "martes"
	DiaMiercoles DiaSemana = "miercoles"
	DiaJueves    DiaSemana = "jueves"
	DiaViernes   DiaSemana = "viernes"
	DiaSabado    DiaSemana = "sabado"
	DiaDomingo   DiaSemana = "domingo"
)

var diasPorWeekday = map[time.Weekday]DiaSemana{
	time.Sunday:    DiaDomingo,
	time.Monday:    DiaLunes,
	time.Tuesday:   DiaMartes,
	time.Wednesday: DiaMiercoles,
	time.Thursday:  DiaJueves,
	time.Friday:    DiaViernes,
	time.Saturday:  DiaSabado,
}

// El mapa inverso se deriva del primero en vez de escribirse a mano: dos
// literales pueden desincronizarse y nadie lo nota hasta que la agenda muestra
// el día equivocado.
var weekdaysPorDia = func() map[DiaSemana]time.Weekday {
	m := make(map[DiaSemana]time.Weekday, len(diasPorWeekday))
	for weekday, dia := range diasPorWeekday {
		m[dia] = weekday
	}
	return m
}()

func (d DiaSemana) EsValido() bool {
	_, existe := weekdaysPorDia[d]
	return existe
}

func (d DiaSemana) AWeekday() time.Weekday {
	return weekdaysPorDia[d]
}

// Orden pone el lunes en cero, que es como se lee una agenda acá. time.Weekday
// arranca en domingo y ordenar por ese número dejaría el domingo primero.
func (d DiaSemana) Orden() int {
	return (int(d.AWeekday()) + 6) % 7
}

func DiaSemanaDe(w time.Weekday) DiaSemana {
	return diasPorWeekday[w]
}
