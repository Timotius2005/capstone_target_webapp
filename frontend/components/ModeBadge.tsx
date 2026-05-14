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
      ? 'inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-semibold tracking-wide border'
      : 'inline-flex items-center gap-2 px-2.5 py-1 rounded-md text-xs font-semibold tracking-wide border'

  const variant = isSandbox
    ? 'bg-red-50 border-red-200 text-red-700 dark:bg-red-900/20 dark:border-red-800/60 dark:text-red-400'
    : 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800/60 dark:text-green-400'

  return (
    <span className={`${base} ${variant} ${className}`}>
      <span
        className={`rounded-full flex-shrink-0 ${size === 'sm' ? 'w-1.5 h-1.5' : 'w-2 h-2'} ${
          isSandbox ? 'bg-red-500' : 'bg-green-600 dark:bg-green-400'
        }`}
      />
      {isSandbox ? 'Vulnerable' : 'Secure'}
    </span>
  )
}
