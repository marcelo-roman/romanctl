import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const src = resolve(dirname(fileURLToPath(import.meta.url)), '../src');

type Layer = 'domain' | 'application' | 'infrastructure' | 'api';

const rules: Record<Layer, { layers: Layer[]; packages: string[] | 'any' }> = {
  domain: { layers: ['domain'], packages: [] },
  application: { layers: ['domain', 'application'], packages: ['@nestjs/common', '@nestjs/cqrs'] },
  infrastructure: { layers: ['domain', 'application', 'infrastructure'], packages: 'any' },
  api: { layers: ['domain', 'application', 'infrastructure', 'api'], packages: 'any' },
};

const importPattern = /(?:\bfrom\s+|\bimport\s*\(?\s*)['"]([^'"]+)['"]/g;

function sourceFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      return sourceFiles(path);
    }
    return path.endsWith('.ts') ? [path] : [];
  });
}

function layerOf(path: string): Layer | undefined {
  const top = relative(src, path).split(/[\\/]/)[0];
  return top && top in rules ? (top as Layer) : undefined;
}

function violations(file: string): string[] {
  const layer = layerOf(file);
  if (!layer) {
    return [];
  }
  const rule = rules[layer];
  const found: string[] = [];
  for (const [, specifier] of readFileSync(file, 'utf8').matchAll(importPattern)) {
    if (!specifier) {
      continue;
    }
    const name = relative(src, file);
    if (specifier.startsWith('.')) {
      const target = layerOf(resolve(dirname(file), specifier));
      if (!target || !rule.layers.includes(target)) {
        found.push(`${name}: ${layer} must not depend on ${specifier}`);
      }
      continue;
    }
    if (rule.packages !== 'any' && !rule.packages.includes(specifier)) {
      found.push(`${name}: ${layer} must not depend on ${specifier}`);
    }
  }
  return found;
}

describe('architecture', () => {
  it('finds the source of every layer', () => {
    const layers = new Set(sourceFiles(src).map(layerOf));
    expect([...Object.keys(rules)].every((layer) => layers.has(layer as Layer))).toBe(true);
  });

  it('keeps dependencies pointing inward', () => {
    expect(sourceFiles(src).flatMap(violations)).toEqual([]);
  });
});
