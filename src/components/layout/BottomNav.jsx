import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Calendar, UserCircle, Settings } from 'lucide-react'

const NAV_ITEMS = [
  { to: '/profesional',        icon: LayoutDashboard, label: 'Inicio' },
  { to: '/profesional/agenda', icon: Calendar,         label: 'Agenda' },
  { to: '/profesional/perfil', icon: UserCircle,       label: 'Perfil' },
]

export default function BottomNav() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-100 safe-area-pb z-40">
      <div className="max-w-lg mx-auto flex">
        {NAV_ITEMS.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/profesional'}
            className={({ isActive }) =>
              `flex-1 flex flex-col items-center gap-0.5 py-2.5 text-xs font-medium transition-colors
               ${isActive ? 'text-sky-500' : 'text-gray-400 hover:text-gray-600'}`
            }
          >
            <Icon size={22} strokeWidth={1.75} />
            {label}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
