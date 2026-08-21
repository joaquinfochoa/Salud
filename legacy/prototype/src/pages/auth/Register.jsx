import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Stethoscope, User } from 'lucide-react'
import Button from '../../components/ui/Button'

export default function Register() {
  const [role, setRole] = useState(null)
  const navigate = useNavigate()

  function handleContinue() {
    if (role === 'profesional') navigate('/profesional/onboarding')
    else navigate('/buscar')
  }

  return (
    <div className="min-h-screen bg-white flex flex-col max-w-lg mx-auto px-6">
      <div className="flex-1 flex flex-col justify-center py-12">
        <div className="mb-10">
          <h1 className="text-2xl font-bold text-gray-900">Crear cuenta</h1>
          <p className="text-gray-400 text-sm mt-1">¿Con qué rol vas a usar la plataforma?</p>
        </div>

        <div className="flex flex-col gap-3">
          <RoleCard
            icon={<Stethoscope size={28} strokeWidth={1.5} />}
            title="Soy profesional de salud"
            description="Quiero gestionar mi agenda, cobrar y acceder a obras sociales"
            selected={role === 'profesional'}
            onClick={() => setRole('profesional')}
          />
          <RoleCard
            icon={<User size={28} strokeWidth={1.5} />}
            title="Soy paciente"
            description="Quiero encontrar profesionales y reservar turnos"
            selected={role === 'paciente'}
            onClick={() => setRole('paciente')}
          />
        </div>

        <Button
          size="full"
          className="mt-6"
          disabled={!role}
          onClick={handleContinue}
        >
          Continuar
        </Button>

        <p className="mt-6 text-center text-sm text-gray-500">
          ¿Ya tenés cuenta?{' '}
          <Link to="/login" className="text-sky-500 font-medium">
            Iniciá sesión
          </Link>
        </p>
      </div>
    </div>
  )
}

function RoleCard({ icon, title, description, selected, onClick }) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-4 rounded-2xl border-2 transition-all flex items-start gap-4
        ${selected
          ? 'border-sky-500 bg-sky-50'
          : 'border-gray-100 bg-white hover:border-gray-200'
        }`}
    >
      <div className={`mt-0.5 ${selected ? 'text-sky-500' : 'text-gray-400'}`}>
        {icon}
      </div>
      <div>
        <p className={`font-semibold text-sm ${selected ? 'text-sky-700' : 'text-gray-800'}`}>{title}</p>
        <p className="text-xs text-gray-400 mt-0.5 leading-relaxed">{description}</p>
      </div>
    </button>
  )
}
