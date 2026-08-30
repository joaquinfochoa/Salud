# Salud API

Backend en Go. Profesionales, disponibilidad y autenticación por sesión.
Sin base de datos: todo vive en memoria.

## Correr

```bash
make run                    # con seed de desarrollo
curl localhost:8080/healthz
```

Configuración en `.env.example`.

## Arquitectura

Cuatro capas. `domain` no importa nada del proyecto.

```
handler ──▶ service ──▶ repository (interfaz)
   │           │              ▲
   └───────────┴──────────────┤
               ▼              │
            domain      repository/memory
```

| Carpeta | Equivalente en Spring / ASP.NET |
|---|---|
| `internal/handler` | `@RestController` |
| `internal/service` | `@Service` |
| `internal/repository/memory` | `@Repository` |
| `internal/domain` | Entidades y value objects |
| `internal/handler/middleware.go` | Filtros de servlet / middleware |
| `cmd/api/main.go` | El contenedor de DI, pero explícito |

## Migrar a PostgreSQL

Implementar las interfaces de `internal/repository/` en
`internal/repository/postgres/` y cambiar las líneas del cableado en
`cmd/api/main.go`. Nada más. Las sesiones además van a querer un `DELETE` por
`expira_en`: hoy solo se borran al detectarlas vencidas.

## Autenticación

La sesión es un token opaco de 32 bytes en una cookie `HttpOnly` llamada
`sesion`, con 7 días de vigencia absoluta. El servidor guarda el **hash** del
token, nunca el token: quien lea el almacenamiento no puede suplantar a nadie.

No es un JWT a propósito. Con un JWT el logout no existe —el token sigue
valiendo hasta vencer— y arreglarlo pide una lista de revocados, que es una
sesión reinventada, o un refresh flow, que es más piezas. Acá cerrar sesión es
borrar una fila. Ver la sección 3.4 del spec de diseño.

**La autorización se decide en el servicio, no en el handler.** El middleware
`Autenticar` resuelve la cookie y deja el `UsuarioID` en el `context`; el
handler se lo pasa al servicio como parámetro explícito y el servicio compara
contra el dueño. Un solo lugar, cubierto por tests que no levantan HTTP.

`RequerirSesion` se aplica ruta por ruta en `router.go`, no global: la mayor
parte del contrato es pública, y una lista de excepciones se desactualiza sola
la primera vez que alguien agrega un endpoint.

Probarlo a mano, con el seed cargado:

```bash
curl -c galletas.txt -X POST localhost:8080/api/v1/sesiones \
  -H 'Content-Type: application/json' \
  -d '{"email":"martin.gonzalez@ejemplo.com","contrasena":"desarrollo123"}'

curl -b galletas.txt localhost:8080/api/v1/usuarios/yo
```

Los cuatro profesionales del seed tienen email `nombre.apellido@ejemplo.com` y
la misma contraseña, `desarrollo123`. Solo existen con `APP_ENV=development`:
en producción el binario arranca vacío.

### Por qué `Content-Type: application/json` es obligatorio

Todo cuerpo que no lo traiga recibe 415. Es la mitad de la defensa contra CSRF:
un formulario de otro sitio solo puede mandar `form-urlencoded`, `multipart` o
`text/plain`, así que no puede forjar una escritura aunque el browser adjunte
la cookie. La otra mitad es `SameSite=Lax`.

## Contrato

`api/openapi.yaml` es la fuente de verdad. Se escribe antes que los handlers.

Para navegarlo con Swagger UI:

```bash
docker run --rm -p 8081:8080 \
  -e SWAGGER_JSON=/spec/openapi.yaml \
  -v "$(pwd)/api:/spec" swaggerapi/swagger-ui
```

## Comandos

| Comando | Qué hace |
|---|---|
| `make run` | Levanta el servidor con seed |
| `make test` | Corre los tests |
| `make test-race` | Tests con detector de carreras |
| `make lint` | golangci-lint |
| `make check` | fmt + test-race + lint. Correr antes de commitear. |
| `make docker-build` | Imagen local |

### El detector de carreras en Windows

`-race` necesita cgo, y cgo necesita un compilador de C. Windows no trae uno.
Sin él, `go test -race` falla con `gcc: executable file not found`.

```powershell
winget install BrechtSanders.WinLibs.POSIX.UCRT
```

El paquete no se agrega solo al PATH. Hay que apuntarlo a mano:

```bash
export PATH="$LOCALAPPDATA/Microsoft/WinGet/Packages/BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe/mingw64/bin:$PATH"
export CGO_ENABLED=1
```

En Linux y macOS no hace falta nada: `gcc` o `clang` ya están.

## Convenciones

- Sin mocks. El repositorio en memoria es el doble de test.
- Dinero en `int64` de centavos, nunca `float64`.
- Todo lo que escribimos nosotros va en español: tipos, funciones, campos,
  constantes, comentarios, mensajes y los nombres de campo del JSON. Quedan
  en inglés los paquetes (`domain`, `service`, ...), `String()` y `Error()`,
  las variables de entorno y las claves `type`/`title`/`status`/`detail`
  del RFC 7807.
- Dos dependencias externas: `github.com/google/uuid` y `golang.org/x/crypto`
  (bcrypt). El hasheo de contraseñas es la única parte del proyecto que no se
  escribe a mano.

### Linting

`misspell` está deliberadamente apagado en `.golangci.yml`: valida contra un
diccionario en inglés y este proyecto escribe en español a propósito, así que
cada hallazgo es una palabra española bien escrita. Las reglas `exported` y
`package-comments` de `revive` también están apagadas porque exigen un
comentario godoc en cada símbolo exportado, justo el relleno que la
convención de arriba rechaza (comentarios que expliquen el porqué, no que
repitan el nombre).
