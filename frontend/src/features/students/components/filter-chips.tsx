import { IconSearchSm } from '@/components/icons'
import { cn } from '@/lib/utils'
import type { StudentFilter } from '@/features/students/hooks'

interface Chip {
  key: StudentFilter
  label: string
  count: number
  red?: boolean
}

/** Baris chip filter (varian WK): Semua / E-Rapor / Observasi / Red flag / Diarsipkan. */
export function FilterChips({
  chips,
  active,
  onToggle,
}: {
  chips: Chip[]
  active: StudentFilter
  onToggle: (f: StudentFilter) => void
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {chips.map((c) => {
        const on = active === c.key
        return (
          <button
            key={c.key}
            type="button"
            onClick={() => onToggle(c.key)}
            aria-pressed={on}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-[8px] px-3 py-2.5 whitespace-nowrap transition-colors',
              // Red chip: selalu putih + border merah, teks 14px medium merah.
              c.red
                ? 'border border-[#ef4444] bg-[#fcfcfe] text-[14px] font-medium text-[#ef4444] hover:bg-[#fef2f2]'
                : on
                  ? // Aktif: bg #e2e2f2 tanpa border.
                    'bg-[#e2e2f2] text-[14px] font-semibold text-violet-dark'
                  : // Nonaktif: putih + border #E2E2F2.
                    'border border-[#E2E2F2] bg-white text-[14px] font-semibold text-[rgba(21,21,58,.7)] hover:bg-[#F6F6FB]',
            )}
          >
            {c.label}
            {c.red ? (
              <span className="flex size-5 items-center justify-center rounded-full bg-[#ef4444] text-[12px] font-semibold text-[#fcfcfe]">
                {c.count}
              </span>
            ) : (
              <span className="text-[12px] font-normal opacity-60">{c.count}</span>
            )}
          </button>
        )
      })}
    </div>
  )
}

/** Toggle "Tampilkan arsip" (varian BK). */
export function ArchiveToggle({ checked, onToggle }: { checked: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className="flex shrink-0 items-center gap-2 rounded-[8px] border border-violet-light bg-base px-3.5 py-3 hover:bg-[#F6F6FB]"
    >
      <span
        className={cn(
          'flex size-4 shrink-0 items-center justify-center rounded-[4px] border-[1.5px] border-violet',
          checked ? 'bg-violet' : 'bg-transparent',
        )}
      >
        {checked && (
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
            <path d="M2 5.2 4 7.2 8 3" stroke="#fcfcfe" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        )}
      </span>
      <span className="text-[14px] font-medium text-violet-dark">Tampilkan arsip</span>
    </button>
  )
}

/** Tombol "Hanya red flag" (varian BK). */
export function RedFlagToggle({ count, active, onToggle }: { count: number; active: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={cn(
        'flex shrink-0 items-center gap-2 rounded-[8px] border border-[#ef4444] bg-base px-3.5 py-3 hover:bg-[#fef2f2]',
        active && 'ring-2 ring-[#ef4444]/20',
      )}
    >
      <span className="flex size-[18px] items-center justify-center rounded-full bg-[#ef4444] text-[12px] font-bold text-base">
        !
      </span>
      <span className="text-[14px] font-medium text-[#ef4444]">Hanya red flag ({count})</span>
    </button>
  )
}

/**
 * Input pencarian siswa. `bk` = gaya header BK (radius 8, bg base);
 * `wk` = gaya `.input` mockup (radius 10, bg putih, min-w 200).
 */
export function StudentSearch({
  value,
  onChange,
  variant = 'bk',
}: {
  value: string
  onChange: (v: string) => void
  variant?: 'bk' | 'wk'
}) {
  const bk = variant === 'bk'
  return (
    <div
      className={cn(
        'flex flex-1 items-center gap-2.5',
        bk
          ? 'min-w-[240px] rounded-[8px] border border-[#e2e2f2] bg-base px-3.5 py-3'
          : 'relative min-w-[200px]',
      )}
    >
      {!bk && (
        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[rgba(21,21,58,.55)]">
          <IconSearchSm />
        </span>
      )}
      {bk && <IconSearchSm className="shrink-0 text-[rgba(21,21,58,.55)]" />}
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Cari nama atau NISN siswa…"
        className={cn(
          'w-full bg-transparent text-[14px] text-violet-dark outline-none placeholder:text-violet-dark placeholder:opacity-50',
          !bk && 'input pl-10',
        )}
      />
    </div>
  )
}
