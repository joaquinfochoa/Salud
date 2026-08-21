# Scout plan - 2026-05-20

## Objetivo

Encontrar evidencia que reduzca incertidumbre antes de construir producto.

## Preguntas prioritarias

- MKT-001 / P5: Cuantas personas en AMBA tienen cobertura y sufren friccion de acceso?
- MKT-002 / P5: Cuanto gastan obras sociales/prepagas por salud mental, rehabilitacion y odontologia?
- USR-001 / P5: Pacientes con cobertura pagarian privado por turno rapido, precio claro y profesional verificado?
- PRO-001 / P5: Profesionales aceptan comision si la plataforma trae pacientes y cobra online?
- LEG-001 / P5: Que requisitos reales tiene operar como Red de Prestadores digital?
- LEG-002 / P5: Que oportunidades y limites abre Mi Argentina y la receta electronica para una plataforma privada de salud?

## Busquedas sugeridas

### Mercado y dinero

- "Superintendencia Servicios de Salud gasto prestaciones salud mental odontologia rehabilitacion"
- "INDEC cobertura salud obra social prepaga AMBA edad ingresos"
- "reclamos Superintendencia Servicios de Salud reintegros turnos copagos"
- "estado contable obra social prestaciones salud mental rehabilitacion odontologia"

### Papers y seguridad IA

python .\startup_ai_lab.py paper-search --source pubmed --query "digital triage primary care patient navigation" --limit 5
python .\startup_ai_lab.py paper-search --source pubmed --query "telemedicine mental health access waiting times" --limit 5
python .\startup_ai_lab.py paper-search --source arxiv --query "large language models clinical triage safety" --limit 5

### Competidores y analogos

- Doctoralia: estudiar modelo, pricing, onboarding y promesa. https://www.doctoralia.com.ar/
- Headway: estudiar modelo, pricing, onboarding y promesa. https://headway.co/
- Docturno: estudiar modelo, pricing, onboarding y promesa. https://www.docturno.com/
- Zocdoc: estudiar modelo, pricing, onboarding y promesa. https://www.zocdoc.com/
- Alma: estudiar modelo, pricing, onboarding y promesa. https://helloalma.com/

### Regulacion critica

- Ley 25.326 / Datos personales sensibles: Datos de salud requieren consentimiento y controles fuertes. https://www.argentina.gob.ar/normativa/nacional/64790/actualizacion
- Ley 26.529 / Derechos del paciente e historia clinica: Define consentimiento, confidencialidad e historia clinica. https://www.argentina.gob.ar/normativa/nacional/160432/actualizacion
- Leyes 23.660 y 23.661 / Obras sociales y Sistema Nacional del Seguro de Salud: Marco de financiadores, prestadores y Fondo Solidario. https://www.sssalud.gob.ar/index/normativas/consulta/000437.pdf
- Disposicion 1/2025 DNSISA / CUIR y visualizacion en Mi Argentina: Define CUIR, datos minimos enviados al Ministerio y endpoint de repositorios para visualizacion de recetas en Mi Argentina. https://www.argentina.gob.ar/normativa/nacional/disposici%C3%B3n-1-2025-415504/texto
- Ley 27.553 / Receta electronica y teleasistencia: Marco para teleasistencia y recetas digitales/electronicas. https://www.argentina.gob.ar/normativa/nacional/ley-27553-340919/texto

## Fuentes ya cargadas para no perder

- INDEC Censo 2022 - cobertura de salud: Argentina tenia 45,9M personas; 60,9% con obra social/prepaga/PAMI y 35,8% solo sistema publico.
- INDEC EPH primer semestre 2025 - cobertura urbana: En 31 aglomerados, 66,5% tiene obra social/prepaga/mutual/emergencia.
- SSS poblacion beneficiaria por provincia: Al 01-04-2026, SSS informa 20,25M beneficiarios; Buenos Aires + CABA concentran mas de 10,5M.
- Informe gasto en salud Argentina 2022: Gasto total estimado USD 66.367M; obras sociales USD 21.203M; gasto privado USD 27.740M.
- Resolucion 1430/2010 - plan de cuentas Agentes del Seguro de Salud: El plan de cuentas desagrega gasto en salud mental, rehabilitacion, odontologia, alta complejidad, medicamentos y otros rubros.

## Checklist de esta semana

- Cargar 5 entrevistas de pacientes.
- Cargar 5 entrevistas de profesionales.
- Guardar 5 fuentes nuevas sobre dinero/coberturas.
- Revisar 3 competidores/analogos y registrar huecos.
- Convertir 1 hallazgo en experimento concreto.
