package domain

import "testing"

func TestEspecialidadEsValida(t *testing.T) {
	validas := []Especialidad{
		EspecialidadPsicologia,
		EspecialidadKinesiologia,
		EspecialidadOdontologia,
	}
	for _, e := range validas {
		if !e.EsValida() {
			t.Errorf("Especialidad(%q) debía ser válida", e)
		}
	}

	invalidas := []Especialidad{"", "cardiologia", "Psicologia", "PSICOLOGIA"}
	for _, e := range invalidas {
		if e.EsValida() {
			t.Errorf("Especialidad(%q) no debía ser válida", e)
		}
	}
}

func TestModalidadEsValida(t *testing.T) {
	validas := []Modalidad{ModalidadTelemedicina, ModalidadPresencial, ModalidadDomicilio}
	for _, m := range validas {
		if !m.EsValida() {
			t.Errorf("Modalidad(%q) debía ser válida", m)
		}
	}

	invalidas := []Modalidad{"", "online", "Presencial"}
	for _, m := range invalidas {
		if m.EsValida() {
			t.Errorf("Modalidad(%q) no debía ser válida", m)
		}
	}
}

func TestEstadoEsValido(t *testing.T) {
	if !EstadoActivo.EsValido() || !EstadoInactivo.EsValido() {
		t.Error("activo e inactivo debían ser válidos")
	}
	if Estado("suspendido").EsValido() {
		t.Error("suspendido todavía no existe: no debía ser válido")
	}
}

func TestEstadoVerificacionEsValido(t *testing.T) {
	validos := []EstadoVerificacion{VerificacionPendiente, VerificacionVerificada, VerificacionRechazada}
	for _, v := range validos {
		if !v.EsValido() {
			t.Errorf("EstadoVerificacion(%q) debía ser válido", v)
		}
	}
	if EstadoVerificacion("desconocido").EsValido() {
		t.Error("desconocido no debía ser válido")
	}
}
