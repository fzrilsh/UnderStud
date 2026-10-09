import { Link, Outlet, useLocation } from 'react-router-dom'
import { Loader2, LogOut } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'
import { useLogout } from '@/features/auth/hooks'

const navItems = [
  { to: '/', label: 'Dashboard' },
  // Tambah menu fitur di sini pas requirement keluar.
]

export default function AppLayout() {
  const user = useAuthStore((s) => s.user)
  const logout = useLogout()
  const location = useLocation()

  return (
    <div className="min-h-svh bg-background">
      <div className="flex">
        {/* Sidebar */}
        <aside className="hidden w-60 shrink-0 border-r md:block">
          <div className="flex h-14 items-center border-b px-4 font-semibold tracking-tight">
            Hackathon Starter
          </div>
          <nav className="flex flex-col gap-1 p-3">
            {navItems.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className={cn(
                  'rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-foreground',
                  location.pathname === item.to && 'bg-accent text-foreground',
                )}
              >
                {item.label}
              </Link>
            ))}
          </nav>
        </aside>

        {/* Main */}
        <div className="flex min-w-0 flex-1 flex-col">
          <header className="flex h-14 items-center justify-between border-b px-4">
            <span className="text-sm text-muted-foreground md:hidden">Hackathon Starter</span>
            <div className="ml-auto flex items-center gap-3">
              <span className="text-sm text-muted-foreground">{user?.name}</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => logout.mutate()}
                disabled={logout.isPending}
              >
                {logout.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <LogOut className="h-4 w-4" />
                )}
                Logout
              </Button>
            </div>
          </header>
          <main className="flex-1 p-6">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  )
}
