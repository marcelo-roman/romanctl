import type { Catalog, Generator, Project, Template } from "../api/types";

export interface StackGroup {
  stack: string;
  templates: Template[];
}

export function groupTemplatesByStack(catalog: Catalog): StackGroup[] {
  const groups = new Map<string, Template[]>();
  for (const stack of catalog.stacks) groups.set(stack, []);
  for (const template of catalog.templates) {
    const list = groups.get(template.stack) ?? [];
    list.push(template);
    groups.set(template.stack, list);
  }
  return [...groups.entries()]
    .filter(([, templates]) => templates.length > 0)
    .map(([stack, templates]) => ({ stack, templates }));
}

export function templateGenerators(catalog: Catalog, template: Template): Generator[] {
  return catalog.generators[`${template.stack}/${template.id}`] ?? [];
}

export function isProjectUp(project: Project): boolean {
  return Boolean(project.up?.alive);
}

export function projectUrl(project: Project): string {
  return project.up?.url || project.url || project.manifest?.url || "";
}

export function projectKind(project: Project): string {
  if (project.manifest) return `${project.manifest.stack}/${project.manifest.template}`;
  return project.type;
}
