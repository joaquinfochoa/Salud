package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func TestEscribirErrorMapeaLosErroresDelDominio(t *testing.T) {
	casos := []struct {
		nombre         string
		err            error
		statusEsperado int
	}{
		{"no encontrado", domain.ErrNoEncontrado, http.StatusNotFound},
		{"matricula tomada", domain.ErrMatriculaEnUso, http.StatusConflict},
		{
			"validacion",
			domain.ErrorValidacion{Campos: []domain.ErrorCampo{{Campo: "zona", Mensaje: "es obligatoria"}}},
			http.StatusUnprocessableEntity,
		},
		{"desconocido", errors.New("algo explotó"), http.StatusInternalServerError},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			escribirError(rec, req, caso.err)

			if rec.Code != caso.statusEsperado {
				t.Errorf("status = %d, se esperaba %d", rec.Code, caso.statusEsperado)
			}
			if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
				t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
			}

			var p Problema
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
				t.Fatalf("el cuerpo no es JSON válido: %v", err)
			}
			if p.Estado != caso.statusEsperado {
				t.Errorf("problem.status = %d, se esperaba %d", p.Estado, caso.statusEsperado)
			}
			if p.Titulo == "" || p.Tipo == "" {
				t.Error("title y type son obligatorios en RFC 7807")
			}
		})
	}
}

func TestEscribirErrorNoFiltraDetallesInternos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// un error interno no puede llegarle al cliente: puede tener nombres de
	// tablas, rutas del servidor o datos de otro usuario
	escribirError(rec, req, errors.New("pq: relation \"profesionales\" does not exist"))

	if strings.Contains(rec.Body.String(), "profesionales") {
		t.Error("el error interno se filtró al cliente")
	}
}

func TestEscribirErrorDeValidacionIncluyeLosCampos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	verr := domain.ErrorValidacion{Campos: []domain.ErrorCampo{
		{Campo: "matricula", Mensaje: "formato inválido"},
		{Campo: "modalidades", Mensaje: "se requiere al menos una"},
	}}
	escribirError(rec, req, verr)

	var p Problema
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	if len(p.Errores) != 2 {
		t.Fatalf("errores tenía %d elementos, se esperaban 2", len(p.Errores))
	}
	if p.Errores[0].Campo != "matricula" {
		t.Errorf("errores[0].campo = %q", p.Errores[0].Campo)
	}
}

func TestDecodificarJSONRechazaCamposDesconocidos(t *testing.T) {
	rec := httptest.NewRecorder()
	// "precioConsulta" en vez de "precioConsultaCentavos" es exactamente el typo
	// que este modo estricto tiene que atrapar
	cuerpo := strings.NewReader(`{"nombre":"Ana","precioConsulta":100}`)
	req := httptest.NewRequest(http.MethodPost, "/", cuerpo)

	var destino peticionProfesional
	if err := decodificarJSON(rec, req, &destino); err == nil {
		t.Error("un campo desconocido debía ser rechazado")
	}
}

func TestDecodificarJSONRechazaBasuraDespuesDelObjeto(t *testing.T) {
	rec := httptest.NewRecorder()
	cuerpo := strings.NewReader(`{"nombre":"Ana"} {"nombre":"Otro"}`)
	req := httptest.NewRequest(http.MethodPost, "/", cuerpo)

	var destino peticionProfesional
	if err := decodificarJSON(rec, req, &destino); err == nil {
		t.Error("dos objetos JSON seguidos debían ser rechazados")
	}
}
