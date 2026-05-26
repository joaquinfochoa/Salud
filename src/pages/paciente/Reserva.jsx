import { useState } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { ArrowLeft, Video, MapPin, CheckCircle, CreditCard, Lock } from 'lucide-react'
import { PROFESIONALES } from '../../data/profesionales'
import Card from '../../components/ui/Card'
import Button from '../../components/ui/Button'
import Input from '../../components/ui/Input'

export default function Reserva() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const hora        = searchParams.get('hora') || '09:00'
  const obraSocial  = searchParams.get('obraSocial') || ''

  const p = PROFESIONALES.find(x => x.id === id)
  const [pagando, setPagando] = useState(false)
  const [confirmado, setConfirmado] = useState(false)
  const [card, setCard] = useState({ numero: '', vencimiento: '', cvv: '', nombre: '' })

  const coseguro = p?.coseguros?.[obraSocial]
  const montoPagar = coseguro !== undefined ? coseguro : p?.precio || 0
  const hoy = new Date()
  hoy.setDate(hoy.getDate() + 1)
  const fechaStr = hoy.toLocaleDateString('es-AR', { weekday: 'long', day: 'numeric', month: 'long' })

  async function handlePagar() {
    setPagando(true)
    await new Promise(r => setTimeout(r, 1800))
    setConfirmado(true)
    setPagando(false)
  }

  if (!p) return null

  if (confirmado) return <Confirmacion profesional={p} hora={hora} fecha={fechaStr} monto={montoPagar} obraSocial={obraSocial} onVolver={() => navigate('/buscar')} />

  return (
    <div className="min-h-screen bg-gray-50 max-w-lg mx-auto">
      <header className="sticky top-0 bg-white border-b border-gray-100 z-30 px-4 h-14 flex items-center gap-3">
        <button onClick={() => navigate(-1)} className="p-1.5 -ml-1.5 rounded-xl text-gray-500 hover:bg-gray-100">
          <ArrowLeft size={20} />
        </button>
        <p className="text-sm font-semibold text-gray-900">Confirmar reserva</p>
      </header>

      <div className="px-4 py-4 flex flex-col gap-4">

        {/* Resumen del turno */}
        <Card className="p-4">
          <p className="text-xs text-gray-400 font-medium uppercase tracking-wide mb-3">Tu turno</p>
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-2xl bg-sky-100 flex items-center justify-center flex-shrink-0">
              <span className="text-base font-bold text-sky-500">
                {p.nombre.split(' ').map(n => n[0]).join('').slice(0, 2)}
              </span>
            </div>
            <div>
              <p className="font-semibold text-gray-900 text-sm">{p.nombre}</p>
              <p className="text-xs text-gray-500">{p.especialidad}</p>
            </div>
          </div>
          <div className="flex flex-col gap-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-500">Fecha</span>
              <span className="font-medium text-gray-800 capitalize">{fechaStr}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Hora</span>
              <span className="font-medium text-gray-800">{hora}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-500">Modalidad</span>
              <div className="flex items-center gap-1">
                <Video size={13} className="text-sky-400" />
                <span className="font-medium text-gray-800">Telemedicina</span>
              </div>
            </div>
          </div>
        </Card>

        {/* Desglose de pago */}
        <Card className="p-4">
          <p className="text-xs text-gray-400 font-medium uppercase tracking-wide mb-3">Detalle del pago</p>
          <div className="flex flex-col gap-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-500">Honorario consulta</span>
              <span className="text-gray-800">${p.precio.toLocaleString('es-AR')}</span>
            </div>
            {obraSocial && coseguro !== undefined && (
              <div className="flex justify-between">
                <span className="text-gray-500">Cobertura {obraSocial}</span>
                <span className="text-sky-500 font-medium">−${(p.precio - coseguro).toLocaleString('es-AR')}</span>
              </div>
            )}
            <div className="flex justify-between border-t border-gray-100 pt-2 mt-1">
              <span className="font-semibold text-gray-800">Total a pagar</span>
              <span className="font-bold text-lg text-gray-900">
                {montoPagar === 0 ? 'Sin cargo' : `$${montoPagar.toLocaleString('es-AR')}`}
              </span>
            </div>
          </div>
          {obraSocial && (
            <p className="text-xs text-gray-400 mt-2 leading-relaxed">
              El reintegro de {obraSocial} se gestiona automáticamente. No tenés que hacer ningún trámite.
            </p>
          )}
        </Card>

        {/* Pago con tarjeta */}
        {montoPagar > 0 && (
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-3">
              <CreditCard size={16} className="text-gray-400" />
              <p className="text-sm font-semibold text-gray-900">Pago con tarjeta</p>
            </div>
            <div className="flex flex-col gap-3">
              <Input
                label="Número de tarjeta"
                placeholder="1234 5678 9012 3456"
                value={card.numero}
                onChange={e => setCard(c => ({ ...c, numero: e.target.value }))}
                maxLength={19}
              />
              <Input
                label="Nombre en la tarjeta"
                placeholder="Como aparece en la tarjeta"
                value={card.nombre}
                onChange={e => setCard(c => ({ ...c, nombre: e.target.value }))}
              />
              <div className="grid grid-cols-2 gap-3">
                <Input label="Vencimiento" placeholder="MM/AA" value={card.vencimiento} onChange={e => setCard(c => ({ ...c, vencimiento: e.target.value }))} />
                <Input label="CVV" placeholder="123" value={card.cvv} onChange={e => setCard(c => ({ ...c, cvv: e.target.value }))} type="password" maxLength={4} />
              </div>
            </div>
          </Card>
        )}

        {/* Seguridad */}
        <div className="flex items-center justify-center gap-1.5 text-xs text-gray-400">
          <Lock size={12} />
          Pago procesado con cifrado SSL
        </div>

        <Button size="full" onClick={handlePagar} loading={pagando}>
          {montoPagar === 0 ? 'Confirmar turno' : `Pagar $${montoPagar.toLocaleString('es-AR')} y reservar`}
        </Button>

        <div className="h-2" />
      </div>
    </div>
  )
}

function Confirmacion({ profesional, hora, fecha, monto, obraSocial, onVolver }) {
  return (
    <div className="min-h-screen bg-white flex flex-col items-center justify-center max-w-lg mx-auto px-6 text-center">
      <div className="w-20 h-20 bg-emerald-100 rounded-full flex items-center justify-center mb-6">
        <CheckCircle size={40} className="text-emerald-500" />
      </div>

      <h1 className="text-2xl font-bold text-gray-900 mb-1">¡Turno reservado!</h1>
      <p className="text-gray-400 text-sm mb-8">Te enviamos la confirmación por email</p>

      <Card className="w-full p-4 text-left mb-4">
        <p className="text-xs text-gray-400 font-medium uppercase tracking-wide mb-3">Detalles del turno</p>
        <div className="flex flex-col gap-2 text-sm">
          <div className="flex justify-between"><span className="text-gray-500">Profesional</span><span className="font-medium">{profesional.nombre}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Fecha</span><span className="font-medium capitalize">{fecha}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Hora</span><span className="font-medium">{hora}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Modalidad</span><span className="font-medium">Videollamada</span></div>
          {obraSocial && <div className="flex justify-between"><span className="text-gray-500">Obra social</span><span className="font-medium text-sky-600">{obraSocial} ✓</span></div>}
          <div className="flex justify-between border-t border-gray-100 pt-2">
            <span className="font-semibold text-gray-700">Pagado</span>
            <span className="font-bold text-emerald-600">{monto === 0 ? 'Sin cargo' : `$${monto.toLocaleString('es-AR')}`}</span>
          </div>
        </div>
      </Card>

      <button onClick={onVolver} className="text-sky-500 font-medium text-sm">
        Volver al inicio
      </button>
    </div>
  )
}
