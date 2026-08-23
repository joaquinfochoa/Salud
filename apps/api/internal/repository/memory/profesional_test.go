package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

var ahoraDePrueba = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

func hacerProfesional(t *testing.T, nombre, apellido, matricula string, esp domain.Especialidad, zona string) domain.Profesional {
	t.Helper()
	p, err := domain.NuevoProfesional(domain.EntradaProfesional{
		Nombre:         nombre,
		Apellido:       apellido,
		Matricula:      matricula,
		Especialidad:   string(esp),
		Bio:            "bio",
		PrecioConsulta: 1000000,
		Modalidades:    []string{"telemedicina"},
		Zona:           zona,
		ObrasSociales:  []string{"OSDE"},
	}, ahoraDePrueba)
	if err != nil {
		t.Fatalf("no se pudo construir el profesional de prueba: %v", err)
	}
	return p
}

func TestCrearYObtenerPorID(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")

	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	obtenido, err := repo.ObtenerPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("ObtenerPorID devolvió error: %v", err)
	}
	if obtenido.ID != p.ID || obtenido.Slug != p.Slug {
		t.Errorf("el profesional recuperado no coincide: %+v", obtenido)
	}
}

func TestObtenerPorIDNoExiste(t *testing.T) {
	_, err := NuevoProfesional().ObtenerPorID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestObtenerPorSlugYPorMatricula(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	porSlug, err := repo.ObtenerPorSlug(ctx, "martin-gonzalez")
	if err != nil {
		t.Fatalf("ObtenerPorSlug devolvió error: %v", err)
	}
	if porSlug.ID != p.ID {
		t.Error("ObtenerPorSlug devolvió otro profesional")
	}

	porMatricula, err := repo.ObtenerPorMatricula(ctx, p.Matricula)
	if err != nil {
		t.Fatalf("ObtenerPorMatricula devolvió error: %v", err)
	}
	if porMatricula.ID != p.ID {
		t.Error("ObtenerPorMatricula devolvió otro profesional")
	}

	if _, err := repo.ObtenerPorSlug(ctx, "no-existe"); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// El test más importante de este paquete. Si el repositorio devuelve algo que
// comparte memoria con lo que guardó, quien lo reciba puede mutar el store
// desde afuera sin enterarse.
func TestElStoreDevuelveCopias(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	// mutar lo que devolvió el repositorio
	obtenido, _ := repo.ObtenerPorID(ctx, p.ID)
	obtenido.Nombre = "MUTADO"
	obtenido.Modalidades[0] = domain.ModalidadDomicilio
	obtenido.ObrasSociales[0] = "MUTADA"

	recargado, _ := repo.ObtenerPorID(ctx, p.ID)
	if recargado.Nombre == "MUTADO" {
		t.Error("mutar el resultado alteró el store")
	}
	if recargado.Modalidades[0] == domain.ModalidadDomicilio {
		t.Error("las modalidades comparten memoria con el store")
	}
	if recargado.ObrasSociales[0] == "MUTADA" {
		t.Error("las obras sociales comparten memoria con el store")
	}

	// y al revés: mutar lo que se pasó a Crear tampoco debe afectar
	p.Modalidades[0] = domain.ModalidadPresencial
	recargado2, _ := repo.ObtenerPorID(ctx, p.ID)
	if recargado2.Modalidades[0] == domain.ModalidadPresencial {
		t.Error("Crear guardó una referencia en vez de una copia")
	}
}

func TestActualizar(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	p.Zona = "GBA Norte"
	if err := repo.Actualizar(ctx, p); err != nil {
		t.Fatalf("Actualizar devolvió error: %v", err)
	}

	obtenido, _ := repo.ObtenerPorID(ctx, p.ID)
	if obtenido.Zona != "GBA Norte" {
		t.Errorf("Zona = %q, se esperaba %q", obtenido.Zona, "GBA Norte")
	}

	desconocido := hacerProfesional(t, "Otro", "Nombre", "MN 11111", domain.EspecialidadOdontologia, "CABA")
	if err := repo.Actualizar(ctx, desconocido); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado al actualizar algo inexistente, se obtuvo %v", err)
	}
}

// La unicidad se decide acá, bajo el lock de escritura, porque es lo que va a
// hacer la constraint UNIQUE de PostgreSQL. Cada conflicto tiene que ser
// distinguible con errors.Is o el handler no puede mapearlo a un 409.
func TestCrearRechazaDuplicados(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()
	p := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	if err := repo.Crear(ctx, p); err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	t.Run("misma matricula", func(t *testing.T) {
		otro := hacerProfesional(t, "Otro", "Nombre", "MN 98234", domain.EspecialidadOdontologia, "CABA")
		if err := repo.Crear(ctx, otro); !errors.Is(err, domain.ErrMatriculaEnUso) {
			t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
		}
	})

	t.Run("mismo slug", func(t *testing.T) {
		homonimo := hacerProfesional(t, "Martín", "González", "MN 11111", domain.EspecialidadPsicologia, "CABA")
		if err := repo.Crear(ctx, homonimo); !errors.Is(err, domain.ErrSlugEnUso) {
			t.Errorf("se esperaba ErrSlugEnUso, se obtuvo %v", err)
		}
	})

	t.Run("mismo id", func(t *testing.T) {
		if err := repo.Crear(ctx, p); !errors.Is(err, domain.ErrIDEnUso) {
			t.Errorf("se esperaba ErrIDEnUso, se obtuvo %v", err)
		}
	})
}

func TestActualizarRechazaLaMatriculaDeOtro(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	uno := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	dos := hacerProfesional(t, "Carolina", "Vega", "MN 11111", domain.EspecialidadPsicologia, "CABA")
	for _, p := range []domain.Profesional{uno, dos} {
		if err := repo.Crear(ctx, p); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	robo := uno
	robo.Matricula = dos.Matricula
	if err := repo.Actualizar(ctx, robo); !errors.Is(err, domain.ErrMatriculaEnUso) {
		t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
	}

	// y el caso que no puede romperse: editar conservando la matrícula propia
	// no es un choque consigo mismo
	uno.Zona = "GBA Norte"
	if err := repo.Actualizar(ctx, uno); err != nil {
		t.Errorf("editar conservando la matrícula propia falló: %v", err)
	}
}

func TestListarFiltros(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	psico := hacerProfesional(t, "Martín", "González", "MN 98234", domain.EspecialidadPsicologia, "CABA")
	kine := hacerProfesional(t, "Pablo", "Moreno", "MN 45321", domain.EspecialidadKinesiologia, "CABA")
	odonto := hacerProfesional(t, "Gabriela", "Ríos", "MN 67890", domain.EspecialidadOdontologia, "GBA Norte")
	// las fechas distintas fuerzan un orden determinista
	kine.CreadoEn = ahoraDePrueba.Add(time.Minute)
	odonto.CreadoEn = ahoraDePrueba.Add(2 * time.Minute)

	for _, p := range []domain.Profesional{psico, kine, odonto} {
		if err := repo.Crear(ctx, p); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	t.Run("sin filtros devuelve todo", func(t *testing.T) {
		obtenido, total, err := repo.Listar(ctx, repository.Filtro{Limite: 10})
		if err != nil {
			t.Fatalf("Listar devolvió error: %v", err)
		}
		if total != 3 || len(obtenido) != 3 {
			t.Errorf("total=%d len=%d, se esperaba 3 y 3", total, len(obtenido))
		}
	})

	t.Run("por especialidad", func(t *testing.T) {
		esp := domain.EspecialidadPsicologia
		obtenido, total, _ := repo.Listar(ctx, repository.Filtro{Especialidad: &esp, Limite: 10})
		if total != 1 || obtenido[0].ID != psico.ID {
			t.Errorf("se esperaba solo el psicólogo, se obtuvo total=%d", total)
		}
	})

	t.Run("por zona sin distinguir acentos", func(t *testing.T) {
		zona := "caba"
		_, total, _ := repo.Listar(ctx, repository.Filtro{Zona: &zona, Limite: 10})
		if total != 2 {
			t.Errorf("total = %d, se esperaban 2 en CABA", total)
		}
	})

	t.Run("busqueda sin acentos", func(t *testing.T) {
		// buscar "gonzalez" tiene que encontrar a "González"
		q := "gonzalez"
		obtenido, total, _ := repo.Listar(ctx, repository.Filtro{Busqueda: &q, Limite: 10})
		if total != 1 || obtenido[0].ID != psico.ID {
			t.Errorf("la búsqueda sin acentos no encontró a González: total=%d", total)
		}
	})

	t.Run("busqueda por nombre parcial", func(t *testing.T) {
		q := "pab"
		_, total, _ := repo.Listar(ctx, repository.Filtro{Busqueda: &q, Limite: 10})
		if total != 1 {
			t.Errorf("total = %d, se esperaba 1", total)
		}
	})

	t.Run("por estado", func(t *testing.T) {
		baja := psico.DarDeBaja(ahoraDePrueba.Add(time.Hour))
		if err := repo.Actualizar(ctx, baja); err != nil {
			t.Fatalf("Actualizar devolvió error: %v", err)
		}

		inactivo := domain.EstadoInactivo
		_, total, _ := repo.Listar(ctx, repository.Filtro{Estado: &inactivo, Limite: 10})
		if total != 1 {
			t.Errorf("total = %d, se esperaba 1 inactivo", total)
		}

		activo := domain.EstadoActivo
		_, total, _ = repo.Listar(ctx, repository.Filtro{Estado: &activo, Limite: 10})
		if total != 2 {
			t.Errorf("total = %d, se esperaban 2 activos", total)
		}
	})
}

func TestListarPaginacionEsEstable(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	for i := range 10 {
		p := hacerProfesional(t, fmt.Sprintf("Nombre%d", i), "Apellido",
			fmt.Sprintf("MN %d", 10000+i), domain.EspecialidadPsicologia, "CABA")
		p.CreadoEn = ahoraDePrueba.Add(time.Duration(i) * time.Minute)
		if err := repo.Crear(ctx, p); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	pagina1, total, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 0})
	if total != 10 || len(pagina1) != 3 {
		t.Fatalf("total=%d len=%d, se esperaba 10 y 3", total, len(pagina1))
	}

	pagina2, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 3})
	if len(pagina2) != 3 {
		t.Fatalf("len(pagina2) = %d, se esperaba 3", len(pagina2))
	}

	// el mapa de Go itera en orden aleatorio: sin ordenar, dos llamadas
	// idénticas devolverían páginas distintas y la paginación no serviría
	for range 5 {
		otraVez, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 0})
		for i := range pagina1 {
			if otraVez[i].ID != pagina1[i].ID {
				t.Fatal("dos llamadas idénticas devolvieron órdenes distintos")
			}
		}
	}

	// las páginas no se solapan
	for _, a := range pagina1 {
		for _, b := range pagina2 {
			if a.ID == b.ID {
				t.Error("la página 1 y la 2 comparten un elemento")
			}
		}
	}

	ultima, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 9})
	if len(ultima) != 1 {
		t.Errorf("la última página tenía %d elementos, se esperaba 1", len(ultima))
	}

	vacio, _, _ := repo.Listar(ctx, repository.Filtro{Limite: 3, Desplazamiento: 50})
	if len(vacio) != 0 {
		t.Errorf("un desplazamiento más allá del total debía devolver vacío, devolvió %d", len(vacio))
	}
}

func TestListarNormalizaPaginacionInvalida(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	for i := range 3 {
		p := hacerProfesional(t, fmt.Sprintf("Nombre%d", i), "Apellido",
			fmt.Sprintf("MN %d", 30000+i), domain.EspecialidadPsicologia, "CABA")
		p.CreadoEn = ahoraDePrueba.Add(time.Duration(i) * time.Minute)
		if err := repo.Crear(ctx, p); err != nil {
			t.Fatalf("Crear devolvió error: %v", err)
		}
	}

	// un Desplazamiento negativo no es válido en SQL, pero acá además hace
	// que Go panique al cortar el slice con un índice bajo negativo
	obtenido, total, err := repo.Listar(ctx, repository.Filtro{Limite: 2, Desplazamiento: -1})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, se esperaba 3", total)
	}
	if len(obtenido) != 2 {
		t.Errorf("len(obtenido) = %d, se esperaba 2 (primera página)", len(obtenido))
	}

	// un Limite negativo se trata como "sin límite", igual que cero
	obtenido, total, err = repo.Listar(ctx, repository.Filtro{Limite: -1})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if total != 3 || len(obtenido) != 3 {
		t.Errorf("total=%d len=%d, se esperaba 3 y 3 con Limite negativo", total, len(obtenido))
	}
}

func TestAccesoConcurrente(t *testing.T) {
	ctx := context.Background()
	repo := NuevoProfesional()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := hacerProfesional(t, fmt.Sprintf("N%d", i), "A",
				fmt.Sprintf("MN %d", 20000+i), domain.EspecialidadPsicologia, "CABA")
			_ = repo.Crear(ctx, p)
			_, _, _ = repo.Listar(ctx, repository.Filtro{Limite: 10})
			_, _ = repo.ObtenerPorID(ctx, p.ID)
		}(i)
	}
	wg.Wait()

	_, total, _ := repo.Listar(ctx, repository.Filtro{Limite: 100})
	if total != 50 {
		t.Errorf("total = %d, se esperaban 50", total)
	}
}
