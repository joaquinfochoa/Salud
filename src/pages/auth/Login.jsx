import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import Button from '../../components/ui/Button'
import Input from '../../components/ui/Input'

export default function Login() {
  const { login, loading, error, user } = useAuth()
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', password: '' })

  // Navegar después de que React confirme el nuevo user en el estado
  useEffect(() => {
    if (!user) return
    if (user.role === 'profesional') navigate('/profesional', { replace: true })
    else if (user.role === 'admin') navigate('/admin', { replace: true })
    else navigate('/buscar', { replace: true })
  }, [user])

  function handleChange(e) {
    setForm(f => ({ ...f, [e.target.name]: e.target.value }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    await login(form.email, form.password)
  }

  return (
    <div className="min-h-screen bg-white flex flex-col max-w-lg mx-auto px-6">
      <div className="flex-1 flex flex-col justify-center py-12">
        {/* Logo / marca */}
        <div className="mb-10">
          <div className="w-12 h-12 bg-sky-500 rounded-2xl flex items-center justify-center mb-4">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Bienvenido</h1>
          <p className="text-gray-400 text-sm mt-1">Iniciá sesión para continuar</p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <Input
            label="Email"
            name="email"
            type="email"
            placeholder="tu@email.com"
            value={form.email}
            onChange={handleChange}
            required
            autoComplete="email"
          />
          <Input
            label="Contraseña"
            name="password"
            type="password"
            placeholder="••••••••"
            value={form.password}
            onChange={handleChange}
            required
            autoComplete="current-password"
          />

          {error && (
            <p className="text-sm text-red-500 text-center">{error}</p>
          )}

          <Button type="submit" size="full" loading={loading} className="mt-2">
            Ingresar
          </Button>
        </form>

        <div className="mt-8 text-center text-sm text-gray-500">
          ¿No tenés cuenta?{' '}
          <Link to="/registro" className="text-sky-500 font-medium">
            Registrate
          </Link>
        </div>

        {/* Hint para desarrollo */}
        <div className="mt-10 p-3 bg-gray-50 rounded-xl border border-gray-100 text-xs text-gray-400">
          <p className="font-medium text-gray-500 mb-1">Accesos de prueba</p>
          <p>profesional@test.com / 123456</p>
          <p>paciente@test.com / 123456</p>
        </div>
      </div>
    </div>
  )
}
