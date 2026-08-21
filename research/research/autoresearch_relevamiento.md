# Relevamiento de karpathy/autoresearch aplicado a la startup

Fecha de corte: 2026-05-18

Repositorio: https://github.com/karpathy/autoresearch

## Que es autoresearch

`karpathy/autoresearch` es un experimento para que agentes de IA hagan
investigacion autonoma sobre entrenamiento de LLMs.

La idea central no es una libreria compleja. Es un sistema de trabajo:

1. Un humano escribe un `program.md` con instrucciones del "laboratorio".
2. El agente modifica un unico archivo permitido (`train.py`).
3. Ejecuta un experimento con tiempo fijo.
4. Mide una metrica objetiva (`val_bpb`, menor es mejor).
5. Si mejora, conserva el cambio.
6. Si empeora o rompe, descarta y prueba otra cosa.
7. Registra cada intento en un log.

## Que partes importan para nosotros

No nos interesa copiar el entrenamiento de modelos ni requerir GPU. Nos interesa
copiar la arquitectura mental:

- instrucciones vivas en Markdown;
- alcance acotado;
- ciclos autonomos;
- una pregunta clara por ciclo;
- fuentes/logs persistentes;
- criterio para avanzar o descartar;
- humano define estrategia, IA ejecuta investigacion.

## Traduccion a nuestro proyecto

| Autoresearch original | Startup healthtech |
|---|---|
| `program.md` define el laboratorio | `program.md` define tesis, riesgos y reglas de decision |
| `train.py` es el archivo editable | preguntas/fuentes/briefs son la superficie de trabajo |
| metrica `val_bpb` | calidad de evidencia, decision accionable, reduccion de incertidumbre |
| corre `uv run train.py` | corre `auto-research` con web search + IA |
| guarda `results.tsv` | guarda reportes y `auto_runs.csv` |
| keep/discard cambios | avanzar/descartar hipotesis |

## Como debe funcionar nuestro agente

Entrada:

- una pregunta de investigacion;
- contexto del proyecto;
- fuentes conocidas;
- dominio de busqueda;
- restricciones de seguridad/legal.

Proceso:

- buscar en internet;
- leer fuentes;
- sintetizar hallazgos;
- citar URLs;
- decir que falta probar;
- proponer un experimento.

Salida:

- reporte Markdown;
- fuentes usadas;
- registro de corrida;
- proximo paso sugerido.

## Retroalimentacion

El comando `reflect` agrega una capa de meta-investigacion:

- lee reportes recientes;
- genera nuevas preguntas si detecta incertidumbres no resueltas;
- propone experimentos de bajo costo;
- puede aplicar esas filas a los CSV con `--apply`;
- deja un reporte `reflection_*.md` para auditar que cambio y por que.

El comando `auto-loop` encadena investigacion, reflexion aplicada y brief. Esa
es la version mas cercana al ciclo autoresearch: el sistema no solo responde,
tambien decide que investigar despues.

El comando `self-analyze` agrega una segunda capa: evalua el comportamiento del
propio agente. Detecta si hay exceso de research, falta de respuestas del
fundador, decisiones candidatas maduras o riesgos que siguen abiertos. Guarda reportes
`self_analysis_*.md` y puede persistir candidatos en `decisions.csv`.

El comando `founder-questions` agrega un canal de retroalimentacion humano:
convierte decisiones candidatas y dudas abiertas en preguntas concretas para el
fundador. Las respuestas se guardan con `founder-answer` y pasan a formar parte
del brief y del self-analysis.

## Que NO hace todavia

- No toma decisiones legales definitivas.
- No reemplaza entrevistas reales.
- No firma convenios ni verifica datos privados.
- No ejecuta pagos ni maneja datos sensibles.
- No puede hacer busqueda web autonoma sin OpenAI o Claude; si esos proveedores
  fallan, solo genera un reporte local con la memoria ya cargada.

## Por que esto sirve

El fundador esta solo. El agente tiene que actuar como un analista que trabaja en
segundo plano:

- busca datos de mercado;
- revisa leyes;
- compara competidores;
- encuentra papers;
- propone experimentos;
- convierte todo en briefs.

Primero usamos IA para construir la startup. Despues decidimos donde IA entra en
el producto.
