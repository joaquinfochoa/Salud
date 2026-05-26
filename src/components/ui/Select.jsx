export default function Select({ label, error, children, className = '', ...props }) {
  return (
    <div className={`flex flex-col gap-1 ${className}`}>
      {label && <label className="text-sm font-medium text-gray-700">{label}</label>}
      <select
        className={`w-full px-3.5 py-2.5 rounded-xl border text-sm bg-white transition-colors outline-none
          focus:ring-2 focus:ring-sky-500 focus:border-sky-500
          ${error ? 'border-red-400' : 'border-gray-200'}
        `}
        {...props}
      >
        {children}
      </select>
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
