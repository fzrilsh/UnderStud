import type { ReactNode } from 'react'

/**
 * Topbar halaman. Dua varian mengikuti mockup:
 * - `gradient` (default): header kustom halaman BK (Dashboard BK, Daftar Siswa,
 *   Detail Siswa) — judul 28/36px `text-gradient-heading` + border bawah.
 * - `plain`: padanan `UI.topbar` mockup (dipakai Dashboard Wali Kelas) —
 *   judul 26/30px solid, tanpa border.
 * Ruang kiri di mobile sudah disediakan untuk tombol hamburger sidebar (fixed).
 */
export default function Topbar({
  title,
  subtitle,
  actions,
  variant = 'gradient',
}: {
  title: string
  subtitle?: ReactNode
  actions?: ReactNode
  variant?: 'gradient' | 'plain'
}) {
  const isGradient = variant === 'gradient'
  return (
    <header
      className={
        isGradient
          ? 'flex flex-wrap items-start justify-between gap-4 border-b border-[#c3c2e3] px-5 pb-6 pt-6 md:px-8 md:pt-9'
          : 'flex flex-wrap items-center gap-4 px-5 pb-4 pt-6 md:px-8 md:pt-8'
      }
    >
      <div className="min-w-0 flex-1 pl-12 md:pl-0">
        <h1
          className={
            isGradient
              ? 'font-heading text-[28px] font-bold leading-[1.25] text-gradient-heading md:text-[36px]'
              : 'text-[26px] font-bold leading-tight text-violet-dark md:text-[30px]'
          }
        >
          {title}
        </h1>
        {subtitle &&
          (isGradient ? (
            <div className="mt-1.5 flex flex-wrap items-center gap-x-1 gap-y-0.5 text-[15px] text-violet-dark md:text-[16px]">
              {subtitle}
            </div>
          ) : (
            <div className="mt-1 text-[13.5px] text-[rgba(21,21,58,.6)]">{subtitle}</div>
          ))}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </header>
  )
}
