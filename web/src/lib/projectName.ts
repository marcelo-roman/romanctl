import type { MessageKey } from "../i18n/messages";

const projectNamePattern = /^[a-z0-9][a-z0-9._-]*$/;
const startsWithLetter = /^[a-z]/;
const entityPattern = /^[A-Za-z][A-Za-z0-9_-]*$/;

export function validateProjectName(name: string, isTemplate: boolean): MessageKey | null {
  if (!name) return "validation.nameRequired";
  if (name.length > 63) return "validation.nameTooLong";
  if (!projectNamePattern.test(name)) return "validation.namePattern";
  if (isTemplate && !startsWithLetter.test(name)) return "validation.nameLetter";
  return null;
}

export function validateEntity(value: string, required: boolean): MessageKey | null {
  if (!value) return required ? "validation.required" : null;
  if (!entityPattern.test(value)) return "validation.entityPattern";
  return null;
}
