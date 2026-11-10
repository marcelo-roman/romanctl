import { useState } from "react";
import type { Generator, JobRequest, Project } from "../../api/types";
import type { MessageKey } from "../../i18n/messages";
import { validateEntity } from "../../lib/projectName";

export interface GenerateErrors {
  entity: MessageKey | null;
  in: MessageKey | null;
  plural: MessageKey | null;
}

export function useGenerateForm(project: Project) {
  const [kind, setKind] = useState(project.generators[0]?.kind ?? "");
  const [entity, setEntity] = useState("");
  const [inside, setInside] = useState("");
  const [plural, setPlural] = useState("");
  const [touched, setTouched] = useState(false);

  const generator: Generator | undefined = project.generators.find((item) => item.kind === kind);
  const needsIn = Boolean(generator?.needsIn);
  const errors: GenerateErrors = {
    entity: validateEntity(entity, true),
    in: needsIn ? validateEntity(inside, true) : null,
    plural: validateEntity(plural, false),
  };
  const valid = Boolean(generator) && !errors.entity && !errors.in && !errors.plural;

  const body = (): JobRequest => ({
    action: "project.generate",
    name: project.name,
    kind,
    entity,
    ...(needsIn && inside ? { in: inside } : {}),
    ...(plural ? { plural } : {}),
  });

  return {
    fields: { kind, entity, inside, plural },
    setters: { setKind, setEntity, setInside, setPlural },
    generator,
    needsIn,
    errors,
    valid,
    touched,
    touch: () => setTouched(true),
    body,
  };
}
