package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

func TestARespuestaSerializaLosCamposDelContrato(t *testing.T) {
	ahora := time.Date(2026, 8, 21, 14, 2, 11, 0, time.UTC)
	precio := int64(1200000)
	p, err := domain.NuevoProfesional(domain.EntradaProfesional{
		Nombre:         "Martín",
		Apellido:       "González",
		Matricula:      "MN 98.234",
		Especialidad:   "psicologia",
		Bio:            "Psicólogo clínico.",
		PrecioConsulta: &precio,
		Modalidades:    []string{"telemedicina", "presencial"},
		Zona:           "CABA",
		ObrasSociales:  []string{"OSDE"},
	}, uuid.New(), ahora)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	crudo, err := json.Marshal(aRespuesta(p))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	cuerpo := string(crudo)

	// el precio viaja como entero, sin punto decimal
	if !strings.Contains(cuerpo, `"precioConsultaCentavos":1200000`) {
		t.Errorf("el precio no viaja como entero de centavos: %s", cuerpo)
	}
	// la matrícula viaja en forma canónica
	if !strings.Contains(cuerpo, `"matricula":"MN 98234"`) {
		t.Errorf("la matrícula no viaja canónica: %s", cuerpo)
	}
	if !strings.Contains(cuerpo, `"dadoDeBajaEn":null`) {
		t.Errorf("dadoDeBajaEn debía estar presente como null: %s", cuerpo)
	}
}

func TestARespuestaNuncaSerializaSlicesComoNull(t *testing.T) {
	// un slice nil se serializa como null y obliga al cliente TypeScript a
	// chequearlo en cada uso
	var p domain.Profesional
	crudo, err := json.Marshal(aRespuesta(p))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	cuerpo := string(crudo)

	if strings.Contains(cuerpo, `"modalidades":null`) {
		t.Error("modalidades se serializó como null en vez de []")
	}
	if strings.Contains(cuerpo, `"obrasSociales":null`) {
		t.Error("obrasSociales se serializó como null en vez de []")
	}
}

func TestARespuestaListadoConListaVacia(t *testing.T) {
	crudo, err := json.Marshal(aRespuestaListado(nil, 0, 20, 0))
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	if strings.Contains(string(crudo), `"datos":null`) {
		t.Errorf("datos se serializó como null en vez de []: %s", crudo)
	}
}

// Sin este chequeo, un formulario HTML de otro sitio puede mandar un POST con
// la cookie de sesión adjunta. Con él, cualquier cuerpo cross-origin muere en
// 415 antes de tocar el servicio: un form solo puede mandar form-urlencoded,
// multipart o text/plain, nunca application/json.
func TestDecodificarJSONExigeContentType(t *testing.T) {
	casos := []struct {
		nombre      string
		contentType string
		esperado    int
	}{
		{"json", "application/json", 0},
		{"json con charset", "application/json; charset=utf-8", 0},
		{"form urlencoded", "application/x-www-form-urlencoded", http.StatusUnsupportedMediaType},
		{"texto plano", "text/plain", http.StatusUnsupportedMediaType},
		{"multipart", "multipart/form-data; boundary=x", http.StatusUnsupportedMediaType},
		{"ausente", "", http.StatusUnsupportedMediaType},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
			if c.contentType != "" {
				req.Header.Set("Content-Type", c.contentType)
			}
			rec := httptest.NewRecorder()

			var destino struct{}
			err := decodificarJSON(rec, req, &destino)

			if c.esperado == 0 {
				if err != nil {
					t.Fatalf("se esperaba éxito, se obtuvo %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("se esperaba error por Content-Type")
			}
			escribirErrorDeDecodificacion(rec, req, err)
			if rec.Code != c.esperado {
				t.Errorf("estado = %d, se esperaba %d", rec.Code, c.esperado)
			}
		})
	}
}
