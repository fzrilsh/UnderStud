import { Link } from 'react-router-dom'
import { ArrowRight, Download } from 'lucide-react'

import StatusBadge from '@/components/status-badge'
import { IconStudentCard } from '@/components/icons'
import MiniChart from '@/features/dashboard/components/mini-chart'
import type { ClassItem, DashboardBK, DashboardWK, SummaryView } from '@/features/dashboard/hooks'
import { fmtDateShortID } from '@/lib/format'
import { cn } from '@/lib/utils'

/** Kartu dasar dashboard (border + shadow). */
export function Panel({ className, children }: { className?: string; children: React.ReactNode }) {
  return (
    <div
      className={cn(
        'flex flex-col rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]',
        className,
      )}
    >
      {children}
    </div>
  )
}

/** Banner red flag (BK & WK). */
export function RedBanner({
  count,
  name,
  kelas,
  tanggal,
}: {
  count: number
  name: string
  kelas: string
  tanggal: string
}) {
  return (
    <div className="flex flex-col gap-4 overflow-hidden rounded-[10px] border-l-4 border-l-[#ef4444] bg-base px-5 py-5 shadow-[0px_0px_1px_0px_rgba(24,24,27,0.2),0px_1px_2px_0px_rgba(24,24,27,0.1)] md:flex-row md:items-center md:px-6">
      <div className="flex min-w-0 flex-1 items-start gap-4">
        <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-[#ef4444] text-[18px] font-bold text-base">
          !
        </span>
        <div className="min-w-0 flex-1">
          <p className="text-[16px] font-semibold leading-[1.5] text-violet-dark">
            {count} siswa perlu segera ditinjau
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-x-3.5 gap-y-1 text-[14px] text-[rgba(21,21,58,.6)] md:text-[16px]">
            <span>{name}</span>
            <span className="text-[rgba(21,21,58,.3)]">·</span>
            <span>{kelas}</span>
            <span className="text-[rgba(21,21,58,.3)]">·</span>
            <span>{tanggal}</span>
          </div>
        </div>
      </div>
      <Link
        to="/students?filter=red_flag"
        className="shrink-0 self-start rounded-[12px] bg-violet px-5 py-2.5 text-center text-[14px] font-medium text-base hover:bg-violet-hover md:self-auto"
      >
        Lihat di Daftar siswa
      </Link>
    </div>
  )
}

/** Metrik ber-bar (Disematkan/Draft/Antrean). */
function BarMetric({ value, label, color }: { value: number; label: string; color: string }) {
  return (
    <div className="flex items-center gap-3">
      <span className="h-[47px] w-[3px] shrink-0 rounded-sm" style={{ background: color }} />
      <span className="w-[62px]">
        <span className="block font-heading text-[24px] font-bold leading-[1.3] text-violet-dark">{value}</span>
        <span className="block text-[10.5px] font-medium leading-[1.5] text-[rgba(21,21,58,.8)]">{label}</span>
      </span>
    </div>
  )
}

/** Kartu "Chat bulan ini" (BK). */
export function ChatCard({ d }: { d: DashboardBK }) {
  return (
    <Panel className="lg:min-w-0 lg:flex-1">
      <div className="flex items-start justify-between gap-4">
        <h2 className="font-heading text-[32px] font-bold leading-[1.3] text-violet-dark">Chat bulan ini</h2>
        <Link
          to="/reports"
          className="shrink-0 rounded-[12px] bg-violet px-5 py-2.5 text-[14px] font-medium text-base hover:bg-violet-hover"
        >
          Mulai tinjau
        </Link>
      </div>
      <div className="mt-5 flex flex-wrap items-start gap-[18px]">
        <BarMetric value={d.chats.pinned} label="Disematkan" color="#ef4444" />
        <BarMetric value={d.chats.drafted} label="Draft" color="#16a34a" />
        <BarMetric value={d.chats.queued} label="Antrean" color="#3d3ba6" />
      </div>
      <div className="mt-5 h-[214px]">
        <MiniChart points={d.chart} />
      </div>
    </Panel>
  )
}

/** Kartu "Daftar Siswa" (BK). */
export function StudentShortcutCard({ redFlags }: { redFlags: number }) {
  return (
    <Panel className="justify-between lg:h-[389px] lg:w-[400px] lg:shrink-0">
      <div className="flex flex-col gap-3">
        <span className="flex size-11 shrink-0 items-center justify-center rounded-[10px] bg-[#ecebf6] text-violet-dark">
          <IconStudentCard />
        </span>
        <p className="font-heading text-[32px] font-bold leading-[1.3] text-violet-dark">Daftar Siswa</p>
        <p className="text-[14px] font-medium leading-[1.5] text-violet-dark opacity-70">
          Lihat data siswa per kelas. Akses Anda hanya untuk melihat.
        </p>
      </div>
      <div className="mt-6 flex items-center justify-between gap-3">
        <StatusBadge tone="red">{redFlags} Red flag</StatusBadge>
        <Link to="/students" className="whitespace-pre text-[14px] font-medium leading-[1.5] text-violet hover:underline">
          Buka Daftar Siswa →
        </Link>
      </div>
    </Panel>
  )
}

/** Badge status kelas (Lengkap/Ditandai/Belum diisi). */
export function ClassStatusBadge({ status }: { status: ClassItem['status'] }) {
  if (status === 'complete') return <StatusBadge tone="green">Lengkap</StatusBadge>
  if (status === 'flagged') return <StatusBadge tone="red">Ditandai</StatusBadge>
  return <StatusBadge tone="orange">Belum diisi</StatusBadge>
}

/** Kartu tabel "Data kelas" (BK). */
export function ClassTableCard({ classes }: { classes: ClassItem[] }) {
  return (
    <Panel className="min-w-0 flex-1 gap-5">
      <div>
        <div className="flex items-center justify-between gap-3">
          <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Data kelas</h2>
          <Link to="/students" className="whitespace-pre text-[14px] font-medium leading-[1.5] text-violet hover:underline">
            Lihat semua →
          </Link>
        </div>
        <p className="mt-1 text-[12px] leading-[1.5] text-violet-dark opacity-65">
          Daftar siswa diisi oleh masing-masing wali kelas. Red flag ditandai manual oleh guru.
        </p>
      </div>
      <div className="overflow-x-auto">
        <div className="min-w-[520px]">
          <div className="flex items-center gap-3 rounded-[8px] bg-[#ecebf6] px-3 py-2.5">
            <p className="w-[110px] shrink-0 text-[12px] font-semibold leading-[1.5] text-violet-dark opacity-70">Kelas</p>
            <p className="min-w-0 flex-1 text-[12px] font-semibold leading-[1.5] text-violet-dark opacity-70">Wali kelas</p>
            <p className="w-[64px] shrink-0 text-[12px] font-semibold leading-[1.5] text-violet-dark opacity-70">Siswa</p>
            <p className="w-[72px] shrink-0 text-[12px] font-semibold leading-[1.5] text-violet-dark opacity-70">Red flag</p>
            <p className="w-[104px] shrink-0 text-[12px] font-semibold leading-[1.5] text-violet-dark opacity-70">Status</p>
          </div>
          {classes.length === 0 && (
            <p className="px-3 py-6 text-center text-[13px] text-[rgba(21,21,58,.45)]">Belum ada kelas.</p>
          )}
          {classes.map((c) => (
            <div key={c.kelas} className="flex items-center gap-3 border-b border-violet-light px-3 py-[13px] last:border-b-0">
              <p className="w-[110px] shrink-0 text-[14px] font-semibold text-violet-dark">{c.kelas}</p>
              <p className="min-w-0 flex-1 text-[14px] text-violet-dark opacity-80">{c.homeroom || '—'}</p>
              <p className="w-[64px] shrink-0 text-[14px] font-medium text-violet-dark">{c.students || '—'}</p>
              <p className={cn('w-[72px] shrink-0 text-[14px] font-medium text-violet-dark', !c.red_flag && 'opacity-45')}>
                {c.red_flag || '0'}
              </p>
              <div className="w-[104px] shrink-0">
                <ClassStatusBadge status={c.status} />
              </div>
            </div>
          ))}
        </div>
      </div>
    </Panel>
  )
}

/** Kartu "Ringkasan terbaru" (BK). */
export function LatestSummariesCard({
  latest,
  onDownload,
}: {
  latest: DashboardBK['latest']
  onDownload: (studentId: string) => void
}) {
  return (
    <Panel className="gap-5 lg:w-[400px] lg:shrink-0">
      <div>
        <div className="flex items-center justify-between gap-3">
          <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Ringkasan terbaru</h2>
          <Link to="/students" className="whitespace-pre text-[14px] font-medium leading-[1.5] text-violet hover:underline">
            Lihat semua →
          </Link>
        </div>
        <p className="mt-0.5 text-[12px] leading-normal text-violet-dark opacity-65">
          Lihat atau unduh laporan ringkasan siswa
        </p>
      </div>
      {latest.length === 0 ? (
        <p className="py-8 text-center text-[13px] text-[rgba(21,21,58,.45)]">Belum ada ringkasan.</p>
      ) : (
        <div className="flex flex-col gap-4">
          {latest.map((x) => (
            <LatestRow key={x.summary.id} student={x.student} summary={x.summary} onDownload={onDownload} />
          ))}
        </div>
      )}
    </Panel>
  )
}

function LatestRow({
  student,
  summary,
  onDownload,
}: {
  student: DashboardBK['latest'][number]['student']
  summary: SummaryView
  onDownload: (studentId: string) => void
}) {
  return (
    <div className="flex flex-col gap-2 border-b border-violet-light pb-4 last:border-b-0 last:pb-0">
      <div className="flex items-center justify-between gap-3">
        <p className="text-[14px] leading-[1.5] text-violet-dark">
          <span className="font-semibold">{student.name}</span>{' '}
          <span className="font-medium opacity-65">· {student.kelas}</span>
        </p>
        <p className="shrink-0 text-[12px] text-violet-dark opacity-55">
          Versi {summary.version} · {fmtDateShortID(summary.created_at ?? '')}
        </p>
      </div>
      <div className="flex items-center justify-between gap-3">
        <p className="text-[12px] leading-[1.5] text-violet-dark opacity-60">oleh {summary.author || 'Wali Kelas'}</p>
        <div className="flex items-center gap-2">
          <Link
            to={`/students/${student.id}?tab=summary`}
            className="rounded-[8px] border border-violet-light bg-base px-3 py-1.5 text-[12px] font-semibold text-violet hover:bg-[#F6F6FB]"
          >
            Lihat
          </Link>
          <button
            type="button"
            onClick={() => onDownload(student.id)}
            className="flex items-center gap-1.5 rounded-[8px] bg-[#ecebf6] py-1.5 pl-2.5 pr-3 text-[12px] font-medium text-violet hover:bg-violet-light"
          >
            <Download className="size-3.5" /> Unduh
          </button>
        </div>
      </div>
    </div>
  )
}

/** Kartu statistik + chart (WK). */
export function WKStatCard({ d }: { d: DashboardWK }) {
  return (
    <Panel className="gap-5 lg:col-span-2">
      <div className="flex items-center justify-between gap-3">
        <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Statistik</h2>
        <a href="#wk-students" className="rounded-[8px] bg-violet px-3.5 py-2 text-[13px] font-medium text-base hover:bg-violet-hover">
          Lanjutkan observasi
        </a>
      </div>
      <div className="flex flex-wrap gap-8">
        <BarMetric value={d.e_rapor_complete} label="E-Rapor" color="#B42318" />
        <BarMetric value={d.observed} label="Diobservasi" color="#0B7A43" />
        <BarMetric value={d.students - d.e_rapor_complete} label="Belum diisi" color="#3D3BA6" />
      </div>
      <div className="h-[230px]">
        <MiniChart points={d.chart} />
      </div>
    </Panel>
  )
}

/** Kartu aksi cepat (WK). */
export function WKQuickActions({
  onImport,
  onAddStudent,
}: {
  onImport: () => void
  onAddStudent: () => void
}) {
  return (
    <Panel className="gap-3">
      <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Aksi cepat</h2>
      <div className="flex flex-col gap-2.5">
        <button
          type="button"
          onClick={onImport}
          className="flex w-full items-center gap-2 rounded-[8px] border border-violet-light bg-white px-3 py-2.5 text-[14px] font-medium text-violet-dark hover:bg-[#F6F6FB]"
        >
          Impor E-Rapor
        </button>
        <button
          type="button"
          onClick={onAddStudent}
          className="flex w-full items-center gap-2 rounded-[8px] border border-violet-light bg-white px-3 py-2.5 text-[14px] font-medium text-violet-dark hover:bg-[#F6F6FB]"
        >
          Tambah siswa
        </button>
        <a
          href="#wk-students"
          className="flex w-full items-center gap-1.5 rounded-[8px] bg-violet-dark px-3 py-2.5 text-[14px] font-medium text-base hover:bg-[#24245c]"
        >
          Tambah observasi
        </a>
      </div>
    </Panel>
  )
}

/** Kartu jumlah red flag (WK). */
export function WKRedFlagCard({ count }: { count: number }) {
  return (
    <Panel className="gap-2">
      <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">{count} Red Flag</h2>
      <a href="#wk-students" className="inline-flex items-center gap-1 text-[13px] font-semibold text-violet hover:underline">
        Lihat observasi <ArrowRight className="size-3.5" />
      </a>
    </Panel>
  )
}
