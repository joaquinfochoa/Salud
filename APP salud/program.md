# Programa de investigacion - Plataforma de Salud Digital

## Mision

Construir una plataforma que haga viable la relacion directa entre paciente y
profesional de salud, empezando con un MVP simple y validable, y avanzando luego
hacia una red prestacional digital con cobro, cobertura y liquidacion resueltos.

La meta no es hacer una app de turnos. La meta es reducir opacidad:

- El paciente entiende a quien consultar, cuando, cuanto cuesta y que opciones
  tiene antes de confirmar.
- El profesional gana pacientes, autonomia, reputacion propia y cobro mas rapido.

## Tesis actual

El diferencial grande no es directorio ni agenda. Es:

1. Orientacion inicial por necesidad.
2. Precio/cobertura visibles antes del turno.
3. Profesional verificado.
4. Pago y comprobante resueltos.
5. Camino futuro hacia red prestacional y liquidacion con financiadores.

## Principio de etapa

No intentar resolver obras sociales desde el dia 1.

Etapa 1:

- Cobro privado o copago informado.
- Comprobante util para reintegro.
- Orientacion asistida/manual.
- Agenda y teleconsulta/presencial.
- Profesionales verificados.

Etapa 2:

- Gestion asistida de reintegros.
- Convenios piloto con financiadores chicos.
- Automatizacion de liquidaciones.

Etapa 3:

- Red de Prestadores operativa.
- Cobro al financiador.
- Adelanto de honorarios al profesional con control financiero.

## Hipotesis prioritarias

H1. Pacientes con cobertura igual pagan privado si reciben turno rapido, precio
claro y profesional confiable.

H2. Profesionales independientes aceptan pagar comision o suscripcion si la
plataforma trae pacientes reales y reduce gestion administrativa.

H3. Salud mental es el mejor vertical inicial por recurrencia, teleconsulta y
demanda sostenida.

H4. Kinesiologia es un segundo vertical atractivo por frecuencia, episodios de
tratamiento y posible atencion domiciliaria.

H5. Odontologia puede ser diferencial por conocimiento del fundador, pero exige
mas presencialidad y mejor gestion de tickets altos.

## Reglas de decision

Una idea pasa a construccion solo si hay evidencia en al menos dos fuentes:

- entrevista con paciente;
- entrevista con profesional;
- dato publico/regulatorio;
- prueba real con pago o compromiso explicito.

No confundir interes con validacion. La senal fuerte es:

- pago;
- agenda tomada;
- profesional que entrega disponibilidad real;
- paciente que vuelve;
- recomendacion organica.

## Experimentos semanales

Cada semana debe producir una de estas salidas:

- 5 entrevistas nuevas;
- 1 experimento de venta;
- 1 busqueda academica o regulatoria que confirme/refute una hipotesis;
- 1 decision de producto;
- 1 riesgo legal/financiero aclarado;
- 1 profesional comprometido.

## Research loop automatico

El patron inspirado en autoresearch es:

1. Elegir una pregunta.
2. Buscar evidencia en fuentes externas.
3. Guardar resultados.
4. Convertirlos en un brief.
5. Decidir si cambia la estrategia.

Ejemplos de preguntas:

- Que evidencia existe sobre teleconsulta en salud mental?
- Los sistemas de orientacion/triage digital reducen consultas innecesarias?
- Que riesgos se reportan al usar IA para orientacion clinica?
- Que modelos de pago mejoran adherencia de profesionales independientes?

La IA puede ayudar a resumir y comparar papers, pero no debe reemplazar criterio
clinico, legal o regulatorio.

## Memoria del agente

El sistema local guarda la investigacion en CSV simples dentro de `data/`:

- `research_questions.csv`: preguntas que deciden estrategia.
- `sources.csv`: fuentes oficiales, informes, notas o datasets.
- `competitors.csv`: competidores directos y analogos internacionales.
- `regulations.csv`: leyes, resoluciones y riesgos.
- `tools.csv`: herramientas para no construir todo desde cero.
- `interviews.csv`: senales de pacientes y profesionales.
- `experiments.csv`: pruebas de mercado.
- `papers.csv`: evidencia academica de PubMed/arXiv.

El comando `scout` propone que buscar; el comando `brief` convierte lo cargado
en una lectura accionable.

## Modo autonomo con IA e internet

El comando `auto-research` ejecuta una pregunta con busqueda web e IA. Debe
usarse para reducir incertidumbre, no para confirmar sesgos.

Si el proveedor con internet falla, el modo local puede ordenar fuentes,
regulaciones y evidencias ya cargadas, pero no debe considerarse respuesta final.

Reglas:

- una pregunta por corrida;
- priorizar fuentes oficiales para datos de mercado y regulacion;
- priorizar papers para evidencia clinica o seguridad de IA;
- exigir URLs citadas;
- separar hechos, inferencias y opiniones;
- terminar siempre con un proximo experimento;
- si la evidencia es insuficiente, decirlo.

## Loop de retroalimentacion

El research no termina en el reporte. Despues de cada bloque de corridas, el
comando `reflect` debe:

1. Leer reportes recientes.
2. Extraer incertidumbres y proximos experimentos.
3. Crear preguntas nuevas solo si cambian decisiones.
4. Crear experimentos concretos ejecutables en 7 a 14 dias.
5. Guardar un reporte de reflexion auditable.

El comando `auto-loop` encadena `auto-scout`, `reflect --apply` y `brief`. Este
modo puede hacer crecer la investigacion, pero no debe reemplazar entrevistas ni
criterio legal/clinico.

## Autoanalisis del agente

El comando `self-analyze` revisa el comportamiento del propio research agent:

- que aprendio realmente;
- que puntos ciegos o sesgos aparecen;
- si esta acumulando research sin accion de campo;
- que decisiones candidatas ya estan maduras;
- que no debe decidir todavia.

Las decisiones que genera se guardan como `candidate`, no como decisiones finales.
Una decision candidata debe tener evidencia, confianza, proximo paso y fecha de
revision. El agente debe marcar como problema cualquier ciclo con muchos reportes
y pocas respuestas del fundador o experimentos ejecutados.

## Preguntas al fundador

Si todavia no queremos hacer entrevistas externas, el agente debe generar
preguntas concretas para el fundador con `founder-questions`. Cada pregunta debe
desbloquear una decision real y quedar guardada en `founder_questions.csv`.

Las respuestas del fundador no reemplazan evidencia de mercado, pero si evitan
que el agente invente preferencias, restricciones, conocimiento de industria o
criterios personales. Antes de seguir investigando una misma zona, el agente debe
mirar si hay preguntas abiertas al fundador sin responder.

## Preguntas a pacientes

1. Contame la ultima vez que necesitaste sacar un turno de salud.
2. Que fue lo mas dificil: saber a quien ir, conseguir turno, pagar, cobertura,
   confianza, distancia u otra cosa?
3. Antes de atenderte, sabias cuanto ibas a pagar?
4. Usaste obra social/prepaga o pagaste privado? Por que?
5. Si una plataforma te dijera especialidad probable, precio, disponibilidad y
   profesional verificado, la usarias?
6. Que tendria que pasar para que pagues por ahi?
7. Que te daria desconfianza?

## Preguntas a profesionales

1. Como conseguis pacientes hoy?
2. Que porcentaje viene por obra social, privado, institucion o recomendacion?
3. Cuanto tardas en cobrar cada canal?
4. Que tarea administrativa te molesta mas?
5. Que te haria probar una plataforma nueva?
6. Preferis pagar comision por consulta o suscripcion mensual?
7. Que no aceptarias delegar nunca?

## Riesgos no negociables

- Datos de salud son sensibles.
- No dar diagnostico automatico.
- No prometer reintegro automatico sin convenio real.
- Verificar matricula y habilitacion.
- Separar orientacion administrativa de acto clinico.
- Confirmar con abogado sanitario la figura de Red de Prestadores, cobros,
  responsabilidad y requisitos de direccion tecnica.

## Norte

Doctoralia hace visible al profesional.

La plataforma debe hacerlo viable.
