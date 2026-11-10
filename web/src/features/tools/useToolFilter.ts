import { useMemo, useState } from "react";
import type { Tool, ToolGroup } from "../../api/types";
import { allGroups, filterTools } from "../../lib/tools";

export interface ToolSection {
  group: ToolGroup;
  tools: Tool[];
}

export function useToolFilter(groups: ToolGroup[], tools: Tool[]) {
  const [group, setGroup] = useState(allGroups);
  const [search, setSearch] = useState("");
  const [installedOnly, setInstalledOnly] = useState(false);

  const visible = useMemo(() => filterTools(tools, { group, search, installedOnly }), [tools, group, search, installedOnly]);

  const sections = useMemo<ToolSection[]>(
    () =>
      groups
        .map((item) => ({ group: item, tools: visible.filter((tool) => tool.group === item.key) }))
        .filter((section) => section.tools.length > 0),
    [groups, visible],
  );

  return { group, setGroup, search, setSearch, installedOnly, setInstalledOnly, visible, sections };
}
