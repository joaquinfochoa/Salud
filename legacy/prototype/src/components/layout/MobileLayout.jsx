import BottomNav from './BottomNav'

export default function MobileLayout({ children, withNav = true }) {
  return (
    <div className="max-w-lg mx-auto min-h-screen bg-gray-50 relative">
      <div className={withNav ? 'pb-20' : ''}>
        {children}
      </div>
      {withNav && <BottomNav />}
    </div>
  )
}
