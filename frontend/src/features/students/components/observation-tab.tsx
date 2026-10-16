import { useMemo, useState } from 'react'

import { IconCalendar, IconCheck } from '@/components/icons'
import { cn } from '@/lib/utils'
import { fmtDateID, fmtDateTimeID } from '@/lib/format'
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

/** "Okt 16" — format kolom tanggal timeline (mockup fmtMonDay). */
const MON_EN = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
function fmtMonDay(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '-'
  return `${MON_EN[d.getMonth()]} ${d.getDate()}`
}

/** "09:33" — jam kolom tanggal timeline (mockup fmtClock). */
function fmtClock(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

/** Label pendek tindakan awal (mockup actionLabelShort). */
function actionLabelShort(v?: string | null): string {
  if (v === 'konsultasi' || v === 'konsultasi_bk') return 'Konsultasi ke BK'
  if (v === 'eskalasi' || v === 'eskalasi_segera') return 'Tandai sekarang'
  return 'Pantau'
}

/** Catatan lanjutan tindakan awal (mockup actionNote). */
function actionNote(o: ObservationView): string {
  if (o.tindakan_awal === 'eskalasi' || o.tindakan_awal === 'eskalasi_segera') {
    return o.handled_at ? `BK diberi tahu ${fmtDateTimeID(o.handled_at)}` : 'BK diberi tahu · menunggu tinjauan'
  }
  if (o.tindakan_awal === 'konsultasi' || o.tindakan_awal === 'konsultasi_bk') return 'Di pertemuan BK berikutnya'
  return 'Terus amati di kelas'
}

const isEskalasi = (v?: string | null) => v === 'eskalasi' || v === 'eskalasi_segera'

/** Kartu Indikator (checklist 28 item) + Detail observasi + riwayat. */
export default function ObservationTab({
  observations,
  checklist,
  canWrite,
  submitting,
  onCreate,
  onHandle,
  studentName,
}: {
  observations: ObservationView[]
  checklist: ChecklistItem[]
  canWrite: boolean
  submitting: boolean
  onCreate: (payload: { items: string[]; catatan?: string; tanggal: string; tindakan_awal?: string }) => void
  onHandle: (observationId: string) => void
  /** Nama siswa — dipakai di strip peringatan red flag. */
  studentName: string
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

  // Kartu `UI.card` mockup: radius 16, p-5, putih, tanpa shadow.
  const card = 'rounded-[16px] border border-[rgba(226,226,242,.9)] bg-white p-5'
  const cardTitle = 'text-[17px] font-bold leading-tight text-violet-dark'
  const cardSub = 'mt-1 text-[12.5px] text-[rgba(21,21,58,.55)]'

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-[20px] font-bold text-violet-dark">Observasi</h2>
        <p className="mt-0.5 text-[12.5px] text-[rgba(21,21,58,.55)]">
          Menyimpan menambah entri baru ke riwayat. Kosongkan bila hanya memperbarui data profil.
        </p>
      </div>

      {canWrite && (
        <div className={card}>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 className={cardTitle}>Indikator</h3>
              <p className={cardSub}>
                Centang hanya yang Anda lihat langsung · <b>{tidakAda ? 0 : selected.length}</b> dari {checklist.length} dipilih ·
                Bukan diagnosis, tanpa skor risiko
              </p>
            </div>
            <label className="chk items-center rounded-xl border border-violet-light px-3 py-1.5">
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
                      <span className="rounded-[6px] bg-[#ecebf6] px-2 py-[3px] text-[12px] font-semibold text-violet">{picked}</span>
                    )}
                  </div>
                  <div className="grid grid-cols-1 gap-x-4 gap-y-1.5 sm:grid-cols-2 lg:grid-cols-4">
                    {items.map((it) => (
                      <label key={it.key} className="chk">
                        <input
                          type="checkbox"
                          checked={selected.includes(it.key)}
                          disabled={tidakAda}
                          onChange={() => toggle(it.key)}
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
                  Needs immediate alert to BK. A red flag never waits for a score.
                </p>
              </div>
              <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
                {redItems.map((it) => (
                  <label key={it.key} className="chk chk-red hover:!bg-[#fef2f2]">
                    <input
                      type="checkbox"
                      checked={selected.includes(it.key)}
                      disabled={tidakAda}
                      onChange={() => toggle(it.key)}
                    />
                    <span className="text-[13px] leading-[1.4] text-[#ef4444]">{it.label}</span>
                  </label>
                ))}
              </div>
            </div>
          </div>

          {/* Strip peringatan red flag (mockup) */}
          <div className="mt-4 rounded-[8px] border border-[#ef4444] bg-[#fcfcfe] px-3 py-2.5">
            <p className="text-[12px] font-medium text-[#ef4444]">
              Menyimpan langsung menandai {studentName} untuk BK dan menandai siswa di daftar Anda.
            </p>
          </div>

          {/* Detail observasi */}
          <div className="mt-6">
            <h3 className="mb-3 text-[15px] font-bold text-violet-dark">Detail observasi</h3>
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1.5 block text-[13px] font-semibold text-violet-dark">Tanggal observasi</label>
                <div className="flex max-w-[240px] items-center gap-2 rounded-[10px] border border-[#E2E2F2] bg-white px-3 py-2.5 text-[14px] text-violet-dark">
                  <IconCalendar className="shrink-0 text-violet-dark" />
                  {fmtDateID(new Date().toISOString())}
                </div>
                <p className="mt-1 text-[11.5px] text-[rgba(21,21,58,.45)]">
                  Otomatis hari ini. Tanggal setelah hari ini tidak bisa dipilih.
                </p>
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
                  className="input min-h-[80px] resize-y"
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
                          className="sr-only"
                        />
                        <span className={cn('obs-radio', t.red && 'obs-radio-red')} data-checked={sel ? '1' : '0'} />
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
              className="mt-4 rounded-[10px] bg-violet px-4 py-2.5 text-[14px] font-semibold text-base hover:bg-violet-hover disabled:opacity-40"
            >
              {submitting ? 'Menyimpan…' : 'Simpan observasi'}
            </button>
          </div>
        </div>
      )}

      {/* Riwayat observasi — heading DI LUAR kartu (mockup). */}
      <div>
        <h2 className="font-heading text-[24px] font-medium text-violet-dark">Riwayat observasi</h2>
        <p className="mt-1 text-[12.5px] text-[rgba(21,21,58,.55)]">
          Setiap entri menyimpan penulis, tanggal, dan jejak edit. BK membaca riwayat yang sama.
        </p>
      </div>

      <div className={card}>
        {/* Filter chips (mockup pill) + "Terbaru dulu". */}
        <div className="mb-5 flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => setHistoryFilter('all')}
            aria-pressed={historyFilter === 'all'}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-[8px] px-3 py-2.5 whitespace-nowrap text-[14px] font-semibold transition-colors',
              historyFilter === 'all'
                ? 'bg-[#e2e2f2] text-violet-dark'
                : 'border border-[#E2E2F2] bg-white text-[rgba(21,21,58,.7)] hover:bg-[#F6F6FB]',
            )}
          >
            Semua <span className="text-[12px] font-normal opacity-60">{sorted.length}</span>
          </button>
          <button
            type="button"
            onClick={() => setHistoryFilter('red_flag')}
            aria-pressed={historyFilter === 'red_flag'}
            className="inline-flex items-center gap-1.5 rounded-[8px] border border-[#ef4444] bg-[#fcfcfe] px-3 py-2.5 whitespace-nowrap text-[14px] font-medium text-[#ef4444] hover:bg-[#fef2f2]"
          >
            Red flag
            <span className="flex size-5 items-center justify-center rounded-full bg-[#ef4444] text-[12px] font-semibold text-[#fcfcfe]">
              {redCount}
            </span>
          </button>
          <span className="ml-auto text-[12px] text-[rgba(21,21,58,.6)]">Terbaru dulu</span>
        </div>

        {list.length === 0 ? (
          <p className="px-4 py-12 text-center text-[14px] text-[rgba(21,21,58,.45)]">Belum ada observasi.</p>
        ) : (
          <div className="flex flex-col">
            {list.map((o, i) => (
              <TimelineEntry
                key={o.id}
                o={o}
                checklist={checklist}
                isFirst={i === 0}
                isLast={i === list.length - 1}
                canWrite={canWrite}
                onHandle={onHandle}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

/** Satu entri timeline: kolom tanggal + rail + kartu body (mockup 318:3657). */
function TimelineEntry({
  o,
  checklist,
  isFirst,
  isLast,
  canWrite,
  onHandle,
}: {
  o: ObservationView
  checklist: ChecklistItem[]
  isFirst: boolean
  isLast: boolean
  canWrite: boolean
  onHandle: (id: string) => void
}) {
  const red = !!o.red_flag
  // Kelompokkan item per kategori (mockup: baris label kategori + chips).
  const byCat = new Map<string, string[]>()
  for (const k of o.items) {
    const cat = k.split(':')[0]
    if (!byCat.has(cat)) byCat.set(cat, [])
    byCat.get(cat)!.push(k)
  }

  return (
    <div className="flex w-full items-start gap-4">
      {/* Kolom tanggal (88px) */}
      <div className="w-[88px] shrink-0 pt-[14px]">
        <p className={cn('text-[14px] font-semibold leading-[1.5]', red ? 'text-[#ef4444]' : 'text-violet-dark')}>
          {fmtMonDay(o.tanggal)}
        </p>
        <p className="text-[12px] leading-[1.5] text-[rgba(21,21,58,.6)]">
          {new Date(o.tanggal).getFullYear()} · {fmtClock(o.tanggal)}
        </p>
      </div>

      {/* Rail: spacer + dot + garis */}
      <div className="flex w-3 shrink-0 flex-col items-center gap-1 self-stretch">
        <span className="h-[14px] w-[2px] shrink-0" style={{ background: isFirst ? 'transparent' : '#e2e2f2' }} />
        <span
          className="size-3 shrink-0 rounded-full"
          style={{
            background: red ? '#ef4444' : '#3d3ba6',
            boxShadow: red ? '0 0 0 3px #fff4f4' : '0 0 0 3px #fcfcfe',
          }}
        />
        <span className="w-[2px] flex-1" style={{ background: isLast ? 'transparent' : '#e2e2f2' }} />
      </div>

      {/* Body card */}
      <div className="min-w-0 flex-1 pb-4">
        <div
          className={cn(
            'flex flex-col gap-3.5 rounded-[8px] border p-4',
            red ? 'border-[#ef4444] bg-[#fff4f4]' : 'border-[#e2e2f2] bg-[#fcfcfe]',
          )}
        >
          {/* Baris penulis */}
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[14px] font-semibold text-violet-dark">{o.author?.name ?? 'Wali Kelas'}</span>
            <span className="rounded-[6px] bg-[#e2e2f2] px-2 py-[3px] text-[12px] font-semibold text-violet-dark">Anda</span>
            <span className="text-[12px] text-[rgba(21,21,58,.6)]">Wali Kelas · {o.items.length} indikator</span>
            {red && (
              <span className="ml-auto rounded-[6px] bg-[#ef4444] px-2 py-[3px] text-[12px] font-semibold text-[#fcfcfe]">
                Red flag
              </span>
            )}
          </div>

          {/* Baris kategori + chips */}
          <div className="flex flex-col gap-2">
            {[...byCat.entries()].map(([cat, keys]) => {
              const isRed = cat === 'red_flag'
              return (
                <div key={cat} className="flex flex-col gap-1.5 sm:flex-row sm:items-center sm:gap-3">
                  <span
                    className={cn(
                      'shrink-0 text-[12px] font-medium sm:w-[160px]',
                      isRed ? 'text-[#ef4444]' : 'text-violet-dark',
                    )}
                  >
                    {KATEGORI_LABEL[cat] ?? cat}
                  </span>
                  <div className="flex flex-wrap gap-1.5">
                    {keys.map((k) => {
                      const it = checklist.find((x) => x.key === k)
                      return (
                        <span
                          key={k}
                          className={cn(
                            'rounded-[6px] px-2 py-[3px] text-[12px] font-semibold',
                            isRed ? 'border border-[#ef4444] bg-[#fcfcfe] text-[#ef4444]' : 'bg-[#ecebf6] text-violet',
                          )}
                        >
                          {it ? it.label : k === 'tidak_ada' ? 'Tidak ada indikator' : k}
                        </span>
                      )
                    })}
                  </div>
                </div>
              )
            })}
          </div>

          {/* Catatan */}
          {o.catatan && (
            <div className={cn('rounded-[8px] px-3.5 py-3', red ? 'bg-white' : 'bg-[#ecebf6]')}>
              <p className="text-[12px] font-medium text-[rgba(21,21,58,.6)]">Catatan</p>
              <p className="mt-0.5 text-[14px] leading-[1.5] text-violet-dark">{o.catatan}</p>
            </div>
          )}

          {/* Footer tindakan awal */}
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[12px] text-[rgba(21,21,58,.6)]">Tindakan awal</span>
            <span
              className={cn(
                'rounded-[6px] px-2 py-[3px] text-[12px] font-semibold',
                isEskalasi(o.tindakan_awal) ? 'bg-[#ef4444] text-[#fcfcfe]' : 'bg-[#ecebf6] text-violet',
              )}
            >
              {actionLabelShort(o.tindakan_awal)}
            </span>
            <span
              className={cn(
                'text-[12px]',
                isEskalasi(o.tindakan_awal) ? 'text-[#ef4444]' : 'text-[rgba(21,21,58,.6)]',
              )}
            >
              {actionNote(o)}
            </span>
          </div>

          {/* Tandai ditangani (red flag belum ditangani) */}
          {red && !o.handled_at && canWrite && (
            <div className="border-t border-[#f3c9c9] pt-3">
              <button
                type="button"
                onClick={() => onHandle(o.id)}
                className="inline-flex items-center gap-1.5 rounded-[8px] border border-[#e2e2f2] bg-[#fcfcfe] px-3 py-1.5 text-[12px] font-medium text-violet hover:bg-[#F6F6FB]"
              >
                <IconCheck /> Tandai ditangani
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
