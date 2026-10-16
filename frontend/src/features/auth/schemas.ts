import { z } from 'zod'

export const loginSchema = z.object({
  email: z.email('Email tidak valid'),
  password: z.string().min(1, 'Kata sandi wajib diisi'),
})

export type LoginInput = z.infer<typeof loginSchema>
