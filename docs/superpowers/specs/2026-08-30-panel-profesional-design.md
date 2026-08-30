# Diseño — Panel del profesional

- **Fecha:** 2026-08-30
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de plan de implementación
- **Depende de:** [apps/web, lado paciente](2026-08-30-apps-web-design.md). Ninguna entidad nueva.

---

## 1. Contexto

`apps/web` cubre el lado paciente: buscar, ver un perfil y reservar. Le falta el
otro lado del marketplace, y sin ese lado el primero no existe: **sin
profesionales cargando su semana no hay nada que buscar.**

Los endpoints ya están. `PUT /horarios`, `POST` y `DELETE /bloqueos`,
`GET /profesionales/{id}/turnos` y `PUT /profesionales/{id}` existen desde las
etapas 2, 3 y 4, y **ninguno tiene todavía un consumidor.** Esta etapa no agrega
dominio: le pone cara a lo que ya funciona.

### El prototipo como referencia

`legacy/prototype` tiene las pantallas del profesional resueltas —dashboard,
agenda y perfil— y son la base de esta etapa. Lo que aporta y se respeta: el
saludo con la especialidad, el nudge de "completá tu perfil", los turnos del día
con los pasados en gris, la tira de semana con punto en los días ocupados, y las
obras sociales como chips que se prenden y apagan.

---

## 2. Una app, tres grupos de rutas

**No dos frontends.** La razón es del dominio, no de conveniencia: `Usuario` no
tiene campo `Rol` —sos profesional si existe un `Profesional` con tu
`UsuarioID`— y eso es deliberado (sección 3.1 de la spec de autenticación).
**La misma persona puede ser paciente de otro profesional.** Dos apps
separadas pelean contra eso: dos builds, dos clientes de API, dos sistemas de
diseño derivando.

Y no se gana lo que parece: Next separa el código por ruta, así que una librería
importada solo en `/panel` no se descarga cuando alguien abre un perfil público.

```
app/
  (publico)/                     header de marketing
    page.tsx                     /                     landing = buscador
    perfiles/[slug]/                                   perfil público
    profesionales/                                     landing de captación
  (paciente)/                    header simple
    turnos/
  (panel)/                       tabs abajo en móvil, barra lateral en escritorio
    panel/                                             hoy
    panel/agenda/
    panel/horarios/
    panel/perfil/
  entrar/                        un solo login
```

Cada grupo tiene su `layout.tsx`. Comparten los tokens, `pedir()` y la sesión.

**Si algún día hay que separar**, con grupos de rutas es mover carpetas. Esta
estructura no cierra esa puerta.

---

## 3. El login único

Uno solo, para los dos lados. Después de entrar, `GET /api/v1/usuarios/yo`
devuelve `perfilProfesionalId`:

| Valor | A dónde va |
|---|---|
| un uuid | `/panel` |
| `null` | `/turnos` |

Ese campo se diseñó en la etapa de autenticación sin saber que iba a resolver
esto. Es la consecuencia de no haber puesto un enum `Rol`: el destino se deriva
del estado real en vez de un dato que se puede desincronizar.

El `?volver=` sigue mandando: si alguien llegó a `/entrar` desde una reserva,
vuelve ahí. La redirección por perfil es solo el default.

---

## 4. Layout: móvil primero, crece a escritorio

Las mismas rutas, un solo layout responsive.

| | Móvil | Escritorio |
|---|---|---|
| Navegación | tabs abajo: Hoy · Agenda · Horarios · Perfil | barra lateral |
| `/panel/horarios` | una columna, un día por vez | grilla de siete columnas |

Son dos usos distintos y los dos son reales: **mirar la agenda del día se hace
en el teléfono entre pacientes; cargar la semana entera se hace sentado.** Un
layout que solo sirve para uno de los dos falla la mitad de las veces.

---

## 5. Las cuatro pantallas

### `/panel` — hoy

Saludo con nombre y especialidad. Los turnos de hoy con hora, paciente, motivo y
modalidad; los que ya pasaron en gris, igual que el prototipo.

**Dos KPIs, y son los que se pueden calcular de verdad:**

| KPI | De dónde sale |
|---|---|
| **Turnos esta semana** | `GET /profesionales/{id}/turnos` con el rango de la semana |
| **Ocupación** | turnos activos sobre huecos ofrecidos en los próximos 7 días |

La ocupación es un número que un profesional independiente mira: dice si le
sobra agenda o le falta. Y es honesto — sale de datos que existen.

**No se muestran ingresos ni ausentismo.** El prototipo tiene "Cobrado hoy
$16.9k" y "2 completados", y son los dos números que un profesional más quiere.
No hay pagos ni existe `atendido`/`ausente`, así que mostrarlos sería inventar.
Cuando existan, esta pantalla es donde van.

### `/panel/agenda`

La tira de días con punto en los que tienen turnos, y la lista del día elegido.
El botón `+` crea un bloqueo sobre el día que se está mirando.

Los turnos cancelados aparecen tachados, y **dice quién canceló**: comparando
`canceladoPor` contra `pacienteId` se sabe si fue el paciente o el propio
profesional. Es el dato que guardamos en la etapa 4 sin que nada lo leyera
todavía.

### `/panel/horarios`

Los siete días con sus bloques: desde, hasta, duración y modalidad. Es la
pantalla donde el profesional se sienta a trabajar.

**Guardar puede cancelar turnos, y hay que avisarlo antes.** `PUT /horarios`
cancela los turnos futuros que quedan fuera del horario nuevo y devuelve cuántos
en `turnosCancelados`. Esta pantalla no puede enterarse después: antes de
guardar consulta los turnos futuros, calcula cuáles quedarían huérfanos con la
semana nueva, y si hay alguno **pide confirmación nombrando el número**.

> Con este cambio se cancelan **3 turnos** ya reservados. Los pacientes lo van a
> ver en la app. ¿Guardar igual?

Sin ese aviso, un profesional que acorta un bloque descubre que canceló seis
turnos recién cuando ve el número en la respuesta.

Lo mismo, más suave, para los bloqueos: `POST /bloqueos` también devuelve
`turnosCancelados`, y la pantalla lo informa.

### `/panel/perfil`

Bio, precio, zona, modalidades y obras sociales como chips. Con "Ver mi perfil
público", que abre `/perfiles/{slug}`.

**La misma pantalla en dos modos.** Un usuario con cuenta pero sin perfil
profesional ve el mismo formulario, con matrícula y especialidad —que solo se
piden al crear, porque cambiarlas resetea la verificación— y el botón dice
*"Crear mi perfil"* en vez de *"Guardar cambios"*.

Hace falta: sin eso, el CTA de la landing de captación no lleva a ningún lado.
Es un modo más de un formulario que ya existe, no una pantalla nueva, y usa el
`POST /api/v1/profesionales` que hoy no tiene consumidor. El alta pensada para
profesionales —con onboarding, subida de matrícula y validación— es su propia
etapa, y esto no la reemplaza: la destraba.

**El precio se edita en pesos y viaja en centavos.** Es el lugar donde más fácil
se cuela un error de dos órdenes de magnitud, así que la conversión pasa por
`formato.ts` y el campo muestra el valor formateado mientras se escribe.

---

## 6. El nudge

La pieza del prototipo que más vale, y nuestro dominio tiene el equivalente
exacto: **un profesional sin horarios cargados no aparece con disponibilidad en
la búsqueda.** Lo descubrimos en la etapa anterior, cuando el seed dejaba a los
cuatro sin agenda y el listado mostraba cuatro tarjetas vacías.

En `/panel`, un banner que dice qué falta y linkea a arreglarlo:

| Estado | Qué dice |
|---|---|
| Sin horarios cargados | *Todavía no cargaste tus horarios, así que nadie puede reservarte un turno.* → **Configurar** |
| Sin bio | *Tu perfil no tiene descripción. Es lo primero que lee un paciente.* → **Completar** |
| `verificacion: pendiente` | *Tu matrícula está pendiente de verificación.* Sin link: no depende del profesional. |

El tercero es informativo a propósito. Hoy nada mueve ese estado —la
integración con REFEPS es su propia etapa— y ofrecer un botón que no hace nada
es peor que no ofrecer ninguno.

---

## 7. La landing

**El hero es el buscador, no un texto.** `/` es la página con más autoridad de
SEO del sitio, y gastarla en un hero de marketing sin contenido indexable es
tirarla. Debajo del buscador van la propuesta de valor y los profesionales, que
es contenido real que Google lee.

Es lo que hacen Doctoralia y Zocdoc, y no es casualidad: en un marketplace de
salud la búsqueda orgánica es el canal de adquisición.

`/profesionales` sí es una landing de verdad: su propio mensaje para captar
profesionales, con un CTA a registrarse. Se enlaza desde el header y el footer
del grupo público.

---

## 8. Fuera de alcance

| Pendiente | Cuándo se activa |
|---|---|
| KPIs de ingresos | Con pagos |
| KPI de ausentismo | Con `atendido` / `ausente` |
| Notificaciones y la campana | Etapa propia, con proveedor externo |
| Foto de perfil | Necesita subida, almacenamiento, moderación y una decisión legal |
| Obra social del paciente | El paciente es un `Usuario` y no la tiene. Las que existen son las que el profesional acepta. |
| Onboarding de profesional | El alta mínima entra (ver 5, `/panel/perfil`). Un onboarding de verdad —pasos, subida de matrícula, validación— es su propia etapa. |
| Vista de admin | No hay endpoints de admin |
| Borrar `legacy/prototype` | Cuando estas pantallas lo reemplacen |

---

## 9. Criterios de aceptación

1. `pnpm build`, `pnpm lint` y `pnpm test` en verde, y el E2E pasa.
2. Entrar con una cuenta que tiene perfil profesional lleva a `/panel`; una sin
   perfil, a `/turnos`. Desde la landing de captación, una cuenta sin perfil
   puede crearlo y queda en `/panel`.
3. `/panel` muestra los turnos de hoy, y los que ya pasaron se distinguen de los
   que faltan.
4. Los dos KPIs muestran números que coinciden con lo que devuelve la API.
5. Un profesional sin horarios ve el nudge, y el link lo lleva a `/panel/horarios`.
6. Se puede cargar la semana, y los huecos aparecen en el perfil público.
7. **Guardar un horario que dejaría turnos huérfanos pide confirmación nombrando
   cuántos**, antes de guardar.
8. Se puede crear y borrar un bloqueo, y la respuesta informa cuántos turnos
   canceló.
9. Editar bio, precio, zona, modalidades y obras sociales se refleja en el perfil
   público.
10. `/` sigue viéndose completa con JavaScript apagado, ahora con la landing.
11. El panel se recorre entero con teclado, con foco visible.
12. El motivo de consulta aparece **solo** en `/panel` y `/panel/agenda` —las
    dos detrás de sesión y solo para el dueño del perfil— y en ningún `<title>`.
    Nunca en una página pública ni en la landing.
