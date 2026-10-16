import { useRef, useState } from 'react'
import { toast } from 'sonner'

import { IconClose, IconFileDoc, IconUploadBig } from '@/components/icons'
import { api } from '@/lib/api'
import { getApiErrorMessage } from '@/lib/errors'
import { useQueryClient } from '@tanstack/react-query'
import { cn } from '@/lib/utils'

/** Modal "Tambah siswa" (mockup home.js `addStudentModal`): Nama + NISN. */
export function AddStudentModal({
  kelas,
  pending,
  onSubmit,
  onClose,
}: {
  kelas: string
  pending: boolean
  onSubmit: (payload: { name: string; nisn: string; kelas: string }) => void
  onClose: () => void
}) {
  const [name, setName] = useState('')
  const [nisn, setNisn] = useState('')
  const [error, setError] = useState('')

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    const n = name.trim()
    const i = nisn.trim()
    setError('')
    if (!n) {
      setError('Nama siswa wajib diisi.')
      return
    }
    if (!/^\d{1,30}$/.test(i)) {
      setError('NISN harus berupa angka (1–30 digit).')
      return
    }
    onSubmit({ name: n, nisn: i, kelas })
  }

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-[rgba(21,21,58,.45)] p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className="relative w-full max-w-[440px] rounded-2xl bg-white p-6 shadow-[0_24px_60px_rgba(21,21,58,.25)]">
        <button
          type="button"
          title="Tutup"
          onClick={onClose}
          className="absolute right-4 top-4 flex size-7 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60"
        >
          <IconClose />
        </button>
        <h3 className="text-[18px] font-bold text-violet-dark">Tambah siswa</h3>
        <p className="mt-1.5 text-[13px] text-[rgba(21,21,58,.7)]">
          Tambah siswa secara manual ke <b>{kelas}</b>. Cukup isi data wajib, nilai dan catatan bisa dilengkapi nanti.
        </p>
        <form onSubmit={submit} className="mt-4 flex flex-col gap-4">
          <div>
            <label className="mb-1.5 block text-[13px] font-semibold text-violet-dark">
              Nama siswa <span className="text-[#B42318]">*</span>
            </label>
            <input
              autoFocus
              value={name}
              maxLength={80}
              onChange={(e) => setName(e.target.value)}
              className="input"
              placeholder="Nama lengkap siswa"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-[13px] font-semibold text-violet-dark">
              NISN <span className="text-[#B42318]">*</span>
            </label>
            <input
              value={nisn}
              maxLength={30}
              onChange={(e) => setNisn(e.target.value)}
              className="input"
              placeholder="Contoh: 2060691511"
            />
            <p className="mt-1 text-[11.5px] text-[rgba(21,21,58,.45)]">
              Hanya angka. Dipakai untuk mencocokkan impor E-Rapor.
            </p>
          </div>
          {error && <p className="text-[13px] font-medium text-[#B42318]">{error}</p>}
          <div className="mt-1 flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-[10px] border border-violet-light bg-white px-4 py-2.5 text-[14px] font-medium text-violet-dark hover:bg-[#F6F6FB]"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={pending}
              className="rounded-[10px] bg-violet px-4 py-2.5 text-[14px] font-semibold text-base hover:bg-violet-hover disabled:opacity-50"
            >
              {pending ? 'Menambah…' : 'Tambah siswa'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

interface ImportRow {
  nisn?: string
  name?: string
  kelas?: string
}

interface ImportPreviewData {
  preview_token: string
  baru?: ImportRow[]
  update?: ImportRow[]
  errors?: Array<{ baris?: number; nisn?: string; pesan?: string }>
  ringkasan?: { baru?: number; update?: number; error?: number }
}

/** Modal "Impor E-Rapor" (mockup home.js `importModal`): pilih file → preview → commit. */
export function ImportModal({ kelas, onClose }: { kelas: string; onClose: () => void }) {
  const qc = useQueryClient()
  const fileRef = useRef<HTMLInputElement>(null)
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<ImportPreviewData | null>(null)
  const [loading, setLoading] = useState(false)
  const [committing, setCommitting] = useState(false)
  const [error, setError] = useState('')
  const [dragOver, setDragOver] = useState(false)

  const handleFile = async (f: File) => {
    setFile(f)
    setPreview(null)
    setError('')
    setLoading(true)
    try {
      const form = new FormData()
      form.append('file', f)
      const res = await api.post<{ data: ImportPreviewData }>('/students/import/preview', form)
      setPreview(res.data.data)
    } catch (err) {
      setError(getApiErrorMessage(err, 'Gagal membaca file leger.'))
      setFile(null)
    } finally {
      setLoading(false)
    }
  }

  const commit = async () => {
    if (!preview) return
    setCommitting(true)
    try {
      await api.post('/students/import/commit', { preview_token: preview.preview_token })
      toast.success('Impor selesai.')
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['students-count'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
      void qc.invalidateQueries({ queryKey: ['classes'] })
      onClose()
    } catch (err) {
      toast.error(getApiErrorMessage(err, 'Gagal menyimpan impor'))
    } finally {
      setCommitting(false)
    }
  }

  const baru = preview?.baru ?? []
  const update = preview?.update ?? []
  const errs = preview?.errors ?? []

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-[rgba(21,21,58,.45)] p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className="scroll-thin relative max-h-[88vh] w-full max-w-[760px] overflow-y-auto rounded-2xl bg-white p-6 shadow-[0_24px_60px_rgba(21,21,58,.25)]">
        <button
          type="button"
          title="Tutup"
          onClick={onClose}
          className="absolute right-4 top-4 flex size-7 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60"
        >
          <IconClose />
        </button>
        <h3 className="text-[18px] font-bold text-violet-dark">Impor E-Rapor</h3>
        <p className="mt-1.5 text-[13px] text-[rgba(21,21,58,.7)]">
          Pilih file leger (.csv atau .xlsx) untuk <b>{kelas}</b>. Baris siswa ditambah bila NISN baru, atau diperbarui
          bila sudah ada.
        </p>

        {/* Drop zone */}
        <label
          onDragEnter={(e) => {
            e.preventDefault()
            setDragOver(true)
          }}
          onDragOver={(e) => {
            e.preventDefault()
            setDragOver(true)
          }}
          onDragLeave={() => setDragOver(false)}
          onDrop={(e) => {
            e.preventDefault()
            setDragOver(false)
            const f = e.dataTransfer.files[0]
            if (f) void handleFile(f)
          }}
          className={cn(
            'mt-4 flex cursor-pointer flex-col items-center justify-center gap-2 rounded-[10px] border-2 border-dashed border-[#c3c2e3] bg-[#fcfcfe] px-6 py-8 text-center transition hover:border-[#3d3ba6] hover:bg-[#f6f6fb]',
            dragOver && 'border-[#3d3ba6] bg-[#f6f6fb]',
          )}
        >
          <span className="flex size-11 items-center justify-center rounded-full bg-[#ecebf6] text-violet">
            <IconUploadBig />
          </span>
          <span className="text-[14px] font-medium text-violet-dark">Click to choose a file</span>
          <span className="text-[12px] text-[rgba(21,21,58,.55)]">or drag &amp; drop · .csv or .xlsx · max 10 MB</span>
          <input
            ref={fileRef}
            type="file"
            accept=".csv,.xlsx,text/csv"
            className="sr-only"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) void handleFile(f)
            }}
          />
        </label>

        {/* Chip file terpilih */}
        {file && (
          <div className="mt-3 flex items-center gap-3 rounded-[10px] border border-[#e2e2f2] bg-white px-3 py-2.5">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-[#ecebf6] text-violet">
              <IconFileDoc />
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-[13.5px] font-medium text-violet-dark">{file.name}</p>
              <p className="text-[11.5px] text-[rgba(21,21,58,.55)]">{Math.round(file.size / 1024)} KB</p>
            </div>
            <button
              type="button"
              onClick={() => {
                setFile(null)
                setPreview(null)
                setError('')
                if (fileRef.current) fileRef.current.value = ''
              }}
              className="flex size-7 items-center justify-center rounded-md text-[rgba(21,21,58,.5)] hover:bg-[#f1f1f5]"
            >
              <IconClose />
            </button>
          </div>
        )}

        {error && <p className="mt-3 text-[13px] font-medium text-[#B42318]">{error}</p>}
        {loading && <p className="mt-3 text-[13px] text-[rgba(21,21,58,.6)]">Membaca file…</p>}

        {/* Preview */}
        {preview && (
          <div className="mt-4">
            <p className="text-[13px] text-violet-dark">
              <b>{baru.length}</b> siswa baru · <b>{update.length}</b> diperbarui
              {errs.length > 0 && (
                <>
                  {' '}
                  · <span className="text-[#B42318]">{errs.length} baris gagal</span>
                </>
              )}
            </p>
            {(baru.length > 0 || update.length > 0) && (
              <div className="mt-3 overflow-x-auto">
                <table className="w-full min-w-[520px] border-collapse text-[13px]">
                  <thead>
                    <tr>
                      <th className="whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-left font-semibold text-[rgba(21,21,58,.55)]">
                        NISN
                      </th>
                      <th className="whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-left font-semibold text-[rgba(21,21,58,.55)]">
                        Nama
                      </th>
                      <th className="whitespace-nowrap bg-[#F6F6FB] px-3.5 py-2.5 text-left font-semibold text-[rgba(21,21,58,.55)]">
                        Status
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {baru.map((r, i) => (
                      <tr key={`b-${i}`} className="border-b border-[#F1F1F5]">
                        <td className="px-3.5 py-2 text-violet-dark">{r.nisn ?? '—'}</td>
                        <td className="px-3.5 py-2 text-violet-dark">{r.name ?? '—'}</td>
                        <td className="px-3.5 py-2">
                          <span className="rounded-[6px] bg-[#ecebf6] px-2 py-[3px] text-[12px] font-semibold text-violet">
                            Baru
                          </span>
                        </td>
                      </tr>
                    ))}
                    {update.map((r, i) => (
                      <tr key={`u-${i}`} className="border-b border-[#F1F1F5]">
                        <td className="px-3.5 py-2 text-violet-dark">{r.nisn ?? '—'}</td>
                        <td className="px-3.5 py-2 text-violet-dark">{r.name ?? '—'}</td>
                        <td className="px-3.5 py-2">
                          <span className="rounded-[6px] bg-violet-light px-2 py-[3px] text-[12px] font-semibold text-violet-dark">
                            Diperbarui
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            {errs.length > 0 && (
              <ul className="mt-3 list-inside list-disc text-[12.5px] text-[#B42318]">
                {errs.slice(0, 5).map((e, i) => (
                  <li key={i}>
                    Baris {e.baris ?? '?'}: {e.pesan ?? 'gagal dibaca'}
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        <div className="mt-5 flex flex-wrap justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-[10px] border border-violet-light bg-white px-4 py-2.5 text-[14px] font-medium text-violet-dark hover:bg-[#F6F6FB]"
          >
            Batal
          </button>
          <button
            type="button"
            disabled={!preview || committing || (baru.length === 0 && update.length === 0)}
            onClick={() => void commit()}
            className="rounded-[10px] bg-violet px-4 py-2.5 text-[14px] font-semibold text-base hover:bg-violet-hover disabled:opacity-50"
          >
            {committing ? 'Menyimpan…' : 'Impor'}
          </button>
        </div>
      </div>
    </div>
  )
}
