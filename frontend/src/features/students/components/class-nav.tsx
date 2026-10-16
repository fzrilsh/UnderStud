import { ChevronDown } from 'lucide-react'

import type { ClassItem } from '@/features/dashboard/hooks'
import { cn } from '@/lib/utils'

/** Panel navigasi kelas (varian BK): grup per tingkat (X/XI/XII) + jumlah siswa. */
export function ClassNav({
  classes,
  active,
  onSelect,
}: {
  classes: ClassItem[]
  active: string
  onSelect: (kelas: string) => void
}) {
  const grades = new Map<string, ClassItem[]>()
  for (const c of classes) {
    const g = c.grade || '—'
    if (!grades.has(g)) grades.set(g, [])
    grades.get(g)!.push(c)
  }
  const groups = [...grades.keys()].sort()

  return (
    <aside className="shrink-0 rounded-[10px] border border-violet-light bg-base px-4 py-5 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)] lg:h-fit lg:w-[248px]">
      <div className="flex flex-col gap-9">
        {groups.map((g) => (
          <div key={g} className="flex flex-col gap-1">
            <div className="flex items-center justify-between pb-2 pl-2">
              <p className="text-[12px] font-semibold leading-[1.5] text-violet-dark opacity-60">Kelas {g}</p>
              <ChevronDown className="size-4 text-[rgba(21,21,58,.5)]" />
            </div>
            {grades.get(g)!.map((c) => {
              const on = c.kelas === active
              return (
                <button
                  key={c.kelas}
                  type="button"
                  onClick={() => onSelect(c.kelas)}
                  className={cn(
                    'flex w-full items-center justify-between overflow-hidden rounded-[8px] px-2.5 py-2 text-left',
                    on ? 'border-l-2 border-l-[#ef4444] bg-violet-light' : 'hover:bg-[#ecebf6]',
                  )}
                >
                  <span className="text-[14px] font-medium leading-[1.5] text-violet-dark">{c.kelas}</span>
                  <span className="text-[12px] leading-[1.5] text-violet-dark opacity-60">
                    {c.students ? c.students : 'Belum diisi'}
                  </span>
                </button>
              )
            })}
          </div>
        ))}
        {classes.length === 0 && <p className="text-[13px] text-[rgba(21,21,58,.45)]">Belum ada kelas.</p>}
      </div>
    </aside>
  )
}
