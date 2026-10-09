import { z } from 'zod'

export const loginSchema = z.object({
  email: z.email('Email tidak valid'),
  password: z.string().min(1, 'Password wajib diisi'),
})

export const registerSchema = z.object({
  name: z.string().min(1, 'Nama wajib diisi'),
  email: z.email('Email tidak valid'),
  password: z.string().min(6, 'Password minimal 6 karakter'),
})

export type LoginInput = z.infer<typeof loginSchema>
export type RegisterInput = z.infer<typeof registerSchema>
