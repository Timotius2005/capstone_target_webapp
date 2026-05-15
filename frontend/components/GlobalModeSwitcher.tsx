'use client'

import { useState } from 'react'
import { useMode } from '@/contexts/ModeContext'
import VulnConfigPanel from '@/components/VulnConfigPanel'

export default function GlobalModeSwitcher() {
  const { mode, isLoading, switchMode, vulnConfig } = useMode()
  const [switching, setSwitching]       = useState(false)
  const [configOpen, setConfigOpen]     = useState(false)
  const [toast, setToast]               = useState<{ msg: string; ok: boolean } | null>(null)

  const isVulnerable = mode === 'sandbox'

  // Count active categories for the badge on the Configure button
  const activeCategories = isVulnerable
    ? Object.values(vulnConfig).filter(Boolean).length
    : 0

  const handleToggle = async () => {
    if (switching || isLoading) return
    setSwitching(true)
    // Close config panel when switching modes
    setConfigOpen(false)
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

        {/* Right: controls */}
        <div className="flex items-center gap-2 flex-shrink-0 ml-4">
          {/* Configure button — only shown in vulnerable mode */}
          {isVulnerable && !isLoading && (
            <button
              onClick={() => setConfigOpen((o) => !o)}
              aria-label="Configure OWASP vulnerability categories"
              aria-expanded={configOpen}
              className="flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-semibold
                         bg-white/15 hover:bg-white/28 transition-colors"
            >
              {/* gear icon */}
              <svg className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round"
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066
                     c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924
                     0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724
                     1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0
                     00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572
                     c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543
                     .826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
              <span className="hidden sm:inline">Configure</span>
              {/* Active-category badge */}
              <span className={`text-[9px] font-bold px-1 py-0.5 rounded-sm ${
                activeCategories < 10 ? 'bg-amber-400 text-amber-900' : 'bg-white/20 text-white'
              }`}>
                {activeCategories}/10
              </span>
            </button>
          )}

          {/* Mode toggle button */}
          <button
            onClick={handleToggle}
            disabled={switching || isLoading}
            className="flex-shrink-0 px-2.5 py-0.5 rounded text-[11px] font-semibold
                       bg-white/15 hover:bg-white/28 transition-colors
                       disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {switching ? 'Switching…' : `Switch to ${isVulnerable ? 'Secure' : 'Vulnerable'}`}
          </button>
        </div>
      </div>

      {/* ── Vuln config panel (flyout, only in vulnerable mode) ── */}
      {configOpen && isVulnerable && (
        <VulnConfigPanel onClose={() => setConfigOpen(false)} />
      )}

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
