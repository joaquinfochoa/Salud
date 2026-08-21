import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { CheckCircle, Banknote } from 'lucide-react'
import MobileLayout from '../../components/layout/MobileLayout'
import TopBar from '../../components/layout/TopBar'
import Card from '../../components/ui/Card'
import Button from '../../components/ui/Button'

export default function Cobro() {
  const navigate = useNavigate()
  const [acreditado, setAcreditado] = useState(false)

  useEffect(() => {
    const t = setTimeout(() => setAcreditado(true), 1500)
    return () => clearTimeout(t)
  }, [])

  return (
    <MobileLayout>
      <TopBar title="Confirmación de cobro" backTo="/profesional" />

      <div className="px-4 py-8 flex flex-col items-center gap-6">

        <div className={`w-20 h-20 rounded-full flex items-center justify-center transition-all duration-700
          ${acreditado ? 'bg-emerald-100 scale-100' : 'bg-gray-100 scale-90'}`}>
          {acreditado
            ? <CheckCircle size={40} className="text-emerald-500" />
            : <Banknote size={40} className="text-gray-300 animate-pulse" />
          }
        </div>

        <div className="text-center">
          <p className={`text-lg font-bold transition-colors ${acreditado ? 'text-emerald-600' : 'text-gray-400'}`}>
            {acreditado ? 'Pago acreditado' : 'Procesando pago...'}
          </p>
          <p className="text-gray-400 text-sm mt-1">
            {acreditado ? 'El monto ya está disponible en tu cuenta' : 'Verificando con la obra social'}
          </p>
        </div>

        <Card className="w-full p-4">
          <p className="text-xs text-gray-400 font-medium mb-3 uppercase tracking-wide">Detalle del cobro</p>
          <div className="flex flex-col gap-2 text-sm">
            <Row label="Paciente" value="Ana Rodríguez" />
            <Row label="Obra social" value="OSDE 210" />
            <Row label="Cobertura OS" value="$9.200" />
            <Row label="Coseguro cobrado" value="$2.800" />
            <div className="border-t border-gray-100 pt-2 mt-1">
              <Row label="Fee plataforma (8%)" value="−$960" secondary />
              <Row label="Acreditado en cuenta" value="$11.040" bold green />
            </div>
          </div>
        </Card>

        {acreditado && (
          <div className="w-full flex flex-col gap-2 animate-in fade-in duration-500">
            <Button size="full" onClick={() => navigate('/profesional')}>
              Volver al inicio
            </Button>
          </div>
        )}
      </div>
    </MobileLayout>
  )
}

function Row({ label, value, secondary, bold, green }) {
  return (
    <div className="flex justify-between items-center">
      <span className={`${secondary ? 'text-xs text-gray-400' : 'text-gray-500'}`}>{label}</span>
      <span className={`
        ${bold ? 'font-bold text-base' : 'font-medium'}
        ${green ? 'text-emerald-600' : secondary ? 'text-gray-400' : 'text-gray-800'}
      `}>{value}</span>
    </div>
  )
}
