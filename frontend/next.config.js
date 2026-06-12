/** @type {import('next').NextConfig} */

// Server-side URL used by Next.js for API rewrites (SSR proxy).
// In Docker, set BACKEND_URL=http://backend:8080 so the frontend container
// can reach the backend container by service name.
//
// On Vercel (hybrid-cloud), DON'T set BACKEND_URL. The frontend calls the
// backend directly via NEXT_PUBLIC_API_URL (the ngrok static domain), so the
// rewrites are unnecessary — and disabling them avoids accidental proxying to
// a non-existent localhost:8080 inside Vercel's serverless runtime.
const BACKEND_URL = process.env.BACKEND_URL

const nextConfig = {
  // Produce a minimal standalone build for Docker — drops node_modules from image.
  output: 'standalone',

  async rewrites() {
    // Only proxy /api/* when BACKEND_URL is explicitly provided (Docker case).
    if (!BACKEND_URL) return []
    return [
      {
        source: '/api/:path*',
        destination: `${BACKEND_URL}/api/:path*`,
      },
      {
        source: '/config/:path*',
        destination: `${BACKEND_URL}/config/:path*`,
      },
    ]
  },
}

module.exports = nextConfig
