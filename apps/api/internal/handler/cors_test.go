package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func handlerConCORS(origenes []string) http.Handler {
	base := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	return CORS(origenes)(base)
}

func TestCORSReflejaElOrigenPermitido(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profesionales", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	handlerConCORS([]string{"http://localhost:3000"}).ServeHTTP(rec, req)

	// Se refleja el origen exacto, nunca "*": con credenciales el browser
	// rechaza el comodín, y además abriría la API a cualquier sitio.
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Allow-Origin = %q, se esperaba el origen exacto", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q: sin esto la cookie de sesión no viaja", got)
	}
	// Sin Vary, un caché intermedio puede servirle a un origen la respuesta
	// que se armó para otro.
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, se esperaba Origin", got)
	}
}

func TestCORSIgnoraUnOrigenNoPermitido(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profesionales", nil)
	req.Header.Set("Origin", "https://sitio-malicioso.com")

	handlerConCORS([]string{"http://localhost:3000"}).ServeHTTP(rec, req)

	// La petición se atiende igual —el servidor no rechaza nada—, pero sin el
	// header el browser no le deja leer la respuesta al script. Rechazarla con
	// un 403 sería peor: le confirmaría al atacante que la ruta existe.
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, no debería venir para un origen ajeno", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("estado = %d: la petición se atiende igual", rec.Code)
	}
}

func TestCORSRespondeElPreflight(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/profesionales", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")

	handlerConCORS([]string{"http://localhost:3000"}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("estado = %d, se esperaba 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Allow-Origin = %q", got)
	}
	// Content-Type tiene que estar permitido o el 415 que exigimos en todos
	// los cuerpos se vuelve imposible de satisfacer desde un browser.
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("el preflight no declaró headers permitidos")
	}
	for _, metodo := range []string{"GET", "POST", "PUT", "DELETE"} {
		if !contieneMetodo(rec.Header().Get("Access-Control-Allow-Methods"), metodo) {
			t.Errorf("el preflight no permite %s", metodo)
		}
	}
}

// El preflight de un origen ajeno tampoco recibe permisos, pero se corta ahí:
// no tiene sentido pasarle un OPTIONS al router.
func TestCORSPreflightDeOrigenAjeno(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/profesionales", nil)
	req.Header.Set("Origin", "https://sitio-malicioso.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	handlerConCORS([]string{"http://localhost:3000"}).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, no debería venir", got)
	}
}

// Sin orígenes configurados el middleware no hace nada. Es el estado de
// producción hasta que exista un front desplegado: una lista vacía no puede
// permitir a nadie por accidente.
func TestCORSSinOrigenesNoAgregaNada(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profesionales", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	handlerConCORS(nil).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q con lista vacía", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("estado = %d: las peticiones sin CORS se siguen atendiendo", rec.Code)
	}
}

// Una petición sin Origin —curl, otro servidor, un health check— no es del
// browser y no lleva ningún header de CORS.
func TestCORSSinHeaderOrigin(t *testing.T) {
	rec := httptest.NewRecorder()
	handlerConCORS([]string{"http://localhost:3000"}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q sin header Origin", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("estado = %d", rec.Code)
	}
}

func contieneMetodo(lista, metodo string) bool {
	for _, m := range strings.Split(lista, ",") {
		if strings.TrimSpace(m) == metodo {
			return true
		}
	}
	return false
}
