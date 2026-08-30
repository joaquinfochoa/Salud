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
func NuevoRouter(ph *ManejadorProfesional, ah *ManejadorAgenda, mh *ManejadorAutenticacion, th *ManejadorTurno) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)

	// Sesión como recurso, no /login: consistente con el resto del contrato.
	// Ninguna de estas cuatro colisiona con lo existente: /usuarios y
	// /sesiones son prefijos nuevos, y usuarios/yo es un literal sin ninguna
	// plantilla que compita.
	mux.HandleFunc("POST /api/v1/usuarios", mh.Registrar)
	mux.HandleFunc("GET /api/v1/usuarios/yo", RequerirSesion(mh.Yo))
	mux.HandleFunc("POST /api/v1/sesiones", mh.IniciarSesion)
	mux.HandleFunc("DELETE /api/v1/sesiones/actual", mh.CerrarSesion)

	mux.HandleFunc("GET /api/v1/profesionales", ph.Listar)
	mux.HandleFunc("POST /api/v1/profesionales", RequerirSesion(ph.Crear))
	mux.HandleFunc("GET /api/v1/profesionales/{id}", ph.ObtenerPorID)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}", RequerirSesion(ph.Actualizar))
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}", RequerirSesion(ph.DarDeBaja))
	mux.HandleFunc("POST /api/v1/profesionales/{id}/reactivar", RequerirSesion(ph.Reactivar))

	// El perfil público vive en su propio recurso, no en /profesionales. Antes
	// vivía ahí, con un segmento fijo delante del slug, y esos cinco
	// segmentos chocaban de forma irresoluble con los otros tres GET de cinco
	// segmentos ({id}/horarios, {id}/bloqueos, {id}/huecos): en cada par el
	// literal nuevo quedaba en una posición distinta a la de esa ruta vieja
	// (uno en el cuarto segmento, el otro en el quinto), y el ServeMux exige
	// que un patrón domine al otro en todas las posiciones donde difieren —
	// ninguno dominaba, así que ambos patrones matcheaban la misma URL.
	// Con /perfiles/{slug} el recurso tiene cuatro segmentos contra los cinco
	// de profesionales: prefijo distinto, sin ambigüedad posible, y la tabla
	// vuelve a ser solo `mux.HandleFunc` sin despacho a mano.
	mux.HandleFunc("GET /api/v1/perfiles/{slug}", ph.ObtenerPorSlug)

	mux.HandleFunc("GET /api/v1/profesionales/{id}/horarios", ah.ListarHorarios)
	mux.HandleFunc("PUT /api/v1/profesionales/{id}/horarios", RequerirSesion(ah.ReemplazarHorarios))
	mux.HandleFunc("GET /api/v1/profesionales/{id}/bloqueos", ah.ListarBloqueos)
	mux.HandleFunc("POST /api/v1/profesionales/{id}/bloqueos", RequerirSesion(ah.CrearBloqueo))
	mux.HandleFunc("DELETE /api/v1/profesionales/{id}/bloqueos/{bloqueoId}", RequerirSesion(ah.EliminarBloqueo))
	mux.HandleFunc("GET /api/v1/profesionales/{id}/huecos", ah.HuecosLibres)

	// Las cuatro son privadas. GET .../turnos en particular se aparta de
	// /horarios y /bloqueos, que sí son públicos: los huecos libres son
	// información de oferta, pero la agenda ocupada dice quién es paciente de
	// quién, y eso es dato de salud.
	mux.HandleFunc("GET /api/v1/turnos", RequerirSesion(th.MisTurnos))
	mux.HandleFunc("DELETE /api/v1/turnos/{id}", RequerirSesion(th.Cancelar))
	mux.HandleFunc("GET /api/v1/profesionales/{id}/turnos", RequerirSesion(th.DeProfesional))
	mux.HandleFunc("POST /api/v1/profesionales/{id}/turnos", RequerirSesion(th.Reservar))

	// El orden es de afuera hacia adentro. IDPeticion va primero para que el
	// log lo tenga; Autenticar va antes que RegistrarPeticiones para que el log
	// de cada request pueda incluir al usuario; RegistrarPeticiones envuelve a
	// RecuperarPanic para que un panic quede registrado con su 500.
	return Encadenar(envolverErroresDeRuteo(mux),
		IDPeticion, mh.Autenticar, RegistrarPeticiones, RecuperarPanic)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
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
