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
              'inline-flex items-center gap-2 rounded-full border px-3.5 py-2 text-[13px] font-medium transition-colors',
              c.red
                ? 'border-[#fca5a5] text-[#ef4444] hover:bg-[#fef2f2]'
                : 'border-violet-light text-violet-dark hover:bg-[#F6F6FB]',
              on && (c.red ? 'bg-[#fee2e2] ring-2 ring-[#ef4444]/20' : 'bg-[#ecebf6] ring-2 ring-violet/20'),
            )}
          >
            {c.red && (
              <span className="flex size-[18px] items-center justify-center rounded-full bg-[#ef4444] text-[12px] font-bold text-base">
                !
              </span>
            )}
            {c.label}
            <span
              className={cn(
                'rounded-full px-2 py-[1px] text-[11.5px] font-semibold',
                c.red ? 'bg-[#ef4444] text-base' : 'bg-violet-light text-violet-dark',
              )}
            >
              {c.count}
            </span>
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

/** Input pencarian siswa. */
export function StudentSearch({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <div className="flex min-w-[240px] flex-1 items-center gap-2.5 rounded-[8px] border border-violet-light bg-base px-3.5 py-3">
      <svg width="16" height="16" viewBox="0 0 18 18" fill="none" className="shrink-0 text-[rgba(21,21,58,.5)]">
        <path
          d="M16.5 16.5 12.11 12.11M14.25 7.875a6.375 6.375 0 1 1-12.75 0 6.375 6.375 0 0 1 12.75 0Z"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Cari nama atau NISN siswa…"
        className="w-full bg-transparent text-[14px] text-violet-dark outline-none placeholder:text-violet-dark placeholder:opacity-50"
      />
    </div>
  )
}
