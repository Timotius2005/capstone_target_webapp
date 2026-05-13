'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useMode } from '@/contexts/ModeContext'

interface NavItem {
  label: string
  labelId: string
  href: string
  icon: string
  sandboxOnly?: boolean
}

const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', labelId: 'Dashboard', href: '/dashboard', icon: '⊞' },
  { label: 'Nasabah', labelId: 'Data Nasabah', href: '/nasabah', icon: '👥' },
  { label: 'Pinjaman', labelId: 'Manajemen Pinjaman', href: '/loans', icon: '💰' },
  {
    label: 'Admin Settings',
    labelId: 'Pengaturan Admin',
    href: '/admin',
    icon: '⚙️',
    sandboxOnly: true,
  },
]

interface SidebarProps {
  open: boolean
}

export default function Sidebar({ open }: SidebarProps) {
  const pathname = usePathname()
  const { mode } = useMode()
  const isSandbox = mode === 'sandbox'

  const items = NAV_ITEMS.filter((item) => !item.sandboxOnly || isSandbox)

  return (
    <>
      {/* Overlay for mobile */}
      {open && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-30 lg:hidden" />
      )}

      <aside
        className={`fixed left-0 top-0 h-screen z-40 transition-all duration-300 ease-in-out ${
          open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0 lg:w-0 lg:overflow-hidden'
        } w-64`}
      >
        <div className="h-full flex flex-col bg-white/80 dark:bg-[#0a1020]/90 backdrop-blur-2xl border-r border-slate-200/30 dark:border-slate-700/30 shadow-glass">
          {/* ── Brand ─────────────────────────────── */}
          <div className="px-6 py-5 border-b border-slate-200/30 dark:border-slate-700/30">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 animated-gradient rounded-xl flex items-center justify-center text-white font-bold text-base shadow-glow flex-shrink-0">
                DS
              </div>
              <div className="min-w-0">
                <p className="font-bold text-slate-800 dark:text-white text-sm leading-tight">
                  PT. Dana Sejahtera
                </p>
                <p className="text-xs text-slate-400 mt-0.5">Fintech Management</p>
              </div>
            </div>
          </div>

          {/* ── Navigation ────────────────────────── */}
          <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
            <p className="px-3 py-2 text-[10px] font-semibold text-slate-400 uppercase tracking-widest">
              Navigasi
            </p>

            {items.map((item) => {
              const isActive =
                pathname === item.href || pathname.startsWith(item.href + '/')
              const isAdmin = item.sandboxOnly

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 group ${
                    isActive
                      ? 'nav-active'
                      : isAdmin
                      ? 'text-amber-400/80 hover:text-amber-400 hover:bg-amber-500/10 border border-dashed border-amber-500/30'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800/50'
                  }`}
                >
                  <span className="text-base w-5 text-center">{item.icon}</span>
                  <span className="flex-1">{item.labelId}</span>
                  {isAdmin && !isActive && (
                    <span className="text-[9px] font-bold text-amber-400/60 uppercase tracking-widest bg-amber-500/10 px-1.5 py-0.5 rounded">
                      sandbox
                    </span>
                  )}
                  {isActive && (
                    <span className="w-1.5 h-1.5 bg-white/60 rounded-full" />
                  )}
                </Link>
              )
            })}
          </nav>

          {/* ── Footer ────────────────────────────── */}
          <div className="px-4 py-4 border-t border-slate-200/20 dark:border-slate-700/30">
            <div className="px-3 py-2 rounded-xl bg-slate-100/50 dark:bg-slate-800/30">
              <p className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider">
                Capstone Project
              </p>
              <p className="text-xs text-slate-500 dark:text-slate-500 mt-0.5">
                PT. Dana Sejahtera © 2024
              </p>
            </div>
          </div>
        </div>
      </aside>
    </>
  )
}
