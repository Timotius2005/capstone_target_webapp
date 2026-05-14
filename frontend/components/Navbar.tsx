'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import ModeBadge from './ModeBadge'
import { useTheme } from './ThemeProvider'
import { useMode } from '@/contexts/ModeContext'
import { authService, type User } from '@/services/auth'

interface NavbarProps {
  title: string
  onMenuToggle: () => void
}

export default function Navbar({ title, onMenuToggle }: NavbarProps) {
  const { theme, toggleTheme } = useTheme()
  const { mode, switchMode } = useMode()
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const [user, setUser] = useState<Partial<User> | null>(null)
  const [switching, setSwitching] = useState(false)
  const [switchError, setSwitchError] = useState<string | null>(null)
  const router = useRouter()

  const isSandbox = mode === 'sandbox'
  const isAdmin = user?.role === 'admin'

  useEffect(() => {
    setUser(authService.getUser())
  }, [])

  const handleLogout = () => {
    authService.logout()
    router.push('/login')
  }

  const handleModeToggle = async () => {
    if (switching) return
    setSwitching(true)
    setSwitchError(null)
    try {
      await switchMode(isSandbox ? 'secure' : 'sandbox')
    } catch (err) {
      setSwitchError(err instanceof Error ? err.message : 'Failed to switch mode')
      setTimeout(() => setSwitchError(null), 4000)
    } finally {
      setSwitching(false)
    }
  }

  const initials = user?.username?.slice(0, 2).toUpperCase() || 'U'

  return (
    <header className="sticky top-8 z-20 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 px-4 sm:px-6 py-0">
      <div className="flex items-center justify-between h-14 gap-4">

        {/* ── Left: hamburger + breadcrumb ── */}
        <div className="flex items-center gap-3 min-w-0">
          <button
            onClick={onMenuToggle}
            className="p-1.5 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            aria-label="Toggle sidebar"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>

          <div className="h-5 w-px bg-slate-200 dark:bg-slate-700 hidden sm:block" />

          <div className="min-w-0 hidden sm:block">
            <h1 className="text-sm font-semibold text-slate-700 dark:text-slate-200 truncate leading-tight">
              {title}
            </h1>
            <p className="text-[10px] text-slate-400 dark:text-slate-500 mt-px">
              PT. Dana Sejahtera — Sistem Manajemen Pinjaman
            </p>
          </div>
        </div>

        {/* ── Right: controls ── */}
        <div className="flex items-center gap-1 sm:gap-2 flex-shrink-0">

          {/* Mode badge */}
          <ModeBadge size="sm" className="hidden sm:inline-flex" />

          {/* Admin mode toggle */}
          {isAdmin && (
            <div className="hidden sm:flex items-center gap-2 pl-2 border-l border-slate-200 dark:border-slate-700">
              {switchError && (
                <span className="text-[10px] text-red-500 font-medium max-w-[100px] truncate">
                  {switchError}
                </span>
              )}
              <button
                onClick={handleModeToggle}
                disabled={switching}
                title={`Switch to ${isSandbox ? 'secure' : 'vulnerable'} mode`}
                className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors
                            focus:outline-none focus:ring-2 focus:ring-offset-1 disabled:opacity-50 ${
                  isSandbox
                    ? 'bg-red-500 focus:ring-red-400'
                    : 'bg-emerald-600 focus:ring-emerald-500'
                }`}
                aria-label="Toggle security mode"
              >
                <span
                  className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow-sm transition-transform ${
                    isSandbox ? 'translate-x-[18px]' : 'translate-x-[3px]'
                  } ${switching ? 'animate-pulse' : ''}`}
                />
              </button>
              <span className={`text-[10px] font-semibold tracking-wide ${isSandbox ? 'text-red-500' : 'text-emerald-600 dark:text-emerald-400'}`}>
                {switching ? '…' : isSandbox ? 'Vulnerable' : 'Secure'}
              </span>
            </div>
          )}

          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            className="p-1.5 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
          >
            {theme === 'dark' ? (
              <svg className="w-[17px] h-[17px]" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Z" />
              </svg>
            ) : (
              <svg className="w-[17px] h-[17px]" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
              </svg>
            )}
          </button>

          {/* User menu */}
          <div className="relative">
            <button
              onClick={() => setUserMenuOpen((v) => !v)}
              className="flex items-center gap-2 pl-2 border-l border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 rounded-md px-2 py-1 transition-colors"
            >
              <div className="w-7 h-7 rounded-full bg-[#1E3A8A] flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                {initials}
              </div>
              <div className="hidden sm:block text-left">
                <p className="text-xs font-semibold text-slate-700 dark:text-slate-200 leading-tight">
                  {user?.username || 'User'}
                </p>
                <p className="text-[10px] text-slate-400 capitalize">{user?.role || 'member'}</p>
              </div>
              <svg className="w-3.5 h-3.5 text-slate-400 hidden sm:block" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="m19 9-7 7-7-7" />
              </svg>
            </button>

            {userMenuOpen && (
              <>
                <div className="fixed inset-0 z-10" onClick={() => setUserMenuOpen(false)} />
                <div className="absolute right-0 top-full mt-1.5 w-52 bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-700 shadow-card-md py-1 z-20 animate-fade-in">

                  <div className="px-4 py-3 border-b border-slate-100 dark:border-slate-800">
                    <p className="text-sm font-semibold text-slate-800 dark:text-white">
                      {user?.username || 'User'}
                    </p>
                    {/* TODO: Vulnerability Injection Point */}
                    {/* Sandbox mode exposes internal user ID and email */}
                    {isSandbox && user?.email && (
                      <p className="text-xs text-red-500/80 mt-0.5 font-mono break-all">
                        {user.email}
                      </p>
                    )}
                    {isSandbox && (user as User)?.id && (
                      <p className="text-[10px] text-red-500/60 mt-0.5 font-mono truncate">
                        ID: {(user as User).id}
                      </p>
                    )}
                    {!isSandbox && (
                      <p className="text-xs text-slate-400 capitalize mt-0.5">{user?.role}</p>
                    )}
                  </div>

                  <div className="px-3 py-2">
                    <ModeBadge size="sm" />
                  </div>

                  {/* Mobile mode toggle */}
                  {isAdmin && (
                    <div className="sm:hidden px-4 py-2 border-t border-slate-100 dark:border-slate-800">
                      <p className="text-[10px] text-slate-400 uppercase tracking-wider mb-2">Security Mode</p>
                      <button
                        onClick={() => { setUserMenuOpen(false); handleModeToggle() }}
                        disabled={switching}
                        className={`w-full flex items-center justify-between px-3 py-2 rounded-md text-xs font-semibold transition-colors ${
                          isSandbox
                            ? 'bg-red-50 text-red-700 hover:bg-red-100 dark:bg-red-900/20 dark:text-red-400'
                            : 'bg-green-50 text-green-800 hover:bg-green-100 dark:bg-green-900/20 dark:text-green-400'
                        }`}
                      >
                        <span>{isSandbox ? '⚠ Vulnerable' : '✓ Secure'}</span>
                        <span className="text-slate-400">Switch →</span>
                      </button>
                    </div>
                  )}

                  <button
                    onClick={handleLogout}
                    className="w-full flex items-center gap-2.5 px-4 py-2.5 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/15 transition-colors border-t border-slate-100 dark:border-slate-800"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
                    </svg>
                    Sign Out
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
