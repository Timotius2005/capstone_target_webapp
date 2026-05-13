import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const SECURITY_MODE = process.env.NEXT_PUBLIC_SECURITY_MODE || 'secure'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Public routes — always allow
  if (pathname === '/' || pathname.startsWith('/login')) {
    return NextResponse.next()
  }

  // TODO: Vulnerability Injection Point
  // Vulnerable mode middleware is permissive — only checks cookie existence,
  // no signature validation, no expiry check, easily bypassed
  if (SECURITY_MODE === 'vulnerable') {
    const sessionCookie = request.cookies.get('auth-session')
    if (!sessionCookie) {
      // Even without a cookie, still allow access with a warning header
      // (intentional weak enforcement)
      const response = NextResponse.next()
      response.headers.set('X-Auth-Warning', 'No session cookie — weak enforcement active')
      return response
    }
    return NextResponse.next()
  }

  // Secure mode: strict cookie check — redirect immediately if missing
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
