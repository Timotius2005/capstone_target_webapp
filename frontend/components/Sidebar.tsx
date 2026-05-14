'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useMode } from '@/contexts/ModeContext'

/* ── SVG Icons ─────────────────────────────────────────────────────────── */

function IconDashboard() {
  return (
    <svg className="w-[17px] h-[17px]" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
      <rect x="3" y="3" width="7" height="7" rx="1.5" />
      <rect x="14" y="3" width="7" height="7" rx="1.5" />
      <rect x="3" y="14" width="7" height="7" rx="1.5" />
      <rect x="14" y="14" width="7" height="7" rx="1.5" />
    </svg>
  )
}

function IconUsers() {
  return (
    <svg className="w-[17px] h-[17px]" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3Zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3Zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5Zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5Z" />
    </svg>
  )
}

function IconCreditCard() {
  return (
    <svg className="w-[17px] h-[17px]" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
      <rect x="2" y="5" width="20" height="14" rx="2" />
      <path strokeLinecap="round" d="M2 10h20M6 15h4" />
    </svg>
  )
}

function IconSettings() {
  return (
    <svg className="w-[17px] h-[17px]" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 0 0 2.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 0 0 1.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 0 0-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 0 0-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 0 0-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 0 0-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 0 0 1.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  )
}

/* ── Nav data ───────────────────────────────────────────────────────────── */

interface NavItem {
  labelId: string
  href: string
  icon: React.ReactNode
  sandboxOnly?: boolean
}

const NAV_ITEMS: NavItem[] = [
  { labelId: 'Dashboard',          href: '/dashboard', icon: <IconDashboard /> },
  { labelId: 'Data Nasabah',       href: '/nasabah',   icon: <IconUsers /> },
  { labelId: 'Manajemen Pinjaman', href: '/loans',     icon: <IconCreditCard /> },
  { labelId: 'Pengaturan Admin',   href: '/admin',     icon: <IconSettings />, sandboxOnly: true },
]

/* ── Component ──────────────────────────────────────────────────────────── */

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
      {/* Mobile overlay */}
      {open && <div className="fixed inset-0 bg-black/50 z-30 lg:hidden" />}

      <aside
        className={`fixed left-0 top-8 h-[calc(100vh-2rem)] z-40 transition-all duration-300 ease-in-out ${
          open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0 lg:w-0 lg:overflow-hidden'
        } w-64`}
      >
        <div className="h-full flex flex-col bg-[#0B1E3D] border-r border-white/[0.07]">

          {/* ── Brand ──────────────────────────────── */}
          <div className="px-5 py-4 border-b border-white/[0.07]">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 bg-[#1E3A8A] rounded-lg flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                DS
              </div>
              <div className="min-w-0">
                <p className="font-semibold text-white text-sm leading-tight tracking-tight">
                  PT. Dana Sejahtera
                </p>
                <p className="text-[10px] text-white/35 mt-0.5 tracking-wide">
                  Fintech Management
                </p>
              </div>
            </div>
          </div>

          {/* ── Navigation ─────────────────────────── */}
          <nav className="flex-1 px-3 py-3 overflow-y-auto space-y-0.5">
            <p className="px-3 pt-2 pb-1.5 text-[10px] font-semibold text-white/25 uppercase tracking-[0.12em]">
              Menu
            </p>

            {items.map((item) => {
              const isActive = pathname === item.href || pathname.startsWith(item.href + '/')
              const isAdmin = item.sandboxOnly

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium transition-colors duration-150 group ${
                    isActive
                      ? 'nav-active'
                      : isAdmin
                      ? 'text-amber-400/70 hover:text-amber-300 hover:bg-white/5 border border-dashed border-amber-500/20'
                      : 'text-white/55 hover:text-white/90 hover:bg-white/6'
                  }`}
                >
                  <span className={`flex-shrink-0 transition-colors ${
                    isActive ? 'text-blue-300' : isAdmin ? 'text-amber-400/60' : 'text-white/35 group-hover:text-white/70'
                  }`}>
                    {item.icon}
                  </span>
                  <span className="flex-1 tracking-tight">{item.labelId}</span>
                  {isAdmin && !isActive && (
                    <span className="text-[9px] font-bold text-amber-400/40 uppercase tracking-widest">
                      sandbox
                    </span>
                  )}
                </Link>
              )
            })}
          </nav>

          {/* ── Footer ─────────────────────────────── */}
          <div className="px-4 py-3 border-t border-white/[0.07]">
            <div className={`flex items-center gap-2 px-2.5 py-2 rounded-md ${
              isSandbox
                ? 'bg-red-500/10 border border-red-500/20'
                : 'bg-white/[0.04] border border-white/[0.06]'
            }`}>
              <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${
                isSandbox ? 'bg-red-400 animate-pulse' : 'bg-green-400'
              }`} />
              <div className="min-w-0">
                <p className={`text-[11px] font-semibold ${isSandbox ? 'text-red-400' : 'text-green-400'}`}>
                  {isSandbox ? 'Vulnerable Mode' : 'Secure Mode'}
                </p>
                <p className="text-[10px] text-white/25 truncate">
                  PT. Dana Sejahtera © {new Date().getFullYear()}
                </p>
              </div>
            </div>
          </div>

        </div>
      </aside>
    </>
  )
}
