import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query'
import { toast } from 'sonner'

import { api } from '@/lib/api'
import { getApiErrorMessage } from '@/lib/errors'
import type { SummaryView } from '@/features/dashboard/hooks'

type DataEnvelope<T> = { data: T }

export type StudentDetail = {
  student: {
    id: string
    nisn: string
    name: string
    kelas: string
    archived_at: string | null
    has_red_flag: boolean
    e_rapor_complete: boolean
    observed: boolean
    has_summary: boolean
  }
  profile: {
    nilai_erapor: Array<{ mapel: string; nilai: number; periode: string }>
    sakit: number
    izin: number
    alpa: number
    minat_bakat: string
    tanggal_lahir: string
    catatan_psikologis: string
    catatan_kasus_prestasi: string
    masukan_bk: string
    pendapat_wali_kelas: string
    updated_at: string
  }
  observations: ObservationView[]
  summaries: SummaryView[]
  latest_summary: SummaryView | null
}

export type ObservationView = {
  id: string
  student_id: string
  items: string[]
  categories: string[]
  red_flag: boolean
  handled_at: string | null
  catatan?: string
  tanggal: string
  tindakan_awal?: string
  author: { name: string }
}

export type ChecklistItem = { key: string; kategori: string; label: string }

/** GET /students/{id} — detail lengkap (profil + observasi + riwayat ringkasan). */
export function useStudentDetail(id: string | undefined): UseQueryResult<StudentDetail> {
  return useQuery({
    queryKey: ['student', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await api.get<DataEnvelope<StudentDetail>>(`/students/${id}`)
      return res.data.data
    },
  })
}

/** GET /checklist — katalog 28 item observasi (statis). */
export function useChecklist(): UseQueryResult<ChecklistItem[]> {
  return useQuery({
    queryKey: ['checklist'],
    staleTime: Infinity,
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ChecklistItem[]>>('/checklist')
      return res.data.data
    },
  })
}

export interface ProfilePayload {
  nilai_erapor: Array<{ mapel: string; nilai: number; periode: string }>
  sakit: number
  izin: number
  alpa: number
  minat_bakat: string
  tanggal_lahir: string
  catatan_psikologis: string
  catatan_kasus_prestasi: string
  masukan_bk: string
  pendapat_wali_kelas: string
}

/** PUT /students/{id}/profile — simpan profil (WK saja). */
/** PATCH /students/{id} — ubah data dasar siswa (nama/kelas/nisn). WK saja. */
export interface UpdateStudentPayload {
  name: string
  kelas: string
  nisn: string
}

export function useUpdateStudent(id: string): UseMutationResult<void, Error, UpdateStudentPayload> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (payload: UpdateStudentPayload) => {
      await api.patch(`/students/${id}`, payload)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['student', id] })
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['students-count'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
      void qc.invalidateQueries({ queryKey: ['classes'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menyimpan data siswa')),
  })
}

export function useSaveProfile(id: string): UseMutationResult<void, Error, ProfilePayload> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (payload: ProfilePayload) => {
      await api.put(`/students/${id}/profile`, payload)
    },
    onSuccess: () => {
      toast.success('Perubahan disimpan.')
      void qc.invalidateQueries({ queryKey: ['student', id] })
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menyimpan perubahan')),
  })
}

export interface ObservationPayload {
  items: string[]
  catatan?: string
  tanggal: string
  tindakan_awal?: string
}

/** POST /students/{id}/observations — tambah entri observasi (WK saja). */
export function useCreateObservation(id: string): UseMutationResult<void, Error, ObservationPayload> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (payload: ObservationPayload) => {
      await api.post(`/students/${id}/observations`, payload)
    },
    onSuccess: () => {
      toast.success('Observasi disimpan.')
      void qc.invalidateQueries({ queryKey: ['student', id] })
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menyimpan observasi')),
  })
}

/** PUT /observations/{id} — ubah entri observasi (WK saja). */
export function useUpdateObservation(studentId: string): UseMutationResult<void, Error, { id: string } & ObservationPayload> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...payload }: { id: string } & ObservationPayload) => {
      await api.put(`/observations/${id}`, payload)
    },
    onSuccess: () => {
      toast.success('Observasi diperbarui.')
      void qc.invalidateQueries({ queryKey: ['student', studentId] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal memperbarui observasi')),
  })
}

/** POST /observations/{id}/handle — tandai red flag sudah ditangani (BK saja). */
export function useHandleRedFlag(studentId: string): UseMutationResult<void, Error, string> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (observationId: string) => {
      await api.post(`/observations/${observationId}/handle`)
    },
    onSuccess: () => {
      toast.success('Red flag ditandai sudah ditangani.')
      void qc.invalidateQueries({ queryKey: ['student', studentId] })
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menandai red flag')),
  })
}

/**
 * POST /students/{id}/summaries — trigger generate AI (asinkron, 202).
 * TANPA polling: hasil muncul saat halaman dimuat ulang berikutnya.
 */
export function useGenerateSummary(id: string): UseMutationResult<void, Error, void> {
  return useMutation({
    mutationFn: async () => {
      await api.post(`/students/${id}/summaries`)
    },
    onSuccess: () => {
      toast.info('Ringkasan sedang dibuat. Tunggu beberapa saat, lalu muat ulang halaman ini.')
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal memulai pembuatan ringkasan')),
  })
}

/** GET /students/import/template?format= — unduh template impor leger (WK). */
export function useImportTemplate(): UseMutationResult<void, Error, 'xlsx' | 'csv'> {
  return useMutation({
    mutationFn: async (format: 'xlsx' | 'csv') => {
      const res = await api.get('/students/import/template', {
        params: { format },
        responseType: 'blob',
      })
      const url = URL.createObjectURL(res.data as Blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `template-leger.${format}`
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal mengunduh template')),
  })
}
