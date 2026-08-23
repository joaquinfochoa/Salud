# Salud API

Backend en Go. CRUD de profesionales, sin base de datos.

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
| `cmd/api/main.go` | El contenedor de DI, pero explícito |

## Migrar a PostgreSQL

Implementar `repository.Profesional` en `internal/repository/postgres/` y
cambiar una línea de `cmd/api/main.go`. Nada más.

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
- Una sola dependencia externa: `github.com/google/uuid`.
