'use client'

import { useState, useEffect, useCallback } from 'react'
import DashboardLayout from '@/components/DashboardLayout'
import DataTable, { type Column } from '@/components/DataTable'
import { api } from '@/services/api'
import { useMode } from '@/contexts/ModeContext'

interface Nasabah {
  id: string
  user_id: string
  full_name: string
  nik: string
  phone: string
  address: string
  date_of_birth: string
  created_at: string
  updated_at: string
}

function maskNIK(nik: string): string {
  if (nik.length < 8) return '••••••••'
  return nik.slice(0, 4) + '••••••••' + nik.slice(-4)
}

export default function NasabahPage() {
  const [nasabah, setNasabah] = useState<Nasabah[]>([])
  const [filtered, setFiltered] = useState<Nasabah[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [rawData, setRawData] = useState<unknown>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [formData, setFormData] = useState({
    full_name: '', nik: '', phone: '', address: '', date_of_birth: ''
  })
  const [formError, setFormError] = useState('')
  const [formSuccess, setFormSuccess] = useState('')

  const { mode } = useMode()
  const vulnerable = mode === 'sandbox'

  const fetchNasabah = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<Nasabah[] | { data: Nasabah[] }>('/api/v1/nasabah')

      // Backend returns { data: [...], total: N } — unwrap the array
      const list: Nasabah[] = Array.isArray(res.data)
        ? res.data
        : (res.data as { data?: Nasabah[] }).data ?? []

      setNasabah(list)
      setFiltered(list)
      if (vulnerable) {
        // TODO: Vulnerability Injection Point
        // Full nasabah data (NIK, address, phone) exposed in debug panel
        setRawData(res.data)
      }
    } catch {
      // Secure: silent fail; Vulnerable: logged by interceptor
    } finally {
      setLoading(false)
    }
  }, [vulnerable])

  useEffect(() => { fetchNasabah() }, [fetchNasabah])

  useEffect(() => {
    const q = search.toLowerCase()
    setFiltered(
      nasabah.filter(
        (n) =>
          (n.full_name || '').toLowerCase().includes(q) ||
          n.phone.includes(q) ||
          (vulnerable ? n.nik.includes(q) : maskNIK(n.nik).includes(q))
      )
    )
  }, [search, nasabah, vulnerable])

  const validateForm = (): boolean => {
    if (vulnerable) return true // TODO: Vulnerability Injection Point — no validation
    if (!formData.full_name.trim() || formData.full_name.length < 2) {
      setFormError('Nama minimal 2 karakter.')
      return false
    }
    if (!/^\d{16}$/.test(formData.nik)) {
      setFormError('NIK harus 16 digit angka.')
      return false
    }
    if (!/^\+?[\d\s-]{8,15}$/.test(formData.phone)) {
      setFormError('Format nomor telepon tidak valid.')
      return false
    }
    return true
  }

  const handleAddNasabah = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    setFormSuccess('')
    if (!validateForm()) return
    try {
      await api.post('/api/v1/nasabah', formData)
      setFormSuccess('Nasabah berhasil ditambahkan.')
      setFormData({ full_name: '', nik: '', phone: '', address: '', date_of_birth: '' })
      setShowAddForm(false)
      fetchNasabah()
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string } } }
      if (vulnerable) {
        // TODO: Vulnerability Injection Point — server error exposed verbatim
        setFormError(axiosErr.response?.data?.error || 'Server error')
      } else {
        setFormError('Gagal menambahkan nasabah. Silakan coba lagi.')
      }
    }
  }

  const COLUMNS: Column[] = [
    {
      key: 'id',
      label: 'UUID',
      sensitiveInSecure: true,
      render: (v) => (
        <span className="font-mono text-xs text-red-400">{String(v).slice(0, 8)}…</span>
      ),
    },
    {
      key: 'user_id',
      label: 'User ID',
      sensitiveInSecure: true,
      render: (v) => (
        <span className="font-mono text-xs text-red-400">{String(v).slice(0, 8)}…</span>
      ),
    },
    {
      key: 'full_name',
      label: 'Nama Nasabah',
      render: (v) => (
        <span className="font-semibold text-slate-800 dark:text-white">{String(v)}</span>
      ),
    },
    {
      key: 'nik',
      label: 'NIK',
      render: (v) => (
        <span className={`font-mono text-xs ${vulnerable ? 'text-red-400' : 'text-slate-400'}`}>
          {/* TODO: Vulnerability Injection Point — full NIK exposed in vulnerable mode */}
          {vulnerable ? String(v) : maskNIK(String(v))}
        </span>
      ),
    },
    { key: 'phone', label: 'Telepon' },
    {
      key: 'address',
      label: 'Alamat',
      render: (v) => (
        <span className="max-w-[160px] truncate block text-slate-400 text-xs">
          {String(v)}
        </span>
      ),
    },
    {
      key: 'created_at',
      label: 'Terdaftar',
      render: (v) => (
        <span className="text-xs text-slate-400">
          {new Date(v as string).toLocaleDateString('id-ID')}
        </span>
      ),
    },
  ]

  return (
    <DashboardLayout title="Data Nasabah">
      {/* ── Page header ────────────────────────── */}
      <div className="flex items-center justify-between mb-6 flex-wrap gap-3">
        <div>
          <h2 className="text-xl font-bold text-slate-800 dark:text-white">
            Manajemen Nasabah
          </h2>
          <p className="text-sm text-slate-400 mt-0.5">
            {loading ? '…' : `${nasabah.length} nasabah terdaftar`}
            {vulnerable && (
              <span className="ml-2 text-red-400 text-xs">[NIK & UUID exposed]</span>
            )}
          </p>
        </div>
        <button
          onClick={() => setShowAddForm((v) => !v)}
          className="px-4 py-2 animated-gradient text-white text-sm font-semibold rounded-xl shadow hover:shadow-glow hover:scale-[1.02] active:scale-[0.98] transition-all duration-200"
        >
          + Tambah Nasabah
        </button>
      </div>

      {/* ── Add nasabah form ─────────────────────── */}
      {showAddForm && (
        <div className="glass-card rounded-2xl p-6 mb-6 animate-fade-in border border-indigo-500/20">
          <h3 className="text-base font-bold text-slate-800 dark:text-white mb-4">
            Form Tambah Nasabah
          </h3>

          {vulnerable && (
            <div className="mb-4 p-3 rounded-xl bg-red-500/8 border border-red-500/20">
              <p className="text-red-400 text-xs">
                ⚠ [VULN] Validasi input dinonaktifkan · Data dikirim tanpa sanitasi
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

          <form onSubmit={handleAddNasabah} noValidate={vulnerable}>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {[
                { label: 'Nama Lengkap', key: 'name', type: 'text', placeholder: 'Budi Santoso' },
                { label: 'NIK (16 digit)', key: 'nik', type: 'text', placeholder: '3201234567890001' },
                { label: 'No. Telepon', key: 'phone', type: 'tel', placeholder: '+6281234567890' },
                { label: 'Tanggal Lahir', key: 'date_of_birth', type: 'date', placeholder: '' },
              ].map((f) => (
                <div key={f.key}>
                  <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1.5 uppercase tracking-wider">
                    {f.label}
                  </label>
                  <input
                    type={f.type}
                    value={formData[f.key as keyof typeof formData]}
                    onChange={(e) => setFormData((d) => ({ ...d, [f.key]: e.target.value }))}
                    placeholder={f.placeholder}
                    required={!vulnerable}
                    className="input-field"
                  />
                </div>
              ))}
              <div className="sm:col-span-2">
                <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1.5 uppercase tracking-wider">
                  Alamat
                </label>
                <textarea
                  value={formData.address}
                  onChange={(e) => setFormData((d) => ({ ...d, address: e.target.value }))}
                  placeholder="Jl. Merdeka No. 1, Jakarta Pusat"
                  rows={2}
                  required={!vulnerable}
                  className="input-field resize-none"
                />
              </div>
            </div>
            <div className="flex gap-3 mt-5">
              <button
                type="submit"
                className="px-5 py-2 animated-gradient text-white text-sm font-semibold rounded-xl hover:shadow-glow transition-all"
              >
                Simpan
              </button>
              <button
                type="button"
                onClick={() => setShowAddForm(false)}
                className="px-5 py-2 text-sm font-semibold text-slate-500 dark:text-slate-400 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              >
                Batal
              </button>
            </div>
          </form>
        </div>
      )}

      {/* ── Search ───────────────────────────────── */}
      <div className="mb-4">
        <div className="relative max-w-sm">
          <svg
            className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400"
            fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
          </svg>
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari nama, telepon, NIK…"
            className="input-field pl-10"
          />
        </div>
      </div>

      {/* ── Table ────────────────────────────────── */}
      <DataTable
        columns={COLUMNS}
        data={filtered as unknown as Record<string, unknown>[]}
        loading={loading}
        emptyMessage="Tidak ada nasabah ditemukan."
        rawApiData={vulnerable ? rawData : undefined}
      />
    </DashboardLayout>
  )
}
