'use client'

import { useCallback } from 'react'
import { useMode, type VulnConfig, defaultVulnConfig } from '@/contexts/ModeContext'

// ── OWASP category metadata ───────────────────────────────────────────────────

interface OWASPCategory {
  key: keyof VulnConfig
  id:  string
  label: string
  desc: string
}

const OWASP_CATEGORIES: OWASPCategory[] = [
  {
    key:   'A01_BrokenAccessControl',
    id:    'A01',
    label: 'Broken Access Control',
    desc:  'BOLA/IDOR — any user reads/modifies any resource without ownership check.',
  },
  {
    key:   'A02_CryptographicFailures',
    id:    'A02',
    label: 'Cryptographic Failures',
    desc:  'Sensitive fields (password_hash, NIK, internal IDs) exposed in API responses.',
  },
  {
    key:   'A03_Injection',
    id:    'A03',
    label: 'Injection / BOPLA',
    desc:  'NIK uniqueness bypass; mass-assignment via unchecked field writes.',
  },
  {
    key:   'A04_InsecureDesign',
    id:    'A04',
    label: 'Insecure Design',
    desc:  'Pagination enforcement disabled — full table dumps; loan-creation rate limits removed.',
  },
  {
    key:   'A05_SecurityMisconfiguration',
    id:    'A05',
    label: 'Security Misconfiguration',
    desc:  'Database query debug logging exposes raw SQL including sensitive parameters.',
  },
  {
    key:   'A06_VulnerableComponents',
    id:    'A06',
    label: 'Vulnerable Components',
    desc:  'Loan amount validation and transaction business-flow rules bypassed.',
  },
  {
    key:   'A07_AuthenticationFailures',
    id:    'A07',
    label: 'Authentication Failures',
    desc:  'Login uses plain-text compare; 100-year JWT; no account lockout; admin routes open to all.',
  },
  {
    key:   'A08_SoftwareDataIntegrityFailures',
    id:    'A08',
    label: 'Software & Data Integrity',
    desc:  'Staff can mass-assign loan status to "approved" — bypasses admin-only rule.',
  },
  {
    key:   'A09_SecurityLoggingFailures',
    id:    'A09',
    label: 'Security Logging Failures',
    desc:  'Login rate-limiting disabled — brute-force attacks go undetected and unblocked.',
  },
  {
    key:   'A10_SSRF',
    id:    'A10',
    label: 'SSRF',
    desc:  'URL allowlist bypassed — arbitrary internal/external URLs reachable via fetch endpoint.',
  },
]

// ── Toggle row ────────────────────────────────────────────────────────────────

interface ToggleRowProps {
  category:  OWASPCategory
  enabled:   boolean
  disabled:  boolean
  onChange:  (key: keyof VulnConfig, value: boolean) => void
}

function ToggleRow({ category, enabled, disabled, onChange }: ToggleRowProps) {
  return (
    <div className={`flex items-start gap-3 px-3 py-2.5 rounded-md transition-colors ${
      enabled
        ? 'bg-red-500/10 border border-red-500/20'
        : 'bg-white/[0.03] border border-white/[0.06]'
    }`}>
      {/* Toggle switch */}
      <button
        type="button"
        role="switch"
        aria-checked={enabled}
        aria-label={`${category.id} ${category.label}: ${enabled ? 'vulnerable (on)' : 'secure (off)'}`}
        disabled={disabled}
        onClick={() => onChange(category.key, !enabled)}
        className={`relative flex-shrink-0 mt-0.5 w-8 h-4 rounded-full transition-colors duration-200
                    focus:outline-none focus:ring-2 focus:ring-red-400 focus:ring-offset-1 focus:ring-offset-transparent
                    disabled:opacity-40 disabled:cursor-not-allowed
                    ${enabled ? 'bg-red-500' : 'bg-white/20'}`}
      >
        <span className={`absolute top-0.5 left-0.5 w-3 h-3 bg-white rounded-full shadow transition-transform duration-200
                          ${enabled ? 'translate-x-4' : 'translate-x-0'}`} />
      </button>

      {/* Label + description */}
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className={`text-[10px] font-bold font-mono px-1 py-0.5 rounded ${
            enabled ? 'text-red-300 bg-red-500/20' : 'text-white/40 bg-white/5'
          }`}>
            {category.id}
          </span>
          <span className={`text-[11px] font-semibold truncate ${
            enabled ? 'text-red-200' : 'text-white/50'
          }`}>
            {category.label}
          </span>
          <span className={`ml-auto text-[9px] font-bold uppercase tracking-widest flex-shrink-0 ${
            enabled ? 'text-red-400' : 'text-green-400/60'
          }`}>
            {enabled ? 'VULN' : 'SAFE'}
          </span>
        </div>
        <p className="text-[10px] text-white/30 mt-0.5 leading-snug">{category.desc}</p>
      </div>
    </div>
  )
}

// ── Panel ─────────────────────────────────────────────────────────────────────

interface VulnConfigPanelProps {
  onClose: () => void
}

export default function VulnConfigPanel({ onClose }: VulnConfigPanelProps) {
  const { mode, vulnConfig, isVulnConfigLoading, updateVulnConfig } = useMode()

  // Only render when in vulnerable mode
  if (mode !== 'sandbox') return null

  const handleToggle = useCallback(
    (key: keyof VulnConfig, value: boolean) => {
      updateVulnConfig({ ...vulnConfig, [key]: value })
    },
    [vulnConfig, updateVulnConfig]
  )

  const handleEnableAll = () => updateVulnConfig(defaultVulnConfig)

  const handleDisableAll = () => {
    const allOff = Object.fromEntries(
      Object.keys(defaultVulnConfig).map((k) => [k, false])
    ) as VulnConfig
    updateVulnConfig(allOff)
  }

  const enabledCount = Object.values(vulnConfig).filter(Boolean).length

  return (
    /* Backdrop */
    <div
      className="fixed inset-0 z-[9998] flex items-start justify-end pt-10 pr-4"
      onClick={(e) => { if (e.target === e.currentTarget) onClose() }}
      role="dialog"
      aria-modal="true"
      aria-label="OWASP Vulnerability Configuration"
    >
      <div
        className="w-[340px] max-h-[calc(100vh-3rem)] flex flex-col
                   bg-[#0B1E3D] border border-red-500/30 rounded-lg shadow-2xl
                   overflow-hidden animate-fade-in"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-red-500/20 bg-red-500/10 flex-shrink-0">
          <div>
            <p className="text-[11px] font-bold text-red-300 uppercase tracking-widest">
              OWASP Top 10 — Vulnerability Config
            </p>
            <p className="text-[10px] text-white/40 mt-0.5">
              {enabledCount}/10 categories active
              {isVulnConfigLoading && ' · Saving…'}
            </p>
          </div>
          <button
            onClick={onClose}
            aria-label="Close vulnerability config panel"
            className="text-white/40 hover:text-white/80 transition-colors p-1 rounded"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Bulk actions */}
        <div className="flex gap-2 px-3 py-2 border-b border-white/[0.06] flex-shrink-0">
          <button
            onClick={handleEnableAll}
            disabled={isVulnConfigLoading}
            className="flex-1 text-[10px] font-semibold py-1.5 px-2 rounded
                       bg-red-500/20 text-red-300 hover:bg-red-500/30 transition-colors
                       disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Enable All
          </button>
          <button
            onClick={handleDisableAll}
            disabled={isVulnConfigLoading}
            className="flex-1 text-[10px] font-semibold py-1.5 px-2 rounded
                       bg-white/[0.06] text-white/60 hover:bg-white/10 transition-colors
                       disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Disable All
          </button>
        </div>

        {/* Category list */}
        <div className="overflow-y-auto flex-1 p-2 space-y-1.5">
          {OWASP_CATEGORIES.map((cat) => (
            <ToggleRow
              key={cat.key}
              category={cat}
              enabled={vulnConfig[cat.key]}
              disabled={isVulnConfigLoading}
              onChange={handleToggle}
            />
          ))}
        </div>

        {/* Footer note */}
        <div className="px-4 py-2.5 border-t border-white/[0.06] bg-[#0B1E3D] flex-shrink-0">
          <p className="text-[9px] text-white/25 leading-snug">
            Disabled categories behave as secure while all others remain vulnerable.
            Config resets to all-on when switching back to Secure Mode.
          </p>
        </div>
      </div>
    </div>
  )
}
