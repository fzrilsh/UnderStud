/** Footer paginasi tabel siswa (mockup `wkFooter` / footer BK). */
export function Pagination({
  page,
  pages,
  shown,
  total,
  onPrev,
  onNext,
  note,
}: {
  page: number
  pages: number
  shown: number
  total: number
  onPrev: () => void
  onNext: () => void
  /** Catatan opsional di kiri (mockup WK: " · Red flag ditandai manual…"). */
  note?: string
}) {
  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
      <p className="text-[12px] leading-[normal] text-violet-dark opacity-60">
        Menampilkan {shown} dari {total} siswa{note ? ` · ${note}` : ''}
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onPrev}
          disabled={page <= 1}
          className="rounded-[8px] bg-[#ecebf6] px-3 py-2 text-[12px] font-medium text-violet hover:bg-violet-light disabled:opacity-40"
        >
          Sebelumnya
        </button>
        <span className="px-3 py-2 text-[12px] font-medium text-violet-dark opacity-70">
          {page} / {pages}
        </span>
        <button
          type="button"
          onClick={onNext}
          disabled={page >= pages}
          className="rounded-[8px] bg-[#ecebf6] px-3 py-2 text-[12px] font-medium text-violet hover:bg-violet-light disabled:opacity-40"
        >
          Berikutnya
        </button>
      </div>
    </div>
  )
}
