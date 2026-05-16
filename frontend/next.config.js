/** @type {import('next').NextConfig} */

// Server-side URL used by Next.js for API rewrites (SSR proxy).
// In Docker, set BACKEND_URL=http://backend:8080 so the frontend container
// can reach the backend container by service name.
// Locally defaults to localhost:8080.
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080'

const nextConfig = {
  // Produce a minimal standalone build for Docker — drops node_modules from image.
  output: 'standalone',

  async rewrites() {
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
