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

// Tiny consumer component to expose context values in tests
function ModeConsumer() {
  const { mode, isLoading, switchMode } = useMode()
  return (
    <div>
      <span data-testid="mode">{mode}</span>
      <span data-testid="loading">{isLoading ? 'loading' : 'ready'}</span>
      <button onClick={() => switchMode('sandbox')}>to-vulnerable</button>
      <button onClick={() => switchMode('secure')}>to-secure</button>
    </div>
  )
}

function mockFetch(responses: { url: string; response: object; status?: number }[]) {
  const mockFn = jest.fn((url: string, options?: RequestInit) => {
    const match = responses.find((r) => url.includes(r.url))
    if (!match) {
      return Promise.reject(new Error(`Unexpected fetch: ${url}`))
    }
    return Promise.resolve({
      ok: (match.status ?? 200) < 400,
      status: match.status ?? 200,
      json: () => Promise.resolve(match.response),
    } as Response)
  })
  global.fetch = mockFn as unknown as typeof fetch
  return mockFn
}

beforeEach(() => {
  // Reset cookies and storage
  document.cookie = '_app_mode=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;'
  localStorage.clear()
  sessionStorage.clear()
})

afterEach(() => {
  jest.restoreAllMocks()
})

// ── Initial mount ─────────────────────────────────────────────────────────────

describe('ModeProvider — initial mount', () => {
  it('fetches current mode from /api/system/mode on mount', async () => {
    const fetchMock = mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },
    ])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining('/api/system/mode'),
        expect.objectContaining({ headers: expect.any(Object) })
      )
    })
  })

  it('sets mode to "secure" when backend returns secure', async () => {
    mockFetch([{ url: '/api/system/mode', response: { mode: 'secure' } }])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => {
      expect(screen.getByTestId('mode').textContent).toBe('secure')
    })
  })

  it('sets mode to "sandbox" when backend returns vulnerable', async () => {
    mockFetch([{ url: '/api/system/mode', response: { mode: 'vulnerable' } }])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => {
      expect(screen.getByTestId('mode').textContent).toBe('sandbox')
    })
  })

  it('isLoading becomes false after fetch completes', async () => {
    mockFetch([{ url: '/api/system/mode', response: { mode: 'secure' } }])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => {
      expect(screen.getByTestId('loading').textContent).toBe('ready')
    })
  })

  it('keeps env-var default when backend is unreachable', async () => {
    global.fetch = jest.fn().mockRejectedValue(new Error('network error'))
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => {
      // isLoading becomes false even on error
      expect(screen.getByTestId('loading').textContent).toBe('ready')
    })
    // Mode remains at the env-var default (secure)
    expect(screen.getByTestId('mode').textContent).toBe('secure')
  })
})

// ── switchMode ────────────────────────────────────────────────────────────────

describe('ModeProvider — switchMode', () => {
  it('calls PUT /api/system/mode without Authorization header', async () => {
    const fetchMock = mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },         // GET
      { url: '/api/system/mode', response: { mode: 'vulnerable' } },     // PUT
    ])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => {
      screen.getByText('to-vulnerable').click()
    })

    const putCall = fetchMock.mock.calls.find(
      ([, opts]) => opts?.method === 'PUT'
    )
    expect(putCall).toBeDefined()
    // Must NOT include Authorization header (public endpoint)
    const headers = (putCall?.[1]?.headers ?? {}) as Record<string, string>
    expect(headers['Authorization']).toBeUndefined()
  })

  it('sends "vulnerable" in body when switching to sandbox', async () => {
    const fetchMock = mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },
      { url: '/api/system/mode', response: { mode: 'vulnerable' } },
    ])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => {
      screen.getByText('to-vulnerable').click()
    })

    const putCall = fetchMock.mock.calls.find(([, opts]) => opts?.method === 'PUT')
    const body = JSON.parse(putCall?.[1]?.body as string)
    expect(body.mode).toBe('vulnerable')
  })

  it('updates mode state immediately after successful switch', async () => {
    mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },
      { url: '/api/system/mode', response: { mode: 'vulnerable' } },
    ])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('mode').textContent).toBe('secure'))

    await act(async () => {
      screen.getByText('to-vulnerable').click()
    })

    await waitFor(() => {
      expect(screen.getByTestId('mode').textContent).toBe('sandbox')
    })
  })

  it('mode switch does NOT require a page reload', async () => {
    mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },
      { url: '/api/system/mode', response: { mode: 'vulnerable' } },
    ])
    // Spy on window.location.reload — it must NOT be called
    const reloadSpy = jest.spyOn(window.location, 'reload').mockImplementation(() => {})

    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => {
      screen.getByText('to-vulnerable').click()
    })

    await waitFor(() => expect(screen.getByTestId('mode').textContent).toBe('sandbox'))
    expect(reloadSpy).not.toHaveBeenCalled()
    reloadSpy.mockRestore()
  })

  it('propagates error when PUT /api/system/mode fails', async () => {
    mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },            // GET
      { url: '/api/system/mode', response: { error: 'server error' }, status: 500 }, // PUT
    ])

    const { container } = render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    let caughtError: Error | undefined
    const TestWrapper = () => {
      const { switchMode } = useMode()
      return (
        <button
          onClick={() =>
            switchMode('sandbox').catch((e) => { caughtError = e })
          }
        >
          switch
        </button>
      )
    }

    render(
      <ModeProvider><TestWrapper /></ModeProvider>
    )
    await waitFor(() => {}) // wait for mount fetch

    await act(async () => {
      screen.getByText('switch').click()
    })

    await waitFor(() => {
      expect(caughtError).toBeDefined()
    })
  })
})

// ── Cookie sync ───────────────────────────────────────────────────────────────

describe('ModeProvider — cookie sync', () => {
  it('sets _app_mode cookie after successful mode switch', async () => {
    mockFetch([
      { url: '/api/system/mode', response: { mode: 'secure' } },
      { url: '/api/system/mode', response: { mode: 'vulnerable' } },
    ])
    render(<ModeProvider><ModeConsumer /></ModeProvider>)
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('ready'))

    await act(async () => {
      screen.getByText('to-vulnerable').click()
    })

    await waitFor(() => expect(screen.getByTestId('mode').textContent).toBe('sandbox'))
    expect(document.cookie).toContain('_app_mode')
  })
})
