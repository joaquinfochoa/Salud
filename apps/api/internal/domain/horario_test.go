package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func entradaHorarioValida() EntradaHorarioSemanal {
	return EntradaHorarioSemanal{
		DiaSemana:   "lunes",
		Desde:       "09:00",
		Hasta:       "13:00",
		DuracionMin: 50,
		Modalidad:   "telemedicina",
	}
}

func TestNuevaSemanaValida(t *testing.T) {
	profesionalID := uuid.New()
	entradas := []EntradaHorarioSemanal{
		{DiaSemana: "miercoles", Desde: "15:00", Hasta: "19:00", DuracionMin: 60, Modalidad: "presencial"},
		{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
	}

	semana, err := NuevaSemana(profesionalID, entradas)
	if err != nil {
		t.Fatalf("NuevaSemana devolvió error: %v", err)
	}
	if len(semana) != 2 {
		t.Fatalf("len(semana) = %d, se esperaba 2", len(semana))
	}

	// se devuelve ordenada, y el lunes va primero aunque haya llegado segundo
	if semana[0].DiaSemana != DiaLunes {
		t.Errorf("el primer bloque es %q, se esperaba lunes", semana[0].DiaSemana)
	}
	if semana[0].ProfesionalID != profesionalID {
		t.Error("el bloque no quedó atado a su profesional")
	}
	if semana[0].Desde.String() != "09:00" || semana[0].Hasta.String() != "13:00" {
		t.Errorf("horas = %q a %q", semana[0].Desde, semana[0].Hasta)
	}
}

func TestNuevaSemanaVacia(t *testing.T) {
	// vaciar la agenda es legítimo: un profesional puede dejar de atender
	semana, err := NuevaSemana(uuid.New(), nil)
	if err != nil {
		t.Fatalf("una semana vacía debía ser válida, devolvió: %v", err)
	}
	if len(semana) != 0 {
		t.Errorf("len(semana) = %d, se esperaba 0", len(semana))
	}
}

func TestNuevaSemanaCamposInvalidos(t *testing.T) {
	casos := []struct {
		nombre        string
		mutar         func(*EntradaHorarioSemanal)
		campoEsperado string
	}{
		{"dia desconocido", func(e *EntradaHorarioSemanal) { e.DiaSemana = "lunez" }, "horarios[0].diaSemana"},
		{"desde mal formado", func(e *EntradaHorarioSemanal) { e.Desde = "9am" }, "horarios[0].desde"},
		{"hasta mal formado", func(e *EntradaHorarioSemanal) { e.Hasta = "25:00" }, "horarios[0].hasta"},
		{"hasta antes que desde", func(e *EntradaHorarioSemanal) { e.Hasta = "08:00" }, "horarios[0].hasta"},
		{"hasta igual a desde", func(e *EntradaHorarioSemanal) { e.Hasta = "09:00" }, "horarios[0].hasta"},
		{"duracion muy corta", func(e *EntradaHorarioSemanal) { e.DuracionMin = 5 }, "horarios[0].duracionMin"},
		{"duracion muy larga", func(e *EntradaHorarioSemanal) { e.DuracionMin = 500 }, "horarios[0].duracionMin"},
		{"modalidad desconocida", func(e *EntradaHorarioSemanal) { e.Modalidad = "carta" }, "horarios[0].modalidad"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := entradaHorarioValida()
			caso.mutar(&entrada)

			_, err := NuevaSemana(uuid.New(), []EntradaHorarioSemanal{entrada})
			if err == nil {
				t.Fatal("se esperaba un error de validación")
			}

			var verr ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
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

func TestNuevaSemanaBloqueSinHuecos(t *testing.T) {
	// "lunes de 9:00 a 9:30" con sesiones de 50 minutos no genera un solo
	// turno. El profesional tiene que enterarse al cargarlo y no dos semanas
	// después, al no recibir nada.
	entrada := entradaHorarioValida()
	entrada.Hasta = "09:30"

	_, err := NuevaSemana(uuid.New(), []EntradaHorarioSemanal{entrada})

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
	encontrado := false
	for _, campo := range verr.Campos {
		if campo.Campo == "horarios[0].duracionMin" {
			encontrado = true
		}
	}
	if !encontrado {
		t.Errorf("se esperaba un error en duracionMin, se obtuvo %+v", verr.Campos)
	}
}

func TestNuevaSemanaBloqueJusto(t *testing.T) {
	// el borde del caso anterior: un bloque donde entra exactamente una sesión
	entrada := entradaHorarioValida()
	entrada.Hasta = "09:50"

	semana, err := NuevaSemana(uuid.New(), []EntradaHorarioSemanal{entrada})
	if err != nil {
		t.Fatalf("un bloque donde entra justo una sesión es válido, devolvió: %v", err)
	}
	if len(semana) != 1 {
		t.Errorf("len(semana) = %d, se esperaba 1", len(semana))
	}
}

func TestNuevaSemanaSolapamiento(t *testing.T) {
	t.Run("dos bloques del mismo dia que se pisan", func(t *testing.T) {
		entradas := []EntradaHorarioSemanal{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
			{DiaSemana: "lunes", Desde: "12:00", Hasta: "15:00", DuracionMin: 50, Modalidad: "telemedicina"},
		}

		_, err := NuevaSemana(uuid.New(), entradas)
		var verr ErrorValidacion
		if !errors.As(err, &verr) {
			t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
		}
	})

	t.Run("dos bloques pegados no se pisan", func(t *testing.T) {
		// los intervalos son semiabiertos: uno que termina 13:00 y otro que
		// empieza 13:00 son contiguos, no solapados
		entradas := []EntradaHorarioSemanal{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
			{DiaSemana: "lunes", Desde: "13:00", Hasta: "17:00", DuracionMin: 50, Modalidad: "telemedicina"},
		}

		if _, err := NuevaSemana(uuid.New(), entradas); err != nil {
			t.Errorf("dos bloques contiguos son válidos, devolvió: %v", err)
		}
	})

	t.Run("mismo horario en dias distintos no se pisa", func(t *testing.T) {
		entradas := []EntradaHorarioSemanal{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
			{DiaSemana: "martes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
		}

		if _, err := NuevaSemana(uuid.New(), entradas); err != nil {
			t.Errorf("el mismo horario en días distintos es válido, devolvió: %v", err)
		}
	})
}

func TestNuevaSemanaAcotaLaCantidadDeBloques(t *testing.T) {
	// Sin este tope, entradas suficientes en un cuerpo de 1 MB llegan al
	// verificador de solapamiento, que es O(n²) y acumula un error por cada
	// par: 101 bloques bastan para probar que el guard corta antes de eso.
	entradas := make([]EntradaHorarioSemanal, 101)
	for i := range entradas {
		entradas[i] = entradaHorarioValida()
	}

	_, err := NuevaSemana(uuid.New(), entradas)

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
	if len(verr.Campos) != 1 || verr.Campos[0].Campo != "horarios" {
		t.Errorf("se esperaba un único error en el campo horarios, se obtuvo %+v", verr.Campos)
	}
}

func TestNuevaSemanaConCienBloquesEsValida(t *testing.T) {
	entradas := make([]EntradaHorarioSemanal, 100)
	for i := range entradas {
		desde := i * 10
		entradas[i] = EntradaHorarioSemanal{
			DiaSemana:   "lunes",
			Desde:       HoraDelDia{Minutos: desde}.String(),
			Hasta:       HoraDelDia{Minutos: desde + 10}.String(),
			DuracionMin: 10,
			Modalidad:   "telemedicina",
		}
	}

	semana, err := NuevaSemana(uuid.New(), entradas)
	if err != nil {
		t.Fatalf("100 bloques que no se pisan tienen que ser válidos, devolvió: %v", err)
	}
	if len(semana) != 100 {
		t.Errorf("len(semana) = %d, se esperaba 100", len(semana))
	}
}

func TestNuevaSemanaAcumulaErroresDeVariosBloques(t *testing.T) {
	entradas := []EntradaHorarioSemanal{
		{DiaSemana: "lunez", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
		{DiaSemana: "martes", Desde: "nueve", Hasta: "13:00", DuracionMin: 50, Modalidad: "telemedicina"},
	}

	_, err := NuevaSemana(uuid.New(), entradas)

	var verr ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T", err)
	}
	// el índice del bloque tiene que estar en el nombre del campo, o el cliente
	// no sabe cuál de los dos corregir
	if len(verr.Campos) != 2 {
		t.Fatalf("se esperaban 2 campos con error, se obtuvieron %d: %+v", len(verr.Campos), verr.Campos)
	}
	if verr.Campos[0].Campo != "horarios[0].diaSemana" || verr.Campos[1].Campo != "horarios[1].desde" {
		t.Errorf("los campos no identifican el bloque: %+v", verr.Campos)
	}
}
