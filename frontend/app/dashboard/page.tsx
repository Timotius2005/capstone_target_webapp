'use client'

import { useState, useEffect } from 'react'
import DashboardLayout from '@/components/DashboardLayout'
import StatCard from '@/components/StatCard'
import DataTable, { type Column } from '@/components/DataTable'
import { api } from '@/services/api'
import { authService } from '@/services/auth'
import { useMode } from '@/contexts/ModeContext'

interface Loan {
  id: string
  nasabah_id: string
  amount: number
  interest_rate: number
  term_months: number
  status: string
  approved_at: string | null
  created_at: string
}

interface Nasabah {
  id: string
  user_id: string
  name: string
  nik: string
  phone: string
  address: string
  created_at: string
}

const STATUS_STYLES: Record<string, string> = {
  active:   'bg-emerald-500/15 text-emerald-400',
  approved: 'bg-emerald-500/15 text-emerald-400',
  pending:  'bg-amber-500/15 text-amber-400',
  rejected: 'bg-red-500/15 text-red-400',
  closed:   'bg-slate-500/15 text-slate-400',
}

function formatIDR(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const LOAN_COLUMNS: Column[] = [
  {
    key: 'id',
    label: 'Loan ID',
    sensitiveInSecure: true,
    render: (v) => (
      <span className="font-mono text-xs text-slate-400">{String(v).slice(0, 8)}…</span>
    ),
  },
  {
    key: 'nasabah_id',
    label: 'Nasabah ID',
    sensitiveInSecure: true,
    render: (v) => (
      <span className="font-mono text-xs text-red-400">{String(v).slice(0, 8)}…</span>
    ),
  },
  {
    key: 'amount',
    label: 'Jumlah',
    render: (v) => <span className="font-semibold">{formatIDR(v as number)}</span>,
  },
  {
    key: 'interest_rate',
    label: 'Bunga',
    render: (v) => <span>{(v as number).toFixed(1)}%</span>,
  },
  {
    key: 'term_months',
    label: 'Tenor',
    render: (v) => <span>{String(v)} bln</span>,
  },
  {
    key: 'status',
    label: 'Status',
    render: (v) => {
      const s = String(v).toLowerCase()
      return (
        <span className={`status-badge ${STATUS_STYLES[s] || 'bg-slate-500/15 text-slate-400'}`}>
          <span className="w-1.5 h-1.5 rounded-full bg-current" />
          {s}
        </span>
      )
    },
  },
  {
    key: 'created_at',
    label: 'Dibuat',
    render: (v) => (
      <span className="text-slate-400 text-xs">
        {new Date(v as string).toLocaleDateString('id-ID')}
      </span>
    ),
  },
]

export default function DashboardPage() {
  const [loans, setLoans] = useState<Loan[]>([])
  const [nasabahCount, setNasabahCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const [rawLoans, setRawLoans] = useState<unknown>(null)
  const [rawNasabah, setRawNasabah] = useState<unknown>(null)

  const { mode } = useMode()
  const vulnerable = mode === 'sandbox'
  const user = authService.getUser()

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [loansRes, nasabahRes] = await Promise.all([
          api.get<Loan[] | { data: Loan[] }>('/api/v1/loans'),
          api.get<Nasabah[] | { data: Nasabah[] }>('/api/v1/nasabah'),
        ])

        // Backend returns { data: [...], total: N } — unwrap the array
        const loansArray = Array.isArray(loansRes.data)
          ? loansRes.data
          : (loansRes.data as { data?: Loan[] }).data ?? []
        const nasabahArray = Array.isArray(nasabahRes.data)
          ? nasabahRes.data
          : (nasabahRes.data as { data?: Nasabah[] }).data ?? []

        setLoans(loansArray)
        setNasabahCount(nasabahArray.length)

        if (vulnerable) {
          // TODO: Vulnerability Injection Point
          // Full raw API responses stored for debug panel rendering
          setRawLoans(loansRes.data)
          setRawNasabah(nasabahRes.data)  // intentionally stores full wrapper incl. metadata
        }
      } catch {
        // Secure: swallow error silently; Vulnerable: logged by axios interceptor
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [vulnerable])

  const activeLoans = loans.filter((l) => l.status === 'active' || l.status === 'approved')
  const pendingLoans = loans.filter((l) => l.status === 'pending')
  const totalPortfolio = loans.reduce((sum, l) => sum + l.amount, 0)

  const recentLoans = [...loans]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 8)

  return (
    <DashboardLayout title="Dashboard">
      {/* ── Page header ────────────────────────── */}
      <div className="mb-5 flex items-start justify-between gap-4 flex-wrap">
        <div>
          <p className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-0.5">
            Loan Management System
          </p>
          <h2 className="text-lg font-bold text-slate-800 dark:text-white tracking-tight">
            Dashboard Overview
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Selamat datang,{' '}
            <span className="font-semibold text-slate-600 dark:text-slate-300">{user?.username || 'Pengguna'}</span>
            {user?.role && (
              <span className="ml-1.5 text-slate-400">({user.role})</span>
            )}
          </p>
        </div>
        {/* TODO: Vulnerability Injection Point */}
        {/* Vulnerable mode: expose user internal ID on dashboard */}
        {vulnerable && (user as { id?: string })?.id && (
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900">
            <span className="w-1.5 h-1.5 rounded-full bg-red-500 animate-pulse" />
            <span className="text-red-700 dark:text-red-400 text-[10px] font-mono">
              [VULN] User UUID: {(user as { id?: string }).id}
            </span>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 mb-6">
        <StatCard
          title="Total Nasabah"
          value={loading ? '—' : nasabahCount.toLocaleString('id-ID')}
          subtitle="Terdaftar dalam sistem"
          icon="👥"
          iconBg="bg-gradient-to-br from-blue-500 to-blue-600"
          trend="+8.2%"
          trendUp
        />
        <StatCard
          title="Total Pinjaman"
          value={loading ? '—' : loans.length.toLocaleString('id-ID')}
          subtitle="Semua status"
          icon="📋"
          iconBg="bg-gradient-to-br from-indigo-500 to-indigo-600"
          trend="+12.4%"
          trendUp
        />
        <StatCard
          title="Pinjaman Aktif"
          value={loading ? '—' : activeLoans.length.toLocaleString('id-ID')}
          subtitle={`${pendingLoans.length} menunggu persetujuan`}
          icon="✅"
          iconBg="bg-gradient-to-br from-emerald-500 to-emerald-600"
          trend="+5.1%"
          trendUp
        />
        <StatCard
          title="Total Portfolio"
          value={loading ? '—' : formatIDR(totalPortfolio)}
          subtitle="Nilai pinjaman aktif"
          icon="💼"
          iconBg="bg-gradient-to-br from-violet-500 to-violet-600"
          trend="+18.7%"
          trendUp
        />
      </div>

      {/* ── Portfolio Distribution ──────────────── */}
      {!loading && loans.length > 0 && (
        <>
          <p className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-3">
            Portfolio Distribution
          </p>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
            {[
              {
                label: 'Aktif / Disetujui',
                count: activeLoans.length,
                color: 'from-emerald-500/20 to-emerald-500/5',
                accent: 'bg-emerald-400',
                pct: loans.length ? Math.round((activeLoans.length / loans.length) * 100) : 0,
              },
              {
                label: 'Menunggu',
                count: pendingLoans.length,
                color: 'from-amber-500/20 to-amber-500/5',
                accent: 'bg-amber-400',
                pct: loans.length ? Math.round((pendingLoans.length / loans.length) * 100) : 0,
              },
              {
                label: 'Ditolak',
                count: loans.filter((l) => l.status === 'rejected').length,
                color: 'from-red-500/20 to-red-500/5',
                accent: 'bg-red-400',
                pct: loans.length
                  ? Math.round(
                      (loans.filter((l) => l.status === 'rejected').length / loans.length) * 100
                    )
                  : 0,
              },
            ].map((item) => (
              <div
                key={item.label}
                className={`enterprise-card rounded-lg p-5 bg-gradient-to-br ${item.color}`}
              >
                <div className="flex items-center justify-between mb-3">
                  <p className="text-sm font-medium text-slate-600 dark:text-slate-300">{item.label}</p>
                  <span className="text-2xl font-bold text-slate-800 dark:text-white">{item.count}</span>
                </div>
                <div className="w-full bg-slate-200/50 dark:bg-slate-700/50 rounded-full h-1.5">
                  <div
                    className={`h-1.5 rounded-full ${item.accent}`}
                    style={{ width: `${item.pct}%` }}
                  />
                </div>
                <p className="text-xs text-slate-400 mt-1.5">{item.pct}% dari total</p>
              </div>
            ))}
          </div>
        </>
      )}

      {/* ── Recent Activity ─────────────────────── */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <p className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
            Recent Loan Activity
          </p>
          {vulnerable && (
            <span className="text-red-500 text-[10px] font-mono bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 px-2 py-0.5 rounded">
              [VULN] ID columns exposed
            </span>
          )}
        </div>

        <DataTable
          columns={LOAN_COLUMNS}
          data={recentLoans as unknown as Record<string, unknown>[]}
          loading={loading}
          emptyMessage="Belum ada data pinjaman."
          rawApiData={vulnerable ? rawLoans : undefined}
        />

        {/* TODO: Vulnerability Injection Point */}
        {/* Nasabah raw data also exposed in vulnerable mode */}
        {vulnerable && rawNasabah != null && (
          <div className="mt-4 enterprise-card rounded-lg p-4 border-l-4 border-red-500 bg-red-50 dark:bg-red-950/20">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
                <span className="text-red-700 dark:text-red-400 text-[10px] font-semibold uppercase tracking-widest">
                  Debug: Raw Nasabah API Data [Vulnerability Injection Point]
                </span>
              </div>
            </div>
            <pre className="text-xs text-red-700/70 dark:text-red-300/70 overflow-x-auto max-h-48 font-mono leading-relaxed">
              {JSON.stringify(rawNasabah, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
