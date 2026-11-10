import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');

interface Dependency {
  specifier: string;
  typeOnly: boolean;
}

interface SourceFile {
  path: string;
  name: string;
  dependencies: Dependency[];
}

type Rule = (file: SourceFile, dependency: Dependency) => string | undefined;

const staticImport = /\b(?:import|export)\s+(type\s+)?[^'";]*?\bfrom\s+['"]([^'"]+)['"]/g;
const bareImport = /\bimport\s*\(?\s*['"]([^'"]+)['"]/g;

function walk(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      return walk(path);
    }
    return /\.tsx?$/.test(path) ? [path] : [];
  });
}

function sources(dir: string): SourceFile[] {
  return walk(join(root, dir)).map((path) => {
    const text = readFileSync(path, 'utf8');
    return {
      path,
      name: relative(root, path).split(sep).join('/'),
      dependencies: [
        ...[...text.matchAll(staticImport)].map(([, type, specifier]) => ({ specifier: specifier ?? '', typeOnly: Boolean(type) })),
        ...[...text.matchAll(bareImport)].map(([, specifier]) => ({ specifier: specifier ?? '', typeOnly: false })),
      ],
    };
  });
}

function target(file: SourceFile, specifier: string): string {
  return relative(root, resolve(dirname(file.path), specifier)).split(sep).join('/');
}

function within(file: SourceFile, dependency: Dependency, prefix: string): boolean {
  return target(file, dependency.specifier).startsWith(prefix);
}

function violations(dir: string, rule: Rule): string[] {
  return sources(dir).flatMap((file) =>
    file.dependencies.flatMap((dependency) => {
      const problem = rule(file, dependency);
      return problem ? [`${file.name}: ${problem} (${dependency.specifier})`] : [];
    }),
  );
}

const builtIn = (specifier: string): boolean => specifier.startsWith('node:');
const isRelative = (specifier: string): boolean => specifier.startsWith('.');

const adapterFolder = /^packages\/db\/src\/(memory|postgres)\//;

const dbRule: Rule =(file, dependency) => {
  const { specifier } = dependency;
  if (isRelative(specifier)) {
    if (!within(file, dependency, 'packages/db/src/')) {
      return 'db must stay inside its own package';
    }
    const adapter = target(file, specifier).match(adapterFolder)?.[1];
    const own = file.name.match(adapterFolder)?.[1];
    return adapter && own && adapter !== own ? `the ${own} adapter must not depend on the ${adapter} adapter` : undefined;
  }
  return builtIn(specifier) || specifier === 'pg' ? undefined : 'db depends only on pg';
};

const apiPackages = ['@trpc/server', 'zod'];

const apiRule: Rule = (file, dependency) => {
  const { specifier, typeOnly } = dependency;
  const inDomain = file.name.startsWith('packages/api/src/domain/');
  if (isRelative(specifier)) {
    if (inDomain && !within(file, dependency, 'packages/api/src/domain/')) {
      return 'the domain depends on nothing outside it';
    }
    return within(file, dependency, 'packages/api/src/') ? undefined : 'api must stay inside its own package';
  }
  if (inDomain) {
    return 'the domain depends on no package';
  }
  if (specifier === '@sample-app/db') {
    return typeOnly ? undefined : 'api reaches db only through `import type` of its interfaces';
  }
  return builtIn(specifier) || apiPackages.includes(specifier) ? undefined : 'api must not depend on the web stack';
};

const webServer = ['apps/web/src/server/', 'apps/web/src/app/api/'];

const webRule: Rule = (file, dependency) => {
  const { specifier, typeOnly } = dependency;
  const onServer = webServer.some((prefix) => file.name.startsWith(prefix));
  if (isRelative(specifier)) {
    return within(file, dependency, 'apps/web/src/') ? undefined : 'web must stay inside its own app';
  }
  if (specifier === 'pg' || specifier === 'zod') {
    return 'web leaves persistence and validation to the packages';
  }
  if (specifier.startsWith('@sample-app/api/')) {
    return 'web imports api only through its entry point';
  }
  if (specifier === '@sample-app/api') {
    return onServer || typeOnly ? undefined : 'the browser side uses only the router type, through the tRPC client';
  }
  if (specifier.startsWith('@sample-app/db')) {
    return onServer ? undefined : 'only src/server picks the repositories';
  }
  return undefined;
};

describe('architecture', () => {
  it('finds the source of every package', () => {
    for (const dir of ['packages/db/src', 'packages/api/src/domain', 'packages/api/src/trpc', 'apps/web/src/server', 'apps/web/src/components']) {
      expect(sources(dir).length, dir).toBeGreaterThan(0);
    }
  });

  it('keeps db independent of api and web', () => {
    expect(violations('packages/db/src', dbRule)).toEqual([]);
  });

  it('keeps api free of Next and React, with a pure domain', () => {
    expect(violations('packages/api/src', apiRule)).toEqual([]);
  });

  it('lets web reach api only through the router type and the client', () => {
    expect(violations('apps/web/src', webRule)).toEqual([]);
  });
});
