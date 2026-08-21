import { useNavigate, useSearchParams } from 'react-router-dom'
import { ArrowLeft, Star, Video, MapPin, Home, Clock } from 'lucide-react'
import { PROFESIONALES } from '../../data/profesionales'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'

export default function Busqueda() {
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const especialidad = params.get('especialidad')
  const obraSocial   = params.get('obraSocial')
  const zona         = params.get('zona')
  const modalidad    = params.get('modalidad')

  const resultados = PROFESIONALES.filter(p => {
    if (especialidad && p.especialidad !== especialidad) return false
    if (obraSocial && obraSocial !== 'Sin obra social' && !p.obrasSociales.includes(obraSocial)) return false
    if (zona && p.zona !== zona) return false
    if (modalidad && !p.modalidades.includes(modalidad)) return false
    return true
  })

  return (
    <div className="min-h-screen bg-gray-50 max-w-lg mx-auto">
      {/* Header */}
      <header className="sticky top-0 bg-white border-b border-gray-100 z-30">
        <div className="px-4 h-14 flex items-center gap-3">
          <button onClick={() => navigate('/buscar')} className="p-1.5 -ml-1.5 rounded-xl text-gray-500 hover:bg-gray-100">
            <ArrowLeft size={20} />
          </button>
          <div className="flex-1">
            <p className="text-sm font-semibold text-gray-900">{especialidad || 'Todos los profesionales'}</p>
            <p className="text-xs text-gray-400">
              {[obraSocial, zona, modalidad].filter(Boolean).join(' · ') || 'Sin filtros aplicados'}
            </p>
          </div>
          <span className="text-xs text-gray-400 font-medium">{resultados.length} resultados</span>
        </div>
      </header>

      <div className="px-4 py-4 flex flex-col gap-3">
        {resultados.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-gray-400 text-sm">No encontramos profesionales con esos filtros.</p>
            <button onClick={() => navigate('/buscar')} className="mt-3 text-sky-500 text-sm font-medium">
              Modificar búsqueda
            </button>
          </div>
        ) : (
          resultados.map(p => (
            <ProfesionalCard
              key={p.id}
              profesional={p}
              obraSocialPaciente={obraSocial}
              onClick={() => navigate(`/buscar/${p.slug}?obraSocial=${obraSocial || ''}`)}
            />
          ))
        )}
      </div>
    </div>
  )
}

function ProfesionalCard({ profesional: p, obraSocialPaciente, onClick }) {
  const coseguro = obraSocialPaciente ? p.coseguros?.[obraSocialPaciente] : undefined
  const tieneOS  = coseguro !== undefined

  return (
    <Card onClick={onClick} className="p-4">
      <div className="flex items-start gap-3">
        {/* Avatar */}
        <div className="w-14 h-14 rounded-2xl bg-sky-100 flex items-center justify-center flex-shrink-0">
          <span className="text-lg font-bold text-sky-500">
            {p.nombre.split(' ').map(n => n[0]).join('').slice(0, 2)}
          </span>
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <p className="font-semibold text-gray-900 text-sm">{p.nombre}</p>
            <div className="flex items-center gap-1 flex-shrink-0">
              <Star size={12} fill="#f59e0b" className="text-amber-400" />
              <span className="text-xs font-semibold text-gray-700">{p.rating}</span>
              <span className="text-xs text-gray-400">({p.totalReseñas})</span>
            </div>
          </div>

          <p className="text-xs text-gray-500 mt-0.5">{p.especialidad} · {p.zona}</p>

          {/* Modalidades */}
          <div className="flex gap-1.5 mt-2 flex-wrap">
            {p.modalidades.map(m => (
              <div key={m} className="flex items-center gap-1">
                {m === 'telemedicina' && <Video size={11} className="text-sky-400" />}
                {m === 'presencial'   && <MapPin size={11} className="text-gray-400" />}
                {m === 'domicilio'    && <Home size={11} className="text-emerald-400" />}
                <span className="text-xs text-gray-400 capitalize">{m === 'telemedicina' ? 'Online' : m}</span>
              </div>
            ))}
          </div>

          {/* Precio */}
          <div className="flex items-end justify-between mt-3">
            <div>
              <div className="flex items-baseline gap-1.5">
                {tieneOS ? (
                  <>
                    <span className="text-base font-bold text-emerald-600">
                      ${coseguro === 0 ? 'Gratis' : coseguro.toLocaleString('es-AR')}
                    </span>
                    <span className="text-xs text-gray-400">con {obraSocialPaciente}</span>
                  </>
                ) : (
                  <span className="text-base font-bold text-gray-900">
                    ${p.precio.toLocaleString('es-AR')}
                  </span>
                )}
              </div>
              {tieneOS && (
                <p className="text-xs text-gray-400">Valor privado ${p.precio.toLocaleString('es-AR')}</p>
              )}
            </div>
            <div className="flex items-center gap-1 text-xs text-sky-500 font-medium">
              <Clock size={12} />
              {p.proximoDisponible}
            </div>
          </div>
        </div>
      </div>
    </Card>
  )
}
