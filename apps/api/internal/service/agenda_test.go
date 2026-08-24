package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

const lunesDePrueba = "2026-08-24" // es lunes

func instante(t *testing.T, s string) time.Time {
	t.Helper()
	momento, err := time.ParseInLocation("2006-01-02 15:04", s, domain.ZonaHoraria)
	if err != nil {
		t.Fatalf("no se pudo parsear %q: %v", s, err)
	}
	return momento
}

func dia(t *testing.T, fecha string) time.Time {
	t.Helper()
	return instante(t, fecha+" 00:00")
}

// bancoDePrueba arma el stack real: repositorios en memoria, sin mocks.
type bancoDePrueba struct {
	profesionales *memory.Profesional
	agenda        *Agenda
	svcProf       *Profesional
}

func nuevoBancoDePrueba() *bancoDePrueba {
	profesionales := memory.NuevoProfesional()
	return &bancoDePrueba{
		profesionales: profesionales,
		svcProf:       NuevoProfesional(profesionales),
		agenda:        NuevaAgenda(profesionales, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo()),
	}
}

func (b *bancoDePrueba) crearProfesional(t *testing.T) domain.Profesional {
	t.Helper()
	p, err := b.svcProf.Crear(context.Background(), entradaValida())
	if err != nil {
		t.Fatalf("no se pudo crear el profesional de prueba: %v", err)
	}
	return p
}

func entradaHorarioLunes() domain.EntradaHorarioSemanal {
	return domain.EntradaHorarioSemanal{
		DiaSemana:   "lunes",
		Desde:       "09:00",
		Hasta:       "13:00",
		DuracionMin: 50,
		Modalidad:   "telemedicina",
	}
}

func TestReemplazarHorarios(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	semana, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()})
	if err != nil {
		t.Fatalf("ReemplazarHorarios devolvió error: %v", err)
	}
	if len(semana) != 1 {
		t.Fatalf("len = %d, se esperaba 1", len(semana))
	}

	guardada, err := banco.agenda.ListarHorarios(ctx, p.ID)
	if err != nil {
		t.Fatalf("ListarHorarios devolvió error: %v", err)
	}
	if len(guardada) != 1 {
		t.Error("la semana no quedó persistida")
	}
}

func TestReemplazarHorariosDeProfesionalInexistente(t *testing.T) {
	_, err := nuevoBancoDePrueba().agenda.ReemplazarHorarios(
		context.Background(), uuid.New(), []domain.EntradaHorarioSemanal{entradaHorarioLunes()})
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestReemplazarHorariosConModalidadQueElProfesionalNoOfrece(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()

	// entradaValida() declara solo telemedicina
	p := banco.crearProfesional(t)

	entrada := entradaHorarioLunes()
	entrada.Modalidad = "domicilio"

	_, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entrada})

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
	if len(verr.Campos) == 0 {
		t.Fatal("el error no nombra ningún campo")
	}
}

func TestCrearYListarBloqueos(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	b, err := banco.agenda.CrearBloqueo(ctx, p.ID, domain.EntradaBloqueo{
		Desde:  dia(t, "2099-09-10"),
		Hasta:  dia(t, "2099-09-20"),
		Motivo: "Vacaciones",
	})
	if err != nil {
		t.Fatalf("CrearBloqueo devolvió error: %v", err)
	}

	obtenidos, err := banco.agenda.ListarBloqueos(ctx, p.ID, dia(t, "2099-09-01"), dia(t, "2099-10-01"))
	if err != nil {
		t.Fatalf("ListarBloqueos devolvió error: %v", err)
	}
	if len(obtenidos) != 1 || obtenidos[0].ID != b.ID {
		t.Errorf("no se recuperó el bloqueo creado: %+v", obtenidos)
	}
}

func TestEliminarBloqueoDeOtroProfesional(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	uno := banco.crearProfesional(t)

	otraEntrada := entradaValida()
	otraEntrada.Matricula = "MN 77777"
	otro, err := banco.svcProf.Crear(ctx, otraEntrada)
	if err != nil {
		t.Fatalf("no se pudo crear el segundo profesional: %v", err)
	}

	b, err := banco.agenda.CrearBloqueo(ctx, uno.ID, domain.EntradaBloqueo{
		Desde: dia(t, "2099-09-10"),
		Hasta: dia(t, "2099-09-20"),
	})
	if err != nil {
		t.Fatalf("CrearBloqueo devolvió error: %v", err)
	}

	// borrar el bloqueo de otro desde la ruta de este es un 404, no un éxito
	if err := banco.agenda.EliminarBloqueo(ctx, otro.ID, b.ID); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}

	// y el bloqueo sigue existiendo
	obtenidos, _ := banco.agenda.ListarBloqueos(ctx, uno.ID, dia(t, "2099-09-01"), dia(t, "2099-10-01"))
	if len(obtenidos) != 1 {
		t.Error("el bloqueo se borró igual")
	}
}

func TestHuecosLibres(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	if _, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()}); err != nil {
		t.Fatalf("ReemplazarHorarios devolvió error: %v", err)
	}

	// el reloj del servicio se fija para que la anticipación mínima no borre
	// los huecos del lunes de prueba
	banco.agenda.ahora = func() time.Time { return instante(t, "2026-08-01 00:00") }

	resultado, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, lunesDePrueba), dia(t, lunesDePrueba))
	if err != nil {
		t.Fatalf("HuecosLibres devolvió error: %v", err)
	}
	if len(resultado.Huecos) != 4 {
		t.Errorf("se obtuvieron %d huecos, se esperaban 4", len(resultado.Huecos))
	}
}

func TestHuecosLibresDeProfesionalInactivo(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	if _, err := banco.agenda.ReemplazarHorarios(ctx, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()}); err != nil {
		t.Fatalf("ReemplazarHorarios devolvió error: %v", err)
	}
	if err := banco.svcProf.DarDeBaja(ctx, p.ID); err != nil {
		t.Fatalf("DarDeBaja devolvió error: %v", err)
	}

	banco.agenda.ahora = func() time.Time { return instante(t, "2026-08-01 00:00") }

	resultado, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, lunesDePrueba), dia(t, lunesDePrueba))
	// un profesional dado de baja no tiene disponibilidad, pero el recurso
	// existe: es una lista vacía, no un error
	if err != nil {
		t.Fatalf("se esperaba una lista vacía, se obtuvo error: %v", err)
	}
	if len(resultado.Huecos) != 0 {
		t.Errorf("se obtuvieron %d huecos de un profesional inactivo", len(resultado.Huecos))
	}
}

func TestHuecosLibresRecortaAlHorizonte(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()

	entrada := entradaValida()
	horizonte := 7
	entrada.HorizonteDias = &horizonte
	p, err := banco.svcProf.Crear(ctx, entrada)
	if err != nil {
		t.Fatalf("no se pudo crear el profesional: %v", err)
	}

	// el reloj se fija en el lunes de prueba: el horizonte se cuenta desde hoy
	banco.agenda.ahora = func() time.Time { return instante(t, lunesDePrueba+" 08:00") }

	// se piden 90 días a alguien que expone 7
	resultado, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, lunesDePrueba), dia(t, "2026-11-24"))
	if err != nil {
		t.Fatalf("HuecosLibres devolvió error: %v", err)
	}

	// recorta, no rechaza, y lo informa
	ultimoEsperado := dia(t, "2026-08-30") // hoy + 6 días
	if !resultado.Hasta.Equal(ultimoEsperado) {
		t.Errorf("Hasta = %v, se esperaba %v", resultado.Hasta, ultimoEsperado)
	}
	if !resultado.Desde.Equal(dia(t, lunesDePrueba)) {
		t.Errorf("Desde = %v, se esperaba el pedido", resultado.Desde)
	}
}

func TestHuecosLibresRangoInvertido(t *testing.T) {
	ctx := context.Background()
	banco := nuevoBancoDePrueba()
	p := banco.crearProfesional(t)

	_, err := banco.agenda.HuecosLibres(ctx, p.ID, dia(t, "2026-09-10"), dia(t, "2026-09-01"))

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Errorf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
}
