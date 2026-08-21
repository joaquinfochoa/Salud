import { useState } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { ArrowLeft, Star, Video, MapPin, Home, CheckCircle } from 'lucide-react'
import { PROFESIONALES } from '../../data/profesionales'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

const DIAS_SEMANA = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb']

export default function PerfilProfesional() {
  const { slug } = useParams()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const obraSocialPaciente = searchParams.get('obraSocial') || ''

  const p = PROFESIONALES.find(x => x.slug === slug)
  const [diaIdx, setDiaIdx] = useState(0)
  const [horaSel, setHoraSel] = useState(null)

  if (!p) return (
    <div className="flex items-center justify-center min-h-screen">
      <p className="text-gray-400">Profesional no encontrado</p>
    </div>
  )

  const coseguro = p.coseguros?.[obraSocialPaciente]
  const tieneOS  = coseguro !== undefined

  function handleReservar() {
    if (!horaSel) return
    navigate(`/reserva/${p.id}?hora=${horaSel}&obraSocial=${obraSocialPaciente}`)
  }

  // Generar días de la semana desde hoy
  const hoy = new Date()
  const dias = Array.from({ length: 5 }, (_, i) => {
    const d = new Date(hoy)
    d.setDate(hoy.getDate() + i + 1)
    return d
  })

  return (
    <div className="min-h-screen bg-gray-50 max-w-lg mx-auto pb-28">
      {/* Header */}
      <header className="sticky top-0 bg-white border-b border-gray-100 z-30 px-4 h-14 flex items-center gap-3">
        <button onClick={() => navigate(-1)} className="p-1.5 -ml-1.5 rounded-xl text-gray-500 hover:bg-gray-100">
          <ArrowLeft size={20} />
        </button>
        <p className="text-sm font-semibold text-gray-900 flex-1 truncate">{p.nombre}</p>
      </header>

      {/* Hero del profesional */}
      <div className="bg-white px-4 pt-6 pb-5 flex flex-col items-center text-center border-b border-gray-100">
        <div className="w-20 h-20 rounded-full bg-sky-100 flex items-center justify-center mb-3">
          <span className="text-2xl font-bold text-sky-500">
            {p.nombre.split(' ').map(n => n[0]).join('').slice(0, 2)}
          </span>
        </div>
        <h1 className="text-lg font-bold text-gray-900">{p.nombre}</h1>
        <p className="text-sm text-gray-500 mt-0.5">{p.especialidad} · {p.matricula}</p>
        <div className="flex items-center gap-1.5 mt-2">
          <Star size={14} fill="#f59e0b" className="text-amber-400" />
          <span className="text-sm font-semibold text-gray-800">{p.rating}</span>
          <span className="text-sm text-gray-400">({p.totalReseñas} reseñas)</span>
        </div>

        {/* Modalidades */}
        <div className="flex items-center gap-3 mt-3">
          {p.modalidades.map(m => (
            <div key={m} className="flex items-center gap-1">
              {m === 'telemedicina' && <Video size={13} className="text-sky-400" />}
              {m === 'presencial'   && <MapPin size={13} className="text-gray-400" />}
              {m === 'domicilio'    && <Home size={13} className="text-emerald-400" />}
              <span className="text-xs text-gray-500 capitalize">{m === 'telemedicina' ? 'Online' : m}</span>
            </div>
          ))}
        </div>
      </div>

      <div className="px-4 py-4 flex flex-col gap-4">

        {/* Descripción */}
        <Card className="p-4">
          <p className="text-sm text-gray-700 leading-relaxed">{p.descripcion}</p>
        </Card>

        {/* Precio */}
        <Card className="p-4">
          <p className="text-xs text-gray-400 font-medium uppercase tracking-wide mb-3">Precio de la consulta</p>
          {tieneOS ? (
            <div className="flex flex-col gap-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500">Honorario total</span>
                <span className="text-sm font-medium text-gray-400 line-through">${p.precio.toLocaleString('es-AR')}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500">Cobertura {obraSocialPaciente}</span>
                <span className="text-sm text-sky-500 font-medium">
                  −${(p.precio - coseguro).toLocaleString('es-AR')}
                </span>
              </div>
              <div className="flex justify-between items-center border-t border-gray-100 pt-2">
                <span className="text-sm font-semibold text-gray-800">Pagás vos</span>
                <span className="text-lg font-bold text-emerald-600">
                  {coseguro === 0 ? 'Sin cargo' : `$${coseguro.toLocaleString('es-AR')}`}
                </span>
              </div>
              <p className="text-xs text-gray-400">El reintegro se gestiona automáticamente</p>
            </div>
          ) : (
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">Consulta privada</span>
              <span className="text-xl font-bold text-gray-900">${p.precio.toLocaleString('es-AR')}</span>
            </div>
          )}
        </Card>

        {/* Obras sociales */}
        <Card className="p-4">
          <p className="text-xs text-gray-400 font-medium uppercase tracking-wide mb-3">Obras sociales habilitadas</p>
          <div className="flex flex-wrap gap-2">
            {p.obrasSociales.map(os => (
              <div key={os} className="flex items-center gap-1">
                <CheckCircle size={13} className="text-emerald-400" />
                <span className="text-xs text-gray-600">{os}</span>
              </div>
            ))}
          </div>
        </Card>

        {/* Disponibilidad */}
        <Card className="p-4">
          <p className="text-xs text-gray-400 font-medium uppercase tracking-wide mb-3">Próxima disponibilidad</p>

          {/* Días */}
          <div className="flex gap-1 mb-3">
            {dias.map((d, i) => (
              <button
                key={i}
                onClick={() => { setDiaIdx(i); setHoraSel(null) }}
                className={`flex-1 py-2 rounded-xl flex flex-col items-center transition-all
                  ${diaIdx === i ? 'bg-sky-500 text-white' : 'bg-gray-50 text-gray-600 hover:bg-gray-100'}`}
              >
                <span className="text-[10px] font-medium">{DIAS_SEMANA[d.getDay()]}</span>
                <span className="text-sm font-bold">{d.getDate()}</span>
              </button>
            ))}
          </div>

          {/* Horarios */}
          <div className="flex flex-wrap gap-2">
            {p.horarios.map(h => (
              <button
                key={h}
                onClick={() => setHoraSel(h)}
                className={`px-3 py-1.5 rounded-xl text-xs font-medium border transition-all
                  ${horaSel === h
                    ? 'bg-sky-500 text-white border-sky-500'
                    : 'bg-white text-gray-600 border-gray-200 hover:border-sky-200'
                  }`}
              >
                {h}
              </button>
            ))}
          </div>
        </Card>

        {/* Reseñas */}
        <div>
          <p className="text-sm font-semibold text-gray-900 mb-3">Reseñas de pacientes</p>
          <div className="flex flex-col gap-2">
            {p.reseñas.map((r, i) => (
              <Card key={i} className="p-4">
                <div className="flex items-center justify-between mb-1.5">
                  <p className="text-sm font-semibold text-gray-800">{r.autor}</p>
                  <div className="flex items-center gap-0.5">
                    {Array.from({ length: r.nota }).map((_, j) => (
                      <Star key={j} size={12} fill="#f59e0b" className="text-amber-400" />
                    ))}
                  </div>
                </div>
                <p className="text-xs text-gray-600 leading-relaxed">{r.texto}</p>
                <p className="text-xs text-gray-400 mt-1">{r.fecha}</p>
              </Card>
            ))}
          </div>
        </div>
      </div>

      {/* CTA fijo abajo */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-lg px-4 py-4 bg-white border-t border-gray-100">
        <Button
          size="full"
          disabled={!horaSel}
          onClick={handleReservar}
        >
          {horaSel ? `Reservar turno · ${horaSel}` : 'Seleccioná un horario'}
        </Button>
      </div>
    </div>
  )
}
