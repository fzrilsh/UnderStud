import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { Trash2 } from 'lucide-react'

import { useConfirm } from '@/components/confirm-provider'
import { IconAttach, IconPencil, IconSend, IconUserLight } from '@/components/icons'
import {
  getGeolocation,
  useCreateLapor,
  useHideLapor,
  useLaporBootstrap,
  useLaporChat,
  useLaporChats,
  useReplyLapor,
  useSaveAnonProfile,
} from '@/features/lapor/hooks'
import { fmtHM } from '@/lib/format'
import { cn } from '@/lib/utils'
import { useSeo } from '@/lib/seo'

export default function LaporPage() {
  useSeo('Chat Anonim', 'Kirim laporan anonim tanpa akun. Guru BK akan meninjau laporanmu.')

  const bootstrap = useLaporBootstrap()
  const chats = useLaporChats()
  const [params, setParams] = useSearchParams()
  const activeId = params.get('chat')

  const setActive = (id: string | null) => {
    const next = new URLSearchParams(params)
    if (id) next.set('chat', id)
    else next.delete('chat')
    setParams(next, { replace: true })
  }

  const list = chats.data ?? []
  const profileName = bootstrap.data?.profile?.name ?? ''

  return (
    <div className="flex h-svh min-h-0 bg-base">
      {/* Sidebar siswa */}
      <aside className="bg-sidebar-gradient flex w-[240px] shrink-0 flex-col rounded-tr-[20px] border-r border-t border-violet-light px-[14px] pt-6">
        <div className="flex h-[68px] items-center gap-2.5 px-1">
          <img src="/understud-logo.png" alt="Understud" className="size-9 shrink-0 object-contain" />
          <span className="flex-1 truncate text-[16px] font-semibold text-violet-dark">Understud</span>
        </div>

        <button
          type="button"
          onClick={() => setActive(null)}
          className="mt-3 flex w-full items-center gap-2.5 rounded-[8px] px-3 py-2 text-[14px] font-medium text-[rgba(21,21,58,.8)] hover:bg-violet-light/60"
        >
          <span className="text-violet-dark">
            <IconPencil />
          </span>{' '}
          Chat baru
        </button>

        <div className="scroll-thin mt-5 flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto">
          <p className="mb-1.5 text-[12px] font-semibold text-[rgba(21,21,58,.45)]">Terbaru</p>
          {list.length === 0 ? (
            <p className="px-2 py-3 text-[12.5px] text-[rgba(21,21,58,.45)]">Belum ada laporan.</p>
          ) : (
            list.map((c) => <RecentItem key={c.id} c={c} active={c.id === activeId} onOpen={setActive} />)
          )}
        </div>
      </aside>

      {/* Main */}
      <main className="bg-chat-gradient relative flex min-w-0 flex-1 flex-col">
        {activeId ? (
          <ConversationView chatId={activeId} />
        ) : (
          <NewChatView profileName={profileName} />
        )}
      </main>
    </div>
  )
}

function RecentItem({
  c,
  active,
  onOpen,
}: {
  c: { id: string; last_message: { body: string } | null; unread_count: number }
  active: boolean
  onOpen: (id: string) => void
}) {
  const confirm = useConfirm()
  const hide = useHideLapor()

  const remove = async () => {
    const ok = await confirm({
      title: 'Hapus chat ini?',
      body: 'Chat ini akan dihapus dari daftar dan tidak bisa dilanjutkan.',
      confirmText: 'Hapus chat',
      danger: true,
    })
    if (ok) hide.mutate(c.id, { onSuccess: () => onOpen('') })
  }

  return (
    <div className="group relative">
      <button
        type="button"
        onClick={() => onOpen(c.id)}
        className={cn(
          'flex w-full items-center justify-between gap-2 rounded-[8px] py-1.5 pl-3 pr-8 text-left',
          active ? 'bg-[rgba(195,194,227,.6)]' : 'hover:bg-[#F6F6FB]',
        )}
      >
        <span className="truncate text-[14px] font-medium leading-[1.5] text-[rgba(21,21,58,.8)]">
          {c.last_message?.body || 'Laporan baru'}
        </span>
        {c.unread_count > 0 && (
          <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-violet text-[11px] font-semibold text-base">
            {c.unread_count}
          </span>
        )}
      </button>
      <button
        type="button"
        title="Hapus chat"
        onClick={remove}
        className="absolute right-1.5 top-1/2 flex size-6 -translate-y-1/2 items-center justify-center rounded-md text-violet-dark opacity-0 transition-opacity hover:bg-violet-light/60 group-hover:opacity-100"
      >
        <Trash2 className="size-3.5" />
      </button>
    </div>
  )
}

/** Composer bersama (kirim laporan baru / balasan). */
function Composer({
  placeholder,
  sending,
  onSubmit,
}: {
  placeholder: string
  sending: boolean
  onSubmit: (body: string, files: File[]) => void
}) {
  const [text, setText] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const inputRef = useRef<HTMLInputElement>(null)

  const submit = (e: { preventDefault: () => void }) => {
    e.preventDefault()
    if (!text.trim() && files.length === 0) return
    onSubmit(text.trim(), files)
    setText('')
    setFiles([])
  }

  return (
    <form onSubmit={submit} className="flex w-full flex-col rounded-[16px] border border-white bg-[rgba(252,252,254,.8)] px-3 py-2">
      {files.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-2">
          {files.map((f, i) => (
            <span key={i} className="flex items-center gap-2 rounded-xl bg-white/80 px-2 py-1 text-[12px] text-violet-dark">
              {f.name}
              <button type="button" onClick={() => setFiles((cur) => cur.filter((_, j) => j !== i))} className="text-violet-dark">
                ×
              </button>
            </span>
          ))}
        </div>
      )}
      <input
        ref={inputRef}
        type="file"
        accept="image/*,application/pdf,text/plain"
        multiple
        className="hidden"
        onChange={(e) => setFiles(Array.from(e.target.files ?? []))}
      />
      <div className="flex w-full items-center justify-between gap-3">
        <div className="flex min-w-0 flex-1 items-center gap-4">
          <button type="button" onClick={() => inputRef.current?.click()} className="shrink-0 text-violet-dark hover:opacity-70">
            <IconAttach className="size-[20px]" />
          </button>
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                submit(e)
              }
            }}
            rows={1}
            placeholder={placeholder}
            className="max-h-40 w-full resize-none bg-transparent text-[14px] font-medium leading-[1.5] text-[rgba(21,21,58,.8)] outline-none placeholder:text-[rgba(21,21,58,.8)] placeholder:opacity-70"
          />
        </div>
        <button
          type="submit"
          disabled={sending}
          className="flex size-9 shrink-0 items-center justify-center rounded-[10px] bg-violet-dark text-base hover:bg-[#24245c] disabled:opacity-40"
        >
          <IconSend className="size-4" />
        </button>
      </div>
    </form>
  )
}

/** Tampilan awal: judul + composer + toggle identitas. */
function NewChatView({ profileName }: { profileName: string }) {
  const [mode, setMode] = useState<'anonim' | 'profil'>('anonim')
  const [profileOpen, setProfileOpen] = useState(false)
  const create = useCreateLapor()

  const submit = async (body: string, files: File[]) => {
    if (mode === 'profil' && !profileName) {
      toast.error('Lengkapi nama dulu, atau pilih mode Anonim.')
      return
    }
    const geo = await getGeolocation()
    create.mutate(
      { body, identityMode: mode, files, ...geo },
      {
        onSuccess: () => toast.success('Laporan terkirim.'),
      },
    )
  }

  const pickProfile = async () => {
    if (!profileName) {
      setProfileOpen(true)
      return
    }
    setMode('profil')
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col items-center justify-center overflow-y-auto px-4 pb-10 md:px-6">
      <div className="flex w-full max-w-[780px] flex-col items-center gap-12 md:gap-16">
        <h1 className="text-gradient-heading w-full text-center font-heading text-[32px] font-medium leading-[1.2] md:text-[48px]">
          Apa yang ingin kamu ceritakan?
        </h1>
        <div className="flex w-full flex-col items-start gap-5">
          <Composer placeholder="Ketik pesan di sini..." sending={create.isPending} onSubmit={submit} />
          <div className="flex items-center gap-4 rounded-[16px] border border-white bg-[rgba(252,252,254,.6)] px-3 py-2">
            <button
              type="button"
              onClick={() => setMode('anonim')}
              className={cn('flex shrink-0 items-center gap-2 rounded-[12px] px-3 py-2.5', mode === 'anonim' && 'bg-violet-dark')}
            >
              <span className={cn('text-[14px] font-medium leading-[1.5]', mode === 'anonim' ? 'text-violet-light' : 'text-violet-dark')}>
                Anonim
              </span>
            </button>
            <button
              type="button"
              onClick={pickProfile}
              className={cn('flex shrink-0 items-center gap-2 rounded-[12px] px-3 py-2.5', mode === 'profil' && 'bg-violet-dark')}
            >
              <span className={cn('text-[14px] font-medium leading-[1.5]', mode === 'profil' ? 'text-violet-light' : 'text-violet-dark')}>
                Kirim sebagai {profileName || 'Profil'}
              </span>
            </button>
          </div>
        </div>
      </div>
      {profileOpen && (
        <ProfileModal
          initialName={profileName}
          onClose={(saved) => {
            setProfileOpen(false)
            if (saved) setMode('profil')
          }}
        />
      )}
    </div>
  )
}

/** Modal identitas pelapor (Nama + Kelas). */
function ProfileModal({
  initialName,
  onClose,
}: {
  initialName: string
  onClose: (saved: boolean) => void
}) {
  const [name, setName] = useState(initialName)
  const [kelas, setKelas] = useState('')
  const save = useSaveAnonProfile()

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-[rgba(21,21,58,.45)] p-4">
      <div className="relative w-full max-w-[440px] rounded-2xl bg-white p-6 shadow-xl">
        <button
          type="button"
          onClick={() => onClose(false)}
          className="absolute right-4 top-4 flex size-7 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60"
        >
          ×
        </button>
        <h3 className="text-gradient-heading font-heading text-[22px] font-semibold leading-tight">Identitas Pelapor</h3>
        <p className="mt-1.5 text-[13px] text-[rgba(21,21,58,.6)]">
          Nama dan kelas dipakai hanya untuk mode "Kirim sebagai". Laporan tetap bisa dikirim secara anonim.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            save.mutate(
              { name: name.trim(), kelas: kelas.trim() },
              { onSuccess: () => onClose(true) },
            )
          }}
          className="mt-6 flex flex-col gap-4"
        >
          <div>
            <label className="mb-1.5 block text-[13px] font-semibold text-violet-dark">Nama</label>
            <input
              value={name}
              maxLength={80}
              onChange={(e) => setName(e.target.value)}
              placeholder="Nama lengkap kamu"
              className="w-full rounded-[10px] border border-violet-light px-3 py-2.5 text-[14px] outline-none focus:border-violet"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-[13px] font-semibold text-violet-dark">Kelas</label>
            <input
              value={kelas}
              maxLength={40}
              onChange={(e) => setKelas(e.target.value)}
              placeholder="Contoh: X TJKT 1"
              className="w-full rounded-[10px] border border-violet-light px-3 py-2.5 text-[14px] outline-none focus:border-violet"
            />
          </div>
          <div className="mt-2 flex gap-3">
            <button
              type="button"
              onClick={() => onClose(false)}
              className="flex-1 rounded-[10px] border border-violet-light bg-white px-3 py-2.5 text-[14px] font-medium text-violet-dark hover:bg-[#F6F6FB]"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={save.isPending}
              className="flex-1 rounded-[10px] bg-violet-dark px-3 py-2.5 text-[14px] font-medium text-base hover:bg-[#24245c] disabled:opacity-50"
            >
              Simpan
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

/** Tampilan percakapan (buka laporan). */
function ConversationView({ chatId }: { chatId: string }) {
  const { data, isLoading } = useLaporChat(chatId)
  const reply = useReplyLapor(chatId)
  const boxRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (boxRef.current) boxRef.current.scrollTop = boxRef.current.scrollHeight
  }, [data])

  if (isLoading || !data) return <p className="p-8 text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</p>

  return (
    <div className="flex h-full min-h-0 flex-col items-center justify-end px-4 pb-6 md:px-8 md:pb-10">
      <div className="flex h-full w-full max-w-[956px] flex-col items-center justify-between gap-8 pt-8">
        <h2 className="shrink-0 text-center font-heading text-[20px] font-semibold leading-[1.4] text-violet-dark">Chat Anonim</h2>
        <div className="flex w-full flex-col items-center gap-10 md:gap-16">
          <div ref={boxRef} className="scroll-thin flex max-h-[52vh] w-full flex-col gap-10 overflow-y-auto md:gap-16">
            {data.messages.map((m) => (
              <Bubble key={m.id} m={m} />
            ))}
          </div>
          <div className="w-full max-w-[780px] shrink-0">
            <Composer
              placeholder="Ketik pesan di sini..."
              sending={reply.isPending}
              onSubmit={(body, files) => reply.mutate({ body, files })}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

function Bubble({ m }: { m: { sender: string; body: string; created_at: string } }) {
  const own = m.sender === 'SISWA' // pesan sendiri → kanan
  return (
    <div className={cn('flex w-full flex-col', own ? 'items-end' : 'items-start')}>
      <div
        className={cn(
          'flex w-[519px] max-w-full flex-col justify-center border border-white bg-[rgba(252,252,254,.8)] py-3',
          own ? 'rounded-[16px] rounded-tr-none pl-5 pr-3' : 'rounded-[16px] rounded-tl-none pl-3 pr-5',
        )}
      >
        <div className="flex flex-col gap-1 text-violet-dark">
          <p className="whitespace-pre-wrap break-words text-[14px] font-medium leading-[1.5] opacity-70">{m.body}</p>
          <span className="self-end text-[12px] font-normal leading-[1.5] opacity-70">{fmtHM(m.created_at)}</span>
        </div>
      </div>
    </div>
  )
}

/** Avatar mengambang di kanan-atas (identitas pelapor). */
export function LaporAvatar() {
  return (
    <span
      className="flex size-12 items-center justify-center rounded-full p-[5px]"
      style={{ backgroundImage: 'linear-gradient(90deg,#3d3ba6 0%,#6462bd 40%,#e9ebf8 100%)' }}
    >
      <span className="flex size-full items-center justify-center rounded-full">
        <IconUserLight />
      </span>
    </span>
  )
}
