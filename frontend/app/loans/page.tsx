'use client'

import { useState, useEffect, useCallback } from 'react'
import DashboardLayout from '@/components/DashboardLayout'
import DataTable, { type Column } from '@/components/DataTable'
import { api } from '@/services/api'
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
  updated_at: string
}

type StatusFilter = 'all' | 'active' | 'pending' | 'rejected' | 'closed'

function formatIDR(n: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency', currency: 'IDR',
    minimumFractionDigits: 0, maximumFractionDigits: 0,
  }).format(n)
}

const STATUS_STYLES: Record<string, string> = {
  active:   'bg-emerald-500/15 text-emerald-400 border-emerald-500/20',
  approved: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/20',
  pending:  'bg-amber-500/15  text-amber-400  border-amber-500/20',
  rejected: 'bg-red-500/15    text-red-400    border-red-500/20',
  closed:   'bg-slate-500/15  text-slate-400  border-slate-500/20',
}

const TAB_LABELS: { key: StatusFilter; label: string }[] = [
  { key: 'all', label: 'Semua' },
  { key: 'active', label: 'Aktif' },
  { key: 'pending', label: 'Menunggu' },
  { key: 'rejected', label: 'Ditolak' },
  { key: 'closed', label: 'Selesai' },
]

export default function LoansPage() {
  const [loans, setLoans] = useState<Loan[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<StatusFilter>('all')
  const [rawData, setRawData] = useState<unknown>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [formData, setFormData] = useState({
    nasabah_id: '', amount: '', interest_rate: '', term_months: ''
  })
  const [formError, setFormError] = useState('')
  const [formSuccess, setFormSuccess] = useState('')

  const { mode } = useMode()
  const vulnerable = mode === 'sandbox'

  const fetchLoans = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<Loan[] | { data: Loan[] }>('/api/v1/loans')

      // Backend returns { data: [...], total: N } — unwrap the array
      const list: Loan[] = Array.isArray(res.data)
        ? res.data
        : (res.data as { data?: Loan[] }).data ?? []

      setLoans(list)
      if (vulnerable) {
        // TODO: Vulnerability Injection Point
        // Full loan data including internal IDs exposed in debug panel
        setRawData(res.data)
      }
    } catch {
      // Secure: silent; Vulnerable: logged by interceptor
    } finally {
      setLoading(false)
    }
  }, [vulnerable])

  useEffect(() => { fetchLoans() }, [fetchLoans])

  const filtered =
    filter === 'all'
      ? loans
      : loans.filter((l) => l.status === filter || (filter === 'active' && l.status === 'approved'))

  const validateForm = (): boolean => {
    if (vulnerable) return true // TODO: Vulnerability Injection Point — disabled validation
    if (!formData.nasabah_id.trim()) { setFormError('Nasabah ID diperlukan.'); return false }
    const amount = parseFloat(formData.amount)
    if (isNaN(amount) || amount < 1_000_000) {
      setFormError('Jumlah pinjaman minimal Rp 1.000.000.')
      return false
    }
    const rate = parseFloat(formData.interest_rate)
    if (isNaN(rate) || rate <= 0 || rate > 50) {
      setFormError('Bunga harus antara 0.1% - 50%.')
      return false
    }
    const term = parseInt(formData.term_months)
    if (isNaN(term) || term < 1 || term > 360) {
      setFormError('Tenor harus antara 1-360 bulan.')
      return false
    }
    return true
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    setFormSuccess('')
    if (!validateForm()) return
    try {
      await api.post('/api/v1/loans', {
        nasabah_id: formData.nasabah_id,
        amount: parseFloat(formData.amount),
        interest_rate: parseFloat(formData.interest_rate),
        term_months: parseInt(formData.term_months),
      })
      setFormSuccess('Pinjaman berhasil dibuat.')
      setFormData({ nasabah_id: '', amount: '', interest_rate: '', term_months: '' })
      setShowCreateForm(false)
      fetchLoans()
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string } } }
      if (vulnerable) {
        // TODO: Vulnerability Injection Point — verbose server error
        setFormError(axiosErr.response?.data?.error || 'Server error')
      } else {
        setFormError('Gagal membuat pinjaman. Silakan coba lagi.')
      }
    }
  }

  const COLUMNS: Column[] = [
    {
      key: 'id',
      label: 'Loan UUID',
      sensitiveInSecure: true,
      render: (v) => (
        <span className="font-mono text-xs text-red-400">{String(v).slice(0, 8)}…</span>
      ),
    },
    {
      key: 'nasabah_id',
      label: 'Nasabah UUID',
      sensitiveInSecure: true,
      render: (v) => (
        <span className="font-mono text-xs text-red-400">{String(v).slice(0, 8)}…</span>
      ),
    },
    {
      key: 'amount',
      label: 'Jumlah Pinjaman',
      render: (v) => (
        <span className="font-bold text-slate-800 dark:text-white">{formatIDR(v as number)}</span>
      ),
    },
    {
      key: 'interest_rate',
      label: 'Bunga (% / thn)',
      render: (v) => <span>{(v as number).toFixed(2)}%</span>,
    },
    {
      key: 'term_months',
      label: 'Tenor',
      render: (v) => <span>{String(v)} bulan</span>,
    },
    {
      key: 'status',
      label: 'Status',
      render: (v) => {
        const s = String(v).toLowerCase()
        const style = STATUS_STYLES[s] || 'bg-slate-500/15 text-slate-400 border-slate-500/20'
        return (
          <span className={`status-badge border ${style}`}>
            <span className="w-1.5 h-1.5 rounded-full bg-current" />
            {s}
          </span>
        )
      },
    },
    {
      key: 'approved_at',
      label: 'Disetujui',
      render: (v) =>
        v ? (
          <span className="text-xs text-emerald-400">
            {new Date(v as string).toLocaleDateString('id-ID')}
          </span>
        ) : (
          <span className="text-xs text-slate-400">—</span>
        ),
    },
    {
      key: 'created_at',
      label: 'Dibuat',
      render: (v) => (
        <span className="text-xs text-slate-400">
          {new Date(v as string).toLocaleDateString('id-ID')}
        </span>
      ),
    },
  ]

  const totalAmount = filtered.reduce((s, l) => s + l.amount, 0)

  return (
    <DashboardLayout title="Manajemen Pinjaman">
      {/* ── Header ─────────────────────────────── */}
      <div className="flex items-center justify-between mb-6 flex-wrap gap-3">
        <div>
          <h2 className="text-xl font-bold text-slate-800 dark:text-white">
            Manajemen Pinjaman
          </h2>
          <p className="text-sm text-slate-400 mt-0.5">
            {loading ? '…' : `${filtered.length} pinjaman · Total ${formatIDR(totalAmount)}`}
            {vulnerable && (
              <span className="ml-2 text-red-400 text-xs">[UUID & data internal exposed]</span>
            )}
          </p>
        </div>
        <button
          onClick={() => setShowCreateForm((v) => !v)}
          className="px-4 py-2 animated-gradient text-white text-sm font-semibold rounded-xl shadow hover:shadow-glow hover:scale-[1.02] active:scale-[0.98] transition-all duration-200"
        >
          + Buat Pinjaman
        </button>
      </div>

      {/* ── Create form ──────────────────────────── */}
      {showCreateForm && (
        <div className="glass-card rounded-2xl p-6 mb-6 animate-fade-in border border-indigo-500/20">
          <h3 className="text-base font-bold text-slate-800 dark:text-white mb-4">
            Form Pengajuan Pinjaman
          </h3>

          {vulnerable && (
            <div className="mb-4 p-3 rounded-xl bg-red-500/8 border border-red-500/20">
              <p className="text-red-400 text-xs">
                ⚠ [VULN] Tidak ada validasi · Tidak ada rate limiting · Nasabah ID bisa ditebak
              </p>
              {/* TODO: Vulnerability Injection Point */}
            </div>
          )}

          {formError && (
            <div className="mb-4 p-3 rounded-xl border border-red-500/30 bg-red-500/10 text-red-400 text-xs font-mono">
              {formError}
            </div>
          )}
          {formSuccess && (
            <div className="mb-4 p-3 rounded-xl border border-emerald-500/30 bg-emerald-500/10 text-emerald-400 text-xs">
              {formSuccess}
            </div>
          )}

          <form onSubmit={handleCreate} noValidate={vulnerable}>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="sm:col-span-2">
                <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1.5 uppercase tracking-wider">
                  Nasabah ID (UUID)
                  {/* TODO: Vulnerability Injection Point */}
                  {/* Vulnerable: UUID field directly exposed in form */}
                  {vulnerable && (
                    <span className="ml-2 text-red-400 font-normal normal-case tracking-normal">
                      [VULN: IDOR risk — predictable ID field]
                    </span>
                  )}
                </label>
                <input
                  type="text"
                  value={formData.nasabah_id}
                  onChange={(e) => setFormData((d) => ({ ...d, nasabah_id: e.target.value }))}
                  placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                  required={!vulnerable}
                  className="input-field font-mono text-sm"
                />
              </div>
              {[
                { label: 'Jumlah Pinjaman (IDR)', key: 'amount', placeholder: '10000000' },
                { label: 'Suku Bunga (% / tahun)', key: 'interest_rate', placeholder: '12.5' },
                { label: 'Tenor (bulan)', key: 'term_months', placeholder: '24' },
              ].map((f) => (
                <div key={f.key}>
                  <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1.5 uppercase tracking-wider">
                    {f.label}
                  </label>
                  <input
                    type="number"
                    value={formData[f.key as keyof typeof formData]}
                    onChange={(e) => setFormData((d) => ({ ...d, [f.key]: e.target.value }))}
                    placeholder={f.placeholder}
                    required={!vulnerable}
                    className="input-field"
                  />
                </div>
              ))}
            </div>
            <div className="flex gap-3 mt-5">
              <button
                type="submit"
                className="px-5 py-2 animated-gradient text-white text-sm font-semibold rounded-xl hover:shadow-glow transition-all"
              >
                Ajukan Pinjaman
              </button>
              <button
                type="button"
                onClick={() => setShowCreateForm(false)}
                className="px-5 py-2 text-sm font-semibold text-slate-500 dark:text-slate-400 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              >
                Batal
              </button>
            </div>
          </form>
        </div>
      )}

      {/* ── Status tabs ──────────────────────────── */}
      <div className="flex gap-1 mb-4 p-1 glass-card rounded-xl w-fit">
        {TAB_LABELS.map((tab) => {
          const count =
            tab.key === 'all'
              ? loans.length
              : loans.filter(
                  (l) =>
                    l.status === tab.key ||
                    (tab.key === 'active' && l.status === 'approved')
                ).length
          return (
            <button
              key={tab.key}
              onClick={() => setFilter(tab.key)}
              className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all duration-200 ${
                filter === tab.key
                  ? 'bg-indigo-600 text-white shadow'
                  : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-white'
              }`}
            >
              {tab.label}
              <span
                className={`ml-1.5 text-[10px] ${
                  filter === tab.key ? 'text-white/70' : 'text-slate-400'
                }`}
              >
                {count}
              </span>
            </button>
          )
        })}
      </div>

      {/* ── Table ────────────────────────────────── */}
      <DataTable
        columns={COLUMNS}
        data={filtered as unknown as Record<string, unknown>[]}
        loading={loading}
        emptyMessage="Tidak ada pinjaman ditemukan."
        rawApiData={vulnerable ? rawData : undefined}
      />
    </DashboardLayout>
  )
}
