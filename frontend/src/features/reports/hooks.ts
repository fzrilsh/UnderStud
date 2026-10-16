import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query'
import { toast } from 'sonner'

import { api } from '@/lib/api'
import { getApiErrorMessage } from '@/lib/errors'

type DataEnvelope<T> = { data: T }
type Meta = { total: number; limit: number; offset: number }

export type LastMessage = { sender: 'SISWA' | 'BK'; body: string; created_at: string }

export type ReportItem = {
  id: string
  status: 'baru' | 'dibuka' | 'dilewati'
  pinned: boolean
  drafted: boolean
  pinned_at: string | null
  drafted_at: string | null
  identity: { mode: 'anonim' | 'profil'; name: string | null; kelas: string | null }
  last_message_at: string
  last_message: LastMessage | null
  unread_by_staff: boolean
}

export type MessageView = {
  id: string
  sender: 'SISWA' | 'BK'
  body: string
  attachments: Array<{ id: string; name: string; mime: string; size_bytes: number }>
  created_at: string
}

export type ReportDetail = { chat: ReportItem; messages: MessageView[] }

/** GET /reports — antrean (Pin ∪ Draft ∪ Antrean, status baru/dibuka). */
export function useReports(limit = 50): UseQueryResult<{ items: ReportItem[]; total: number }> {
  return useQuery({
    queryKey: ['reports', limit],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ReportItem[]> & { meta: Meta }>('/reports', {
        params: { limit, offset: 0 },
      })
      return { items: res.data.data, total: res.data.meta.total }
    },
  })
}

/** GET /reports/history — riwayat (dilewati + di-hide BK). */
export function useReportHistory(limit = 50): UseQueryResult<{ items: ReportItem[]; total: number }> {
  return useQuery({
    queryKey: ['reports-history', limit],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ReportItem[]> & { meta: Meta }>('/reports/history', {
        params: { limit, offset: 0 },
      })
      return { items: res.data.data, total: res.data.meta.total }
    },
  })
}

/** GET /reports/{id} — detail percakapan (membuka → status baru→dibuka). */
export function useReport(id: string | null): UseQueryResult<ReportDetail> {
  return useQuery({
    queryKey: ['report', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ReportDetail>>(`/reports/${id}`)
      return res.data.data
    },
  })
}

function useReportAction(
  fn: (id: string) => Promise<unknown>,
  message: string,
): UseMutationResult<void, Error, string> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await fn(id)
    },
    onSuccess: (_d, id) => {
      toast.success(message)
      void qc.invalidateQueries({ queryKey: ['reports'] })
      void qc.invalidateQueries({ queryKey: ['reports-history'] })
      void qc.invalidateQueries({ queryKey: ['report', id] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal memproses laporan')),
  })
}

export const usePinReport = () => useReportAction((id) => api.post(`/reports/${id}/pin`), 'Disematkan.')
export const useUnpinReport = () => useReportAction((id) => api.delete(`/reports/${id}/pin`), 'Sematan dilepas.')
export const useDraftReport = () => useReportAction((id) => api.post(`/reports/${id}/draft`), 'Disimpan ke draft.')
export const useUndraftReport = () => useReportAction((id) => api.delete(`/reports/${id}/draft`), 'Dikeluarkan dari draft.')
export const useNextReport = () => useReportAction((id) => api.post(`/reports/${id}/next`), 'Chat diselesaikan.')
export const useSkipReport = () => useReportAction((id) => api.post(`/reports/${id}/skip`), 'Chat dilewati.')
export const useHideReport = () => useReportAction((id) => api.delete(`/reports/${id}`), 'Laporan disembunyikan.')

/** POST /reports/{id}/messages — balas sebagai BK. */
export function useReplyReport(id: string): UseMutationResult<void, Error, string> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (body: string) => {
      await api.post(`/reports/${id}/messages`, { body })
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['report', id] })
      void qc.invalidateQueries({ queryKey: ['reports'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal mengirim balasan')),
  })
}
