import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query'
import { toast } from 'sonner'

import { api } from '@/lib/api'
import { getApiErrorMessage } from '@/lib/errors'
import type { components } from '@/types/api.gen'

export type StudentListItem = components['schemas']['StudentItem'] & {
  e_rapor_complete: boolean
  observed: boolean
  has_summary: boolean
}

export type StudentFilter = 'all' | 'red_flag' | 'archived' | 'incomplete' | 'not_observed'

type DataEnvelope<T> = { data: T }
type Meta = { total: number; limit: number; offset: number }

export interface StudentListParams {
  kelas?: string
  q?: string
  filter?: StudentFilter
  includeArchived?: boolean
  limit?: number
  offset?: number
}

/** GET /students — daftar siswa (BK semua, WK kelasnya). */
export function useStudents(params: StudentListParams): UseQueryResult<{ items: StudentListItem[]; total: number }> {
  return useQuery({
    queryKey: ['students', params],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<StudentListItem[]> & { meta: Meta }>('/students', {
        params: {
          kelas: params.kelas || undefined,
          q: params.q || undefined,
          filter: params.filter && params.filter !== 'all' ? params.filter : undefined,
          include_archived: params.includeArchived ? 'true' : undefined,
          limit: params.limit ?? 10,
          offset: params.offset ?? 0,
        },
      })
      return { items: res.data.data, total: res.data.meta.total }
    },
  })
}

/** Jumlah siswa untuk badge filter (limit 1 → ambil meta.total saja). */
export function useStudentCount(
  filter: StudentFilter,
  kelas?: string,
  includeArchived = false,
): UseQueryResult<number> {
  return useQuery({
    queryKey: ['students-count', filter, kelas ?? '', includeArchived],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<StudentListItem[]> & { meta: Meta }>('/students', {
        params: {
          kelas: kelas || undefined,
          filter: filter !== 'all' ? filter : undefined,
          include_archived: includeArchived ? 'true' : undefined,
          limit: 1,
          offset: 0,
        },
      })
      return res.data.meta.total
    },
  })
}

/** POST /students/{id}/archive. */
export function useArchiveStudent(): UseMutationResult<void, Error, string> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/students/${id}/archive`)
    },
    onSuccess: () => {
      toast.success('Siswa diarsipkan.')
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['students-count'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
      void qc.invalidateQueries({ queryKey: ['classes'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal mengarsipkan siswa')),
  })
}

/** POST /students/{id}/restore. */
export function useRestoreStudent(): UseMutationResult<void, Error, string> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/students/${id}/restore`)
    },
    onSuccess: () => {
      toast.success('Siswa dikembalikan ke daftar aktif.')
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['students-count'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
      void qc.invalidateQueries({ queryKey: ['classes'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal membuka arsip siswa')),
  })
}

/** POST /students — tambah siswa manual (WK). */
export interface CreateStudentInput {
  name: string
  nisn: string
  kelas: string
}

export function useCreateStudent(): UseMutationResult<void, Error, CreateStudentInput> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input) => {
      await api.post('/students', input)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['students'] })
      void qc.invalidateQueries({ queryKey: ['students-count'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
      void qc.invalidateQueries({ queryKey: ['classes'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menambah siswa')),
  })
}

/** POST /students/{id}/summaries — trigger generate AI (asinkron). */
export function useGenerateStudentSummary(): UseMutationResult<void, Error, string> {
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/students/${id}/summaries`)
    },
    onSuccess: () => {
      toast.info('Ringkasan sedang dibuat. Tunggu beberapa saat, lalu muat ulang halaman.')
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal memulai pembuatan ringkasan')),
  })
}
