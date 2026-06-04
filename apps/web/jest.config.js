module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
  moduleNameMapper: {
    // Force all axios imports to use the CommonJS build so Jest can parse them
    '^axios$': 'axios/dist/node/axios.cjs',
    '^axios/(.*)$': 'axios/dist/node/axios.cjs',
  },
  // The closing paren was missing in the original — fixed here
  transformIgnorePatterns: [
    'node_modules/(?!(axios)/)',
  ],
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.{js,jsx,ts,tsx}',
    '<rootDir>/src/**/*.{test,spec}.{js,jsx,ts,tsx}',
  ],
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70,
    },
  },
};
