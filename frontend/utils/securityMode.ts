// Runtime mode variable — updated dynamically by ModeContext after fetching
// the current mode from the backend. Non-React callers (api.ts, auth.ts) read
// this directly; React components should use useMode() for reactive re-renders.
let _runtimeMode: 'secure' | 'sandbox' =
  typeof process !== 'undefined' &&
  (process.env.NEXT_PUBLIC_SECURITY_MODE === 'sandbox' ||
    process.env.NEXT_PUBLIC_SECURITY_MODE === 'vulnerable')
    ? 'sandbox'
    : 'secure'

/** Called by ModeContext after a successful GET or PUT /config/mode. */
export const setRuntimeMode = (mode: 'secure' | 'sandbox'): void => {
  _runtimeMode = mode
}

export const getRuntimeMode = (): 'secure' | 'sandbox' => _runtimeMode

export const isSecure = (): boolean => _runtimeMode === 'secure'

/** True when mode is "sandbox". Preserved for OWASP injection-point code. */
export const isVulnerable = (): boolean => _runtimeMode === 'sandbox'

export const getMode = (): 'secure' | 'sandbox' => _runtimeMode
