import { Outlet } from 'react-router-dom'

import Sidebar from '@/components/layout/sidebar'
import { useConfirm } from '@/components/confirm-provider'
import { useLogout } from '@/features/auth/hooks'
import { useAuthStore } from '@/stores/auth-store'

export default function AppLayout() {
  const user = useAuthStore((s) => s.user)
  const logout = useLogout()
  const confirm = useConfirm()

  const handleLogout = async () => {
    const ok = await confirm({
      title: 'Keluar dari akun?',
      body: 'Kamu akan kembali ke halaman masuk.',
      confirmText: 'Keluar',
    })
    if (ok) logout.mutate()
  }

  return (
    <div className="flex min-h-svh bg-base">
      <Sidebar user={user} onLogout={handleLogout} />
      <main className="min-w-0 flex-1">
        <Outlet />
      </main>
    </div>
  )
}
