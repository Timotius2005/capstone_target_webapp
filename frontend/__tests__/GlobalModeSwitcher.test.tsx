/**
 * GlobalModeSwitcher unit tests
 */
import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import GlobalModeSwitcher from '@/components/GlobalModeSwitcher'
import { ModeContext, defaultVulnConfig } from '@/contexts/ModeContext'

function renderSwitcher(
  mode: 'secure' | 'sandbox',
  switchMode: jest.Mock = jest.fn().mockResolvedValue(undefined)
) {
  const ctx = {
    mode,
    isLoading:           false,
    vulnConfig:          defaultVulnConfig,
    isVulnConfigLoading: false,
    switchMode,
    updateVulnConfig:    jest.fn().mockResolvedValue(undefined),
  }
  return {
    switchMode,
    ...render(
      <ModeContext.Provider value={ctx}>
        <GlobalModeSwitcher />
      </ModeContext.Provider>
    ),
  }
}

// ── Rendering ─────────────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — rendering', () => {
  it('shows SECURE MODE label in secure mode', () => {
    renderSwitcher('secure')
    // Component renders two spans with the mode text (desktop + mobile)
    // — use getAllByText and check at least one is present
    const labels = screen.getAllByText(/SECURE MODE/i)
    expect(labels.length).toBeGreaterThan(0)
  })

  it('shows VULNERABLE label in vulnerable mode', () => {
    renderSwitcher('sandbox')
    const labels = screen.getAllByText(/VULNERABLE/i)
    expect(labels.length).toBeGreaterThan(0)
  })

  it('offers "Switch to Vulnerable" toggle button when currently secure', () => {
    renderSwitcher('secure')
    expect(screen.getByRole('button', { name: /Switch to Vulnerable/i })).toBeInTheDocument()
  })

  it('offers "Switch to Secure" toggle button when currently vulnerable', () => {
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
    await waitFor(() => expect(switchMode).toHaveBeenCalledWith('sandbox'))
  })

  it('calls switchMode("secure") when toggled from vulnerable', async () => {
    const { switchMode } = renderSwitcher('sandbox')
    fireEvent.click(screen.getByRole('button', { name: /Switch to Secure/i }))
    await waitFor(() => expect(switchMode).toHaveBeenCalledWith('secure'))
  })

  it('switchMode is called exactly once per click', async () => {
    const { switchMode } = renderSwitcher('secure')
    fireEvent.click(screen.getByRole('button', { name: /Switch to Vulnerable/i }))
    await waitFor(() => expect(switchMode).toHaveBeenCalledTimes(1))
  })
})

// ── Success toast ─────────────────────────────────────────────────────────────

describe('GlobalModeSwitcher — success toast', () => {
  it('shows a toast after a successful switch', async () => {
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
    const ctx = {
      mode:                'secure' as const,
      isLoading:           true,
      vulnConfig:          defaultVulnConfig,
      isVulnConfigLoading: false,
      switchMode:          jest.fn(),
      updateVulnConfig:    jest.fn().mockResolvedValue(undefined),
    }
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
  it('banner text updates when context switches from secure to vulnerable', () => {
    const base = {
      isLoading: false, vulnConfig: defaultVulnConfig,
      isVulnConfigLoading: false, updateVulnConfig: jest.fn(),
    }
    const { rerender } = render(
      <ModeContext.Provider value={{ ...base, mode: 'secure', switchMode: jest.fn() }}>
        <GlobalModeSwitcher />
      </ModeContext.Provider>
    )
    expect(screen.getAllByText(/SECURE MODE/i).length).toBeGreaterThan(0)

    rerender(
      <ModeContext.Provider value={{ ...base, mode: 'sandbox', switchMode: jest.fn() }}>
        <GlobalModeSwitcher />
      </ModeContext.Provider>
    )

    expect(screen.getAllByText(/VULNERABLE/i).length).toBeGreaterThan(0)
    expect(screen.queryAllByText(/SECURE MODE/i).length).toBe(0)
  })
})
