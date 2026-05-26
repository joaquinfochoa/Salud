export default function Button({ children, variant = 'primary', size = 'md', className = '', disabled, loading, ...props }) {
  const base = 'inline-flex items-center justify-center font-medium rounded-xl transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none'

  const variants = {
    primary:  'bg-sky-500 text-white hover:bg-sky-600 shadow-sm',
    secondary:'bg-emerald-500 text-white hover:bg-emerald-600 shadow-sm',
    outline:  'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50',
    ghost:    'text-gray-600 hover:bg-gray-100',
    danger:   'bg-red-500 text-white hover:bg-red-600',
  }

  const sizes = {
    sm: 'text-sm px-3 py-1.5 gap-1.5',
    md: 'text-sm px-4 py-2.5 gap-2',
    lg: 'text-base px-5 py-3 gap-2',
    full: 'text-base px-5 py-3.5 gap-2 w-full',
  }

  return (
    <button
      className={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <svg className="animate-spin -ml-1 mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      )}
      {children}
    </button>
  )
}
