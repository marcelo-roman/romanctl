import { useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import { useInfo } from "../api/queries";
import type { Lang } from "../api/types";
import { I18nContext, translate, type I18n } from "./context";

const storageKey = "romanctl.lang";

function readStoredLang(): Lang | null {
  try {
    const value = window.localStorage.getItem(storageKey);
    return value === "pt" || value === "en" ? value : null;
  } catch {
    return null;
  }
}

function storeLang(lang: Lang): void {
  try {
    window.localStorage.setItem(storageKey, lang);
  } catch {
    return;
  }
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const info = useInfo();
  const [override, setOverride] = useState<Lang | null>(readStoredLang);
  const lang: Lang = override ?? info.data?.lang ?? "pt";

  useEffect(() => {
    document.documentElement.lang = lang === "en" ? "en" : "pt-BR";
  }, [lang]);

  const setLang = useCallback((next: Lang) => {
    storeLang(next);
    setOverride(next);
  }, []);

  const value = useMemo<I18n>(
    () => ({ lang, setLang, t: (key, vars) => translate(lang, key, vars) }),
    [lang, setLang],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}
