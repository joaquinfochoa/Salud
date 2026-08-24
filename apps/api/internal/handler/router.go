package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// NuevoRouter arma la tabla de rutas.
//
// Usa el ServeMux de la stdlib: desde Go 1.22 entiende método y parámetros de
// ruta, así que no hace falta chi, gin ni echo para esto.
func NuevoRouter(ph *ManejadorProfesional, ah *ManejadorAgenda) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)

	mux.HandleFunc("GET /api/v1/profesionales", ph.Listar)
	mux.HandleFunc("POST /api/v1/profesionales", ph.Crear)
	mux.HandleFunc("GET /api/v1/profesionales/{id}", ph.ObtenerPorID)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}", ph.Actualizar)
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}", ph.DarDeBaja)
	mux.HandleFunc("POST /api/v1/profesionales/{id}/reactivar", ph.Reactivar)

	// Esta ruta y la de reactivar tienen la misma cantidad de segmentos y la
	// misma forma: `.../profesionales/por-slug/reactivar` encajaría en las dos.
	// Lo que las separa es el método, no la especificidad — por eso todos los
	// patrones acá llevan el verbo adelante. Si alguno lo pierde, el ServeMux
	// entra en pánico al registrar, no al recibir la petición.
	//
	// Esa separación por método no alcanza para por-slug frente a los GET
	// nuevos: por-slug/{slug}, {id}/horarios, {id}/bloqueos y {id}/huecos son
	// los cuatro GET de cinco segmentos, y en cada par el literal nuevo queda
	// en una posición distinta a la de por-slug (uno lo tiene en el cuarto
	// segmento, el otro en el quinto). El ServeMux exige que un patrón domine
	// al otro en todas las posiciones donde difieren, y acá ninguno domina —
	// para "/profesionales/por-slug/horarios" harían falta los dos a la vez.
	// Van los cuatro bajo un único patrón, resuelto a mano.
	mux.HandleFunc("GET /api/v1/profesionales/{a}/{b}", despacharSegundoSegmento(ph, ah))

	mux.HandleFunc("PUT /api/v1/profesionales/{id}/horarios", ah.ReemplazarHorarios)
	mux.HandleFunc("POST /api/v1/profesionales/{id}/bloqueos", ah.CrearBloqueo)
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}/bloqueos/{bloqueoId}", ah.EliminarBloqueo)

	// El orden es de afuera hacia adentro. IDPeticion va primero para que el
	// log lo tenga; RegistrarPeticiones envuelve a RecuperarPanic para que un panic
	// quede registrado con su 500.
	return Encadenar(envolverErroresDeRuteo(mux), IDPeticion, RegistrarPeticiones, RecuperarPanic)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}

// despacharSegundoSegmento resuelve a mano la ambigüedad de especificidad
// entre por-slug/{slug} y los tres GET nuevos: mira el valor real de cada
// segmento y llama al handler que corresponde, reponiendo el nombre de
// parámetro que ese handler espera (id o slug).
//
// Esto sigue siendo ruteo, no una decisión de negocio: ningún handler queda
// al tanto de que existe esta ambigüedad.
func despacharSegundoSegmento(ph *ManejadorProfesional, ah *ManejadorAgenda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		primero, segundo := r.PathValue("a"), r.PathValue("b")

		if primero == "por-slug" {
			r.SetPathValue("slug", segundo)
			ph.ObtenerPorSlug(w, r)
			return
		}

		r.SetPathValue("id", primero)
		switch segundo {
		case "horarios":
			ah.ListarHorarios(w, r)
		case "bloqueos":
			ah.ListarBloqueos(w, r)
		case "huecos":
			ah.HuecosLibres(w, r)
		default:
			http.NotFound(w, r)
		}
	}
}

// envolverErroresDeRuteo convierte a application/problem+json los dos casos
// donde ServeMux responde solo, sin pasar por ningún handler nuestro: ruta sin
// match (404) y método sin match (405). Ambos salen de la stdlib en texto
// plano; todo lo demás en el contrato es problem+json, y un cliente que
// parsea errores de forma uniforme revienta justo en estos dos.
func envolverErroresDeRuteo(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siguiente.ServeHTTP(&interceptorRuteo{ResponseWriter: w}, r)
	})
}

// interceptorRuteo deja pasar cualquier respuesta que ya sea problem+json —
// toda la nuestra— y solo reemplaza el cuerpo cuando detecta el 404 o 405 que
// arma la stdlib antes de llegar a un handler nuestro. Se distingue por el
// Content-Type: el nuestro siempre es problem+json antes de escribir el
// estado; el de la stdlib es texto plano.
type interceptorRuteo struct {
	http.ResponseWriter
	interceptando bool
	codigo        int
	yaEscrito     bool
}

func (w *interceptorRuteo) WriteHeader(codigo int) {
	if (codigo == http.StatusNotFound || codigo == http.StatusMethodNotAllowed) &&
		!strings.HasPrefix(w.Header().Get("Content-Type"), tipoContenidoProblema) {
		w.interceptando = true
		w.codigo = codigo
		return // se decide el cuerpo real recién en Write
	}
	w.ResponseWriter.WriteHeader(codigo)
}

func (w *interceptorRuteo) Write(b []byte) (int, error) {
	if !w.interceptando {
		return w.ResponseWriter.Write(b)
	}
	if w.yaEscrito {
		return len(b), nil // el cuerpo de reemplazo ya salió
	}
	w.yaEscrito = true

	tipo, titulo, detalle := tipoNoEncontrado, "No encontrado", "La ruta o el recurso solicitado no existe"
	if w.codigo == http.StatusMethodNotAllowed {
		tipo, titulo, detalle = tipoMetodoNoPermitido, "Método no permitido", "El método HTTP no está soportado para esta ruta"
	}

	w.ResponseWriter.Header().Set("Content-Type", tipoContenidoProblema)
	w.ResponseWriter.WriteHeader(w.codigo)
	_ = json.NewEncoder(w.ResponseWriter).Encode(Problema{
		Tipo:    tipo,
		Titulo:  titulo,
		Estado:  w.codigo,
		Detalle: detalle,
	})
	return len(b), nil
}
