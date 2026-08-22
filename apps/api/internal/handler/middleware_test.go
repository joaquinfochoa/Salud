package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIDPeticionGeneraUnoSiNoViene(t *testing.T) {
	var visto string
	h := IDPeticion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visto = IDPeticionDe(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if visto == "" {
		t.Error("no se generó un request id")
	}
	if rec.Header().Get("X-Request-ID") != visto {
		t.Error("el request id tenía que volver en el header de la respuesta")
	}
}

func TestIDPeticionRespetaElQueViene(t *testing.T) {
	var visto string
	h := IDPeticion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visto = IDPeticionDe(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "trae-el-suyo")

	h.ServeHTTP(httptest.NewRecorder(), req)

	// permite seguir un request a través de varios servicios
	if visto != "trae-el-suyo" {
		t.Errorf("request id = %q, se esperaba el que vino en el header", visto)
	}
}

func TestRecuperarPanicDevuelve500(t *testing.T) {
	h := RecuperarPanic(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("algo explotó")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, se esperaba 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}
	// el mensaje del panic no puede llegarle al cliente
	if strings.Contains(rec.Body.String(), "algo explotó") {
		t.Error("el mensaje del panic se filtró al cliente")
	}
}

func TestEncadenarAplicaEnOrden(t *testing.T) {
	var orden []string

	marcar := func(nombre string) func(http.Handler) http.Handler {
		return func(siguiente http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				orden = append(orden, nombre)
				siguiente.ServeHTTP(w, r)
			})
		}
	}

	final := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		orden = append(orden, "handler")
	})

	Encadenar(final, marcar("primero"), marcar("segundo")).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	esperado := []string{"primero", "segundo", "handler"}
	if len(orden) != len(esperado) {
		t.Fatalf("orden = %v, se esperaba %v", orden, esperado)
	}
	for i := range esperado {
		if orden[i] != esperado[i] {
			t.Fatalf("orden = %v, se esperaba %v", orden, esperado)
		}
	}
}
