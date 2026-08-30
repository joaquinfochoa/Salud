package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/config"
	"github.com/joaquinfochoa/Salud/apps/api/internal/handler"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository/memory"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

func main() {
	if err := ejecutar(); err != nil {
		fmt.Fprintln(os.Stderr, "error fatal:", err)
		os.Exit(1)
	}
}

func ejecutar() error {
	cfg, err := config.Cargar()
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.NivelLog,
	})))

	// El cableado de dependencias, explícito y de arriba abajo. Sin
	// anotaciones ni contenedor: no hay magia que debuggear a las 3 de la
	// mañana.
	//
	// Migrar a PostgreSQL es cambiar esta línea por
	// postgres.NuevoProfesional(db). Nada más.
	repo := memory.NuevoProfesional()
	repoHorarios := memory.NuevoHorarioSemanal()
	repoBloqueos := memory.NuevoBloqueo()
	repoUsuarios := memory.NuevoUsuario()
	repoSesiones := memory.NuevaSesion()
	repoTurnos := memory.NuevoTurno()

	svc := service.NuevoProfesional(repo)
	svcAgenda := service.NuevaAgenda(repo, repoHorarios, repoBloqueos)
	svcAuth := service.NuevaAutenticacion(repoUsuarios, repoSesiones)
	svcTurnos := service.NuevoTurno(repoTurnos, repo, svcAgenda)

	// El seed va después del servicio y escribe a través de él: así queda
	// sujeto a las mismas reglas que cualquier alta, y esta función sigue
	// siendo el único lugar que sabe qué repositorio está en juego.
	//
	// Los repositorios de agenda no se siembran: un profesional sembrado
	// arranca sin horarios, y cargarlos es parte de probar la API.
	if cfg.EsDesarrollo() {
		if err := sembrar(context.Background(), svcAuth, svc); err != nil {
			return fmt.Errorf("cargando el seed: %w", err)
		}
		slog.Info("seed de desarrollo cargado")
	}

	router := handler.NuevoRouter(
		handler.NuevoProfesional(svc),
		handler.NuevaAgenda(svcAgenda),
		// La cookie va con Secure salvo en desarrollo: sin TLS el browser la
		// descarta y el login "no anda" sin ningún error visible.
		handler.NuevaAutenticacion(svcAuth, svc, !cfg.EsDesarrollo()),
		handler.NuevoTurno(svcTurnos),
	)

	srv := &http.Server{
		Addr:    ":" + cfg.Puerto,
		Handler: router,
		// Los valores por defecto de http.Server son cero, o sea sin límite:
		// una conexión lenta puede quedarse tomada para siempre.
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	errServidor := make(chan error, 1)
	go func() {
		slog.Info("servidor escuchando", "direccion", srv.Addr, "entorno", cfg.Entorno)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errServidor <- err
		}
	}()

	select {
	case err := <-errServidor:
		return fmt.Errorf("el servidor falló: %w", err)

	case <-ctx.Done():
		// Apagado gracioso: sin esto, cada deploy corta los requests que
		// están a mitad de camino.
		slog.Info("apagando", "timeout", cfg.TimeoutApagado)

		ctxApagado, cancelar := context.WithTimeout(context.Background(), cfg.TimeoutApagado)
		defer cancelar()

		if err := srv.Shutdown(ctxApagado); err != nil {
			return fmt.Errorf("apagado forzado: %w", err)
		}
		slog.Info("apagado limpio")
		return nil
	}
}
