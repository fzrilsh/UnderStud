import { Sparkles } from 'lucide-react'

import type { SummaryView } from '@/features/dashboard/hooks'
import type { StudentDetail } from '@/features/students/detail-hooks'
import { fmtDateID, fmtDateTimeID } from '@/lib/format'
import { cn } from '@/lib/utils'

/** Konten ringkasan terstruktur (dibangun backend). */
type Content = {
  headline: string
  narrative: string
  metrics: Array<{ label: string; value: string }>
  insights: Array<{
    icon: string
    title: string
    subtitle: string
    items: Array<{ text: string }>
    sources: string[]
    muted: boolean
  }>
  data_used: Array<{ source: string; details: string; updated: string; status: string }>
  red_flag: { title: string; note: string } | null
}

function contentOf(sm: SummaryView): Content {
  return (sm.content ?? {}) as unknown as Content
}

import { IconFavourite, IconIdea, IconInfoCircle, IconStar } from '@/components/icons'

function InsightIcon({ icon, className }: { icon: string; className?: string }) {
  if (icon === 'star') return <IconStar className={className} />
  if (icon === 'favourite') return <IconFavourite className={className} />
  if (icon === 'infoCircle') return <IconInfoCircle className={className} />
  return <IconIdea className={className} />
}

/** Tab "Ringkasan AI" — hero, banner red flag, temuan, data dipakai, riwayat versi. */
export default function SummaryTab({
  detail,
  onGenerate,
  onDownload,
  onViewVersion,
  generating,
}: {
  detail: StudentDetail
  onGenerate: () => void
  onDownload: (summaryId: string) => void
  onViewVersion: (sm: SummaryView) => void
  generating: boolean
}) {
  const sm = detail.latest_summary
  const summaries = detail.summaries ?? []

  if (!sm) {
    return (
      <div className="rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
        <h2 className="font-heading text-[24px] font-semibold leading-[1.4] text-violet-dark">Ringkasan AI</h2>
        <p className="mt-1 text-[12px] text-[rgba(21,21,58,.65)]">
          Belum ada ringkasan untuk siswa ini. Buat ringkasan dari data siswa, observasi, dan nilai E-Rapor.
        </p>
        <button
          type="button"
          onClick={onGenerate}
          disabled={generating}
          className="mt-4 flex items-center gap-1.5 rounded-[8px] bg-violet px-3.5 py-2.5 text-[14px] font-medium text-base hover:bg-violet-hover disabled:opacity-50"
        >
          <Sparkles className="size-4" /> {generating ? 'Membuat…' : 'Buat ringkasan'}
        </button>
      </div>
    )
  }

  const c = contentOf(sm)
  const ins = c.insights ?? []
  const dataUsed = c.data_used ?? []
  const red = c.red_flag
  const usedCount = dataUsed.filter((r) => r.status === 'used').length

  return (
    <div className="flex flex-col gap-4">
      {/* Hero */}
      <section className="relative overflow-hidden rounded-[10px] bg-[#1b1b4b] p-8 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
        <div className="pointer-events-none absolute -top-[380px] right-0 size-[480px] translate-x-1/3 rounded-full bg-violet opacity-70 blur-[120px]" />
        <div className="pointer-events-none absolute bottom-0 right-[300px] size-[320px] translate-y-1/2 rounded-full bg-violet opacity-60 blur-[110px]" />
        <div className="relative">
          <div className="flex items-center gap-3">
            <span className="flex size-10 shrink-0 items-center justify-center rounded-[8px] bg-[rgba(252,252,254,0.1)] text-base">
              <Sparkles className="size-5" />
            </span>
            <div className="min-w-0 flex-1">
              <p className="text-[14px] font-semibold text-base">Ringkasan profil</p>
              <p className="mt-1 text-[12px] text-base opacity-70">
                Dibuat oleh {sm.author} · {fmtDateID(sm.created_at)} · {sm.sources_used} dari {sm.sources_total} sumber data dipakai
              </p>
            </div>
            <span className="shrink-0 rounded-full bg-[rgba(252,252,254,0.1)] px-3 py-1.5 text-[12px] font-medium text-base">
              Versi {sm.version} · Terbaru
            </span>
          </div>
          <p className="mt-6 max-w-[760px] font-heading text-[28px] font-medium leading-[1.3] text-base">{c.headline}</p>
          <p className="mt-4 max-w-[820px] whitespace-pre-wrap text-[14px] leading-[1.6] text-base opacity-80">{c.narrative}</p>
          <div className="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4">
            {(c.metrics ?? []).map((m) => (
              <div key={m.label} className="rounded-[8px] bg-[rgba(252,252,254,0.1)] px-4 py-[14px]">
                <p className="font-heading text-[24px] font-medium leading-normal text-base">{m.value}</p>
                <p className="mt-1 text-[12px] leading-normal text-base opacity-70">{m.label}</p>
              </div>
            ))}
          </div>
          <div className="mt-6 flex items-center gap-2.5 rounded-[8px] border border-[rgba(252,252,254,0.16)] bg-[rgba(252,252,254,0.05)] px-3.5 py-3">
            <p className="text-[12px] font-medium leading-normal text-base opacity-85">
              Ringkasan ini membantu penilaian Anda, bukan diagnosis atau keputusan otomatis. Data yang kosong tidak pernah
              dianggap sebagai fakta.
            </p>
          </div>
        </div>
      </section>

      {/* Banner red flag */}
      {red && (
        <div className="flex flex-wrap items-center gap-4 rounded-[10px] border border-[#ef4444] bg-base px-6 py-5 shadow-[0px_0px_1px_0px_rgba(24,24,27,0.2),0px_1px_2px_0px_rgba(24,24,27,0.1)]">
          <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-[#ef4444] text-[18px] font-bold text-base">
            !
          </span>
          <div className="min-w-0 flex-1">
            <p className="text-[14px] font-semibold text-violet-dark">{red.title}</p>
            <p className="mt-2 text-[12px] text-[#a52d18]">{red.note}</p>
          </div>
        </div>
      )}

      {/* Temuan */}
      <section className="mt-4">
        <div className="mb-4">
          <h2 className="font-heading text-[24px] font-medium text-violet-dark">Temuan</h2>
          <p className="mt-1 text-[12px] text-violet-dark opacity-60">
            Dikelompokkan per tema. Setiap poin menunjukkan sumber datanya.
          </p>
        </div>
        <div className="rounded-[10px] border border-violet-light bg-base px-6 py-1 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
          {ins.map((g, gi) => (
            <div
              key={g.title + gi}
              className={cn('flex flex-col gap-4 py-5 lg:flex-row lg:gap-8', gi !== ins.length - 1 && 'border-b border-violet-light')}
            >
              <div className="flex w-full shrink-0 items-center gap-3 lg:w-[280px]">
                <span className={cn('flex size-9 shrink-0 items-center justify-center rounded-[8px] text-[16px]', g.muted ? 'bg-violet-light' : 'bg-[#ecebf6]')}>
                  <InsightIcon icon={g.icon} className="size-5 text-violet-dark" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-[16px] font-semibold text-violet-dark">{g.title}</p>
                  <p className="text-[12px] text-violet-dark opacity-60">{g.subtitle}</p>
                </div>
              </div>
              <div className="flex min-w-0 flex-1 flex-col gap-[10px]">
                {(g.items ?? []).map((it, ii) => (
                  <div key={ii} className="flex items-start gap-[10px]">
                    <span className={cn('mt-[7px] size-1.5 shrink-0 rounded-full', g.muted ? 'bg-[#a1a1aa]' : 'bg-violet')} />
                    <p className={cn('text-[14px] leading-[1.5] text-violet-dark', g.muted && 'opacity-75')}>{it.text}</p>
                  </div>
                ))}
                {g.sources?.length > 0 && (
                  <div className="flex flex-wrap items-center gap-1.5 pt-1">
                    <span className="text-[12px] text-violet-dark opacity-60">Sumber</span>
                    {g.sources.map((s) => (
                      <span
                        key={s}
                        className={cn('rounded-[6px] px-2 py-[3px] text-[12px] font-medium', g.muted ? 'bg-violet-light text-violet-dark' : 'bg-[#ecebf6] text-violet')}
                      >
                        {s}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Data yang dipakai */}
      <section className="mt-4">
        <div className="mb-4">
          <h2 className="font-heading text-[24px] font-medium text-violet-dark">Data yang dipakai</h2>
          <p className="mt-1 text-[12px] text-violet-dark opacity-60">
            Data diambil {fmtDateTimeID(sm.created_at)}. Ubah data di halaman siswa, lalu buat ulang ringkasan.
          </p>
        </div>
        <div className="rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
          <div className="overflow-x-auto rounded-[8px]">
            <div className="min-w-[680px]">
              <div className="flex items-center gap-3 bg-[#ecebf6] px-3 py-2.5">
                <p className="min-w-0 flex-1 text-[12px] font-semibold text-violet-dark">Sumber</p>
                <p className="w-[320px] shrink-0 text-[12px] font-semibold text-violet-dark">Detail</p>
                <p className="w-[140px] shrink-0 text-[12px] font-semibold text-violet-dark">Diperbarui</p>
                <p className="w-[110px] shrink-0 text-[12px] font-semibold text-violet-dark">Status</p>
              </div>
              {dataUsed.map((r) => (
                <div key={r.source} className="flex items-center gap-3 border-b border-violet-light p-3 last:border-b-0">
                  <p className="min-w-0 flex-1 text-[14px] font-medium text-violet-dark">{r.source}</p>
                  <p className="w-[320px] shrink-0 text-[12px] text-violet-dark opacity-60">{r.details}</p>
                  <p className="w-[140px] shrink-0 text-[12px] text-violet-dark opacity-60">{r.updated}</p>
                  <div className="w-[110px] shrink-0">
                    <span
                      className={cn(
                        'rounded-[6px] px-2 py-[3px] text-[12px] font-medium',
                        r.status === 'used' ? 'bg-[#ecebf6] text-violet' : 'bg-violet-light text-violet-dark',
                      )}
                    >
                      {r.status === 'used' ? 'Dipakai' : r.status === 'not_sent' ? 'Tidak dikirim' : 'Kosong'}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
            <p className="text-[12px] text-violet-dark opacity-60">
              {usedCount} dari {dataUsed.length} sumber dipakai · Kolom kosong dilewati, tidak pernah ditebak
            </p>
          </div>
        </div>
      </section>

      {/* Riwayat versi */}
      <section className="mt-4">
        <div className="mb-4">
          <h2 className="font-heading text-[24px] font-medium text-violet-dark">Riwayat versi</h2>
          <p className="mt-1 text-[12px] text-violet-dark opacity-60">BK hanya bisa mengunduh versi terbaru.</p>
        </div>
        <div className="rounded-[10px] border border-violet-light bg-base px-6 py-1 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
          {summaries.map((v, i) => {
            const current = i === 0
            return (
              <div key={v.id} className={cn('flex flex-wrap items-center gap-4 py-4', i !== summaries.length - 1 && 'border-b border-violet-light')}>
                <div className="w-full min-w-0 sm:w-[360px] sm:shrink-0">
                  <div className="flex items-center gap-2">
                    <p className="text-[14px] font-semibold text-violet-dark">Versi {v.version}</p>
                    {current && (
                      <span className="rounded-[6px] bg-[#ecebf6] px-2 py-[3px] text-[12px] font-medium text-violet">Terbaru</span>
                    )}
                  </div>
                  <p className="mt-0.5 text-[12px] text-violet-dark opacity-60">
                    {fmtDateID(v.created_at)} · oleh {v.author}
                  </p>
                </div>
                <p className="min-w-0 flex-1 text-[12px] text-violet-dark opacity-60">{v.note}</p>
                {current ? (
                  <button
                    type="button"
                    onClick={() => onDownload(v.id)}
                    className="shrink-0 rounded-[8px] bg-[#ecebf6] px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-violet-light"
                  >
                    Unduh
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => onViewVersion(v)}
                    className="shrink-0 rounded-[8px] border border-violet-light bg-base px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-[#F6F6FB]"
                  >
                    Lihat
                  </button>
                )}
              </div>
            )
          })}
        </div>
      </section>
    </div>
  )
}
