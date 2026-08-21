# Analisis inicial del mercado de salud en Argentina

Fecha de corte: 2026-05-18

Este documento es una primera base de trabajo para entender el terreno real:
tamanio de poblacion, cobertura, demanda de atencion, financiadores,
regulacion, barreras y oportunidad para una plataforma de salud digital.

## 1. Foto poblacional y cobertura

### Censo 2022 - total pais

Poblacion censada definitiva: 45.892.285 personas.

Distribucion de cobertura de salud:

- Obra social o prepaga, incluyendo PAMI: 27.787.124 personas, 60,9%.
- Programas o planes estatales de salud: 1.514.231 personas, 3,3%.
- Solo sistema publico: 16.317.432 personas, 35,8%.

Lectura para la startup:

- Hay un mercado masivo con cobertura formal, pero eso no significa acceso
  eficiente.
- Hay tambien una poblacion enorme sin cobertura formal, pero probablemente con
  menor capacidad de pago directo.
- La oportunidad inicial no es "reemplazar el sistema publico"; es resolver
  fricciones de personas con algun poder de pago y/o cobertura que igual no
  logran navegar bien el sistema.

Fuente:
https://biblioteca.indec.gob.ar/bases/minde/1c2022_2.pdf

### EPH - 31 aglomerados urbanos, primer semestre 2025

Total personas relevadas/proyectadas en 31 aglomerados: 29.900.963.

Cobertura medica:

- Obra social, prepaga, mutual o servicio de emergencia: 66,5%.
- Solo sistema publico: 33,3%.

Por edad:

- Hasta 17 anios: 56,1% con cobertura; 43,7% solo publico.
- 18 a 64 anios: 64,7% con cobertura; 35,1% solo publico.
- 65 anios y mas: 97,5% con cobertura; 2,5% solo publico.

Lectura para la startup:

- El adulto mayor esta fuertemente cubierto, principalmente por PAMI u otros
  esquemas.
- Ninios y jovenes tienen mucha mayor exposicion al sistema publico.
- El segmento 18-64 es clave: concentra decision, pago, empleo formal,
  monotributo, prepagas y dolor de acceso.

Fuente:
https://www.indec.gob.ar/uploads/informesdeprensa/eph_indicadores_hogares_11_2514558E0A45.pdf

## 2. Seguridad social y prepagas bajo Superintendencia

La Superintendencia de Servicios de Salud informa, al 01-04-2026, un total de
20.250.487 beneficiarios por provincia, con 14.826.033 titulares y 5.424.454
familiares.

Principales concentraciones:

- Buenos Aires: 8.190.490 beneficiarios.
- CABA: 2.346.206 beneficiarios.
- Cordoba: 1.711.783 beneficiarios.
- Santa Fe: 1.593.274 beneficiarios.
- Mendoza: 823.478 beneficiarios.

Lectura para la startup:

- CABA + Buenos Aires concentran una masa critica suficiente para empezar.
- El primer mercado geografico razonable es AMBA, no "Argentina".
- El acceso a financiadores y cartillas tiene valor porque el padron bajo
  regulacion es enorme.

Fuente:
https://www.sssalud.gob.ar/index.php?cat=consultas&page=poblacion

## 3. Demanda real de consultas

No hay una fuente nacional reciente y simple para todas las consultas privadas y
publicas. Como proxy fuerte, se usa la Encuesta Provincial de Acceso, Uso y
Satisfaccion con los Servicios de Salud de Provincia de Buenos Aires (EPAUSS
2023), por ser una provincia enorme y con muestra grande.

Resultados clave PBA:

- 61,1% de la poblacion solicito consultas medicas en los ultimos 12 meses.
- Eso representa 11.015.809 bonaerenses.
- Mujeres: 66,5%; varones: 55,4%.
- Personas con obra social/prepaga: 66,3%.
- Personas con cobertura publica exclusiva: 52,0%.

Motivo principal de consulta:

- Control de salud: 55,1%.
- Dolor o malestar fisico: 21,0%.
- Seguimiento o continuacion de tratamiento: 6,8%.
- Apto medico/fisico: 6,5%.
- Accidente o lesion: 6,1%.
- Rehabilitacion o terapia: 1,0%.
- Malestar emocional: 0,9%.

Lectura para la startup:

- El mercado no es solo urgencia. La mayoria de consultas son planificables.
- Hay espacio para agenda, orientacion, seguimiento y derivacion.
- La brecha entre cobertura privada/social y cobertura publica exclusiva
  sugiere que tener cobertura aumenta uso, pero no necesariamente calidad de
  acceso.
- Salud mental aparece chica en "motivo principal" de consulta general, pero
  puede estar subregistrada o canalizada por circuitos especificos. Hay que
  validarla con datos propios.

Fuente:
https://www.ioma.gba.gob.ar/wp-content/uploads/2024/07/Informe-EPAUSS-Ministerio-de-Salud-1.pdf

## 4. Gasto en salud y flujo de dinero

Entre 2017 y 2022, Argentina exhibio un gasto total en salud entre 10% y 11% del
PIB.

En 2022, el gasto en salud fue estimado en USD 66.367 millones, compuesto por:

- Gasto publico: USD 17.425 millones.
- Obras sociales: USD 21.203 millones.
- Gasto privado: USD 27.740 millones.

El informe destaca que desde 2018 el gasto privado asume el rol de principal
financiador, salvo en 2020. El gasto privado incluye cuotas de medicina prepaga,
pagos complementarios, pagos directos, copagos y coseguros.

Lectura para la startup:

- Hay mucho dinero moviendose fuera del esquema puro de obra social.
- El problema de la plataforma no es solo "turnos"; es informacion + pago +
  cobertura + administracion.
- Si el producto ayuda a decidir entre cobertura, pago privado, copago o
  reintegro, toca un flujo economico real.

Fuente:
https://www.argentina.gob.ar/sites/default/files/2025/05/informe_gasto_en_salud_2022_dnesa.pdf

## 5. Profesionales y verificacion

En abril de 2026, el Ministerio de Salud simplifico el acceso al Buscador
Nacional de Profesionales de la Salud basado en REFEPS.

El buscador informa datos de mas de 1.200.000 profesionales habilitados y permite
consultar nombre, DNI o matricula, habilitaciones vigentes, especialidades e
inhabilitaciones.

Lectura para la startup:

- La verificacion de profesionales no tiene que empezar manualmente desde cero.
- La matricula verificada puede ser una pieza central de confianza.
- Toda orientacion, agenda o cobro deberia depender de un profesional verificado.

Fuente:
https://www.argentina.gob.ar/noticias/salud-simplifica-el-acceso-al-buscador-nacional-de-profesionales-de-salud

## 6. Regulacion relevante

### Obras sociales, prepagas y financiadores

- Ley 23.660: regimen de obras sociales.
- Ley 23.661: Sistema Nacional del Seguro de Salud.
- Ley 26.682: marco regulatorio de medicina prepaga.
- Resolucion 1/2025: derivacion directa de aportes y contribuciones a la entidad
  contratada por el beneficiario.
- Resolucion 1926/2024: deja sin efecto aranceles vigentes de coseguros del PMO
  y permite que sean fijados libremente por entidades comprendidas.
- Resolucion 645/2025: prepagas deben informar mensualmente precios de planes.
- DNU 70/2023: marco de desregulacion que impacta en planes y coseguros.

Fuentes:
https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1-2025-408968
https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1926-2024-400745
https://www.argentina.gob.ar/sssalud/valores-de-planes

### Salud digital y datos

- Ley 27.553 y Decreto 98/2023: receta electronica/digital y teleasistencia.
- Resolucion 1959/2024: crea ReNaPDiS.
- Resolucion 2214/2025: lineamientos, estandares tecnicos y CUIR.
- Disposicion 1/2025: requisitos complementarios para sistemas de receta
  electronica, validaciones con registros nacionales.
- Ley 25.326: proteccion de datos personales; salud es dato sensible.
- Ley 26.529: derechos del paciente, historia clinica y consentimiento informado.

Fuentes:
https://www.argentina.gob.ar/normativa/nacional/ley-27553-340919/texto
https://www.argentina.gob.ar/normativa/nacional/400825/texto
https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-2214-2025-415349/texto
https://www.argentina.gob.ar/normativa/nacional/disposici%C3%B3n-1-2025-415504/texto
https://www.argentina.gob.ar/normativa/nacional/64790/actualizacion
https://www.argentina.gob.ar/normativa/nacional/160432/actualizacion

### Red de prestadores

La inscripcion de redes de prestadores ante SSS exige contrato constitutivo,
listado de prestadores, certificados de inscripcion en Registro de Prestadores y
actas de adhesion.

El Anexo IV tambien indica que, para ciertas figuras asociativas, debe quedar
establecida responsabilidad frente a Agentes del Seguro de Salud por actos u
omisiones realizados con motivo de los convenios de la Red.

Fuente:
https://www.sssalud.gob.ar/prestadores/formularios/prest_red_anIV.pdf

## 7. Barreras principales

1. Fragmentacion institucional: publico, seguridad social, prepagas, provincias,
   colegios profesionales y prestadores.
2. Acceso a financiadores: los convenios son la barrera real de largo plazo.
3. Confianza: salud exige verificacion de matricula, reputacion y claridad del
   rol de la plataforma.
4. Datos sensibles: el producto no puede tratar datos de salud como una app
   comun.
5. Responsabilidad: si la plataforma orienta, cobra, agenda y administra, no es
   un directorio pasivo.
6. Liquidez: pagar al profesional en el dia antes de cobrar al financiador exige
   capital de trabajo.
7. Marketplace: hay que tener densidad por vertical y zona.
8. Regulacion cambiante: 2024-2026 muestra un sistema en reordenamiento.

## 8. Oportunidad preliminar

La oportunidad no parece ser "mas turnos online". La oportunidad aparece en la
interseccion de:

- poblacion con cobertura que igual tiene friccion de acceso;
- profesionales que quieren pacientes y mejor cobro;
- gasto privado creciente y poco transparente;
- regulacion que esta empujando mas transparencia de precios;
- herramientas publicas de verificacion profesional;
- necesidad de orientacion antes de consumir una prestacion.

Hipotesis de posicionamiento:

La plataforma no debe venderse primero como cartilla. Debe venderse como capa de
decision y ejecucion: "decidi a quien consultar, sabe cuanto te cuesta, consegui
turno y deja el pago/comprobante resuelto".

## 9. Proximas busquedas

Datos faltantes:

- Cantidad nacional reciente de consultas ambulatorias publicas + privadas.
- Tiempos de espera por especialidad en obras sociales/prepagas.
- Ticket promedio privado por especialidad: psicologia, kinesiologia,
  odontologia, clinica.
- Cantidad de profesionales activos por especialidad y provincia.
- Precios reales de planes de prepagas por edad/region desde la herramienta SSS.
- Reclamos ante SSS por tipo: turnos, cobertura, reintegros, copagos, negativas.
- Evidencia sobre orientacion digital/triage no diagnostico y reduccion de
  consultas innecesarias.
- Experiencias comparables: Zocdoc, Headway, Alma, Grow Therapy, Doctoralia,
  Docto, Docturno.

## 10. Dinero de obras sociales por tipo de cobertura

Pregunta: se puede ver cuanto dinero destinan las obras sociales a distintos
tipos de cobertura?

Respuesta corta: parcialmente si. El dato existe dentro del sistema, pero no
todo esta publicado como una tabla publica limpia y reutilizable.

### Lo que si se puede ver publicamente

1. Transferencias y subsidios desde SSS / FSR.

La Superintendencia publica transferencias recurrentes y no recurrentes:

- Sistema Unico de Reintegros (SUR + SURGE): reintegros a obras sociales por
  prestaciones de alto impacto economico, baja incidencia y tratamientos
  prolongados.
- Mecanismo de Integracion: financiamiento directo de prestaciones por
  discapacidad.
- SANO: subsidio automatico nominativo ajustado por edad, sexo e ingreso
  promedio del grupo familiar.
- SUMA, SUMARTE, SUMA 65, oficios judiciales, HPGD y otros esquemas.

Esto permite analizar una parte muy sensible del sistema: alto costo,
discapacidad y compensaciones entre obras sociales. No muestra todo el gasto
normal en consultas ambulatorias.

Fuentes:
https://www.argentina.gob.ar/sssalud/transparencia/subsidios
https://www.argentina.gob.ar/sssalud/transparencia/subsidios/sur
https://www.argentina.gob.ar/sssalud/transparencia/subsidios/mecanismo-integracion
https://www.argentina.gob.ar/sssalud/transparencia/subsidios/sano

2. Matriz SANO.

La matriz SANO permite ver cuanto se transfiere por capita segun grupo etario y
sexo. La pagina de SSS publica valores por grupo:

- 0 a 14
- 15 a 49
- 50 a 64
- 65 o mas

Esto no es "gasto por prestacion", pero sirve para entender como el sistema
compensa riesgo poblacional.

Fuente:
https://www.argentina.gob.ar/sssalud/transparencia/subsidios/sano

3. Gasto agregado por financiador.

El informe de gasto en salud muestra gasto publico, obras sociales y gasto
privado. En 2022 estima:

- Gasto publico: USD 17.425 millones.
- Obras sociales: USD 21.203 millones.
- Gasto privado: USD 27.740 millones.
- Total salud: USD 66.367 millones.

Esto sirve para dimensionar el mercado, pero no baja al detalle de "odontologia"
o "salud mental".

Fuente:
https://www.argentina.gob.ar/sites/default/files/2025/05/informe_gasto_en_salud_2022_dnesa.pdf

### Lo que el sistema exige, pero no siempre publica

1. Ley 23.660.

Las obras sociales deben presentar ante la autoridad:

- programa de prestaciones medico-asistenciales;
- presupuesto de gastos y recursos;
- memoria y balance;
- contratos de prestaciones.

Ademas, deben destinar como minimo el 80% de sus recursos brutos, deducido el
aporte al Fondo Solidario de Redistribucion, a prestaciones de salud.

Fuente:
https://www.sssalud.gob.ar/index/normativas/consulta/000437.pdf

2. Resolucion 1430/2010 - plan de cuentas.

Esta resolucion aprueba un plan de cuentas para Agentes del Seguro de Salud.
Lo mas importante para nuestra investigacion es que desagrega gastos en
prestaciones medico-asistenciales. En servicios contratados aparecen cuentas
como:

- convenios globales;
- atencion medica primaria;
- programas de prevencion;
- atencion secundaria;
- salud mental;
- rehabilitacion;
- odontologia;
- otras coberturas;
- optica;
- protesis y ortesis;
- servicios de ambulancia;
- alta complejidad;
- medicamentos cronicos;
- resto de medicamentos.

Esto significa que, si se accede a estados contables detallados o anexos
presentados ante SSS, se podria estimar gasto por tipo de cobertura.

Fuente:
https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1430-2010-176808/texto

3. Resolucion 650/1997 - estadisticas prestacionales.

Las obras sociales deben presentar estadisticas de prestaciones medicas. Esta
normativa apunta mas a cantidades/uso que a dinero, pero ayuda a cruzar gasto
con volumen: egresos, practicas ambulatorias, laboratorio, radiologia, etc.

Fuente:
https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-650-1997-42684/actualizacion

### Lectura para la startup

Para nuestro proyecto, la pregunta clave no es solo "cuanto gastan", sino:

- que rubros son caros para las obras sociales;
- que rubros generan mas reclamos o judicializacion;
- que rubros tienen baja transparencia para el paciente;
- que rubros tienen profesionales con demora de cobro;
- que rubros pueden moverse a un modelo mas eficiente de pago directo,
  copago/reintegro o red digital.

Los rubros que mas nos interesan por el MVP son:

- salud mental;
- rehabilitacion/kinesiologia;
- odontologia;
- medicamentos cronicos;
- consultas ambulatorias y atencion primaria;
- reintegros y prestaciones fuera de cartilla.

### Proximo paso operativo

Armar un dataset con tres capas:

1. Transferencias publicas SSS:
   SUR/SURGE, Integracion, SANO, SUMA/SUMARTE.
2. Normativa contable:
   cuentas de gasto por cobertura segun Resolucion 1430/2010.
3. Solicitud de acceso a informacion publica:
   pedir a SSS estados contables/anexos agregados por rubro o, al menos,
   totales anonimizados por cuenta de gasto del plan de cuentas.

Esto nos permitiria inferir donde hay mas dinero, mas friccion y mas espacio
para una plataforma.
