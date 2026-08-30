package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	precio := int64(1200000)
	return domain.EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico.",
		PrecioConsulta: &precio,
		Modalidades:    []string{"telemedicina"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE"},
	}
}

func TestCrear(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, uuid.New(), entradaValida())
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

	_, err := nuevoServicioDePrueba().Crear(context.Background(), uuid.New(), entrada)

	var verr domain.ErrorValidacion
	if !errors.As(err, &verr) {
		t.Fatalf("se esperaba ErrorValidacion, se obtuvo %T: %v", err, err)
	}
}

func TestCrearMatriculaDuplicada(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	if _, err := svc.Crear(ctx, uuid.New(), entradaValida()); err != nil {
		t.Fatalf("el primer alta falló: %v", err)
	}

	// la misma matrícula escrita distinto sigue siendo la misma matrícula
	otro := entradaValida()
	otro.Nombre = "Otro"
	otro.Apellido = "Profesional"
	otro.Matricula = "m.n. 98234"

	_, err := svc.Crear(ctx, uuid.New(), otro)
	if !errors.Is(err, domain.ErrMatriculaEnUso) {
		t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
	}
}

func TestCrearSlugUnico(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	// tres homónimos: dos "Martín González" son perfectamente posibles y no
	// pueden ser un error para el cliente
	primero, err := svc.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	entrada2 := entradaValida()
	entrada2.Matricula = "MN 11111"
	segundo, err := svc.Crear(ctx, uuid.New(), entrada2)
	if err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	entrada3 := entradaValida()
	entrada3.Matricula = "MN 22222"
	tercero, err := svc.Crear(ctx, uuid.New(), entrada3)
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

// enLargada lanza n goroutines y las hace arrancar todas juntas. Sin la
// barrera, la primera termina antes de que la última empiece y el test no
// llega a solapar nada.
func enLargada(n int, cuerpo func(i int)) {
	var listas sync.WaitGroup
	var terminadas sync.WaitGroup
	listas.Add(n)
	terminadas.Add(n)

	largada := make(chan struct{})
	for i := range n {
		go func() {
			defer terminadas.Done()
			listas.Done()
			<-largada
			cuerpo(i)
		}()
	}

	listas.Wait()
	close(largada)
	terminadas.Wait()
}

// TestCrearConcurrenteConMatriculaCompartida prueba que la unicidad de la
// matrícula no puede depender de un chequeo previo al alta: verificar y
// escribir bajo dos locks distintos no es una invariante, y entre los dos
// entra otro request con la misma matrícula.
//
// El detector de carreras no ve nada acá: no hay carrera de datos, todos los
// accesos están bien sincronizados. Es una carrera de lógica, y solo la
// atrapa contar los resultados.
func TestCrearConcurrenteConMatriculaCompartida(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	const n = 24
	errores := make([]error, n)

	enLargada(n, func(i int) {
		entrada := entradaValida() // todos con la misma matrícula
		// nombres distintos: así el único conflicto posible es la matrícula
		entrada.Nombre = fmt.Sprintf("Nombre%d", i)
		_, errores[i] = svc.Crear(ctx, uuid.New(), entrada)
	})

	exitos := 0
	for _, err := range errores {
		switch {
		case err == nil:
			exitos++
		case errors.Is(err, domain.ErrMatriculaEnUso):
		default:
			t.Errorf("se esperaba nil o ErrMatriculaEnUso, se obtuvo %v", err)
		}
	}
	if exitos != 1 {
		t.Errorf("altas exitosas = %d, se esperaba exactamente 1: la matrícula es única", exitos)
	}

	_, total, err := svc.Listar(ctx, repository.Filtro{Limite: 100})
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if total != 1 {
		t.Errorf("quedaron %d profesionales guardados con la misma matrícula, se esperaba 1", total)
	}
}

// TestCrearConcurrenteConMismoNombre prueba el otro lado de la moneda: un
// choque de slug nunca es un error para el cliente, así que las n altas tienen
// que entrar todas, y con slugs distintos. Si dos comparten slug, la ficha
// pública de esa URL cambia de persona entre request y request.
func TestCrearConcurrenteConMismoNombre(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	const n = 24
	slugs := make([]string, n)
	errores := make([]error, n)

	enLargada(n, func(i int) {
		entrada := entradaValida() // todos "Martín González"
		// matrículas distintas: así el único conflicto posible es el slug
		entrada.Matricula = fmt.Sprintf("MN %d", 40000+i)
		p, err := svc.Crear(ctx, uuid.New(), entrada)
		slugs[i], errores[i] = p.Slug, err
	})

	vistos := make(map[string]bool, n)
	for i, err := range errores {
		if err != nil {
			t.Errorf("alta %d falló: %v — dos homónimos son válidos", i, err)
			continue
		}
		if vistos[slugs[i]] {
			t.Errorf("slug repetido: %q", slugs[i])
		}
		vistos[slugs[i]] = true
	}

	// la secuencia visible desde afuera no cambia: base, base-2, base-3...
	esperados := map[string]bool{"martin-gonzalez": true}
	for i := 2; i <= n; i++ {
		esperados[fmt.Sprintf("martin-gonzalez-%d", i)] = true
	}
	for slug := range vistos {
		if !esperados[slug] {
			t.Errorf("slug fuera de la secuencia esperada: %q", slug)
		}
	}
	if len(vistos) != n {
		t.Errorf("slugs distintos = %d, se esperaban %d", len(vistos), n)
	}
}

func TestObtenerPorSlug(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, uuid.New(), entradaValida())
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
		if _, err := svc.Crear(ctx, uuid.New(), entrada); err != nil {
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

func TestNormalizarFiltro(t *testing.T) {
	t.Run("limite por encima del maximo se recorta", func(t *testing.T) {
		f := NormalizarFiltro(repository.Filtro{Limite: 5000})
		if f.Limite != LimiteMaximo {
			t.Errorf("Limite = %d, se esperaba %d", f.Limite, LimiteMaximo)
		}
	})

	t.Run("limite cero toma el default", func(t *testing.T) {
		f := NormalizarFiltro(repository.Filtro{Limite: 0})
		if f.Limite != LimitePorDefecto {
			t.Errorf("Limite = %d, se esperaba el default %d", f.Limite, LimitePorDefecto)
		}
	})

	t.Run("limite negativo toma el default", func(t *testing.T) {
		f := NormalizarFiltro(repository.Filtro{Limite: -5})
		if f.Limite != LimitePorDefecto {
			t.Errorf("Limite = %d, se esperaba el default %d", f.Limite, LimitePorDefecto)
		}
	})

	t.Run("desplazamiento negativo se pisa en cero", func(t *testing.T) {
		f := NormalizarFiltro(repository.Filtro{Desplazamiento: -10})
		if f.Desplazamiento != 0 {
			t.Errorf("Desplazamiento = %d, se esperaba 0", f.Desplazamiento)
		}
	})

	t.Run("estado nil toma activo por defecto", func(t *testing.T) {
		f := NormalizarFiltro(repository.Filtro{})
		if f.Estado == nil || *f.Estado != domain.EstadoActivo {
			t.Errorf("Estado = %v, se esperaba activo", f.Estado)
		}
	})

	t.Run("un estado explícito no se toca", func(t *testing.T) {
		inactivo := domain.EstadoInactivo
		f := NormalizarFiltro(repository.Filtro{Estado: &inactivo})
		if f.Estado == nil || *f.Estado != domain.EstadoInactivo {
			t.Errorf("Estado = %v, se esperaba que se respetara inactivo", f.Estado)
		}
	})

	t.Run("es idempotente", func(t *testing.T) {
		una := NormalizarFiltro(repository.Filtro{Limite: 5000, Desplazamiento: -10})
		dos := NormalizarFiltro(una)
		if una != dos {
			t.Errorf("normalizar dos veces dio distinto de normalizar una: %+v vs %+v", una, dos)
		}
	})
}

func TestActualizar(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	entrada := entradaValida()
	entrada.Bio = "Bio actualizada."
	entrada.Zona = "GBA Norte"

	actualizado, err := svc.Actualizar(ctx, p.UsuarioID, p.ID, entrada)
	if err != nil {
		t.Fatalf("Actualizar devolvió error: %v", err)
	}
	if actualizado.Bio != "Bio actualizada." || actualizado.Zona != "GBA Norte" {
		t.Error("los campos editables no se aplicaron")
	}
	if actualizado.Slug != p.Slug {
		t.Error("el slug es una URL pública y no debía cambiar")
	}

	// tiene que haber quedado persistido
	obtenido, _ := svc.ObtenerPorID(ctx, p.ID)
	if obtenido.Zona != "GBA Norte" {
		t.Error("el cambio no quedó guardado")
	}
}

func TestActualizarNoExiste(t *testing.T) {
	_, err := nuevoServicioDePrueba().Actualizar(context.Background(), uuid.New(), uuid.New(), entradaValida())
	if !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestActualizarMatriculaDeOtro(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	primero, err := svc.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("alta 1 falló: %v", err)
	}

	entrada2 := entradaValida()
	entrada2.Nombre = "Carolina"
	entrada2.Apellido = "Vega"
	entrada2.Matricula = "MN 11111"
	if _, err := svc.Crear(ctx, uuid.New(), entrada2); err != nil {
		t.Fatalf("alta 2 falló: %v", err)
	}

	// el primero intenta quedarse con la matrícula del segundo
	robo := entradaValida()
	robo.Matricula = "MN 11111"
	if _, err := svc.Actualizar(ctx, primero.UsuarioID, primero.ID, robo); !errors.Is(err, domain.ErrMatriculaEnUso) {
		t.Errorf("se esperaba ErrMatriculaEnUso, se obtuvo %v", err)
	}
}

func TestActualizarConservaLaPropiaMatricula(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	// editar sin cambiar la matrícula no puede chocar consigo mismo
	entrada := entradaValida()
	entrada.Bio = "Otra bio."
	if _, err := svc.Actualizar(ctx, p.UsuarioID, p.ID, entrada); err != nil {
		t.Errorf("editar conservando la matrícula propia falló: %v", err)
	}
}

func TestActualizarResetaVerificacion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProfesional()
	svc := NuevoProfesional(repo)

	p, err := svc.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	// simular que ya fue verificado
	p.Verificacion = domain.VerificacionVerificada
	if err := repo.Actualizar(ctx, p); err != nil {
		t.Fatalf("no se pudo preparar el estado: %v", err)
	}

	entrada := entradaValida()
	entrada.Matricula = "MN 55555"
	actualizado, err := svc.Actualizar(ctx, p.UsuarioID, p.ID, entrada)
	if err != nil {
		t.Fatalf("Actualizar devolvió error: %v", err)
	}
	if actualizado.Verificacion != domain.VerificacionPendiente {
		t.Error("cambiar la matrícula tenía que invalidar la verificación")
	}
}

func TestDarDeBajaYReactivar(t *testing.T) {
	ctx := context.Background()
	svc := nuevoServicioDePrueba()

	p, err := svc.Crear(ctx, uuid.New(), entradaValida())
	if err != nil {
		t.Fatalf("Crear devolvió error: %v", err)
	}

	if err := svc.DarDeBaja(ctx, p.UsuarioID, p.ID); err != nil {
		t.Fatalf("DarDeBaja devolvió error: %v", err)
	}

	// el registro sigue existiendo: un turno pasado apunta acá
	obtenido, err := svc.ObtenerPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("el profesional dado de baja tenía que seguir existiendo: %v", err)
	}
	if obtenido.Estado != domain.EstadoInactivo {
		t.Errorf("Estado = %q, se esperaba inactivo", obtenido.Estado)
	}
	if obtenido.DadoDeBajaEn == nil {
		t.Error("DadoDeBajaEn debía sellarse")
	}

	// pero no aparece en el listado por defecto
	_, total, _ := svc.Listar(ctx, repository.Filtro{})
	if total != 0 {
		t.Errorf("el listado por defecto devolvió %d, se esperaba 0", total)
	}

	// filtrando explícitamente sí aparece
	inactivo := domain.EstadoInactivo
	_, total, _ = svc.Listar(ctx, repository.Filtro{Estado: &inactivo})
	if total != 1 {
		t.Errorf("el listado de inactivos devolvió %d, se esperaba 1", total)
	}

	// dar de baja dos veces es idempotente, no un error
	if err := svc.DarDeBaja(ctx, p.UsuarioID, p.ID); err != nil {
		t.Errorf("la segunda baja debía ser idempotente, devolvió %v", err)
	}

	reactivado, err := svc.Reactivar(ctx, p.UsuarioID, p.ID)
	if err != nil {
		t.Fatalf("Reactivar devolvió error: %v", err)
	}
	if reactivado.Estado != domain.EstadoActivo || reactivado.DadoDeBajaEn != nil {
		t.Error("reactivar debía dejarlo activo y limpiar DadoDeBajaEn")
	}

	_, total, _ = svc.Listar(ctx, repository.Filtro{})
	if total != 1 {
		t.Errorf("después de reactivar el listado devolvió %d, se esperaba 1", total)
	}
}

func TestDarDeBajaNoExiste(t *testing.T) {
	if err := nuevoServicioDePrueba().DarDeBaja(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, domain.ErrNoEncontrado) {
		t.Errorf("se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}
