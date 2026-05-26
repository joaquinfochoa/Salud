import { useNavigate } from 'react-router-dom'
import { Bell, ChevronRight, Clock, Video, MapPin, DollarSign, Users, AlertCircle } from 'lucide-react'
import { useAuth } from '../../context/AuthContext'
import MobileLayout from '../../components/layout/MobileLayout'
import TopBar from '../../components/layout/TopBar'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'

const TURNOS_HOY = [
  { id: '1', hora: '09:00', nombre: 'Ana Rodríguez', motivo: 'Ansiedad generalizada', modalidad: 'telemedicina', obraSocial: 'OSDE', coseguro: 2800 },
  { id: '2', hora: '10:30', nombre: 'Carlos Méndez',  motivo: 'Depresión, seguimiento', modalidad: 'telemedicina', obraSocial: 'Swiss Medical', coseguro: 3200 },
  { id: '3', hora: '12:00', nombre: 'Sofía Torres',   motivo: 'Primera consulta',      modalidad: 'presencial',   obraSocial: null, coseguro: 8500 },
  { id: '4', hora: '15:30', nombre: 'Diego Fernández', motivo: 'Estrés laboral',       modalidad: 'telemedicina', obraSocial: 'Galeno', coseguro: 2400 },
]

const COBRADO_HOY = TURNOS_HOY.reduce((acc, t) => acc + t.coseguro, 0)
const TURNOS_COMPLETADOS = 2

export default function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const now = new Date()
  const hora = now.getHours()
  const saludo = hora < 12 ? 'Buenos días' : hora < 19 ? 'Buenas tardes' : 'Buenas noches'

  return (
    <MobileLayout>
      <TopBar
        title={`${saludo}, ${user?.nombre}`}
        subtitle={user?.especialidad}
        actions={
          <button className="relative p-1.5 rounded-xl text-gray-500 hover:bg-gray-100">
            <Bell size={20} />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
          </button>
        }
      />

      <div className="px-4 py-4 flex flex-col gap-4">

        {/* Alerta */}
        <div className="flex items-start gap-3 bg-amber-50 border border-amber-100 rounded-2xl px-4 py-3">
          <AlertCircle size={18} className="text-amber-500 flex-shrink-0 mt-0.5" />
          <p className="text-sm text-amber-700">
            Completá tu perfil para aparecer en los resultados de búsqueda.{' '}
            <button onClick={() => navigate('/profesional/perfil')} className="font-semibold underline underline-offset-2">
              Completar ahora
            </button>
          </p>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-3">
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-8 h-8 bg-sky-50 rounded-xl flex items-center justify-center">
                <Users size={16} className="text-sky-500" />
              </div>
              <p className="text-xs text-gray-400 font-medium">Turnos hoy</p>
            </div>
            <p className="text-2xl font-bold text-gray-900">{TURNOS_HOY.length}</p>
            <p className="text-xs text-gray-400 mt-0.5">{TURNOS_COMPLETADOS} completados</p>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-8 h-8 bg-emerald-50 rounded-xl flex items-center justify-center">
                <DollarSign size={16} className="text-emerald-500" />
              </div>
              <p className="text-xs text-gray-400 font-medium">Cobrado hoy</p>
            </div>
            <p className="text-2xl font-bold text-gray-900">
              ${(COBRADO_HOY / 1000).toFixed(1)}k
            </p>
            <p className="text-xs text-emerald-500 font-medium mt-0.5">En tu cuenta</p>
          </Card>
        </div>

        {/* Turnos del día */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold text-gray-900">Turnos de hoy</h2>
            <button
              onClick={() => navigate('/profesional/agenda')}
              className="text-xs text-sky-500 font-medium flex items-center gap-0.5"
            >
              Ver agenda <ChevronRight size={14} />
            </button>
          </div>

          <div className="flex flex-col gap-2">
            {TURNOS_HOY.map((turno, i) => (
              <TurnoCard key={turno.id} turno={turno} completed={i < TURNOS_COMPLETADOS} onClick={() => navigate(`/profesional/consulta/${turno.id}`)} />
            ))}
          </div>
        </div>

      </div>
    </MobileLayout>
  )
}

function TurnoCard({ turno, completed, onClick }) {
  return (
    <Card
      onClick={!completed ? onClick : undefined}
      className={`px-4 py-3 flex items-center gap-3 ${completed ? 'opacity-50' : ''}`}
    >
      {/* Hora */}
      <div className="flex flex-col items-center min-w-[44px]">
        <p className="text-sm font-bold text-gray-900">{turno.hora}</p>
        <div className={`w-2 h-2 rounded-full mt-1 ${completed ? 'bg-gray-300' : 'bg-sky-500'}`} />
      </div>

      {/* Separador */}
      <div className="w-px h-10 bg-gray-100" />

      {/* Info */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-gray-800 truncate">{turno.nombre}</p>
        <p className="text-xs text-gray-400 truncate">{turno.motivo}</p>
        <div className="flex items-center gap-2 mt-1">
          {turno.modalidad === 'telemedicina'
            ? <Video size={12} className="text-sky-400" />
            : <MapPin size={12} className="text-gray-400" />
          }
          <span className="text-xs text-gray-400">
            {turno.modalidad === 'telemedicina' ? 'Online' : 'Presencial'}
          </span>
          {turno.obraSocial
            ? <Badge variant="primary">{turno.obraSocial}</Badge>
            : <Badge variant="default">Privado</Badge>
          }
        </div>
      </div>

      {/* Monto */}
      <div className="text-right">
        <p className="text-sm font-bold text-emerald-600">
          ${turno.coseguro.toLocaleString('es-AR')}
        </p>
        {completed
          ? <p className="text-xs text-gray-400">Cobrado</p>
          : <ChevronRight size={16} className="text-gray-300 ml-auto" />
        }
      </div>
    </Card>
  )
}
