import { useSearchParams, useLocation } from 'react-router-dom'

import { useReports, type ReportItem } from '@/features/reports/hooks'
import { useCurrentUser } from '@/features/auth/hooks'
import { fmtHM } from '@/lib/format'
import { cn } from '@/lib/utils'

/** Section chat BK di sidebar (Disematkan/Draft/Antrean) — hanya di halaman /reports. */
export default function ChatSections() {
  const location = useLocation()
  const user = useCurrentUser()
  const [params, setParams] = useSearchParams()
  const { data } = useReports()

  // Hanya tampil untuk BK di halaman /reports.
  if (user?.role !== 'BK' || !location.pathname.startsWith('/reports')) return null

  const items = data?.items ?? []
  const pinned = items.filter((c) => c.pinned && !c.drafted && c.status !== 'dilewati')
  const drafts = items.filter((c) => c.drafted)
  const queue = items.filter((c) => (c.status === 'baru' || c.status === 'dibuka') && !c.pinned && !c.drafted)
  const activeId = params.get('chat')

  const open = (id: string) => {
    const next = new URLSearchParams(params)
    next.set('chat', id)
    setParams(next, { replace: true })
  }

  return (
    <>
      <Section title="Disematkan" items={pinned} activeId={activeId} onOpen={open} empty="Tidak ada chat disematkan." />
      <Section title="Draft" items={drafts} activeId={activeId} onOpen={open} empty="Tidak ada draft." />
      <Section title="Antrean" items={queue} activeId={activeId} onOpen={open} empty="Antrean kosong." />
    </>
  )
}

function Section({
  title,
  items,
  activeId,
  onOpen,
  empty,
}: {
  title: string
  items: ReportItem[]
  activeId: string | null
  onOpen: (id: string) => void
  empty: string
}) {
  return (
    <div className="mt-3">
      <p className="mb-1.5 text-[12px] font-semibold text-[rgba(21,21,58,.45)]">{title}</p>
      {items.length === 0 ? (
        <p className="px-2 py-1.5 text-[12px] text-[rgba(21,21,58,.45)]">{empty}</p>
      ) : (
        items.map((c) => {
          const label = c.identity.mode === 'profil' ? `${c.identity.name}${c.identity.kelas ? ` - ${c.identity.kelas}` : ''}` : 'Anonim'
          return (
            <button
              key={c.id}
              type="button"
              onClick={() => onOpen(c.id)}
              className={cn(
                'flex w-full items-start justify-between gap-2 rounded-[8px] px-3 py-1.5 text-left',
                c.id === activeId ? 'bg-[rgba(195,194,227,.6)]' : 'hover:bg-[#F6F6FB]',
              )}
            >
              <span className="flex min-w-0 flex-1 flex-col gap-0.5">
                <span className="truncate text-[14px] font-semibold leading-[1.5] text-[rgba(21,21,58,.8)]">{label}</span>
                <span className="truncate text-[12px] leading-[1.5] text-[rgba(21,21,58,.6)]">{c.last_message?.body ?? ''}</span>
              </span>
              <span className="shrink-0 pt-0.5 text-[12px] text-[rgba(21,21,58,.8)]">
                {fmtHM(c.last_message?.created_at ?? c.last_message_at)}
              </span>
            </button>
          )
        })
      )}
    </div>
  )
}
