'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Suspense } from 'react'
import ModeBadge from '@/components/ModeBadge'
import { authService } from '@/services/auth'
import { useMode } from '@/contexts/ModeContext'

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

  const { mode } = useMode()
  const vulnerable = mode === 'sandbox'

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

      {/* ── Left panel — solid navy branding ─────────────────────── */}
      <div className="hidden lg:flex lg:w-5/12 xl:w-[42%] bg-[#0B1E3D] flex-col justify-between p-12 relative overflow-hidden">

        {/* Subtle geometric texture */}
        <div className="absolute inset-0 pointer-events-none opacity-[0.03]"
          style={{
            backgroundImage: 'repeating-linear-gradient(0deg, #fff 0px, #fff 1px, transparent 1px, transparent 40px), repeating-linear-gradient(90deg, #fff 0px, #fff 1px, transparent 1px, transparent 40px)',
          }}
        />

        {/* Top branding */}
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-8">
            <div className="w-11 h-11 bg-[#1E3A8A] rounded-lg flex items-center justify-center text-white font-bold text-lg border border-white/10">
              DS
            </div>
            <div>
              <p className="text-white font-bold text-xl leading-tight tracking-tight">
                PT. Dana Sejahtera
              </p>
              <p className="text-white/40 text-xs tracking-wide mt-0.5">
                Fintech Loan Management System
              </p>
            </div>
          </div>
          <div>
            <ModeBadge className="!bg-white/8 !border-white/15 !text-white/80" />
          </div>
        </div>

        {/* Center copy */}
        <div className="relative z-10">
          <div className="w-10 h-0.5 bg-white/20 mb-6" />
          <p className="text-white/75 text-xl font-light leading-relaxed mb-4 tracking-tight">
            Empowering financial futures through trusted and transparent lending solutions.
          </p>
          <p className="text-white/30 text-sm">— PT. Dana Sejahtera, Est. 2016</p>
        </div>

        {/* Stats row */}
        <div className="relative z-10 grid grid-cols-3 gap-4 pt-6 border-t border-white/10">
          {[
            { label: 'Nasabah Aktif', value: '12.4K+' },
            { label: 'Portfolio', value: 'Rp 2.4T' },
            { label: 'Tahun Beroperasi', value: '8+' },
          ].map((s) => (
            <div key={s.label}>
              <p className="text-white font-bold text-xl tracking-tight">{s.value}</p>
              <p className="text-white/35 text-xs mt-0.5">{s.label}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── Right panel — login form ──────────────────────────────── */}
      <div className="flex-1 flex items-center justify-center p-6 sm:p-10 bg-[#EEF2F7] dark:bg-slate-950">
        <div className="w-full max-w-sm">

          {/* Mobile logo */}
          <div className="lg:hidden flex items-center gap-3 mb-8 justify-center">
            <div className="w-10 h-10 bg-[#1E3A8A] rounded-lg flex items-center justify-center text-white font-bold text-base">
              DS
            </div>
            <p className="font-bold text-lg text-slate-800 dark:text-white tracking-tight">
              PT. Dana Sejahtera
            </p>
          </div>

          {/* Card */}
          <div className="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 shadow-card p-8">
            <div className="mb-6">
              <h2 className="text-xl font-bold text-slate-800 dark:text-white mb-1 tracking-tight">
                System Login
              </h2>
              <p className="text-slate-400 text-sm">
                Masuk ke sistem manajemen pinjaman
              </p>
              <div className="mt-3 lg:hidden">
                <ModeBadge size="sm" />
              </div>
            </div>

            {/* Vulnerable warning banner */}
            {vulnerable && (
              <div className="mb-5 p-3 rounded-md border-l-4 border-red-500 bg-red-50 dark:bg-red-950/30">
                <div className="flex gap-2">
                  <span className="text-red-600 font-bold text-sm flex-shrink-0">⚠</span>
                  <div>
                    <p className="text-red-700 dark:text-red-400 text-xs font-semibold">
                      Vulnerable Mode — Security Disabled
                    </p>
                    <p className="text-red-600/60 dark:text-red-400/50 text-xs mt-0.5">
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
                  className={`p-3 rounded-md text-xs font-medium border-l-4 ${
                    vulnerable
                      ? 'border-red-500 bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-400 font-mono'
                      : 'border-red-400 bg-red-50 dark:bg-red-950/20 text-red-700 dark:text-red-400'
                  }`}
                >
                  {/* TODO: Vulnerability Injection Point */}
                  {error}
                </div>
              )}

              {/* Submit */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 bg-[#1E3A8A] hover:bg-[#1E40AF] active:bg-[#1D3480]
                           text-white font-semibold rounded-md text-sm
                           transition-colors duration-150 shadow-sm
                           disabled:opacity-50 disabled:cursor-not-allowed mt-1"
              >
                {loading ? (
                  <span className="flex items-center justify-center gap-2">
                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Memproses…
                  </span>
                ) : (
                  'Masuk ke Sistem'
                )}
              </button>
            </form>

            {/* TODO: Vulnerability Injection Point */}
            {/* Debug panel — JWT and internal IDs rendered in UI in vulnerable mode */}
            {vulnerable && debugPanel && (
              <div className="mt-5 p-4 rounded-md border-l-4 border-red-500 bg-red-50 dark:bg-red-950/20 animate-fade-in">
                <div className="flex items-center gap-2 mb-3">
                  <span className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
                  <span className="text-red-700 dark:text-red-400 text-[10px] font-semibold uppercase tracking-widest">
                    Debug: Login Response [Vulnerability Injection Point]
                  </span>
                </div>
                <div className="space-y-1.5">
                  {Object.entries(debugPanel).map(([k, v]) => (
                    <div key={k} className="flex gap-2 text-xs font-mono">
                      <span className="text-red-500/60 flex-shrink-0 w-24">{k}:</span>
                      <span className="text-red-700/80 dark:text-red-300/80 break-all">{v}</span>
                    </div>
                  ))}
                </div>
                <p className="text-red-400/50 text-[10px] mt-3">
                  Redirecting to dashboard in 2.5s...
                </p>
              </div>
            )}
          </div>

          <p className="text-center text-[11px] text-slate-400 mt-5">
            PT. Dana Sejahtera &copy; {new Date().getFullYear()} &middot; Internal System
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
