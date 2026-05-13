import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Public routes — always allow
  if (pathname === '/' || pathname.startsWith('/login')) {
    return NextResponse.next()
  }

  // Resolve runtime mode: ModeContext writes '_app_mode' cookie on every change
  // and on page load after fetching from the backend. Falls back to the build-time
  // env var for the very first request before the cookie is set.
  const modeCookie = request.cookies.get('_app_mode')?.value
  const envMode =
    process.env.NEXT_PUBLIC_SECURITY_MODE === 'sandbox' ||
    process.env.NEXT_PUBLIC_SECURITY_MODE === 'vulnerable'
      ? 'sandbox'
      : 'secure'
  const mode = modeCookie === 'sandbox' || modeCookie === 'secure' ? modeCookie : envMode

  if (mode === 'sandbox') {
    // TODO: Vulnerability Injection Point
    // Sandbox mode middleware is permissive — only checks cookie existence,
    // no signature validation, no expiry check, easily bypassed.
    const sessionCookie = request.cookies.get('auth-session')
    if (!sessionCookie) {
      const response = NextResponse.next()
      response.headers.set(
        'X-Auth-Warning',
        'No session cookie — sandbox enforcement active'
      )
      return response
    }
    return NextResponse.next()
  }

  // Secure mode: strict cookie check — redirect immediately if missing or wrong value
  const sessionCookie = request.cookies.get('auth-session')
  if (!sessionCookie || sessionCookie.value !== '1') {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('next', pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/dashboard/:path*', '/nasabah/:path*', '/loans/:path*', '/admin/:path*'],
}
