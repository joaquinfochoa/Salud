package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const maxBytesCuerpo = 1 << 20 // 1 MB

// peticionProfesional es lo que entra. Deliberadamente no incluye id, slug,
// estado, verificacion ni marcas de tiempo: son campos que el servidor decide,
// y aceptarlos sería dejar que el cliente se autoverifique.
//
// PrecioConsultaCentavos es un puntero: 0 es un precio válido, pero el campo
// ausente no lo es. Con int64 plano las dos formas decodifican igual y un alta
// sin el campo terminaba cobrando gratis sin que el 422 lo avisara.
type peticionProfesional struct {
	Nombre                 string   `json:"nombre"`
	Apellido               string   `json:"apellido"`
	Matricula              string   `json:"matricula"`
	Especialidad           string   `json:"especialidad"`
	Bio                    string   `json:"bio"`
	PrecioConsultaCentavos *int64   `json:"precioConsultaCentavos"`
	Modalidades            []string `json:"modalidades"`
	Zona                   string   `json:"zona"`
	ObrasSociales          []string `json:"obrasSociales"`
	AnticipacionMinimaMin  *int     `json:"anticipacionMinimaMin"`
	HorizonteDias          *int     `json:"horizonteDias"`
}

func (r peticionProfesional) aEntrada() domain.EntradaProfesional {
	return domain.EntradaProfesional{
		Nombre:                r.Nombre,
		Apellido:              r.Apellido,
		Matricula:             r.Matricula,
		Especialidad:          r.Especialidad,
		Bio:                   r.Bio,
		PrecioConsulta:        r.PrecioConsultaCentavos,
		Modalidades:           r.Modalidades,
		Zona:                  r.Zona,
		ObrasSociales:         r.ObrasSociales,
		AnticipacionMinimaMin: r.AnticipacionMinimaMin,
		HorizonteDias:         r.HorizonteDias,
	}
}

type respuestaProfesional struct {
	ID                     string     `json:"id"`
	Slug                   string     `json:"slug"`
	Nombre                 string     `json:"nombre"`
	Apellido               string     `json:"apellido"`
	Matricula              string     `json:"matricula"`
	Especialidad           string     `json:"especialidad"`
	Bio                    string     `json:"bio"`
	PrecioConsultaCentavos int64      `json:"precioConsultaCentavos"`
	Modalidades            []string   `json:"modalidades"`
	Zona                   string     `json:"zona"`
	ObrasSociales          []string   `json:"obrasSociales"`
	Estado                 string     `json:"estado"`
	Verificacion           string     `json:"verificacion"`
	CreadoEn               time.Time  `json:"creadoEn"`
	ActualizadoEn          time.Time  `json:"actualizadoEn"`
	DadoDeBajaEn           *time.Time `json:"dadoDeBajaEn"`
	AnticipacionMinimaMin  int        `json:"anticipacionMinimaMin"`
	HorizonteDias          int        `json:"horizonteDias"`
}

func aRespuesta(p domain.Profesional) respuestaProfesional {
	// make con len 0 en vez de nil: un slice nil se serializa como null y el
	// cliente TypeScript tendría que chequearlo en cada uso
	mods := make([]string, 0, len(p.Modalidades))
	for _, m := range p.Modalidades {
		mods = append(mods, string(m))
	}

	obras := make([]string, 0, len(p.ObrasSociales))
	obras = append(obras, p.ObrasSociales...)

	return respuestaProfesional{
		ID:                     p.ID.String(),
		Slug:                   p.Slug,
		Nombre:                 p.Nombre,
		Apellido:               p.Apellido,
		Matricula:              p.Matricula.String(),
		Especialidad:           string(p.Especialidad),
		Bio:                    p.Bio,
		PrecioConsultaCentavos: int64(p.PrecioConsulta),
		Modalidades:            mods,
		Zona:                   p.Zona,
		ObrasSociales:          obras,
		Estado:                 string(p.Estado),
		Verificacion:           string(p.Verificacion),
		CreadoEn:               p.CreadoEn,
		ActualizadoEn:          p.ActualizadoEn,
		DadoDeBajaEn:           p.DadoDeBajaEn,
		AnticipacionMinimaMin:  p.AnticipacionMinimaMin,
		HorizonteDias:          p.HorizonteDias,
	}
}

type respuestaPaginacion struct {
	Total          int `json:"total"`
	Limite         int `json:"limite"`
	Desplazamiento int `json:"desplazamiento"`
}

type respuestaListado struct {
	Datos      []respuestaProfesional `json:"datos"`
	Paginacion respuestaPaginacion    `json:"paginacion"`
}

func aRespuestaListado(ps []domain.Profesional, total, limite, desplazamiento int) respuestaListado {
	datos := make([]respuestaProfesional, 0, len(ps))
	for _, p := range ps {
		datos = append(datos, aRespuesta(p))
	}
	return respuestaListado{
		Datos:      datos,
		Paginacion: respuestaPaginacion{Total: total, Limite: limite, Desplazamiento: desplazamiento},
	}
}

// decodificarJSON lee el cuerpo en modo estricto.
//
// DisallowUnknownFields atrapa el typo más probable de esta API: mandar
// "precioConsulta" en vez de "precioConsultaCentavos" y que el precio quede en cero
// sin que nadie se entere.
func decodificarJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	if err := verificarContentTypeJSON(r); err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytesCuerpo)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(destino); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("el cuerpo debe contener un único objeto JSON")
	}
	return nil
}

// errTipoDeContenido lo distingue escribirErrorDeDecodificacion para responder
// 415 en vez de 400.
var errTipoDeContenido = errors.New("el Content-Type debe ser application/json")

// verificarContentTypeJSON es media defensa contra CSRF, y la otra media es
// SameSite=Lax en la cookie. Un formulario de otro sitio solo puede mandar
// form-urlencoded, multipart o text/plain: ninguno pasa de acá, así que no
// puede forjar una escritura aunque el browser adjunte la cookie de sesión.
//
// Se parsea en vez de comparar la cadena entera porque
// "application/json; charset=utf-8" es legítimo y un cliente real lo manda.
func verificarContentTypeJSON(r *http.Request) error {
	tipo, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || tipo != "application/json" {
		return errTipoDeContenido
	}
	return nil
}

// escribirErrorDeDecodificacion traduce un error de decodificarJSON al
// problema HTTP que corresponde. El error de Go puede nombrar el struct
// interno y el tipo real del campo (p. ej. "json: cannot unmarshal string
// into Go struct field peticionProfesional.precioConsultaCentavos of type
// int64"): eso va al log, nunca al cliente. El nombre del campo del contrato
// (Field, sin el prefijo del struct) sí es información del cliente y se
// conserva porque hace el mensaje accionable.
func escribirErrorDeDecodificacion(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "cuerpo JSON inválido",
		"error", err,
		"metodo", r.Method,
		"ruta", r.URL.Path,
	)

	if errors.Is(err, errTipoDeContenido) {
		escribirProblema(w, Problema{
			Tipo:    tipoTipoDeContenido,
			Titulo:  "Tipo de contenido no soportado",
			Estado:  http.StatusUnsupportedMediaType,
			Detalle: "el cuerpo tiene que ser application/json",
		})
		return
	}

	var errTamaño *http.MaxBytesError
	if errors.As(err, &errTamaño) {
		escribirProblema(w, Problema{
			Tipo:    tipoCuerpoDemasiadoGrande,
			Titulo:  "Cuerpo demasiado grande",
			Estado:  http.StatusRequestEntityTooLarge,
			Detalle: "el cuerpo no puede superar los " + strconv.Itoa(maxBytesCuerpo) + " bytes",
		})
		return
	}

	detalle := "el cuerpo no es un JSON válido"
	var errTipo *json.UnmarshalTypeError
	if errors.As(err, &errTipo) && errTipo.Field != "" {
		detalle += ": el campo " + errTipo.Field + " tiene un tipo inválido"
	}
	escribirPeticionInvalida(w, detalle)
}
