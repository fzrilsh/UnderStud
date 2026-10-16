import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { NavLink } from 'react-router-dom'

import {
  IconChat,
  IconChevrons,
  IconClose,
  IconDashboard,
  IconLogout,
  IconMenu,
  IconStudents,
  IconUserLight,
} from '@/components/icons'
import { useDashboardWK } from '@/features/dashboard/hooks'
import { cn } from '@/lib/utils'
import type { ApiUser } from '@/stores/auth-store'

const PREF_KEY = 'understud_sidebar_collapsed'

/** Preferensi collapse sidebar (tersimpan di localStorage). */
function collapsedPref(): boolean {
  try {
    return localStorage.getItem(PREF_KEY) === '1'
  } catch {
    return false
  }
}

export interface SidebarProps {
  user: ApiUser | null
  /** Slot tambahan di bawah nav (mis. section chat / kelas saya). */
  children?: ReactNode
  /** Callback logout (dipanggil setelah konfirmasi). */
  onLogout: () => void
}

const navClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    'flex items-center gap-2.5 rounded-[10px] px-2.5 py-2 text-[14px] font-medium text-violet-dark hover:bg-violet-light/60',
    isActive && 'bg-[rgba(195,194,227,.6)]',
  )

/** Panel "Kelas saya" untuk Wali Kelas (mockup UI.sidebar wkClass). */
function KelasSaya({ kelas }: { kelas: string }) {
  const { data } = useDashboardWK()
  return (
    <div className="mt-3 flex flex-col gap-3 pl-1.5">
      <p className="text-[14px] font-semibold text-violet-dark">Kelas saya</p>
      <div className="flex flex-col gap-1 rounded-[8px] bg-[#e2e2f2] px-3 py-2.5">
        <p className="text-[14px] font-semibold text-violet-dark">{kelas}</p>
        <p className="text-[12px] text-[rgba(21,21,58,.7)]">
          {data?.students ?? 0} siswa · Sem. ganjil 2026/27
        </p>
      </div>
    </div>
  )
}

/**
 * Sidebar staf (Figma 84:993 / 474:2926): 240px, gradient, bisa collapse ke 64px
 * (logo jadi tombol expand). Mobile: drawer dengan overlay + tombol close.
 */
export default function Sidebar({ user, children, onLogout }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(collapsedPref)
  const [drawerOpen, setDrawerOpen] = useState(false)

  const toggle = useCallback(() => {
    setCollapsed((v) => {
      const next = !v
      try {
        localStorage.setItem(PREF_KEY, next ? '1' : '0')
      } catch {
        /* ignore */
      }
      return next
    })
  }, [])

  const closeDrawer = useCallback(() => setDrawerOpen(false), [])

  // Tutup drawer saat resize ke desktop.
  useEffect(() => {
    const onResize = () => {
      if (window.innerWidth >= 768) setDrawerOpen(false)
    }
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const isBK = user?.role === 'BK'

  return (
    <>
      {/* Hamburger mobile (tombol ini yang membuka drawer). */}
      <button
        type="button"
        title="Menu"
        onClick={() => setDrawerOpen(true)}
        className="fixed left-4 top-4 z-30 flex size-10 items-center justify-center rounded-xl border border-white/60 bg-white/70 text-violet-dark backdrop-blur md:hidden"
      >
        <IconMenu />
      </button>

      {/* Overlay transparan (menangkap klik di luar drawer). */}
      {drawerOpen && (
        <div
          className="fixed inset-0 z-[65] md:hidden"
          onClick={closeDrawer}
          aria-hidden="true"
        />
      )}

      <aside
        data-collapsed={collapsed ? '1' : '0'}
        className={cn(
          // Sticky setinggi layar (padanan mockup `sticky top-0 h-screen`).
          'bg-sidebar-gradient sticky top-0 z-[70] flex h-svh w-[240px] shrink-0 flex-col rounded-tr-[20px] border-r border-t border-violet-light px-[14px] pt-6 transition-[width] duration-[250ms] ease-out',
          collapsed && 'w-16 px-2',
          // Mobile: drawer tersembunyi kecuali dibuka.
          drawerOpen
            ? 'fixed inset-y-0 left-0 flex h-full md:sticky md:h-svh'
            : 'hidden md:flex',
        )}
      >
        {/* Header: logo + nama + toggle */}
        <div className={cn('flex h-[68px] w-full items-center gap-2.5 px-1', collapsed && 'justify-center px-0')}>
          {collapsed ? (
            <button
              type="button"
              title="Bentangkan"
              onClick={toggle}
              className="flex size-9 shrink-0 items-center justify-center rounded-lg hover:bg-violet-light/60"
            >
              <img src="/understud-logo.png" alt="Understud" className="size-9 object-contain" />
            </button>
          ) : (
            <>
              <img src="/understud-logo.png" alt="Understud" className="size-9 shrink-0 object-contain" />
              <span className="flex-1 truncate text-[16px] font-semibold text-violet-dark">Understud</span>
              <button
                type="button"
                title="Ciutkan"
                onClick={toggle}
                className="hidden size-6 shrink-0 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60 md:flex"
              >
                <IconChevrons className="size-6" />
              </button>
            </>
          )}
          <button
            type="button"
            title="Tutup"
            onClick={closeDrawer}
            className="flex size-8 shrink-0 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60 md:hidden"
          >
            <IconClose />
          </button>
        </div>

        {/* Nav */}
        <nav className="mt-4 flex w-full flex-col gap-0.5">
          <NavLink to="/" end className={(s) => cn(navClass(s), collapsed && 'justify-center px-0')} title="Dashboard">
            <span className="shrink-0 text-violet-dark">
              <IconDashboard />
            </span>
            {!collapsed && <span className="truncate">Dashboard</span>}
          </NavLink>
          {isBK && (
            <NavLink to="/reports" className={(s) => cn(navClass(s), collapsed && 'justify-center px-0')} title="Chat">
              <span className="shrink-0 text-violet-dark">
                <IconChat />
              </span>
              {!collapsed && <span className="truncate">Chat</span>}
            </NavLink>
          )}
          <NavLink to="/students" className={(s) => cn(navClass(s), collapsed && 'justify-center px-0')} title="Daftar siswa">
            <span className="shrink-0 text-violet-dark">
              <IconStudents />
            </span>
            {!collapsed && <span className="truncate">Daftar siswa</span>}
          </NavLink>
        </nav>

        {/* Panel "Kelas saya" (Wali Kelas) — mockup UI.sidebar wkClass. */}
        {!collapsed && !isBK && user?.kelas && <KelasSaya kelas={user.kelas} />}

        {/* Slot tambahan (section chat) */}
        {!collapsed && children && (
          <div className="scroll-thin mt-4 flex min-h-0 w-full flex-1 flex-col gap-1 overflow-y-auto">
            {children}
          </div>
        )}
        {collapsed && <div className="flex-1" />}

        {/* Footer: avatar + identitas + logout */}
        <div className={cn('mt-2 mb-4 w-full border-t border-violet-light pt-3 pb-2', collapsed && 'border-t-0')}>
          <div className={cn('flex items-center gap-2.5 px-2', collapsed && 'justify-center px-0')}>
            <span
              className="flex size-9 shrink-0 items-center justify-center rounded-full p-[3.75px]"
              style={{ backgroundImage: 'linear-gradient(90deg,#3D3BA6 0%,#6462BD 40%,#E9EBF8 100%)' }}
            >
              <span className="flex size-full items-center justify-center rounded-full">
                <IconUserLight />
              </span>
            </span>
            {!collapsed && (
              <>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-[13.5px] font-semibold leading-tight text-violet-dark">
                    {user?.name ?? ''}
                  </p>
                  <p className="truncate text-[12px] text-[rgba(21,21,58,.6)]">
                    {user?.role === 'BK' ? 'Guru BK' : `Wali Kelas${user?.kelas ? ` · ${user.kelas}` : ''}`}
                  </p>
                </div>
                <button
                  type="button"
                  title="Keluar"
                  onClick={() => {
                    closeDrawer()
                    onLogout()
                  }}
                  className="flex size-7 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60"
                >
                  <IconLogout />
                </button>
              </>
            )}
          </div>
        </div>
      </aside>
    </>
  )
}
