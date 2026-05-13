'use client'

import type { ReactNode } from 'react'
import { isVulnerable } from '@/utils/securityMode'

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
  <div className="space-y-3 p-6">
    {Array.from({ length: 5 }).map((_, i) => (
      <div key={i} className="h-10 bg-slate-200 dark:bg-slate-700/50 rounded-lg animate-pulse" />
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
  const vulnerable = isVulnerable()

  const visibleColumns = columns.filter(
    (c) => !(c.sensitiveInSecure && !vulnerable)
  )

  return (
    <div className="space-y-4">
      <div className="glass-card rounded-2xl overflow-hidden">
        {loading ? (
          <Skeleton />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-slate-200/30 dark:border-slate-700/40">
                  {visibleColumns.map((col) => (
                    <th
                      key={col.key}
                      className="px-6 py-4 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-widest whitespace-nowrap"
                    >
                      {col.label}
                      {vulnerable && col.sensitiveInSecure && (
                        <span className="ml-1.5 text-red-400 font-normal normal-case tracking-normal">
                          [exposed]
                        </span>
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200/20 dark:divide-slate-700/30">
                {data.length === 0 ? (
                  <tr>
                    <td
                      colSpan={visibleColumns.length}
                      className="px-6 py-16 text-center text-slate-400 text-sm"
                    >
                      <div className="flex flex-col items-center gap-2">
                        <span className="text-3xl">📭</span>
                        {emptyMessage}
                      </div>
                    </td>
                  </tr>
                ) : (
                  data.map((row, idx) => (
                    <tr
                      key={idx}
                      className="hover:bg-slate-50/80 dark:hover:bg-slate-800/30 transition-colors duration-150"
                    >
                      {visibleColumns.map((col) => (
                        <td
                          key={col.key}
                          className="px-6 py-3.5 text-sm text-slate-700 dark:text-slate-300 whitespace-nowrap"
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
        <div className="glass-card rounded-2xl p-4 border border-red-500/25 bg-red-500/5">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 bg-red-400 rounded-full animate-pulse" />
              <span className="text-red-400 text-xs font-semibold uppercase tracking-wider">
                Debug: Full API Response
              </span>
            </div>
            <span className="text-red-400/50 text-[10px] font-mono">
              [Vulnerability Injection Point]
            </span>
          </div>
          <pre className="text-xs text-red-300/70 overflow-x-auto max-h-56 leading-relaxed font-mono">
            {JSON.stringify(rawApiData, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
