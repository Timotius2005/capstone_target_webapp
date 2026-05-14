/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        navy: {
          50:  '#EEF2F7',
          100: '#D5E0EE',
          200: '#ADC1DD',
          500: '#2D5FA4',
          600: '#1E3A8A',
          700: '#162D6E',
          800: '#0B1E3D',
          900: '#060E20',
        },
      },
      animation: {
        'fade-in':  'fadeIn 0.2s ease-out',
        'slide-up': 'slideUp 0.22s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%':   { opacity: '0', transform: 'translateY(4px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideUp: {
          '0%':   { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', '-apple-system', 'sans-serif'],
      },
      boxShadow: {
        // Enterprise shadows — subtle, no neon
        card:     '0 1px 3px rgba(0,0,0,0.07), 0 1px 2px rgba(0,0,0,0.04)',
        'card-md':'0 4px 12px rgba(0,0,0,0.09), 0 2px 4px rgba(0,0,0,0.04)',
        // Aliases kept for backward-compat — all nullified
        glass:      '0 1px 3px rgba(0,0,0,0.07)',
        'glass-lg': '0 4px 12px rgba(0,0,0,0.08)',
        glow:       '0 0 0 transparent',
        'glow-red': '0 0 0 transparent',
        'glow-green':'0 0 0 transparent',
      },
    },
  },
  plugins: [],
}
