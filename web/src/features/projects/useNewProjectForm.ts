import { useState } from "react";
import type { Template } from "../../api/types";
import { buildCreateBody, type ProjectSource } from "../../lib/jobs";
import { validateProjectName } from "../../lib/projectName";

export function sourceKey(source: ProjectSource | null): string {
  if (!source) return "";
  return source.kind === "simple" ? `simple:${source.type}` : `template:${source.template.stack}/${source.template.id}`;
}

export function useNewProjectForm() {
  const [source, setSource] = useState<ProjectSource | null>(null);
  const [name, setName] = useState("");
  const [services, setServices] = useState<string[]>([]);
  const [git, setGit] = useState(true);
  const [touched, setTouched] = useState(false);

  const isTemplate = source?.kind === "template";
  const nameError = validateProjectName(name, isTemplate);
  const valid = Boolean(source) && !nameError;

  const chooseSimple = (type: string) => {
    setSource({ kind: "simple", type });
    setServices([]);
  };

  const chooseTemplate = (template: Template) => {
    setSource({ kind: "template", template });
    setServices([...(template.services ?? [])]);
  };

  const toggleService = (service: string) => {
    setServices((current) => (current.includes(service) ? current.filter((item) => item !== service) : [...current, service]));
  };

  const body = () => {
    if (!source) return null;
    return buildCreateBody({ source, name, services, git });
  };

  return {
    source,
    selectedKey: sourceKey(source),
    name,
    setName,
    services,
    toggleService,
    git,
    setGit,
    nameError,
    valid,
    touched,
    touch: () => setTouched(true),
    chooseSimple,
    chooseTemplate,
    body,
  };
}
