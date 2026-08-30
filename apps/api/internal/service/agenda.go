package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// Agenda resuelve los casos de uso que necesitan mirar más de una entidad: el
// horario, los bloqueos y el profesional dueño de los dos.
type Agenda struct {
	profesionales repository.Profesional
	horarios      repository.HorarioSemanal
	bloqueos      repository.Bloqueo

	ahora func() time.Time
}

// ventanaBloqueosAniosPorDefecto es cuánto hacia adelante se listan los
// bloqueos vigentes y futuros cuando el cliente no pide un rango explícito.
//
// El dominio no le pone techo a cuán lejos se puede cargar un bloqueo (a
// diferencia del horizonte de huecos, que sí lo tiene), pero "sin límite" en
// la práctica es una ventana con un largo defendible, no "para siempre": dos
// años cubre cualquier agenda médica real (vacaciones, licencias, cierres
// programados) sin arrastrar el default a una fecha arbitrariamente lejana.
const ventanaBloqueosAniosPorDefecto = 2

// ConReloj fija el reloj que usa el servicio en vez del real. Sirve para que
// los tests —de servicio o de handler— no dependan de time.Now ni tengan que
// calcular fechas relativas al día en que corren.
func ConReloj(ahora func() time.Time) func(*Agenda) {
	return func(a *Agenda) { a.ahora = ahora }
}

func NuevaAgenda(profesionales repository.Profesional, horarios repository.HorarioSemanal, bloqueos repository.Bloqueo, opciones ...func(*Agenda)) *Agenda {
	a := &Agenda{
		profesionales: profesionales,
		horarios:      horarios,
		bloqueos:      bloqueos,
		ahora:         func() time.Time { return time.Now().In(domain.ZonaHoraria) },
	}
	for _, opcion := range opciones {
		opcion(a)
	}
	return a
}

// ResultadoHuecos lleva los huecos y el rango que de verdad se usó, que puede
// ser menor al pedido si el horizonte del profesional lo recortó.
type ResultadoHuecos struct {
	Huecos []domain.Hueco
	Desde  time.Time
	Hasta  time.Time
}

func (s *Agenda) ReemplazarHorarios(ctx context.Context, usuarioID, profesionalID uuid.UUID, entradas []domain.EntradaHorarioSemanal) ([]domain.HorarioSemanal, error) {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return nil, err
	}
	if err := verificarPropiedad(profesional, usuarioID); err != nil {
		return nil, err
	}

	semana, err := domain.NuevaSemana(profesionalID, entradas)
	if err != nil {
		return nil, err
	}

	if err := verificarModalidades(semana, profesional); err != nil {
		return nil, err
	}

	if err := s.horarios.ReemplazarDeProfesional(ctx, profesionalID, semana); err != nil {
		return nil, err
	}
	return semana, nil
}

func (s *Agenda) ListarHorarios(ctx context.Context, profesionalID uuid.UUID) ([]domain.HorarioSemanal, error) {
	if _, err := s.profesionales.ObtenerPorID(ctx, profesionalID); err != nil {
		return nil, err
	}
	return s.horarios.ListarDeProfesional(ctx, profesionalID)
}

func (s *Agenda) CrearBloqueo(ctx context.Context, usuarioID, profesionalID uuid.UUID, entrada domain.EntradaBloqueo) (domain.Bloqueo, error) {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return domain.Bloqueo{}, err
	}
	if err := verificarPropiedad(profesional, usuarioID); err != nil {
		return domain.Bloqueo{}, err
	}

	bloqueo, err := domain.NuevoBloqueo(profesionalID, entrada, s.ahora())
	if err != nil {
		return domain.Bloqueo{}, err
	}

	if err := s.bloqueos.Crear(ctx, bloqueo); err != nil {
		return domain.Bloqueo{}, err
	}
	return bloqueo, nil
}

// ListarBloqueos acepta desde y hasta opcionales: cuando el cliente no pide un
// rango (nil), se listan los vigentes y futuros, una ventana que resuelve
// s.ahora() y no la capa HTTP — es una regla de negocio, no de transporte, y
// así queda testeable con el reloj fijo.
func (s *Agenda) ListarBloqueos(ctx context.Context, profesionalID uuid.UUID, desde, hasta *time.Time) ([]domain.Bloqueo, error) {
	if _, err := s.profesionales.ObtenerPorID(ctx, profesionalID); err != nil {
		return nil, err
	}

	ahora := s.ahora()
	d, h := ahora, ahora.AddDate(ventanaBloqueosAniosPorDefecto, 0, 0)
	if desde != nil {
		d = *desde
	}
	if hasta != nil {
		h = *hasta
	}

	return s.bloqueos.ListarDeProfesional(ctx, profesionalID, d, h)
}

// EliminarBloqueo exige que el bloqueo sea de ese profesional.
//
// Sin autenticación cualquiera puede llamar a esto, pero al menos la ruta y el
// recurso tienen que ser coherentes: borrar el bloqueo de otro desde la ruta de
// este es un 404, no un éxito silencioso.
func (s *Agenda) EliminarBloqueo(ctx context.Context, usuarioID, profesionalID, bloqueoID uuid.UUID) error {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return err
	}
	if err := verificarPropiedad(profesional, usuarioID); err != nil {
		return err
	}

	bloqueo, err := s.bloqueos.ObtenerPorID(ctx, bloqueoID)
	if err != nil {
		return err
	}
	if bloqueo.ProfesionalID != profesionalID {
		return domain.ErrNoEncontrado
	}

	return s.bloqueos.Eliminar(ctx, bloqueoID)
}

// HuecosLibres calcula los turnos reservables de un profesional en un rango.
//
// desde y hasta son fechas y las dos entran: pedir del 25 al 27 incluye los
// tres días.
func (s *Agenda) HuecosLibres(ctx context.Context, profesionalID uuid.UUID, desde, hasta time.Time) (ResultadoHuecos, error) {
	profesional, err := s.profesionales.ObtenerPorID(ctx, profesionalID)
	if err != nil {
		return ResultadoHuecos{}, err
	}

	desde = domain.InicioDelDia(desde)
	hasta = domain.InicioDelDia(hasta)

	// Un hueco en el pasado nunca es reservable, así que pedir desde 1900 no
	// puede producir nada — pero sí puede recorrer 700.000 días. El horizonte
	// acota hacia adelante; esto acota hacia atrás.
	if hoy := domain.InicioDelDia(s.ahora()); desde.Before(hoy) {
		desde = hoy
	}

	// El handler ya rechaza con 400 un rango invertido tal como lo mandó el
	// cliente (ver ManejadorAgenda.HuecosLibres): esto es la defensa que
	// sostiene la invariante para cualquier otro llamador del servicio, y
	// también cubre el caso en que el recorte de arriba deja desde por delante
	// de un hasta que era válido antes de recortar.
	if hasta.Before(desde) {
		return ResultadoHuecos{}, domain.ErrorValidacion{Campos: []domain.ErrorCampo{
			{Campo: "hasta", Mensaje: "tiene que ser posterior o igual a desde"},
		}}
	}

	// El horizonte se cuenta desde hoy, no desde la fecha pedida: es cuánto de
	// su agenda el profesional expone hacia adelante. Contarlo desde `desde`
	// dejaría que un cliente pidiera septiembre de 2099 y reservara turnos a
	// tres años vista.
	//
	// Recorta, no rechaza, igual que paginacion.limite en el listado de
	// profesionales, y el resultado informa el rango que de verdad se usó.
	ultimoDia := domain.InicioDelDia(s.ahora()).AddDate(0, 0, profesional.HorizonteDias-1)
	if hasta.After(ultimoDia) {
		hasta = ultimoDia
	}

	// Si el rango entero cae más allá del horizonte no queda nada que
	// calcular, pero el recurso existe: es una lista vacía, no un error.
	if desde.After(ultimoDia) {
		return ResultadoHuecos{Huecos: []domain.Hueco{}, Desde: desde, Hasta: desde}, nil
	}

	resultado := ResultadoHuecos{Huecos: []domain.Hueco{}, Desde: desde, Hasta: hasta}

	// Un profesional dado de baja no tiene disponibilidad, sin importar lo que
	// digan sus reglas. No es un error: el recurso existe, simplemente no opera.
	if profesional.Estado != domain.EstadoActivo {
		return resultado, nil
	}

	horarios, err := s.horarios.ListarDeProfesional(ctx, profesionalID)
	if err != nil {
		return ResultadoHuecos{}, err
	}

	// La modalidad se valida al guardar el horario, pero el profesional puede
	// cambiar sus modalidades después por otro camino (service.Profesional.
	// Actualizar, que no conoce el repositorio de horarios). Filtrar en la
	// lectura es lo único que sobrevive a cualquier orden de escritura.
	horarios = slices.DeleteFunc(horarios, func(h domain.HorarioSemanal) bool {
		return !slices.Contains(profesional.Modalidades, h.Modalidad)
	})

	// el cálculo trabaja con el intervalo semiabierto [desde, finExclusivo)
	finExclusivo := hasta.AddDate(0, 0, 1)

	bloqueos, err := s.bloqueos.ListarDeProfesional(ctx, profesionalID, desde, finExclusivo)
	if err != nil {
		return ResultadoHuecos{}, err
	}

	resultado.Huecos = domain.CalculoHuecos{
		Horarios:              horarios,
		Bloqueos:              bloqueos,
		Desde:                 desde,
		Hasta:                 finExclusivo,
		AnticipacionMinimaMin: profesional.AnticipacionMinimaMin,
		Ahora:                 s.ahora(),
	}.Generar()

	return resultado, nil
}

// verificarModalidades comprueba que cada bloque use una modalidad que el
// profesional declara ofrecer.
//
// Es una regla entre dos entidades y por eso no puede vivir en el dominio: un
// bloque solo no sabe qué ofrece su profesional. Cargar un bloque presencial en
// un perfil que solo hace telemedicina produce huecos que el paciente ve y no
// puede usar.
//
// Identifica el bloque por su día y hora y no por su índice: NuevaSemana
// devuelve la semana ordenada, así que el índice ya no coincide con el orden en
// que el cliente los mandó.
func verificarModalidades(semana []domain.HorarioSemanal, profesional domain.Profesional) error {
	var campos []domain.ErrorCampo

	for _, bloque := range semana {
		if slices.Contains(profesional.Modalidades, bloque.Modalidad) {
			continue
		}
		campos = append(campos, domain.ErrorCampo{
			Campo: "horarios",
			Mensaje: fmt.Sprintf("el bloque de %s %s usa la modalidad %q, que el profesional no ofrece (ofrece: %s)",
				bloque.DiaSemana, bloque.Desde, bloque.Modalidad, listarModalidades(profesional.Modalidades)),
		})
	}

	if len(campos) > 0 {
		return domain.ErrorValidacion{Campos: campos}
	}
	return nil
}

func listarModalidades(modalidades []domain.Modalidad) string {
	nombres := make([]string, 0, len(modalidades))
	for _, m := range modalidades {
		nombres = append(nombres, string(m))
	}
	return strings.Join(nombres, ", ")
}
