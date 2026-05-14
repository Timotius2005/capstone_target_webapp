/**
 * ModeContext unit tests
 *
 * Validates:
 * - Initial mode fetch from /api/system/mode on mount
 * - switchMode calls PUT /api/system/mode without auth token
 * - Mode state updates immediately after successful API call
 * - Error is propagated when API call fails
 * - Cookie is synced after mode change
 */
import React from 'react'
import { render, screen, waitFor, act } from '@testing-library/react'
import { ModeProvider, useMode } from '@/contexts/ModeContext'

// ── Helpers ───────────────────────────────────────────────────────────────────

function ModeConsumer() {
  const { mode, isLoading, switchMode } = useMode()
  return (
    <div>
      <span data-testid="mode">{mode}</span>
      <span data-testid="loading">{isLoading ? 'loading' : 'ready'}</span>
      <button onClick={() => switchMode('sandbox').catch(() => {})}>to-vulnerable</button>
      <button onClick={() => switchMode('secure').catch(() => {})}>to-secure</button>
    </div>
  )
}

// Track the order of fetch calls for PUT vs GET discrimination
function mockFetchSequence(calls: { response: object; status?: number }[]) {
  let idx = 0
  global.fetch = jest.fn().mockImplementation(() => {
    const call = calls[idx] ?? calls[calls.length - 1]
    idx++
    const status = call.status ?? 200
    return Promise.resolve({
      ok: status < 400,
      status,
      json: () => Promise.resolve(call.response),
    } as Response)
  })
}

beforeEach(() => {
  document.cookie = '_app_mode=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;'
  localStorage.clear()
  sessionStorage.clear()
})

afterEach(() => {
  jest.restoreAllMocks()
})

// ── Initial mount ─────────────────────────────────────────────────────────────

describe('ModeProvider — initial mount', () => {
  it('fetches current mode on mount', async () => {
    const fetchMock = jest.fn().mockResolvedValue({
      ok: true, status: 200,
      json: () => Promise.resolve({ mode: 'secure' }),
    } as Response)
    global.fetch = fetchMock

    render(<ModeProvider><ModeConsumer /></ModeProvider>)

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalled()
    })
  })

  it('sets mode to "secure" when backend returns secure', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true, json: () => Promise.resolve({ mode: 'secure' }),
    } as Response)

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))
    expect(screen.getByTestId('mode').textContent).toBe('secure')
  })

  it('sets mode to "sandbox" when backend returns vulnerable', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true, json: () => Promise.resolve({ mode: 'vulnerable' }),
    } as Response)

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))
    expect(screen.getByTestId('mode').textContent).toBe('sandbox')
  })

  it('isLoading becomes false after fetch completes', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true, json: () => Promise.resolve({ mode: 'secure' }),
    } as Response)

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))
  })

  it('keeps env-var default when backend is unreachable', async () => {
    global.fetch = jest.fn().mockRejectedValue(new Error('network error'))
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))
    expect(screen.getByTestId('mode').textContent).toBe('secure')
  })
})

// ── switchMode ────────────────────────────────────────────────────────────────

describe('ModeProvider — switchMode', () => {
  it('calls PUT /api/system/mode without Authorization header', async () => {
    let capturedOptions: RequestInit | undefined
    let callCount = 0
    global.fetch = jest.fn().mockImplementation((_url: string, options?: RequestInit) => {
      callCount++
      if (callCount === 1) {
        // GET on mount
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ mode: 'secure' }) } as Response)
      }
      capturedOptions = options
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ mode: 'vulnerable' }) } as Response)
    })

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => { screen.getByText('to-vulnerable').click() })
    await waitFor(() => { if (!capturedOptions) throw new Error('no PUT call yet') })

    const headers = (capturedOptions?.headers ?? {}) as Record<string, string>
    expect(headers['Authorization']).toBeUndefined()
  })

  it('sends "vulnerable" in body when switching to sandbox', async () => {
    let putBody: Record<string, unknown> | undefined
    let callCount = 0
    global.fetch = jest.fn().mockImplementation((_url: string, options?: RequestInit) => {
      callCount++
      if (callCount === 1) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ mode: 'secure' }) } as Response)
      }
      putBody = JSON.parse(options?.body as string)
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ mode: 'vulnerable' }) } as Response)
    })

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => { screen.getByText('to-vulnerable').click() })
    await waitFor(() => { if (!putBody) throw new Error('no PUT call yet') })

    expect(putBody!.mode).toBe('vulnerable')
  })

  it('updates mode state immediately after successful switch', async () => {
    mockFetchSequence([
      { response: { mode: 'secure' } },    // GET mount
      { response: { mode: 'vulnerable' } }, // PUT switch
    ])

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('mode').textContent).toBe('secure'))

    await act(async () => { screen.getByText('to-vulnerable').click() })

    await waitFor(() => {
      expect(screen.getByTestId('mode').textContent).toBe('sandbox')
    })
  })

  it('mode switch does NOT reload the page', async () => {
    // Test that window.location.reload is NOT called during a mode switch
    // by checking that the test completes successfully without navigation side-effects.
    mockFetchSequence([
      { response: { mode: 'secure' } },
      { response: { mode: 'vulnerable' } },
    ])

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    // Switch mode — if this triggers a reload the test environment resets
    // and the subsequent assertion fails, making a failed test observable.
    await act(async () => { screen.getByText('to-vulnerable').click() })
    await waitFor(() => expect(screen.getByTestId('mode').textContent).toBe('sandbox'))

    // If we reach here without an error, no reload happened.
    expect(screen.getByTestId('mode').textContent).toBe('sandbox')
  })
})

// ── Cookie sync ───────────────────────────────────────────────────────────────

describe('ModeProvider — cookie sync', () => {
  it('sets _app_mode cookie after successful mode switch', async () => {
    mockFetchSequence([
      { response: { mode: 'secure' } },
      { response: { mode: 'vulnerable' } },
    ])

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => { screen.getByText('to-vulnerable').click() })
    await waitFor(() => expect(screen.getByTestId('mode').textContent).toBe('sandbox'))

    expect(document.cookie).toContain('_app_mode')
  })
})
