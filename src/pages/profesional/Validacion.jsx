import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { CheckCircle, Loader } from 'lucide-react'
import { useAuth } from '../../context/AuthContext'
import Button from '../../components/ui/Button'

const CHECKS = [
  { id: 'refeps',   label: 'Validando matrícula en REFEPS',     delay: 1000 },
  { id: 'renaper',  label: 'Verificando identidad en RENAPER',  delay: 2200 },
  { id: 'sss',      label: 'Consultando estado en SSS',         delay: 3400 },
  { id: 'convenios',label: 'Habilitando convenios de la Red',   delay: 4400 },
]

export default function Validacion() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [done, setDone] = useState([])
  const [finished, setFinished] = useState(false)

  useEffect(() => {
    CHECKS.forEach(({ id, delay }) => {
      setTimeout(() => {
        setDone(d => [...d, id])
        if (id === CHECKS[CHECKS.length - 1].id) {
          setTimeout(() => setFinished(true), 600)
        }
      }, delay)
    })
  }, [])

  return (
    <div className="min-h-screen bg-white flex flex-col max-w-lg mx-auto px-6">
      <div className="flex-1 flex flex-col justify-center py-12">
        {/* Logo */}
        <div className="w-14 h-14 bg-sky-500 rounded-2xl flex items-center justify-center mb-8">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
          </svg>
        </div>

        <h1 className="text-2xl font-bold text-gray-900 mb-1">
          Validando tu perfil
        </h1>
        <p className="text-gray-400 text-sm mb-8">
          Verificamos tu matrícula y habilitaciones automáticamente
        </p>

        <div className="flex flex-col gap-4">
          {CHECKS.map(({ id, label }) => {
            const isDone = done.includes(id)
            const isActive = !isDone && done.length === CHECKS.findIndex(c => c.id === id)
            return (
              <div key={id} className="flex items-center gap-3">
                <div className="w-6 h-6 flex items-center justify-center flex-shrink-0">
                  {isDone
                    ? <CheckCircle size={22} className="text-emerald-500" />
                    : isActive
                      ? <Loader size={22} className="text-sky-500 animate-spin" />
                      : <div className="w-5 h-5 rounded-full border-2 border-gray-200" />
                  }
                </div>
                <p className={`text-sm transition-colors ${isDone ? 'text-gray-700 font-medium' : 'text-gray-400'}`}>
                  {label}
                </p>
                {isDone && <span className="ml-auto text-xs text-emerald-500 font-medium">OK</span>}
              </div>
            )
          })}
        </div>

        {finished && (
          <div className="mt-10 transition-opacity duration-500">
            <div className="bg-emerald-50 border border-emerald-100 rounded-2xl p-4 mb-6">
              <p className="text-emerald-700 font-semibold text-sm">Matrícula verificada</p>
              <p className="text-emerald-600 text-xs mt-0.5">
                {user?.matricula} · Vigente · Habilitada para telemedicina
              </p>
            </div>
            <Button size="full" onClick={() => navigate('/profesional')}>
              Ir al dashboard
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
