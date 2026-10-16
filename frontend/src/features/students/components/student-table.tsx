import { Link } from 'react-router-dom'

import { IconAiMagic, IconArchive, IconDownloadVioletSm, IconTaskAdd, IconUnarchive } from '@/components/icons'
import type { StudentListItem } from '@/features/students/hooks'
import { cn } from '@/lib/utils'

/** Badge kecil status kolom (mockup `.badge`: teks violet; `todo`: bg violet-light). */
function Badge({ children, variant = 'done' }: { children: React.ReactNode; variant?: 'done' | 'todo' }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-[6px] px-2 py-[3px] text-[12px] font-semibold',
        variant === 'done' ? 'bg-[#ecebf6] text-violet' : 'bg-violet-light text-violet-dark',
      )}
    >
      {children}
    </span>
  )
}

/** Badge "Red Flag" dengan dot (mockup `.badge-redflag`: teks merah). */
export function RedFlagBadge() {
  return (
    <span className="inline-flex items-center gap-2 rounded-[6px] bg-[#fee2e2] px-2 py-[3px]">
      <span className="size-2.5 rounded-full bg-[#dc2626]" />
      <span className="text-[12px] font-semibold text-[#ef4444]">Red Flag</span>
    </span>
  )
}

function StatusCell({ ok, doneLabel, todoLabel }: { ok: boolean; doneLabel: string; todoLabel: string }) {
  return ok ? <Badge>{doneLabel}</Badge> : <Badge variant="todo">{todoLabel}</Badge>
}

/** Tombol outline kecil (mockup `.tbl-btn-outline`). */
const btnOutline =
  'inline-flex items-center gap-1.5 rounded-[8px] border border-violet-light bg-base px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-[#F6F6FB]'

/** Tabel daftar siswa varian WK (kolom status + aksi). */
export function StudentTableWK({
  items,
  onGenerate,
  onArchive,
  onRestore,
  generatingId,
}: {
  items: StudentListItem[]
  onGenerate: (id: string) => void
  onArchive: (id: string) => void
  onRestore: (id: string) => void
  generatingId?: string
}) {
  if (items.length === 0) {
    return <p className="px-4 py-12 text-center text-[14px] text-[rgba(21,21,58,.45)]">Tidak ada siswa yang cocok.</p>
  }
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[880px] border-collapse">
        <thead>
          {/* Mockup `.table-ui thead th`: bg #F6F6FB, padding 10x14, nowrap. */}
          <tr>
            {['Siswa / NISN', 'E-Rapor', 'Observasi', 'Ringkasan AI', 'Red flag', ''].map((h, i) => (
              <th
                key={h}
                className={cn(
                  'whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-[12px] font-semibold text-[rgba(21,21,58,.55)]',
                  i === 5 ? 'text-right' : 'text-left',
                )}
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {items.map((st) => (
            <tr
              key={st.id}
              className={cn('border-b border-[#F1F1F5] hover:bg-[#FAFAFE]', st.has_red_flag && 'border-l-2 border-l-[#ef4444]')}
            >
              <td className="px-3.5 py-3">
                <div className="flex flex-col gap-1">
                  <p className="text-[14px] font-semibold text-violet-dark">{st.name}</p>
                  <p className="text-[12px] text-[rgba(21,21,58,.6)]">
                    {st.nisn ? `NISN ${st.nisn}` : 'Tanpa NISN · Ditambah manual'}
                  </p>
                </div>
              </td>
              <td className="px-3.5 py-3">
                <StatusCell ok={!!st.e_rapor_complete} doneLabel="Lengkap" todoLabel={st.nisn ? 'Belum lengkap' : 'Belum diimpor'} />
              </td>
              <td className="px-3.5 py-3">
                <StatusCell ok={!!st.observed} doneLabel="Diobservasi" todoLabel="Belum diobservasi" />
              </td>
              <td className="px-3.5 py-3">
                {st.has_summary ? (
                  <Badge>Dibuat</Badge>
                ) : (
                  // Mockup `.tbl-btn`: filled #ecebf6, teks violet, ikon aiMagic.
                  <button
                    type="button"
                    onClick={() => onGenerate(st.id)}
                    disabled={generatingId === st.id}
                    className="inline-flex items-center gap-1.5 rounded-[8px] bg-[#ecebf6] px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-violet-light disabled:opacity-50"
                  >
                    <IconAiMagic />
                    {generatingId === st.id ? 'Membuat…' : 'Buat'}
                  </button>
                )}
              </td>
              <td className="px-3.5 py-3">
                {st.has_red_flag ? <RedFlagBadge /> : <span className="text-[rgba(21,21,58,.4)]">—</span>}
              </td>
              <td className="px-3.5 py-3">
                <div className="flex items-center justify-end gap-2">
                  <Link to={`/students/${st.id}#observation`} className="inline-flex items-center gap-1.5 rounded-[8px] bg-[#ecebf6] px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-violet-light">
                    <IconTaskAdd /> Observasi
                  </Link>
                  {st.has_summary ? (
                    <Link to={`/students/${st.id}?tab=summary`} className={btnOutline}>
                      Lihat
                    </Link>
                  ) : (
                    <button type="button" disabled className={cn(btnOutline, 'cursor-not-allowed opacity-40')}>
                      Lihat
                    </button>
                  )}
                  {st.archived_at ? (
                    <button
                      type="button"
                      title="Buka arsip"
                      onClick={() => onRestore(st.id)}
                      className="flex size-8 items-center justify-center rounded-[8px] border border-violet-light bg-base text-violet hover:bg-[#F6F6FB]"
                    >
                      <IconUnarchive />
                    </button>
                  ) : (
                    <button
                      type="button"
                      title="Arsipkan siswa"
                      onClick={() => onArchive(st.id)}
                      className="flex size-8 items-center justify-center rounded-[8px] border border-violet-light bg-base text-violet hover:bg-[#F6F6FB]"
                    >
                      <IconArchive />
                    </button>
                  )}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/** Tabel detail kelas varian BK (kolom observasi/ringkasan/tanda guru). */
export function StudentTableBK({
  items,
  onDownload,
}: {
  items: StudentListItem[]
  onDownload: (id: string) => void
}) {
  if (items.length === 0) {
    return <p className="px-4 py-12 text-center text-[14px] text-[rgba(21,21,58,.45)]">Tidak ada siswa yang cocok.</p>
  }
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[720px] border-collapse">
        <thead>
          {/* Mockup `.table-ui thead.th-violet th`: bg #ecebf6 + sudut membulat di ujung. */}
          <tr className="bg-[#ecebf6]">
            <th className="rounded-l-[8px] px-3 py-2.5 text-left text-[12px] font-semibold text-[rgba(21,21,58,.7)]">
              Siswa / NISN
            </th>
            <th className="w-[130px] px-3 py-2.5 text-left text-[12px] font-semibold text-[rgba(21,21,58,.7)]">
              Observasi terakhir
            </th>
            <th className="w-[234px] px-3 py-2.5 text-left text-[12px] font-semibold text-[rgba(21,21,58,.7)]">
              Ringkasan AI
            </th>
            <th className="w-[170px] rounded-r-[8px] px-3 py-2.5 text-left text-[12px] font-semibold text-[rgba(21,21,58,.7)]">
              Tanda guru
            </th>
          </tr>
        </thead>
        <tbody>
          {items.map((st) => (
            <tr
              key={st.id}
              className={cn('border-b border-[#F1F1F5] hover:bg-[#FAFAFE]', st.has_red_flag && 'border-l-4 border-l-[#ef4444]')}
            >
              <td className="px-3 py-3">
                <div className="flex flex-col gap-1">
                  <p className="text-[14px] font-semibold text-violet-dark">{st.name}</p>
                  <p className="text-[12px] text-[rgba(21,21,58,.6)]">{st.nisn ? `NISN ${st.nisn}` : 'Tanpa NISN'}</p>
                </div>
              </td>
              <td className="px-3 py-3">
                {st.observed ? (
                  <Link to={`/students/${st.id}#observation`} className={btnOutline}>
                    Lihat
                  </Link>
                ) : (
                  <span className="text-[12px] text-[rgba(21,21,58,.45)]">—</span>
                )}
              </td>
              <td className="px-3 py-3">
                {st.has_summary ? (
                  <div className="flex items-center gap-2">
                    <Link to={`/students/${st.id}?tab=summary`} className={btnOutline}>
                      Lihat
                    </Link>
                    <button
                      type="button"
                      onClick={() => onDownload(st.id)}
                      className="flex items-center gap-1.5 rounded-[8px] bg-[#ecebf6] py-1.5 pl-2.5 pr-3 text-[12px] font-medium text-violet hover:bg-violet-light"
                    >
                      <IconDownloadVioletSm /> Unduh
                    </button>
                  </div>
                ) : (
                  <span className="text-[12px] text-[rgba(21,21,58,.5)]">Belum dibuat</span>
                )}
              </td>
              <td className="px-3 py-3">{st.has_red_flag ? <RedFlagBadge /> : ''}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
