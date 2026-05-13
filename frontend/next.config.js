/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
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