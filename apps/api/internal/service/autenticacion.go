package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/repository"
)

// Autenticacion resuelve el registro, el login y la vida de las sesiones. No
// sabe nada de HTTP ni de cookies: devuelve el token en claro y el handler
// decide cómo transportarlo.
type Autenticacion struct {
	usuarios repository.Usuario
	sesiones repository.Sesion

	// ahora es inyectable para que los casos no dependan del reloj.
	ahora func() time.Time
}

// ConRelojAuth inyecta el reloj. Se llama así y no ConReloj porque ese nombre
// ya lo usa el servicio de agenda en este mismo paquete.
func ConRelojAuth(ahora func() time.Time) func(*Autenticacion) {
	return func(a *Autenticacion) { a.ahora = ahora }
}

func NuevaAutenticacion(usuarios repository.Usuario, sesiones repository.Sesion, opciones ...func(*Autenticacion)) *Autenticacion {
	a := &Autenticacion{
		usuarios: usuarios,
		sesiones: sesiones,
		ahora:    func() time.Time { return time.Now().UTC() },
	}
	for _, o := range opciones {
		o(a)
	}
	return a
}

// Registrar crea el usuario y le abre una sesión de una. Pedirle que se loguee
// inmediatamente después de registrarse es un paso que no informa nada.
func (a *Autenticacion) Registrar(ctx context.Context, entrada domain.EntradaUsuario) (domain.Usuario, string, error) {
	u, err := domain.NuevoUsuario(entrada, a.ahora())
	if err != nil {
		return domain.Usuario{}, "", err
	}
	if err := a.usuarios.Crear(ctx, u); err != nil {
		return domain.Usuario{}, "", err
	}

	token, err := a.abrirSesion(ctx, u)
	if err != nil {
		return domain.Usuario{}, "", err
	}
	return u, token, nil
}

// ActualizarPerfil cambia los datos personales de la propia cuenta.
//
// Recibe el id de la sesión y no acepta uno por parámetro: no existe forma de
// que esto edite la cuenta de otra persona.
//
// **Las sesiones abiertas siguen valiendo, incluso al cambiar el email.** El
// token identifica al usuario por su id, no por su email, así que cambiarlo no
// invalida nada. Cerrarlas sería lo correcto si esto pudiera cambiar la
// contraseña —una credencial nueva debería expulsar a quien tuviera la vieja—
// pero la contraseña no se toca acá, a propósito.
func (a *Autenticacion) ActualizarPerfil(ctx context.Context, usuarioID uuid.UUID, entrada domain.EntradaPerfil) (domain.Usuario, error) {
	actual, err := a.usuarios.ObtenerPorID(ctx, usuarioID)
	if err != nil {
		return domain.Usuario{}, err
	}

	nuevo, err := actual.ConPerfil(entrada)
	if err != nil {
		return domain.Usuario{}, err
	}
	if err := a.usuarios.Actualizar(ctx, nuevo); err != nil {
		return domain.Usuario{}, err
	}
	return nuevo, nil
}

// IniciarSesion valida las credenciales y abre una sesión.
//
// Todos los caminos de fallo devuelven ErrCredencialesInvalidas. Distinguir
// "ese email no existe" de "esa contraseña está mal" convierte al login en un
// oráculo: probando direcciones se arma el padrón de usuarios sin adivinar una
// sola contraseña.
//
// ponytail: queda un canal lateral por tiempo. Cuando el email no existe se
// vuelve sin llamar a bcrypt, así que la respuesta llega bastante antes que
// cuando sí existe. Cerrarlo es hashear contra un hash de relleno; no se hace
// todavía porque explotarlo requiere muchos intentos medidos, que es
// exactamente lo que frena el rate limiting del login —ya anotado en la spec
// como requisito previo a exponer la API—. Si el rate limiting se posterga,
// esto se construye.
func (a *Autenticacion) IniciarSesion(ctx context.Context, email, contrasena string) (domain.Usuario, string, error) {
	e, err := domain.ParsearEmail(email)
	if err != nil {
		return domain.Usuario{}, "", domain.ErrCredencialesInvalidas
	}

	u, err := a.usuarios.ObtenerPorEmail(ctx, e)
	switch {
	case errors.Is(err, domain.ErrNoEncontrado):
		return domain.Usuario{}, "", domain.ErrCredencialesInvalidas
	case err != nil:
		return domain.Usuario{}, "", err
	}

	if !u.VerificarContrasena(contrasena) {
		return domain.Usuario{}, "", domain.ErrCredencialesInvalidas
	}

	token, err := a.abrirSesion(ctx, u)
	if err != nil {
		return domain.Usuario{}, "", err
	}
	return u, token, nil
}

// CerrarSesion borra la sesión. Es idempotente: reintentar un logout no es un
// error, y un token inventado tampoco.
func (a *Autenticacion) CerrarSesion(ctx context.Context, token string) error {
	return a.sesiones.Eliminar(ctx, domain.HashearToken(token))
}

// ResolverSesion devuelve el usuario detrás de un token, o ErrNoEncontrado si
// el token no existe o la sesión venció.
//
// La sesión vencida se borra al detectarla: es la única limpieza que hay, y
// alcanza porque una sesión a la que nadie vuelve tampoco molesta a nadie.
func (a *Autenticacion) ResolverSesion(ctx context.Context, token string) (domain.Usuario, error) {
	if token == "" {
		return domain.Usuario{}, domain.ErrNoEncontrado
	}

	hash := domain.HashearToken(token)
	s, err := a.sesiones.ObtenerPorTokenHash(ctx, hash)
	if err != nil {
		return domain.Usuario{}, err
	}

	if s.EstaVencida(a.ahora()) {
		if err := a.sesiones.Eliminar(ctx, hash); err != nil {
			return domain.Usuario{}, err
		}
		return domain.Usuario{}, domain.ErrNoEncontrado
	}

	return a.usuarios.ObtenerPorID(ctx, s.UsuarioID)
}

// UsuarioPorID existe para que el handler de "quién soy" no tenga que hablarle
// al repositorio por atrás. El middleware ya resolvió la sesión, pero deja en
// el contexto solo el ID: meter el domain.Usuario entero haría circular el
// hash de la contraseña por cada request sin que nadie lo necesite.
func (a *Autenticacion) UsuarioPorID(ctx context.Context, id uuid.UUID) (domain.Usuario, error) {
	return a.usuarios.ObtenerPorID(ctx, id)
}

func (a *Autenticacion) abrirSesion(ctx context.Context, u domain.Usuario) (string, error) {
	s, token, err := domain.NuevaSesion(u.ID, a.ahora())
	if err != nil {
		return "", err
	}
	if err := a.sesiones.Crear(ctx, s); err != nil {
		return "", err
	}
	return token, nil
}
