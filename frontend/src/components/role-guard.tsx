import { Loader2 } from 'lucide-react'
import { Navigate, Outlet } from 'react-router-dom'

import { useCurrentUser } from '@/features/auth/hooks'

/**
 * Guard per-peran: batasi halaman staf berdasarkan `user.role`.
 * Contoh: halaman Chat BK hanya untuk role BK.
 * Pakai `useCurrentUser` (cache React Query) — hindari race `useAuthStore`
 * yang di-set lewat useEffect (render pertama masih null → salah redirect).
 */
export default function RoleGuard({ roles }: { roles: Array<'BK' | 'WALI_KELAS'> }) {
  const user = useCurrentUser()

  // /me masih dimuat → jangan putuskan dulu.
  if (!user) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <Loader2 className="size-8 animate-spin text-[rgba(21,21,58,.4)]" />
      </div>
    )
  }
  if (!roles.includes(user.role)) return <Navigate to="/" replace />
  return <Outlet />
}
