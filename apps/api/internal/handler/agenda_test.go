package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const cuerpoHorarios = `{
  "horarios": [
    {"diaSemana": "lunes", "desde": "09:00", "hasta": "13:00", "duracionMin": 50, "modalidad": "telemedicina"}
  ]
}`

// servidorConAgenda cablea el stack real completo, sin mocks.
func servidorConAgenda(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(nuevoStackDePrueba())
	t.Cleanup(srv.Close)
	conSesion(t, srv, "agenda@ejemplo.com")
	return srv
}

// proximoLunes devuelve el lunes siguiente a dentro de una semana, en formato
// AAAA-MM-DD.
//
// Los tests de handler corren contra el reloj real —el servicio no está
// intervenido— y el horizonte por defecto son 60 días desde hoy, así que una
// fecha clavada como "2099-09-07" quedaría fuera de rango y además dependería
// de que ese día caiga lunes. Calcularla lo vuelve robusto al paso del tiempo.
func proximoLunes(t *testing.T) string {
	t.Helper()
	fecha := time.Now().In(domain.ZonaHoraria).AddDate(0, 0, 7)
	for fecha.Weekday() != time.Monday {
		fecha = fecha.AddDate(0, 0, 1)
	}
	return fecha.Format("2006-01-02")
}

func crearProfesionalPorHTTP(t *testing.T, srv *httptest.Server) respuestaProfesional {
	t.Helper()
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo crear el profesional: status %d", resp.StatusCode)
	}
	return decodificarProfesional(t, resp)
}

func TestPutHorarios(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpoHorarios)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	var lista respuestaHorarios
	if err := json.NewDecoder(resp.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if len(lista.Horarios) != 1 {
		t.Fatalf("len = %d, se esperaba 1", len(lista.Horarios))
	}
	if lista.Horarios[0].Desde != "09:00" {
		t.Errorf("desde = %q, se esperaba 09:00", lista.Horarios[0].Desde)
	}
}

// TestPutHorariosCuerpoDemasiadoGrande cubre que ReemplazarHorarios use
// escribirErrorDeDecodificacion como el resto de los handlers: antes de eso,
// un cuerpo de más de 1 MB daba 400 en esta ruta y 413 en las de profesional.
func TestPutHorariosCuerpoDemasiadoGrande(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	entrada := `{"diaSemana":"lunes","desde":"09:00","hasta":"13:00","duracionMin":50,"modalidad":"telemedicina"},`
	var sb strings.Builder
	sb.WriteString(`{"horarios":[`)
	for sb.Len() <= maxBytesCuerpo {
		sb.WriteString(entrada)
	}
	cuerpo := strings.TrimSuffix(sb.String(), ",") + "]}"

	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpo)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, se esperaba 413", resp.StatusCode)
	}
}

func TestPutHorariosConBloqueSinHuecos(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	cuerpo := strings.Replace(cuerpoHorarios, `"hasta": "13:00"`, `"hasta": "09:30"`, 1)
	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, se esperaba 422", resp.StatusCode)
	}

	var problema Problema
	if err := json.NewDecoder(resp.Body).Decode(&problema); err != nil {
		t.Fatalf("no se pudo decodificar el problema: %v", err)
	}
	if len(problema.Errores) == 0 {
		t.Error("el 422 no nombra ningún campo")
	}
}

func TestPutHorariosConModalidadNoOfrecida(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	cuerpo := strings.Replace(cuerpoHorarios, `"telemedicina"`, `"domicilio"`, 1)
	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, se esperaba 422", resp.StatusCode)
	}
}

func TestGetHorariosDeProfesionalSinAgenda(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/horarios")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	cuerpo := new(bytes.Buffer)
	if _, err := cuerpo.ReadFrom(resp.Body); err != nil {
		t.Fatalf("no se pudo leer el cuerpo: %v", err)
	}
	// lista vacía, nunca null
	if strings.Contains(cuerpo.String(), `"horarios":null`) {
		t.Errorf("horarios llegó como null: %s", cuerpo.String())
	}
}

func TestHorariosDeProfesionalInexistente(t *testing.T) {
	srv := servidorConAgenda(t)
	resp := obtener(t, srv, "/api/v1/profesionales/6ba7b810-9dad-11d1-80b4-00c04fd430c8/horarios")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, se esperaba 404", resp.StatusCode)
	}
}

func TestCicloDeBloqueo(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)
	base := "/api/v1/profesionales/" + p.ID + "/bloqueos"

	// El listado se pide sin desde/hasta más abajo, así que cae en la ventana
	// por defecto (dos años desde hoy): el bloqueo tiene que caer dentro de
	// eso, no en una fecha clavada como 2099 que hoy queda afuera y además
	// envejece mal.
	ahora := time.Now().In(domain.ZonaHoraria)
	desde := ahora.AddDate(0, 0, 30).Format(time.RFC3339)
	hasta := ahora.AddDate(0, 0, 40).Format(time.RFC3339)
	cuerpo := `{"desde": "` + desde + `", "hasta": "` + hasta + `", "motivo": "Vacaciones"}`
	resp := postear(t, srv, base, cuerpo)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201", resp.StatusCode)
	}

	var bloqueo respuestaBloqueo
	if err := json.NewDecoder(resp.Body).Decode(&bloqueo); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if bloqueo.Motivo != "Vacaciones" {
		t.Errorf("motivo = %q", bloqueo.Motivo)
	}
	if loc := resp.Header.Get("Location"); loc != base+"/"+bloqueo.ID {
		t.Errorf("Location = %q, se esperaba %q", loc, base+"/"+bloqueo.ID)
	}

	listado := obtener(t, srv, base)
	var lista respuestaBloqueos
	if err := json.NewDecoder(listado.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar el listado: %v", err)
	}
	if len(lista.Datos) != 1 {
		t.Errorf("el listado devolvió %d bloqueos, se esperaba 1", len(lista.Datos))
	}

	borrado := ejecutar(t, srv, http.MethodDelete, base+"/"+bloqueo.ID, "")
	if borrado.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, se esperaba 204", borrado.StatusCode)
	}

	deNuevo := ejecutar(t, srv, http.MethodDelete, base+"/"+bloqueo.ID, "")
	if deNuevo.StatusCode != http.StatusNotFound {
		t.Errorf("borrar dos veces debía dar 404, dio %d", deNuevo.StatusCode)
	}
}

// TestCrearBloqueoCuerpoDemasiadoGrande es la misma cobertura que
// TestPutHorariosCuerpoDemasiadoGrande, para CrearBloqueo.
func TestCrearBloqueoCuerpoDemasiadoGrande(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	motivo := strings.Repeat("x", maxBytesCuerpo+1)
	cuerpo := `{"desde": "2099-01-01T00:00:00-03:00", "hasta": "2099-01-10T00:00:00-03:00", "motivo": "` + motivo + `"}`

	resp := postear(t, srv, "/api/v1/profesionales/"+p.ID+"/bloqueos", cuerpo)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, se esperaba 413", resp.StatusCode)
	}
}

func TestCrearBloqueoEnElPasado(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	cuerpo := `{"desde": "2020-01-01T00:00:00-03:00", "hasta": "2020-01-10T00:00:00-03:00"}`
	resp := postear(t, srv, "/api/v1/profesionales/"+p.ID+"/bloqueos", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, se esperaba 422", resp.StatusCode)
	}
}

func TestGetHuecos(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	if resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+p.ID+"/horarios", cuerpoHorarios); resp.StatusCode != http.StatusOK {
		t.Fatalf("no se pudo cargar el horario: status %d", resp.StatusCode)
	}

	lunes := proximoLunes(t)
	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde="+lunes+"&hasta="+lunes)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	var lista respuestaHuecos
	if err := json.NewDecoder(resp.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if lista.Rango.Desde != lunes || lista.Rango.Hasta != lunes {
		t.Errorf("rango = %+v, se esperaba el pedido", lista.Rango)
	}
	if len(lista.Datos) != 4 {
		t.Errorf("se obtuvieron %d huecos, se esperaban 4", len(lista.Datos))
	}
}

func TestGetHuecosSinParametros(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestGetHuecosConFechaMalFormada(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde=ayer&hasta=2099-09-07")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

// TestGetHuecosRangoInvertidoEs400 cubre el ajuste del contrato: un rango
// invertido es un query param mal formado, no una entidad semánticamente
// inválida, así que el handler lo rechaza con 400 antes de llegar al
// servicio (que devolvería 422 vía ErrorValidacion).
func TestGetHuecosRangoInvertidoEs400(t *testing.T) {
	srv := servidorConAgenda(t)
	p := crearProfesionalPorHTTP(t, srv)

	resp := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde=2099-09-10&hasta=2099-09-01")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestGetHuecosSeRecortaAlHorizonte(t *testing.T) {
	srv := servidorConAgenda(t)

	// cuerpo propio en vez de parchear cuentaValido con un replace: el replace
	// depende de cómo esté formateada esa constante y se rompe en silencio si
	// alguien la reindenta
	cuerpo := `{
	  "nombre": "Ana", "apellido": "Pérez",
	  "matricula": "MP 55.123", "especialidad": "odontologia",
	  "bio": "Odontóloga general.", "precioConsultaCentavos": 1800000,
	  "modalidades": ["presencial"], "zona": "CABA", "obrasSociales": [],
	  "horizonteDias": 7
	}`
	resp := postear(t, srv, "/api/v1/profesionales", cuerpo)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no se pudo crear el profesional: status %d", resp.StatusCode)
	}
	p := decodificarProfesional(t, resp)

	// El horizonte se cuenta desde hoy, así que las fechas se calculan
	// relativas al reloj real: el servicio no está intervenido en este test.
	hoy := time.Now().In(domain.ZonaHoraria)
	desde := hoy.Format("2006-01-02")
	hasta := hoy.AddDate(0, 3, 0).Format("2006-01-02")
	ultimoEsperado := domain.InicioDelDia(hoy).AddDate(0, 0, 6).Format("2006-01-02")

	huecos := obtener(t, srv, "/api/v1/profesionales/"+p.ID+"/huecos?desde="+desde+"&hasta="+hasta)
	if huecos.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200: el horizonte recorta, no rechaza", huecos.StatusCode)
	}

	var lista respuestaHuecos
	if err := json.NewDecoder(huecos.Body).Decode(&lista); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if lista.Rango.Hasta != ultimoEsperado {
		t.Errorf("rango.hasta = %q, se esperaba %q (siete días contados desde hoy)", lista.Rango.Hasta, ultimoEsperado)
	}
}
