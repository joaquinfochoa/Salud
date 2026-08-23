package memory

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func semanaDePrueba(profesionalID uuid.UUID) []domain.HorarioSemanal {
	return []domain.HorarioSemanal{
		{
			ProfesionalID: profesionalID,
			DiaSemana:     domain.DiaLunes,
			Desde:         domain.HoraDelDia{Minutos: 9 * 60},
			Hasta:         domain.HoraDelDia{Minutos: 13 * 60},
			DuracionMin:   50,
			Modalidad:     domain.ModalidadTelemedicina,
		},
	}
}

func TestReemplazarYListarHorarios(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("ReemplazarDeProfesional devolvió error: %v", err)
	}

	obtenida, err := repo.ListarDeProfesional(ctx, profesionalID)
	if err != nil {
		t.Fatalf("ListarDeProfesional devolvió error: %v", err)
	}
	if len(obtenida) != 1 || obtenida[0].DiaSemana != domain.DiaLunes {
		t.Errorf("la semana recuperada no coincide: %+v", obtenida)
	}
}

func TestReemplazarPisaLaSemanaAnterior(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("primer reemplazo falló: %v", err)
	}

	nueva := semanaDePrueba(profesionalID)
	nueva[0].DiaSemana = domain.DiaMartes

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, nueva); err != nil {
		t.Fatalf("segundo reemplazo falló: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, profesionalID)
	if len(obtenida) != 1 {
		t.Fatalf("len = %d, se esperaba 1: reemplazar no debe acumular", len(obtenida))
	}
	if obtenida[0].DiaSemana != domain.DiaMartes {
		t.Errorf("día = %q, se esperaba martes", obtenida[0].DiaSemana)
	}
}

func TestReemplazarConListaVaciaDejaSinHorarios(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("ReemplazarDeProfesional devolvió error: %v", err)
	}
	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, nil); err != nil {
		t.Fatalf("vaciar la semana devolvió error: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, profesionalID)
	if len(obtenida) != 0 {
		t.Errorf("len = %d, se esperaba 0", len(obtenida))
	}
}

func TestListarHorariosDeProfesionalSinCargar(t *testing.T) {
	obtenida, err := NuevoHorarioSemanal().ListarDeProfesional(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("devolvió error: %v", err)
	}
	// lista vacía, no nil: el handler la serializa como [] y el cliente no
	// tiene que chequear null
	if obtenida == nil {
		t.Error("se esperaba una lista vacía, no nil")
	}
}

func TestHorariosNoSeMezclanEntreProfesionales(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	uno, otro := uuid.New(), uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, uno, semanaDePrueba(uno)); err != nil {
		t.Fatalf("devolvió error: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, otro)
	if len(obtenida) != 0 {
		t.Errorf("el otro profesional no tenía horarios y devolvió %d", len(obtenida))
	}
}

func TestElStoreDeHorariosDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	repo := NuevoHorarioSemanal()
	profesionalID := uuid.New()

	if err := repo.ReemplazarDeProfesional(ctx, profesionalID, semanaDePrueba(profesionalID)); err != nil {
		t.Fatalf("devolvió error: %v", err)
	}

	obtenida, _ := repo.ListarDeProfesional(ctx, profesionalID)
	obtenida[0].DuracionMin = 999

	fresca, _ := repo.ListarDeProfesional(ctx, profesionalID)
	if fresca[0].DuracionMin == 999 {
		t.Error("mutar el resultado alteró el store")
	}
}
