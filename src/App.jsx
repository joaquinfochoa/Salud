import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { ProtectedRoute } from './routes/ProtectedRoute'

import Login from './pages/auth/Login'
import Register from './pages/auth/Register'

// Profesional
import Dashboard from './pages/profesional/Dashboard'
import Registro from './pages/profesional/Registro'
import Validacion from './pages/profesional/Validacion'
import Perfil from './pages/profesional/Perfil'
import Agenda from './pages/profesional/Agenda'
import Consulta from './pages/profesional/Consulta'
import Cobro from './pages/profesional/Cobro'
import PerfilPublico from './pages/profesional/PerfilPublico'

// Paciente
import Onboarding from './pages/paciente/Onboarding'
import Busqueda from './pages/paciente/Busqueda'
import PerfilProfesional from './pages/paciente/PerfilProfesional'
import Reserva from './pages/paciente/Reserva'
import ConsultaPaciente from './pages/paciente/Consulta'
import PostConsulta from './pages/paciente/PostConsulta'

// Admin
import AdminDashboard from './pages/admin/Dashboard'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Públicas */}
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="/login" element={<Login />} />
          <Route path="/registro" element={<Register />} />

          {/* Flujo profesional */}
          <Route path="/profesional" element={
            <ProtectedRoute allowedRoles={['profesional']}>
              <Dashboard />
            </ProtectedRoute>
          } />
          <Route path="/profesional/onboarding" element={<Registro />} />
          <Route path="/profesional/validacion" element={<Validacion />} />
          <Route path="/profesional/perfil" element={
            <ProtectedRoute allowedRoles={['profesional']}>
              <Perfil />
            </ProtectedRoute>
          } />
          <Route path="/profesional/agenda" element={
            <ProtectedRoute allowedRoles={['profesional']}>
              <Agenda />
            </ProtectedRoute>
          } />
          <Route path="/profesional/consulta/:id" element={
            <ProtectedRoute allowedRoles={['profesional']}>
              <Consulta />
            </ProtectedRoute>
          } />
          <Route path="/profesional/cobro/:id" element={
            <ProtectedRoute allowedRoles={['profesional']}>
              <Cobro />
            </ProtectedRoute>
          } />

          {/* Perfil público del profesional — sin auth */}
          <Route path="/p/:slug" element={<PerfilPublico />} />

          {/* Flujo paciente — búsqueda pública, reserva sin bloqueo en demo */}
          <Route path="/buscar" element={<Onboarding />} />
          <Route path="/buscar/resultados" element={<Busqueda />} />
          <Route path="/buscar/:slug" element={<PerfilProfesional />} />
          <Route path="/reserva/:id" element={<Reserva />} />
          <Route path="/consulta/:id" element={
            <ProtectedRoute allowedRoles={['paciente']}>
              <ConsultaPaciente />
            </ProtectedRoute>
          } />
          <Route path="/consulta/:id/post" element={
            <ProtectedRoute allowedRoles={['paciente']}>
              <PostConsulta />
            </ProtectedRoute>
          } />

          {/* Admin */}
          <Route path="/admin" element={
            <ProtectedRoute allowedRoles={['admin']}>
              <AdminDashboard />
            </ProtectedRoute>
          } />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
