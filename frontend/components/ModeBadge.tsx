'use client'

import { useMode } from '@/contexts/ModeContext'

interface Props {
  className?: string
  size?: 'sm' | 'md'
}

export default function ModeBadge({ className = '', size = 'md' }: Props) {
  const { mode } = useMode()
  const isSandbox = mode === 'sandbox'

  const base =
    size === 'sm'
      ? 'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-semibold tracking-wider border'
      : 'inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full text-xs font-semibold tracking-wide border'

  // Green = Secure, Yellow = Sandbox (task spec)
  const variant = isSandbox
    ? 'bg-amber-500/10 border-amber-500/30 text-amber-400'
    : 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400 secure-pulse'

  return (
    <span className={`${base} ${variant} ${className}`}>
      <span
        className={`rounded-full flex-shrink-0 ${size === 'sm' ? 'w-1.5 h-1.5' : 'w-2 h-2'} ${
          isSandbox ? 'bg-amber-400' : 'bg-emerald-400'
        } animate-pulse`}
      />
      {isSandbox ? '⚠ Sandbox Mode' : '🔒 Secure Mode'}
    </span>
  )
}
