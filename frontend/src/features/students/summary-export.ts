import { api } from '@/lib/api'

/**
 * Unduh ekspor ringkasan (PDF/TXT) lalu picu download di browser.
 * Butuh id siswa + id ringkasan; server mengembalikan file biner.
 */
export async function downloadSummary(
  studentId: string,
  summaryId: string,
  format: 'pdf' | 'txt' = 'pdf',
): Promise<void> {
  const res = await api.get(
    `/students/${studentId}/summaries/${summaryId}/export`,
    { params: { format }, responseType: 'blob' },
  )
  const url = URL.createObjectURL(res.data as Blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ringkasan-${format === 'pdf' ? 'ringkasan.pdf' : 'ringkasan.txt'}`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
