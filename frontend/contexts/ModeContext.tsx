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

type Mode = 'secure' | 'sandbox'

interface ModeCtx {
  mode: Mode
  isLoading: boolean
  switchMode: (mode: Mode) => Promise<void>
}

const ModeContext = createContext<ModeCtx>({
  mode: 'secure',
  isLoading: true,
  switchMode: async () => {},
})

export const useMode = () => useContext(ModeContext)

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

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

export function ModeProvider({ children }: { children: ReactNode }) {
  const [mode, setMode] = useState<Mode>(getRuntimeMode)
  const [isLoading, setIsLoading] = useState(true)

  // On mount, fetch the authoritative mode from the backend and sync state.
  useEffect(() => {
    const fetchMode = async () => {
      try {
        const res = await fetch(`${API_BASE}/config/mode`)
        if (res.ok) {
          const data = (await res.json()) as { mode: string }
          const resolved: Mode = data.mode === 'secure' ? 'secure' : 'sandbox'
          setRuntimeMode(resolved)
          setMode(resolved)
          syncModeCookie(resolved)
        }
      } catch {
        // Backend unreachable — keep the env-var default set in securityMode.ts
      } finally {
        setIsLoading(false)
      }
    }
    fetchMode()
  }, [])

  // switchMode: calls PUT /config/mode (admin-only), then updates all local state.
  // OWASP A01 — backend enforces admin role; frontend just sends the JWT.
  const switchMode = useCallback(
    async (newMode: Mode) => {
      // Retrieve the token from whichever storage the current mode uses.
      const token =
        getRuntimeMode() === 'sandbox'
          ? localStorage.getItem('auth_token')
          : sessionStorage.getItem('_sess_t')

      if (!token) throw new Error('Not authenticated')

      const res = await fetch(`${API_BASE}/config/mode`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ mode: newMode }),
      })

      if (!res.ok) {
        const err = await res
          .json()
          .catch(() => ({ error: 'unknown error' })) as { error?: string }
        throw new Error(err.error ?? `HTTP ${res.status}`)
      }

      const data = (await res.json()) as { mode: string }
      const resolved: Mode = data.mode === 'secure' ? 'secure' : 'sandbox'

      // Migrate token storage BEFORE updating the module variable so the old
      // storage key is still readable during migration.
      migrateTokenStorage(token, mode, resolved)

      setRuntimeMode(resolved)
      setMode(resolved)
      syncModeCookie(resolved)
    },
    [mode]
  )

  return (
    <ModeContext.Provider value={{ mode, isLoading, switchMode }}>
      {children}
    </ModeContext.Provider>
  )
}
