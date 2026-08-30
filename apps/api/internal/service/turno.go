package service

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// Turno resuelve reservar, cancelar y los dos listados.
//
// Depende de *Agenda y no al revés: para reservar necesita saber qué huecos
// hay, y esa cuenta ya la sabe hacer Agenda. La dirección importa —Agenda
// habla con repository.Turno, nunca con este servicio— porque al revés sería
// un ciclo.
type Turno struct {
	repo          repository.Turno
	profesionales repository.Profesional
	agenda        *Agenda

	// ahora es inyectable para que los casos no dependan del reloj.
	ahora func() time.Time
}

// ConRelojTurno inyecta el reloj. Se llama así y no ConReloj porque ese nombre
// ya lo usa el servicio de agenda en este mismo paquete.
func ConRelojTurno(ahora func() time.Time) func(*Turno) {
	return func(s *Turno) { s.ahora = ahora }
}

func NuevoTurno(repo repository.Turno, profesionales repository.Profesional, agenda *Agenda, opciones ...func(*Turno)) *Turno {
	s := &Turno{
		repo:          repo,
		profesionales: profesionales,
		agenda:        agenda,
		ahora:         func() time.Time { return time.Now().UTC() },
	}
	for _, o := range opciones {
		o(s)
	}
	return s
}

// Reservar toma un hueco a nombre del paciente.
//
// La única validación de calendario es que el inicio pedido coincida con un
// hueco libre. Eso arrastra gratis, sin reimplementar nada: el horario semanal,
// los bloqueos, la anticipación mínima, el horizonte, el profesional inactivo y
// los turnos ya tomados. Todo eso ya lo decide HuecosLibres.
//
// La alternativa —validar cada regla otra vez acá— duplicaría seis reglas en un
// segundo lugar que se desincroniza del primero en el primer cambio.
func (s *Turno) Reservar(ctx context.Context, pacienteID, profesionalID uuid.UUID, inicio time.Time, motivo string) (domain.Turno, error) {
	// Se piden los huecos de ese único día. HuecosLibres valida que el
	// profesional exista, así que uno inexistente sale por acá con
	// ErrNoEncontrado sin que este método sepa nada del asunto.
	dia := domain.InicioDelDia(inicio)
	resultado, err := s.agenda.HuecosLibres(ctx, profesionalID, dia, dia)
	if err != nil {
		return domain.Turno{}, err
	}

	i := slices.IndexFunc(resultado.Huecos, func(h domain.Hueco) bool {
		return h.Inicio.Equal(inicio)
	})
	if i < 0 {
		return domain.Turno{}, domain.ErrorValidacion{Campos: []domain.ErrorCampo{
			{Campo: "inicio", Mensaje: "no corresponde a un turno disponible"},
		}}
	}

	turno, err := domain.NuevoTurno(profesionalID, pacienteID, resultado.Huecos[i], motivo, s.ahora())
	if err != nil {
		return domain.Turno{}, err
	}

	// Lo de arriba es el camino rápido que da el error lindo; entre leer los
	// huecos y escribir acá entra otro paciente. La garantía la da el
	// repositorio bajo el lock de escritura, con ErrHuecoTomado.
	if err := s.repo.Crear(ctx, turno); err != nil {
		return domain.Turno{}, err
	}
	return turno, nil
}

// Cancelar lo pueden hacer las dos partes. Un turno no tiene un dueño sino dos,
// y por eso no alcanza con verificarPropiedad.
func (s *Turno) Cancelar(ctx context.Context, usuarioID, turnoID uuid.UUID) error {
	turno, err := s.repo.ObtenerPorID(ctx, turnoID)
	if err != nil {
		return err
	}
	if err := s.verificarParte(ctx, turno, usuarioID); err != nil {
		return err
	}

	cancelado, err := turno.Cancelar(usuarioID, s.ahora())
	if err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, cancelado)
}

// ListarDePaciente devuelve los turnos del paciente, incluidos los cancelados.
//
// No recibe un pacienteID que venga del cliente: el handler le pasa el de la
// sesión. Aceptarlo como filtro convertiría este endpoint en una forma de leer
// la agenda de cualquiera.
func (s *Turno) ListarDePaciente(ctx context.Context, pacienteID uuid.UUID, desde, hasta *time.Time) ([]domain.Turno, error) {
	d, h := s.ventana(desde, hasta)
	return s.repo.ListarDePaciente(ctx, pacienteID, d, h)
}

// ListarDeProfesional devuelve la agenda ocupada, y solo al dueño del perfil.
//
// Es privado a diferencia de ListarHorarios y ListarBloqueos, que son públicos.
// La diferencia no es caprichosa: los huecos libres son información de oferta,
// pero la agenda ocupada dice quién es paciente de quién, y eso es dato de
// salud bajo Ley 25.326.
func (s *Turno) ListarDeProfesional(ctx context.Context, usuarioID, profesionalID uuid.UUID, desde, hasta *time.Time) ([]domain.Turno, error) {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return nil, err
	}
	if err := verificarPropiedad(profesional, usuarioID); err != nil {
		return nil, err
	}

	d, h := s.ventana(desde, hasta)
	return s.repo.ListarDeProfesional(ctx, profesionalID, d, h)
}

// verificarParte acepta al paciente del turno y al dueño del perfil
// profesional. Cualquier otro recibe ErrTurnoAjeno.
func (s *Turno) verificarParte(ctx context.Context, t domain.Turno, usuarioID uuid.UUID) error {
	if t.PacienteID == usuarioID {
		return nil
	}

	profesional, err := s.profesionales.ObtenerPorID(ctx, t.ProfesionalID)
	if err != nil {
		return err
	}
	if profesional.UsuarioID == usuarioID {
		return nil
	}
	return domain.ErrTurnoAjeno
}

// ventana traduce el rango opcional al mismo default que ListarBloqueos: sin
// rango, los vigentes y futuros. Es una regla de negocio y no de transporte,
// así que vive acá y queda testeable con el reloj fijo.
//
// Quien quiera su historial pasa un desde: por defecto nadie quiere abrir la
// app y ver los turnos del año pasado primero.
func (s *Turno) ventana(desde, hasta *time.Time) (time.Time, time.Time) {
	ahora := s.ahora()
	d, h := ahora, ahora.AddDate(ventanaBloqueosAniosPorDefecto, 0, 0)
	if desde != nil {
		d = *desde
	}
	if hasta != nil {
		h = *hasta
	}
	return d, h
}
