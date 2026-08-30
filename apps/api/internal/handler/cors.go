package handler

import (
	"net/http"
	"slices"
)

// maxEdadPreflight es cuánto puede cachear el browser el resultado de un
// preflight. Diez minutos: suficiente para que una sesión de uso no dispare un
// OPTIONS por request, corto para que cambiar la lista de orígenes se note en
// el día, no en la semana.
const maxEdadPreflight = "600"

// metodosPermitidos y headersPermitidos son los que la API realmente usa. Se
// declaran explícitos y no reflejando lo que el browser pida: reflejar
// convierte al preflight en un sí automático, que es justo lo que no queremos
// que sea.
//
// Content-Type es obligatorio en la lista. Sin él, el browser bloquea
// cualquier cuerpo JSON, y la API exige application/json en todos: el 415
// pasaría a ser imposible de satisfacer.
var (
	metodosPermitidos = "GET, POST, PUT, DELETE, OPTIONS"
	headersPermitidos = "Content-Type, X-Request-ID"
)

// CORS deja que un front en otro origen le hable a la API llevando la cookie
// de sesión.
//
// Tres decisiones que conviene no revertir sin pensarlas:
//
//  1. **Nunca "*".** Con Allow-Credentials el browser rechaza el comodín, así
//     que sería inútil; y sin credenciales, abriría toda la API a cualquier
//     sitio. Se refleja el origen exacto o no se manda nada.
//
//  2. **Un origen no permitido no se rechaza, se ignora.** La petición se
//     atiende y el browser bloquea la lectura por falta del header, que es
//     exactamente para lo que CORS existe. Devolver un 403 sería peor: le
//     confirmaría a quien sondea que la ruta existe, y rompería a los clientes
//     que no son browsers y no mandan Origin.
//
//  3. **Lista vacía apaga el middleware.** Es el estado por defecto en
//     producción hasta que haya un front desplegado: una lista vacía no puede
//     permitir a nadie por accidente, mientras que un valor por defecto
//     "razonable" sí.
//
// CORS no reemplaza a ninguna otra defensa: no protege contra CSRF —de eso se
// ocupan SameSite=Lax y el 415 de decodificarJSON— ni sustituye a
// RequerirSesion.
func CORS(origenes []string) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origen := r.Header.Get("Origin")

			// Sin Origin no es una petición de browser: curl, otro servicio, un
			// health check. No lleva headers de CORS.
			if origen == "" || len(origenes) == 0 {
				siguiente.ServeHTTP(w, r)
				return
			}

			// Vary va aunque el origen no esté permitido: sin esto, un caché
			// intermedio puede servirle a un origen la respuesta que se armó
			// para otro, con o sin permiso adentro.
			w.Header().Add("Vary", "Origin")

			if !slices.Contains(origenes, origen) {
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				siguiente.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origen)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", metodosPermitidos)
				w.Header().Set("Access-Control-Allow-Headers", headersPermitidos)
				w.Header().Set("Access-Control-Max-Age", maxEdadPreflight)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			siguiente.ServeHTTP(w, r)
		})
	}
}
