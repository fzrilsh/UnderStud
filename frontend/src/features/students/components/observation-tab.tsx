import { useMemo, useState } from 'react'

import { cn } from '@/lib/utils'
import { fmtDateID } from '@/lib/format'
import type { ChecklistItem, ObservationView } from '@/features/students/detail-hooks'

const KATEGORI_LABEL: Record<string, string> = {
  emosi: 'Emosi',
  sosial: 'Sosial',
  akademik: 'Akademik & motivasi',
  kehadiran: 'Kehadiran & disiplin',
  fisik: 'Fisik & kebiasaan',
  keluarga: 'Situasi keluarga',
  red_flag: 'Red flag',
}

const NORMAL_CATS = ['emosi', 'sosial', 'akademik', 'kehadiran', 'fisik', 'keluarga']

const TINDAKAN = [
  { value: 'pantau', title: 'Pantau', sub: 'Terus amati di kelas', red: false },
  { value: 'konsultasi_bk', title: 'Konsultasi ke BK', sub: 'Di pertemuan BK berikutnya', red: false },
  { value: 'eskalasi_segera', title: 'Tandai sekarang', sub: 'BK langsung diberi tahu', red: true },
] as const

/** Kartu Indikator (checklist 28 item) + Detail observasi + riwayat. */
export default function ObservationTab({
  observations,
  checklist,
  canWrite,
  submitting,
  onCreate,
  onHandle,
  onEdit,
}: {
  observations: ObservationView[]
  checklist: ChecklistItem[]
  canWrite: boolean
  submitting: boolean
  onCreate: (payload: { items: string[]; catatan?: string; tanggal: string; tindakan_awal?: string }) => void
  onHandle: (observationId: string) => void
  onEdit: (o: ObservationView) => void
}) {
  const [selected, setSelected] = useState<string[]>([])
  const [tidakAda, setTidakAda] = useState(false)
  const [catatan, setCatatan] = useState('')
  const [tindakan, setTindakan] = useState<string>('pantau')
  const [historyFilter, setHistoryFilter] = useState<'all' | 'red_flag'>('all')

  const byCat = useMemo(() => {
    const m = new Map<string, ChecklistItem[]>()
    for (const it of checklist) {
      if (!m.has(it.kategori)) m.set(it.kategori, [])
      m.get(it.kategori)!.push(it)
    }
    return m
  }, [checklist])

  const toggle = (key: string) => {
    setTidakAda(false)
    setSelected((cur) => (cur.includes(key) ? cur.filter((k) => k !== key) : [...cur, key]))
  }

  const reset = () => {
    setSelected([])
    setTidakAda(false)
    setCatatan('')
    setTindakan('pantau')
  }

  const submit = () => {
    const items = tidakAda ? ['tidak_ada'] : selected
    if (items.length === 0) return
    onCreate({
      items,
      catatan: catatan.trim() || undefined,
      tanggal: new Date().toISOString().slice(0, 10),
      tindakan_awal: tindakan,
    })
    reset()
  }

  const redItems = byCat.get('red_flag') ?? []
  const selectedRed = redItems.filter((it) => selected.includes(it.key)).length
  const sorted = [...observations].sort((a, b) => new Date(b.tanggal).getTime() - new Date(a.tanggal).getTime())
  const redCount = sorted.filter((o) => o.red_flag).length
  const list = historyFilter === 'red_flag' ? sorted.filter((o) => o.red_flag) : sorted

  const card = 'rounded-[10px] border border-violet-light bg-base p-6 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]'

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-[20px] font-bold text-violet-dark">Observasi</h2>
        <p className="mt-0.5 text-[12.5px] text-[rgba(21,21,58,.55)]">
          Menyimpan menambah entri baru ke riwayat.
        </p>
      </div>

      {canWrite && (
        <div className={card}>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 className="font-heading text-[20px] font-semibold text-violet-dark">Indikator</h3>
              <p className="mt-0.5 text-[12px] text-[rgba(21,21,58,.6)]">
                Centang hanya yang Anda lihat langsung · <b>{tidakAda ? 0 : selected.length}</b> dari {checklist.length} dipilih ·
                Bukan diagnosis, tanpa skor risiko
              </p>
            </div>
            <label className="flex items-center gap-2 rounded-xl border border-violet-light px-3 py-1.5">
              <input
                type="checkbox"
                checked={tidakAda}
                onChange={(e) => {
                  setTidakAda(e.target.checked)
                  if (e.target.checked) setSelected([])
                }}
              />
              <span className="text-[13px] font-medium text-violet-dark">Tidak ada indikator yang terlihat</span>
            </label>
          </div>

          <div className="mt-4">
            {NORMAL_CATS.map((cat) => {
              const items = byCat.get(cat) ?? []
              const picked = items.filter((it) => selected.includes(it.key)).length
              return (
                <div key={cat} className="grid grid-cols-1 items-start gap-2 border-t border-[#F1F1F5] py-3 md:grid-cols-[190px_1fr] md:gap-4">
                  <div className="flex items-center gap-2">
                    <span className="text-[13.5px] font-semibold text-violet-dark">{KATEGORI_LABEL[cat] ?? cat}</span>
                    {picked > 0 && (
                      <span className="rounded-[6px] bg-[#ecebf6] px-2 py-[2px] text-[12px] font-semibold text-violet">{picked}</span>
                    )}
                  </div>
                  <div className="grid grid-cols-1 gap-x-4 gap-y-1.5 sm:grid-cols-2 lg:grid-cols-4">
                    {items.map((it) => (
                      <label key={it.key} className="flex cursor-pointer items-start gap-2.5 rounded-[8px] px-2 py-1.5 hover:bg-[#F6F6FB]">
                        <input
                          type="checkbox"
                          checked={selected.includes(it.key)}
                          disabled={tidakAda}
                          onChange={() => toggle(it.key)}
                          className="mt-0.5"
                        />
                        <span className="text-[13px] leading-[1.4] text-violet-dark">{it.label}</span>
                      </label>
                    ))}
                  </div>
                </div>
              )
            })}
          </div>

          {/* Panel Red flag */}
          <div className="mt-4 rounded-[8px] border border-[#ef4444] bg-[rgba(239,68,68,0.02)] p-4">
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-[224px_1fr]">
              <div className="flex flex-col gap-2">
                <div className="flex w-full items-center gap-2">
                  <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-[#ef4444] text-[12px] font-semibold text-base">
                    !
                  </span>
                  <span className="min-w-0 flex-1 text-[14px] font-semibold text-[#ef4444]">Red flag</span>
                  {selectedRed > 0 && (
                    <span className="shrink-0 rounded-[6px] bg-[#ef4444] px-2 py-[3px] text-[12px] font-semibold text-base">
                      {selectedRed}
                    </span>
                  )}
                </div>
                <p className="text-[12px] leading-snug text-[#ef4444]">
                  Perlu pemberitahuan segera ke BK. Red flag tidak menunggu skor.
                </p>
              </div>
              <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
                {redItems.map((it) => (
                  <label key={it.key} className="flex cursor-pointer items-start gap-2.5 rounded-[8px] px-2 py-1.5 hover:bg-[#fef2f2]">
                    <input
                      type="checkbox"
                      checked={selected.includes(it.key)}
                      disabled={tidakAda}
                      onChange={() => toggle(it.key)}
                      className="mt-0.5 accent-[#ef4444]"
                    />
                    <span className="text-[13px] leading-[1.4] text-[#ef4444]">{it.label}</span>
                  </label>
                ))}
              </div>
            </div>
          </div>

          {/* Detail observasi */}
          <div className="mt-6">
            <h3 className="mb-3 text-[15px] font-bold text-violet-dark">Detail observasi</h3>
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1.5 block text-[13px] font-semibold text-violet-dark">Tanggal observasi</label>
                <div className="flex max-w-[240px] items-center gap-2 rounded-[10px] border border-violet-light px-3 py-2.5 text-[14px] text-violet-dark">
                  {fmtDateID(new Date().toISOString())}
                </div>
                <p className="mt-1 text-[11.5px] text-[rgba(21,21,58,.45)]">Otomatis hari ini.</p>
              </div>
              <div>
                <div className="mb-1.5 flex items-baseline justify-between gap-2">
                  <label className="text-[13px] font-semibold text-violet-dark">Catatan observasi</label>
                  <span className="text-[11.5px] text-[rgba(21,21,58,.45)]">{catatan.length} / 1000</span>
                </div>
                <textarea
                  value={catatan}
                  maxLength={1000}
                  rows={3}
                  onChange={(e) => setCatatan(e.target.value)}
                  placeholder="Tulis di sini…"
                  className="w-full rounded-[10px] border border-violet-light bg-white px-3 py-2.5 text-[14px] text-violet-dark outline-none focus:border-violet"
                />
                <p className="mt-1 text-[11.5px] text-[rgba(21,21,58,.45)]">Opsional · Tulis yang Anda lihat, bukan kesimpulan.</p>
              </div>
              <div>
                <div className="mb-1.5 flex items-baseline justify-between gap-2">
                  <label className="text-[13px] font-semibold text-violet-dark">Tindakan awal</label>
                  <span className="text-[11.5px] text-[rgba(21,21,58,.45)]">Opsional</span>
                </div>
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                  {TINDAKAN.map((t) => {
                    const sel = tindakan === t.value
                    return (
                      <label
                        key={t.value}
                        className={cn(
                          'flex cursor-pointer items-center gap-[10px] rounded-[8px] border px-3 py-2.5 transition',
                          t.red ? 'border-[#ef4444] bg-[#fff0eb]' : 'border-violet-light bg-base hover:bg-[#F6F6FB]',
                        )}
                      >
                        <input
                          type="radio"
                          name="tindakan"
                          value={t.value}
                          checked={sel}
                          onChange={() => setTindakan(t.value)}
                          className="size-4 accent-[#3d3ba6]"
                        />
                        <span className={cn('text-[14px] font-medium', t.red ? 'text-[#ef4444]' : 'text-violet-dark')}>
                          {t.title}
                        </span>
                        <span className={cn('ml-auto text-right text-[12px]', t.red ? 'text-[#ef4444]' : 'text-[rgba(21,21,58,.6)]')}>
                          {t.sub}
                        </span>
                      </label>
                    )
                  })}
                </div>
              </div>
            </div>
            <button
              type="button"
              onClick={submit}
              disabled={submitting || (!tidakAda && selected.length === 0)}
              className="mt-4 rounded-[8px] bg-violet-dark px-3.5 py-2.5 text-[14px] font-medium text-base hover:bg-[#24245c] disabled:opacity-40"
            >
              {submitting ? 'Menyimpan…' : 'Simpan observasi'}
            </button>
          </div>
        </div>
      )}

      {/* Riwayat observasi */}
      <div className={card}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h3 className="font-heading text-[20px] font-semibold text-violet-dark">Riwayat observasi</h3>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setHistoryFilter('all')}
              className={cn(
                'rounded-full px-3 py-1.5 text-[13px] font-medium',
                historyFilter === 'all' ? 'bg-[#ecebf6] text-violet' : 'text-[rgba(21,21,58,.6)] hover:bg-[#F6F6FB]',
              )}
            >
              Semua ({sorted.length})
            </button>
            <button
              type="button"
              onClick={() => setHistoryFilter('red_flag')}
              className={cn(
                'rounded-full px-3 py-1.5 text-[13px] font-medium',
                historyFilter === 'red_flag' ? 'bg-[#fee2e2] text-[#ef4444]' : 'text-[rgba(21,21,58,.6)] hover:bg-[#F6F6FB]',
              )}
            >
              Red flag ({redCount})
            </button>
          </div>
        </div>

        {list.length === 0 ? (
          <p className="py-8 text-center text-[13px] text-[rgba(21,21,58,.45)]">Belum ada observasi.</p>
        ) : (
          <div className="mt-4 flex flex-col">
            {list.map((o) => (
              <div key={o.id} className={cn('flex flex-col gap-3 border-b border-violet-light py-4 last:border-b-0')}>
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className={cn('text-[14px] font-semibold', o.red_flag ? 'text-[#ef4444]' : 'text-violet-dark')}>
                      {fmtDateID(o.tanggal)}
                      {o.red_flag && <span className="ml-2 rounded-[6px] bg-[#fee2e2] px-2 py-[2px] text-[11px] text-[#ef4444]">Red flag</span>}
                    </p>
                    <p className="text-[12px] text-[rgba(21,21,58,.6)]">
                      oleh {o.author?.name ?? 'Wali Kelas'}
                      {o.handled_at && ' · sudah ditangani BK'}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    {canWrite && (
                      <button
                        type="button"
                        onClick={() => onEdit(o)}
                        className="rounded-[8px] border border-violet-light bg-base px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-[#F6F6FB]"
                      >
                        Ubah
                      </button>
                    )}
                    {o.red_flag && !o.handled_at && (
                      <button
                        type="button"
                        onClick={() => onHandle(o.id)}
                        className="rounded-[8px] bg-[#ef4444] px-3 py-1.5 text-[12px] font-medium text-base hover:bg-[#dc2626]"
                      >
                        Tandai ditangani
                      </button>
                    )}
                  </div>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {o.items.map((k) => {
                    const it = checklist.find((x) => x.key === k)
                    const isRed = k.startsWith('red_flag:')
                    return (
                      <span
                        key={k}
                        className={cn(
                          'rounded-[6px] px-2 py-[3px] text-[12px] font-semibold',
                          isRed ? 'border border-[#ef4444] bg-base text-[#ef4444]' : 'bg-[#ecebf6] text-violet',
                        )}
                      >
                        {it ? it.label : k === 'tidak_ada' ? 'Tidak ada indikator' : k}
                      </span>
                    )
                  })}
                </div>
                {o.catatan && <p className="text-[13.5px] text-violet-dark">{o.catatan}</p>}
                {o.tindakan_awal && (
                  <p className="text-[12px] text-[rgba(21,21,58,.6)]">
                    Tindakan awal:{' '}
                    {TINDAKAN.find((t) => t.value === o.tindakan_awal)?.title ?? o.tindakan_awal}
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
