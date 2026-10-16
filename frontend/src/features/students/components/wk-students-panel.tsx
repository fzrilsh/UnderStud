import { useState } from 'react'
import { toast } from 'sonner'

import { useConfirm } from '@/components/confirm-provider'
import { FilterChips, StudentSearch } from '@/features/students/components/filter-chips'
import { StudentTableWK } from '@/features/students/components/student-table'
import { Pagination } from '@/features/students/components/pagination'
import {
  useArchiveStudent,
  useCreateStudent,
  useGenerateStudentSummary,
  useRestoreStudent,
  useStudentCount,
  useStudents,
  type StudentFilter,
} from '@/features/students/hooks'
import { AddStudentModal, ImportModal } from '@/features/students/components/student-modals'
import { fmtDateID } from '@/lib/format'
import { cn } from '@/lib/utils'

const PAGE_SIZE = 10

/**
 * Panel "Siswa · {kelas}" varian Wali Kelas — kartu putih berisi heading,
 * toolbar (search + chip filter), tabel, dan footer paginasi.
 * Dipakai di dua tempat (mockup juga begitu): dashboard WK (`#wk-students`)
 * dan halaman Daftar Siswa.
 */
export function WKStudentsPanel({ kelas, className }: { kelas: string; className?: string }) {
  const confirm = useConfirm()
  const [q, setQ] = useState('')
  const [filter, setFilter] = useState<StudentFilter>('all')
  const [offset, setOffset] = useState(0)
  const [generatingId, setGeneratingId] = useState<string>()
  const [addOpen, setAddOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)

  const students = useStudents({ kelas, q, filter, limit: PAGE_SIZE, offset })
  const total = useStudentCount('all', kelas)
  const incomplete = useStudentCount('incomplete', kelas)
  const notObserved = useStudentCount('not_observed', kelas)
  const redCount = useStudentCount('red_flag', kelas)
  const archived = useStudentCount('archived', kelas)

  const archive = useArchiveStudent()
  const restore = useRestoreStudent()
  const generate = useGenerateStudentSummary()
  const createStudent = useCreateStudent()

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
    <div className={cn('rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]', className)}>
      <div>
        <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Siswa · {kelas}</h2>
        <p className="text-[12px] text-[rgba(21,21,58,.65)]">
          Impor dari E-Rapor (.xlsx / .csv) atau tambah siswa manual · Impor terakhir{' '}
          {fmtDateID(new Date().toISOString())}
        </p>
      </div>

      {/* Search + chips dalam satu baris (mockup wkToolbar). */}
      <div className="my-3 flex flex-wrap items-center gap-2">
        <StudentSearch
          variant="wk"
          value={q}
          onChange={(v) => {
            setQ(v)
            setOffset(0)
          }}
        />
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
        note="Red flag ditandai manual lewat checklist observasi, bukan oleh AI. BK bisa melihat siswa yang ditandai."
        onPrev={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
        onNext={() => setOffset(offset + PAGE_SIZE)}
      />

      {/* Modal (dibuka dari tombol dashboard WK / import-ulang). */}
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
    </div>
  )
}
