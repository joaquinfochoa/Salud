package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const headerIDPeticion = "X-Request-ID"

// maxLargoIDPeticion acota lo que se acepta del cliente. Sin este límite, un
// cliente puede adjuntar hasta 1 MB de basura (el tope de maxBytesCuerpo no
// aplica a los headers) que termina en cada línea de log de ese request.
const maxLargoIDPeticion = 128

type claveContexto string

const claveIDPeticion claveContexto = "requestID"

const claveUsuarioID claveContexto = "usuarioID"

// UsuarioIDDe devuelve el usuario autenticado. El segundo valor dice si había
// sesión, y distinguirlo del uuid.Nil importa: uuid.Nil comparado contra un
// UsuarioID real da "no autorizado" y no "no autenticado", que son dos códigos
// HTTP distintos y dos mensajes distintos para el que llama.
func UsuarioIDDe(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(claveUsuarioID).(uuid.UUID)
	return id, ok
}

// RequerirSesion corta con 401 si el middleware Autenticar no dejó un usuario.
//
// Se aplica por ruta y no globalmente: la mayoría del contrato es pública, y
// una lista de excepciones se desactualiza en silencio la primera vez que
// alguien agrega un endpoint. Envolver el handler es visible en la tabla de
// rutas.
func RequerirSesion(siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UsuarioIDDe(r.Context()); !ok {
			escribirNoAutenticado(w)
			return
		}
		siguiente(w, r)
	}
}

// usuarioAutenticado es lo que usa cada handler privado. Existe para no repetir
// el mismo bloque ocho veces: ocho copias son ocho lugares donde escribir mal
// el código de estado.
//
// Con la ruta envuelta en RequerirSesion el segundo valor siempre es true. El
// chequeo queda igual porque si alguien saca el RequerirSesion de la tabla de
// rutas, esto responde 401 en vez de tratar uuid.Nil como un usuario legítimo.
func usuarioAutenticado(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, ok := UsuarioIDDe(r.Context())
	if !ok {
		escribirNoAutenticado(w)
	}
	return id, ok
}

// Encadenar envuelve el handler. El primer middleware de la lista es el más
// externo: el primero en ver el request y el último en ver la respuesta.
func Encadenar(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// IDPeticion asegura que todo request tenga un identificador. Si el cliente
// manda uno con forma razonable, se respeta: permite seguir una operación a
// través de varios servicios en los logs.
//
// No autenticamos al cliente en esta etapa, así que el valor que manda no es
// de fiar: Go ya neutraliza CRLF al escribir el header de respuesta y slog
// escapa el JSON al loguearlo, pero nada impide que sea gigante (hasta 1 MB,
// el límite de maxBytesCuerpo no cubre headers) o que sea el mismo ID que otro
// request está usando, lo que arruina la correlación en los logs sin que se
// note. Por eso se valida forma, no solo presencia.
func IDPeticion(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerIDPeticion)
		if !idPeticionValida(id) {
			id = uuid.NewString()
		}

		w.Header().Set(headerIDPeticion, id)
		siguiente.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claveIDPeticion, id)))
	})
}

// idPeticionValida acepta el ID del cliente solo si tiene una forma
// razonable: no vacío, acotado en longitud y compuesto por ASCII imprimible.
// Cualquier otra cosa se descarta a favor de un UUID propio.
func idPeticionValida(id string) bool {
	if id == "" || len(id) > maxLargoIDPeticion {
		return false
	}
	for i := 0; i < len(id); i++ {
		if id[i] < 0x20 || id[i] > 0x7e {
			return false
		}
	}
	return true
}

// IDPeticionDe devuelve el identificador del request, o cadena vacía si el
// middleware no corrió.
func IDPeticionDe(ctx context.Context) string {
	id, _ := ctx.Value(claveIDPeticion).(string)
	return id
}

// grabadorEstado captura el código de estado para poder loguearlo: el
// http.ResponseWriter no lo expone.
type grabadorEstado struct {
	http.ResponseWriter
	estado int
	bytes  int
}

func (w *grabadorEstado) WriteHeader(codigo int) {
	w.estado = codigo
	w.ResponseWriter.WriteHeader(codigo)
}

func (w *grabadorEstado) Write(b []byte) (int, error) {
	if w.estado == 0 {
		w.estado = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func RegistrarPeticiones(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rec := &grabadorEstado{ResponseWriter: w, estado: http.StatusOK}

		siguiente.ServeHTTP(rec, r)

		slog.InfoContext(r.Context(), "peticion",
			"idPeticion", IDPeticionDe(r.Context()),
			"metodo", r.Method,
			"ruta", r.URL.Path,
			"estado", rec.estado,
			"bytes", rec.bytes,
			"duracionMs", time.Since(inicio).Milliseconds(),
		)
	})
}

// RecuperarPanic evita que un panic en un handler tire el proceso entero.
//
// Va por dentro de RegistrarPeticiones para que el log registre el 500 resultante.
func RecuperarPanic(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// el detalle del panic va al log, nunca al cliente
			slog.ErrorContext(r.Context(), "panic recuperado",
				"idPeticion", IDPeticionDe(r.Context()),
				"panic", rec,
				"metodo", r.Method,
				"ruta", r.URL.Path,
			)

			escribirProblema(w, Problema{
				Tipo:    tipoInterno,
				Titulo:  "Error interno",
				Estado:  http.StatusInternalServerError,
				Detalle: "Ocurrió un error inesperado. Volvé a intentar.",
			})
		}()

		siguiente.ServeHTTP(w, r)
	})
}
