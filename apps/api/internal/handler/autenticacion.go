package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

const nombreCookieSesion = "sesion"

// ManejadorAutenticacion traduce entre HTTP y el servicio de autenticación, y
// es el dueño de la cookie: es el único archivo que sabe que la sesión viaja
// ahí. Cambiar a un header Authorization sería tocar solo este archivo.
type ManejadorAutenticacion struct {
	svc           *service.Autenticacion
	profesionales *service.Profesional
	cookieSegura  bool
}

// NuevaAutenticacion recibe cookieSegura en vez de decidirlo solo.
//
// El flag Secure impide que la cookie viaje por HTTP sin TLS, que es lo que se
// quiere en producción. En desarrollo contra http://localhost los browsers la
// aceptan igual —localhost es un origen de confianza— pero cualquier otra
// dirección de desarrollo, como una IP de la LAN para probar desde el
// teléfono, la descarta en silencio y el login "no anda" sin ningún error
// visible.
func NuevaAutenticacion(svc *service.Autenticacion, profesionales *service.Profesional, cookieSegura bool) *ManejadorAutenticacion {
	return &ManejadorAutenticacion{svc: svc, profesionales: profesionales, cookieSegura: cookieSegura}
}

// Autenticar resuelve la cookie y deja el UsuarioID en el contexto.
//
// No rechaza nada: un request sin cookie o con una cookie vencida sigue de
// largo sin usuario. Rechazar es trabajo de RequerirSesion, ruta por ruta,
// porque la mayor parte del contrato es pública y un middleware global que
// corte tendría que llevar una lista de excepciones que se desactualiza sola.
func (h *ManejadorAutenticacion) Autenticar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(nombreCookieSesion)
		if err != nil {
			siguiente.ServeHTTP(w, r)
			return
		}

		u, err := h.svc.ResolverSesion(r.Context(), cookie.Value)
		if err != nil {
			// Token inválido o sesión vencida: se sigue como anónimo y se borra
			// la cookie muerta para que el browser deje de mandarla en cada
			// request por los próximos siete días.
			h.borrarCookie(w)
			siguiente.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), claveUsuarioID, u.ID)
		siguiente.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *ManejadorAutenticacion) Registrar(w http.ResponseWriter, r *http.Request) {
	var req peticionRegistro
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	u, token, err := h.svc.Registrar(r.Context(), req.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}

	// Registrarse loguea: pedirle al usuario que inmediatamente después mande
	// sus credenciales otra vez es un paso que no informa nada.
	h.ponerCookie(w, token)
	escribirJSON(w, http.StatusCreated, aRespuestaUsuario(u))
}

func (h *ManejadorAutenticacion) IniciarSesion(w http.ResponseWriter, r *http.Request) {
	var req peticionLogin
	if err := decodificarJSON(w, r, &req); err != nil {
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	u, token, err := h.svc.IniciarSesion(r.Context(), req.Email, req.Contrasena)
	if err != nil {
		escribirError(w, r, err)
		return
	}

	h.ponerCookie(w, token)
	escribirJSON(w, http.StatusCreated, aRespuestaUsuario(u))
}

// CerrarSesion responde 204 siempre, incluso sin cookie. El logout es
// idempotente: un cliente que reintenta no está haciendo nada malo, y decirle
// "no estabas logueado" no le sirve para nada.
func (h *ManejadorAutenticacion) CerrarSesion(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(nombreCookieSesion); err == nil {
		if err := h.svc.CerrarSesion(r.Context(), cookie.Value); err != nil {
			escribirError(w, r, err)
			return
		}
	}

	h.borrarCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Yo devuelve quién es el usuario de la sesión y si tiene perfil profesional.
// ActualizarPerfil cambia los datos personales de la propia cuenta.
//
// No recibe un id: el usuario sale de la sesión, así que no hay forma de que
// esto edite la cuenta de otra persona ni siquiera con la petición armada a
// mano.
func (h *ManejadorAutenticacion) ActualizarPerfil(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	var p peticionPerfil
	if err := decodificarJSON(w, r, &p); err != nil {
		// Y no escribirError: un cuerpo mal formado o con un campo de más es
		// culpa de quien lo mandó, no del servidor. Con escribirError esto
		// devolvía 500, que además de estar mal esconde el motivo.
		escribirErrorDeDecodificacion(w, r, err)
		return
	}

	u, err := h.svc.ActualizarPerfil(r.Context(), usuarioID, p.aEntrada())
	if err != nil {
		escribirError(w, r, err)
		return
	}
	escribirJSON(w, http.StatusOK, aRespuestaUsuario(u))
}

func (h *ManejadorAutenticacion) Yo(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	u, err := h.svc.UsuarioPorID(r.Context(), usuarioID)
	if err != nil {
		escribirError(w, r, err)
		return
	}

	respuesta := respuestaUsuarioActual{respuestaUsuario: aRespuestaUsuario(u)}

	// No tener perfil profesional no es un error: es la mitad de los usuarios.
	// Solo un fallo real del repositorio sale por acá.
	switch p, err := h.profesionales.ObtenerPorUsuarioID(r.Context(), usuarioID); {
	case err == nil:
		id := p.ID.String()
		respuesta.PerfilProfesionalID = &id
	case !errors.Is(err, domain.ErrNoEncontrado):
		escribirError(w, r, err)
		return
	}

	escribirJSON(w, http.StatusOK, respuesta)
}

func (h *ManejadorAutenticacion) ponerCookie(w http.ResponseWriter, token string) {
	// gosec G124 exige Secure como literal true y acá sale de la config: en
	// desarrollo se sirve por HTTP y una cookie Secure no volvería nunca. El
	// atributo está puesto; lo que el linter no puede probar es su valor.
	//nolint:gosec // Secure viene de cfg.EsDesarrollo(); ver NuevaAutenticacion
	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookieSesion,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().UTC().Add(domain.DuracionSesion),
		HttpOnly: true, // invisible para JS: un XSS no se lleva la sesión
		Secure:   h.cookieSegura,
		SameSite: http.SameSiteLaxMode, // media defensa contra CSRF; la otra media es el 415
	})
}

func (h *ManejadorAutenticacion) borrarCookie(w http.ResponseWriter) {
	//nolint:gosec // mismo caso que ponerCookie
	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookieSesion,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSegura,
		SameSite: http.SameSiteLaxMode,
	})
}
