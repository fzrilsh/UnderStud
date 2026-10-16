import { useEffect, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import Topbar from '@/components/layout/topbar'
import { IconAiMagic, IconDotSm, IconDownloadLight, IconRefresh } from '@/components/icons'
import { useConfirm } from '@/components/confirm-provider'
import DataTab, { DataActionBar, initForm, type DataFormState } from '@/features/students/components/data-tab'
import ObservationTab from '@/features/students/components/observation-tab'
import SummaryTab from '@/features/students/components/summary-tab'
import {
  useChecklist,
  useCreateObservation,
  useGenerateSummary,
  useHandleRedFlag,
  useSaveProfile,
  useStudentDetail,
  useUpdateStudent,
} from '@/features/students/detail-hooks'
import { ImportModal } from '@/features/students/components/student-modals'
import { downloadSummary } from '@/features/students/summary-export'
import { useCurrentUser } from '@/features/auth/hooks'
import { fmtDateID, fmtDateTimeID } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { SummaryView } from '@/features/dashboard/hooks'
import { useSeo } from '@/lib/seo'

type Tab = 'data' | 'summary'

export default function StudentDetailPage() {
  useSeo('Detail Siswa', 'Data siswa, observasi, dan ringkasan AI.')

  const { id } = useParams<{ id: string }>()
  const [params, setParams] = useSearchParams()
  const user = useCurrentUser()
  const detail = useStudentDetail(id)
  const checklist = useChecklist()

  const tab: Tab = params.get('tab') === 'summary' ? 'summary' : 'data'
  const setTab = (t: Tab) => setParams({ tab: t === 'summary' ? 'summary' : 'data' }, { replace: true })

  const canWrite = user?.role === 'WALI_KELAS'

  const [form, setForm] = useState<DataFormState | null>(null)
  useEffect(() => {
    if (detail.data && !form) setForm(initForm(detail.data))
  }, [detail.data, form])

  const saveProfile = useSaveProfile(id ?? '')
  const updateStudent = useUpdateStudent(id ?? '')
  const createObs = useCreateObservation(id ?? '')
  const handleRed = useHandleRedFlag(id ?? '')
  const generate = useGenerateSummary(id ?? '')
  const confirm = useConfirm()
  const [importOpen, setImportOpen] = useState(false)

  const d = detail.data

  if (detail.isLoading) {
    return <div className="p-8 text-[14px] text-[rgba(21,21,58,.6)]">Memuat…</div>
  }
  if (detail.isError || !d || !form) {
    return (
      <div className="mx-auto mt-10 max-w-[420px] rounded-[10px] border border-violet-light bg-base p-10 text-center">
        <h1 className="text-[20px] font-bold text-violet-dark">Siswa tidak ditemukan</h1>
        <Link to="/students" className="mt-5 inline-flex rounded-[8px] bg-violet-dark px-4 py-2.5 text-[14px] font-medium text-base">
          Kembali ke daftar siswa
        </Link>
      </div>
    )
  }

  const lastSaved = d.summaries.length ? fmtDateID(d.summaries[0].created_at) : fmtDateID(d.profile.updated_at)
  const hint = d.latest_summary
    ? `Ringkasan saat ini: versi ${d.latest_summary.version} · ${fmtDateTimeID(d.latest_summary.created_at)}. Membuat ulang menambah versi baru.`
    : 'Belum ada ringkasan AI.'

  const saveOnly = async () => {
    // Data siswa (nama/kelas/nisn) disimpan via PATCH; profil via PUT.
    if (form.nama !== d.student.name || form.kelas !== d.student.kelas || form.nisn !== d.student.nisn) {
      await updateStudent.mutateAsync({ name: form.nama.trim(), kelas: form.kelas.trim(), nisn: form.nisn.trim() })
    }
    await saveProfile.mutateAsync({
      nilai_erapor: form.nilai_erapor,
      sakit: form.sakit,
      izin: form.izin,
      alpa: form.alpa,
      minat_bakat: form.minat_bakat,
      tanggal_lahir: form.tanggal_lahir,
      catatan_psikologis: form.catatan_psikologis,
      catatan_kasus_prestasi: form.catatan_kasus_prestasi,
      masukan_bk: form.masukan_bk,
      pendapat_wali_kelas: form.pendapat_wali_kelas,
    })
  }

  const saveAndGenerate = async () => {
    try {
      await saveOnly()
      generate.mutate()
    } catch {
      /* toast sudah dari hook */
    }
  }

  const handleDownload = async (summaryId: string) => {
    if (!id) return
    try {
      await downloadSummary(id, summaryId, 'pdf', { nisn: d.student.nisn, name: d.student.name })
      toast.success('Ringkasan diunduh.')
    } catch {
      toast.error('Gagal mengunduh ringkasan.')
    }
  }

  return (
    // Gradient halaman Ringkasan menutupi seluruh halaman termasuk header (mockup
    // menempelkan `bg-summary-page` di <html>).
    <div className={cn(tab === 'summary' && 'bg-summary-gradient min-h-svh')}>
      <Topbar
        title={tab === 'summary' ? 'Ringkasan AI' : 'Data siswa & observasi'}
        subtitle={
          <>
            <span className="whitespace-nowrap">{d.student.name}</span>
            <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
            <span className="whitespace-nowrap">NISN {d.student.nisn}</span>
            <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
            <span className="whitespace-nowrap">{d.student.kelas}</span>
            {tab === 'data' && (
              <>
                <IconDotSm className="shrink-0 text-[rgba(21,21,58,.6)]" />
                <span className="whitespace-nowrap">Terakhir disimpan {lastSaved}</span>
              </>
            )}
          </>
        }
        actions={
          <div className="flex w-full flex-wrap items-center gap-2 md:w-auto">
            <Link
              to="/students"
              className="flex flex-1 items-center justify-center gap-1.5 rounded-[8px] bg-[#ecebf6] px-3.5 py-2.5 text-[14px] font-medium text-violet hover:bg-violet-light md:flex-none"
            >
              ← Kembali ke daftar siswa
            </Link>
            {tab === 'data' ? (
              <button
                type="button"
                onClick={() => setTab('summary')}
                className="flex flex-1 items-center justify-center gap-1.5 rounded-[8px] bg-[#ecebf6] px-3.5 py-2.5 text-[14px] font-medium text-violet hover:bg-violet-light md:flex-none"
              >
                <IconAiMagic /> Lihat ringkasan AI
              </button>
            ) : (
              <>
                <button
                  type="button"
                  onClick={() => generate.mutate()}
                  disabled={generate.isPending}
                  className="flex flex-1 items-center justify-center gap-1.5 rounded-[8px] bg-[#ecebf6] px-3.5 py-2.5 text-[14px] font-medium text-violet hover:bg-violet-light disabled:opacity-50 md:flex-none"
                >
                  <IconRefresh />
                  {generate.isPending ? 'Membuat…' : 'Buat ulang'}
                </button>
                {d.latest_summary && (
                  <button
                    type="button"
                    onClick={() => handleDownload(d.latest_summary!.id)}
                    className="flex flex-1 items-center justify-center gap-1.5 rounded-[8px] bg-violet-dark px-3.5 py-2.5 text-[14px] font-medium text-base hover:bg-[#24245c] md:flex-none"
                  >
                    <IconDownloadLight /> Unduh PDF
                  </button>
                )}
              </>
            )}
          </div>
        }
      />

      <div className="px-5 pb-10 pt-6 md:px-8">
        {tab === 'data' ? (
          <>
            <DataTab detail={d} form={form} setForm={setForm} canWrite={canWrite} onImport={() => setImportOpen(true)} />
            <ObservationTab
              observations={d.observations}
              checklist={checklist.data ?? []}
              canWrite={canWrite}
              submitting={createObs.isPending}
              onCreate={(payload) => createObs.mutate(payload)}
              onHandle={(oid) => handleRed.mutate(oid)}
              studentName={d.student.name}
            />
            {canWrite && (
              <DataActionBar
                hint={hint}
                saving={saveProfile.isPending}
                onCancel={() => setForm(initForm(d))}
                onSave={() => void saveOnly()}
                onSaveAndGenerate={() => void saveAndGenerate()}
              />
            )}
          </>
        ) : (
          <SummaryTab
            detail={d}
            generating={generate.isPending}
            onGenerate={() => generate.mutate()}
            onDownload={handleDownload}
            onViewData={() => setTab('data')}
            onViewVersion={(sm: SummaryView) => {
              const c = sm.content as { headline?: string; narrative?: string }
              void confirm({
                title: `Versi ${sm.version} · ${fmtDateID(sm.created_at)}`,
                body: `${c?.headline ?? ''} — ${c?.narrative ?? ''}`,
                cancelText: 'Tutup',
                confirmText: 'Unduh',
              }).then((ok) => {
                if (ok) handleDownload(sm.id)
              })
            }}
          />
        )}
      </div>

      {importOpen && <ImportModal kelas={d.student.kelas} onClose={() => setImportOpen(false)} />}
    </div>
  )
}
