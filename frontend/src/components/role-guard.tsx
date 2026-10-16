import { Navigate, Outlet } from 'react-router-dom'

import { useAuthStore } from '@/stores/auth-store'

/**
 * Guard per-peran: batasi halaman staf berdasarkan `user.role`.
 * Contoh: halaman Chat BK hanya untuk role BK.
 */
export default function RoleGuard({ roles }: { roles: Array<'BK' | 'WALI_KELAS'> }) {
  const user = useAuthStore((s) => s.user)

  if (!user) return <Navigate to="/login" replace />
  if (!roles.includes(user.role)) return <Navigate to="/" replace />
  return <Outlet />
}
