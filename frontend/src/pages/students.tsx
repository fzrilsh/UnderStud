import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import Topbar from '@/components/layout/topbar'
import { IconClock, IconDotSm, IconTeaching, IconUsersGroup } from '@/components/icons'
import { ClassNav } from '@/features/students/components/class-nav'
import { ArchiveToggle, RedFlagToggle, StudentSearch } from '@/features/students/components/filter-chips'
import { StudentTableBK } from '@/features/students/components/student-table'
import { WKStudentsPanel } from '@/features/students/components/wk-students-panel'
import { Pagination } from '@/features/students/components/pagination'
import { useStudentCount, useStudents, type StudentFilter } from '@/features/students/hooks'
import { useCurrentUser } from '@/features/auth/hooks'
import { useClasses } from '@/features/dashboard/hooks'
import { api } from '@/lib/api'
import { downloadSummary } from '@/features/students/summary-export'
import { fmtDateID } from '@/lib/format'
import { useSeo } from '@/lib/seo'

const PAGE_SIZE = 10

export default function StudentsPage() {
  useSeo('Daftar Siswa', 'Data siswa, status E-Rapor, observasi, dan red flag.')

  const user = useCurrentUser()
  if (!user) return null
  return user.role === 'BK' ? <StudentsBK /> : <StudentsWK kelas={user.kelas ?? ''} />
}

// ---------------------------------------------------------------- BK variant
function StudentsBK() {
  const [params, setParams] = useSearchParams()
  const classes = useClasses()
  const list = classes.data ?? []

  const activeKelas = params.get('kelas') || list[0]?.kelas || ''
  const q = params.get('q') || ''
  const filter = (params.get('filter') || 'all') as StudentFilter
  const archived = params.get('archived') === '1'
  const offset = Number(params.get('offset') || 0)

  const setParam = (patch: Record<string, string | null>) => {
    const next = new URLSearchParams(params)
    for (const [k, v] of Object.entries(patch)) {
      if (v === null) next.delete(k)
      else next.set(k, v)
    }
    setParams(next)
  }

  const students = useStudents({
    kelas: activeKelas,
    q,
    filter,
    includeArchived: archived,
    limit: PAGE_SIZE,
    offset,
  })
  const redCount = useStudentCount('red_flag')
  const totalStudents = useStudentCount('all', undefined, true)

  const items = students.data?.items ?? []
  const total = students.data?.total ?? 0
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const page = Math.floor(offset / PAGE_SIZE) + 1

  const handleDownload = async (studentId: string) => {
    try {
      const st = items.find((s) => s.id === studentId)
      const res = await api.get<{ data: { id: string } }>(`/students/${studentId}/summaries/latest`)
      await downloadSummary(studentId, res.data.data.id, 'pdf', { nisn: st?.nisn, name: st?.name })
      toast.success('Ringkasan diunduh.')
    } catch {
      toast.error('Belum ada ringkasan untuk diunduh.')
    }
  }

  return (
    <>
      {/* Header BK: satu blok flex-col gap-6 dengan filter di dalamnya (mockup). */}
      <header className="flex flex-col gap-6 border-b border-[#c3c2e3] px-5 pb-6 pt-6 md:px-8 md:pt-9">
        <div className="flex items-start gap-3">
          <div className="min-w-0 flex-1 pl-12 md:pl-0">
            <h1 className="font-heading text-[28px] font-bold leading-[1.25] text-gradient-heading md:text-[36px]">
              Daftar Siswa
            </h1>
            <p className="mt-1.5 flex flex-wrap items-center gap-x-1 gap-y-0.5 text-[15px] text-violet-dark md:text-[16px]">
              <span className="whitespace-nowrap">Semua siswa di sekolah</span>
              <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
              <span className="whitespace-nowrap">{list.length} kelas</span>
              <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
              <span className="whitespace-nowrap">{totalStudents.data ?? 0} siswa</span>
            </p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <StudentSearch value={q} onChange={(v) => setParam({ q: v || null, offset: null })} />
          <RedFlagToggle
            count={redCount.data ?? 0}
            active={filter === 'red_flag'}
            onToggle={() => setParam({ filter: filter === 'red_flag' ? null : 'red_flag', offset: null })}
          />
          <ArchiveToggle
            checked={archived}
            onToggle={() => setParam({ archived: archived ? null : '1', offset: null })}
          />
        </div>
      </header>
      <div className="px-5 pb-10 pt-6 md:px-8">
        <div className="flex flex-col gap-4 lg:flex-row">
          <ClassNav
            classes={list}
            active={activeKelas}
            onSelect={(kelas) => setParam({ kelas, offset: null })}
          />
          <div className="min-w-0 flex-1">
            <div className="flex flex-col gap-5 rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
              <ClassDetailHead kelas={activeKelas} total={total} homeroom={list.find((c) => c.kelas === activeKelas)?.homeroom} />
              {students.isLoading && <p className="text-[13px] text-[rgba(21,21,58,.6)]">Memuat…</p>}
              {students.isError && <p className="text-[13px] text-destructive">Gagal memuat daftar siswa.</p>}
              {students.data && <StudentTableBK items={items} onDownload={handleDownload} />}
              <Pagination
                page={page}
                pages={pages}
                shown={items.length}
                total={total}
                onPrev={() => setParam({ offset: String(Math.max(0, offset - PAGE_SIZE)) })}
                onNext={() => setParam({ offset: String(offset + PAGE_SIZE) })}
              />
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

function ClassDetailHead({ kelas, total, homeroom }: { kelas: string; total: number; homeroom?: string }) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">{kelas}</h2>
        <div className="mt-1 flex flex-wrap items-center gap-3 text-[12px] leading-[1.5] text-violet-dark opacity-80">
          <span className="flex items-center gap-1.5">
            <IconTeaching />
            {homeroom || 'Belum ada wali kelas'}
          </span>
          <span className="flex items-center gap-1.5">
            <IconUsersGroup />
            {total} siswa
          </span>
          <span className="flex items-center gap-1.5">
            <IconClock />
            Diperbarui {fmtDateID(new Date().toISOString())}
          </span>
        </div>
      </div>
      <span className="shrink-0 rounded-[6px] bg-[#ecebf6] px-2 py-[3px] text-[12px] font-semibold text-violet">Hanya lihat</span>
    </div>
  )
}

// ---------------------------------------------------------------- WK variant
function StudentsWK({ kelas }: { kelas: string }) {
  const total = useStudentCount('all', kelas)
  return (
    <>
      <Topbar
        title="Daftar Siswa"
        subtitle={
          <>
            <span className="whitespace-nowrap">Kelas perwalian</span>
            <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
            <span className="whitespace-nowrap">{kelas}</span>
            <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
            <span className="whitespace-nowrap">{total.data ?? 0} siswa</span>
          </>
        }
      />
      <div className="px-5 pb-10 pt-6 md:px-8">
        <WKStudentsPanel kelas={kelas} />
      </div>
    </>
  )
}
