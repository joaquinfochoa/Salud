package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// nuevoServidorDePrueba cablea el stack real de punta a punta.
func nuevoServidorDePrueba(t *testing.T) *httptest.Server {
	t.Helper()
	repo := memory.NuevoProfesional()
	svc := service.NuevoProfesional(repo)
	srv := httptest.NewServer(NuevoRouter(NuevoProfesional(svc)))
	t.Cleanup(srv.Close)
	return srv
}

const cuerpoValido = `{
  "nombre": "Martín",
  "apellido": "González",
  "matricula": "MN 98.234",
  "especialidad": "psicologia",
  "bio": "Psicólogo clínico.",
  "precioConsultaCentavos": 1200000,
  "modalidades": ["telemedicina", "presencial"],
  "zona": "CABA",
  "obrasSociales": ["OSDE"]
}`

func postear(t *testing.T, srv *httptest.Server, ruta, cuerpo string) *http.Response {
	t.Helper()
	resp, err := srv.Client().Post(srv.URL+ruta, "application/json", strings.NewReader(cuerpo))
	if err != nil {
		t.Fatalf("POST %s falló: %v", ruta, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func obtener(t *testing.T, srv *httptest.Server, ruta string) *http.Response {
	t.Helper()
	resp, err := srv.Client().Get(srv.URL + ruta)
	if err != nil {
		t.Fatalf("GET %s falló: %v", ruta, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func ejecutar(t *testing.T, srv *httptest.Server, metodo, ruta, cuerpo string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(metodo, srv.URL+ruta, strings.NewReader(cuerpo))
	if err != nil {
		t.Fatalf("no se pudo armar el request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s falló: %v", metodo, ruta, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decodificarProfesional(t *testing.T, resp *http.Response) respuestaProfesional {
	t.Helper()
	var p respuestaProfesional
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	return p
}

func TestHealthz(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := obtener(t, srv, "/healthz")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, se esperaba 200", resp.StatusCode)
	}
}

func TestCrearDevuelve201YLocation(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201", resp.StatusCode)
	}

	p := decodificarProfesional(t, resp)
	if p.Slug != "martin-gonzalez" {
		t.Errorf("slug = %q, se esperaba martin-gonzalez", p.Slug)
	}
	if p.Verificacion != "pendiente" {
		t.Errorf("verificacion = %q, se esperaba pendiente", p.Verificacion)
	}

	ubicacion := resp.Header.Get("Location")
	if ubicacion != "/api/v1/profesionales/"+p.ID {
		t.Errorf("Location = %q, se esperaba la URL del recurso creado", ubicacion)
	}
}

func TestCrearJSONMalformadoDevuelve400(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := postear(t, srv, "/api/v1/profesionales", `{"nombre":`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != tipoContenidoProblema {
		t.Errorf("Content-Type = %q, se esperaba %q", ct, tipoContenidoProblema)
	}
}

func TestCrearDatosInvalidosDevuelve422(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	// JSON perfectamente válido, datos mal: es otro problema que un 400
	cuerpo := `{
	  "nombre": "",
	  "apellido": "González",
	  "matricula": "roto",
	  "especialidad": "cardiologia",
	  "precioConsultaCentavos": -5,
	  "modalidades": [],
	  "zona": ""
	}`
	resp := postear(t, srv, "/api/v1/profesionales", cuerpo)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, se esperaba 422", resp.StatusCode)
	}

	var p Problema
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("no se pudo decodificar el problem: %v", err)
	}
	if len(p.Errores) < 5 {
		t.Errorf("se esperaban al menos 5 campos con error, se obtuvieron %d: %+v", len(p.Errores), p.Errores)
	}
}

func TestCrearMatriculaDuplicadaDevuelve409(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	postear(t, srv, "/api/v1/profesionales", cuerpoValido)
	resp := postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, se esperaba 409", resp.StatusCode)
	}
}

func TestObtenerPorIDYPorSlug(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))

	porID := obtener(t, srv, "/api/v1/profesionales/"+creado.ID)
	if porID.StatusCode != http.StatusOK {
		t.Errorf("GET por id: status = %d, se esperaba 200", porID.StatusCode)
	}

	porSlug := obtener(t, srv, "/api/v1/profesionales/por-slug/"+creado.Slug)
	if porSlug.StatusCode != http.StatusOK {
		t.Errorf("GET por slug: status = %d, se esperaba 200", porSlug.StatusCode)
	}
	if decodificarProfesional(t, porSlug).ID != creado.ID {
		t.Error("GET por slug devolvió otro profesional")
	}
}

func TestObtenerIDInexistenteDevuelve404(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	resp := obtener(t, srv, "/api/v1/profesionales/6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, se esperaba 404", resp.StatusCode)
	}
}

func TestObtenerIDMalFormadoDevuelve400(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	// no es un UUID: es un problema del cliente, no un recurso que falta
	resp := obtener(t, srv, "/api/v1/profesionales/no-soy-un-uuid")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestListarPaginaYFiltra(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	segundo := strings.Replace(cuerpoValido, `"MN 98.234"`, `"MN 11111"`, 1)
	segundo = strings.Replace(segundo, `"psicologia"`, `"odontologia"`, 1)
	postear(t, srv, "/api/v1/profesionales", segundo)

	t.Run("sin filtros", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales")
		var listado respuestaListado
		if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if listado.Paginacion.Total != 2 {
			t.Errorf("total = %d, se esperaba 2", listado.Paginacion.Total)
		}
		if listado.Paginacion.Limite != service.LimitePorDefecto {
			t.Errorf("limite = %d, se esperaba el default %d", listado.Paginacion.Limite, service.LimitePorDefecto)
		}
	})

	t.Run("por especialidad", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?especialidad=odontologia")
		var listado respuestaListado
		if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if listado.Paginacion.Total != 1 {
			t.Errorf("total = %d, se esperaba 1", listado.Paginacion.Total)
		}
	})

	t.Run("busqueda sin acentos", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?busqueda=gonzalez")
		var listado respuestaListado
		if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
			t.Fatalf("no se pudo decodificar: %v", err)
		}
		if listado.Paginacion.Total != 2 {
			t.Errorf("total = %d, se esperaban 2: ambos apellidan González", listado.Paginacion.Total)
		}
	})

	t.Run("limite invalido devuelve 400", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?limite=abc")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
		}
	})

	t.Run("especialidad invalida devuelve 400", func(t *testing.T) {
		resp := obtener(t, srv, "/api/v1/profesionales?especialidad=cardiologia")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, se esperaba 400", resp.StatusCode)
		}
	})
}

// TestListarInformaElLimiteRealmenteAplicado es la prueba de la regresión:
// el filtro se pasaba por valor al servicio, así que el recorte a
// LimiteMaximo que hacía el servicio quedaba en su copia y nunca volvía al
// handler. La respuesta terminaba diciendo "limite": 5000 mientras devolvía
// 100 registros, y quien pagina sumando ese número se saltea el resto sin
// que nada falle.
func TestListarInformaElLimiteRealmenteAplicado(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	resp := obtener(t, srv, "/api/v1/profesionales?limite=5000")
	var listado respuestaListado
	if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}

	if listado.Paginacion.Limite > service.LimiteMaximo {
		t.Errorf("paginacion.limite = %d, no puede superar el máximo %d", listado.Paginacion.Limite, service.LimiteMaximo)
	}
	if len(listado.Datos) > listado.Paginacion.Limite {
		t.Errorf("len(datos) = %d, no puede superar el limite informado %d", len(listado.Datos), listado.Paginacion.Limite)
	}
}

func TestListarSinLimiteInformaElDefault(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	postear(t, srv, "/api/v1/profesionales", cuerpoValido)

	resp := obtener(t, srv, "/api/v1/profesionales")
	var listado respuestaListado
	if err := json.NewDecoder(resp.Body).Decode(&listado); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}

	if listado.Paginacion.Limite != service.LimitePorDefecto {
		t.Errorf("paginacion.limite = %d, se esperaba el default %d", listado.Paginacion.Limite, service.LimitePorDefecto)
	}
}

func TestActualizar(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))

	cuerpo := strings.Replace(cuerpoValido, `"CABA"`, `"GBA Norte"`, 1)
	resp := ejecutar(t, srv, http.MethodPut, "/api/v1/profesionales/"+creado.ID, cuerpo)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}
	actualizado := decodificarProfesional(t, resp)
	if actualizado.Zona != "GBA Norte" {
		t.Errorf("zona = %q, se esperaba GBA Norte", actualizado.Zona)
	}
	if actualizado.Slug != creado.Slug {
		t.Error("el slug es una URL pública y no debía cambiar")
	}
}

func TestDeleteEsBajaLogicaYIdempotente(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))

	resp := ejecutar(t, srv, http.MethodDelete, "/api/v1/profesionales/"+creado.ID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, se esperaba 204", resp.StatusCode)
	}

	// el recurso sigue existiendo: no fue un borrado
	despues := obtener(t, srv, "/api/v1/profesionales/"+creado.ID)
	if despues.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: el profesional dado de baja tenía que seguir existiendo", despues.StatusCode)
	}
	if p := decodificarProfesional(t, despues); p.Estado != "inactivo" || p.DadoDeBajaEn == nil {
		t.Error("debía quedar inactivo con dadoDeBajaEn sellado")
	}

	// pero no aparece en el listado por defecto
	respListado := obtener(t, srv, "/api/v1/profesionales")
	var listado respuestaListado
	if err := json.NewDecoder(respListado.Body).Decode(&listado); err != nil {
		t.Fatalf("no se pudo decodificar: %v", err)
	}
	if listado.Paginacion.Total != 0 {
		t.Errorf("total = %d, se esperaba 0", listado.Paginacion.Total)
	}

	// y una segunda baja no es un error
	otraVez := ejecutar(t, srv, http.MethodDelete, "/api/v1/profesionales/"+creado.ID, "")
	if otraVez.StatusCode != http.StatusNoContent {
		t.Errorf("la segunda baja devolvió %d, se esperaba 204", otraVez.StatusCode)
	}
}

func TestReactivar(t *testing.T) {
	srv := nuevoServidorDePrueba(t)
	creado := decodificarProfesional(t, postear(t, srv, "/api/v1/profesionales", cuerpoValido))
	ejecutar(t, srv, http.MethodDelete, "/api/v1/profesionales/"+creado.ID, "")

	resp := postear(t, srv, "/api/v1/profesionales/"+creado.ID+"/reactivar", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	p := decodificarProfesional(t, resp)
	if p.Estado != "activo" || p.DadoDeBajaEn != nil {
		t.Error("debía quedar activo y con dadoDeBajaEn en null")
	}
}
