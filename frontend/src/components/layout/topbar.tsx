import type { ReactNode } from 'react'

/**
 * Topbar halaman (padanan UI.topbar mockup): judul + subjudul opsional + slot aksi.
 * Ruang kiri di mobile sudah disediakan untuk tombol hamburger sidebar (fixed).
 */
export default function Topbar({
  title,
  subtitle,
  actions,
}: {
  title: string
  subtitle?: ReactNode
  actions?: ReactNode
}) {
  return (
    <header className="flex flex-wrap items-center gap-4 px-5 pb-4 pt-6 md:px-8 md:pt-8">
      <div className="min-w-0 flex-1 pl-12 md:pl-0">
        <h1 className="font-heading text-[26px] font-bold leading-tight text-violet-dark md:text-[30px]">
          {title}
        </h1>
        {subtitle && <div className="mt-1 text-[13.5px] text-[rgba(21,21,58,.6)]">{subtitle}</div>}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </header>
  )
}
