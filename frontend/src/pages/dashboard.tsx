import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/stores/auth-store'

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)

  return (
    <div className="grid gap-4">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Halaman protected — ganti ini dengan fiturmu saat requirement keluar.
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Akun kamu</CardTitle>
          <CardDescription>Data dari GET /api/v1/auth/me</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-1 text-sm">
          <p>
            <span className="text-muted-foreground">Nama:</span> {user?.name}
          </p>
          <p>
            <span className="text-muted-foreground">Email:</span> {user?.email}
          </p>
          <p>
            <span className="text-muted-foreground">ID:</span> {user?.id}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
