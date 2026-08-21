import { createContext, useContext, useState } from 'react'

const AuthContext = createContext(null)

const MOCK_USERS = {
  'profesional@test.com': { id: '1', role: 'profesional', nombre: 'Martín', apellido: 'González', especialidad: 'Psicología', matricula: 'MN 98.234' },
  'admin@test.com':       { id: '2', role: 'admin', nombre: 'Admin', apellido: '' },
  'paciente@test.com':    { id: '3', role: 'paciente', nombre: 'Laura', apellido: 'Martínez' },
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  async function login(email, password) {
    setLoading(true)
    setError(null)
    await new Promise(r => setTimeout(r, 800))
    const found = MOCK_USERS[email.toLowerCase()]
    if (found && password === '123456') {
      setUser(found)
      setLoading(false)
      return found
    }
    setError('Email o contraseña incorrectos')
    setLoading(false)
    return null
  }

  function logout() {
    setUser(null)
  }

  function setUserData(data) {
    setUser(data)
  }

  return (
    <AuthContext.Provider value={{ user, loading, error, login, logout, setUserData }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
