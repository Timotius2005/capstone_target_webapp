'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Suspense } from 'react'
import ModeBadge from '@/components/ModeBadge'
import { authService } from '@/services/auth'
import { isVulnerable } from '@/utils/securityMode'

interface DebugInfo {
  token?: string
  userId?: string
  role?: string
  storageKey?: string
  storageType?: string
}

function LoginForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const nextPath = searchParams.get('next') || '/dashboard'

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [debugPanel, setDebugPanel] = useState<DebugInfo | null>(null)

  const vulnerable = isVulnerable()

  useEffect(() => {
    // Secure: redirect immediately if already logged in
    if (!vulnerable && authService.isAuthenticated()) {
      router.replace('/dashboard')
    }
    // TODO: Vulnerability Injection Point
    // Vulnerable: no redirect check — allows multiple sessions to accumulate
  }, [router, vulnerable])

  const validate = (): boolean => {
    if (vulnerable) {
      // TODO: Vulnerability Injection Point
      // Validation completely disabled in vulnerable mode
      return true
    }
    if (!username.trim() || username.length < 3) {
      setError('Username minimal 3 karakter.')
      return false
    }
    if (!password || password.length < 6) {
      setError('Password minimal 6 karakter.')
      return false
    }
    return true
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setDebugPanel(null)
    if (!validate()) return

    setLoading(true)
    try {
      const data = await authService.login(username, password)

      if (vulnerable) {
        // TODO: Vulnerability Injection Point
        // Full login response rendered in the UI (JWT, internal IDs)
        setDebugPanel({
          token: data.token,
          userId: data.user.id,
          role: data.user.role,
          storageKey: 'auth_token',
          storageType: 'localStorage (XSS accessible)',
        })
        setTimeout(() => router.push(nextPath), 2500)
      } else {
        router.push(nextPath)
      }
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string }; status?: number } }

      if (vulnerable) {
        // TODO: Vulnerability Injection Point
        // Verbose error from server exposed — reveals valid usernames via timing/message
        const serverMsg = axiosErr.response?.data?.error
        const status = axiosErr.response?.status
        setError(serverMsg || `HTTP ${status}: Authentication failed`)
      } else {
        // Secure: always generic message
        setError('Username atau password salah. Silakan coba lagi.')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex">
      {/* ── Left panel ────────────────────────────────── */}
      <div className="hidden lg:flex lg:w-5/12 xl:w-1/2 animated-gradient flex-col justify-between p-12 relative overflow-hidden">
        {/* Decorative blobs */}
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute -top-48 -right-48 w-96 h-96 bg-white/5 rounded-full blur-3xl" />
          <div className="absolute -bottom-48 -left-32 w-80 h-80 bg-white/5 rounded-full blur-2xl" />
          <div className="absolute top-1/3 right-1/4 w-64 h-64 bg-white/3 rounded-full blur-3xl" />
        </div>

        {/* Top branding */}
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-8">
            <div className="w-12 h-12 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center text-white font-bold text-xl border border-white/30 shadow-glass">
              DS
            </div>
            <div>
              <p className="text-white font-bold text-xl leading-tight">PT. Dana Sejahtera</p>
              <p className="text-white/60 text-sm">Fintech Loan Management System</p>
            </div>
          </div>
          <ModeBadge className="!bg-white/10 !border-white/25 !text-white" />
        </div>

        {/* Center quote */}
        <div className="relative z-10">
          <div className="text-5xl mb-6 opacity-60">"</div>
          <p className="text-white/90 text-xl font-light italic leading-relaxed mb-4">
            Empowering financial futures through trusted and transparent lending solutions.
          </p>
          <p className="text-white/40 text-sm">— PT. Dana Sejahtera, Est. 2016</p>
        </div>

        {/* Stats row */}
        <div className="relative z-10 grid grid-cols-3 gap-6 pt-6 border-t border-white/15">
          {[
            { label: 'Nasabah Aktif', value: '12.4K+' },
            { label: 'Portfolio', value: 'Rp 2.4T' },
            { label: 'Tahun Beroperasi', value: '8+' },
          ].map((s) => (
            <div key={s.label}>
              <p className="text-white font-bold text-2xl">{s.value}</p>
              <p className="text-white/50 text-xs mt-0.5">{s.label}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── Right panel (form) ─────────────────────── */}
      <div className="flex-1 flex items-center justify-center p-6 sm:p-10 bg-slate-50 dark:bg-[#080d1a]">
        <div className="w-full max-w-[400px]">
          {/* Mobile logo */}
          <div className="lg:hidden flex items-center gap-3 mb-8 justify-center">
            <div className="w-12 h-12 animated-gradient rounded-2xl flex items-center justify-center text-white font-bold text-xl shadow-glow">
              DS
            </div>
            <p className="font-bold text-xl text-slate-800 dark:text-white">PT. Dana Sejahtera</p>
          </div>

          {/* Card */}
          <div className="glass-card rounded-3xl p-8 shadow-glass-lg">
            <div className="mb-7">
              <h2 className="text-2xl font-bold text-slate-800 dark:text-white mb-1.5">
                Selamat Datang 👋
              </h2>
              <p className="text-slate-400 text-sm">Masuk ke sistem manajemen pinjaman</p>
              <div className="mt-3 lg:hidden">
                <ModeBadge size="sm" />
              </div>
            </div>

            {/* Vulnerable warning banner */}
            {vulnerable && (
              <div className="mb-5 p-3.5 rounded-xl border border-red-500/30 bg-red-500/8">
                <div className="flex gap-2.5">
                  <span className="text-red-400 text-base flex-shrink-0 mt-0.5">⚠</span>
                  <div>
                    <p className="text-red-400 text-xs font-semibold">
                      Vulnerable Mode — Security Disabled
                    </p>
                    <p className="text-red-300/60 text-xs mt-0.5 leading-relaxed">
                      Form validation off · JWT → localStorage · Full errors exposed · See DevTools Console
                    </p>
                    {/* TODO: Vulnerability Injection Point */}
                  </div>
                </div>
              </div>
            )}

            <form onSubmit={handleSubmit} noValidate={vulnerable} className="space-y-4">
              {/* Username */}
              <div>
                <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1.5 uppercase tracking-wider">
                  Username
                </label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required={!vulnerable}
                  minLength={vulnerable ? undefined : 3}
                  placeholder="Masukkan username"
                  autoComplete="username"
                  className="input-field"
                />
              </div>

              {/* Password */}
              <div>
                <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1.5 uppercase tracking-wider">
                  Password
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required={!vulnerable}
                  minLength={vulnerable ? undefined : 6}
                  placeholder="Masukkan password"
                  autoComplete="current-password"
                  className="input-field"
                />
              </div>

              {/* Error */}
              {error && (
                <div
                  className={`p-3 rounded-xl text-xs font-medium border ${
                    vulnerable
                      ? 'border-red-500/40 bg-red-500/10 text-red-400 font-mono'
                      : 'border-red-200 bg-red-50 dark:border-red-800/40 dark:bg-red-900/20 text-red-600 dark:text-red-400'
                  }`}
                >
                  {/* TODO: Vulnerability Injection Point */}
                  {/* Vulnerable: server error message verbatim; Secure: generic */}
                  {error}
                </div>
              )}

              {/* Submit */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 animated-gradient text-white font-semibold rounded-xl shadow-lg hover:shadow-glow hover:scale-[1.015] active:scale-[0.985] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 disabled:hover:shadow-lg mt-2"
              >
                {loading ? (
                  <span className="flex items-center justify-center gap-2">
                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Memproses...
                  </span>
                ) : (
                  'Masuk ke Sistem'
                )}
              </button>
            </form>

            {/* TODO: Vulnerability Injection Point */}
            {/* Debug panel — JWT and internal IDs rendered in UI in vulnerable mode */}
            {vulnerable && debugPanel && (
              <div className="mt-5 p-4 rounded-xl border border-red-500/30 bg-red-500/5 animate-fade-in">
                <div className="flex items-center gap-2 mb-3">
                  <span className="w-2 h-2 bg-red-400 rounded-full animate-pulse" />
                  <span className="text-red-400 text-[10px] font-semibold uppercase tracking-widest">
                    Debug: Login Response [Vulnerability Injection Point]
                  </span>
                </div>
                <div className="space-y-1.5">
                  {Object.entries(debugPanel).map(([k, v]) => (
                    <div key={k} className="flex gap-2 text-xs font-mono">
                      <span className="text-red-400/60 flex-shrink-0 w-24">{k}:</span>
                      <span className="text-red-300/80 break-all">{v}</span>
                    </div>
                  ))}
                </div>
                <p className="text-red-300/40 text-[10px] mt-3">
                  Redirecting to dashboard in 2.5s...
                </p>
              </div>
            )}
          </div>

          <p className="text-center text-[11px] text-slate-400 mt-5">
            PT. Dana Sejahtera © {new Date().getFullYear()} · All rights reserved
          </p>
        </div>
      </div>
    </div>
  )
}

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  )
}
