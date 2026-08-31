package domain

import (
	"testing"
	"time"
)

func TestEspecialidadEsValida(t *testing.T) {
	for _, e := range Especialidades {
		if !e.EsValida() {
			t.Errorf("Especialidad(%q) debía ser válida", e)
		}
	}

	// La lista dejó de ser tres cuando la plataforma pasó a ser multi-área.
	// Cardiología es válida ahora, y este test lo fija: si alguien la saca de
	// Especialidades, un profesional deja de poder registrarse en silencio.
	for _, e := range []Especialidad{"cardiologia", "nutricion", "pediatria"} {
		if !e.EsValida() {
			t.Errorf("Especialidad(%q) debía ser válida", e)
		}
	}

	// Lo que sigue sin valer: la mayúscula, el acento y lo que no está.
	invalidas := []Especialidad{"", "Psicologia", "PSICOLOGIA", "psicología", "astrologia"}
	for _, e := range invalidas {
		if e.EsValida() {
			t.Errorf("Especialidad(%q) no debía ser válida", e)
		}
	}
}

// Sin repetidas y sin acentos ni ñ, que es la regla de todo identificador del
// proyecto. Una repetida duplicaría la opción en el select del alta.
func TestEspecialidadesEsUnaListaSana(t *testing.T) {
	vistas := make(map[Especialidad]bool, len(Especialidades))
	for _, e := range Especialidades {
		if vistas[e] {
			t.Errorf("Especialidad(%q) está repetida", e)
		}
		vistas[e] = true

		for _, r := range e {
			if r > 127 {
				t.Errorf("Especialidad(%q) tiene un carácter no ASCII: %q", e, r)
			}
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

func TestDiaSemanaEsValido(t *testing.T) {
	validos := []DiaSemana{DiaLunes, DiaMartes, DiaMiercoles, DiaJueves, DiaViernes, DiaSabado, DiaDomingo}
	for _, d := range validos {
		if !d.EsValido() {
			t.Errorf("DiaSemana(%q) debía ser válido", d)
		}
	}

	invalidos := []DiaSemana{"", "Lunes", "LUNES", "miércoles", "monday"}
	for _, d := range invalidos {
		if d.EsValido() {
			t.Errorf("DiaSemana(%q) no debía ser válido", d)
		}
	}
}

func TestDiaSemanaIdaYVuelta(t *testing.T) {
	// la conversión contra time.Weekday tiene que cerrar en los dos sentidos,
	// o el cálculo de huecos va a mirar el día equivocado
	casos := map[DiaSemana]time.Weekday{
		DiaDomingo:   time.Sunday,
		DiaLunes:     time.Monday,
		DiaMartes:    time.Tuesday,
		DiaMiercoles: time.Wednesday,
		DiaJueves:    time.Thursday,
		DiaViernes:   time.Friday,
		DiaSabado:    time.Saturday,
	}

	for dia, weekday := range casos {
		if obtenido := dia.AWeekday(); obtenido != weekday {
			t.Errorf("%q.AWeekday() = %v, se esperaba %v", dia, obtenido, weekday)
		}
		if obtenido := DiaSemanaDe(weekday); obtenido != dia {
			t.Errorf("DiaSemanaDe(%v) = %q, se esperaba %q", weekday, obtenido, dia)
		}
	}
}

func TestDiaSemanaOrdenArrancaEnLunes(t *testing.T) {
	// time.Weekday pone el domingo en cero, que no es cómo se lee una agenda
	// en Argentina
	esperado := []DiaSemana{DiaLunes, DiaMartes, DiaMiercoles, DiaJueves, DiaViernes, DiaSabado, DiaDomingo}
	for i, dia := range esperado {
		if dia.Orden() != i {
			t.Errorf("%q.Orden() = %d, se esperaba %d", dia, dia.Orden(), i)
		}
	}
}
