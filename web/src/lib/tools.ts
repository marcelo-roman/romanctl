import type { Tool } from "../api/types";
import { isNewer } from "./version";

export const allGroups = "all";

export interface ToolFilter {
  group: string;
  search: string;
  installedOnly: boolean;
}

function matches(tool: Tool, search: string): boolean {
  if (!search) return true;
  const haystack = `${tool.key} ${tool.name} ${tool.description}`.toLowerCase();
  return search
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)
    .every((word) => haystack.includes(word));
}

export function filterTools(tools: Tool[], filter: ToolFilter): Tool[] {
  return tools.filter((tool) => {
    if (filter.group !== allGroups && tool.group !== filter.group) return false;
    if (filter.installedOnly && !tool.installed) return false;
    return matches(tool, filter.search.trim());
  });
}

export function hasUpdate(tool: Tool): boolean {
  return tool.installed && isNewer(tool.latest, tool.version);
}

export function countByGroup(tools: Tool[]): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const tool of tools) counts[tool.group] = (counts[tool.group] ?? 0) + 1;
  return counts;
}
