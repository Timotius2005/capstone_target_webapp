'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import ModeBadge from './ModeBadge'
import { useTheme } from './ThemeProvider'
import { authService, type User } from '@/services/auth'
import { isVulnerable } from '@/utils/securityMode'

interface NavbarProps {
  title: string
  onMenuToggle: () => void
}

export default function Navbar({ title, onMenuToggle }: NavbarProps) {
  const { theme, toggleTheme } = useTheme()
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const [user, setUser] = useState<Partial<User> | null>(null)
  const router = useRouter()
  const vulnerable = isVulnerable()

  useEffect(() => {
    setUser(authService.getUser())
  }, [])

  const handleLogout = () => {
    authService.logout()
    router.push('/login')
  }

  const initials = user?.username?.slice(0, 2).toUpperCase() || 'U'

  return (
    <header className="sticky top-0 z-20 glass-card border-b border-slate-200/20 dark:border-slate-700/30 px-4 sm:px-6 py-3.5">
      <div className="flex items-center justify-between gap-4">
        {/* ── Left: hamburger + title ─── */}
        <div className="flex items-center gap-3 min-w-0">
          <button
            onClick={onMenuToggle}
            className="p-2 rounded-lg text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            aria-label="Toggle sidebar"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>

          <div className="min-w-0">
            <h1 className="text-base font-bold text-slate-800 dark:text-white truncate">{title}</h1>
            <p className="text-xs text-slate-400 hidden sm:block">
              {new Date().toLocaleDateString('id-ID', {
                weekday: 'long',
                day: 'numeric',
                month: 'long',
                year: 'numeric',
              })}
            </p>
          </div>
        </div>

        {/* ── Right: badge + controls ─── */}
        <div className="flex items-center gap-2 sm:gap-3 flex-shrink-0">
          <ModeBadge size="sm" className="hidden sm:inline-flex" />

          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            className="p-2 rounded-xl text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200"
            title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
          >
            {theme === 'dark' ? (
              <svg className="w-4.5 h-4.5 w-[18px] h-[18px]" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Z" />
              </svg>
            ) : (
              <svg className="w-[18px] h-[18px]" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
              </svg>
            )}
          </button>

          {/* User menu */}
          <div className="relative">
            <button
              onClick={() => setUserMenuOpen((v) => !v)}
              className="flex items-center gap-2 px-2.5 py-1.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200"
            >
              <div className="w-8 h-8 rounded-full animated-gradient flex items-center justify-center text-white text-xs font-bold shadow flex-shrink-0">
                {initials}
              </div>
              <div className="hidden sm:block text-left">
                <p className="text-xs font-semibold text-slate-700 dark:text-white leading-tight">
                  {user?.username || 'User'}
                </p>
                <p className="text-[10px] text-slate-400 capitalize">{user?.role || 'member'}</p>
              </div>
              <svg className="w-3.5 h-3.5 text-slate-400 hidden sm:block" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="m19 9-7 7-7-7" />
              </svg>
            </button>

            {userMenuOpen && (
              <>
                <div
                  className="fixed inset-0 z-10"
                  onClick={() => setUserMenuOpen(false)}
                />
                <div className="absolute right-0 top-full mt-2 w-56 glass-card rounded-xl border border-slate-200/30 dark:border-slate-700/50 shadow-glass-lg py-1.5 z-20 animate-fade-in">
                  <div className="px-4 py-3 border-b border-slate-200/20 dark:border-slate-700/30">
                    <p className="text-sm font-semibold text-slate-800 dark:text-white">
                      {user?.username || 'User'}
                    </p>
                    {/* TODO: Vulnerability Injection Point */}
                    {/* Vulnerable mode exposes internal user ID and email */}
                    {vulnerable && user?.email && (
                      <p className="text-xs text-red-400/80 mt-0.5 font-mono break-all">
                        {user.email}
                      </p>
                    )}
                    {vulnerable && (user as User)?.id && (
                      <p className="text-[10px] text-red-400/60 mt-0.5 font-mono truncate">
                        ID: {(user as User).id}
                      </p>
                    )}
                    {!vulnerable && (
                      <p className="text-xs text-slate-400 capitalize mt-0.5">{user?.role}</p>
                    )}
                  </div>
                  <ModeBadge size="sm" className="mx-3 my-2" />
                  <button
                    onClick={handleLogout}
                    className="w-full flex items-center gap-2.5 px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 transition-colors"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
                    </svg>
                    Logout
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
