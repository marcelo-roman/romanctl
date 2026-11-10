import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { useI18n } from "../i18n/context";

async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

export function CopyButton({ value }: { value: string }) {
  const { t } = useI18n();
  const [copied, setCopied] = useState(false);

  const onClick = async () => {
    if (!(await copyText(value))) return;
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1500);
  };

  return (
    <button type="button" className="icon-button" aria-label={t("common.copyValue", { value })} title={copied ? t("common.copied") : t("common.copy")} onClick={onClick}>
      {copied ? <Check size={14} aria-hidden="true" /> : <Copy size={14} aria-hidden="true" />}
    </button>
  );
}
