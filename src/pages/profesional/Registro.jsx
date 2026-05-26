import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import Button from '../../components/ui/Button'
import Input from '../../components/ui/Input'
import Select from '../../components/ui/Select'

const ESPECIALIDADES = ['Psicología', 'Kinesiología / Fisioterapia', 'Odontología']
const MODALIDADES = [
  { id: 'telemedicina', label: 'Telemedicina', desc: 'Atención por videollamada' },
  { id: 'presencial',   label: 'Presencial',   desc: 'En tu consultorio' },
  { id: 'domicilio',    label: 'A domicilio',   desc: 'En el domicilio del paciente' },
]
const ZONAS = ['CABA', 'GBA Norte', 'GBA Sur', 'GBA Oeste']

const TOTAL_STEPS = 3

export default function Registro() {
  const { setUserData } = useAuth()
  const navigate = useNavigate()
  const [step, setStep] = useState(1)
  const [form, setForm] = useState({
    nombre: '', apellido: '', email: '', password: '',
    especialidad: '', matricula: '', numeroMatricula: '',
    modalidades: [], zona: '',
  })
  const [errors, setErrors] = useState({})

  function set(field, value) {
    setForm(f => ({ ...f, [field]: value }))
    setErrors(e => ({ ...e, [field]: undefined }))
  }

  function toggleModalidad(id) {
    setForm(f => ({
      ...f,
      modalidades: f.modalidades.includes(id)
        ? f.modalidades.filter(m => m !== id)
        : [...f.modalidades, id]
    }))
  }

  function validateStep() {
    const errs = {}
    if (step === 1) {
      if (!form.nombre.trim())    errs.nombre = 'Requerido'
      if (!form.apellido.trim())  errs.apellido = 'Requerido'
      if (!form.email.includes('@')) errs.email = 'Email inválido'
      if (form.password.length < 6)  errs.password = 'Mínimo 6 caracteres'
    }
    if (step === 2) {
      if (!form.especialidad)         errs.especialidad = 'Seleccioná una especialidad'
      if (!form.numeroMatricula.trim()) errs.numeroMatricula = 'Ingresá tu número de matrícula'
    }
    if (step === 3) {
      if (form.modalidades.length === 0) errs.modalidades = 'Seleccioná al menos una modalidad'
      if (!form.zona)                    errs.zona = 'Seleccioná tu zona'
    }
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  function next() {
    if (!validateStep()) return
    if (step < TOTAL_STEPS) setStep(s => s + 1)
    else handleSubmit()
  }

  function handleSubmit() {
    setUserData({
      id: 'new-1',
      role: 'profesional',
      nombre: form.nombre,
      apellido: form.apellido,
      especialidad: form.especialidad,
      matricula: `MN ${form.numeroMatricula}`,
    })
    navigate('/profesional/validacion')
  }

  return (
    <div className="min-h-screen bg-white flex flex-col max-w-lg mx-auto">
      {/* Header */}
      <div className="px-6 pt-10 pb-6">
        <div className="flex items-center gap-2 mb-6">
          {Array.from({ length: TOTAL_STEPS }).map((_, i) => (
            <div
              key={i}
              className={`h-1 flex-1 rounded-full transition-colors ${
                i < step ? 'bg-sky-500' : 'bg-gray-100'
              }`}
            />
          ))}
        </div>
        <p className="text-xs font-medium text-sky-500 uppercase tracking-wide mb-1">
          Paso {step} de {TOTAL_STEPS}
        </p>
        <h1 className="text-xl font-bold text-gray-900">
          {step === 1 && 'Tus datos personales'}
          {step === 2 && 'Datos profesionales'}
          {step === 3 && 'Modalidad y zona'}
        </h1>
        <p className="text-sm text-gray-400 mt-1">
          {step === 1 && 'Creá tu cuenta en la plataforma'}
          {step === 2 && 'Ingresá tu especialidad y matrícula'}
          {step === 3 && 'Definí cómo y dónde atendés'}
        </p>
      </div>

      {/* Form */}
      <div className="flex-1 px-6 flex flex-col gap-4">
        {step === 1 && (
          <>
            <div className="flex gap-3">
              <Input label="Nombre" value={form.nombre} onChange={e => set('nombre', e.target.value)} error={errors.nombre} placeholder="Martín" className="flex-1" />
              <Input label="Apellido" value={form.apellido} onChange={e => set('apellido', e.target.value)} error={errors.apellido} placeholder="García" className="flex-1" />
            </div>
            <Input label="Email" type="email" value={form.email} onChange={e => set('email', e.target.value)} error={errors.email} placeholder="tu@email.com" />
            <Input label="Contraseña" type="password" value={form.password} onChange={e => set('password', e.target.value)} error={errors.password} placeholder="Mínimo 6 caracteres" hint="Usada para iniciar sesión" />
          </>
        )}

        {step === 2 && (
          <>
            <Select
              label="Especialidad"
              value={form.especialidad}
              onChange={e => set('especialidad', e.target.value)}
              error={errors.especialidad}
            >
              <option value="">Seleccioná una especialidad</option>
              {ESPECIALIDADES.map(e => <option key={e} value={e}>{e}</option>)}
            </Select>
            <Input
              label="Número de matrícula nacional"
              value={form.numeroMatricula}
              onChange={e => set('numeroMatricula', e.target.value)}
              error={errors.numeroMatricula}
              placeholder="Ej: 98.234"
              hint="Lo validaremos automáticamente contra REFEPS"
            />
          </>
        )}

        {step === 3 && (
          <>
            <div>
              <p className="text-sm font-medium text-gray-700 mb-2">Modalidades de atención</p>
              <div className="flex flex-col gap-2">
                {MODALIDADES.map(({ id, label, desc }) => {
                  const active = form.modalidades.includes(id)
                  return (
                    <button
                      key={id}
                      type="button"
                      onClick={() => toggleModalidad(id)}
                      className={`w-full text-left px-4 py-3 rounded-xl border-2 transition-all
                        ${active ? 'border-sky-500 bg-sky-50' : 'border-gray-100 hover:border-gray-200'}`}
                    >
                      <p className={`text-sm font-semibold ${active ? 'text-sky-700' : 'text-gray-700'}`}>{label}</p>
                      <p className="text-xs text-gray-400">{desc}</p>
                    </button>
                  )
                })}
              </div>
              {errors.modalidades && <p className="text-xs text-red-500 mt-1">{errors.modalidades}</p>}
            </div>

            <Select
              label="Zona principal de atención"
              value={form.zona}
              onChange={e => set('zona', e.target.value)}
              error={errors.zona}
            >
              <option value="">Seleccioná una zona</option>
              {ZONAS.map(z => <option key={z} value={z}>{z}</option>)}
            </Select>
          </>
        )}
      </div>

      {/* Footer */}
      <div className="px-6 py-6 flex gap-3">
        {step > 1 && (
          <Button variant="outline" size="lg" onClick={() => setStep(s => s - 1)} className="flex-1">
            Atrás
          </Button>
        )}
        <Button size="lg" onClick={next} className="flex-1">
          {step === TOTAL_STEPS ? 'Finalizar' : 'Continuar'}
        </Button>
      </div>
    </div>
  )
}
