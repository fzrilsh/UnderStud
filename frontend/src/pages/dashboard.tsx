import Topbar from '@/components/layout/topbar'
import { useAuthStore } from '@/stores/auth-store'

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)

  return (
    <>
      <Topbar
        title="Ringkasan konseling"
        subtitle={
          user?.role === 'BK'
            ? 'Guru BK · Semua kelas'
            : `Wali Kelas${user?.kelas ? ` · ${user.kelas}` : ''}`
        }
      />
      <div className="px-5 pb-10 md:px-8">
        <p className="text-sm text-muted-foreground">Dashboard sedang dibangun.</p>
      </div>
    </>
  )
}
