import { useState } from 'react'
import { toast } from 'sonner'

import Topbar from '@/components/layout/topbar'
import {
  RedBanner,
  WKQuickActions,
  WKRedFlagCard,
  WKStatCard,
} from '@/features/dashboard/components/dashboard-panels'
import { AddStudentModal, ImportModal } from '@/features/students/components/student-modals'
import { WKStudentsPanel } from '@/features/students/components/wk-students-panel'
import { useCreateStudent } from '@/features/students/hooks'
import { useDashboardWK } from '@/features/dashboard/hooks'
import { useCurrentUser } from '@/features/auth/hooks'
import { useSeo } from '@/lib/seo'
import { fmtDateID } from '@/lib/format'

export default function DashboardWKPage() {
  useSeo('Dashboard', 'Statistik kelas perwalian: E-Rapor, observasi, dan red flag.')

  const user = useCurrentUser()
  const q = useDashboardWK()
  const d = q.data

  const kelas = user?.kelas ?? ''
  const [addOpen, setAddOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const createStudent = useCreateStudent()

  // "Tambah observasi"/"Lanjutkan observasi"/"Lihat observasi" menggulir ke
  // tabel siswa di bawah (mockup: `location.hash = 'wk-students'`).
  const scrollToStudents = () => {
    document.getElementById('wk-students')?.scrollIntoView({ behavior: 'smooth' })
  }

  return (
    <>
      <Topbar
        variant="plain"
        title="Dashboard"
        subtitle={`Kelas perwalian · ${kelas} · ${fmtDateID(new Date().toISOString())}`}
      />
      <div className="px-5 pb-10 pt-6 md:px-8">
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
              <WKStatCard d={d} onContinue={scrollToStudents} />
              <div className="flex flex-col gap-6">
                <WKQuickActions
                  onImport={() => setImportOpen(true)}
                  onAddStudent={() => setAddOpen(true)}
                  onAddObservation={scrollToStudents}
                />
                <WKRedFlagCard count={d.red_flags} onViewObservations={scrollToStudents} />
              </div>
            </div>

            {/* Tabel "Siswa · {kelas}" (mockup: kartu penuh di bawah grid). */}
            <div id="wk-students" className="scroll-mt-6">
              <WKStudentsPanel kelas={kelas} />
            </div>
          </div>
        )}
      </div>

      {addOpen && (
        <AddStudentModal
          kelas={kelas}
          pending={createStudent.isPending}
          onSubmit={(payload) =>
            createStudent.mutate(payload, {
              onSuccess: () => {
                toast.success('Siswa ditambahkan.')
                setAddOpen(false)
              },
            })
          }
          onClose={() => setAddOpen(false)}
        />
      )}
      {importOpen && <ImportModal kelas={kelas} onClose={() => setImportOpen(false)} />}
    </>
  )
}
