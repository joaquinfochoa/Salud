import { useState } from 'react'
import { Camera, ChevronRight } from 'lucide-react'
import { useAuth } from '../../context/AuthContext'
import MobileLayout from '../../components/layout/MobileLayout'
import TopBar from '../../components/layout/TopBar'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import Input from '../../components/ui/Input'

const OBRAS_SOCIALES = ['OSDE', 'Swiss Medical', 'Galeno', 'OMINT', 'Medifé', 'IOMA', 'PAMI', 'Sanitas']

export default function Perfil() {
  const { user, logout } = useAuth()
  const [precio, setPrecio] = useState('8500')
  const [desc, setDesc] = useState('')
  const [obrasSel, setObrasSel] = useState(['OSDE', 'Swiss Medical', 'Galeno'])
  const [saving, setSaving] = useState(false)

  function toggleOS(os) {
    setObrasSel(prev => prev.includes(os) ? prev.filter(o => o !== os) : [...prev, os])
  }

  async function handleSave() {
    setSaving(true)
    await new Promise(r => setTimeout(r, 800))
    setSaving(false)
  }

  return (
    <MobileLayout>
      <TopBar title="Mi perfil" />

      <div className="px-4 py-4 flex flex-col gap-4">

        {/* Avatar */}
        <div className="flex flex-col items-center py-4">
          <div className="relative">
            <div className="w-20 h-20 rounded-full bg-sky-100 flex items-center justify-center">
              <span className="text-2xl font-bold text-sky-500">
                {user?.nombre?.[0]}{user?.apellido?.[0]}
              </span>
            </div>
            <button className="absolute -bottom-1 -right-1 w-7 h-7 bg-sky-500 rounded-full flex items-center justify-center">
              <Camera size={14} className="text-white" />
            </button>
          </div>
          <p className="font-semibold text-gray-900 mt-3">{user?.nombre} {user?.apellido}</p>
          <p className="text-sm text-gray-400">{user?.especialidad}</p>
          <Badge variant="success" className="mt-1">{user?.matricula}</Badge>
        </div>

        {/* Sobre mí */}
        <Card className="p-4">
          <p className="text-sm font-semibold text-gray-900 mb-3">Sobre mí</p>
          <textarea
            value={desc}
            onChange={e => setDesc(e.target.value)}
            placeholder="Describí tu enfoque, formación y tipo de pacientes que atendés..."
            className="w-full text-sm text-gray-700 placeholder-gray-300 resize-none outline-none min-h-[80px] leading-relaxed"
            rows={3}
          />
        </Card>

        {/* Precio */}
        <Card className="p-4">
          <p className="text-sm font-semibold text-gray-900 mb-3">Precio por consulta</p>
          <Input
            label="Valor (ARS)"
            type="number"
            value={precio}
            onChange={e => setPrecio(e.target.value)}
            hint="Precio de consulta privada"
          />
        </Card>

        {/* Obras sociales */}
        <Card className="p-4">
          <p className="text-sm font-semibold text-gray-900 mb-1">Obras sociales habilitadas</p>
          <p className="text-xs text-gray-400 mb-3">Seleccioná con cuáles querés trabajar a través de la Red</p>
          <div className="flex flex-wrap gap-2">
            {OBRAS_SOCIALES.map(os => {
              const active = obrasSel.includes(os)
              return (
                <button
                  key={os}
                  onClick={() => toggleOS(os)}
                  className={`px-3 py-1.5 rounded-full text-xs font-medium border transition-all
                    ${active ? 'bg-sky-500 text-white border-sky-500' : 'bg-white text-gray-500 border-gray-200 hover:border-sky-200'}`}
                >
                  {os}
                </button>
              )
            })}
          </div>
        </Card>

        {/* Ver perfil público */}
        <button className="flex items-center justify-between px-4 py-3.5 bg-white border border-gray-100 rounded-2xl shadow-sm text-sm font-medium text-gray-700 hover:shadow-md transition-all">
          Ver mi perfil público
          <ChevronRight size={16} className="text-gray-400" />
        </button>

        <Button size="full" onClick={handleSave} loading={saving}>
          Guardar cambios
        </Button>

        {/* Cerrar sesión */}
        <button onClick={logout} className="text-sm text-red-400 text-center py-2 hover:text-red-500">
          Cerrar sesión
        </button>

        <div className="h-2" />
      </div>
    </MobileLayout>
  )
}
