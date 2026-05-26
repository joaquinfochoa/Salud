import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search, X } from 'lucide-react'
import { SINTOMAS, ESPECIALIDADES_LISTA, OBRAS_SOCIALES_LISTA } from '../../data/profesionales'
import Button from '../../components/ui/Button'
import Select from '../../components/ui/Select'

const MODALIDADES = [
  { id: 'telemedicina', label: 'Online' },
  { id: 'presencial',   label: 'Presencial' },
  { id: 'domicilio',    label: 'A domicilio' },
]

const ZONAS = ['CABA', 'GBA Norte', 'GBA Sur', 'GBA Oeste']

export default function Onboarding() {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [especialidad, setEspecialidad] = useState('')
  const [obraSocial, setObraSocial] = useState('')
  const [zona, setZona] = useState('')
  const [modalidad, setModalidad] = useState('')
  const [showSugerencias, setShowSugerencias] = useState(false)

  const sugerencias = query.length > 1
    ? SINTOMAS.filter(s => s.sintoma.toLowerCase().includes(query.toLowerCase()))
    : []

  function seleccionarSintoma(item) {
    setQuery(item.sintoma)
    setEspecialidad(item.especialidad)
    setShowSugerencias(false)
  }

  function buscar() {
    const params = new URLSearchParams()
    if (especialidad) params.set('especialidad', especialidad)
    if (obraSocial)   params.set('obraSocial', obraSocial)
    if (zona)         params.set('zona', zona)
    if (modalidad)    params.set('modalidad', modalidad)
    navigate(`/buscar/resultados?${params.toString()}`)
  }

  return (
    <div className="min-h-screen bg-white flex flex-col max-w-lg mx-auto">
      {/* Hero */}
      <div className="bg-sky-500 px-6 pt-12 pb-8">
        <h1 className="text-2xl font-bold text-white leading-snug mb-1">
          Encontrá el profesional<br />que necesitás
        </h1>
        <p className="text-sky-100 text-sm">
          Con tu obra social. Con precio transparente. Turno en minutos.
        </p>
      </div>

      <div className="px-6 -mt-4 flex flex-col gap-4 pb-8">
        {/* Buscador de síntoma */}
        <div className="relative">
          <div className="flex items-center gap-3 bg-white rounded-2xl shadow-md px-4 py-3 border border-gray-100">
            <Search size={18} className="text-gray-400 flex-shrink-0" />
            <input
              value={query}
              onChange={e => { setQuery(e.target.value); setShowSugerencias(true) }}
              onFocus={() => setShowSugerencias(true)}
              placeholder="¿Qué síntoma tenés o qué especialidad buscás?"
              className="flex-1 text-sm outline-none text-gray-700 placeholder-gray-400"
            />
            {query && (
              <button onClick={() => { setQuery(''); setEspecialidad('') }}>
                <X size={16} className="text-gray-400" />
              </button>
            )}
          </div>

          {showSugerencias && sugerencias.length > 0 && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden z-20">
              {sugerencias.map((s, i) => (
                <button
                  key={i}
                  onClick={() => seleccionarSintoma(s)}
                  className="w-full text-left px-4 py-3 hover:bg-gray-50 transition-colors flex items-start justify-between gap-3"
                >
                  <span className="text-sm text-gray-700">{s.sintoma}</span>
                  <span className="text-xs text-sky-500 font-medium whitespace-nowrap">{s.especialidad}</span>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Chips de especialidades */}
        <div>
          <p className="text-xs font-medium text-gray-500 mb-2">O elegí una especialidad</p>
          <div className="flex gap-2 flex-wrap">
            {ESPECIALIDADES_LISTA.map(e => (
              <button
                key={e}
                onClick={() => { setEspecialidad(esp => esp === e ? '' : e); setQuery('') }}
                className={`px-3 py-1.5 rounded-full text-xs font-medium border transition-all
                  ${especialidad === e
                    ? 'bg-sky-500 text-white border-sky-500'
                    : 'bg-white text-gray-600 border-gray-200 hover:border-sky-200'
                  }`}
              >
                {e}
              </button>
            ))}
          </div>
        </div>

        {/* Filtros */}
        <div className="flex flex-col gap-3 mt-1">
          <Select
            label="Tu obra social"
            value={obraSocial}
            onChange={e => setObraSocial(e.target.value)}
          >
            <option value="">Todas las obras sociales</option>
            {OBRAS_SOCIALES_LISTA.map(os => <option key={os} value={os}>{os}</option>)}
          </Select>

          <div className="grid grid-cols-2 gap-3">
            <Select label="Zona" value={zona} onChange={e => setZona(e.target.value)}>
              <option value="">Cualquier zona</option>
              {ZONAS.map(z => <option key={z} value={z}>{z}</option>)}
            </Select>

            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium text-gray-700">Modalidad</p>
              <div className="flex gap-1 flex-wrap">
                {MODALIDADES.map(m => (
                  <button
                    key={m.id}
                    onClick={() => setModalidad(mod => mod === m.id ? '' : m.id)}
                    className={`px-2.5 py-1.5 rounded-xl text-xs font-medium border transition-all
                      ${modalidad === m.id
                        ? 'bg-sky-500 text-white border-sky-500'
                        : 'bg-white text-gray-600 border-gray-200'
                      }`}
                  >
                    {m.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        <Button size="full" onClick={buscar} className="mt-2">
          <Search size={16} />
          Buscar profesionales
        </Button>
      </div>
    </div>
  )
}
