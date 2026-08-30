package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// 2026-08-24 es lunes. Todas las fechas de este archivo se apoyan en eso.
const (
	lunes         = "2026-08-24"
	martes        = "2026-08-25"
	lunesQueViene = "2026-08-31"
)

func bloqueLunes() HorarioSemanal {
	return HorarioSemanal{
		ProfesionalID: uuid.New(),
		DiaSemana:     DiaLunes,
		Desde:         HoraDelDia{Minutos: 9 * 60},
		Hasta:         HoraDelDia{Minutos: 13 * 60},
		DuracionMin:   50,
		Modalidad:     ModalidadTelemedicina,
	}
}

// rango arma el intervalo semiabierto que espera CalculoHuecos a partir de dos
// fechas inclusivas, igual que hace el servicio.
func rango(t *testing.T, desdeFecha, hastaFecha string) (time.Time, time.Time) {
	t.Helper()
	desde := momento(t, desdeFecha+" 00:00")
	hasta := momento(t, hastaFecha+" 00:00").AddDate(0, 0, 1)
	return desde, hasta
}

func inicios(huecos []Hueco) []string {
	salida := make([]string, 0, len(huecos))
	for _, h := range huecos {
		salida = append(salida, h.Inicio.Format("2006-01-02 15:04"))
	}
	return salida
}

func compararInicios(t *testing.T, huecos []Hueco, esperado []string) {
	t.Helper()
	obtenido := inicios(huecos)
	if len(obtenido) != len(esperado) {
		t.Fatalf("se obtuvieron %d huecos %v, se esperaban %d %v", len(obtenido), obtenido, len(esperado), esperado)
	}
	for i := range esperado {
		if obtenido[i] != esperado[i] {
			t.Errorf("hueco %d = %q, se esperaba %q", i, obtenido[i], esperado[i])
		}
	}
}

func TestGenerarUnBloque(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	c := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}

	// de 09:00 a 13:00 con sesiones de 50 minutos entran cuatro: la quinta
	// arrancaría 12:20 y terminaría 13:10, pasada del final del bloque
	compararInicios(t, c.Generar(), []string{
		lunes + " 09:00",
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarLlevaLaModalidadDelBloque(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)
	bloque := bloqueLunes()
	bloque.Modalidad = ModalidadDomicilio

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloque},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) == 0 {
		t.Fatal("no se generó ningún hueco")
	}
	for _, h := range huecos {
		if h.Modalidad != ModalidadDomicilio {
			t.Errorf("modalidad = %q, se esperaba domicilio", h.Modalidad)
		}
	}
}

func TestGenerarBloqueDondeNoEntraNinguno(t *testing.T) {
	// la validación rechaza esto al cargarlo, pero el cálculo tampoco tiene
	// que inventar un hueco que se pasa del bloque
	desde, hasta := rango(t, lunes, lunes)
	bloque := bloqueLunes()
	bloque.Hasta = HoraDelDia{Minutos: 9*60 + 30}

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloque},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("se generaron %d huecos %v, se esperaban 0", len(huecos), inicios(huecos))
	}
}

func TestGenerarDosBloquesElMismoDia(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	tarde := bloqueLunes()
	tarde.Desde = HoraDelDia{Minutos: 15 * 60}
	tarde.Hasta = HoraDelDia{Minutos: 17 * 60}
	tarde.DuracionMin = 60

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{tarde, bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	// salen ordenados aunque los bloques hayan llegado al revés
	compararInicios(t, huecos, []string{
		lunes + " 09:00",
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
		lunes + " 15:00",
		lunes + " 16:00",
	})
}

func TestGenerarIgnoraLosBloquesDeOtroDia(t *testing.T) {
	desde, hasta := rango(t, martes, martes)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("un bloque de lunes no debía producir nada un martes: %v", inicios(huecos))
	}
}

func TestGenerarRepiteElBloqueCadaSemana(t *testing.T) {
	desde, hasta := rango(t, lunes, lunesQueViene)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	// dos lunes en el rango, cuatro huecos cada uno
	if len(huecos) != 8 {
		t.Fatalf("se obtuvieron %d huecos %v, se esperaban 8", len(huecos), inicios(huecos))
	}
	if inicios(huecos)[0] != lunes+" 09:00" || inicios(huecos)[4] != lunesQueViene+" 09:00" {
		t.Errorf("el bloque no se repitió como corresponde: %v", inicios(huecos))
	}
}

func TestGenerarBloqueoQueTapaTodoElDia(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Bloqueos: []Bloqueo{{
			Desde: momento(t, lunes+" 00:00"),
			Hasta: momento(t, martes+" 00:00"),
		}},
		Desde: desde,
		Hasta: hasta,
		Ahora: momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("se generaron %d huecos %v con el día entero bloqueado", len(huecos), inicios(huecos))
	}
}

func TestGenerarBloqueoParcial(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Bloqueos: []Bloqueo{{
			Desde: momento(t, lunes+" 09:30"),
			Hasta: momento(t, lunes+" 10:30"),
		}},
		Desde: desde,
		Hasta: hasta,
		Ahora: momento(t, "2026-08-01 00:00"),
	}.Generar()

	// 09:00-09:50 y 09:50-10:40 pisan el bloqueo; 10:40 y 11:30 sobreviven
	compararInicios(t, huecos, []string{
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarBordeDelBloqueo(t *testing.T) {
	// El test que más importa de este archivo. Un bloqueo que arranca
	// exactamente cuando termina un hueco no lo pisa, y uno que termina
	// exactamente cuando el hueco arranca tampoco.
	desde, hasta := rango(t, lunes, lunes)

	t.Run("empieza cuando el hueco termina", func(t *testing.T) {
		huecos := CalculoHuecos{
			Horarios: []HorarioSemanal{bloqueLunes()},
			Bloqueos: []Bloqueo{{
				Desde: momento(t, lunes+" 09:50"),
				Hasta: momento(t, lunes+" 23:00"),
			}},
			Desde: desde,
			Hasta: hasta,
			Ahora: momento(t, "2026-08-01 00:00"),
		}.Generar()

		// solo sobrevive el de 09:00, que termina justo a las 09:50
		compararInicios(t, huecos, []string{lunes + " 09:00"})
	})

	t.Run("termina cuando el hueco empieza", func(t *testing.T) {
		huecos := CalculoHuecos{
			Horarios: []HorarioSemanal{bloqueLunes()},
			Bloqueos: []Bloqueo{{
				Desde: momento(t, lunes+" 00:00"),
				Hasta: momento(t, lunes+" 09:00"),
			}},
			Desde: desde,
			Hasta: hasta,
			Ahora: momento(t, "2026-08-01 00:00"),
		}.Generar()

		// no se pierde ninguno: el bloqueo termina justo cuando arranca el primero
		compararInicios(t, huecos, []string{
			lunes + " 09:00",
			lunes + " 09:50",
			lunes + " 10:40",
			lunes + " 11:30",
		})
	})
}

func TestGenerarAnticipacionMinima(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios:              []HorarioSemanal{bloqueLunes()},
		Desde:                 desde,
		Hasta:                 hasta,
		AnticipacionMinimaMin: 120,
		// son las 08:00 del mismo lunes: con dos horas de anticipación, lo
		// más temprano reservable son las 10:00
		Ahora: momento(t, lunes+" 08:00"),
	}.Generar()

	compararInicios(t, huecos, []string{
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarAnticipacionMinimaEnElBordeExacto(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	huecos := CalculoHuecos{
		Horarios:              []HorarioSemanal{bloqueLunes()},
		Desde:                 desde,
		Hasta:                 hasta,
		AnticipacionMinimaMin: 60,
		// a las 08:50 con una hora de anticipación, el mínimo cae exactamente
		// en 09:50: ese hueco entra
		Ahora: momento(t, lunes+" 08:50"),
	}.Generar()

	compararInicios(t, huecos, []string{
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

func TestGenerarRangoInvertido(t *testing.T) {
	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Desde:    momento(t, lunesQueViene+" 00:00"),
		Hasta:    momento(t, lunes+" 00:00"),
		Ahora:    momento(t, "2026-08-01 00:00"),
	}.Generar()

	if len(huecos) != 0 {
		t.Errorf("un rango invertido no produce nada, dio %v", inicios(huecos))
	}
}

// TestGenerarRespetaLaHoraDeRelojEnUnCambioDeHuso es la única prueba del
// repositorio que distingue construir el hueco con time.Date de sumarle una
// duración al inicio del día. El 2008-10-19 (domingo) Argentina adelantó los
// relojes: sumar aritmética de instantes desde la medianoche corre la hora en
// vez de respetar la hora de reloj. Con fechas de 2026 esto no se nota porque
// el país no tiene horario de verano desde 2009.
//
// El rango se pide más ancho que un solo día (17 al 20) para no chocar contra
// el propio salto: pedir justo el 19 de 00:00 a 00:00 del día siguiente cae
// exactamente en el instante que no existe, y el rango terminaría vacío antes
// de llegar a generar nada.
func TestGenerarRespetaLaHoraDeRelojEnUnCambioDeHuso(t *testing.T) {
	desde, hasta := rango(t, "2008-10-18", "2008-10-20")

	bloque := bloqueLunes()
	bloque.DiaSemana = DiaDomingo

	huecos := CalculoHuecos{
		Horarios: []HorarioSemanal{bloque},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2008-10-01 00:00"),
	}.Generar()

	compararInicios(t, huecos, []string{
		"2008-10-19 09:00",
		"2008-10-19 09:50",
		"2008-10-19 10:40",
		"2008-10-19 11:30",
	})
}

func TestGenerarSinHorarios(t *testing.T) {
	desde, hasta := rango(t, lunes, lunesQueViene)

	huecos := CalculoHuecos{
		Desde: desde,
		Hasta: hasta,
		Ahora: momento(t, "2026-08-01 00:00"),
	}.Generar()

	// devuelve una lista vacía, no nil: el handler la serializa como [] y el
	// cliente no tiene que chequear null
	if huecos == nil {
		t.Error("se esperaba una lista vacía, no nil")
	}
	if len(huecos) != 0 {
		t.Errorf("se esperaban 0 huecos, se obtuvieron %d", len(huecos))
	}
}

// turnoEn arma un turno tomado que arranca en la hora dada del lunes.
func turnoEn(t *testing.T, hora string, estado EstadoTurno) Turno {
	t.Helper()
	inicio := momento(t, lunes+" "+hora)
	return Turno{Inicio: inicio, Fin: inicio.Add(50 * time.Minute), Estado: estado}
}

func TestHuecosRestanTurnosTomados(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	c := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Turnos:   []Turno{turnoEn(t, "09:50", TurnoReservado)},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}

	// los mismos cuatro de TestGenerarUnBloque menos el reservado
	compararInicios(t, c.Generar(), []string{
		lunes + " 09:00",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

// Cancelar libera el hueco: es el punto de cancelar.
func TestHuecosNoRestanTurnosCancelados(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	c := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Turnos:   []Turno{turnoEn(t, "09:50", TurnoCancelado)},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}

	compararInicios(t, c.Generar(), []string{
		lunes + " 09:00",
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}

// El mismo borde que los bloqueos: un turno que termina 09:50 no tapa el hueco
// que empieza 09:50. Un <= en cualquiera de las dos comparaciones de
// SeSolapaCon haría desaparecer un hueco por turno.
func TestElBordeDelTurnoNoComeElHuecoSiguiente(t *testing.T) {
	desde, hasta := rango(t, lunes, lunes)

	c := CalculoHuecos{
		Horarios: []HorarioSemanal{bloqueLunes()},
		Turnos:   []Turno{turnoEn(t, "09:00", TurnoReservado)},
		Desde:    desde,
		Hasta:    hasta,
		Ahora:    momento(t, "2026-08-01 00:00"),
	}

	compararInicios(t, c.Generar(), []string{
		lunes + " 09:50",
		lunes + " 10:40",
		lunes + " 11:30",
	})
}
