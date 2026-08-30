# Diseño — Las dos partes del turno

- **Fecha:** 2026-08-30
- **Rama:** `refactor-gian`
- **Estado:** aprobado, pendiente de implementación
- **Depende de:** [Turno](2026-08-29-turno-design.md)
- **Desbloquea:** [Panel del profesional](2026-08-30-panel-profesional-design.md)

Etapa corta: un DTO y dos lecturas. Sin plan aparte — es una sola tarea, y un
documento de plan para una tarea es ceremonia.

---

## 1. El problema

`Turno` viaja con `profesionalId` y `pacienteId`, y no existe ningún endpoint
para resolver un usuario por id. Los dos listados quedan a medias:

- **`GET /api/v1/turnos`** no dice **con quién** es el turno. La pantalla
  `/turnos` del front ya está construida y muestra hora, día y motivo — nunca el
  nombre del profesional. **Es un defecto que ya está en la rama**, y pasó los
  tests y el E2E porque verifican que el turno *aparezca*, no que sea *útil*.
- **`GET /api/v1/profesionales/{id}/turnos`** no dice **quién viene**.

Es el cuarto hallazgo de la misma clase —después de CORS, del seed sin horarios
y del N+1 del listado— y sigue el mismo patrón: **un contrato diseñado sin
consumidor.** Los ids alcanzan para el dominio y no alcanzan para una pantalla.

---

## 2. La forma

Cada listado devuelve **la otra parte**, no las dos.

| Endpoint | Qué agrega |
|---|---|
| `GET /api/v1/turnos` | `profesional: { id, nombre, apellido, slug, especialidad }` |
| `GET /api/v1/profesionales/{id}/turnos` | `paciente: { id, nombre, apellido }` |

**Dos tipos, no uno con la mitad vacía.** Un `Turno` con `profesional` y
`paciente` ambos opcionales obliga a cada consumidor a saber cuál viene lleno
según de qué endpoint lo sacó. Dos schemas lo dicen en el tipo.

`POST /api/v1/profesionales/{id}/turnos` sigue devolviendo un `Turno` pelado: el
paciente acaba de venir de la página de ese profesional, así que ya sabe con
quién reservó.

---

## 3. Por qué esto no filtra nada

Los dos listados exigen sesión y solo devuelven turnos donde quien pregunta es
una de las dos partes: `ListarDePaciente` filtra por el `pacienteID` de la
sesión, y `ListarDeProfesional` exige ser el dueño del perfil.

El nombre de la otra parte es un dato al que ya tenés derecho — **no tenerlo es
lo raro**: reservaste un turno con esa persona.

**El email del paciente no viaja.** El profesional necesita saber quién viene, no
cómo contactarlo por fuera de la plataforma. El día que haga falta —para avisar
una cancelación, por ejemplo— es una decisión aparte y va con notificaciones.

---

## 4. Dónde se enriquece

**En el servicio, no en el handler.** El handler traduce entre HTTP y el
servicio; componer dos entidades es trabajo del servicio, y ahí lo cubren los
tests sin levantar un servidor.

`service.Turno` gana `repository.Usuario` como dependencia. No hay ciclo: el
servicio ya depende de `repository.Profesional`.

```go
type TurnoConProfesional struct {
    domain.Turno
    Profesional domain.Profesional
}

type TurnoConPaciente struct {
    domain.Turno
    Paciente domain.Usuario
}
```

**Los ids se deduplican antes de resolverlos.** Un paciente con seis turnos con
el mismo profesional hace una lectura, no seis. Es un `map` y cinco líneas, y
evita que el N+1 crezca con el historial en vez de con la cantidad de personas.

**Una parte que no se puede resolver no rompe el listado.** Si un usuario fue
borrado —hoy imposible, mañana con una purga por Ley 25.326 no— el turno se
devuelve igual con la parte en su valor cero. Que un dato faltante haga
desaparecer la agenda entera sería peor que mostrarla incompleta.

---

## 5. Fuera de alcance

| Pendiente | Cuándo |
|---|---|
| `GET /usuarios/{id}` | Nunca, si se puede evitar: un endpoint que resuelve cualquier usuario por id es una superficie que hoy nadie necesita. |
| Email o teléfono del paciente | Con notificaciones |
| Obra social del paciente | El paciente es un `Usuario` y no la tiene |
| Filtrar el listado del profesional por estado | Cuando alguien lo pida |

---

## 6. Criterios de aceptación

1. `make check` en verde: `-race` y lint.
2. `GET /api/v1/turnos` devuelve, por cada turno, el nombre, apellido, slug y
   especialidad del profesional.
3. `GET /api/v1/profesionales/{id}/turnos` devuelve el nombre y apellido del
   paciente, y **no su email**.
4. Seis turnos con el mismo profesional hacen **una** lectura del repositorio,
   no seis.
5. Un turno cuya otra parte no existe se devuelve igual, sin romper el listado.
6. El contrato valida y las dos formas nuevas están en `openapi.yaml`.
7. **`/turnos` del front muestra con quién es cada turno.** Es la pantalla que
   esta etapa vino a arreglar; dejarla sin cambiar sería agregar un campo que
   nadie lee.
