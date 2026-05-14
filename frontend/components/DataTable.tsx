'use client'

import type { ReactNode } from 'react'
import { useMode } from '@/contexts/ModeContext'

export interface Column {
  key: string
  label: string
  sensitiveInSecure?: boolean
  render?: (value: unknown, row: Record<string, unknown>) => ReactNode
}

interface DataTableProps {
  columns: Column[]
  data: Record<string, unknown>[]
  loading?: boolean
  emptyMessage?: string
  rawApiData?: unknown
}

const Skeleton = () => (
  <div className="space-y-2 p-5">
    {Array.from({ length: 5 }).map((_, i) => (
      <div key={i} className="h-9 bg-slate-100 dark:bg-slate-800 rounded animate-pulse" />
    ))}
  </div>
)

export default function DataTable({
  columns,
  data,
  loading = false,
  emptyMessage = 'Tidak ada data tersedia.',
  rawApiData,
}: DataTableProps) {
  const { mode } = useMode()
  const vulnerable = mode === 'sandbox'

  const rows: Record<string, unknown>[] = Array.isArray(data)
    ? data
    : Array.isArray((data as unknown as { data?: unknown }).data)
      ? ((data as unknown as { data: Record<string, unknown>[] }).data)
      : []

  const visibleColumns = columns.filter(
    (c) => !(c.sensitiveInSecure && !vulnerable)
  )

  return (
    <div className="space-y-4">
      <div className="enterprise-card rounded-lg overflow-hidden">
        {loading ? (
          <Skeleton />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/60">
                  {visibleColumns.map((col) => (
                    <th
                      key={col.key}
                      className="px-5 py-3 text-left text-[11px] font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-widest whitespace-nowrap"
                    >
                      {col.label}
                      {vulnerable && col.sensitiveInSecure && (
                        <span className="ml-1.5 text-red-500 font-normal normal-case tracking-normal text-[10px]">
                          [exposed]
                        </span>
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                {rows.length === 0 ? (
                  <tr>
                    <td
                      colSpan={visibleColumns.length}
                      className="px-5 py-12 text-center text-slate-400 text-sm"
                    >
                      <div className="flex flex-col items-center gap-2">
                        <span className="text-2xl opacity-40">—</span>
                        {emptyMessage}
                      </div>
                    </td>
                  </tr>
                ) : (
                  rows.map((row, idx) => (
                    <tr
                      key={idx}
                      className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors duration-100"
                    >
                      {visibleColumns.map((col) => (
                        <td
                          key={col.key}
                          className="px-5 py-3 text-sm text-slate-700 dark:text-slate-300 whitespace-nowrap"
                        >
                          {col.render
                            ? col.render(row[col.key], row)
                            : String(row[col.key] ?? '—')}
                        </td>
                      ))}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* TODO: Vulnerability Injection Point */}
      {/* Debug panel — exposes raw API response with hidden/sensitive fields */}
      {vulnerable && rawApiData != null && (
        <div className="enterprise-card rounded-lg p-4 border-l-4 border-red-500 bg-red-50 dark:bg-red-950/20 dark:border-red-900">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
              <span className="text-red-700 dark:text-red-400 text-xs font-semibold uppercase tracking-wider">
                Debug: Full API Response
              </span>
            </div>
            <span className="text-red-400/60 text-[10px] font-mono">
              [Vulnerability Injection Point]
            </span>
          </div>
          <pre className="text-xs text-red-700/70 dark:text-red-300/70 overflow-x-auto max-h-56 leading-relaxed font-mono">
            {JSON.stringify(rawApiData, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
