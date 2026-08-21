export default function Badge({ children, variant = 'default', className = '' }) {
  const variants = {
    default:  'bg-gray-100 text-gray-600',
    primary:  'bg-sky-100 text-sky-700',
    success:  'bg-emerald-100 text-emerald-700',
    warning:  'bg-amber-100 text-amber-700',
    danger:   'bg-red-100 text-red-700',
    purple:   'bg-purple-100 text-purple-700',
  }
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${variants[variant]} ${className}`}>
      {children}
    </span>
  )
}
