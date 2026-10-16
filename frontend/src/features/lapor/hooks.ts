import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query'
import { toast } from 'sonner'

import { api } from '@/lib/api'
import { getApiErrorMessage } from '@/lib/errors'

type DataEnvelope<T> = { data: T }
type Meta = { total: number; limit: number; offset: number }

export type LastMessage = { sender: 'SISWA' | 'BK'; body: string; created_at: string }

export type ChatSummary = {
  id: string
  identity_mode: 'anonim' | 'profil'
  status: string
  last_message_at: string
  last_message: LastMessage | null
  has_new_reply: boolean
  unread_count: number
}

export type MessageView = {
  id: string
  sender: 'SISWA' | 'BK'
  body: string
  attachments: Array<{ id: string; name: string; mime: string; size_bytes: number }>
  created_at: string
}

export type ChatDetail = {
  chat: { id: string; identity_mode: 'anonim' | 'profil'; status: string }
  messages: MessageView[]
}

export type AnonProfile = { name?: string; kelas?: string }

export type LaporBootstrap = { profile: AnonProfile; chats: ChatSummary[] }

/** GET /lapor — bootstrap sesi anon (set cookie) + profil + daftar chat. */
export function useLaporBootstrap(): UseQueryResult<LaporBootstrap> {
  return useQuery({
    queryKey: ['lapor', 'bootstrap'],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<LaporBootstrap>>('/lapor')
      return res.data.data
    },
    staleTime: 60_000,
  })
}

/** GET /lapor/chats — daftar laporan siswa. */
export function useLaporChats(): UseQueryResult<ChatSummary[]> {
  return useQuery({
    queryKey: ['lapor', 'chats'],
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ChatSummary[]> & { meta: Meta }>('/lapor/chats', {
        params: { limit: 50, offset: 0 },
      })
      return res.data.data
    },
  })
}

/** GET /lapor/chats/{id} — detail percakapan (membuka → reset unread). */
export function useLaporChat(id: string | null): UseQueryResult<ChatDetail> {
  const qc = useQueryClient()
  return useQuery({
    queryKey: ['lapor', 'chat', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await api.get<DataEnvelope<ChatDetail>>(`/lapor/chats/${id}`)
      // Badge unread di-reset server-side → segarkan daftar.
      void qc.invalidateQueries({ queryKey: ['lapor', 'chats'] })
      return res.data.data
    },
  })
}

export interface GeoInput {
  lat?: number
  lng?: number
  accuracy?: number
}

export interface CreateLaporInput extends GeoInput {
  body: string
  identityMode: 'anonim' | 'profil'
  files?: File[]
}

/** POST /lapor/chats — kirim laporan baru (multipart). */
export function useCreateLapor(): UseMutationResult<{ id: string }, Error, CreateLaporInput> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateLaporInput) => {
      const fd = new FormData()
      fd.append('body', input.body)
      fd.append('identity_mode', input.identityMode)
      if (input.lat !== undefined) fd.append('lat', String(input.lat))
      if (input.lng !== undefined) fd.append('lng', String(input.lng))
      if (input.accuracy !== undefined) fd.append('accuracy', String(input.accuracy))
      for (const f of input.files ?? []) fd.append('files', f)
      const res = await api.post<DataEnvelope<{ id: string }>>('/lapor/chats', fd)
      return res.data.data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['lapor', 'chats'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal mengirim laporan')),
  })
}

/** POST /lapor/chats/{id}/messages — balas sebagai siswa. */
export function useReplyLapor(id: string): UseMutationResult<void, Error, { body: string; files?: File[] }> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ body, files }) => {
      const fd = new FormData()
      fd.append('body', body)
      for (const f of files ?? []) fd.append('files', f)
      await api.post(`/lapor/chats/${id}/messages`, fd)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['lapor', 'chat', id] })
      void qc.invalidateQueries({ queryKey: ['lapor', 'chats'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal mengirim pesan')),
  })
}

/** DELETE /lapor/chats/{id} — sembunyikan laporan di sisi siswa. */
export function useHideLapor(): UseMutationResult<void, Error, string> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/lapor/chats/${id}`)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['lapor', 'chats'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menghapus laporan')),
  })
}

/** PUT /lapor/profile — simpan identitas pelapor. */
export function useSaveAnonProfile(): UseMutationResult<void, Error, AnonProfile> {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (profile: AnonProfile) => {
      await api.put('/lapor/profile', profile)
    },
    onSuccess: () => {
      toast.success('Identitas disimpan.')
      void qc.invalidateQueries({ queryKey: ['lapor', 'bootstrap'] })
    },
    onError: (err) => toast.error(getApiErrorMessage(err, 'Gagal menyimpan identitas')),
  })
}

/** Ambil geolokasi browser (timeout 10s). Kembalikan {} bila gagal. */
export function getGeolocation(): Promise<GeoInput> {
  return new Promise((resolve) => {
    if (!('geolocation' in navigator)) return resolve({})
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve({ lat: pos.coords.latitude, lng: pos.coords.longitude, accuracy: pos.coords.accuracy }),
      () => resolve({}),
      { timeout: 10_000, maximumAge: 60_000 },
    )
  })
}
