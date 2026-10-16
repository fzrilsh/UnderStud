import { api } from '@/lib/api'

/**
 * Nama file unduhan: `Summary-{NISN}-{NamaTanpaSpasi}` (mockup: Summary-...).
 * NISN kosong → pakai id siswa sebagai fallback agar nama tetap unik.
 */
function fileName(nisn: string, name: string, format: 'pdf' | 'txt'): string {
  const namaTanpaSpasi = name.replace(/\s+/g, '')
  const id = nisn?.trim() || 'tanpa-nisn'
  return `Summary-${id}-${namaTanpaSpasi}.${format}`
}

/**
 * Unduh ekspor ringkasan (PDF/TXT) lalu picu download di browser.
 * Butuh id siswa + id ringkasan; server mengembalikan file biner.
 * `student` (nisn + nama) dipakai untuk menamai file.
 */
export async function downloadSummary(
  studentId: string,
  summaryId: string,
  format: 'pdf' | 'txt' = 'pdf',
  student?: { nisn?: string; name?: string },
): Promise<void> {
  const res = await api.get(
    `/students/${studentId}/summaries/${summaryId}/export`,
    { params: { format }, responseType: 'blob' },
  )
  const url = URL.createObjectURL(res.data as Blob)
  const a = document.createElement('a')
  a.href = url
  a.download = fileName(student?.nisn ?? '', student?.name ?? '', format)
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
