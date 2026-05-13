export const SECURITY_MODE = (
  process.env.NEXT_PUBLIC_SECURITY_MODE || 'secure'
) as 'secure' | 'vulnerable'

export const IS_SECURE = SECURITY_MODE === 'secure'
export const IS_VULNERABLE = SECURITY_MODE === 'vulnerable'
