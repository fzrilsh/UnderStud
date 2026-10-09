import { create } from 'zustand'
import type { components } from '@/types/api.gen'

export type ApiUser = components['schemas']['User']

interface AuthState {
  user: ApiUser | null
  setUser: (user: ApiUser | null) => void
  clear: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  clear: () => set({ user: null }),
}))
