import js from '@eslint/js';
import boundaries from 'eslint-plugin-boundaries';
import reactHooks from 'eslint-plugin-react-hooks';
import { defineConfig, globalIgnores } from 'eslint/config';
import globals from 'globals';
import tseslint from 'typescript-eslint';

const layers = ['app', 'pages', 'widgets', 'features', 'entities', 'shared'];

const slicedLayers = ['pages', 'widgets', 'features', 'entities'];

const lowerLayers = (layer) => (layer === 'shared' ? ['shared'] : layers.slice(layers.indexOf(layer) + 1));

const publicApi = (type) => ({ element: { type, fileInternalPath: 'index.ts' } });

export default defineConfig(
  globalIgnores(['dist', 'coverage', 'node_modules']),
  js.configs.recommended,
  tseslint.configs.recommended,
  reactHooks.configs.flat['recommended-latest'],
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: { globals: globals.browser },
  },
  {
    files: ['*.config.{js,ts}'],
    languageOptions: { globals: globals.node },
  },
  {
    files: ['src/**/*.{ts,tsx}', 'tests/**/*.{ts,tsx}'],
    plugins: { boundaries },
    settings: {
      'import/resolver': { typescript: { project: './tsconfig.json' } },
      'boundaries/include': ['src/**/*', 'tests/**/*'],
      'boundaries/elements': [
        { type: 'app', pattern: 'src/app', partialMatch: false },
        ...slicedLayers.map((type) => ({ type, pattern: `src/${type}/*`, partialMatch: false, capture: ['slice'] })),
        { type: 'shared', pattern: 'src/shared/*', partialMatch: false, capture: ['segment'] },
        { type: 'tests', pattern: 'tests', partialMatch: false },
      ],
      'boundaries/files': [{ pattern: '**/*.test.{ts,tsx}', category: 'test' }],
    },
    rules: {
      'boundaries/no-unknown-files': 'error',
      'boundaries/no-unknown-dependencies': 'error',
      'boundaries/dependencies': [
        'error',
        {
          default: 'disallow',
          policies: [
            ...layers.map((layer) => ({
              from: { element: { type: layer } },
              allow: { to: lowerLayers(layer).map(publicApi) },
            })),
            { from: { file: { categories: 'test' } }, allow: { to: { element: { type: 'tests' } } } },
          ],
        },
      ],
    },
  },
);
