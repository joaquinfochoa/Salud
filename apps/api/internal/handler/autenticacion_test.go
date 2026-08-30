package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const cuerpoRegistro = `{
  "email": "juan@ejemplo.com",
  "contrasena": "unaclave8",
  "nombre": "Juan",
  "apellido": "Pérez"
}`

func cookieDeSesion(t *testing.T, resp *http.Response) *http.Cookie {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == nombreCookieSesion {
			return c
		}
	}
	return nil
}

func TestRegistroDevuelveCookie(t *testing.T) {
	srv := servidorAnonimo(t)

	resp := postear(t, srv, "/api/v1/usuarios", cuerpoRegistro)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("estado = %d, se esperaba 201", resp.StatusCode)
	}

	cookie := cookieDeSesion(t, resp)
	if cookie == nil {
		t.Fatal("no vino la cookie de sesión")
	}
	if cookie.Value == "" {
		t.Error("la cookie vino vacía")
	}
	if !cookie.HttpOnly {
		t.Error("la cookie no es HttpOnly: cualquier XSS puede leerla")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Error("la cookie no es SameSite=Lax")
	}
	if cookie.Path != "/" {
		t.Errorf("Path = %q, se esperaba /", cookie.Path)
	}

	cuerpo, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("leyendo el cuerpo: %v", err)
	}
	// El token viaja SOLO en el Set-Cookie. Si además aparece en el JSON,
	// cualquier script de la página puede leerlo y el HttpOnly no sirvió de
	// nada.
	if strings.Contains(string(cuerpo), cookie.Value) {
		t.Error("el token de sesión salió en el cuerpo de la respuesta")
	}
	if strings.Contains(string(cuerpo), "unaclave8") || strings.Contains(string(cuerpo), "hash") {
		t.Error("la respuesta filtró la contraseña o el hash")
	}
}

func TestRegistroEmailDuplicado(t *testing.T) {
	srv := servidorAnonimo(t)

	primero := postear(t, srv, "/api/v1/usuarios", cuerpoRegistro)
	if primero.StatusCode != http.StatusCreated {
		t.Fatalf("el primer registro devolvió %d", primero.StatusCode)
	}

	segundo := postear(t, srv, "/api/v1/usuarios", cuerpoRegistro)
	if segundo.StatusCode != http.StatusConflict {
		t.Errorf("estado = %d, se esperaba 409", segundo.StatusCode)
	}
}

func TestRegistroInvalidoNombraElCampo(t *testing.T) {
	srv := servidorAnonimo(t)

	resp := postear(t, srv, "/api/v1/usuarios", `{
	  "email": "juan@ejemplo.com",
	  "contrasena": "corta",
	  "nombre": "Juan",
	  "apellido": "Pérez"
	}`)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("estado = %d, se esperaba 422", resp.StatusCode)
	}

	var p Problema
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	for _, e := range p.Errores {
		if e.Campo == "contrasena" {
			return
		}
	}
	t.Errorf("no se nombró el campo contrasena: %+v", p.Errores)
}

func TestLoginCorrecto(t *testing.T) {
	srv := servidorAnonimo(t)
	postear(t, srv, "/api/v1/usuarios", cuerpoRegistro)

	resp := postear(t, srv, "/api/v1/sesiones", `{"email":"juan@ejemplo.com","contrasena":"unaclave8"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("estado = %d, se esperaba 201", resp.StatusCode)
	}
	if cookieDeSesion(t, resp) == nil {
		t.Error("el login no dejó cookie")
	}
}

// El 401 tiene que ser indistinguible entre "ese email no existe" y "esa
// contraseña está mal". Si se separan, probando direcciones se arma el padrón
// de usuarios sin adivinar una sola contraseña.
func TestLoginIncorrectoNoDistingueLaCausa(t *testing.T) {
	srv := servidorAnonimo(t)
	postear(t, srv, "/api/v1/usuarios", cuerpoRegistro)

	cuerpos := map[string]string{
		"email inexistente":     `{"email":"otro@ejemplo.com","contrasena":"unaclave8"}`,
		"contrasena incorrecta": `{"email":"juan@ejemplo.com","contrasena":"otraclave"}`,
	}

	respuestas := make(map[string]string, len(cuerpos))
	for nombre, cuerpo := range cuerpos {
		resp := postear(t, srv, "/api/v1/sesiones", cuerpo)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: estado = %d, se esperaba 401", nombre, resp.StatusCode)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("%s: leyendo el cuerpo: %v", nombre, err)
		}
		respuestas[nombre] = string(b)
		if cookieDeSesion(t, resp) != nil {
			t.Errorf("%s: un login fallido dejó cookie", nombre)
		}
	}

	if respuestas["email inexistente"] != respuestas["contrasena incorrecta"] {
		t.Errorf("las dos causas dan respuestas distintas:\n  %s\n  %s",
			respuestas["email inexistente"], respuestas["contrasena incorrecta"])
	}
}

func TestYoSinSesion(t *testing.T) {
	srv := servidorAnonimo(t)

	if resp := obtener(t, srv, "/api/v1/usuarios/yo"); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("estado = %d, se esperaba 401", resp.StatusCode)
	}
}

func TestYoConSesionSinPerfil(t *testing.T) {
	srv := servidorAnonimo(t)
	conSesion(t, srv, "juan@ejemplo.com")

	resp := obtener(t, srv, "/api/v1/usuarios/yo")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("estado = %d, se esperaba 200", resp.StatusCode)
	}

	var yo respuestaUsuarioActual
	if err := json.NewDecoder(resp.Body).Decode(&yo); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if yo.Email != "juan@ejemplo.com" {
		t.Errorf("email = %q", yo.Email)
	}
	// No tener perfil profesional no es un error: es la mitad de los usuarios.
	if yo.PerfilProfesionalID != nil {
		t.Errorf("perfilProfesionalId = %v, se esperaba null", *yo.PerfilProfesionalID)
	}
}

func TestYoConPerfil(t *testing.T) {
	srv := nuevoServidorDePrueba(t) // ya viene con sesión
	if resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido); resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo crear el perfil")
	}

	resp := obtener(t, srv, "/api/v1/usuarios/yo")
	var yo respuestaUsuarioActual
	if err := json.NewDecoder(resp.Body).Decode(&yo); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if yo.PerfilProfesionalID == nil {
		t.Fatal("perfilProfesionalId = null, se esperaba el id del perfil recién creado")
	}
}

func TestLogoutInvalidaLaSesion(t *testing.T) {
	srv := servidorAnonimo(t)
	conSesion(t, srv, "juan@ejemplo.com")

	if resp := obtener(t, srv, "/api/v1/usuarios/yo"); resp.StatusCode != http.StatusOK {
		t.Fatalf("la sesión no servía antes del logout: %d", resp.StatusCode)
	}

	if resp := ejecutar(t, srv, http.MethodDelete, "/api/v1/sesiones/actual", ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("estado = %d, se esperaba 204", resp.StatusCode)
	}

	if resp := obtener(t, srv, "/api/v1/usuarios/yo"); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("después del logout la sesión sigue viva: %d", resp.StatusCode)
	}
}

// El logout es idempotente: un cliente que reintenta no está haciendo nada mal.
func TestLogoutSinSesion(t *testing.T) {
	srv := servidorAnonimo(t)

	if resp := ejecutar(t, srv, http.MethodDelete, "/api/v1/sesiones/actual", ""); resp.StatusCode != http.StatusNoContent {
		t.Errorf("estado = %d, se esperaba 204", resp.StatusCode)
	}
}

// Criterio de aceptación 3, medido en HTTP: el dueño puede, un segundo usuario
// recibe 403 en el perfil, los horarios y los bloqueos.
func TestUnIntrusoNoPuedeEditarUnPerfilAjeno(t *testing.T) {
	srv := nuevoServidorDePrueba(t)

	var creado respuestaProfesional
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	if err := json.NewDecoder(resp.Body).Decode(&creado); err != nil {
		t.Fatalf("no se pudo decodificar el alta: %v", err)
	}

	// conSesion cambia el jar del cliente: a partir de acá somos otro usuario
	conSesion(t, srv, "intruso@ejemplo.com")

	base := "/api/v1/profesionales/" + creado.ID
	casos := []struct {
		nombre string
		metodo string
		ruta   string
		cuerpo string
	}{
		{"editar perfil", http.MethodPut, base, cuerpoValido},
		{"dar de baja", http.MethodDelete, base, ""},
		{"reactivar", http.MethodPost, base + "/reactivar", ""},
		{"cargar horarios", http.MethodPut, base + "/horarios", cuerpoHorarios},
		{"crear bloqueo", http.MethodPost, base + "/bloqueos", `{"desde":"2099-01-01T00:00:00-03:00","hasta":"2099-01-10T00:00:00-03:00"}`},
		{"eliminar bloqueo", http.MethodDelete, base + "/bloqueos/6ba7b810-9dad-11d1-80b4-00c04fd430c8", ""},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resp := ejecutar(t, srv, c.metodo, c.ruta, c.cuerpo)
			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("estado = %d, se esperaba 403", resp.StatusCode)
			}
		})
	}
}

// Criterio de aceptación 4: sin cookie, todo lo público sigue igual.
func TestLasRutasPublicasNoPidenSesion(t *testing.T) {
	srv := nuevoServidorDePrueba(t)

	var creado respuestaProfesional
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	if err := json.NewDecoder(resp.Body).Decode(&creado); err != nil {
		t.Fatalf("no se pudo decodificar el alta: %v", err)
	}

	// un cliente sin jar nunca manda cookies
	anonimo := &http.Client{}
	base := srv.URL + "/api/v1/profesionales/" + creado.ID

	rutas := []string{
		srv.URL + "/api/v1/profesionales",
		base,
		base + "/horarios",
		base + "/bloqueos",
		srv.URL + "/api/v1/perfiles/" + creado.Slug,
	}

	for _, ruta := range rutas {
		t.Run(ruta, func(t *testing.T) {
			resp, err := anonimo.Get(ruta)
			if err != nil {
				t.Fatalf("GET %s: %v", ruta, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("estado = %d, se esperaba 200", resp.StatusCode)
			}
		})
	}
}

// Sin sesión, las privadas dan 401 y no 403: 401 es "no sé quién sos", 403 es
// "sé quién sos y no te alcanza".
func TestLasRutasPrivadasSinSesionDan401(t *testing.T) {
	srv := servidorAnonimo(t)
	id := "/api/v1/profesionales/6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	casos := []struct{ metodo, ruta, cuerpo string }{
		{http.MethodPost, "/api/v1/profesionales", cuerpoValido},
		{http.MethodPut, id, cuerpoValido},
		{http.MethodDelete, id, ""},
		{http.MethodPost, id + "/reactivar", ""},
		{http.MethodPut, id + "/horarios", cuerpoHorarios},
		{http.MethodPost, id + "/bloqueos", `{"desde":"2099-01-01T00:00:00-03:00","hasta":"2099-01-10T00:00:00-03:00"}`},
	}

	for _, c := range casos {
		t.Run(c.metodo+" "+c.ruta, func(t *testing.T) {
			resp := ejecutar(t, srv, c.metodo, c.ruta, c.cuerpo)
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("estado = %d, se esperaba 401", resp.StatusCode)
			}
		})
	}
}

// Criterio de aceptación 7, en el stack real y no solo en decodificarJSON.
func TestPostConContentTypeIncorrectoDa415(t *testing.T) {
	srv := nuevoServidorDePrueba(t)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/profesionales", strings.NewReader(cuerpoValido))
	if err != nil {
		t.Fatalf("no se pudo armar el request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST falló: %v", err)
	}

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("estado = %d, se esperaba 415", resp.StatusCode)
	}
}
