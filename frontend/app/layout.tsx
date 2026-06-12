import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { ThemeProvider } from '@/components/ThemeProvider'
import { ModeProvider } from '@/contexts/ModeContext'
import GlobalModeSwitcher from '@/components/GlobalModeSwitcher'

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' })

export const metadata: Metadata = {
  title: 'PT. Dana Sejahtera — Fintech System',
  description: 'Sistem Manajemen Pinjaman PT. Dana Sejahtera',
}

// URL backend API (disuntikkan saat build via NEXT_PUBLIC_API_URL).
// Ditaruh sebagai preconnect/dns-prefetch di <head> agar koneksi ke backend
// lebih cepat — sekaligus membuat origin backend muncul di HTML awal yang
// di-render server (relevan untuk skenario hybrid cloud).
const API_ORIGIN = process.env.NEXT_PUBLIC_API_URL || ''

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id" suppressHydrationWarning>
      <head>
        {API_ORIGIN && (
          <>
            <link rel="preconnect" href={API_ORIGIN} crossOrigin="anonymous" />
            <link rel="dns-prefetch" href={API_ORIGIN} />
          </>
        )}
      </head>
      <body className={`${inter.variable} font-sans antialiased pt-8`} suppressHydrationWarning>
        <ThemeProvider>
          <ModeProvider>
            {children}
            <GlobalModeSwitcher />
          </ModeProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
