import { Link } from 'react-router-dom'
import { buttonVariants } from '@/components/ui/button'

export default function NotFoundPage() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-4">
      <p className="text-6xl font-bold tracking-tight">404</p>
      <p className="text-muted-foreground">Halaman tidak ditemukan.</p>
      <Link to="/" className={buttonVariants({ variant: 'default' })}>
        Kembali ke Dashboard
      </Link>
    </div>
  )
}
