export const PROFESIONALES = [
  {
    id: '1',
    slug: 'martin-gonzalez',
    nombre: 'Martín González',
    especialidad: 'Psicología',
    matricula: 'MN 98.234',
    descripcion: 'Psicólogo clínico con orientación cognitivo-conductual. Atiendo adultos y adolescentes con ansiedad, depresión y crisis vitales. Más de 8 años de experiencia.',
    rating: 4.8,
    totalReseñas: 47,
    precio: 12000,
    modalidades: ['telemedicina', 'presencial'],
    zona: 'CABA',
    obrasSociales: ['OSDE', 'Swiss Medical', 'Galeno', 'Medifé'],
    coseguros: { OSDE: 2800, 'Swiss Medical': 3200, Galeno: 2400, Medifé: 2600 },
    proximoDisponible: 'Mañana 09:00',
    horarios: ['09:00', '09:50', '10:40', '14:00', '15:00', '16:00'],
    reseñas: [
      { autor: 'Ana R.', nota: 5, texto: 'Excelente profesional. Me ayudó mucho a manejar la ansiedad.', fecha: 'Hace 3 días' },
      { autor: 'Carlos M.', nota: 5, texto: 'Muy empático y puntual. Las sesiones online funcionan muy bien.', fecha: 'Hace 1 semana' },
      { autor: 'Lucía P.', nota: 4, texto: 'Buen profesional, se nota la experiencia.', fecha: 'Hace 2 semanas' },
    ],
  },
  {
    id: '2',
    slug: 'carolina-vega',
    nombre: 'Carolina Vega',
    especialidad: 'Psicología',
    matricula: 'MN 112.087',
    descripcion: 'Especializada en terapia sistémica y de parejas. Trabajo con adultos en procesos de cambio, duelos y conflictos relacionales.',
    rating: 4.9,
    totalReseñas: 83,
    precio: 14000,
    modalidades: ['telemedicina'],
    zona: 'GBA Norte',
    obrasSociales: ['OSDE', 'OMINT', 'Swiss Medical', 'Sanitas'],
    coseguros: { OSDE: 3200, OMINT: 3800, 'Swiss Medical': 3600, Sanitas: 2900 },
    proximoDisponible: 'Hoy 18:00',
    horarios: ['18:00', '19:00', '20:00'],
    reseñas: [
      { autor: 'Sofía T.', nota: 5, texto: 'La mejor terapeuta que tuve. 100% recomendable.', fecha: 'Hace 5 días' },
      { autor: 'Diego F.', nota: 5, texto: 'Muy profesional y cálida. Me ayudó muchísimo.', fecha: 'Hace 2 semanas' },
    ],
  },
  {
    id: '3',
    slug: 'pablo-moreno',
    nombre: 'Pablo Moreno',
    especialidad: 'Kinesiología / Fisioterapia',
    matricula: 'MN 45.321',
    descripcion: 'Kinesiólogo especializado en traumatología deportiva y rehabilitación postquirúrgica. Atiendo a domicilio y en consultorio.',
    rating: 4.7,
    totalReseñas: 31,
    precio: 9500,
    modalidades: ['presencial', 'domicilio'],
    zona: 'CABA',
    obrasSociales: ['OSDE', 'Galeno', 'IOMA', 'PAMI'],
    coseguros: { OSDE: 1200, Galeno: 900, IOMA: 800, PAMI: 0 },
    proximoDisponible: 'Hoy 15:30',
    horarios: ['15:30', '16:30', '17:30'],
    reseñas: [
      { autor: 'Valentina R.', nota: 5, texto: 'Me operé la rodilla y la rehabilitación fue perfecta.', fecha: 'Hace 1 semana' },
      { autor: 'Marcos L.', nota: 4, texto: 'Muy bueno, viene a domicilio sin problema.', fecha: 'Hace 3 semanas' },
    ],
  },
  {
    id: '4',
    slug: 'gabriela-rios',
    nombre: 'Gabriela Ríos',
    especialidad: 'Odontología',
    matricula: 'MN 67.890',
    descripcion: 'Odontóloga general con especialización en estética dental. Trabajo con materiales de primera calidad en un consultorio moderno en Palermo.',
    rating: 4.6,
    totalReseñas: 58,
    precio: 15000,
    modalidades: ['presencial'],
    zona: 'CABA',
    obrasSociales: ['OSDE', 'Swiss Medical', 'Galeno', 'Medifé', 'OMINT'],
    coseguros: { OSDE: 0, 'Swiss Medical': 1500, Galeno: 0, Medifé: 0, OMINT: 2000 },
    proximoDisponible: 'Pasado mañana 10:00',
    horarios: ['10:00', '11:00', '12:00', '16:00', '17:00'],
    reseñas: [
      { autor: 'Romina V.', nota: 5, texto: 'Excelente. Por primera vez no le tuve miedo al dentista.', fecha: 'Hace 4 días' },
      { autor: 'Guillermo P.', nota: 4, texto: 'Muy profesional y el consultorio es muy cómodo.', fecha: 'Hace 2 semanas' },
    ],
  },
]

export const OBRAS_SOCIALES_LISTA = [
  'OSDE', 'Swiss Medical', 'Galeno', 'Medifé', 'OMINT', 'IOMA', 'PAMI', 'Sanitas', 'Sin obra social'
]

export const ESPECIALIDADES_LISTA = [
  'Psicología', 'Kinesiología / Fisioterapia', 'Odontología'
]

export const SINTOMAS = [
  { sintoma: 'Ansiedad o ataques de pánico', especialidad: 'Psicología' },
  { sintoma: 'Tristeza o depresión',          especialidad: 'Psicología' },
  { sintoma: 'Estrés laboral',                especialidad: 'Psicología' },
  { sintoma: 'Dolor de espalda o columna',    especialidad: 'Kinesiología / Fisioterapia' },
  { sintoma: 'Dolor de rodilla',              especialidad: 'Kinesiología / Fisioterapia' },
  { sintoma: 'Recuperación postquirúrgica',   especialidad: 'Kinesiología / Fisioterapia' },
  { sintoma: 'Dolor de muela o encías',       especialidad: 'Odontología' },
  { sintoma: 'Limpieza dental',               especialidad: 'Odontología' },
  { sintoma: 'Blanqueamiento',                especialidad: 'Odontología' },
]
