import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const apps = ['api', 'bff', 'web'] as const;
type App = (typeof apps)[number];

const importPattern = /\b(?:import|export)\s+(?:type\s+)?(?:[^'";]*?\bfrom\s+)?['"]([^'"]+)['"]|\bimport\s*\(\s*['"]([^'"]+)['"]/g;
const skipped = new Set(['node_modules', 'dist', 'coverage']);

function appDir(app: App): string {
  return join(root, 'apps', app);
}

function sourceFiles(dir: string): string[] {
  if (!existsSync(dir)) {
    return [];
  }
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      return skipped.has(entry.name) ? [] : sourceFiles(path);
    }
    return /\.(ts|tsx)$/.test(entry.name) ? [path] : [];
  });
}

function specifiers(file: string): string[] {
  const text = readFileSync(file, 'utf8');
  return [...text.matchAll(importPattern)].map(([, fromImport, dynamic]) => fromImport ?? dynamic ?? '');
}

function workspaceNames(): Map<string, App> {
  const names = new Map<string, App>();
  for (const app of apps) {
    const pkg = JSON.parse(readFileSync(join(appDir(app), 'package.json'), 'utf8')) as { name: string };
    names.set(pkg.name, app);
  }
  return names;
}

function appOf(path: string): App | undefined {
  const [top, app] = relative(root, path).split(sep);
  return top === 'apps' && apps.includes(app as App) ? (app as App) : undefined;
}

function crossings(app: App): string[] {
  const names = workspaceNames();
  return sourceFiles(appDir(app)).flatMap((file) =>
    specifiers(file).flatMap((specifier) => {
      const name = relative(root, file);
      if (specifier.startsWith('.')) {
        const target = appOf(resolve(dirname(file), specifier));
        return target === app ? [] : [`${name}: ${app} must not import ${specifier}`];
      }
      const workspace = [...names.entries()].find(([pkg]) => specifier === pkg || specifier.startsWith(`${pkg}/`));
      return workspace && workspace[1] !== app ? [`${name}: ${app} must not import ${specifier}`] : [];
    }),
  );
}

function siblingDependencies(app: App): string[] {
  const names = workspaceNames();
  const pkg = JSON.parse(readFileSync(join(appDir(app), 'package.json'), 'utf8')) as Record<string, Record<string, string> | undefined>;
  const declared = [...Object.keys(pkg.dependencies ?? {}), ...Object.keys(pkg.devDependencies ?? {})];
  return declared.filter((dependency) => names.has(dependency));
}

describe('apps talk only over HTTP', () => {
  it('finds the source of every app', () => {
    expect(apps.filter((app) => sourceFiles(join(appDir(app), 'src')).length === 0)).toEqual([]);
  });

  for (const app of apps) {
    it(`keeps ${app} free of the other apps' source`, () => {
      expect(crossings(app)).toEqual([]);
    });

    it(`keeps ${app} free of dependencies on the other apps`, () => {
      expect(siblingDependencies(app)).toEqual([]);
    });
  }

  it('keeps the api unaware of the bff and the web', () => {
    const mentions = sourceFiles(join(appDir('api'), 'src')).filter((file) => /\b(bff|web)\b/i.test(readFileSync(file, 'utf8')));
    expect(mentions.map((file) => relative(root, file))).toEqual([]);
  });

  it('keeps the web calling only the screens and actions of the bff', () => {
    const files = sourceFiles(join(appDir('web'), 'src'));
    const fetchers = files.filter((file) => /\bfetch\(/.test(readFileSync(file, 'utf8')));
    expect(fetchers.map((file) => relative(root, file))).toEqual([join('apps', 'web', 'src', 'bff', 'bff-client.ts')]);

    const paths = files.flatMap((file) => [...readFileSync(file, 'utf8').matchAll(/bffRequest(?:<[^>]*>)?\(\s*[`'"]([^`'"$]*)/g)].map(([, path]) => path ?? ''));
    expect(paths.length).toBeGreaterThan(0);
    expect(paths.filter((path) => !/^\/(screens|actions)(\/|$)/.test(path))).toEqual([]);
  });
});
