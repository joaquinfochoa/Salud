# Salud

Plataforma de salud digital. Monorepo.

## Estructura

| Carpeta | Qué es |
|---|---|
| `apps/api` | Backend en Go. Ver su README. |
| `apps/web` | Frontend en Next.js. Lado paciente: buscar, ver un perfil y reservar. |
| `research/` | Lab de investigación en Python y el relevamiento de mercado. No es producto. |
| `legacy/prototype/` | Prototipo React + Vite, mockeado. Es la especificación visual de los flujos. Se elimina cuando `apps/web` lo reemplace. |
| `docs/superpowers/` | Specs de diseño y planes de implementación. |

## Levantar todo

Son dos procesos: la API en el 8080 y el front en el 3000.

```bash
cd apps/api && make run          # terminal 1
pnpm dev                         # terminal 2
```

El front espera la API en `http://localhost:8080`. Se cambia con
`NEXT_PUBLIC_API_URL`.

La API arranca con cuatro profesionales de prueba. Para entrar como uno de
ellos: `nombre.apellido@ejemplo.com`, contraseña `desarrollo123`.

## Prototipo (referencia visual)

```bash
cd legacy/prototype
npm install && npm run dev
```
