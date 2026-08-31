package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

// servidorConTurnos arma un servidor con un profesional que ya tiene horario
// cargado, y devuelve su id y el primer hueco libre. El cliente queda logueado
// como el dueño de ese perfil.
func servidorConTurnos(t *testing.T) (*httptest.Server, string, string) {
	t.Helper()
	srv := servidorConAgenda(t) // ya viene con sesión del profesional

	var creado respuestaProfesional
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	if err := json.NewDecoder(resp.Body).Decode(&creado); err != nil {
		t.Fatalf("no se pudo decodificar el alta: %v", err)
	}

	horarios := `{"horarios":[{"diaSemana":"lunes","desde":"09:00","hasta":"13:00","duracionMin":50,"modalidad":"telemedicina"}]}`
	if r := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+creado.ID+"/horarios", horarios); r.StatusCode != http.StatusOK {
		t.Fatalf("cargando horarios: estado %d", r.StatusCode)
	}

	return srv, creado.ID, primerHueco(t, srv, creado.ID)
}

func huecosDe(t *testing.T, srv *httptest.Server, profesionalID string) []string {
	t.Helper()
	lunes := proximoLunes(t)
	resp := obtener(t, srv, "/api/v1/profesionales/"+profesionalID+"/huecos?desde="+lunes+"&hasta="+lunes)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET huecos devolvió %d", resp.StatusCode)
	}

	var cuerpo struct {
		Datos []struct {
			Inicio string `json:"inicio"`
		} `json:"datos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cuerpo); err != nil {
		t.Fatalf("no se pudo decodificar los huecos: %v", err)
	}

	inicios := make([]string, 0, len(cuerpo.Datos))
	for _, h := range cuerpo.Datos {
		inicios = append(inicios, h.Inicio)
	}
	return inicios
}

func primerHueco(t *testing.T, srv *httptest.Server, profesionalID string) string {
	t.Helper()
	huecos := huecosDe(t, srv, profesionalID)
	if len(huecos) == 0 {
		t.Fatal("el profesional de prueba no tiene ningún hueco")
	}
	return huecos[0]
}

func reservar(t *testing.T, srv *httptest.Server, profesionalID, inicio string) *http.Response {
	t.Helper()
	return postear(t, srv, "/api/v1/profesionales/"+profesionalID+"/turnos",
		`{"inicio":"`+inicio+`","motivo":"dolor lumbar"}`)
}

func TestReservarDevuelve201YLocation(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)
	conSesion(t, srv, "paciente@ejemplo.com")

	resp := reservar(t, srv, profesionalID, hueco)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("estado = %d, se esperaba 201", resp.StatusCode)
	}

	var turno respuestaTurno
	if err := json.NewDecoder(resp.Body).Decode(&turno); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if loc := resp.Header.Get("Location"); loc != "/api/v1/turnos/"+turno.ID {
		t.Errorf("Location = %q, se esperaba la URL del turno", loc)
	}
	if turno.Estado != "reservado" {
		t.Errorf("estado = %q, se esperaba reservado", turno.Estado)
	}
	// Fin y modalidad los pone el servidor, no el cliente.
	if turno.Fin.IsZero() || turno.Modalidad == "" {
		t.Error("el servidor no completó fin y modalidad desde el hueco")
	}
	// Las dos claves de cancelación salen aunque estén vacías: van en required.
	if turno.CanceladoEn != nil || turno.CanceladoPor != nil {
		t.Error("un turno nuevo no puede venir cancelado")
	}
}

func TestCancelarDejaElTurnoCancelado(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)
	conSesion(t, srv, "paciente@ejemplo.com")

	resp := reservar(t, srv, profesionalID, hueco)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo reservar: %d", resp.StatusCode)
	}
	var turno respuestaTurno
	if err := json.NewDecoder(resp.Body).Decode(&turno); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}

	if r := ejecutar(t, srv, http.MethodDelete, "/api/v1/turnos/"+turno.ID, ""); r.StatusCode != http.StatusNoContent {
		t.Fatalf("cancelar devolvió %d, se esperaba 204", r.StatusCode)
	}

	// Sigue en el listado, cancelado: es parte del historial de las dos
	// partes, y esconderlo sería esconderle al paciente que perdió su turno.
	var lista respuestaListaTurnosConProfesional
	if err := json.NewDecoder(obtener(t, srv, "/api/v1/turnos").Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if len(lista.Datos) != 1 {
		t.Fatalf("se devolvieron %d turnos, se esperaba 1", len(lista.Datos))
	}
	if lista.Datos[0].Estado != "cancelado" {
		t.Errorf("estado = %q, se esperaba cancelado", lista.Datos[0].Estado)
	}
	if lista.Datos[0].CanceladoEn == nil || lista.Datos[0].CanceladoPor == nil {
		t.Error("no quedó registrado cuándo ni quién canceló")
	}

	// Cancelar dos veces no es idempotente, a propósito.
	if r := ejecutar(t, srv, http.MethodDelete, "/api/v1/turnos/"+turno.ID, ""); r.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("segunda cancelación = %d, se esperaba 422", r.StatusCode)
	}
}

func TestReservarUnInicioInvalidoDa422(t *testing.T) {
	srv, profesionalID, _ := servidorConTurnos(t)
	conSesion(t, srv, "paciente@ejemplo.com")

	resp := reservar(t, srv, profesionalID, proximoLunes(t)+"T03:00:00-03:00")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("estado = %d, se esperaba 422", resp.StatusCode)
	}

	var p Problema
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}
	for _, e := range p.Errores {
		if e.Campo == "inicio" {
			return
		}
	}
	t.Errorf("no se nombró el campo inicio: %+v", p.Errores)
}

func TestReservarDosVecesDa409(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)

	conSesion(t, srv, "uno@ejemplo.com")
	if r := reservar(t, srv, profesionalID, hueco); r.StatusCode != http.StatusCreated {
		t.Fatalf("primera reserva: %d", r.StatusCode)
	}

	conSesion(t, srv, "dos@ejemplo.com")
	if r := reservar(t, srv, profesionalID, hueco); r.StatusCode != http.StatusConflict {
		t.Errorf("estado = %d, se esperaba 409", r.StatusCode)
	}
}

func TestReservarSinSesionDa401(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)

	anonimo := &http.Client{}
	resp, err := anonimo.Post(srv.URL+"/api/v1/profesionales/"+profesionalID+"/turnos",
		"application/json", strings.NewReader(`{"inicio":"`+hueco+`"}`))
	if err != nil {
		t.Fatalf("POST falló: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("estado = %d, se esperaba 401", resp.StatusCode)
	}
}

func TestCancelarComoTerceroDa403(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)

	conSesion(t, srv, "paciente@ejemplo.com")
	resp := reservar(t, srv, profesionalID, hueco)
	var turno respuestaTurno
	if err := json.NewDecoder(resp.Body).Decode(&turno); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}

	conSesion(t, srv, "tercero@ejemplo.com")
	if r := ejecutar(t, srv, http.MethodDelete, "/api/v1/turnos/"+turno.ID, ""); r.StatusCode != http.StatusForbidden {
		t.Errorf("estado = %d, se esperaba 403", r.StatusCode)
	}
}

func TestMisTurnosSoloDevuelveLosMios(t *testing.T) {
	srv, profesionalID, _ := servidorConTurnos(t)
	huecos := huecosDe(t, srv, profesionalID)

	conSesion(t, srv, "uno@ejemplo.com")
	if r := reservar(t, srv, profesionalID, huecos[0]); r.StatusCode != http.StatusCreated {
		t.Fatalf("reserva de uno: %d", r.StatusCode)
	}

	conSesion(t, srv, "dos@ejemplo.com")
	if r := reservar(t, srv, profesionalID, huecos[1]); r.StatusCode != http.StatusCreated {
		t.Fatalf("reserva de dos: %d", r.StatusCode)
	}

	// seguimos siendo "dos"
	resp := obtener(t, srv, "/api/v1/turnos")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("estado = %d, se esperaba 200", resp.StatusCode)
	}

	var lista respuestaListaTurnosConProfesional
	if err := json.NewDecoder(resp.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if len(lista.Datos) != 1 {
		t.Fatalf("se devolvieron %d turnos, se esperaba solo el propio", len(lista.Datos))
	}
	// Con quién es el turno: sin esto la pantalla muestra una hora y nada más,
	// que es la mitad de lo que el paciente necesita saber.
	if lista.Datos[0].Profesional.Nombre == "" || lista.Datos[0].Profesional.Slug == "" {
		t.Errorf("el turno vino sin el profesional: %+v", lista.Datos[0].Profesional)
	}
}

func TestMisTurnosSinSesionDa401(t *testing.T) {
	srv := servidorAnonimo(t)

	if resp := obtener(t, srv, "/api/v1/turnos"); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("estado = %d, se esperaba 401", resp.StatusCode)
	}
}

// La agenda ocupada es privada del dueño, a diferencia de /horarios y
// /bloqueos: dice quién es paciente de quién.
func TestAgendaDelProfesionalRequiereSerDueno(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)

	conSesion(t, srv, "paciente@ejemplo.com")
	if r := reservar(t, srv, profesionalID, hueco); r.StatusCode != http.StatusCreated {
		t.Fatalf("reserva: %d", r.StatusCode)
	}

	// el paciente no es el dueño del perfil
	if r := obtener(t, srv, "/api/v1/profesionales/"+profesionalID+"/turnos"); r.StatusCode != http.StatusForbidden {
		t.Errorf("estado = %d, se esperaba 403", r.StatusCode)
	}
}

func TestReservarConContentTypeInvalidoDa415(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)
	conSesion(t, srv, "paciente@ejemplo.com")

	req, err := http.NewRequest(http.MethodPost,
		srv.URL+"/api/v1/profesionales/"+profesionalID+"/turnos",
		strings.NewReader(`{"inicio":"`+hueco+`"}`))
	if err != nil {
		t.Fatalf("no se pudo armar el request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST falló: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("estado = %d, se esperaba 415", resp.StatusCode)
	}
}

// Criterios de aceptación 2 y 7 medidos en HTTP: el hueco desaparece al
// reservar y vuelve al cancelar.
func TestElHuecoDesapareceYVuelveAlCancelar(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)
	antes := len(huecosDe(t, srv, profesionalID))

	conSesion(t, srv, "paciente@ejemplo.com")
	resp := reservar(t, srv, profesionalID, hueco)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo reservar: %d", resp.StatusCode)
	}
	var turno respuestaTurno
	if err := json.NewDecoder(resp.Body).Decode(&turno); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}

	if despues := len(huecosDe(t, srv, profesionalID)); despues != antes-1 {
		t.Fatalf("después de reservar hay %d huecos, se esperaban %d", despues, antes-1)
	}

	if r := ejecutar(t, srv, http.MethodDelete, "/api/v1/turnos/"+turno.ID, ""); r.StatusCode != http.StatusNoContent {
		t.Fatalf("cancelar devolvió %d, se esperaba 204", r.StatusCode)
	}

	// Cancelar libera el hueco: es el punto de cancelar.
	if despues := len(huecosDe(t, srv, profesionalID)); despues != antes {
		t.Errorf("después de cancelar hay %d huecos, se esperaban %d", despues, antes)
	}
}

// El teléfono del paciente es dato personal bajo Ley 25.326. La regla es que
// lo ve el profesional con quien reservó, y nadie más.
//
// Este test existe porque la regla vive repartida entre dos DTO, y es
// exactamente el tipo de cosa que se rompe agregando un campo "para
// debuggear" y no se nota hasta que alguien la encuentra en el HTML.
func TestElTelefonoDelPacienteNoSeFiltra(t *testing.T) {
	srv, profesionalID, hueco := servidorConTurnos(t)

	perfil := decodificarProfesional(t, obtener(t, srv, "/api/v1/profesionales/"+profesionalID))

	conSesion(t, srv, "paciente.telefono@ejemplo.com")
	if r := reservar(t, srv, profesionalID, hueco); r.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo reservar: %d", r.StatusCode)
	}

	// La página pública del profesional no puede mencionar ningún teléfono: la
	// ve cualquiera, incluidos los rastreadores.
	publico := leerTodo(t, obtener(t, srv, "/api/v1/perfiles/"+perfil.Slug))
	if strings.Contains(publico, "telefono") {
		t.Errorf("el perfil público menciona un teléfono: %s", publico)
	}

	// El paciente, mirando SUS turnos, tampoco ve el del profesional.
	mios := leerTodo(t, obtener(t, srv, "/api/v1/turnos"))
	if strings.Contains(mios, "telefono") {
		t.Errorf("el listado del paciente trae un teléfono: %s", mios)
	}

	// El profesional, en cambio, sí ve el de quien le reservó: es con quien
	// tiene el turno, y es cómo avisa si se le cae la mañana.
	entrarComo(t, srv, "agenda@ejemplo.com")
	agenda := leerTodo(t, obtener(t, srv, "/api/v1/profesionales/"+profesionalID+"/turnos"))
	// El número, no la clave: `"telefono":""` también contiene `"telefono"`, y
	// una primera versión de este test pasaba con el campo vacío.
	if !strings.Contains(agenda, "1234-5678") {
		t.Errorf("la agenda del profesional no trae el teléfono del paciente: %s", agenda)
	}
}

// entrarComo vuelve a una cuenta que YA existe. conSesion registra una nueva,
// que para el profesional del servidor de prueba devuelve 409.
func entrarComo(t *testing.T, srv *httptest.Server, email string) {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("no se pudo crear el cookie jar: %v", err)
	}
	srv.Client().Jar = jar

	cuerpo := `{"email":"` + email + `","contrasena":"desarrollo123"}`
	if r := postear(t, srv, "/api/v1/sesiones", cuerpo); r.StatusCode != http.StatusCreated {
		t.Fatalf("login de %s devolvió %d", email, r.StatusCode)
	}
}

func leerTodo(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("no se pudo leer el cuerpo: %v", err)
	}
	return string(b)
}
