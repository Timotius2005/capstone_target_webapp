/** @type {import('next').NextConfig} */
const nextConfig = {
  // appDir is stable in Next.js 14 — no longer needs experimental flag
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
      {
        source: '/config/:path*',
        destination: 'http://localhost:8080/config/:path*',
      },
    ]
  },
}

module.exports = nextConfig
