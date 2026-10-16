import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import Topbar from '@/components/layout/topbar'
import { useConfirm } from '@/components/confirm-provider'
import { ClassNav } from '@/features/students/components/class-nav'
import { ArchiveToggle, FilterChips, RedFlagToggle, StudentSearch } from '@/features/students/components/filter-chips'
import { StudentTableBK, StudentTableWK } from '@/features/students/components/student-table'
import {
  useArchiveStudent,
  useGenerateStudentSummary,
  useRestoreStudent,
  useStudentCount,
  useStudents,
  type StudentFilter,
} from '@/features/students/hooks'
// useArchiveStudent dipakai di varian WK.
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
        title="Daftar Siswa"
        subtitle={
          <span className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
            <span className="whitespace-nowrap">Semua siswa di sekolah</span>
            <span className="text-[rgba(21,21,58,.3)]">·</span>
            <span className="whitespace-nowrap">{list.length} kelas</span>
            <span className="text-[rgba(21,21,58,.3)]">·</span>
            <span className="whitespace-nowrap">{totalStudents.data ?? 0} siswa</span>
          </span>
        }
        actions={
          <div className="flex w-full flex-wrap items-center gap-3 lg:w-auto">
            <StudentSearch
              value={q}
              onChange={(v) => setParam({ q: v || null, offset: null })}
            />
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
        }
      />
      <div className="px-5 pb-10 md:px-8">
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
              <p className="text-[12px] text-[rgba(21,21,58,.65)]">
                Red flag ditandai manual lewat checklist observasi, bukan oleh AI. BK bisa melihat siswa yang ditandai.
              </p>
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
          <span>{homeroom || 'Belum ada wali kelas'}</span>
          <span>·</span>
          <span>{total} siswa</span>
          <span>·</span>
          <span>Diperbarui {fmtDateID(new Date().toISOString())}</span>
        </div>
      </div>
      <span className="shrink-0 rounded-[6px] bg-[#ecebf6] px-2 py-[3px] text-[12px] font-semibold text-violet">Hanya lihat</span>
    </div>
  )
}

// ---------------------------------------------------------------- WK variant
function StudentsWK({ kelas }: { kelas: string }) {
  const confirm = useConfirm()
  const [q, setQ] = useState('')
  const [filter, setFilter] = useState<StudentFilter>('all')
  const [offset, setOffset] = useState(0)
  const [generatingId, setGeneratingId] = useState<string>()

  const students = useStudents({ kelas, q, filter, limit: PAGE_SIZE, offset })
  const total = useStudentCount('all', kelas)
  const incomplete = useStudentCount('incomplete', kelas)
  const notObserved = useStudentCount('not_observed', kelas)
  const redCount = useStudentCount('red_flag', kelas)
  const archived = useStudentCount('archived', kelas)

  const archive = useArchiveStudent()
  const restore = useRestoreStudent()
  const generate = useGenerateStudentSummary()

  const items = students.data?.items ?? []
  const count = students.data?.total ?? 0
  const pages = Math.max(1, Math.ceil(count / PAGE_SIZE))
  const page = Math.floor(offset / PAGE_SIZE) + 1

  const handleArchive = async (id: string) => {
    const st = items.find((s) => s.id === id)
    const ok = await confirm({
      title: 'Arsipkan siswa?',
      body: `${st?.name ?? 'Siswa'} akan diarsipkan. Data tetap tersimpan dan bisa dipulihkan.`,
      confirmText: 'Arsipkan',
      danger: true,
    })
    if (ok) archive.mutate(id)
  }

  const handleRestore = async (id: string) => {
    const st = items.find((s) => s.id === id)
    const ok = await confirm({
      title: 'Buka arsip siswa?',
      body: `${st?.name ?? 'Siswa'} akan dikembalikan ke daftar aktif.`,
      confirmText: 'Buka arsip',
    })
    if (ok) restore.mutate(id)
  }

  const handleGenerate = (id: string) => {
    setGeneratingId(id)
    generate.mutate(id, { onSettled: () => setGeneratingId(undefined) })
  }

  const toggleFilter = (f: StudentFilter) => {
    setFilter((cur) => (cur === f ? 'all' : f))
    setOffset(0)
  }

  return (
    <>
      <Topbar
        title="Daftar Siswa"
        subtitle={`Kelas perwalian · ${kelas} · ${total.data ?? 0} siswa`}
      />
      <div className="px-5 pb-10 md:px-8">
        <div className="rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
          <div>
            <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Siswa · {kelas}</h2>
            <p className="text-[12px] text-[rgba(21,21,58,.65)]">
              Impor dari E-Rapor (.xlsx / .csv) atau tambah siswa manual
            </p>
          </div>

          <div className="mt-5">
            <StudentSearch
              value={q}
              onChange={(v) => {
                setQ(v)
                setOffset(0)
              }}
            />
          </div>

          <div className="mt-3">
            <FilterChips
              active={filter}
              onToggle={toggleFilter}
              chips={[
                { key: 'all', label: 'Semua', count: total.data ?? 0 },
                { key: 'incomplete', label: 'E-Rapor belum lengkap', count: incomplete.data ?? 0 },
                { key: 'not_observed', label: 'Belum diobservasi', count: notObserved.data ?? 0 },
                { key: 'red_flag', label: 'Red flag', count: redCount.data ?? 0, red: true },
                { key: 'archived', label: 'Diarsipkan', count: archived.data ?? 0 },
              ]}
            />
          </div>

          <div className="mt-4">
            {students.isLoading && <p className="text-[13px] text-[rgba(21,21,58,.6)]">Memuat…</p>}
            {students.isError && <p className="text-[13px] text-destructive">Gagal memuat daftar siswa.</p>}
            {students.data && (
              <StudentTableWK
                items={items}
                onGenerate={handleGenerate}
                onArchive={handleArchive}
                onRestore={handleRestore}
                generatingId={generatingId}
              />
            )}
          </div>

          <Pagination
            page={page}
            pages={pages}
            shown={items.length}
            total={count}
            onPrev={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            onNext={() => setOffset(offset + PAGE_SIZE)}
          />
        </div>
      </div>
    </>
  )
}

function Pagination({
  page,
  pages,
  shown,
  total,
  onPrev,
  onNext,
}: {
  page: number
  pages: number
  shown: number
  total: number
  onPrev: () => void
  onNext: () => void
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <p className="text-[12px] text-[rgba(21,21,58,.65)]">
        Menampilkan {shown} dari {total}
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onPrev}
          disabled={page <= 1}
          className="rounded-[8px] bg-[#ecebf6] px-3 py-2 text-[12px] font-medium text-violet hover:bg-violet-light disabled:opacity-40"
        >
          Sebelumnya
        </button>
        <span className="px-3 py-2 text-[12px] font-medium text-violet-dark">
          {page} / {pages}
        </span>
        <button
          type="button"
          onClick={onNext}
          disabled={page >= pages}
          className="rounded-[8px] bg-[#ecebf6] px-3 py-2 text-[12px] font-medium text-violet hover:bg-violet-light disabled:opacity-40"
        >
          Berikutnya
        </button>
      </div>
    </div>
  )
}
