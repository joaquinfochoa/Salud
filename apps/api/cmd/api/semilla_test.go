package main

import (
	"context"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

func TestSembrar(t *testing.T) {
	ctx := context.Background()
	repoProf := memory.NuevoProfesional()
	svc := service.NuevoProfesional(repoProf)
	agenda := service.NuevaAgenda(repoProf, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo(), memory.NuevoTurno())
	auth := service.NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion())

	if err := sembrar(ctx, auth, svc, agenda); err != nil {
		t.Fatalf("sembrar devolvió error: %v", err)
	}

	ps, total, err := svc.Listar(ctx, repository.Filtro{Limite: 100})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if total != 4 {
		t.Fatalf("total = %d, se esperaban 4 profesionales de prueba", total)
	}

	// El seed pasa por las mismas reglas que cualquier alta, así que estas dos
	// invariantes valen también acá: sumar un homónimo o repetir una matrícula
	// falla en sembrar, no cuando alguien abra la ficha pública.
	slugs := make(map[string]bool, len(ps))
	matriculas := make(map[string]bool, len(ps))
	for _, p := range ps {
		if slugs[p.Slug] {
			t.Errorf("slug repetido en el seed: %q", p.Slug)
		}
		slugs[p.Slug] = true

		if matriculas[p.Matricula.String()] {
			t.Errorf("matrícula repetida en el seed: %q", p.Matricula)
		}
		matriculas[p.Matricula.String()] = true
	}
}

// El seed deja a los cuatro con la semana cargada. Sin esto el front arranca
// mostrando "todavía no publicó horarios" en cada tarjeta, que es un producto
// que se ve roto en cada reinicio.
func TestSembrarCargaHorarios(t *testing.T) {
	ctx := context.Background()
	repoProf := memory.NuevoProfesional()
	svc := service.NuevoProfesional(repoProf)
	agenda := service.NuevaAgenda(repoProf, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo(), memory.NuevoTurno())
	auth := service.NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion())

	if err := sembrar(ctx, auth, svc, agenda); err != nil {
		t.Fatalf("sembrar devolvió error: %v", err)
	}

	ps, _, err := svc.Listar(ctx, repository.Filtro{Limite: 100})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}

	for _, p := range ps {
		semana, err := agenda.ListarHorarios(ctx, p.ID)
		if err != nil {
			t.Fatalf("ListarHorarios de %s: %v", p.NombreCompleto(), err)
		}
		if len(semana) == 0 {
			t.Errorf("%s no tiene ningún bloque cargado", p.NombreCompleto())
		}
	}
}
