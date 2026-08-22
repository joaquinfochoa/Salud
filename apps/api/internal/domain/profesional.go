package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxLargoNombre = 100
	maxLargoBio    = 2000
	maxLargoZona   = 100
)

// Profesional es un profesional de la salud dado de alta en la plataforma.
//
// Invariante del paquete: no se puede construir uno inválido desde afuera.
// No hay setters públicos; la única puerta de entrada es NuevoProfesional, y la
// única forma de modificarlo es AplicarCambios, que revalida todo.
type Profesional struct {
	ID             uuid.UUID
	Slug           string
	Nombre         string
	Apellido       string
	Matricula      Matricula
	Especialidad   Especialidad
	Bio            string
	PrecioConsulta Dinero
	Modalidades    []Modalidad
	Zona           string
	ObrasSociales  []string
	Estado         Estado
	Verificacion   EstadoVerificacion
	CreadoEn       time.Time
	ActualizadoEn  time.Time
	DadoDeBajaEn   *time.Time
}

// EntradaProfesional es la entrada cruda, en tipos primitivos. Que sea primitiva
// no es descuido: obliga a que todo el parseo y toda la validación ocurran acá
// adentro, y no repartidos por los handlers.
type EntradaProfesional struct {
	Nombre         string
	Apellido       string
	Matricula      string
	Especialidad   string
	Bio            string
	PrecioConsulta int64
	Modalidades    []string
	Zona           string
	ObrasSociales  []string
}

// NuevoProfesional valida la entrada y devuelve un profesional consistente o un
// ErrorValidacion con todos los campos que fallaron.
func NuevoProfesional(entrada EntradaProfesional, ahora time.Time) (Profesional, error) {
	p, verr := construir(entrada)
	if verr.tieneErrores() {
		return Profesional{}, verr
	}

	p.ID = uuid.New()
	p.Slug = GenerarSlug(p.NombreCompleto())
	if p.Slug == "" {
		// el nombre pasó la validación pero no dejó ningún carácter usable
		// (por ejemplo "..."). Sin esto quedaría un slug vacío y la URL
		// pública del profesional colisionaría con la de cualquier otro.
		p.Slug = p.ID.String()
	}
	p.Estado = EstadoActivo
	p.Verificacion = VerificacionPendiente
	p.CreadoEn = ahora
	p.ActualizadoEn = ahora

	return p, nil
}

// AplicarCambios reemplaza los campos editables y devuelve el resultado sin tocar
// el receptor. ID, Slug, Estado, CreadoEn y DadoDeBajaEn no son editables.
func (p Profesional) AplicarCambios(entrada EntradaProfesional, ahora time.Time) (Profesional, error) {
	actualizado, verr := construir(entrada)
	if verr.tieneErrores() {
		return Profesional{}, verr
	}

	actualizado.ID = p.ID
	actualizado.Slug = p.Slug
	actualizado.Estado = p.Estado
	actualizado.CreadoEn = p.CreadoEn
	actualizado.DadoDeBajaEn = p.DadoDeBajaEn
	actualizado.ActualizadoEn = ahora

	// La verificación se hizo sobre una matrícula y una especialidad
	// concretas. Si cambian, deja de valer: toda orientación, agenda o cobro
	// depende de que el profesional esté verificado.
	if actualizado.Matricula != p.Matricula || actualizado.Especialidad != p.Especialidad {
		actualizado.Verificacion = VerificacionPendiente
	} else {
		actualizado.Verificacion = p.Verificacion
	}

	return actualizado, nil
}

// DarDeBaja da de baja al profesional. No es un borrado: los turnos y
// comprobantes históricos siguen apuntando a este registro. Es idempotente y
// no corre la fecha de la primera baja.
func (p Profesional) DarDeBaja(ahora time.Time) Profesional {
	if p.Estado == EstadoInactivo {
		return p
	}
	p.Estado = EstadoInactivo
	p.DadoDeBajaEn = &ahora
	p.ActualizadoEn = ahora
	return p
}

// Reactivar revierte la baja. Idempotente.
func (p Profesional) Reactivar(ahora time.Time) Profesional {
	if p.Estado == EstadoActivo {
		return p
	}
	p.Estado = EstadoActivo
	p.DadoDeBajaEn = nil
	p.ActualizadoEn = ahora
	return p
}

// Clonar devuelve una copia profunda.
//
// Una copia superficial comparte el array que hay debajo de los slices, y deja
// que quien la reciba mute el original desde afuera sin enterarse. Es el bug
// número uno de un repositorio en memoria.
func (p Profesional) Clonar() Profesional {
	c := p
	c.Modalidades = slices.Clone(p.Modalidades)
	c.ObrasSociales = slices.Clone(p.ObrasSociales)
	if p.DadoDeBajaEn != nil {
		t := *p.DadoDeBajaEn
		c.DadoDeBajaEn = &t
	}
	return c
}

func (p Profesional) NombreCompleto() string {
	return p.Nombre + " " + p.Apellido
}

// construir parsea y valida la entrada, acumulando todos los errores. Es la única
// implementación de las reglas: la comparten el alta y la edición.
func construir(entrada EntradaProfesional) (Profesional, ErrorValidacion) {
	var p Profesional
	var verr ErrorValidacion

	p.Nombre = validarNombre(entrada.Nombre, "nombre", &verr)
	p.Apellido = validarNombre(entrada.Apellido, "apellido", &verr)

	if m, err := ParsearMatricula(entrada.Matricula); err != nil {
		verr.agregar("matricula", err.Error())
	} else {
		p.Matricula = m
	}

	esp := Especialidad(strings.ToLower(strings.TrimSpace(entrada.Especialidad)))
	if !esp.EsValida() {
		verr.agregar("especialidad", "debe ser psicologia, kinesiologia u odontologia")
	} else {
		p.Especialidad = esp
	}

	p.Bio = strings.TrimSpace(entrada.Bio)
	if utf8.RuneCountInString(p.Bio) > maxLargoBio {
		verr.agregar("bio", fmt.Sprintf("no puede superar los %d caracteres", maxLargoBio))
	}

	if entrada.PrecioConsulta < 0 {
		verr.agregar("precioConsultaCentavos", "no puede ser negativo")
	} else {
		p.PrecioConsulta = Dinero(entrada.PrecioConsulta)
	}

	p.Modalidades = construirModalidades(entrada.Modalidades, &verr)

	p.Zona = strings.TrimSpace(entrada.Zona)
	switch {
	case p.Zona == "":
		verr.agregar("zona", "es obligatoria")
	case utf8.RuneCountInString(p.Zona) > maxLargoZona:
		verr.agregar("zona", fmt.Sprintf("no puede superar los %d caracteres", maxLargoZona))
	}

	p.ObrasSociales = construirObrasSociales(entrada.ObrasSociales, &verr)

	return p, verr
}

func validarNombre(crudo, campo string, verr *ErrorValidacion) string {
	nombre := strings.TrimSpace(crudo)
	switch {
	case nombre == "":
		verr.agregar(campo, "es obligatorio")
	case utf8.RuneCountInString(nombre) > maxLargoNombre:
		verr.agregar(campo, fmt.Sprintf("no puede superar los %d caracteres", maxLargoNombre))
	}
	return nombre
}

func construirModalidades(crudo []string, verr *ErrorValidacion) []Modalidad {
	if len(crudo) == 0 {
		verr.agregar("modalidades", "se requiere al menos una")
		return nil
	}

	visto := make(map[Modalidad]bool, len(crudo))
	salida := make([]Modalidad, 0, len(crudo))

	for _, r := range crudo {
		m := Modalidad(strings.ToLower(strings.TrimSpace(r)))
		switch {
		case !m.EsValida():
			verr.agregar("modalidades", fmt.Sprintf("%q no es una modalidad válida", r))
		case visto[m]:
			verr.agregar("modalidades", fmt.Sprintf("%q está repetida", r))
		default:
			visto[m] = true
			salida = append(salida, m)
		}
	}
	return salida
}

func construirObrasSociales(crudo []string, verr *ErrorValidacion) []string {
	// puede estar vacía: un profesional que solo atiende privado es válido
	visto := make(map[string]bool, len(crudo))
	salida := make([]string, 0, len(crudo))

	for _, r := range crudo {
		v := strings.TrimSpace(r)
		if v == "" {
			continue
		}
		// "OSDE" y "osde" son la misma obra social
		clave := Normalizar(v)
		if visto[clave] {
			verr.agregar("obrasSociales", fmt.Sprintf("%q está repetida", v))
			continue
		}
		visto[clave] = true
		salida = append(salida, v)
	}
	return salida
}
