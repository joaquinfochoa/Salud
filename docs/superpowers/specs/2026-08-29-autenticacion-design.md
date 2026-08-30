# Diseño — Identidad y autenticación

- **Fecha:** 2026-08-29
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de plan de implementación
- **Depende de:** [CRUD de Profesional](2026-08-21-professional-crud-go-design.md), [Disponibilidad](2026-08-22-disponibilidad-design.md)
- **Desbloquea:** `Turno`

---

## 1. Contexto

Las dos etapas anteriores construyeron cosas públicas: un directorio y una
agenda. Ninguna necesitaba saber quién preguntaba, y por eso diferir
autenticación no costó nada.

`Turno` rompe esa frontera —reservar, cancelar y listar preguntan "¿de
quién?"— así que la sección 9 de disponibilidad ya lo dejó escrito: **antes de
`Turno`, autenticación**.

Esta etapa además **paga la deuda registrada en la sección 10 de esa spec**: hoy
cualquiera con la URL da de baja a un profesional, le vacía la agenda o se la
llena de bloqueos.

---

## 2. Modelo

Dos entidades nuevas y un campo nuevo en `Profesional`.

| `Usuario` | | | `Sesion` | |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | | `TokenHash` | `[32]byte` |
| `Email` | `Email` (VO, único) | | `UsuarioID` | `uuid.UUID` |
| `Hash` | `[]byte`, **nil permitido** | | `CreadaEn` | `time.Time` |
| `Nombre`, `Apellido` | `string` | | `ExpiraEn` | `time.Time` |
| `CreadoEn` | `time.Time` | | | |

```
Profesional
  UsuarioID  uuid.UUID   ← nuevo. Obligatorio y único.
```

Archivos nuevos: `domain/usuario.go`, `domain/email.go`, `domain/sesion.go`,
`repository/{usuario,sesion}.go` y sus implementaciones en `repository/memory/`.
Mismo patrón que los tres repositorios existentes.

---

## 3. Decisiones

### 3.1 No hay campo `Rol`

Sos profesional si existe un `Profesional` con tu `UsuarioID`. Sos paciente
siempre. Un enum `Rol` almacenado puede desincronizarse del perfil real
(`Rol=profesional` sin perfil, o al revés) y no dice nada que la referencia no
diga ya. `EsAdmin` aparece el día que exista un endpoint que lo necesite; hoy no
hay ninguno.

### 3.2 `Profesional.UsuarioID` es obligatorio, no opcional

La spec de disponibilidad anticipaba "una referencia opcional". Se descarta: un
`Profesional` sin dueño es exactamente el bug que esta etapa arregla, y nullable
deja la puerta abierta a recrearlo. No hay datos en producción que migrar.

**Consecuencia — cambio de contrato:** `POST /api/v1/profesionales` deja de ser
público. Pasa a significar "crear **mi** perfil profesional": el `UsuarioID`
sale de la sesión, **nunca** del body. Un usuario tiene como máximo un perfil.

### 3.3 `Nombre` y `Apellido` quedan duplicados

`Usuario.Nombre` es cómo te llamás; `Profesional.Nombre` es cómo te presentás
profesionalmente, y pueden diferir de verdad. Pero la razón honesta es que
moverlos de `Profesional` toca el slug, el seed, el contrato y sus tests a
cambio de pureza. Se deja duplicado a propósito.

### 3.4 Sesión opaca en cookie, no JWT

El eje real es doble: transporte (cookie vs `Bearer`) y formato (opaco vs JWT).
Se eligen por separado.

| | Sesión opaca | JWT |
|---|---|---|
| Dependencias | **Ninguna.** `crypto/rand`, `crypto/sha256`, `net/http` | Una (`golang-jwt`; a mano no, ahí vive `alg=none`) |
| Logout | Borrar la fila | No existe sin lista de revocados o refresh flow |
| Piezas | Un token | Access + refresh, endpoint de refresh, rotación |
| Lo que se paga | Almacenamiento de sesiones + limpieza de vencidas + estado compartido el día que haya dos instancias | Una clave secreta que rotar, y un token robado sirve hasta que vence |

Decide la revocación. Son datos de salud —`research/` registra Ley 25.326
(datos sensibles) y Ley 26.529 (confidencialidad) entre los riesgos
principales— y un token que no se puede invalidar es mala combinación. Las
ventajas de JWT (verificación sin consultar, multi-servicio, terceros) aplican a
una topología que este proyecto no tiene: un binario y un `map`.

La cookie va por el consumidor: un SPA del mismo sitio. `HttpOnly` es invisible
para JS; un token en `localStorage` lo lee cualquier XSS.

### 3.5 SSO con Google: no se construye acá, pero el modelo lo espera

SSO llega en la etapa siguiente. Esta spec solo reserva lo que después sería
caro de cambiar, y **nada más**.

**Por qué encaja sin fricción:** Google reemplaza el *login*, no la *sesión*. Se
verifica la identidad una vez y a partir de ahí se emite la cookie de siempre;
el token de Google no vuelve a aparecer. `Sesion`, `Autenticar`,
`RequerirSesion`, la autorización entera y las firmas de servicio no se tocan.
Es una consecuencia directa de 3.4: con JWT habría dos sistemas de tokens
conviviendo.

**Lo único que se reserva hoy:** `Usuario.Hash` admite `nil`. Un usuario de
Google no tiene contraseña, y volver el campo opcional después toca el
constructor, la validación y sus tests.

**Lo que NO se agrega hoy:** los campos de identidad externa (`Proveedor`,
`SubjectExterno`). Serían dos campos siempre vacíos, y sin base de datos
agregarlos más adelante es editar un struct. Dejarlos ahora es flexibilidad
muerta.

**El invariante lo sostienen los constructores, no el tipo.** `Hash` nil es un
estado válido del modelo pero hoy inalcanzable: `NuevoUsuario` exige contraseña.
Mañana `NuevoUsuarioConGoogle` exigirá un `sub`. La regla —*todo `Usuario` tiene
al menos una forma de autenticarse*— vive en los constructores, que son los
únicos que arman la entidad.

**La regla de vinculación, escrita antes de necesitarla.** Alguien se registra
con `juan@gmail.com` y contraseña; un mes después entra con Google y el mismo
email. Es la misma cuenta **solo si el proveedor afirma `email_verified: true`**
(Google lo hace). Vincular por email sin esa verificación es un secuestro de
cuenta directo: me registro con tu email en un proveedor flojo y heredo la tuya.
Y la identidad se ancla al `sub` de Google, **nunca al email**: el email cambia,
el `sub` no.

**Anticipo de implementación, para que no sorprenda:** Authorization Code +
PKCE, nunca implicit. Sin librería —son dos llamadas HTTP con `net/http`,
`net/url` y `encoding/json`— y sin verificar la firma del ID token: OIDC Core lo
permite explícitamente cuando el token llega directo del token endpoint sobre
TLS, porque validar el certificado ya prueba con quién hablaste. Sin JWKS, sin
rotación de claves, sin dependencias nuevas.

---

## 4. Autorización

Vive **en el servicio, no en el handler**. El handler pasa el `UsuarioID` de la
sesión y el servicio compara contra el dueño: un solo lugar, y los tests del
servicio lo cubren sin levantar HTTP.

| Operación | Quién |
|---|---|
| `GET` de profesionales, perfiles, horarios y huecos | Cualquiera. No cambia nada. |
| `PUT` / `DELETE /profesionales/{id}`, `POST .../reactivar` | Solo el dueño |
| `PUT .../horarios`, `POST` / `DELETE .../bloqueos` | Solo el dueño |
| `POST /profesionales` | Cualquier usuario con sesión. El segundo intento del mismo usuario da **409**. |

Sin sesión, las operaciones privadas dan **401**; con sesión ajena, **403**. La
distinción importa: 401 es "no sé quién sos", 403 es "sé quién sos y no te
alcanza".

Centinelas nuevos, siguiendo el patrón de `ErrNoEncontrado` y `ErrMatriculaEnUso`:

| Centinela | HTTP | Cuándo |
|---|---|---|
| `ErrNoAutorizado` | 403 | El usuario de la sesión no es el dueño |
| `ErrEmailEnUso` | 409 | Registro con un email ya registrado |
| `ErrYaTienePerfil` | 409 | Segundo `POST /profesionales` del mismo usuario |
| `ErrCredencialesInvalidas` | 401 | Login fallido. **Un solo error para email inexistente y contraseña incorrecta**: distinguirlos convierte el login en un oráculo de qué emails están registrados. |

---

## 5. Mecanismo

```go
func NuevoToken() (token string, hash [32]byte, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", [32]byte{}, err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, sha256.Sum256([]byte(token)), nil
}
```

- **Login:** `NuevoToken()` → guardar `{hash, usuarioID, expiraEn}` → `http.SetCookie`.
- **Cada request:** leer cookie → `sha256` → buscar → `UsuarioID` al `context`.
- **Logout:** borrar la sesión.

Se guarda **el hash del token, nunca el token**: si alguien lee el
almacenamiento, no puede suplantar a nadie.

| Pieza | Decisión |
|---|---|
| Contraseña | Value object. Mínimo 12 caracteres, máximo 72 **bytes** (no caracteres: bcrypt trunca en silencio en el byte 72, y una contraseña con acentos llega antes de lo que parece). Sin reglas de composición —mayúscula, número, símbolo—: alargan sin agregar entropía real y empujan a la gente a `Password1!`. |
| Hash de contraseña | `golang.org/x/crypto/bcrypt`, cost default. **Segunda dependencia del proyecto**; actualizar la convención del README. No se escribe a mano. |
| Email | `net/mail.ParseAddress` (stdlib) + minúsculas + tope de largo. Sin regex propia: validar email "bien" rompe direcciones válidas. |
| Cookie | `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, expiración explícita. |
| Vigencia | 7 días, absoluta. Sin renovación deslizante — es menos código y se agrega después sin cambiar firmas. |
| Vencidas | Se rechazan al leer. Sin proceso de limpieza: el `map` de una demo no crece lo suficiente para importar. |
| Middleware | `Autenticar` resuelve la cookie y pone el `UsuarioID` en el `context`; **no rechaza**. `RequerirSesion` es el que devuelve 401. Se encadenan con el `Encadenar` existente. |
| CSRF | `SameSite=Lax` bloquea la cookie en POST cross-site. Además `decodificarJSON` pasa a **exigir `Content-Type: application/json`** (415 si no): un form cross-origin solo puede mandar `form-urlencoded`, `multipart` o `text/plain`. Dos defensas independientes. |

---

## 6. Endpoints

```
POST   /api/v1/usuarios          registro. Setea la cookie: registrarse loguea.
POST   /api/v1/sesiones          login. Setea la cookie.
DELETE /api/v1/sesiones/actual   logout. Borra la sesión y vence la cookie.
GET    /api/v1/usuarios/yo       quién soy, y si tengo perfil profesional. 401 sin sesión.
```

Ninguna respuesta incluye nunca el hash ni el token: el token viaja **solo** en
el `Set-Cookie`.

Sesión como recurso, no `/login`: consistente con el resto del contrato.
`api/openapi.yaml` se escribe antes que los handlers, como en las dos etapas
anteriores.

---

## 7. Fuera de alcance

| Pendiente | Cuándo se activa |
|---|---|
| **SSO con Google** | **Etapa siguiente.** Ver 3.5: puramente aditivo, dos endpoints y dos campos. Nada que migrar. |
| Verificación de email | Cuando haya que mandar un mail que importe. Los usuarios de Google no la necesitan: ya viene verificado. |
| Recuperar contraseña | Antes de tener usuarios reales. Solo para quien tenga contraseña. |
| Rate limiting del login | **Antes de exponer la API.** Sin deploy no protege nada, pero no puede quedar olvidado. |
| Renovación deslizante de sesión | Si 7 días molesta en uso real |
| 2FA, refresh tokens, otros proveedores (Apple, etc.) | Sin fecha. Ninguno cambia una firma de servicio. Un segundo proveedor sí pediría mover la identidad externa a su propia entidad. |
| Rol `admin` y permisos granulares | Cuando exista un endpoint de admin |
| Limpieza de sesiones vencidas | Con PostgreSQL |

---

## 8. Criterios de aceptación

1. `make test-race` y `make lint` en verde.
2. Registro → login → crear perfil profesional → el perfil aparece en el listado
   público.
3. Un segundo usuario recibe **403** al editar ese perfil, sus horarios y sus
   bloqueos.
4. Sin cookie, todos los `GET` públicos siguen funcionando igual que antes.
5. Logout invalida la sesión: usarla después da **401**.
6. Una sesión vencida da **401** sin que nadie la haya borrado.
7. Un `POST` con `Content-Type` que no sea `application/json` da **415**.
8. El seed crea los 4 usuarios con sus perfiles vinculados y `make run` sigue
   arrancando.
