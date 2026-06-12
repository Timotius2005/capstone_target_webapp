'use client'

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import Cookies from 'js-cookie'

import {
  getRuntimeMode,
  setRuntimeMode,
} from '@/utils/securityMode'

// ── Types ─────────────────────────────────────────────────────────────────────

type Mode = 'secure' | 'sandbox'

/**
 * Per-category OWASP Top 10 vulnerability configuration.
 * Only applied when mode === 'sandbox'. In secure mode every category is
 * always treated as disabled regardless of this object.
 */
export interface VulnConfig {
  A01_BrokenAccessControl: boolean
  A02_CryptographicFailures: boolean
  A03_Injection: boolean
  A04_InsecureDesign: boolean
  A05_SecurityMisconfiguration: boolean
  A06_VulnerableComponents: boolean
  A07_AuthenticationFailures: boolean
  A08_SoftwareDataIntegrityFailures: boolean
  A09_SecurityLoggingFailures: boolean
  A10_SSRF: boolean
}

/** All categories enabled — identical to the original all-or-nothing vulnerable mode. */
export const defaultVulnConfig: VulnConfig = {
  A01_BrokenAccessControl:           true,
  A02_CryptographicFailures:         true,
  A03_Injection:                     true,
  A04_InsecureDesign:                true,
  A05_SecurityMisconfiguration:      true,
  A06_VulnerableComponents:          true,
  A07_AuthenticationFailures:        true,
  A08_SoftwareDataIntegrityFailures: true,
  A09_SecurityLoggingFailures:       true,
  A10_SSRF:                          true,
}

interface ModeCtx {
  mode: Mode
  isLoading: boolean
  vulnConfig: VulnConfig
  isVulnConfigLoading: boolean
  switchMode: (mode: Mode) => Promise<void>
  updateVulnConfig: (config: VulnConfig) => Promise<void>
}

export const ModeContext = createContext<ModeCtx>({
  mode:                'secure',
  isLoading:           true,
  vulnConfig:          defaultVulnConfig,
  isVulnConfigLoading: false,
  switchMode:          async () => {},
  updateVulnConfig:    async () => {},
})

export const useMode = () => useContext(ModeContext)

const API_BASE = process.env.NEXT_PUBLIC_API_URL || ''

// Store a cookie so Next.js middleware (Edge Runtime) can read the current mode
// without an extra backend fetch on every request.
function syncModeCookie(m: Mode) {
  Cookies.set('_app_mode', m, { sameSite: 'Lax', expires: 365 })
}

// After a mode switch the JWT may be in the wrong storage location.
// Migrate it so subsequent API calls stay authenticated.
function migrateTokenStorage(token: string, from: Mode, to: Mode) {
  if (typeof window === 'undefined') return

  if (from === 'sandbox' && to === 'secure') {
    // Move from localStorage → sessionStorage
    const rawUser = localStorage.getItem('auth_user')
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    sessionStorage.setItem('_sess_t', token)
    if (rawUser) {
      try {
        const u = JSON.parse(rawUser) as { username?: string; role?: string }
        sessionStorage.setItem(
          '_sess_u',
          JSON.stringify({ username: u.username, role: u.role })
        )
      } catch {
        // malformed stored user — ignore
      }
    }
    Cookies.set('auth-session', '1', { sameSite: 'Strict' })
  } else if (from === 'secure' && to === 'sandbox') {
    // Move from sessionStorage → localStorage
    const rawUser = sessionStorage.getItem('_sess_u')
    sessionStorage.removeItem('_sess_t')
    sessionStorage.removeItem('_sess_u')
    localStorage.setItem('auth_token', token)
    if (rawUser) {
      localStorage.setItem('auth_user', rawUser)
    }
    Cookies.set('auth-session', token, { expires: 7, sameSite: 'Lax' })
  }
}

// ── Provider ──────────────────────────────────────────────────────────────────

export function ModeProvider({ children }: { children: ReactNode }) {
  const [mode, setMode]                             = useState<Mode>(getRuntimeMode)
  const [isLoading, setIsLoading]                   = useState(true)
  const [vulnConfig, setVulnConfig]                 = useState<VulnConfig>(defaultVulnConfig)
  const [isVulnConfigLoading, setIsVulnConfigLoading] = useState(false)

  // ── Mount: fetch authoritative mode + vuln config from backend ─────────────
  useEffect(() => {
    const fetchMode = async () => {
      try {
        const headers: Record<string, string> = { 'ngrok-skip-browser-warning': 'true' }
        const labKey = process.env.NEXT_PUBLIC_LAB_KEY
        if (labKey) headers['X-LAB-KEY'] = labKey

        const res = await fetch(`${API_BASE}/api/system/mode`, { headers })
        if (res.ok) {
          const data = (await res.json()) as { mode: string }
          const resolved: Mode = data.mode === 'secure' ? 'secure' : 'sandbox'
          setRuntimeMode(resolved)
          setMode(resolved)
          syncModeCookie(resolved)

          // If we started in vulnerable mode, also load the current vuln config.
          if (resolved === 'sandbox') {
            fetchVulnConfigSilent(headers)
          }
        }
      } catch {
        // Backend unreachable — keep the env-var default set in securityMode.ts
      } finally {
        setIsLoading(false)
      }
    }
    fetchMode()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Internal helper: loads vuln config without touching isLoading.
  const fetchVulnConfigSilent = useCallback(
    async (headers: Record<string, string> = {}) => {
      try {
        headers['ngrok-skip-browser-warning'] = 'true'
        const labKey = process.env.NEXT_PUBLIC_LAB_KEY
        if (labKey) headers['X-LAB-KEY'] = labKey
        const res = await fetch(`${API_BASE}/api/system/vuln-config`, { headers })
        if (res.ok) {
          const data = (await res.json()) as { config: VulnConfig }
          setVulnConfig(data.config)
        }
      } catch {
        // ignore — keep current config
      }
    },
    []
  )

  // ── switchMode ─────────────────────────────────────────────────────────────
  const switchMode = useCallback(
    async (newMode: Mode) => {
      const headers: Record<string, string> = { 'Content-Type': 'application/json', 'ngrok-skip-browser-warning': 'true' }
      const labKey = process.env.NEXT_PUBLIC_LAB_KEY
      if (labKey) headers['X-LAB-KEY'] = labKey

      // Map internal 'sandbox' alias to the public API contract value 'vulnerable'.
      const apiMode = newMode === 'sandbox' ? 'vulnerable' : 'secure'

      const res = await fetch(`${API_BASE}/api/system/mode`, {
        method:  'PUT',
        headers,
        body:    JSON.stringify({ mode: apiMode }),
      })

      if (!res.ok) {
        const err = await res
          .json()
          .catch(() => ({ error: 'unknown error' })) as { error?: string }
        throw new Error(err.error ?? `HTTP ${res.status}`)
      }

      const data = (await res.json()) as { mode: string }
      const resolved: Mode = data.mode === 'secure' ? 'secure' : 'sandbox'

      // Migrate JWT storage if the user is currently authenticated.
      const token =
        getRuntimeMode() === 'sandbox'
          ? localStorage.getItem('auth_token')
          : sessionStorage.getItem('_sess_t')
      if (token) {
        migrateTokenStorage(token, mode, resolved)
      }

      setRuntimeMode(resolved)
      setMode(resolved)
      syncModeCookie(resolved)

      // When switching to vulnerable mode, fetch the current config.
      // When switching back to secure, reset config to defaults so the next
      // vulnerable session starts clean.
      if (resolved === 'sandbox') {
        fetchVulnConfigSilent()
      } else {
        setVulnConfig(defaultVulnConfig)
      }
    },
    [mode, fetchVulnConfigSilent]
  )

  // ── updateVulnConfig ───────────────────────────────────────────────────────
  // Sends the new config to the backend and updates local state on success.
  // Only callable when mode === 'sandbox' (backend also enforces this).
  const updateVulnConfig = useCallback(
    async (config: VulnConfig) => {
      setIsVulnConfigLoading(true)
      try {
        const headers: Record<string, string> = { 'Content-Type': 'application/json', 'ngrok-skip-browser-warning': 'true' }
        const labKey = process.env.NEXT_PUBLIC_LAB_KEY
        if (labKey) headers['X-LAB-KEY'] = labKey

        const res = await fetch(`${API_BASE}/api/system/vuln-config`, {
          method:  'PUT',
          headers,
          body:    JSON.stringify(config),
        })

        if (!res.ok) {
          const err = await res
            .json()
            .catch(() => ({ error: 'unknown error' })) as { error?: string }
          throw new Error(err.error ?? `HTTP ${res.status}`)
        }

        const data = (await res.json()) as { config: VulnConfig }
        setVulnConfig(data.config)
      } finally {
        setIsVulnConfigLoading(false)
      }
    },
    []
  )

  return (
    <ModeContext.Provider
      value={{ mode, isLoading, vulnConfig, isVulnConfigLoading, switchMode, updateVulnConfig }}
    >
      {children}
    </ModeContext.Provider>
  )
}
