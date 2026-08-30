package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestIDPeticionGeneraUnoSiNoViene(t *testing.T) {
	var visto string
	h := IDPeticion(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
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
	h := IDPeticion(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
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

// TestIDPeticionReemplazaUnoInvalido es hermano de
// TestIDPeticionRespetaElQueViene: el que viene se respeta solo si tiene
// forma razonable. Sin este límite, un cliente sin autenticar puede adjuntar
// hasta 1 MB de basura en cada request, o forjar el mismo ID que otro para
// arruinar la correlación de logs sin que nadie lo note.
func TestIDPeticionReemplazaUnoInvalido(t *testing.T) {
	casos := []struct {
		nombre string
		id     string
	}{
		{"demasiado largo", strings.Repeat("a", 129)},
		{"con caracteres no imprimibles", "abc\ndef"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			var visto string
			h := IDPeticion(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				visto = IDPeticionDe(r.Context())
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Request-ID", caso.id)

			h.ServeHTTP(httptest.NewRecorder(), req)

			if visto == caso.id {
				t.Error("un id inválido no debía respetarse")
			}
			if visto == "" {
				t.Error("tenía que generarse un id de reemplazo")
			}
		})
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

func TestGrabadorEstadoRegistraElEstadoExplicito(t *testing.T) {
	rec := &grabadorEstado{ResponseWriter: httptest.NewRecorder()}
	rec.WriteHeader(http.StatusCreated)

	if rec.estado != http.StatusCreated {
		t.Errorf("estado = %d, se esperaba %d", rec.estado, http.StatusCreated)
	}
}

func TestGrabadorEstadoUsa200PorDefectoSiNuncaSeLlamaWriteHeader(t *testing.T) {
	rec := &grabadorEstado{ResponseWriter: httptest.NewRecorder()}
	rec.Write([]byte("hola"))

	// nadie llamó WriteHeader: net/http asume 200 en ese caso, y grabadorEstado
	// tiene que replicar esa suposición para que el log no mienta
	if rec.estado != http.StatusOK {
		t.Errorf("estado = %d, se esperaba %d", rec.estado, http.StatusOK)
	}
}

func TestGrabadorEstadoNoPisaUnEstadoExplicitoAlEscribirDespues(t *testing.T) {
	rec := &grabadorEstado{ResponseWriter: httptest.NewRecorder()}
	rec.WriteHeader(http.StatusNotFound)
	rec.Write([]byte("no encontrado"))

	// una implementación que ponga 200 en cada Write sin chequear pisaría esto
	if rec.estado != http.StatusNotFound {
		t.Errorf("estado = %d, se esperaba %d", rec.estado, http.StatusNotFound)
	}
}

func TestGrabadorEstadoAcumulaBytesEntreVariasEscrituras(t *testing.T) {
	rec := &grabadorEstado{ResponseWriter: httptest.NewRecorder()}
	rec.Write([]byte("hola"))   // 4 bytes
	rec.Write([]byte("mundo!")) // 6 bytes

	if rec.bytes != 10 {
		t.Errorf("bytes = %d, se esperaba 10", rec.bytes)
	}
}

func TestRegistrarPeticionesPasaElRequestSinTocarlo(t *testing.T) {
	var metodoVisto, rutaVista string
	h := RegistrarPeticiones(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metodoVisto = r.Method
		rutaVista = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("cuerpo de respuesta"))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/algo", nil))

	if metodoVisto != http.MethodPost {
		t.Errorf("metodo visto por el handler interno = %q, se esperaba %q", metodoVisto, http.MethodPost)
	}
	if rutaVista != "/algo" {
		t.Errorf("ruta vista por el handler interno = %q, se esperaba %q", rutaVista, "/algo")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, se esperaba %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "cuerpo de respuesta" {
		t.Errorf("body = %q, se esperaba %q", rec.Body.String(), "cuerpo de respuesta")
	}
}

func TestRequerirSesionSinUsuario(t *testing.T) {
	llamado := false
	h := RequerirSesion(func(_ http.ResponseWriter, _ *http.Request) {
		llamado = true
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/api/v1/profesionales", nil))

	if llamado {
		t.Error("el handler corrió sin sesión")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("estado = %d, se esperaba 401", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}
}

func TestRequerirSesionConUsuario(t *testing.T) {
	usuarioID := uuid.New()
	var visto uuid.UUID

	h := RequerirSesion(func(_ http.ResponseWriter, r *http.Request) {
		visto, _ = UsuarioIDDe(r.Context())
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profesionales", nil)
	req = req.WithContext(context.WithValue(req.Context(), claveUsuarioID, usuarioID))

	h(httptest.NewRecorder(), req)

	if visto != usuarioID {
		t.Errorf("UsuarioIDDe = %v, se esperaba %v", visto, usuarioID)
	}
}

func TestUsuarioIDDeSinValor(t *testing.T) {
	if _, ok := UsuarioIDDe(context.Background()); ok {
		t.Error("UsuarioIDDe devolvió ok sobre un contexto vacío")
	}
}
