package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// DuracionSesion es cuánto vale una sesión desde que se crea. Es absoluta: no
// se renueva con el uso. La renovación deslizante es más cómoda y se agrega
// después sin cambiar ninguna firma; empezar sin ella es menos código y un
// techo de exposición conocido.
const DuracionSesion = 7 * 24 * time.Hour

// bytesToken son 256 bits de entropía. Un token de sesión es la credencial
// completa: quien lo tiene es el usuario, sin nada más que presentar.
const bytesToken = 32

// Sesion es una autenticación vigente.
//
// Guarda el hash del token, nunca el token. Si alguien lee el almacenamiento
// —un dump, un log, una réplica— se lleva hashes que no sirven para
// autenticarse. Es la misma razón por la que no se guardan contraseñas en
// claro, y cuesta una línea.
type Sesion struct {
	TokenHash [32]byte
	UsuarioID uuid.UUID
	CreadaEn  time.Time
	ExpiraEn  time.Time
}

// NuevaSesion devuelve la sesión a guardar y el token en claro, que es lo
// único que se le manda al cliente. El token no se puede recuperar después:
// esta es la única vez que existe.
func NuevaSesion(usuarioID uuid.UUID, ahora time.Time) (Sesion, string, error) {
	b := make([]byte, bytesToken)
	if _, err := rand.Read(b); err != nil {
		return Sesion{}, "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)

	return Sesion{
		TokenHash: HashearToken(token),
		UsuarioID: usuarioID,
		CreadaEn:  ahora,
		ExpiraEn:  ahora.Add(DuracionSesion),
	}, token, nil
}

// HashearToken convierte un token en la clave con la que se lo busca.
//
// SHA-256 pelado y no bcrypt: acá no hace falta un hash lento. bcrypt existe
// para que una contraseña de baja entropía no se pueda romper por fuerza
// bruta; un token de 256 bits aleatorios no tiene ese problema, y un hash
// lento en cada request sería un costo por nada.
func HashearToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}

// EstaVencida cierra el intervalo hacia afuera: en el instante exacto del
// vencimiento la sesión ya no vale.
func (s Sesion) EstaVencida(ahora time.Time) bool {
	return !ahora.Before(s.ExpiraEn)
}
