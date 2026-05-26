import { useState } from 'react'
import { ChevronLeft, ChevronRight, Plus, Video, MapPin } from 'lucide-react'
import MobileLayout from '../../components/layout/MobileLayout'
import TopBar from '../../components/layout/TopBar'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

const DIAS = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb']
const MESES = ['Enero','Febrero','Marzo','Abril','Mayo','Junio','Julio','Agosto','Septiembre','Octubre','Noviembre','Diciembre']

function buildWeek(base) {
  const start = new Date(base)
  start.setDate(base.getDate() - base.getDay() + 1) // lunes
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    return d
  })
}

const TURNOS_SEMANA = {
  1: [{ hora: '09:00', nombre: 'Ana Rodríguez', modalidad: 'telemedicina' }, { hora: '10:30', nombre: 'Carlos Méndez', modalidad: 'telemedicina' }],
  2: [{ hora: '11:00', nombre: 'Lucía Pérez', modalidad: 'presencial' }],
  4: [{ hora: '09:00', nombre: 'Sofía Torres', modalidad: 'presencial' }, { hora: '14:00', nombre: 'Diego Fernández', modalidad: 'telemedicina' }],
  5: [{ hora: '10:00', nombre: 'Valentina Ruiz', modalidad: 'telemedicina' }],
}

export default function Agenda() {
  const today = new Date()
  const [selectedDate, setSelectedDate] = useState(today)
  const [weekBase, setWeekBase] = useState(today)
  const week = buildWeek(weekBase)

  const dayIndex = selectedDate.getDay() === 0 ? 6 : selectedDate.getDay() - 1
  const turnos = TURNOS_SEMANA[dayIndex] || []

  function prevWeek() {
    const d = new Date(weekBase)
    d.setDate(weekBase.getDate() - 7)
    setWeekBase(d)
  }
  function nextWeek() {
    const d = new Date(weekBase)
    d.setDate(weekBase.getDate() + 7)
    setWeekBase(d)
  }

  return (
    <MobileLayout>
      <TopBar
        title="Agenda"
        actions={
          <button className="w-8 h-8 bg-sky-500 rounded-xl flex items-center justify-center text-white">
            <Plus size={18} />
          </button>
        }
      />

      <div className="px-4 py-4 flex flex-col gap-4">

        {/* Navegador de semana */}
        <Card className="p-3">
          <div className="flex items-center justify-between mb-3">
            <button onClick={prevWeek} className="p-1 rounded-lg hover:bg-gray-50 text-gray-400">
              <ChevronLeft size={18} />
            </button>
            <p className="text-sm font-semibold text-gray-700">
              {MESES[weekBase.getMonth()]} {weekBase.getFullYear()}
            </p>
            <button onClick={nextWeek} className="p-1 rounded-lg hover:bg-gray-50 text-gray-400">
              <ChevronRight size={18} />
            </button>
          </div>

          <div className="grid grid-cols-7 gap-1">
            {week.map((day, i) => {
              const isToday = day.toDateString() === today.toDateString()
              const isSelected = day.toDateString() === selectedDate.toDateString()
              const hasTurnos = TURNOS_SEMANA[i] && TURNOS_SEMANA[i].length > 0
              return (
                <button
                  key={i}
                  onClick={() => setSelectedDate(day)}
                  className={`flex flex-col items-center py-2 rounded-xl transition-all
                    ${isSelected ? 'bg-sky-500 text-white' : isToday ? 'bg-sky-50 text-sky-600' : 'text-gray-600 hover:bg-gray-50'}`}
                >
                  <span className="text-[10px] font-medium">{DIAS[day.getDay()]}</span>
                  <span className="text-sm font-bold">{day.getDate()}</span>
                  {hasTurnos && (
                    <div className={`w-1 h-1 rounded-full mt-0.5 ${isSelected ? 'bg-white' : 'bg-sky-400'}`} />
                  )}
                </button>
              )
            })}
          </div>
        </Card>

        {/* Turnos del día seleccionado */}
        <div>
          <p className="text-sm font-semibold text-gray-900 mb-3">
            {selectedDate.toLocaleDateString('es-AR', { weekday: 'long', day: 'numeric', month: 'long' })}
            {' · '}
            <span className="text-gray-400 font-normal">{turnos.length} turnos</span>
          </p>

          {turnos.length === 0 ? (
            <div className="text-center py-12 text-gray-300">
              <p className="text-sm">Sin turnos para este día</p>
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {turnos.map((t, i) => (
                <Card key={i} className="px-4 py-3 flex items-center gap-3">
                  <p className="text-sm font-bold text-gray-900 min-w-[44px]">{t.hora}</p>
                  <div className="w-px h-8 bg-gray-100" />
                  <div className="flex-1">
                    <p className="text-sm font-semibold text-gray-800">{t.nombre}</p>
                    <div className="flex items-center gap-1 mt-0.5">
                      {t.modalidad === 'telemedicina'
                        ? <Video size={12} className="text-sky-400" />
                        : <MapPin size={12} className="text-gray-400" />
                      }
                      <span className="text-xs text-gray-400 capitalize">{t.modalidad}</span>
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Configurar disponibilidad */}
        <Button variant="outline" size="full" className="mt-2">
          Configurar disponibilidad
        </Button>

        <div className="h-2" />
      </div>
    </MobileLayout>
  )
}
