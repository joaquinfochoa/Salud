import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'

export default function TopBar({ title, subtitle, backTo, actions }) {
  const navigate = useNavigate()

  function handleBack() {
    if (backTo) navigate(backTo)
    else navigate(-1)
  }

  return (
    <header className="sticky top-0 bg-white border-b border-gray-100 z-30">
      <div className="max-w-lg mx-auto px-4 h-14 flex items-center gap-3">
        {backTo !== undefined && (
          <button onClick={handleBack} className="p-1.5 -ml-1.5 rounded-xl text-gray-500 hover:bg-gray-100 transition-colors">
            <ArrowLeft size={20} />
          </button>
        )}
        <div className="flex-1 min-w-0">
          <h1 className="text-base font-semibold text-gray-900 leading-tight truncate">{title}</h1>
          {subtitle && <p className="text-xs text-gray-400 truncate">{subtitle}</p>}
        </div>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
    </header>
  )
}
