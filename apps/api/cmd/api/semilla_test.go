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
	svc := service.NuevoProfesional(memory.NuevoProfesional())

	if err := sembrar(ctx, service.NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion()), svc); err != nil {
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
