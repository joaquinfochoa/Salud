package memory

import (
	"context"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

func TestSembrar(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	if err := Sembrar(ctx, repo); err != nil {
		t.Fatalf("Sembrar devolvió error: %v", err)
	}

	_, total, err := repo.Listar(ctx, repository.Filtro{Limite: 100})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if total != 4 {
		t.Errorf("total = %d, se esperaban 4 profesionales de prueba", total)
	}
}

func TestSembrarGeneraSlugsUnicos(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	if err := Sembrar(ctx, repo); err != nil {
		t.Fatalf("Sembrar devolvió error: %v", err)
	}

	ps, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 100})

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
