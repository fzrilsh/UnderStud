import { useEffect, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { useConfirm } from '@/components/confirm-provider'
import {
  IconAttach,
  IconChatArrowRight,
  IconChatArrowRightMuted,
  IconChatDot,
  IconChatMore,
  IconChatPin,
  IconChatSave,
  IconChatSaveOff,
  IconChatUnpin,
  IconClose,
  IconDocAttach,
  IconSend,
} from '@/components/icons'
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
import { useSeo } from '@/lib/seo'

export default function ReportsPage() {
  useSeo('Chat', 'Antrean laporan siswa untuk ditinjau Guru BK.')

  const user = useCurrentUser()

  // Guard SEBELUM memanggil useReports() — agar non-BK tidak ikut fetch
  // `/reports` (yang akan membalas 403 dan mengotori konsol).
  if (user && user.role !== 'BK') {
    return (
      <div className="flex min-h-svh items-center justify-center px-6">
        <div className="card max-w-[420px] p-10 text-center">
          <h1 className="text-[22px] font-bold text-violet-dark">Anda tidak punya akses ke halaman ini</h1>
          <p className="mt-2 text-[13.5px] text-[rgba(21,21,58,.6)]">Area ini hanya untuk Guru BK.</p>
          <Link
            to="/"
            className="mt-6 inline-flex items-center justify-center rounded-[10px] bg-violet px-4 py-2.5 text-[14px] font-semibold text-base hover:bg-violet-hover"
          >
            Kembali ke dashboard
          </Link>
        </div>
      </div>
    )
  }
  return <ReportsInner />
}

function ReportsInner() {
  const [params, setParams] = useSearchParams()
  const activeId = params.get('chat')
  const { data, isLoading } = useReports()

  const setActive = (id: string | null) => {
    const next = new URLSearchParams(params)
    if (id) next.set('chat', id)
    else next.delete('chat')
    setParams(next, { replace: true })
  }

  const items = data?.items ?? []
  const queue = items.filter((c) => (c.status === 'baru' || c.status === 'dibuka') && !c.pinned && !c.drafted)

  const openNext = () => {
    if (queue.length > 0) setActive(queue[0].id)
  }

  // Auto-buka laporan antrean pertama saat halaman dibuka (mockup: openNextInQueue).
  const autoOpened = useRef(false)
  useEffect(() => {
    if (!autoOpened.current && !activeId && !isLoading && queue.length > 0) {
      autoOpened.current = true
      setActive(queue[0].id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, activeId, queue.length])

  return (
    <div className="bg-chat-gradient relative flex h-svh min-h-0 flex-col">
      {isLoading ? (
        <p className="p-8 text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</p>
      ) : activeId ? (
        <ReportPanel chatId={activeId} onClose={() => setActive(null)} onNext={openNext} />
      ) : (
        <div className="flex h-full flex-col items-center justify-center">
          <p className="px-4 py-12 text-center text-[14px] text-[rgba(21,21,58,.45)]">
            Semua sudah ditinjau — tidak ada laporan di antrean.
          </p>
        </div>
      )}
    </div>
  )
}

function ReportPanel({ chatId, onClose, onNext }: { chatId: string; onClose: () => void; onNext: () => void }) {
  const { data, isLoading } = useReport(chatId)
  const confirm = useConfirm()
  const [text, setText] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const fileRef = useRef<HTMLInputElement>(null)
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
    if (!body && files.length === 0) return
    reply.mutate(
      { body, files },
      {
        onSuccess: () => {
          setText('')
          setFiles([])
        },
      },
    )
  }

  // Simpan ke draft dengan dialog konfirmasi (mockup: "Simpan ke draft?").
  const saveDraft = async () => {
    const ok = await confirm({
      title: 'Simpan ke draft?',
      body: 'Chat akan dijeda dan disimpan di draft supaya bisa ditinjau nanti.',
      cancelText: 'Batal',
      confirmText: 'Simpan draft',
      closeX: true,
    })
    if (ok) draft.mutate(chatId)
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
          <IconChatDot className="shrink-0 text-violet-dark" />
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
            {isPinned ? <IconChatUnpin /> : <IconChatPin />}
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
              <IconChatSaveOff />
              <span className="text-[16px] font-semibold leading-[1.5] text-violet-dark">Keluarkan dari draft</span>
            </button>
          )}
          <button
            type="button"
            title="Opsi"
            onClick={hideChat}
            className="flex h-[42px] items-center justify-center rounded-[12px] border border-[#c3c2e3] bg-base px-3 py-2 hover:bg-[#F6F6FB]"
          >
            <IconChatMore className="text-violet-dark" />
          </button>
        </div>
      </header>

      {/* Konten. `min-h-0` pada kolom & daftar pesan wajib agar daftar yang
          lebih tinggi dari ruang tersedia benar-benar bisa di-scroll (tanpa
          ini konten meluber ke atas dan menabrak header). */}
      <div className="flex min-h-0 flex-1 flex-col items-center justify-end overflow-hidden px-4 pb-6 md:px-6 md:pb-[100px]">
        <div className="flex min-h-0 w-full max-w-[958px] flex-col gap-8 md:gap-16">
          <div
            ref={boxRef}
            className="scroll-thin flex min-h-0 max-h-[520px] flex-1 flex-col gap-8 overflow-y-auto pb-1 md:gap-16"
          >
            {messages.length === 0 ? (
              <p className="px-4 py-12 text-center text-[14px] text-[rgba(21,21,58,.45)]">Belum ada pesan.</p>
            ) : (
              messages.map((m) => <Bubble key={m.id} m={m} />)
            )}
          </div>

          <div className="flex shrink-0 flex-col gap-5">
            {/* Composer (BK juga bisa melampirkan file — mockup: tombol attach). */}
            <div className="flex w-full flex-col rounded-[16px] border border-white bg-[rgba(252,252,254,.8)] px-3 py-2">
              {files.length > 0 && (
                <div className="mb-2 flex flex-wrap gap-2">
                  {files.map((f, i) => (
                    <span key={i} className="flex items-center gap-2 rounded-xl bg-white/80 p-2 pr-3 text-[12px] text-violet-dark">
                      {f.type.startsWith('image/') ? (
                        <img src={URL.createObjectURL(f)} alt={f.name} className="size-14 shrink-0 rounded-[10px] object-cover" />
                      ) : (
                        <span className="flex size-14 shrink-0 items-center justify-center rounded-[10px] bg-[#ecebf6] text-violet">
                          <IconAttach className="size-5" />
                        </span>
                      )}
                      <span className="flex min-w-0 flex-col">
                        <span className="max-w-[160px] truncate font-medium">{f.name}</span>
                        <span className="text-[11px] text-[rgba(21,21,58,.55)]">{Math.round(f.size / 1024)} KB</span>
                      </span>
                      <button
                        type="button"
                        onClick={() => setFiles((cur) => cur.filter((_, j) => j !== i))}
                        className="flex size-7 shrink-0 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60"
                        title="Hapus lampiran"
                      >
                        <IconClose />
                      </button>
                    </span>
                  ))}
                </div>
              )}
              <input
                ref={fileRef}
                type="file"
                accept="image/*,application/pdf,text/plain"
                multiple
                className="hidden"
                onChange={(e) => setFiles(Array.from(e.target.files ?? []))}
              />
              <div className="flex items-end gap-2">
                <button
                  type="button"
                  onClick={() => fileRef.current?.click()}
                  className="mb-1.5 shrink-0 text-violet-dark hover:opacity-70"
                  title="Lampirkan file"
                >
                  <IconAttach className="size-5" />
                </button>
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
                  className="scroll-thin max-h-40 w-full flex-1 resize-none bg-transparent px-1 py-1.5 text-[14px] font-medium leading-[1.5] text-[rgba(21,21,58,.8)] outline-none placeholder:text-[rgba(21,21,58,.8)] placeholder:opacity-70"
                />
                <button
                  type="button"
                  onClick={send}
                  disabled={(!text.trim() && files.length === 0) || reply.isPending}
                  className="flex size-9 shrink-0 items-center justify-center rounded-[10px] bg-violet-dark text-base hover:bg-[#24245c] disabled:opacity-40"
                >
                  <IconSend />
                </button>
              </div>
            </div>

            {/* Aksi */}
            <div className="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between">
              <button
                type="button"
                onClick={() => (isDraft ? undraft.mutate(chatId) : saveDraft())}
                className="flex min-w-0 flex-1 flex-col items-center justify-center rounded-[12px] bg-base px-3 py-4 hover:bg-[#F6F6FB]"
              >
                <span className="flex items-center gap-2 text-[16px] font-semibold leading-[1.5] text-violet-dark">
                  <IconChatSave />
                  Simpan ke draft
                </span>
              </button>
              <button
                type="button"
                onClick={resolve}
                className={cn(
                  'flex min-w-0 flex-1 flex-col items-center justify-center rounded-[12px] px-3 py-4',
                  hasReply ? 'bg-violet-dark hover:bg-[#24245c]' : 'bg-[#d4d4d8] hover:bg-[#c9c9ce]',
                )}
              >
                <span className="flex items-center gap-2">
                  <span className={cn('text-[16px] font-semibold leading-[1.5]', hasReply ? 'text-[#e2e2f2]' : 'text-[#71717a]')}>
                    Selesai &amp; lanjut
                  </span>
                  {hasReply ? <IconChatArrowRight /> : <IconChatArrowRightMuted />}
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
              {m.attachments.map((a) =>
                a.mime?.startsWith('image/') ? (
                  <div key={a.id} className="size-[150px] overflow-hidden rounded-[12px]">
                    <img src={`/api/v1/attachments/${a.id}`} alt={a.name} className="size-full object-cover" />
                  </div>
                ) : (
                  <div key={a.id} className="flex w-[120px] flex-col overflow-hidden rounded-[12px]">
                    <div className="flex h-[79px] items-center justify-center border border-b-0 border-violet-light bg-base text-[12px] text-violet-dark">
                      {(a.name.split('.').pop() || 'FILE').toUpperCase().slice(0, 4)}
                    </div>
                    <div className="flex h-[27px] items-center justify-center gap-1 rounded-b-[12px] bg-violet-dark px-2">
                      <IconDocAttach className="shrink-0 text-white" />
                      <span className="truncate text-[10px] leading-[1.5] text-white">{a.name.slice(0, 14)}</span>
                    </div>
                  </div>
                ),
              )}
            </div>
          )}
          <p className="whitespace-pre-wrap break-words text-[14px] font-medium leading-[1.5] opacity-70">{m.body}</p>
          <span className="self-end text-[12px] font-normal leading-[1.5] opacity-70">{fmtHM(m.created_at)}</span>
        </div>
      </div>
    </div>
  )
}
