import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

const bff = process.env.BFF_URL || 'http://localhost:3301';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3300,
    strictPort: true,
    proxy: {
      '/screens': bff,
      '/actions': bff,
    },
  },
  test: {
    environment: 'jsdom',
    include: ['test/**/*.test.{ts,tsx}'],
    setupFiles: ['test/setup.ts'],
  },
});
