/**
 * El nombre para mostrar de cada especialidad.
 *
 * Vive acá y no en el contrato porque los valores del dominio no llevan
 * acentos ni ñ —es la regla de todo identificador del proyecto— y "Alergia e
 * inmunología" no se deduce de "alergia-e-inmunologia" sin una tabla.
 *
 * El orden es alfabético por el nombre visible, que es el orden en que
 * aparecen en el select del alta y en el filtro de la búsqueda. Ordenar por la
 * clave pondría "clinica-medica" antes que "cardiologia".
 */
export const ESPECIALIDADES: Record<string, string> = {
  "alergia-e-inmunologia": "Alergia e inmunología",
  cardiologia: "Cardiología",
  "clinica-medica": "Clínica médica",
  dermatologia: "Dermatología",
  endocrinologia: "Endocrinología",
  enfermeria: "Enfermería",
  fonoaudiologia: "Fonoaudiología",
  gastroenterologia: "Gastroenterología",
  ginecologia: "Ginecología",
  kinesiologia: "Kinesiología",
  neumonologia: "Neumonología",
  neurologia: "Neurología",
  nutricion: "Nutrición",
  obstetricia: "Obstetricia",
  odontologia: "Odontología",
  oftalmologia: "Oftalmología",
  otorrinolaringologia: "Otorrinolaringología",
  pediatria: "Pediatría",
  podologia: "Podología",
  psicologia: "Psicología",
  psicopedagogia: "Psicopedagogía",
  psiquiatria: "Psiquiatría",
  reumatologia: "Reumatología",
  "terapia-ocupacional": "Terapia ocupacional",
  traumatologia: "Traumatología",
  urologia: "Urología",
};
