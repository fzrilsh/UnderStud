import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'

import Topbar from '@/components/layout/topbar'
import { api } from '@/lib/api'
import { downloadSummary } from '@/features/students/summary-export'
import {
  ChatCard,
  ClassTableCard,
  LatestSummariesCard,
  RedBanner,
  StudentShortcutCard,
} from '@/features/dashboard/components/dashboard-panels'
import { useDashboardBK } from '@/features/dashboard/hooks'
import { useCurrentUser } from '@/features/auth/hooks'
import DashboardWKPage from '@/pages/dashboard-wk'

export default function DashboardPage() {
  const user = useCurrentUser()
  // Tunggu /me selesai sebelum memilih varian dashboard (hindari request salah peran).
  if (!user) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <Loader2 className="size-8 animate-spin text-[rgba(21,21,58,.4)]" />
      </div>
    )
  }
  if (user.role === 'BK') return <DashboardBKPage />
  return <DashboardWKPage />
}

function DashboardBKPage() {
  const q = useDashboardBK()
  const d = q.data

  const handleDownload = async (studentId: string) => {
    // Ambil ringkasan terbaru siswa lalu ekspor PDF.
    try {
      const res = await api.get<{ data: { id: string } }>(`/students/${studentId}/summaries/latest`)
      await downloadSummary(studentId, res.data.data.id, 'pdf')
      toast.success('Ringkasan diunduh.')
    } catch {
      toast.error('Belum ada ringkasan untuk diunduh.')
    }
  }

  return (
    <>
      <Topbar
        title="Dashboard"
        subtitle={
          <span className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
            <span className="whitespace-nowrap">Ringkasan konseling</span>
            <span className="text-[rgba(21,21,58,.3)]">·</span>
            <span className="whitespace-nowrap">{new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}</span>
          </span>
        }
      />
      <div className="px-5 pb-10 md:px-8">
        {q.isLoading && <p className="text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</p>}
        {q.isError && <p className="text-[14px] text-destructive">Gagal memuat dashboard.</p>}
        {d && (
          <div className="flex flex-col gap-9">
            {d.red_top && (
              <RedBanner
                count={d.red_flags}
                name={d.red_top.student.name}
                kelas={d.red_top.student.kelas}
                tanggal={new Date(d.red_top.tanggal).toLocaleDateString('id-ID', {
                  day: 'numeric',
                  month: 'short',
                  year: 'numeric',
                })}
              />
            )}
            <div className="flex flex-col gap-4 lg:flex-row">
              <ChatCard d={d} />
              <StudentShortcutCard redFlags={d.red_flags} />
            </div>
            <div className="flex flex-col gap-4 lg:flex-row">
              <ClassTableCard classes={d.classes} />
              <LatestSummariesCard latest={d.latest} onDownload={handleDownload} />
            </div>
          </div>
        )}
      </div>
    </>
  )
}
