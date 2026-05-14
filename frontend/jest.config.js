/** @type {import('jest').Config} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',

  // Run after Jest is installed but before each test file
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],

  testMatch: ['**/__tests__/**/*.test.(ts|tsx)'],

  moduleNameMapper: {
    // Path aliases (mirrors tsconfig paths)
    '^@/(.*)$': '<rootDir>/$1',
    // CSS modules
    '\\.(css|scss|sass)$': 'identity-obj-proxy',
  },

  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
        esModuleInterop: true,
      },
    }],
  },

  // Collect coverage from these paths
  collectCoverageFrom: [
    'components/**/*.{ts,tsx}',
    'contexts/**/*.{ts,tsx}',
    'app/**/*.{ts,tsx}',
    '!app/layout.tsx',
    '!**/*.d.ts',
  ],

  coverageReporters: ['text', 'lcov', 'html'],
  coverageDirectory: '../reports/frontend-coverage',
}
