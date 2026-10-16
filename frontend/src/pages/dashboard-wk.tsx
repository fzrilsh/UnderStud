import Topbar from '@/components/layout/topbar'
import {
  RedBanner,
  WKQuickActions,
  WKRedFlagCard,
  WKStatCard,
} from '@/features/dashboard/components/dashboard-panels'
import { useDashboardWK } from '@/features/dashboard/hooks'
import { useCurrentUser } from '@/features/auth/hooks'
import { useSeo } from '@/lib/seo'
import { fmtDateID } from '@/lib/format'

export default function DashboardWKPage() {
  useSeo('Dashboard', 'Statistik kelas perwalian: E-Rapor, observasi, dan red flag.')

  const user = useCurrentUser()
  const q = useDashboardWK()
  const d = q.data

  return (
    <>
      <Topbar
        variant="plain"
        title="Dashboard"
        subtitle={`Kelas perwalian · ${user?.kelas ?? ''} · ${fmtDateID(new Date().toISOString())}`}
      />
      <div className="px-5 pb-10 md:px-8">
        {q.isLoading && <p className="text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</p>}
        {q.isError && <p className="text-[14px] text-destructive">Gagal memuat dashboard.</p>}
        {d && (
          <div className="flex flex-col gap-6">
            {d.red_top && (
              <RedBanner
                count={d.red_flags}
                name={d.red_top.name}
                kelas={d.red_top.kelas}
                tanggal={fmtDateID(new Date().toISOString())}
              />
            )}
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
              <WKStatCard d={d} />
              <div className="flex flex-col gap-6">
                <WKQuickActions
                  onImport={() => {
                    // Impor leger diimplementasikan di halaman Daftar siswa (T4).
                    window.location.href = '/students?import=1'
                  }}
                  onAddStudent={() => {
                    window.location.href = '/students?add=1'
                  }}
                />
                <WKRedFlagCard count={d.red_flags} />
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
