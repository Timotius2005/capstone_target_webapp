import { SECURITY_MODE } from '@/config/security'

export const isSecure = (): boolean => SECURITY_MODE === 'secure'
export const isVulnerable = (): boolean => SECURITY_MODE === 'vulnerable'
export const getMode = (): 'secure' | 'vulnerable' =>
  SECURITY_MODE === 'vulnerable' ? 'vulnerable' : 'secure'
