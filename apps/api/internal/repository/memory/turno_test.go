package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// El lunes 2026-09-07 a las 10:00, con turnos de 50 minutos.
func huecoRepo(desplazamiento time.Duration) domain.Hueco {
	inicio := time.Date(2026, 9, 7, 10, 0, 0, 0, domain.ZonaHoraria).Add(desplazamiento)
	return domain.Hueco{Inicio: inicio, Fin: inicio.Add(50 * time.Minute), Modalidad: domain.ModalidadTelemedicina}
}

func turnoRepo(t *testing.T, profesionalID, pacienteID uuid.UUID, desplazamiento time.Duration) domain.Turno {
	t.Helper()
	turno, err := domain.NuevoTurno(profesionalID, pacienteID, huecoRepo(desplazamiento), "", time.Now().UTC())
	if err != nil {
		t.Fatalf("domain.NuevoTurno: %v", err)
	}
	return turno
}

// ventana cubre todo el día del hueco.
func ventana() (time.Time, time.Time) {
	dia := time.Date(2026, 9, 7, 0, 0, 0, 0, domain.ZonaHoraria)
	return dia, dia.AddDate(0, 0, 1)
}

func TestTurnoCrearYObtener(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	turno := turnoRepo(t, uuid.New(), uuid.New(), 0)

	if err := r.Crear(ctx, turno); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	obtenido, err := r.ObtenerPorID(ctx, turno.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if !obtenido.Inicio.Equal(turno.Inicio) || obtenido.PacienteID != turno.PacienteID {
		t.Error("el turno no volvió igual")
	}

	if _, err := r.ObtenerPorID(ctx, uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestTurnoHuecoTomado(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID := uuid.New()

	if err := r.Crear(ctx, turnoRepo(t, profesionalID, uuid.New(), 0)); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// otro paciente, mismo profesional, mismo intervalo. Chequear la
	// disponibilidad en una llamada y escribir en otra son dos operaciones, y
	// entre las dos entra el otro paciente. Bajo un mismo lock son una sola.
	err := r.Crear(ctx, turnoRepo(t, profesionalID, uuid.New(), 0))
	if !errors.Is(err, domain.ErrHuecoTomado) {
		t.Errorf("se esperaba ErrHuecoTomado, se obtuvo %v", err)
	}
}

func TestTurnoHuecoLiberadoPorCancelacion(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID := uuid.New()

	primero := turnoRepo(t, profesionalID, uuid.New(), 0)
	if err := r.Crear(ctx, primero); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	cancelado, err := primero.Cancelar(uuid.New(), primero.Inicio.Add(-time.Hour))
	if err != nil {
		t.Fatalf("Cancelar: %v", err)
	}
	if err := r.Actualizar(ctx, cancelado); err != nil {
		t.Fatalf("Actualizar: %v", err)
	}

	if err := r.Crear(ctx, turnoRepo(t, profesionalID, uuid.New(), 0)); err != nil {
		t.Errorf("el hueco debería estar libre después de cancelar: %v", err)
	}
}

func TestTurnoPacienteSolapado(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	pacienteID := uuid.New()

	if err := r.Crear(ctx, turnoRepo(t, uuid.New(), pacienteID, 0)); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// distinto profesional, mismo paciente, misma hora
	err := r.Crear(ctx, turnoRepo(t, uuid.New(), pacienteID, 0))
	if !errors.Is(err, domain.ErrPacienteSolapado) {
		t.Errorf("se esperaba ErrPacienteSolapado, se obtuvo %v", err)
	}
}

// Semiabierto: uno termina 10:50 y el otro empieza 10:50. Los dos entran, con
// el mismo profesional y con el mismo paciente.
func TestTurnoPegadosNoChocan(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID, pacienteID := uuid.New(), uuid.New()

	if err := r.Crear(ctx, turnoRepo(t, profesionalID, pacienteID, 0)); err != nil {
		t.Fatalf("Crear: %v", err)
	}
	if err := r.Crear(ctx, turnoRepo(t, profesionalID, pacienteID, 50*time.Minute)); err != nil {
		t.Errorf("dos turnos pegados no se pisan: %v", err)
	}
}

func TestTurnoIDDuplicado(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	turno := turnoRepo(t, uuid.New(), uuid.New(), 0)

	if err := r.Crear(ctx, turno); err != nil {
		t.Fatalf("Crear: %v", err)
	}
	if err := r.Crear(ctx, turno); !errors.Is(err, domain.ErrIDEnUso) {
		t.Errorf("se esperaba ErrIDEnUso, se obtuvo %v", err)
	}
}

func TestTurnoListados(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID, pacienteID := uuid.New(), uuid.New()
	desde, hasta := ventana()

	mio := turnoRepo(t, profesionalID, pacienteID, 0)
	ajeno := turnoRepo(t, uuid.New(), uuid.New(), 50*time.Minute)
	for _, turno := range []domain.Turno{mio, ajeno} {
		if err := r.Crear(ctx, turno); err != nil {
			t.Fatalf("Crear: %v", err)
		}
	}

	delProfesional, err := r.ListarDeProfesional(ctx, profesionalID, desde, hasta)
	if err != nil {
		t.Fatalf("ListarDeProfesional: %v", err)
	}
	if len(delProfesional) != 1 || delProfesional[0].ID != mio.ID {
		t.Errorf("ListarDeProfesional devolvió %d turnos, se esperaba solo el propio", len(delProfesional))
	}

	delPaciente, err := r.ListarDePaciente(ctx, pacienteID, desde, hasta)
	if err != nil {
		t.Fatalf("ListarDePaciente: %v", err)
	}
	if len(delPaciente) != 1 || delPaciente[0].ID != mio.ID {
		t.Errorf("ListarDePaciente devolvió %d turnos, se esperaba solo el propio", len(delPaciente))
	}
}

func TestTurnoListadosRecortanPorRango(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID := uuid.New()

	turno := turnoRepo(t, profesionalID, uuid.New(), 0)
	if err := r.Crear(ctx, turno); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	// una ventana del día siguiente no lo alcanza
	desde, hasta := ventana()
	fuera, err := r.ListarDeProfesional(ctx, profesionalID, hasta, hasta.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("ListarDeProfesional: %v", err)
	}
	if len(fuera) != 0 {
		t.Errorf("se devolvieron %d turnos fuera del rango", len(fuera))
	}

	dentro, err := r.ListarDeProfesional(ctx, profesionalID, desde, hasta)
	if err != nil {
		t.Fatalf("ListarDeProfesional: %v", err)
	}
	if len(dentro) != 1 {
		t.Errorf("se devolvieron %d turnos dentro del rango, se esperaba 1", len(dentro))
	}
}

// Un turno cancelado es parte del historial de las dos partes. Esconderlo
// sería esconderle al paciente que le cancelaron algo.
func TestTurnoListadosIncluyenCancelados(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID, pacienteID := uuid.New(), uuid.New()
	desde, hasta := ventana()

	turno := turnoRepo(t, profesionalID, pacienteID, 0)
	if err := r.Crear(ctx, turno); err != nil {
		t.Fatalf("Crear: %v", err)
	}
	cancelado, err := turno.Cancelar(pacienteID, turno.Inicio.Add(-time.Hour))
	if err != nil {
		t.Fatalf("Cancelar: %v", err)
	}
	if err := r.Actualizar(ctx, cancelado); err != nil {
		t.Fatalf("Actualizar: %v", err)
	}

	lista, err := r.ListarDePaciente(ctx, pacienteID, desde, hasta)
	if err != nil {
		t.Fatalf("ListarDePaciente: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("se devolvieron %d turnos, se esperaba 1 aunque esté cancelado", len(lista))
	}
	if lista[0].EstaActivo() {
		t.Error("el turno devuelto debería estar cancelado")
	}
}

func TestTurnoActualizarInexistente(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()

	err := r.Actualizar(ctx, turnoRepo(t, uuid.New(), uuid.New(), 0))
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El repositorio guarda y devuelve copias: sin esto, quien recibe un turno le
// puede correr la fecha de cancelación al que está guardado.
func TestTurnoDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()

	turno := turnoRepo(t, uuid.New(), uuid.New(), 0)
	cancelado, err := turno.Cancelar(uuid.New(), turno.Inicio.Add(-time.Hour))
	if err != nil {
		t.Fatalf("Cancelar: %v", err)
	}
	if err := r.Crear(ctx, cancelado); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	*cancelado.CanceladoEn = cancelado.CanceladoEn.Add(time.Hour)

	guardado, err := r.ObtenerPorID(ctx, turno.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if guardado.CanceladoEn.Equal(*cancelado.CanceladoEn) {
		t.Fatal("Crear guardó el puntero del llamador en vez de una copia")
	}

	*guardado.CanceladoEn = guardado.CanceladoEn.Add(time.Hour)
	otra, err := r.ObtenerPorID(ctx, turno.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID: %v", err)
	}
	if otra.CanceladoEn.Equal(*guardado.CanceladoEn) {
		t.Error("ObtenerPorID devolvió el puntero guardado en vez de una copia")
	}
}

// El mapa de Go itera en orden aleatorio. Sin orden total, dos llamadas
// idénticas devuelven listas distintas.
func TestTurnoOrdenTotal(t *testing.T) {
	ctx := context.Background()
	r := NuevoTurno()
	profesionalID := uuid.New()
	desde, hasta := ventana()

	for i := range 5 {
		if err := r.Crear(ctx, turnoRepo(t, profesionalID, uuid.New(), time.Duration(i)*50*time.Minute)); err != nil {
			t.Fatalf("Crear: %v", err)
		}
	}

	primera, err := r.ListarDeProfesional(ctx, profesionalID, desde, hasta)
	if err != nil {
		t.Fatalf("ListarDeProfesional: %v", err)
	}
	for range 5 {
		otra, err := r.ListarDeProfesional(ctx, profesionalID, desde, hasta)
		if err != nil {
			t.Fatalf("ListarDeProfesional: %v", err)
		}
		for i := range primera {
			if otra[i].ID != primera[i].ID {
				t.Fatalf("el orden cambió entre dos llamadas idénticas, posición %d", i)
			}
		}
	}
	// y está ordenado por inicio
	for i := 1; i < len(primera); i++ {
		if primera[i].Inicio.Before(primera[i-1].Inicio) {
			t.Errorf("el turno %d empieza antes que el %d", i, i-1)
		}
	}
}
