import type { components } from '@/types/api.gen'

type ErrorResponse = components['schemas']['ErrorResponse']

/**
 * Ambil pesan error dari axios error (shape backend {"error":{code,message}}).
 * Padanan getApiErrorMessage di features/auth/hooks.ts — dipindah ke lib agar
 * dipakai semua fitur tanpa impor lintas-fitur.
 */
export function getApiErrorMessage(err: unknown, fallback: string): string {
  const maybe = err as { response?: { data?: ErrorResponse } }
  return maybe?.response?.data?.error?.message ?? fallback
}

/** Ambil kode error backend (mis. OUT_OF_RANGE, AI_NOT_CONFIGURED). */
export function getApiErrorCode(err: unknown): string | undefined {
  const maybe = err as { response?: { data?: ErrorResponse } }
  return maybe?.response?.data?.error?.code
}
