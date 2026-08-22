package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
)

// No hay mocks en este proyecto. El repositorio en memoria es rápido y
// determinista, así que es el doble de test: se prueba contra la
// implementación de verdad. Si un test pareciera necesitar un mock, la
// frontera está mal dibujada.
func nuevoServicioDePrueba() *Profesional {
	return NuevoProfesional(memory.NuevoProfesional())
}

func entradaValida() domain.EntradaProfesional {
	return domain.EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico.",
		PrecioConsulta: 1200000,
		Modalidades:    []string{"telemedicina"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE"},
	}
}

func TestCrear(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}
	if p.Slug != "martin-gonzalez" {
		t.Errorf("Slug = %q, se esperaba %q", p.Slug, "martin-gonzalez")
	}
	if p.Estado != domain.EstadoActivo || p.Verificacion != domain.VerificacionPendiente {
		t.Error("un profesional nuevo nace activo y sin verificar")
	}

	// tiene que quedar realmente guardado
	obtenido, err := svc.ObtenerPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID devolvió error: %v", err)
	}
	if obtenido.ID != p.ID {
		t.Error("el profesional no quedó persistido")
	}
}

func TestCrearValidacion(t *testing.T) {
	entrada := entradaValida()
	entrada.Nombre = ""

	_, err := nuevoServicioDePrueba().Crear(context.Background(), entrada)

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
}

func TestCrearMatriculaDuplicada(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	if _, err := svc.Crear(ctx, entradaValida()); err != nil {
		t.Fatalf("el primer alta falló: %v", err)
	}

	// la misma matrícula escrita distinto sigue siendo la misma matrícula
	otro := entradaValida()
	otro.Nombre = "Otro"
	otro.Apellido = "Profesional"
	otro.Matricula = "m.n. 98234"

	_, err := svc.Crear(ctx, otro)
	if !errors.Is(err, domain.ErrMatriculaEnUso) {
		t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
	}
}

func TestCrearSlugUnico(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	// tres homónimos: dos "Martín González" son perfectamente posibles y no
	// pueden ser un error para el cliente
	primero, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	entrada2 := entradaValida()
	entrada2.Matricula = "MN 11111"
	segundo, err := svc.Crear(ctx, entrada2)
	if err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	entrada3 := entradaValida()
	entrada3.Matricula = "MN 22222"
	tercero, err := svc.Crear(ctx, entrada3)
	if err != nil {
		t.Fatalf("alta 3 falló: %v", err)
	}

	if primero.Slug != "martin-gonzalez" {
		t.Errorf("slug 1 = %q", primero.Slug)
	}
	if segundo.Slug != "martin-gonzalez-2" {
		t.Errorf("slug 2 = %q, se esperaba martin-gonzalez-2", segundo.Slug)
	}
	if tercero.Slug != "martin-gonzalez-3" {
		t.Errorf("slug 3 = %q, se esperaba martin-gonzalez-3", tercero.Slug)
	}
}

func TestObtenerPorSlug(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenido, err := svc.ObtenerPorSlug(ctx, p.Slug)
	if err != nil {
		t.Fatalf("ObtenerPorSlug devolvió error: %v", err)
	}
	if obtenido.ID != p.ID {
		t.Error("ObtenerPorSlug devolvió otro profesional")
	}

	if _, err := svc.ObtenerPorSlug(ctx, "no-existe"); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestObtenerPorIDNoExiste(t *testing.T) {
	_, err := nuevoServicioDePrueba().ObtenerPorID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestListarDefaults(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	for i := range 3 {
		entrada := entradaValida()
		entrada.Matricula = "MN 3000" + string(rune('0'+i))
		if _, err := svc.Crear(ctx, entrada); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	t.Run("limite por defecto", func(t *testing.T) {
		f := repository.Filtro{}
		_, total, err := svc.Listar(ctx, f)
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if total != 3 {
			t.Errorf("total = %d, se esperaba 3", total)
		}
	})

	t.Run("limite recortado al maximo", func(t *testing.T) {
		obtenido, _, err := svc.Listar(ctx, repository.Filtro{Limite: 5000})
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if len(obtenido) > LimiteMaximo {
			t.Errorf("devolvió %d elementos, el máximo es %d", len(obtenido), LimiteMaximo)
		}
	})

	t.Run("desplazamiento negativo se normaliza", func(t *testing.T) {
		obtenido, _, err := svc.Listar(ctx, repository.Filtro{Desplazamiento: -10})
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if len(obtenido) != 3 {
			t.Errorf("len = %d, se esperaba 3", len(obtenido))
		}
	})
}
