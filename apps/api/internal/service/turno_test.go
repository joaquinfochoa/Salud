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

// El reloj de estos tests está clavado en un martes, y todos los turnos caen
// el lunes siguiente: lejos de la anticipación mínima (2 h) y dentro del
// horizonte (60 días).
var (
	ahoraFijo   = time.Date(2026, 9, 1, 9, 0, 0, 0, domain.ZonaHoraria)
	lunesTurnos = time.Date(2026, 9, 7, 0, 0, 0, 0, domain.ZonaHoraria)
)

func aLas(hora, minuto int) time.Time {
	return time.Date(2026, 9, 7, hora, minuto, 0, 0, domain.ZonaHoraria)
}

type bancoTurnos struct {
	profesionales *memory.Profesional
	turnos        *memory.Turno
	svcProf       *Profesional
	agenda        *Agenda
	svc           *Turno
}

func nuevoBancoTurnos() *bancoTurnos {
	reloj := func() time.Time { return ahoraFijo }

	profesionales := memory.NuevoProfesional()
	turnos := memory.NuevoTurno()
	agenda := NuevaAgenda(profesionales, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo(), ConReloj(reloj))

	return &bancoTurnos{
		profesionales: profesionales,
		turnos:        turnos,
		svcProf:       NuevoProfesional(profesionales),
		agenda:        agenda,
		svc:           NuevoTurno(turnos, profesionales, agenda, ConRelojTurno(reloj)),
	}
}

// conProfesional crea un profesional con horario de lunes 09:00-13:00 cargado,
// que genera cuatro huecos: 09:00, 09:50, 10:40 y 11:30.
func (b *bancoTurnos) conProfesional(t *testing.T, matricula string) (domain.Profesional, uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	usuarioID := uuid.New()
	entrada := entradaValida()
	entrada.Matricula = matricula

	p, err := b.svcProf.Crear(ctx, usuarioID, entrada)
	if err != nil {
		t.Fatalf("creando el profesional: %v", err)
	}
	if _, err := b.agenda.ReemplazarHorarios(ctx, usuarioID, p.ID, []domain.EntradaHorarioSemanal{entradaHorarioLunes()}); err != nil {
		t.Fatalf("cargando el horario: %v", err)
	}
	return p, usuarioID
}

func TestReservarUnHueco(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, _ := b.conProfesional(t, "MN 100001")
	pacienteID := uuid.New()

	turno, err := b.svc.Reservar(ctx, pacienteID, p.ID, aLas(9, 0), "dolor lumbar")
	if err != nil {
		t.Fatalf("Reservar devolvió error: %v", err)
	}

	if turno.PacienteID != pacienteID || turno.ProfesionalID != p.ID {
		t.Error("las dos partes no quedaron guardadas")
	}
	if !turno.Inicio.Equal(aLas(9, 0)) {
		t.Errorf("Inicio = %v, se esperaba 09:00", turno.Inicio)
	}
	// Fin y Modalidad salen del hueco, no del cliente.
	if !turno.Fin.Equal(aLas(9, 50)) {
		t.Errorf("Fin = %v, se esperaba 09:50 (50 minutos del bloque)", turno.Fin)
	}
	if turno.Modalidad != domain.ModalidadTelemedicina {
		t.Errorf("Modalidad = %q, se esperaba la del bloque", turno.Modalidad)
	}
	if turno.Motivo != "dolor lumbar" {
		t.Errorf("Motivo = %q", turno.Motivo)
	}
}

func TestReservarUnInicioQueNoEsHueco(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, _ := b.conProfesional(t, "MN 100002")

	casos := map[string]time.Time{
		"fuera del bloque":    aLas(3, 0),
		"desalineado":         aLas(9, 25),
		"despues del bloque":  aLas(14, 0),
		"un dia sin atencion": lunesTurnos.AddDate(0, 0, 1).Add(10 * time.Hour),
	}

	for nombre, inicio := range casos {
		t.Run(nombre, func(t *testing.T) {
			_, err := b.svc.Reservar(ctx, uuid.New(), p.ID, inicio, "")

			var verr domain.ErrorValidacion
			if !errors.As(err, &verr) {
				t.Fatalf("se esperaba ErrorValidacion, se obtuvo %v", err)
			}
			if verr.Campos[0].Campo != "inicio" {
				t.Errorf("campo = %q, se esperaba inicio", verr.Campos[0].Campo)
			}
		})
	}
}

// La anticipación mínima no está escrita en ningún lado del servicio de
// turnos: sale gratis porque HuecosLibres ya no genera ese hueco. Si alguien
// "optimiza" Reservar para no consultar los huecos y validar a mano, este test
// lo atrapa.
func TestReservarFueraDeLaAnticipacionMinima(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()

	// El reloj está en el martes 09:00. Un profesional que atiende los martes
	// con dos horas de anticipación mínima no puede recibir un turno hoy a las
	// 09:50: faltan 50 minutos.
	usuarioID := uuid.New()
	entrada := entradaValida()
	entrada.Matricula = "MN 100003"
	p, err := b.svcProf.Crear(ctx, usuarioID, entrada)
	if err != nil {
		t.Fatalf("creando el profesional: %v", err)
	}

	horarioMartes := entradaHorarioLunes()
	horarioMartes.DiaSemana = "martes"
	if _, err := b.agenda.ReemplazarHorarios(ctx, usuarioID, p.ID, []domain.EntradaHorarioSemanal{horarioMartes}); err != nil {
		t.Fatalf("cargando el horario: %v", err)
	}

	hoyALas950 := time.Date(2026, 9, 1, 9, 50, 0, 0, domain.ZonaHoraria)
	_, err = b.svc.Reservar(ctx, uuid.New(), p.ID, hoyALas950, "")

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion por la anticipación mínima, se obtuvo %v", err)
	}
}

// Mismo argumento: el estado del profesional lo aplica HuecosLibres, no este
// servicio.
func TestReservarContraProfesionalInactivo(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, usuarioID := b.conProfesional(t, "MN 100004")

	if err := b.svcProf.DarDeBaja(ctx, usuarioID, p.ID); err != nil {
		t.Fatalf("DarDeBaja: %v", err)
	}

	_, err := b.svc.Reservar(ctx, uuid.New(), p.ID, aLas(9, 0), "")
	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %v", err)
	}
}

func TestReservarContraProfesionalInexistente(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()

	_, err := b.svc.Reservar(ctx, uuid.New(), uuid.New(), aLas(9, 0), "")
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestReservarDosVecesElMismoHueco(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, _ := b.conProfesional(t, "MN 100005")

	if _, err := b.svc.Reservar(ctx, uuid.New(), p.ID, aLas(9, 0), ""); err != nil {
		t.Fatalf("primera reserva: %v", err)
	}

	_, err := b.svc.Reservar(ctx, uuid.New(), p.ID, aLas(9, 0), "")
	if !errors.Is(err, domain.ErrHuecoTomado) {
		t.Errorf("se esperaba ErrHuecoTomado, se obtuvo %v", err)
	}
}

func TestReservarSolapandoAlPaciente(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	uno, _ := b.conProfesional(t, "MN 100006")
	otro, _ := b.conProfesional(t, "MN 100007")
	pacienteID := uuid.New()

	if _, err := b.svc.Reservar(ctx, pacienteID, uno.ID, aLas(9, 0), ""); err != nil {
		t.Fatalf("primera reserva: %v", err)
	}

	// mismo paciente, misma hora, otro profesional
	_, err := b.svc.Reservar(ctx, pacienteID, otro.ID, aLas(9, 0), "")
	if !errors.Is(err, domain.ErrPacienteSolapado) {
		t.Errorf("se esperaba ErrPacienteSolapado, se obtuvo %v", err)
	}
}

func TestReservarSolapandoUnTurnoCancelado(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	uno, _ := b.conProfesional(t, "MN 100008")
	otro, _ := b.conProfesional(t, "MN 100009")
	pacienteID := uuid.New()

	primero, err := b.svc.Reservar(ctx, pacienteID, uno.ID, aLas(9, 0), "")
	if err != nil {
		t.Fatalf("primera reserva: %v", err)
	}
	if err := b.svc.Cancelar(ctx, pacienteID, primero.ID); err != nil {
		t.Fatalf("Cancelar: %v", err)
	}

	if _, err := b.svc.Reservar(ctx, pacienteID, otro.ID, aLas(9, 0), ""); err != nil {
		t.Errorf("cancelar libera al paciente igual que libera el hueco: %v", err)
	}
}

func TestCancelarLoPuedenLasDosPartes(t *testing.T) {
	ctx := context.Background()

	casos := map[string]func(p domain.Profesional, duenoID, pacienteID uuid.UUID) uuid.UUID{
		"el paciente":    func(_ domain.Profesional, _, pacienteID uuid.UUID) uuid.UUID { return pacienteID },
		"el profesional": func(_ domain.Profesional, duenoID, _ uuid.UUID) uuid.UUID { return duenoID },
	}

	for nombre, quien := range casos {
		t.Run(nombre, func(t *testing.T) {
			b := nuevoBancoTurnos()
			p, duenoID := b.conProfesional(t, "MN 100010")
			pacienteID := uuid.New()

			turno, err := b.svc.Reservar(ctx, pacienteID, p.ID, aLas(9, 0), "")
			if err != nil {
				t.Fatalf("Reservar: %v", err)
			}

			if err := b.svc.Cancelar(ctx, quien(p, duenoID, pacienteID), turno.ID); err != nil {
				t.Fatalf("Cancelar: %v", err)
			}

			guardado, err := b.turnos.ObtenerPorID(ctx, turno.ID)
			if err != nil {
				t.Fatalf("ObtenerPorID: %v", err)
			}
			if guardado.EstaActivo() {
				t.Error("el turno sigue activo después de cancelarlo")
			}
			if guardado.CanceladoPor == nil || *guardado.CanceladoPor != quien(p, duenoID, pacienteID) {
				t.Error("no quedó registrado quién canceló")
			}
		})
	}
}

func TestCancelarComoTercero(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, _ := b.conProfesional(t, "MN 100011")

	turno, err := b.svc.Reservar(ctx, uuid.New(), p.ID, aLas(9, 0), "")
	if err != nil {
		t.Fatalf("Reservar: %v", err)
	}

	if err := b.svc.Cancelar(ctx, uuid.New(), turno.ID); !errors.Is(err, domain.ErrTurnoAjeno) {
		t.Errorf("se esperaba ErrTurnoAjeno, se obtuvo %v", err)
	}
}

func TestCancelarUnTurnoInexistente(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()

	if err := b.svc.Cancelar(ctx, uuid.New(), uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestListarDePacienteSoloDevuelveLosSuyos(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, _ := b.conProfesional(t, "MN 100012")
	mio, ajeno := uuid.New(), uuid.New()

	if _, err := b.svc.Reservar(ctx, mio, p.ID, aLas(9, 0), ""); err != nil {
		t.Fatalf("Reservar: %v", err)
	}
	if _, err := b.svc.Reservar(ctx, ajeno, p.ID, aLas(9, 50), ""); err != nil {
		t.Fatalf("Reservar: %v", err)
	}

	lista, err := b.svc.ListarDePaciente(ctx, mio, nil, nil)
	if err != nil {
		t.Fatalf("ListarDePaciente: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("se devolvieron %d turnos, se esperaba solo el propio", len(lista))
	}
	if lista[0].PacienteID != mio {
		t.Error("se devolvió el turno de otro paciente")
	}
}

func TestListarDeProfesionalRequiereSerDueno(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, duenoID := b.conProfesional(t, "MN 100013")

	if _, err := b.svc.Reservar(ctx, uuid.New(), p.ID, aLas(9, 0), ""); err != nil {
		t.Fatalf("Reservar: %v", err)
	}

	lista, err := b.svc.ListarDeProfesional(ctx, duenoID, p.ID, nil, nil)
	if err != nil {
		t.Fatalf("el dueño no debería recibir error: %v", err)
	}
	if len(lista) != 1 {
		t.Errorf("se devolvieron %d turnos, se esperaba 1", len(lista))
	}

	if _, err := b.svc.ListarDeProfesional(ctx, uuid.New(), p.ID, nil, nil); !errors.Is(err, domain.ErrNoAutorizado) {
		t.Errorf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
	}
}

// Los cancelados siguen en el listado: es parte del historial de las dos
// partes, y esconderlo sería esconderle al paciente que le cancelaron algo.
func TestListarIncluyeLosCancelados(t *testing.T) {
	ctx := context.Background()
	b := nuevoBancoTurnos()
	p, _ := b.conProfesional(t, "MN 100014")
	pacienteID := uuid.New()

	turno, err := b.svc.Reservar(ctx, pacienteID, p.ID, aLas(9, 0), "")
	if err != nil {
		t.Fatalf("Reservar: %v", err)
	}
	if err := b.svc.Cancelar(ctx, pacienteID, turno.ID); err != nil {
		t.Fatalf("Cancelar: %v", err)
	}

	lista, err := b.svc.ListarDePaciente(ctx, pacienteID, nil, nil)
	if err != nil {
		t.Fatalf("ListarDePaciente: %v", err)
	}
	if len(lista) != 1 {
		t.Errorf("se devolvieron %d turnos, se esperaba 1 aunque esté cancelado", len(lista))
	}
}
