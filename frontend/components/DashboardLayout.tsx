'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Sidebar from './Sidebar'
import Navbar from './Navbar'
import { authService } from '@/services/auth'
import { isVulnerable } from '@/utils/securityMode'

interface DashboardLayoutProps {
  children: React.ReactNode
  title: string
}

export default function DashboardLayout({ children, title }: DashboardLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [mounted, setMounted] = useState(false)
  const router = useRouter()

  useEffect(() => {
    setMounted(true)

    if (isVulnerable()) {
      // TODO: Vulnerability Injection Point
      // Weak auth check — only reads localStorage, no token validation
      // Redirect is delayed and bypassable
      const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null
      if (!token) {
        console.warn('[🔴 VULNERABLE] No auth_token in localStorage — weak redirect in 1500ms...')
        setTimeout(() => router.push('/login'), 1500)
      }
    } else {
      // Secure: immediate strict redirect
      if (!authService.isAuthenticated()) {
        router.push('/login')
      }
    }
  }, [router])

  // Collapse sidebar on small screens by default
  useEffect(() => {
    const mediaQuery = window.matchMedia('(max-width: 1024px)')
    setSidebarOpen(!mediaQuery.matches)

    const handler = (e: MediaQueryListEvent) => setSidebarOpen(!e.matches)
    mediaQuery.addEventListener('change', handler)
    return () => mediaQuery.removeEventListener('change', handler)
  }, [])

  if (!mounted) return null

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-[#080d1a]">
      <Sidebar open={sidebarOpen} />

      <div
        className={`flex flex-col min-h-screen transition-all duration-300 ${
          sidebarOpen ? 'lg:pl-64' : 'pl-0'
        }`}
      >
        <Navbar title={title} onMenuToggle={() => setSidebarOpen((v) => !v)} />

        <main className="flex-1 p-4 sm:p-6 page-enter">
          {children}
        </main>
      </div>
    </div>
  )
}
