import { useQuery } from '@tanstack/react-query'
import type { UseQueryResult } from '@tanstack/react-query'

import { api } from '@/lib/api'
import type { components } from '@/types/api.gen'

// Backend selalu mengirim field ini (openapi tidak menandai `required`, jadi
// tipe generated semuanya opsional) — di sini ditegaskan agar FE aman.
export type ChartPoint = {
  label: string
  pinned: number
  drafted: number
  queued: number
}

export type ClassItem = {
  kelas: string
  grade: string
  homeroom: string
  students: number
  red_flag: number
  status: 'flagged' | 'complete' | 'not_filled'
}

export type StudentItem = components['schemas']['StudentItem']
export type SummaryView = components['schemas']['SummaryView'] & {
  version: number
  author: string
  note: string
  sources_used: number
  sources_total: number
}

export type DashboardBK = {
  chats: { waiting: number; pinned: number; drafted: number; queued: number; unqueued: number }
  red_flags: number
  summaries: number
  latest: Array<{ student: StudentItem; summary: SummaryView }>
  red_top: { student: StudentItem; tanggal: string } | null
  chart: ChartPoint[]
  classes: ClassItem[]
}

export type DashboardWK = {
  students: number
  e_rapor_complete: number
  observed: number
  red_flags: number
  red_top: StudentItem | null
  chart: ChartPoint[]
}

type DataEnvelope<T> = { data: T }

/** GET /dashboard/bk — agregat dashboard Guru BK. */
export function useDashboardBK(): UseQueryResult<DashboardBK> {
  return useQuery({
    queryKey: ['dashboard', 'bk'],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<DashboardBK>>('/dashboard/bk')
      return res.data.data
    },
  })
}

/** GET /dashboard/wk — agregat dashboard Wali Kelas (scoped ke kelasnya). */
export function useDashboardWK(): UseQueryResult<DashboardWK> {
  return useQuery({
    queryKey: ['dashboard', 'wk'],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<DashboardWK>>('/dashboard/wk')
      return res.data.data
    },
  })
}

/** GET /classes — daftar kelas + statistik. */
export function useClasses(): UseQueryResult<ClassItem[]> {
  return useQuery({
    queryKey: ['classes'],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ClassItem[]>>('/classes')
      return res.data.data
    },
  })
}
