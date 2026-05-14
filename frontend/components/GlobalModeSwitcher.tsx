'use client'

import { useState } from 'react'
import { useMode } from '@/contexts/ModeContext'

export default function GlobalModeSwitcher() {
  const { mode, isLoading, switchMode } = useMode()
  const [switching, setSwitching] = useState(false)
  const [toast, setToast] = useState<{ msg: string; ok: boolean } | null>(null)

  const isVulnerable = mode === 'sandbox'

  const handleToggle = async () => {
    if (switching || isLoading) return
    setSwitching(true)
    try {
      const next: 'secure' | 'sandbox' = isVulnerable ? 'secure' : 'sandbox'
      await switchMode(next)
      setToast({
        msg: `Switched to ${next === 'secure' ? 'Secure' : 'Vulnerable'} mode`,
        ok: true,
      })
    } catch (err) {
      setToast({
        msg: err instanceof Error ? err.message : 'Failed to switch mode',
        ok: false,
      })
    } finally {
      setSwitching(false)
      setTimeout(() => setToast(null), 3000)
    }
  }

  return (
    <>
      {/* ── Thin status bar — fixed, always on top of all content ── */}
      <div
        role="banner"
        aria-label={`Security mode: ${isVulnerable ? 'Vulnerable' : 'Secure'}`}
        className={`fixed top-0 left-0 right-0 h-8 z-[9999] flex items-center justify-between px-4 text-white text-xs font-medium select-none transition-colors duration-300 ${
          isVulnerable
            ? 'bg-red-700'
            : 'bg-[#14532d]'
        }`}
      >
        {/* Left: status label */}
        <div className="flex items-center gap-2 min-w-0">
          <span
            className={`flex-shrink-0 w-1.5 h-1.5 rounded-full bg-white/80 ${
              isLoading || switching ? 'opacity-30' : 'animate-pulse'
            }`}
          />
          <span className="font-semibold tracking-wide hidden sm:block truncate">
            {isLoading
              ? 'PT. Dana Sejahtera  ·  Checking system mode…'
              : isVulnerable
              ? 'VULNERABLE MODE  ·  OWASP protections disabled  ·  Authorised testing only'
              : 'SECURE MODE  ·  All OWASP API Top 10 protections active'}
          </span>
          {/* Mobile short label */}
          <span className="font-semibold tracking-wide sm:hidden">
            {isVulnerable ? '⚠ VULNERABLE MODE' : '✓ SECURE MODE'}
          </span>
        </div>

        {/* Right: toggle button */}
        <button
          onClick={handleToggle}
          disabled={switching || isLoading}
          className="flex-shrink-0 ml-4 px-2.5 py-0.5 rounded text-[11px] font-semibold
                     bg-white/15 hover:bg-white/28 transition-colors
                     disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {switching ? 'Switching…' : `Switch to ${isVulnerable ? 'Secure' : 'Vulnerable'}`}
        </button>
      </div>

      {/* ── Toast notification ── */}
      {toast && (
        <div
          className={`fixed top-10 right-4 z-[9999] px-4 py-2 rounded-md
                      text-xs font-semibold text-white shadow-md animate-fade-in ${
            toast.ok ? 'bg-[#14532d]' : 'bg-red-700'
          }`}
        >
          {toast.msg}
        </div>
      )}
    </>
  )
}
