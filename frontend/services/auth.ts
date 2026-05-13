import Cookies from 'js-cookie'
import { api } from './api'
import { isVulnerable } from '@/utils/securityMode'

export interface User {
  id: string
  username: string
  email: string
  role: string
  created_at?: string
  updated_at?: string
}

export interface LoginResponse {
  token: string
  user: User
}

export const authService = {
  async login(username: string, password: string): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/api/v1/auth/login', {
      username,
      password,
    })
    const { token, user } = response.data

    if (isVulnerable()) {
      // TODO: Vulnerability Injection Point
      // JWT stored in localStorage — persists across browser sessions, XSS accessible
      localStorage.setItem('auth_token', token)
      localStorage.setItem('auth_user', JSON.stringify(user))

      // TODO: Vulnerability Injection Point
      // Full JWT and user object (with internal UUID) exposed to console
      console.log('[🔴 VULNERABLE] ===== LOGIN SUCCESS =====')
      console.log('[🔴 VULNERABLE] JWT Token (stored in localStorage):', token)
      console.log('[🔴 VULNERABLE] Full User Object (incl. internal ID):', user)
      console.log('[🔴 VULNERABLE] User ID:', user.id)
      console.log('[🔴 VULNERABLE] Role:', user.role)

      // Cookie without HttpOnly or Secure flags — JS accessible
      Cookies.set('auth-session', token, { expires: 7, sameSite: 'Lax' })
    } else {
      // Secure: sessionStorage only (tab-scoped, not persistent)
      sessionStorage.setItem('_sess_t', token)
      // Only store non-sensitive display fields — never the UUID
      sessionStorage.setItem(
        '_sess_u',
        JSON.stringify({ username: user.username, role: user.role })
      )
      // Flag cookie only — NOT the actual JWT
      Cookies.set('auth-session', '1', { sameSite: 'Strict' })
    }

    return { token, user }
  },

  logout(): void {
    if (typeof window === 'undefined') return

    if (isVulnerable()) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_user')
    } else {
      sessionStorage.removeItem('_sess_t')
      sessionStorage.removeItem('_sess_u')
    }

    Cookies.remove('auth-session')
  },

  getToken(): string | null {
    if (typeof window === 'undefined') return null
    if (isVulnerable()) return localStorage.getItem('auth_token')
    return sessionStorage.getItem('_sess_t')
  },

  getUser(): Partial<User> | null {
    if (typeof window === 'undefined') return null

    if (isVulnerable()) {
      const raw = localStorage.getItem('auth_user')
      return raw ? (JSON.parse(raw) as User) : null
    }

    const raw = sessionStorage.getItem('_sess_u')
    return raw ? (JSON.parse(raw) as Partial<User>) : null
  },

  isAuthenticated(): boolean {
    return !!this.getToken()
  },
}
