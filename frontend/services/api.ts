import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import { isVulnerable } from '@/utils/securityMode'

const createApiInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    headers: { 'Content-Type': 'application/json' },
    timeout: 10000,
  })

  instance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    if (typeof window === 'undefined') return config

    let token: string | null = null

    if (isVulnerable()) {
      // TODO: Vulnerability Injection Point
      // JWT read from localStorage — accessible by any JS on page (XSS risk)
      token = localStorage.getItem('auth_token')
      if (token) {
        // TODO: Vulnerability Injection Point
        // Full JWT leaked to browser console on every request
        console.log('[🔴 VULNERABLE] Attaching JWT from localStorage:', token)
        console.log('[🔴 VULNERABLE] Request to:', config.url)
      }
    } else {
      // Secure: sessionStorage is tab-scoped and cleared on tab close
      token = sessionStorage.getItem('_sess_t')
    }

    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  })

  instance.interceptors.response.use(
    (response) => {
      if (isVulnerable()) {
        // TODO: Vulnerability Injection Point
        // Full API response logged — exposes hidden/sensitive fields
        console.log('[🔴 VULNERABLE] Full API Response:', {
          url: response.config.url,
          status: response.status,
          headers: response.headers,
          data: response.data,
        })
      }
      return response
    },
    (error) => {
      if (isVulnerable()) {
        // TODO: Vulnerability Injection Point
        // Detailed error info exposed — reveals API internals & stack traces
        console.error('[🔴 VULNERABLE] API Error Details:', {
          status: error.response?.status,
          data: error.response?.data,
          config: { url: error.config?.url, method: error.config?.method },
        })
      }
      return Promise.reject(error)
    }
  )

  return instance
}

export const api = createApiInstance()
export default api
