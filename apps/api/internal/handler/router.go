package handler

import "net/http"

// NuevoRouter arma la tabla de rutas.
//
// Usa el ServeMux de la stdlib: desde Go 1.22 entiende método y parámetros de
// ruta, así que no hace falta chi, gin ni echo para esto.
func NuevoRouter(ph *ManejadorProfesional) http.Handler {
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
	mux.HandleFunc("GET /api/v1/profesionales/por-slug/{slug}", ph.ObtenerPorSlug)

	// El orden es de afuera hacia adentro. IDPeticion va primero para que el
	// log lo tenga; RegistrarPeticiones envuelve a RecuperarPanic para que un panic
	// quede registrado con su 500.
	return Encadenar(mux, IDPeticion, RegistrarPeticiones, RecuperarPanic)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}
