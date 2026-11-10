import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const src = resolve(dirname(fileURLToPath(import.meta.url)), '../src');

type Layer = 'models' | 'repositories' | 'services' | 'routes';

interface Rule {
  layers: Layer[];
  packages: (specifier: string) => boolean;
}

const builtIn = (specifier: string): boolean => specifier.startsWith('node:');

const rules: Record<Layer, Rule> = {
  models: { layers: ['models'], packages: () => false },
  repositories: { layers: ['models', 'repositories'], packages: () => true },
  services: { layers: ['models', 'repositories', 'services'], packages: builtIn },
  routes: { layers: ['models', 'services', 'routes'], packages: (specifier) => builtIn(specifier) || specifier === 'fastify' },
};

interface Dependency {
  specifier: string;
  typeOnly: boolean;
}

const staticImport = /\b(?:import|export)\s+(type\s+)?[^'";]*?\bfrom\s+['"]([^'"]+)['"]/g;
const bareImport = /\bimport\s*\(?\s*['"]([^'"]+)['"]/g;

function sourceFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      return sourceFiles(path);
    }
    return path.endsWith('.ts') ? [path] : [];
  });
}

function dependencies(file: string): Dependency[] {
  const text = readFileSync(file, 'utf8');
  return [
    ...[...text.matchAll(staticImport)].map(([, type, specifier]) => ({ specifier: specifier ?? '', typeOnly: Boolean(type) })),
    ...[...text.matchAll(bareImport)].map(([, specifier]) => ({ specifier: specifier ?? '', typeOnly: false })),
  ];
}

function pathParts(path: string): string[] {
  return relative(src, path).split(sep);
}

function layerOf(path: string): Layer | undefined {
  const [top, ...rest] = pathParts(path);
  return top && rest.length > 0 && top in rules ? (top as Layer) : undefined;
}

function isRepositoryInterface(path: string): boolean {
  const parts = pathParts(path);
  return parts.length === 2 && parts[0] === 'repositories';
}

function adapterOf(path: string): string | undefined {
  const parts = pathParts(path);
  return parts[0] === 'repositories' && parts.length > 2 ? parts[1] : undefined;
}

function violations(file: string): string[] {
  const layer = layerOf(file);
  if (!layer) {
    return [];
  }
  const rule = rules[layer];
  const name = relative(src, file);
  return dependencies(file).flatMap(({ specifier, typeOnly }) => {
    if (!specifier.startsWith('.')) {
      return rule.packages(specifier) ? [] : [`${name}: ${layer} must not depend on ${specifier}`];
    }
    const target = resolve(dirname(file), specifier);
    const targetLayer = layerOf(target);
    if (!targetLayer || !rule.layers.includes(targetLayer)) {
      return [`${name}: ${layer} must not depend on ${specifier}`];
    }
    if (layer === 'services' && targetLayer === 'repositories' && (!isRepositoryInterface(target) || !typeOnly)) {
      return [`${name}: services must depend on repositories only through \`import type\` of their interfaces (${specifier})`];
    }
    const adapter = adapterOf(target);
    if (adapter && adapterOf(file) !== adapter) {
      return [`${name}: must not depend on the ${adapter} adapter (${specifier})`];
    }
    return [];
  });
}

describe('layering', () => {
  const files = sourceFiles(src);

  it('finds the source of every layer', () => {
    const layers = new Set(files.map(layerOf));
    expect(Object.keys(rules).filter((layer) => !layers.has(layer as Layer))).toEqual([]);
  });

  it('keeps routes away from repositories and services behind repository interfaces', () => {
    expect(files.flatMap(violations)).toEqual([]);
  });
});
