import { Outlet } from 'react-router-dom'

import Sidebar from '@/components/layout/sidebar'
import { useConfirm } from '@/components/confirm-provider'
import { useCurrentUser, useLogout } from '@/features/auth/hooks'

export default function AppLayout() {
  const user = useCurrentUser()
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
