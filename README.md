# Salud

Plataforma de salud digital. Monorepo.

## Estructura

| Carpeta | Qué es |
|---|---|
| `apps/api` | Backend en Go. Ver su README. |
| `apps/web` | Frontend en Next.js. Todavía no existe. |
| `research/` | Lab de investigación en Python y el relevamiento de mercado. No es producto. |
| `legacy/prototype/` | Prototipo React + Vite, mockeado. Es la especificación visual de los flujos. Se elimina cuando `apps/web` lo reemplace. |
| `docs/superpowers/` | Specs de diseño y planes de implementación. |

## Backend

```bash
cd apps/api
go run ./cmd/api
```

## Prototipo (referencia visual)

```bash
cd legacy/prototype
npm install && npm run dev
```
