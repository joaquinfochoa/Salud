package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

const maxBytesCuerpo = 1 << 20 // 1 MB

// peticionProfesional es lo que entra. Deliberadamente no incluye id, slug,
// estado, verificacion ni marcas de tiempo: son campos que el servidor decide,
// y aceptarlos sería dejar que el cliente se autoverifique.
type peticionProfesional struct {
	Nombre                 string   `json:"nombre"`
	Apellido               string   `json:"apellido"`
	Matricula              string   `json:"matricula"`
	Especialidad           string   `json:"especialidad"`
	Bio                    string   `json:"bio"`
	PrecioConsultaCentavos int64    `json:"precioConsultaCentavos"`
	Modalidades            []string `json:"modalidades"`
	Zona                   string   `json:"zona"`
	ObrasSociales          []string `json:"obrasSociales"`
}

func (r peticionProfesional) aEntrada() domain.EntradaProfesional {
	return domain.EntradaProfesional{
		Nombre:         r.Nombre,
		Apellido:       r.Apellido,
		Matricula:      r.Matricula,
		Especialidad:   r.Especialidad,
		Bio:            r.Bio,
		PrecioConsulta: r.PrecioConsultaCentavos,
		Modalidades:    r.Modalidades,
		Zona:           r.Zona,
		ObrasSociales:  r.ObrasSociales,
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
