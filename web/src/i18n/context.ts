import { createContext, useContext } from "react";
import type { Lang } from "../api/types";
import { dictionaries, type MessageKey } from "./messages";

export type Vars = Record<string, string | number>;

export interface I18n {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: MessageKey, vars?: Vars) => string;
}

export function translate(lang: Lang, key: MessageKey, vars?: Vars): string {
  const template = dictionaries[lang][key];
  if (!vars) return template;
  return template.replace(/\{(\w+)\}/g, (match, name: string) => (name in vars ? String(vars[name]) : match));
}

export const I18nContext = createContext<I18n>({
  lang: "pt",
  setLang: () => undefined,
  t: (key, vars) => translate("pt", key, vars),
});

export function useI18n(): I18n {
  return useContext(I18nContext);
}
