# Auto research - LEG-002

## Veredicto

Reporte local preliminar: hay material cargado para orientar la decision, pero falta una busqueda web nueva antes de cerrar la pregunta.

## Hallazgos con fuentes

- Ministerio de Salud - Receta Electronica en Mi Argentina: Mi Argentina permite consultar prescripciones electronicas de medicamentos emitidas por plataformas ReNaPDiS durante los 60 dias previos, con autorizacion del usuario. (https://www.argentina.gob.ar/noticias/la-receta-electronica-se-incorpora-la-app-mi-argentina)
- Ministerio de Salud - Receta Electronica: La receta electronica es la modalidad vigente para medicamentos y las plataformas deben estar inscriptas y aprobadas por ReNaPDiS. (https://www.argentina.gob.ar/salud/digital/renapdis/receta-electronica)
- INDEC Censo 2022 - cobertura de salud: Argentina tenia 45,9M personas; 60,9% con obra social/prepaga/PAMI y 35,8% solo sistema publico. (https://biblioteca.indec.gob.ar/bases/minde/1c2022_2.pdf)
- Informe gasto en salud Argentina 2022: Gasto total estimado USD 66.367M; obras sociales USD 21.203M; gasto privado USD 27.740M. (https://www.argentina.gob.ar/sites/default/files/2025/05/informe_gasto_en_salud_2022_dnesa.pdf)
- Resolucion 1430/2010 - plan de cuentas Agentes del Seguro de Salud: El plan de cuentas desagrega gasto en salud mental, rehabilitacion, odontologia, alta complejidad, medicamentos y otros rubros. (https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1430-2010-176808/texto)
- SSS poblacion beneficiaria por provincia: Al 01-04-2026, SSS informa 20,25M beneficiarios; Buenos Aires + CABA concentran mas de 10,5M. (https://www.sssalud.gob.ar/index.php?cat=consultas&page=poblacion)
- Resolucion 5744/2024 / Repositorios de receta electronica: Exige interoperabilidad y APIs para que plataformas y farmacias accedan a repositorios segun cobertura del paciente. (https://www.argentina.gob.ar/normativa/nacional/406757/texto)
- Resolucion 2214/2025 / Extension de receta electronica: Extiende la receta electronica a indicaciones medicas, practicas, procedimientos y dispositivos. (https://www.argentina.gob.ar/normativa/nacional/415349/texto)
- Ley 25.326 / Datos personales sensibles: Datos de salud requieren consentimiento y controles fuertes. (https://www.argentina.gob.ar/normativa/nacional/64790/actualizacion)
- Ley 27.553 / Receta electronica y teleasistencia: Marco para teleasistencia y recetas digitales/electronicas. (https://www.argentina.gob.ar/normativa/nacional/ley-27553-340919/texto)
- Leyes 23.660 y 23.661 / Obras sociales y Sistema Nacional del Seguro de Salud: Marco de financiadores, prestadores y Fondo Solidario. (https://www.sssalud.gob.ar/index/normativas/consulta/000437.pdf)
- Disposicion 1/2025 DNSISA / CUIR y visualizacion en Mi Argentina: Define CUIR, datos minimos enviados al Ministerio y endpoint de repositorios para visualizacion de recetas en Mi Argentina. (https://www.argentina.gob.ar/normativa/nacional/disposici%C3%B3n-1-2025-415504/texto)

## Implicancias para la startup

- Tratar cualquier dato de salud como sensible: consentimiento, minimizacion, trazabilidad y revision legal antes de integrar.
- Usar las fuentes oficiales cargadas para definir alcance del MVP y separar lo posible ahora de lo que requiere integracion formal.
- Comparar el hueco competitivo contra analogos antes de construir funcionalidad propia.
- Convertir la pregunta en una prueba chica: una landing, entrevista, pedido de datos o consulta legal concreta.

## Riesgos e incertidumbres

- Este modo local no navega internet ni verifica cambios recientes.
- Las fuentes cargadas pueden estar incompletas o desactualizadas.
- No reemplaza validacion legal, clinica ni entrevistas con usuarios reales.

## Proximo experimento

Correr la misma pregunta con `--provider openai` o `--provider claude_cli` y contrastar el reporte con una entrevista a un profesional y una consulta legal breve.

## Fuentes consultadas

- https://www.argentina.gob.ar/noticias/la-receta-electronica-se-incorpora-la-app-mi-argentina
- https://www.argentina.gob.ar/salud/digital/renapdis/receta-electronica
- https://biblioteca.indec.gob.ar/bases/minde/1c2022_2.pdf
- https://www.argentina.gob.ar/sites/default/files/2025/05/informe_gasto_en_salud_2022_dnesa.pdf
- https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1430-2010-176808/texto
- https://www.sssalud.gob.ar/index.php?cat=consultas&page=poblacion
- https://www.argentina.gob.ar/normativa/nacional/406757/texto
- https://www.argentina.gob.ar/normativa/nacional/415349/texto
- https://www.argentina.gob.ar/normativa/nacional/64790/actualizacion
- https://www.argentina.gob.ar/normativa/nacional/ley-27553-340919/texto
- https://www.sssalud.gob.ar/index/normativas/consulta/000437.pdf
- https://www.argentina.gob.ar/normativa/nacional/disposici%C3%B3n-1-2025-415504/texto
- https://www.doctoralia.com.ar/
- https://headway.co/
- https://www.docturno.com/
- https://helloalma.com/

## Configuracion

- Dominio pedido: official_ar
- Proveedor: local

