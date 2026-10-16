import { useMemo } from 'react'

import { IconAiMagicLight, IconUpload } from '@/components/icons'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { fmtDateID } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { StudentDetail } from '@/features/students/detail-hooks'

export interface DataFormState {
  nama: string
  kelas: string
  nisn: string
  tanggal_lahir: string
  nilai_erapor: Array<{ mapel: string; nilai: number; periode: string }>
  sakit: number
  izin: number
  alpa: number
  minat_bakat: string
  pendapat_wali_kelas: string
  masukan_bk: string
  catatan_kasus_prestasi: string
  catatan_psikologis: string
}

export function initForm(d: StudentDetail): DataFormState {
  return {
    nama: d.student.name,
    kelas: d.student.kelas,
    nisn: d.student.nisn,
    tanggal_lahir: d.profile.tanggal_lahir || '',
    nilai_erapor: d.profile.nilai_erapor ?? [],
    sakit: d.profile.sakit,
    izin: d.profile.izin,
    alpa: d.profile.alpa,
    minat_bakat: d.profile.minat_bakat,
    pendapat_wali_kelas: d.profile.pendapat_wali_kelas,
    masukan_bk: d.profile.masukan_bk,
    catatan_kasus_prestasi: d.profile.catatan_kasus_prestasi,
    catatan_psikologis: d.profile.catatan_psikologis,
  }
}

function NoteField({
  label,
  value,
  max,
  restricted,
  canWrite,
  onChange,
}: {
  label: string
  value: string
  max: number
  restricted?: boolean
  canWrite: boolean
  onChange: (v: string) => void
}) {
  return (
    <div>
      <div className="mb-1.5 flex items-baseline justify-between gap-2">
        <Label className="text-[13px] font-semibold">
          {label}
          {restricted && (
            <span className="ml-2 rounded-[6px] bg-[#ecebf6] px-2 py-[3px] text-[12px] font-semibold text-violet">
              Restricted
            </span>
          )}
        </Label>
        {canWrite && (
          <span className="text-[11px] text-[rgba(21,21,58,.4)]">
            {value.length} / {max}
          </span>
        )}
      </div>
      {canWrite ? (
        <textarea
          value={value}
          maxLength={max}
          rows={2}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Tulis di sini…"
          className="input min-h-[46px] resize-y"
        />
      ) : (
        <div className="min-h-[46px] whitespace-pre-wrap rounded-[10px] border border-[#E2E2F2] bg-[#FAFAFE] px-3 py-2.5 text-[14px] text-[rgba(21,21,58,.75)]">
          {value || '—'}
        </div>
      )}
    </div>
  )
}

/** Tab "Data siswa & observasi": kartu Data siswa + Nilai E-Rapor + Catatan guru. */
export default function DataTab({
  detail,
  form,
  setForm,
  canWrite,
}: {
  detail: StudentDetail
  form: DataFormState
  setForm: (f: DataFormState) => void
  canWrite: boolean
}) {
  const age = useMemo(() => {
    if (!form.tanggal_lahir) return null
    const diff = Date.now() - new Date(form.tanggal_lahir).getTime()
    const years = Math.floor(diff / (365.25 * 24 * 3600 * 1000))
    return years > 0 && years < 100 ? years : null
  }, [form.tanggal_lahir])

  const patch = (p: Partial<DataFormState>) => setForm({ ...form, ...p })

  // Kartu `UI.card` mockup: radius 16, p-5, putih, tanpa shadow.
  const card = 'rounded-[16px] border border-[rgba(226,226,242,.9)] bg-white p-5'
  const cardTitle = 'text-[17px] font-bold leading-tight text-violet-dark'
  const cardSub = 'mt-1 text-[12.5px] text-[rgba(21,21,58,.55)]'

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-[20px] font-bold text-violet-dark">Data profil</h2>
        <p className="mt-0.5 text-[12.5px] text-[rgba(21,21,58,.55)]">
          Disimpan ke profil siswa dan dipakai sebagai masukan ringkasan AI.
        </p>
      </div>

      {/* Data siswa */}
      <div className={card}>
        <h3 className={cardTitle}>Data siswa</h3>
        <p className={cardSub}>Nama dan kelas wajib diisi.</p>
        <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <Label className="mb-1.5 text-[13px] font-semibold">Nama</Label>
            <Input value={form.nama} readOnly className="input bg-[#FAFAFE]" />
          </div>
          <div>
            <Label className="mb-1.5 text-[13px] font-semibold">Kelas & jurusan</Label>
            <Input value={form.kelas} readOnly className="input bg-[#FAFAFE]" />
          </div>
          <div>
            <div className="mb-1.5 flex items-baseline justify-between">
              <Label className="text-[13px] font-semibold">NISN</Label>
              <span className="text-[11px] text-[rgba(21,21,58,.4)]">Opsional</span>
            </div>
            <Input value={form.nisn} readOnly className="input bg-[#FAFAFE]" />
          </div>
          <div>
            <div className="mb-1.5 flex items-baseline justify-between">
              <Label className="text-[13px] font-semibold">Tanggal lahir</Label>
              <span className="text-[11px] text-[rgba(21,21,58,.4)]">Opsional</span>
            </div>
            {canWrite ? (
              <Input
                type="date"
                value={form.tanggal_lahir}
                onChange={(e) => patch({ tanggal_lahir: e.target.value })}
                className="input"
              />
            ) : (
              <div className="rounded-[10px] border border-[#E2E2F2] bg-[#FAFAFE] px-3 py-2.5 text-[14px] text-[rgba(21,21,58,.75)]">
                {form.tanggal_lahir ? fmtDateID(form.tanggal_lahir) : '—'}
              </div>
            )}
            {age !== null && <p className="mt-1 text-[11.5px] text-[rgba(21,21,58,.45)]">Age {age} · auto-calculated</p>}
          </div>
        </div>
      </div>

      {/* Nilai E-Rapor */}
      <div className={card}>
        <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
          <div className="min-w-0">
            <h3 className={cardTitle}>Nilai E-Rapor</h3>
            <p className={cardSub}>
              Diimpor {fmtDateID(detail.profile.updated_at)} dari .xlsx · Nilai tetap bisa diubah
            </p>
          </div>
          {canWrite && (
            <button
              type="button"
              className="inline-flex shrink-0 items-center gap-1.5 rounded-[8px] bg-[#ecebf6] px-3 py-1.5 text-[13px] font-medium text-violet hover:bg-violet-light"
            >
              <IconUpload /> Impor ulang
            </button>
          )}
        </div>
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-[13.5px]">
            <thead>
              <tr>
                <th className="whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-left font-semibold text-[rgba(21,21,58,.55)]">
                  Mata pelajaran
                </th>
                <th className="whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-left font-semibold text-[rgba(21,21,58,.55)]">
                  Nilai
                </th>
                <th className="whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-left font-semibold text-[rgba(21,21,58,.55)]">
                  Periode
                </th>
              </tr>
            </thead>
            <tbody>
              {form.nilai_erapor.length === 0 && (
                <tr>
                  <td colSpan={3} className="py-6 text-center text-[13px] text-[rgba(21,21,58,.45)]">
                    Belum ada nilai.
                  </td>
                </tr>
              )}
              {form.nilai_erapor.map((r, i) => (
                <tr key={r.mapel + i} className="border-b border-[#F1F1F5] hover:bg-[#FAFAFE]">
                  <td className="px-3.5 py-2.5 font-medium text-violet-dark">{r.mapel}</td>
                  <td className="px-3.5 py-2.5">
                    {canWrite ? (
                      <input
                        type="number"
                        value={r.nilai}
                        onChange={(e) => {
                          const next = [...form.nilai_erapor]
                          next[i] = { ...r, nilai: Number(e.target.value) }
                          patch({ nilai_erapor: next })
                        }}
                        className="w-[90px] rounded-[8px] border border-[#E2E2F2] px-2 py-1.5 text-center font-semibold text-violet-dark outline-none focus:border-violet"
                      />
                    ) : (
                      <span className="font-semibold text-violet-dark">{r.nilai}</span>
                    )}
                  </td>
                  <td className="px-3.5 py-2.5 text-[rgba(21,21,58,.55)]">{r.periode}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
          {canWrite ? (
            <button
              type="button"
              className="inline-flex items-center gap-1 rounded-[8px] bg-[#ecebf6] px-3 py-1.5 text-[13px] font-medium text-violet hover:bg-violet-light"
            >
              + Tambah mapel
            </button>
          ) : (
            <span />
          )}
          <span className="text-[12px] text-[rgba(21,21,58,.5)]">
            Kehadiran dari E-Rapor · Sakit {form.sakit} · Izin {form.izin} · Alpa {form.alpa}
          </span>
        </div>
      </div>

      {/* Catatan guru */}
      <div className={card}>
        <h3 className={cardTitle}>Catatan guru</h3>
        <p className={cardSub}>Semua opsional · Tulis yang Anda ketahui, bukan kesimpulan</p>
        <div className="mt-4 flex flex-col gap-5">
          <NoteField
            label="Minat & bakat"
            value={form.minat_bakat}
            max={1000}
            canWrite={canWrite}
            onChange={(v) => patch({ minat_bakat: v })}
          />
          <NoteField
            label="Pendapat wali kelas"
            value={form.pendapat_wali_kelas}
            max={2000}
            canWrite={canWrite}
            onChange={(v) => patch({ pendapat_wali_kelas: v })}
          />
          <NoteField
            label="Masukan BK"
            value={form.masukan_bk}
            max={2000}
            canWrite={canWrite}
            onChange={(v) => patch({ masukan_bk: v })}
          />
          <NoteField
            label="Catatan kasus & prestasi"
            value={form.catatan_kasus_prestasi}
            max={2000}
            canWrite={canWrite}
            onChange={(v) => patch({ catatan_kasus_prestasi: v })}
          />
          <NoteField
            label="Catatan psikologis"
            value={form.catatan_psikologis}
            max={2000}
            restricted
            canWrite={canWrite}
            onChange={(v) => patch({ catatan_psikologis: v })}
          />
        </div>
        <p className="mt-3 text-[11.5px] text-[rgba(21,21,58,.5)]">
          Hanya Anda dan BK yang bisa melihat ini · Tidak pernah dikirim ke AI
        </p>
      </div>
    </div>
  )
}

/** Kartu aksi bawah (Batal / Simpan perubahan / Simpan & buat ringkasan AI). */
export function DataActionBar({
  hint,
  saving,
  onCancel,
  onSave,
  onSaveAndGenerate,
}: {
  hint: string
  saving: boolean
  onCancel: () => void
  onSave: () => void
  onSaveAndGenerate: () => void
}) {
  return (
    <div className="mt-6 flex flex-wrap items-center gap-4 rounded-[10px] border border-violet-light bg-base px-6 py-4 shadow-[0px_4px_8px_0px_rgba(0,0,0,0.1)]">
      <p className="min-w-0 flex-1 text-[12px] text-[rgba(21,21,58,.6)]">{hint}</p>
      <div className="flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-[8px] bg-[#ecebf6] px-3.5 py-2.5 text-[14px] font-medium text-violet hover:bg-violet-light"
        >
          Batal
        </button>
        <button
          type="button"
          onClick={onSave}
          disabled={saving}
          className={cn(
            'rounded-[8px] bg-[#ecebf6] px-3.5 py-2.5 text-[14px] font-medium text-violet hover:bg-violet-light',
            saving && 'opacity-50',
          )}
        >
          Simpan perubahan
        </button>
        <button
          type="button"
          onClick={onSaveAndGenerate}
          disabled={saving}
          className={cn(
            'inline-flex items-center gap-1.5 rounded-[8px] bg-violet-dark px-3.5 py-2.5 text-[14px] font-medium text-base hover:bg-[#24245c]',
            saving && 'opacity-50',
          )}
        >
          <IconAiMagicLight /> Simpan &amp; buat ringkasan AI
        </button>
      </div>
    </div>
  )
}
