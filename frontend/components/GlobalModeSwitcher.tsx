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
        msg: `Switched to ${next === 'secure' ? 'SECURE' : 'VULNERABLE'} mode`,
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
    <div className="fixed bottom-5 right-5 z-[9999] flex flex-col items-end gap-2 pointer-events-none">
      {/* Toast notification */}
      {toast && (
        <div
          className={`pointer-events-auto px-4 py-2 rounded-xl text-xs font-semibold shadow-xl border backdrop-blur-md animate-fade-in ${
            toast.ok
              ? 'bg-emerald-900/80 border-emerald-500/40 text-emerald-300'
              : 'bg-red-900/80 border-red-500/40 text-red-300'
          }`}
        >
          {toast.msg}
        </div>
      )}

      {/* Main badge + toggle */}
      <div
        role="button"
        tabIndex={0}
        onClick={handleToggle}
        onKeyDown={(e) => e.key === 'Enter' && handleToggle()}
        aria-label={`Security mode: ${isVulnerable ? 'Vulnerable' : 'Secure'}. Click to toggle.`}
        title={`Click to switch to ${isVulnerable ? 'Secure' : 'Vulnerable'} mode`}
        className={`pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-2xl shadow-2xl border cursor-pointer select-none transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-offset-2 ${
          isVulnerable
            ? 'bg-red-950/70 border-red-500/50 hover:bg-red-900/80 focus:ring-red-500 active:scale-95'
            : 'bg-emerald-950/70 border-emerald-500/50 hover:bg-emerald-900/80 focus:ring-emerald-500 active:scale-95'
        } backdrop-blur-md`}
      >
        {/* Animated status dot */}
        <span
          className={`w-3 h-3 rounded-full flex-shrink-0 ${
            isVulnerable ? 'bg-red-400' : 'bg-emerald-400'
          } ${switching || isLoading ? 'opacity-40' : 'animate-pulse'}`}
        />

        {/* Label stack */}
        <div className="flex flex-col leading-tight">
          <span className="text-[9px] uppercase tracking-[0.15em] font-bold text-slate-400">
            Security Mode
          </span>
          <span
            className={`text-sm font-extrabold tracking-wide ${
              isVulnerable ? 'text-red-400' : 'text-emerald-400'
            }`}
          >
            {isLoading
              ? '…'
              : switching
                ? 'Switching…'
                : isVulnerable
                  ? '⚠  VULNERABLE'
                  : '🔒 SECURE'}
          </span>
        </div>

        {/* Toggle pill */}
        <div
          className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-300 flex-shrink-0 ${
            isVulnerable ? 'bg-red-500' : 'bg-emerald-500'
          } ${switching || isLoading ? 'opacity-50' : ''}`}
        >
          <span
            className={`absolute inline-block h-4 w-4 rounded-full bg-white shadow-md transition-transform duration-300 ${
              isVulnerable ? 'translate-x-6' : 'translate-x-1'
            } ${switching ? 'animate-pulse' : ''}`}
          />
        </div>
      </div>
    </div>
  )
}
