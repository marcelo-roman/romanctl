import { existsSync } from 'node:fs';
import { resolve } from 'node:path';
import type { NextConfig } from 'next';

const workspaceRoot = resolve(process.cwd(), '..', '..');
const workspaceEnv = resolve(workspaceRoot, '.env');

if (existsSync(workspaceEnv)) {
  process.loadEnvFile(workspaceEnv);
}

const config: NextConfig = {
  output: 'standalone',
  outputFileTracingRoot: workspaceRoot,
  turbopack: { root: workspaceRoot },
  transpilePackages: ['@sample-app/api', '@sample-app/db'],
};

export default config;
