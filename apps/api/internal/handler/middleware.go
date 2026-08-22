package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const headerIDPeticion = "X-Request-ID"

type claveContexto string

const claveIDPeticion claveContexto = "requestID"

// Encadenar envuelve el handler. El primer middleware de la lista es el más
// externo: el primero en ver el request y el último en ver la respuesta.
func Encadenar(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// IDPeticion asegura que todo request tenga un identificador. Si el cliente
// manda uno, se respeta: permite seguir una operación a través de varios
// servicios en los logs.
func IDPeticion(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerIDPeticion)
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set(headerIDPeticion, id)
		siguiente.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claveIDPeticion, id)))
	})
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
