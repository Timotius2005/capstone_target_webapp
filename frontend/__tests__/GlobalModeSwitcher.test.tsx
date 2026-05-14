/**
 * GlobalModeSwitcher unit tests
 *
 * Validates:
 * - Banner renders with correct text and colour per mode
 * - Toggle button calls switchMode with the opposite mode
 * - Toast notification appears after successful switch
 * - Error toast appears when switchMode rejects
 * - UI is accessible (role, aria-label)
 */
import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import GlobalModeSwitcher from '@/components/GlobalModeSwitcher'
import { ModeContext } from '@/contexts/ModeContext'

function renderSwitcher(
  mode: 'secure' | 'sandbox',
  switchMode: jest.Mock = jest.fn().mockResolvedValue(undefined)
) {
  const ctx = { mode, isLoading: false, switchMode }
  return { switchMode, ...render(
    <ModeContext.Provider value={ctx}>
      <GlobalModeSwitcher />
    </ModeContext.Provider>
  )}
}

// ── Rendering ─────────────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — rendering', () => {
  it('shows SECURE label in secure mode', () => {
    renderSwitcher('secure')
    expect(screen.getByText(/SECURE MODE/i)).toBeInTheDocument()
  })

  it('shows VULNERABLE label in vulnerable mode', () => {
    renderSwitcher('sandbox')
    expect(screen.getByText(/VULNERABLE/i)).toBeInTheDocument()
  })

  it('offers "Switch to Vulnerable" when currently secure', () => {
    renderSwitcher('secure')
    expect(screen.getByRole('button', { name: /Switch to Vulnerable/i })).toBeInTheDocument()
  })

  it('offers "Switch to Secure" when currently vulnerable', () => {
    renderSwitcher('sandbox')
    expect(screen.getByRole('button', { name: /Switch to Secure/i })).toBeInTheDocument()
  })

  it('has accessible banner role', () => {
    renderSwitcher('secure')
    expect(screen.getByRole('banner')).toBeInTheDocument()
  })
})

// ── Toggle interaction ────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — toggle interaction', () => {
  it('calls switchMode("sandbox") when toggled from secure', async () => {
    const { switchMode } = renderSwitcher('secure')
    fireEvent.click(screen.getByRole('button', { name: /Switch to Vulnerable/i }))
    await waitFor(() => {
      expect(switchMode).toHaveBeenCalledWith('sandbox')
    })
  })

  it('calls switchMode("secure") when toggled from vulnerable', async () => {
    const { switchMode } = renderSwitcher('sandbox')
    fireEvent.click(screen.getByRole('button', { name: /Switch to Secure/i }))
    await waitFor(() => {
      expect(switchMode).toHaveBeenCalledWith('secure')
    })
  })

  it('switchMode is called exactly once per click', async () => {
    const { switchMode } = renderSwitcher('secure')
    fireEvent.click(screen.getByRole('button', { name: /Switch to Vulnerable/i }))
    await waitFor(() => expect(switchMode).toHaveBeenCalledTimes(1))
  })
})

// ── Success toast ─────────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — success toast', () => {
  it('shows success toast after switching to vulnerable', async () => {
    renderSwitcher('secure', jest.fn().mockResolvedValue(undefined))
    fireEvent.click(screen.getByRole('button', { name: /Switch to Vulnerable/i }))
    await waitFor(() => {
      expect(screen.getByText(/Switched to/i)).toBeInTheDocument()
    })
  })
})

// ── Error toast ───────────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — error handling', () => {
  it('shows error toast when switchMode rejects', async () => {
    const failingSwitch = jest.fn().mockRejectedValue(new Error('Network error'))
    renderSwitcher('secure', failingSwitch)
    fireEvent.click(screen.getByRole('button', { name: /Switch to Vulnerable/i }))
    await waitFor(() => {
      expect(screen.getByText(/Network error/i)).toBeInTheDocument()
    })
  })
})

// ── Loading state ─────────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — loading state', () => {
  it('disables button while isLoading', () => {
    const ctx = { mode: 'secure' as const, isLoading: true, switchMode: jest.fn() }
    render(
      <ModeContext.Provider value={ctx}>
        <GlobalModeSwitcher />
      </ModeContext.Provider>
    )
    const btn = screen.getByRole('button', { name: /Switch to Vulnerable/i })
    expect(btn).toBeDisabled()
  })
})

// ── Mode change propagates to context ────────────────────────────────────────

describe('GlobalModeSwitcher — context propagation', () => {
  it('mode change is reflected in the banner text after context update', () => {
    // Simulate context re-render after switchMode resolves
    const { rerender } = render(
      <ModeContext.Provider value={{ mode: 'secure', isLoading: false, switchMode: jest.fn() }}>
        <GlobalModeSwitcher />
      </ModeContext.Provider>
    )
    expect(screen.getByText(/SECURE MODE/i)).toBeInTheDocument()

    // Context updates to vulnerable
    rerender(
      <ModeContext.Provider value={{ mode: 'sandbox', isLoading: false, switchMode: jest.fn() }}>
        <GlobalModeSwitcher />
      </ModeContext.Provider>
    )
    expect(screen.getByText(/VULNERABLE/i)).toBeInTheDocument()
    expect(screen.queryByText(/SECURE MODE/i)).not.toBeInTheDocument()
  })
})
