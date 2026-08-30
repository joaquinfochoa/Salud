package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

func nuevoRouterDePrueba() http.Handler {
	return nuevoStackDePrueba()
}

// nuevoStackDePrueba cablea el stack real de punta a punta, sin mocks. Es el
// único lugar del paquete que sabe cómo se arma el router: los tres helpers de
// servidor lo usan.
//
// cookieSegura en false porque httptest sirve por HTTP: con Secure el cliente
// nunca devolvería la cookie y todos los tests privados darían 401 por el
// motivo equivocado.
func nuevoStackDePrueba() http.Handler {
	repo := memory.NuevoProfesional()
	svc := service.NuevoProfesional(repo)
	svcAgenda := service.NuevaAgenda(repo, memory.NuevoHorarioSemanal(), memory.NuevoBloqueo())
	svcAuth := service.NuevaAutenticacion(memory.NuevoUsuario(), memory.NuevaSesion())
	svcTurnos := service.NuevoTurno(memory.NuevoTurno(), repo, svcAgenda)

	return NuevoRouter(
		NuevoProfesional(svc),
		NuevaAgenda(svcAgenda),
		NuevaAutenticacion(svcAuth, svc, false),
		NuevoTurno(svcTurnos),
	)
}

// TestRutaInexistenteDevuelveProblemJSON cubre el 404 que arma ServeMux
// solo, sin pasar por ningún handler nuestro. La stdlib lo devuelve en texto
// plano; el contrato promete problem+json en todos los errores.
func TestRutaInexistenteDevuelveProblemJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	nuevoRouterDePrueba().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-existe", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}

	var p Problema
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	if p.Estado != http.StatusNotFound || p.Titulo == "" || p.Tipo == "" {
		t.Errorf("problema incompleto: %+v", p)
	}
}

// TestMetodoNoPermitidoDevuelveProblemJSON cubre el 405 que arma ServeMux
// cuando la ruta existe pero el método no.
func TestMetodoNoPermitidoDevuelveProblemJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	// /healthz solo registra GET
	nuevoRouterDePrueba().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/healthz", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, se esperaba 405", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}

	var p Problema
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	if p.Estado != http.StatusMethodNotAllowed {
		t.Errorf("problem.status = %d, se esperaba 405", p.Estado)
	}
}

// TestNuestrosPropios404NoSeTocan verifica que el interceptor no reescribe un
// 404 que ya viene en problem+json desde nuestro propio handler: solo
// reemplaza el que arma la stdlib.
func TestNuestrosPropios404NoSeTocan(t *testing.T) {
	rec := httptest.NewRecorder()
	nuevoRouterDePrueba().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/profesionales/6ba7b810-9dad-11d1-80b4-00c04fd430c8", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404", rec.Code)
	}

	var p Problema
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	if p.Detalle != "El profesional solicitado no existe" {
		t.Errorf("detalle = %q, se esperaba el mensaje propio del handler, no el genérico de ruteo", p.Detalle)
	}
}
