package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var ahoraDePrueba = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

// entradaValida devuelve una entrada correcta. Cada test la copia y rompe un
// solo campo, así el que falla es siempre el campo bajo prueba.
func entradaValida() EntradaProfesional {
	precio := int64(1200000)
	return EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico con orientación cognitivo-conductual.",
		PrecioConsulta: &precio,
		Modalidades:    []string{"telemedicina", "presencial"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE", "Swiss Medical"},
	}
}

func TestNuevoProfesionalValido(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("el ID debía generarse")
	}
	if p.Slug != "martin-gonzalez" {
		t.Errorf("Slug = %q, se esperaba %q", p.Slug, "martin-gonzalez")
	}
	if p.Matricula.String() != "MN 98234" {
		t.Errorf("Matricula = %q, se esperaba %q", p.Matricula, "MN 98234")
	}
	if p.PrecioConsulta != Dinero(1200000) {
		t.Errorf("PrecioConsulta = %d, se esperaba 1200000", p.PrecioConsulta)
	}
	if p.Estado != EstadoActivo {
		t.Errorf("Estado = %q, se esperaba activo", p.Estado)
	}
	// nadie nace verificado: la verificación es un acto contra REFEPS
	if p.Verificacion != VerificacionPendiente {
		t.Errorf("Verificacion = %q, se esperaba pendiente", p.Verificacion)
	}
	if !p.CreadoEn.Equal(ahoraDePrueba) || !p.ActualizadoEn.Equal(ahoraDePrueba) {
		t.Error("las marcas de tiempo debían ser el ahora recibido")
	}
	if p.DadoDeBajaEn != nil {
		t.Error("DadoDeBajaEn debía ser nil")
	}
}

func TestNuevoProfesionalCamposInvalidos(t *testing.T) {
	casos := []struct {
		nombre        string
		mutar         func(*EntradaProfesional)
		campoEsperado string
	}{
		{"nombre vacio", func(entrada *EntradaProfesional) { entrada.Nombre = "   " }, "nombre"},
		{"nombre muy largo", func(entrada *EntradaProfesional) { entrada.Nombre = strings.Repeat("a", 101) }, "nombre"},
		{"apellido vacio", func(entrada *EntradaProfesional) { entrada.Apellido = "" }, "apellido"},
		{"matricula invalida", func(entrada *EntradaProfesional) { entrada.Matricula = "XX 123" }, "matricula"},
		{"especialidad desconocida", func(entrada *EntradaProfesional) { entrada.Especialidad = "cardiologia" }, "especialidad"},
		{"bio muy larga", func(entrada *EntradaProfesional) { entrada.Bio = strings.Repeat("a", 2001) }, "bio"},
		{"precio negativo", func(entrada *EntradaProfesional) { v := int64(-1); entrada.PrecioConsulta = &v }, "precioConsultaCentavos"},
		{"precio ausente", func(entrada *EntradaProfesional) { entrada.PrecioConsulta = nil }, "precioConsultaCentavos"},
		{"sin modalidades", func(entrada *EntradaProfesional) { entrada.Modalidades = nil }, "modalidades"},
		{"modalidad desconocida", func(entrada *EntradaProfesional) { entrada.Modalidades = []string{"online"} }, "modalidades"},
		{"modalidad repetida", func(entrada *EntradaProfesional) { entrada.Modalidades = []string{"presencial", "presencial"} }, "modalidades"},
		{"zona vacia", func(entrada *EntradaProfesional) { entrada.Zona = "" }, "zona"},
		{"obra social repetida", func(entrada *EntradaProfesional) { entrada.ObrasSociales = []string{"OSDE", "osde"} }, "obrasSociales"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := entradaValida()
			caso.mutar(&entrada)

			_, err := NuevoProfesional(entrada, ahoraDePrueba)
			if err == nil {
				t.Fatal("se esperaba un error de validación")
			}

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
			}

			encontrado := false
			for _, f := range verr.Campos {
				if f.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestNuevoProfesionalAcumulaErrores(t *testing.T) {
	entrada := entradaValida()
	entrada.Nombre = ""
	entrada.Matricula = "roto"
	entrada.Zona = ""

	_, err := NuevoProfesional(entrada, ahoraDePrueba)

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
	}
	// el punto de acumular: el cliente corrige los tres de una
	if len(verr.Campos) != 3 {
		t.Errorf("se esperaban 3 campos con error, se obtuvieron %d: %+v", len(verr.Campos), verr.Campos)
	}
}

func TestNuevoProfesionalNormalizaEntrada(t *testing.T) {
	entrada := entradaValida()
	entrada.Nombre = "  Martín  "
	entrada.Especialidad = "  PSICOLOGIA  "
	entrada.Modalidades = []string{" Telemedicina "}

	p, err := NuevoProfesional(entrada, ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Nombre != "Martín" {
		t.Errorf("Nombre = %q, se esperaba sin espacios", p.Nombre)
	}
	if p.Especialidad != EspecialidadPsicologia {
		t.Errorf("Especialidad = %q, se esperaba psicologia", p.Especialidad)
	}
	if len(p.Modalidades) != 1 || p.Modalidades[0] != ModalidadTelemedicina {
		t.Errorf("Modalidades = %v, se esperaba [telemedicina]", p.Modalidades)
	}
}

func TestAplicarCambiosResetaVerificacion(t *testing.T) {
	base, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	base.Verificacion = VerificacionVerificada

	masTarde := ahoraDePrueba.Add(time.Hour)

	t.Run("cambiar la matricula resetea", func(t *testing.T) {
		entrada := entradaValida()
		entrada.Matricula = "MN 11111"

		actualizado, err := base.AplicarCambios(entrada, masTarde)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if actualizado.Verificacion != VerificacionPendiente {
			t.Error("cambiar la matrícula tenía que volver la verificación a pendiente")
		}
	})

	t.Run("cambiar la especialidad resetea", func(t *testing.T) {
		entrada := entradaValida()
		entrada.Especialidad = "odontologia"

		actualizado, err := base.AplicarCambios(entrada, masTarde)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if actualizado.Verificacion != VerificacionPendiente {
			t.Error("cambiar la especialidad tenía que volver la verificación a pendiente")
		}
	})

	t.Run("cambiar la bio no resetea", func(t *testing.T) {
		entrada := entradaValida()
		entrada.Bio = "Otra bio."

		actualizado, err := base.AplicarCambios(entrada, masTarde)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if actualizado.Verificacion != VerificacionVerificada {
			t.Error("editar la bio no tenía por qué tocar la verificación")
		}
	})
}

func TestAplicarCambiosPreservaCamposNoEditables(t *testing.T) {
	base, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	entrada := entradaValida()
	entrada.Nombre = "Otro"
	entrada.Apellido = "Nombre"

	masTarde := ahoraDePrueba.Add(time.Hour)
	actualizado, err := base.AplicarCambios(entrada, masTarde)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if actualizado.ID != base.ID {
		t.Error("el ID no es editable")
	}
	// el slug es una URL pública: regenerarlo al cambiar el nombre rompe
	// enlaces y posicionamiento
	if actualizado.Slug != base.Slug {
		t.Errorf("el slug no debía cambiar: %q → %q", base.Slug, actualizado.Slug)
	}
	if !actualizado.CreadoEn.Equal(base.CreadoEn) {
		t.Error("CreadoEn no es editable")
	}
	if !actualizado.ActualizadoEn.Equal(masTarde) {
		t.Error("ActualizadoEn debía avanzar")
	}
}

func TestDarDeBajaReactivar(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	masTarde := ahoraDePrueba.Add(time.Hour)
	baja := p.DarDeBaja(masTarde)

	if baja.Estado != EstadoInactivo {
		t.Errorf("Estado = %q, se esperaba inactivo", baja.Estado)
	}
	if baja.DadoDeBajaEn == nil || !baja.DadoDeBajaEn.Equal(masTarde) {
		t.Error("DadoDeBajaEn debía sellarse con el momento de la baja")
	}
	// valor receiver: el original no se toca
	if p.Estado != EstadoActivo {
		t.Error("DarDeBaja no debía mutar el receptor")
	}

	// idempotente: dar de baja algo ya dado de baja no es un error ni
	// corre la fecha original
	muchoMasTarde := masTarde.Add(time.Hour)
	otraVez := baja.DarDeBaja(muchoMasTarde)
	if !otraVez.DadoDeBajaEn.Equal(masTarde) {
		t.Error("una segunda baja no debía correr la fecha de la primera")
	}

	reactivado := baja.Reactivar(muchoMasTarde)
	if reactivado.Estado != EstadoActivo {
		t.Errorf("Estado = %q, se esperaba activo", reactivado.Estado)
	}
	if reactivado.DadoDeBajaEn != nil {
		t.Error("reactivar debía limpiar DadoDeBajaEn")
	}
}

func TestClonarEsCopiaProfunda(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	c := p.Clonar()
	c.Modalidades[0] = ModalidadDomicilio
	c.ObrasSociales[0] = "MUTADA"

	if p.Modalidades[0] == ModalidadDomicilio {
		t.Error("mutar el clon alteró las modalidades del original")
	}
	if p.ObrasSociales[0] == "MUTADA" {
		t.Error("mutar el clon alteró las obras sociales del original")
	}

	baja := p.DarDeBaja(ahoraDePrueba)
	clonBaja := baja.Clonar()
	*clonBaja.DadoDeBajaEn = ahoraDePrueba.Add(time.Hour)
	if baja.DadoDeBajaEn.Equal(ahoraDePrueba.Add(time.Hour)) {
		t.Error("mutar el clon alteró el DadoDeBajaEn del original")
	}
}

func TestNombreCompleto(t *testing.T) {
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if obtenido := p.NombreCompleto(); obtenido != "Martín González" {
		t.Errorf("NombreCompleto() = %q, se esperaba %q", obtenido, "Martín González")
	}
}

func TestNuevoProfesionalConfiguracionDeAgendaPorDefecto(t *testing.T) {
	// no mandar los campos es lo normal: el profesional no debería tener que
	// decidir esto al registrarse
	p, err := NuevoProfesional(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.AnticipacionMinimaMin != anticipacionMinimaPorDefecto {
		t.Errorf("AnticipacionMinimaMin = %d, se esperaba %d", p.AnticipacionMinimaMin, anticipacionMinimaPorDefecto)
	}
	if p.HorizonteDias != horizonteDiasPorDefecto {
		t.Errorf("HorizonteDias = %d, se esperaba %d", p.HorizonteDias, horizonteDiasPorDefecto)
	}
}

func TestNuevoProfesionalConfiguracionDeAgendaExplicita(t *testing.T) {
	entrada := entradaValida()
	anticipacion := 30
	horizonte := 90
	entrada.AnticipacionMinimaMin = &anticipacion
	entrada.HorizonteDias = &horizonte

	p, err := NuevoProfesional(entrada, ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.AnticipacionMinimaMin != 30 {
		t.Errorf("AnticipacionMinimaMin = %d, se esperaba 30", p.AnticipacionMinimaMin)
	}
	if p.HorizonteDias != 90 {
		t.Errorf("HorizonteDias = %d, se esperaba 90", p.HorizonteDias)
	}
}

func TestNuevoProfesionalConfiguracionDeAgendaInvalida(t *testing.T) {
	casos := []struct {
		nombre        string
		mutar         func(*EntradaProfesional)
		campoEsperado string
	}{
		{"anticipacion negativa", func(e *EntradaProfesional) { v := -1; e.AnticipacionMinimaMin = &v }, "anticipacionMinimaMin"},
		{"anticipacion mayor a una semana", func(e *EntradaProfesional) { v := 10081; e.AnticipacionMinimaMin = &v }, "anticipacionMinimaMin"},
		{"horizonte en cero", func(e *EntradaProfesional) { v := 0; e.HorizonteDias = &v }, "horizonteDias"},
		{"horizonte sobre el tope", func(e *EntradaProfesional) { v := 181; e.HorizonteDias = &v }, "horizonteDias"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := entradaValida()
			caso.mutar(&entrada)

			_, err := NuevoProfesional(entrada, ahoraDePrueba)

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
			}
			encontrado := false
			for _, campo := range verr.Campos {
				if campo.Campo == caso.campoEsperado {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("se esperaba un error en %q, se obtuvo %+v", caso.campoEsperado, verr.Campos)
			}
		})
	}
}

func TestAplicarCambiosVuelveAlDefaultSiNoMandanLaConfiguracion(t *testing.T) {
	// PUT es reemplazo total: omitir un campo lo devuelve a su valor por
	// defecto, igual que pasa con el resto
	entrada := entradaValida()
	anticipacion := 30
	entrada.AnticipacionMinimaMin = &anticipacion

	base, err := NuevoProfesional(entrada, ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if base.AnticipacionMinimaMin != 30 {
		t.Fatalf("no se pudo preparar el estado: %d", base.AnticipacionMinimaMin)
	}

	actualizado, err := base.AplicarCambios(entradaValida(), ahoraDePrueba)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if actualizado.AnticipacionMinimaMin != anticipacionMinimaPorDefecto {
		t.Errorf("AnticipacionMinimaMin = %d, se esperaba el default %d",
			actualizado.AnticipacionMinimaMin, anticipacionMinimaPorDefecto)
	}
}
