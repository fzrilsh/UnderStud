import { Loader2 } from 'lucide-react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useMe } from '@/features/auth/hooks'

/**
 * Guard untuk route protected.
 * - /me pending → spinner
 * - sukses → render child (Outlet)
 * - 401/error → Navigate ke /login dengan state.from (balik ke sini setelah login)
 */
export default function ProtectedRoute() {
  const location = useLocation()
  const me = useMe()

  if (me.isPending) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (me.isError) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <Outlet />
}
