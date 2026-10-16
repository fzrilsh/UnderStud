import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Send } from 'lucide-react'

import { useConfirm } from '@/components/confirm-provider'
import {
  useDraftReport,
  useHideReport,
  useNextReport,
  usePinReport,
  useReplyReport,
  useReport,
  useReports,
  useSkipReport,
  useUndraftReport,
  useUnpinReport,
} from '@/features/reports/hooks'
import { useCurrentUser } from '@/features/auth/hooks'
import { fmtDateShortID, fmtHM } from '@/lib/format'
import { cn } from '@/lib/utils'

export default function ReportsPage() {
  const user = useCurrentUser()
  const [params, setParams] = useSearchParams()
  const activeId = params.get('chat')
  const { data, isLoading } = useReports()

  const setActive = (id: string | null) => {
    const next = new URLSearchParams(params)
    if (id) next.set('chat', id)
    else next.delete('chat')
    setParams(next, { replace: true })
  }

  // Guard: hanya BK.
  if (user && user.role !== 'BK') {
    return (
      <div className="flex min-h-svh items-center justify-center px-6">
        <div className="max-w-[420px] rounded-[10px] border border-violet-light bg-base p-10 text-center">
          <h1 className="text-[22px] font-bold text-violet-dark">Anda tidak punya akses ke halaman ini</h1>
          <p className="mt-2 text-[13.5px] text-[rgba(21,21,58,.6)]">Area ini hanya untuk Guru BK.</p>
        </div>
      </div>
    )
  }

  const items = data?.items ?? []
  const queue = items.filter((c) => (c.status === 'baru' || c.status === 'dibuka') && !c.pinned && !c.drafted)

  const openNext = () => {
    if (queue.length > 0) setActive(queue[0].id)
  }

  return (
    <div className="flex h-svh min-h-0 flex-col">
      {isLoading ? (
        <p className="p-8 text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</p>
      ) : activeId ? (
        <ReportPanel chatId={activeId} onClose={() => setActive(null)} onNext={openNext} />
      ) : (
        <div className="flex h-full flex-col items-center justify-center gap-3">
          <p className="text-[14px] text-[rgba(21,21,58,.45)]">
            {queue.length > 0 ? 'Pilih laporan untuk ditinjau.' : 'Semua sudah ditinjau — tidak ada laporan di antrean.'}
          </p>
          {queue.length > 0 && (
            <button
              type="button"
              onClick={openNext}
              className="rounded-[10px] bg-violet px-5 py-2.5 text-[14px] font-medium text-base hover:bg-violet-hover"
            >
              Mulai tinjau
            </button>
          )}
        </div>
      )}
    </div>
  )
}

function ReportPanel({ chatId, onClose, onNext }: { chatId: string; onClose: () => void; onNext: () => void }) {
  const { data, isLoading } = useReport(chatId)
  const confirm = useConfirm()
  const [text, setText] = useState('')
  const boxRef = useRef<HTMLDivElement>(null)

  const pin = usePinReport()
  const unpin = useUnpinReport()
  const draft = useDraftReport()
  const undraft = useUndraftReport()
  const next = useNextReport()
  const skip = useSkipReport()
  const hide = useHideReport()
  const reply = useReplyReport(chatId)

  useEffect(() => {
    if (boxRef.current) boxRef.current.scrollTop = boxRef.current.scrollHeight
  }, [data])

  if (isLoading || !data) return <p className="p-8 text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</p>

  const c = data.chat
  const messages = data.messages
  const title = c.identity.mode === 'profil' ? `${c.identity.name}${c.identity.kelas ? ` - ${c.identity.kelas}` : ''}` : 'Anonim'
  const isPinned = c.pinned && !c.drafted
  const isDraft = c.drafted
  const hasReply = messages.some((m) => m.sender === 'BK')

  const send = () => {
    const body = text.trim()
    if (!body) return
    reply.mutate(body, { onSuccess: () => setText('') })
  }

  const resolve = async () => {
    if (hasReply) {
      next.mutate(chatId, { onSuccess: onNext })
      return
    }
    const ok = await confirm({
      title: 'Lewati ke chat berikutnya?',
      body: 'Percakapan ini tetap ada di antrean, jadi bisa dibuka lagi nanti.',
      cancelText: 'Batal',
      confirmText: 'Berikutnya',
    })
    if (ok) skip.mutate(chatId, { onSuccess: onNext })
  }

  const hideChat = async () => {
    const ok = await confirm({
      title: 'Sembunyikan chat ini?',
      body: 'Chat akan hilang dari daftar kamu.',
      confirmText: 'Sembunyikan',
      danger: true,
    })
    if (ok) hide.mutate(chatId, { onSuccess: onClose })
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      {/* Header */}
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-[#c3c2e3] bg-base px-4 pb-5 pt-5 md:px-8 md:pb-6 md:pt-9">
        <div className="flex min-w-0 flex-wrap items-center gap-1">
          <h2 className="truncate font-heading text-[20px] font-semibold leading-[1.4] text-violet-dark">{title}</h2>
          <span className="text-[rgba(21,21,58,.4)]">·</span>
          <span className="shrink-0 text-[14px] font-medium leading-[1.5] text-[rgba(27,27,75,.6)]">
            {fmtDateShortID(c.last_message_at)} {new Date(c.last_message_at).getFullYear()}, {fmtHM(c.last_message_at)}
          </span>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <button
            type="button"
            onClick={() => (isPinned ? unpin.mutate(chatId) : pin.mutate(chatId))}
            className="flex items-center gap-2 rounded-[12px] border border-[#c3c2e3] bg-base px-3 py-2 hover:bg-[#F6F6FB]"
          >
            <span className="text-[16px] font-semibold leading-[1.5] text-violet-dark">
              {isPinned ? 'Lepas sematan' : 'Sematkan'}
            </span>
          </button>
          {isDraft && (
            <button
              type="button"
              onClick={() => undraft.mutate(chatId)}
              className="flex items-center gap-2 rounded-[12px] border border-[#c3c2e3] bg-base px-3 py-2 hover:bg-[#F6F6FB]"
            >
              <span className="text-[16px] font-semibold leading-[1.5] text-violet-dark">Keluarkan dari draft</span>
            </button>
          )}
          <button
            type="button"
            title="Sembunyikan"
            onClick={hideChat}
            className="flex h-[42px] items-center justify-center rounded-[12px] border border-[#c3c2e3] bg-base px-3 py-2 hover:bg-[#F6F6FB]"
          >
            ⋯
          </button>
        </div>
      </header>

      {/* Konten */}
      <div className="flex min-h-0 flex-1 flex-col items-center justify-end px-4 pb-6 md:px-6 md:pb-[100px]">
        <div className="flex w-full max-w-[958px] flex-col gap-8 md:gap-16">
          <div ref={boxRef} className="scroll-thin flex max-h-[520px] flex-col gap-8 overflow-y-auto md:gap-16">
            {messages.length === 0 ? (
              <p className="py-8 text-center text-[14px] text-[rgba(21,21,58,.45)]">Belum ada pesan.</p>
            ) : (
              messages.map((m) => <Bubble key={m.id} m={m} />)
            )}
          </div>

          <div className="flex shrink-0 flex-col gap-5">
            {/* Composer */}
            <div className="flex items-end gap-2 rounded-[16px] border border-white bg-[rgba(252,252,254,.8)] px-3 py-2">
              <textarea
                value={text}
                onChange={(e) => setText(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault()
                    send()
                  }
                }}
                rows={1}
                placeholder="Ketik pesan di sini…"
                className="max-h-32 min-h-[36px] flex-1 resize-none bg-transparent px-1 py-1.5 text-[14px] text-violet-dark outline-none placeholder:text-[rgba(21,21,58,.45)]"
              />
              <button
                type="button"
                onClick={send}
                disabled={!text.trim() || reply.isPending}
                className="flex size-9 shrink-0 items-center justify-center rounded-[10px] bg-violet-dark text-base hover:bg-[#24245c] disabled:opacity-40"
              >
                <Send className="size-4" />
              </button>
            </div>

            {/* Aksi */}
            <div className="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between">
              <button
                type="button"
                onClick={() => draft.mutate(chatId)}
                className="flex min-w-0 flex-1 items-center justify-center rounded-[12px] bg-base px-3 py-4 hover:bg-[#F6F6FB]"
              >
                <span className="text-[16px] font-semibold leading-[1.5] text-violet-dark">Simpan ke draft</span>
              </button>
              <button
                type="button"
                onClick={resolve}
                className={cn(
                  'flex min-w-0 flex-1 items-center justify-center rounded-[12px] px-3 py-4',
                  hasReply ? 'bg-violet-dark hover:bg-[#24245c]' : 'bg-[#d4d4d8] hover:bg-[#c9c9ce]',
                )}
              >
                <span className={cn('text-[16px] font-semibold leading-[1.5]', hasReply ? 'text-[#e2e2f2]' : 'text-[#71717a]')}>
                  Selesai &amp; lanjut
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function Bubble({ m }: { m: { sender: string; body: string; created_at: string; attachments: Array<{ id: string; name: string; mime: string }> } }) {
  const fromStudent = m.sender === 'SISWA'
  return (
    <div className={cn('flex w-full flex-col', fromStudent ? 'items-start' : 'items-end')}>
      <div
        className={cn(
          'flex w-[519px] max-w-full flex-col justify-center border border-white bg-[rgba(252,252,254,.8)] py-3',
          fromStudent ? 'rounded-[16px] rounded-tr-none pl-3 pr-5' : 'rounded-[16px] rounded-tl-none pl-5 pr-3',
        )}
      >
        <div className="flex flex-col gap-1 text-violet-dark">
          {m.attachments?.length > 0 && (
            <div className="flex flex-col gap-3">
              {m.attachments.map((a) => (
                <div key={a.id} className="flex w-[120px] flex-col overflow-hidden rounded-[12px]">
                  <div className="flex h-[79px] items-center justify-center border border-b-0 border-violet-light bg-base text-[12px] text-violet-dark">
                    {(a.name.split('.').pop() || 'FILE').toUpperCase().slice(0, 4)}
                  </div>
                  <div className="flex h-[27px] items-center justify-center rounded-b-[12px] bg-violet-dark px-2">
                    <span className="truncate text-[10px] leading-[1.5] text-white">{a.name.slice(0, 14)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
          <p className="whitespace-pre-wrap break-words text-[14px] font-medium leading-[1.5] opacity-70">{m.body}</p>
          <span className="self-end text-[12px] font-normal leading-[1.5] opacity-70">{fmtHM(m.created_at)}</span>
        </div>
      </div>
    </div>
  )
}
