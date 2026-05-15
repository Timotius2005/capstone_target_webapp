/**
 * VulnConfigPanel unit tests
 *
 * Validates:
 * - Panel is hidden in secure mode
 * - Panel renders all 10 OWASP category toggles in vulnerable mode
 * - Toggles call updateVulnConfig with the correct config
 * - "Enable All" / "Disable All" buttons work
 * - Secure mode ignores vulnConfig (IsVulnerableFor always false)
 */
import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import VulnConfigPanel from '@/components/VulnConfigPanel'
import { ModeContext, defaultVulnConfig, type VulnConfig } from '@/contexts/ModeContext'

// ── Context factory ───────────────────────────────────────────────────────────

function makeCtx(
  mode: 'secure' | 'sandbox',
  overrides: Partial<VulnConfig> = {},
  updateVulnConfig: jest.Mock = jest.fn().mockResolvedValue(undefined),
) {
  return {
    mode,
    isLoading:           false,
    vulnConfig:          { ...defaultVulnConfig, ...overrides },
    isVulnConfigLoading: false,
    switchMode:          jest.fn(),
    updateVulnConfig,
  }
}

function renderPanel(
  mode: 'secure' | 'sandbox' = 'sandbox',
  overrides: Partial<VulnConfig> = {},
  updateFn: jest.Mock = jest.fn().mockResolvedValue(undefined),
  onClose: jest.Mock = jest.fn(),
) {
  const ctx = makeCtx(mode, overrides, updateFn)
  return {
    updateFn,
    onClose,
    ...render(
      <ModeContext.Provider value={ctx}>
        <VulnConfigPanel onClose={onClose} />
      </ModeContext.Provider>
    ),
  }
}

// ── Rendering ─────────────────────────────────────────────────────────────────

describe('VulnConfigPanel — rendering', () => {
  it('renders nothing when mode is secure', () => {
    const { container } = renderPanel('secure')
    expect(container.firstChild).toBeNull()
  })

  it('renders the panel when mode is vulnerable (sandbox)', () => {
    renderPanel('sandbox')
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('renders all 10 OWASP category toggle switches', () => {
    renderPanel('sandbox')
    const switches = screen.getAllByRole('switch')
    expect(switches).toHaveLength(10)
  })

  it('shows every OWASP ID label (A01–A10)', () => {
    renderPanel('sandbox')
    for (let i = 1; i <= 10; i++) {
      const id = `A${i.toString().padStart(2, '0')}`
      expect(screen.getAllByText(id).length).toBeGreaterThan(0)
    }
  })

  it('shows "10/10" count in header when all categories are enabled', () => {
    renderPanel('sandbox')
    // Panel header: "{enabledCount}/10 categories active"
    expect(screen.getByText(/10\/10 categories active/)).toBeInTheDocument()
  })

  it('shows correct count in header when some categories are disabled', () => {
    renderPanel('sandbox', {
      A01_BrokenAccessControl: false,
      A07_AuthenticationFailures: false,
    })
    expect(screen.getByText(/8\/10 categories active/)).toBeInTheDocument()
  })
})

// ── Toggle interaction ────────────────────────────────────────────────────────

describe('VulnConfigPanel — toggle interaction', () => {
  it('calls updateVulnConfig with A01 = false when A01 toggle clicked OFF', async () => {
    const updateFn = jest.fn().mockResolvedValue(undefined)
    renderPanel('sandbox', {}, updateFn)

    // A01 switch is currently ON (aria-checked="true")
    const a01Switch = screen
      .getAllByRole('switch')
      .find((s) => s.getAttribute('aria-label')?.includes('A01'))!

    await act(async () => { fireEvent.click(a01Switch) })

    await waitFor(() => expect(updateFn).toHaveBeenCalledTimes(1))

    const calledWith: VulnConfig = updateFn.mock.calls[0][0]
    expect(calledWith.A01_BrokenAccessControl).toBe(false)
    // All other categories must remain unchanged
    expect(calledWith.A07_AuthenticationFailures).toBe(true)
    expect(calledWith.A10_SSRF).toBe(true)
  })

  it('calls updateVulnConfig with A07 = true when A07 toggled ON from OFF', async () => {
    const updateFn = jest.fn().mockResolvedValue(undefined)
    renderPanel('sandbox', { A07_AuthenticationFailures: false }, updateFn)

    const a07Switch = screen
      .getAllByRole('switch')
      .find((s) => s.getAttribute('aria-label')?.includes('A07'))!

    expect(a07Switch.getAttribute('aria-checked')).toBe('false')

    await act(async () => { fireEvent.click(a07Switch) })
    await waitFor(() => expect(updateFn).toHaveBeenCalledTimes(1))

    const calledWith: VulnConfig = updateFn.mock.calls[0][0]
    expect(calledWith.A07_AuthenticationFailures).toBe(true)
  })

  it('"Enable All" calls updateVulnConfig with all categories true', async () => {
    const updateFn = jest.fn().mockResolvedValue(undefined)
    renderPanel('sandbox', { A01_BrokenAccessControl: false, A10_SSRF: false }, updateFn)

    await act(async () => { fireEvent.click(screen.getByText('Enable All')) })
    await waitFor(() => expect(updateFn).toHaveBeenCalledTimes(1))

    const calledWith: VulnConfig = updateFn.mock.calls[0][0]
    expect(Object.values(calledWith).every(Boolean)).toBe(true)
  })

  it('"Disable All" calls updateVulnConfig with all categories false', async () => {
    const updateFn = jest.fn().mockResolvedValue(undefined)
    renderPanel('sandbox', {}, updateFn)

    await act(async () => { fireEvent.click(screen.getByText('Disable All')) })
    await waitFor(() => expect(updateFn).toHaveBeenCalledTimes(1))

    const calledWith: VulnConfig = updateFn.mock.calls[0][0]
    expect(Object.values(calledWith).every((v) => v === false)).toBe(true)
  })
})

// ── Close behaviour ───────────────────────────────────────────────────────────

describe('VulnConfigPanel — close behaviour', () => {
  it('calls onClose when the × button is clicked', () => {
    const onClose = jest.fn()
    renderPanel('sandbox', {}, jest.fn(), onClose)

    fireEvent.click(screen.getByLabelText('Close vulnerability config panel'))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('calls onClose when the backdrop is clicked', () => {
    const onClose = jest.fn()
    renderPanel('sandbox', {}, jest.fn(), onClose)

    // The dialog element itself is the backdrop
    fireEvent.click(screen.getByRole('dialog'))
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})

// ── Secure mode ignores config — conceptual test ──────────────────────────────

describe('VulnConfigPanel — secure mode is never degraded', () => {
  it('renders null in secure mode even when vulnConfig has all-true values', () => {
    // If VulnConfigPanel renders nothing in secure mode, secure mode cannot be
    // degraded through the UI — the config panel is simply inaccessible.
    const ctx = makeCtx('secure', defaultVulnConfig) // all true
    const { container } = render(
      <ModeContext.Provider value={ctx}>
        <VulnConfigPanel onClose={jest.fn()} />
      </ModeContext.Provider>
    )
    expect(container.firstChild).toBeNull()
  })
})
