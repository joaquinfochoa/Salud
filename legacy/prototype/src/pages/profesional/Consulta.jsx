import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Video, Phone, FileText, Pill } from 'lucide-react'
import MobileLayout from '../../components/layout/MobileLayout'
import TopBar from '../../components/layout/TopBar'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

const MOCK_CONSULTA = {
  id: '1',
  hora: '09:00',
  paciente: { nombre: 'Ana Rodríguez', cuil: '27-38.291.445-3', obraSocial: 'OSDE 210', edad: 34 },
  motivo: 'Ansiedad generalizada — seguimiento mensual',
  antecedentes: 'Tratamiento desde Marzo 2024. Refiere mejora en episodios nocturnos.',
  modalidad: 'telemedicina',
  honorarios: { privado: 12000, obraSocial: 9200, coseguro: 2800 },
}

export default function Consulta() {
  const { id } = useParams()
  const navigate = useNavigate()
  const c = MOCK_CONSULTA
  const [nota, setNota] = useState('')
  const [saving, setSaving] = useState(false)

  async function handleCerrar() {
    setSaving(true)
    await new Promise(r => setTimeout(r, 800))
    navigate(`/profesional/cobro/${id}`)
  }

  return (
    <MobileLayout>
      <TopBar title="Consulta" subtitle={`${c.hora} · ${c.paciente.nombre}`} backTo="/profesional" />

      <div className="px-4 py-4 flex flex-col gap-4">

        {/* Paciente */}
        <Card className="p-4">
          <div className="flex items-start justify-between mb-3">
            <div>
              <p className="font-semibold text-gray-900">{c.paciente.nombre}</p>
              <p className="text-xs text-gray-400 mt-0.5">CUIL {c.paciente.cuil} · {c.paciente.edad} años</p>
            </div>
            <Badge variant="primary">{c.paciente.obraSocial}</Badge>
          </div>
          <p className="text-sm text-gray-600 leading-relaxed">{c.motivo}</p>
          {c.antecedentes && (
            <p className="text-xs text-gray-400 mt-2 leading-relaxed border-t border-gray-50 pt-2">
              {c.antecedentes}
            </p>
          )}
        </Card>

        {/* Acción principal */}
        {c.modalidad === 'telemedicina' && (
          <button className="flex items-center justify-center gap-2 w-full py-3.5 bg-sky-500 hover:bg-sky-600 text-white font-semibold rounded-2xl transition-colors active:scale-95">
            <Video size={20} />
            Iniciar videollamada
          </button>
        )}

        {/* Nota clínica */}
        <Card className="p-4">
          <div className="flex items-center gap-2 mb-3">
            <FileText size={16} className="text-gray-400" />
            <p className="text-sm font-semibold text-gray-900">Nota clínica</p>
          </div>
          <textarea
            value={nota}
            onChange={e => setNota(e.target.value)}
            placeholder="Registrá los hallazgos, evolución y plan terapéutico..."
            className="w-full text-sm text-gray-700 placeholder-gray-300 resize-none outline-none min-h-[100px] leading-relaxed"
            rows={4}
          />
        </Card>

        {/* Prescripción (placeholder) */}
        <button className="flex items-center gap-3 px-4 py-3.5 border-2 border-dashed border-gray-200 rounded-2xl text-gray-400 hover:border-sky-200 hover:text-sky-400 transition-colors w-full">
          <Pill size={18} />
          <span className="text-sm font-medium">Agregar prescripción con CUIR</span>
        </button>

        {/* Cobro estimado */}
        <Card className="p-4">
          <p className="text-xs text-gray-400 font-medium mb-2 uppercase tracking-wide">Liquidación estimada</p>
          <div className="flex justify-between text-sm mb-1">
            <span className="text-gray-500">Honorario total</span>
            <span className="font-medium">${c.honorarios.privado.toLocaleString('es-AR')}</span>
          </div>
          <div className="flex justify-between text-sm mb-1">
            <span className="text-gray-500">Cobertura {c.paciente.obraSocial}</span>
            <span className="text-sky-600">−${c.honorarios.obraSocial.toLocaleString('es-AR')}</span>
          </div>
          <div className="flex justify-between text-sm border-t border-gray-100 pt-2 mt-2">
            <span className="font-semibold text-gray-700">Coseguro paciente</span>
            <span className="font-bold text-emerald-600">${c.honorarios.coseguro.toLocaleString('es-AR')}</span>
          </div>
        </Card>

        <Button size="full" onClick={handleCerrar} loading={saving}>
          Cerrar consulta y cobrar
        </Button>

        <div className="h-2" />
      </div>
    </MobileLayout>
  )
}
