# apps/web

Frontend de Salud en Next.js. Lado paciente: buscar un profesional, ver su
perfil y reservar un turno.

Ver el README de la raíz para levantarlo. La documentación de por qué está
armado así vive en
[la spec de diseño](../../docs/superpowers/specs/2026-08-30-apps-web-design.md).

## Lo que hay que saber antes de tocar

- **`/` y `/perfiles/[slug]` son Server Components y tienen que seguir
  siéndolo.** Son las dos páginas que indexa Google, que es la razón por la que
  el front es Next y no Vite. Un `useState` mal puesto las convierte en
  componentes de cliente sin que nada falle, y el SEO desaparece en silencio.
  Se verifica apagando JavaScript en el browser, no leyendo el código.
- **Los tipos de la API se generan**, no se escriben: `pnpm contrato` los
  regenera desde `apps/api/api/openapi.yaml`. `lib/contrato.ts` no se edita a
  mano.
- **El acento (`--color-accion`) va solo en la acción principal de cada
  pantalla.** El resto de las reglas visuales están en la sección 6 de la spec.
