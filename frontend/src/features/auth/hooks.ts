import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query'
import { toast } from 'sonner'

import { api } from '@/lib/api'
import { getApiErrorMessage } from '@/lib/errors'
import { useAuthStore, type ApiUser } from '@/stores/auth-store'
import type { components } from '@/types/api.gen'
import type { LoginInput } from './schemas'

type AuthResponse = components['schemas']['AuthResponse']

export { getApiErrorMessage }

/** GET /auth/me — dipakai ProtectedRoute; sinkron dengan zustand store. */
export function useMe(): UseQueryResult<ApiUser> {
  const setUser = useAuthStore((s) => s.setUser)
  const clear = useAuthStore((s) => s.clear)
  const navigate = useNavigate()

  const query = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const res = await api.get<AuthResponse>('/auth/me')
      return res.data.user
    },
    retry: 0,
    staleTime: 30_000,
  })

  // Sync store saat data berubah.
  useEffect(() => {
    if (query.data) setUser(query.data)
  }, [query.data, setUser])

  // 401 dari interceptor → clear state + redirect.
  useEffect(() => {
    const handler = () => {
      clear()
      navigate('/login', { replace: true })
    }
    window.addEventListener('auth:unauthorized', handler)
    return () => window.removeEventListener('auth:unauthorized', handler)
  }, [clear, navigate])

  return query
}

/** POST /auth/login. */
export function useLogin(): UseMutationResult<ApiUser, Error, LoginInput> {
  const setUser = useAuthStore((s) => s.setUser)
  const navigate = useNavigate()

  return useMutation({
    mutationFn: async (input: LoginInput) => {
      const res = await api.post<AuthResponse>('/auth/login', input)
      return res.data.user
    },
    onSuccess: (user) => {
      setUser(user)
      toast.success('Login berhasil')
      navigate('/', { replace: true })
    },
    onError: (err) => {
      toast.error(getApiErrorMessage(err, 'Login gagal'))
    },
  })
}

/** POST /auth/logout — clear cookie server-side + state lokal. */
export function useLogout(): UseMutationResult<{ ok: boolean }, Error, void> {
  const clear = useAuthStore((s) => s.clear)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      const res = await api.post<{ ok: boolean }>('/auth/logout')
      return res.data
    },
    onSuccess: () => {
      clear()
      queryClient.clear()
      navigate('/login', { replace: true })
    },
  })
}
