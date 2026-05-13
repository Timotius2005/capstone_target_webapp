'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import DashboardLayout from '@/components/DashboardLayout'
import { api } from '@/services/api'
import { isVulnerable } from '@/utils/securityMode'

interface UserRecord {
  id: string
  username: string
  email: string
  role: string
  created_at: string
}

// TODO: Vulnerability Injection Point
// This entire page is only visible/accessible in vulnerable mode
// In secure mode, it redirects immediately
export default function AdminPage() {
  const router = useRouter()
  const vulnerable = isVulnerable()
  const [users, setUsers] = useState<UserRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [rawConfig, setRawConfig] = useState<Record<string, unknown> | null>(null)
  const [activeTab, setActiveTab] = useState<'users' | 'system' | 'tokens'>('users')
  const [storedToken, setStoredToken] = useState<string | null>(null)
  const [storedUser, setStoredUser] = useState<string | null>(null)
  const [cookies, setCookies] = useState<string>('')

  useEffect(() => {
    // Secure mode: redirect immediately — admin page is hidden
    if (!vulnerable) {
      router.replace('/dashboard')
      return
    }

    // TODO: Vulnerability Injection Point
    // Vulnerable: admin UI freely accessible
    // Reads all stored credentials and exposes them in the UI
    const token = localStorage.getItem('auth_token')
    const user = localStorage.getItem('auth_user')
    setStoredToken(token)
    setStoredUser(user)
    setCookies(document.cookie)

    console.log('[🔴 VULNERABLE] Admin page accessed')
    console.log('[🔴 VULNERABLE] localStorage auth_token:', token)
    console.log('[🔴 VULNERABLE] document.cookie:', document.cookie)

    const fetchUsers = async () => {
      try {
        const res = await api.get('/api/v1/users/profile')
        // Wrap single profile in array for display
        setUsers([res.data as UserRecord])
      } catch {
        // Try alternate endpoint
      } finally {
        setLoading(false)
      }
    }

    // TODO: Vulnerability Injection Point
    // Exposing simulated system config — would never appear in secure mode
    setRawConfig({
      database_url: 'postgresql://admin:s3cr3t@localhost:5432/dana_sejahtera',
      jwt_secret: '[retrieved from environment]',
      debug_mode: true,
      log_level: 'debug',
      cors_origins: ['*'],
      rate_limiting: false,
      auth_bypass_enabled: true,
    })

    fetchUsers()
  }, [router, vulnerable])

  if (!vulnerable) {
    return null // Redirect is in progress
  }

  const TABS = [
    { key: 'users' as const, label: '👤 Manajemen User', danger: false },
    { key: 'system' as const, label: '⚙️ Konfigurasi Sistem', danger: true },
    { key: 'tokens' as const, label: '🔑 Token & Session', danger: true },
  ]

  return (
    <DashboardLayout title="Admin Settings">
      {/* ── Warning banner ──────────────────────── */}
      <div className="mb-6 p-4 rounded-2xl border border-red-500/40 bg-red-500/8 animate-fade-in">
        <div className="flex gap-3">
          <div className="text-2xl flex-shrink-0">⚠️</div>
          <div>
            <p className="text-red-400 font-bold">Halaman Admin — Vulnerable Mode Only</p>
            <p className="text-red-300/70 text-sm mt-1 leading-relaxed">
              Halaman ini hanya muncul dalam <strong>Vulnerable Mode</strong>. Dalam Secure Mode,
              halaman ini tidak ada dalam navigasi dan router langsung redirect ke{' '}
              <code className="text-red-300 font-mono text-xs">/dashboard</code>.
              <br />
              <span className="text-red-400/60 text-xs mt-1 block">
                [Vulnerability Injection Points]: Exposed credentials · Hidden admin UI ·
                System config leak · Token display · No RBAC enforcement
              </span>
            </p>
          </div>
        </div>
      </div>

      {/* ── Tabs ────────────────────────────────── */}
      <div className="flex gap-1 p-1 glass-card rounded-xl w-fit mb-6">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all duration-200 ${
              activeTab === tab.key
                ? tab.danger
                  ? 'bg-red-600 text-white shadow'
                  : 'bg-indigo-600 text-white shadow'
                : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-white'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* ── TAB: User Management ─────────────────── */}
      {activeTab === 'users' && (
        <div className="space-y-4 animate-fade-in">
          <div className="glass-card rounded-2xl p-6">
            <h3 className="text-base font-bold text-slate-800 dark:text-white mb-4 flex items-center gap-2">
              Profil Pengguna Aktif
              <span className="text-red-400 text-xs font-mono font-normal">
                [VULN: No RBAC — any user can view this]
              </span>
            </h3>
            {loading ? (
              <div className="h-20 bg-slate-200 dark:bg-slate-700/50 rounded-xl animate-pulse" />
            ) : users.length > 0 ? (
              <div className="space-y-3">
                {users.map((u) => (
                  <div
                    key={u.id}
                    className="p-4 rounded-xl bg-slate-100/50 dark:bg-slate-800/30 border border-slate-200/30 dark:border-slate-700/30"
                  >
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-sm">
                      {[
                        { label: 'UUID', value: u.id, mono: true, sensitive: true },
                        { label: 'Username', value: u.username, mono: false, sensitive: false },
                        { label: 'Email', value: u.email, mono: false, sensitive: true },
                        { label: 'Role', value: u.role, mono: false, sensitive: false },
                        { label: 'Created', value: new Date(u.created_at).toLocaleDateString('id-ID'), mono: false, sensitive: false },
                      ].map((field) => (
                        <div key={field.label}>
                          <p className="text-xs text-slate-400 uppercase tracking-wider mb-0.5">{field.label}</p>
                          <p
                            className={`font-${field.mono ? 'mono' : 'medium'} text-xs break-all ${
                              field.sensitive
                                ? 'text-red-400'
                                : 'text-slate-800 dark:text-white'
                            }`}
                          >
                            {/* TODO: Vulnerability Injection Point — internal IDs and email exposed */}
                            {field.value || '—'}
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-slate-400 text-sm">Tidak ada data pengguna.</p>
            )}
          </div>
        </div>
      )}

      {/* ── TAB: System Config ───────────────────── */}
      {activeTab === 'system' && (
        <div className="space-y-4 animate-fade-in">
          {/* TODO: Vulnerability Injection Point */}
          {/* System config including DB credentials exposed in the UI */}
          <div className="glass-card rounded-2xl p-6 border border-red-500/30">
            <h3 className="text-base font-bold text-red-400 mb-1 flex items-center gap-2">
              ⚠ Konfigurasi Sistem
              <span className="text-red-400/50 text-xs font-mono font-normal">
                [VULN: Config Exposure]
              </span>
            </h3>
            <p className="text-xs text-red-300/60 mb-4">
              Dalam sistem nyata, data berikut TIDAK BOLEH ditampilkan di frontend.
            </p>
            {rawConfig && (
              <div className="space-y-2">
                {Object.entries(rawConfig).map(([k, v]) => (
                  <div
                    key={k}
                    className="flex items-start gap-3 p-3 rounded-lg bg-slate-900/50 dark:bg-black/30"
                  >
                    <span className="font-mono text-xs text-red-400/70 flex-shrink-0 w-40">{k}:</span>
                    <span
                      className={`font-mono text-xs break-all ${
                        typeof v === 'boolean'
                          ? v
                            ? 'text-red-400'
                            : 'text-emerald-400'
                          : 'text-amber-400'
                      }`}
                    >
                      {String(v)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Quick actions */}
          <div className="glass-card rounded-2xl p-6">
            <h3 className="text-base font-bold text-slate-800 dark:text-white mb-4">
              Aksi Sistem
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {[
                { label: 'Reset Semua Token', desc: 'Invalidasi semua JWT aktif', danger: true },
                { label: 'Flush Cache', desc: 'Bersihkan cache aplikasi', danger: false },
                { label: 'Export Database', desc: 'Download dump database penuh', danger: true },
                { label: 'Toggle Debug Mode', desc: 'Aktifkan/nonaktifkan debug logging', danger: false },
              ].map((action) => (
                <button
                  key={action.label}
                  onClick={() => alert(`[DEMO] Action: ${action.label}`)}
                  className={`p-4 rounded-xl text-left border transition-all duration-200 ${
                    action.danger
                      ? 'border-red-500/30 bg-red-500/5 hover:bg-red-500/10 hover:border-red-500/50'
                      : 'border-slate-200/30 dark:border-slate-700/30 hover:bg-slate-100/50 dark:hover:bg-slate-800/30'
                  }`}
                >
                  <p className={`text-sm font-semibold ${action.danger ? 'text-red-400' : 'text-slate-800 dark:text-white'}`}>
                    {action.label}
                  </p>
                  <p className="text-xs text-slate-400 mt-0.5">{action.desc}</p>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* ── TAB: Tokens & Session ────────────────── */}
      {activeTab === 'tokens' && (
        <div className="space-y-4 animate-fade-in">
          {/* TODO: Vulnerability Injection Point */}
          {/* JWT and session cookie displayed verbatim in the UI */}
          <div className="glass-card rounded-2xl p-6 border border-red-500/30">
            <h3 className="text-base font-bold text-red-400 mb-4 flex items-center gap-2">
              🔑 Token & Cookie Storage
              <span className="text-red-400/50 text-xs font-mono font-normal">
                [VULN: Token Exposure]
              </span>
            </h3>

            <div className="space-y-4">
              <div>
                <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  localStorage[&apos;auth_token&apos;]
                </p>
                <div className="p-3 bg-slate-900/60 dark:bg-black/40 rounded-xl border border-red-500/20">
                  <p className="font-mono text-xs text-red-300 break-all leading-relaxed">
                    {storedToken || '(kosong — tidak ada token)'}
                  </p>
                </div>
                <p className="text-[10px] text-red-400/60 mt-1">
                  ⚠ JWT ini dapat dicuri via XSS dan digunakan untuk impersonasi
                </p>
              </div>

              <div>
                <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  localStorage[&apos;auth_user&apos;]
                </p>
                <div className="p-3 bg-slate-900/60 dark:bg-black/40 rounded-xl border border-red-500/20">
                  <pre className="font-mono text-xs text-amber-300 break-all leading-relaxed overflow-x-auto">
                    {storedUser
                      ? JSON.stringify(JSON.parse(storedUser), null, 2)
                      : '(kosong)'}
                  </pre>
                </div>
              </div>

              <div>
                <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  document.cookie
                </p>
                <div className="p-3 bg-slate-900/60 dark:bg-black/40 rounded-xl border border-red-500/20">
                  <p className="font-mono text-xs text-yellow-300 break-all leading-relaxed">
                    {cookies || '(tidak ada cookie)'}
                  </p>
                </div>
                <p className="text-[10px] text-red-400/60 mt-1">
                  ⚠ Cookie tanpa flag HttpOnly — dapat diakses oleh JavaScript
                </p>
              </div>
            </div>
          </div>

          {/* Mitigation guide */}
          <div className="glass-card rounded-2xl p-6 border border-emerald-500/20">
            <h3 className="text-sm font-bold text-emerald-400 mb-3 flex items-center gap-2">
              🔒 Mitigasi (Secure Mode)
            </h3>
            <ul className="space-y-2 text-xs text-slate-400 list-none">
              {[
                'JWT disimpan di sessionStorage dengan key terobfuskasi (bukan localStorage)',
                'Cookie menggunakan flag HttpOnly + SameSite=Strict',
                'Halaman Admin ini tidak ada di navigasi dan route diblokir middleware',
                'Token tidak pernah di-log ke console',
                'Error message selalu generik (tidak mengekspos detail server)',
                'Input divalidasi ketat sebelum dikirim ke API',
              ].map((item, i) => (
                <li key={i} className="flex items-start gap-2">
                  <span className="text-emerald-400 mt-0.5 flex-shrink-0">✓</span>
                  {item}
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </DashboardLayout>
  )
}
