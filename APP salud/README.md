# Salud Digital Startup Lab

Kit inicial para investigar, validar y ordenar la startup de salud digital.

La idea esta inspirada en el patron de `karpathy/autoresearch`: no empezar con una
app enorme, sino con un sistema chico que convierta contexto, entrevistas,
evidencia y experimentos en decisiones semanales.

## Archivos principales

- `program.md`: instrucciones vivas del proyecto. Actua como "sistema operativo"
  de investigacion.
- `startup_ai_lab.py`: script local para registrar entrevistas, evidencia,
  experimentos y generar briefs.
- `data/`: carpeta creada por el script para guardar CSV y reportes.

## Primer uso

```powershell
python .\startup_ai_lab.py init
python .\startup_ai_lab.py seed
python .\startup_ai_lab.py prompts
python .\startup_ai_lab.py scout
python .\startup_ai_lab.py brief
```

## Research agent

El agente local separa la investigacion en piezas simples:

- preguntas estrategicas;
- fuentes oficiales o academicas;
- competidores y analogos;
- regulacion;
- herramientas reutilizables;
- entrevistas;
- experimentos.

Sembrar la base inicial:

```powershell
python .\startup_ai_lab.py seed
```

Ver preguntas abiertas:

```powershell
python .\startup_ai_lab.py questions
```

Generar plan de busqueda semanal:

```powershell
python .\startup_ai_lab.py scout
```

Generar brief de decision:

```powershell
python .\startup_ai_lab.py brief
```

## Modo autonomo con internet + IA

Este modo sigue el patron de `karpathy/autoresearch`: una pregunta, busqueda
web, sintesis, fuentes, decision y log.

### Opcion A: OpenAI API

Requisito:

```powershell
$env:OPENAI_API_KEY="tu_api_key"
```

Tambien podes usar el archivo `.env` ya creado en la carpeta del proyecto:

```env
OPENAI_API_KEY=sk-proj-tu_api_key_real
OPENAI_MODEL=gpt-5
```

Verificar si quedo cargada:

```powershell
python .\startup_ai_lab.py check-api
```

Probar conexion real con OpenAI:

```powershell
python .\startup_ai_lab.py check-api --live
```

Investigar una pregunta ya cargada:

```powershell
python .\startup_ai_lab.py auto-research --question-id MKT-002 --domain-bundle official_ar
```

### Opcion B: Claude Code CLI sin API key

Si tenes Claude Code instalado y logueado en VS Code/terminal:

```powershell
python .\startup_ai_lab.py check-claude
python .\startup_ai_lab.py check-claude --live
python .\startup_ai_lab.py auto-research --provider claude_cli --question-id MKT-002 --domain-bundle official_ar
```

Tambien podes dejarlo por defecto en `.env`:

```env
AI_PROVIDER=claude_cli
CLAUDE_MODEL=sonnet
```

Este modo usa tu sesion de Claude Code, no una `ANTHROPIC_API_KEY`.

### Opcion C: fallback local sin internet

Si todavia no hay API key o el proveedor falla, podes generar un reporte
preliminar usando solo la memoria local (`sources`, `regulations`, `papers`,
`competitors`, `evidence`):

```powershell
python .\startup_ai_lab.py auto-research --provider local --question-id LEG-002
```

Este modo no reemplaza la busqueda web: sirve para no perder el hilo, ordenar
lo ya cargado y decidir que fuente o entrevista falta.

Investigar una pregunta libre:

```powershell
python .\startup_ai_lab.py auto-research --question "Cuantos psicologos matriculados hay en AMBA y que evidencia hay de demora de turnos?" --domain-bundle all
```

Investigar automaticamente las 3 preguntas abiertas mas importantes:

```powershell
python .\startup_ai_lab.py auto-scout --limit 3 --domain-bundle official_ar
```

## Loop de retroalimentacion

Despues de investigar, el agente puede leer reportes recientes, detectar
incertidumbres, crear nuevas preguntas y proponer experimentos:

```powershell
python .\startup_ai_lab.py reflect --provider local
python .\startup_ai_lab.py reflect --provider claude_cli --apply --reports 4
```

`reflect` sin `--apply` solo genera un reporte. Con `--apply`, agrega nuevas
filas a `research_questions.csv` y `experiments.csv`, evitando duplicados.

Para correr un ciclo completo de investigacion + reflexion + brief:

```powershell
python .\startup_ai_lab.py auto-loop --provider claude_cli --cycles 1 --research-limit 2 --domain-bundle official_ar
```

El agente tambien puede autoanalizarse: revisa lo investigado, detecta sesgos,
propone autocorrecciones y guarda decisiones candidatas en `decisions.csv`.

```powershell
python .\startup_ai_lab.py self-analyze --provider local
python .\startup_ai_lab.py self-analyze --provider claude_cli --apply-candidates
```

`auto-loop` ya incluye self-analysis salvo que uses `--skip-self-analysis`.
Este es el modo mas parecido al patron `autoresearch`: el sistema investiga,
registra lo aprendido, se hace nuevas preguntas, se autoevalua y vuelve a
priorizar.

Si no queres hacer entrevistas externas todavia, el agente puede generar
preguntas para el fundador y guardar tus respuestas como memoria:

```powershell
python .\startup_ai_lab.py founder-questions --provider local --apply --limit 7
python .\startup_ai_lab.py founder-answer --id FQ-20260520-001 --answer "..."
python .\startup_ai_lab.py brief
```

Estas respuestas alimentan el proximo `self-analyze` y ayudan a decidir sin que
el agente invente preferencias, restricciones o conocimiento que solo tiene el
fundador.

Domain bundles disponibles:

- `official_ar`: fuentes oficiales argentinas.
- `academic`: PubMed, arXiv, OPS/OMS.
- `healthtech`: competidores y analogos.
- `all`: internet abierto.

Los reportes se guardan en `data/reports/` y las corridas en
`data/auto_runs.csv`. Si OpenAI o Claude fallan, `auto-research` deja un
reporte local de fallback en lugar de cortar solo con error.

## Registrar entrevistas

```powershell
python .\startup_ai_lab.py interview --kind profesional --segment psicologia --name "Lic. ejemplo" --pain 5 --urgency 4 --willingness 4 --notes "Cobra tarde, usa agenda manual, quiere pacientes privados."
```

```powershell
python .\startup_ai_lab.py interview --kind paciente --segment salud_mental --name "Paciente 1" --pain 4 --urgency 5 --willingness 3 --notes "No sabe si necesita psicologo o psiquiatra. Le importa saber precio antes."
```

## Registrar evidencia

```powershell
python .\startup_ai_lab.py evidence --claim "Salud mental tiene alta recurrencia" --source "entrevistas propias" --confidence 4 --notes "Tres pacientes mencionaron frecuencia semanal o quincenal."
```

## Registrar una fuente

```powershell
python .\startup_ai_lab.py source --id "SRC-SSS-RECLAMOS" --category "reclamos" --source-type "oficial" --title "Reclamos SSS" --url "https://..." --credibility 5 --finding "Los reintegros y coberturas aparecen como foco de friccion." --next-action "Cruzar con entrevistas."
```

## Registrar un competidor o analogo

```powershell
python .\startup_ai_lab.py competitor --name "Ejemplo" --geography "LatAm" --model "Marketplace de salud" --target "pacientes" --strengths "Buena UX" --weaknesses "No resuelve liquidacion" --relevance 4 --url "https://..."
```

## Registrar una norma

```powershell
python .\startup_ai_lab.py regulation --id "REG-XYZ" --norm "Norma ejemplo" --topic "Datos de salud" --impact "Define obligaciones del producto." --risk-level 5 --url "https://..."
```

## Registrar experimentos

```powershell
python .\startup_ai_lab.py experiment --name "Beta concierge salud mental" --hypothesis "Pacientes pagan si reciben orientacion y turno en menos de 72h" --metric "consultas pagas" --target "10 consultas en 14 dias"
```

## IA opcional

El script funciona sin IA. Si mas adelante queres sumar un modelo, podes definir:

```powershell
$env:OPENAI_API_KEY="tu_api_key"
$env:OPENAI_MODEL="gpt-5.2"
python .\startup_ai_lab.py brief --ai
```

Sin `OPENAI_API_KEY`, el brief se genera de forma local con reglas simples.

## Buscar papers y evidencia academica

Esto toma la idea de `autoresearch` como esqueleto: en vez de entrenar un modelo,
el script busca evidencia, la guarda y la mete en el brief.

```powershell
python .\startup_ai_lab.py paper-search --source pubmed --query "telemedicine mental health access appointment scheduling" --limit 5
python .\startup_ai_lab.py paper-search --source arxiv --query "clinical triage large language models safety" --limit 5
python .\startup_ai_lab.py brief
```

Fuentes soportadas en esta primera version:

- `pubmed`: papers biomedicos y de salud.
- `arxiv`: papers tecnicos, IA y computacion.
- `all`: consulta ambas fuentes.
