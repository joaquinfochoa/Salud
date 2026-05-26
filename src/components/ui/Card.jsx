export default function Card({ children, className = '', onClick, ...props }) {
  const interactive = onClick ? 'cursor-pointer hover:shadow-md active:scale-[0.99] transition-all' : ''
  return (
    <div
      className={`bg-white rounded-2xl shadow-sm border border-gray-100 ${interactive} ${className}`}
      onClick={onClick}
      {...props}
    >
      {children}
    </div>
  )
}
