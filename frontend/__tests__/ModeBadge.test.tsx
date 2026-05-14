/**
 * ModeBadge unit tests
 *
 * Validates that ModeBadge correctly reflects the current security mode
 * from ModeContext and applies the right visual styling.
 */
import React from 'react'
import { render, screen } from '@testing-library/react'
import ModeBadge from '@/components/ModeBadge'
import { ModeContext } from '@/contexts/ModeContext'

// Helper: wrap a component with a mock ModeContext value
function renderWithMode(
  mode: 'secure' | 'sandbox',
  props: React.ComponentProps<typeof ModeBadge> = {}
) {
  const ctx = {
    mode,
    isLoading: false,
    switchMode: jest.fn().mockResolvedValue(undefined),
  }
  return render(
    <ModeContext.Provider value={ctx}>
      <ModeBadge {...props} />
    </ModeContext.Provider>
  )
}

// ── Secure mode ───────────────────────────────────────────────────────────────

describe('ModeBadge — Secure mode', () => {
  it('renders "Secure" label', () => {
    renderWithMode('secure')
    expect(screen.getByText('Secure')).toBeInTheDocument()
  })

  it('does NOT render vulnerable label when secure', () => {
    renderWithMode('secure')
    expect(screen.queryByText('Vulnerable')).not.toBeInTheDocument()
  })

  it('applies green colour classes in secure mode', () => {
    const { container } = renderWithMode('secure')
    const badge = container.firstElementChild as HTMLElement
    expect(badge.className).toMatch(/green/)
  })
})

// ── Vulnerable / Sandbox mode ─────────────────────────────────────────────────

describe('ModeBadge — Vulnerable (sandbox) mode', () => {
  it('renders "Vulnerable" label', () => {
    renderWithMode('sandbox')
    expect(screen.getByText('Vulnerable')).toBeInTheDocument()
  })

  it('does NOT render secure label when vulnerable', () => {
    renderWithMode('sandbox')
    expect(screen.queryByText('Secure')).not.toBeInTheDocument()
  })

  it('applies red colour classes in vulnerable mode', () => {
    const { container } = renderWithMode('sandbox')
    const badge = container.firstElementChild as HTMLElement
    expect(badge.className).toMatch(/red/)
  })
})

// ── Size prop ─────────────────────────────────────────────────────────────────

describe('ModeBadge — size prop', () => {
  it('renders with sm size without error', () => {
    renderWithMode('secure', { size: 'sm' })
    expect(screen.getByText('Secure')).toBeInTheDocument()
  })

  it('renders with md size without error', () => {
    renderWithMode('sandbox', { size: 'md' })
    expect(screen.getByText('Vulnerable')).toBeInTheDocument()
  })
})

// ── Loading state ─────────────────────────────────────────────────────────────

describe('ModeBadge — isLoading state', () => {
  it('still renders label while loading (shows last-known mode)', () => {
    const ctx = { mode: 'secure' as const, isLoading: true, switchMode: jest.fn() }
    render(
      <ModeContext.Provider value={ctx}>
        <ModeBadge />
      </ModeContext.Provider>
    )
    // Badge renders the current mode regardless of loading state
    expect(screen.getByText('Secure')).toBeInTheDocument()
  })
})
