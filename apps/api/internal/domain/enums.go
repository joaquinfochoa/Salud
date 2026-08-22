package domain

// Especialidad son los tres verticales de lanzamiento definidos en
// research/data/vertical_scores.csv. Es un enum cerrado a propósito: con texto
// libre terminás con "Psicología", "psicologia" y "Psicólogo clínico" como tres
// especialidades distintas, y los filtros dejan de servir.
//
// Agregar una cuarta es una constante y un caso más en EsValida().
type Especialidad string

const (
	EspecialidadPsicologia   Especialidad = "psicologia"
	EspecialidadKinesiologia Especialidad = "kinesiologia"
	EspecialidadOdontologia  Especialidad = "odontologia"
)

func (e Especialidad) EsValida() bool {
	switch e {
	case EspecialidadPsicologia, EspecialidadKinesiologia, EspecialidadOdontologia:
		return true
	}
	return false
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
