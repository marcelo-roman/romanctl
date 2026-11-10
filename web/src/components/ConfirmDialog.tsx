import { useId, useState, type ReactNode } from "react";
import { useI18n } from "../i18n/context";
import { Dialog } from "./Dialog";

interface ConfirmDialogProps {
  title: string;
  children: ReactNode;
  confirmLabel: string;
  onConfirm: () => void;
  onClose: () => void;
  danger?: boolean;
  typeToConfirm?: string;
  busy?: boolean;
}

export function ConfirmDialog({ title, children, confirmLabel, onConfirm, onClose, danger, typeToConfirm, busy }: ConfirmDialogProps) {
  const { t } = useI18n();
  const [typed, setTyped] = useState("");
  const inputId = useId();
  const blocked = Boolean(typeToConfirm && typed !== typeToConfirm) || Boolean(busy);

  const footer = (
    <>
      <button type="button" className="button" onClick={onClose}>
        {t("common.cancel")}
      </button>
      <button type="button" className={danger ? "button button-danger" : "button button-primary"} disabled={blocked} onClick={onConfirm}>
        {confirmLabel}
      </button>
    </>
  );

  return (
    <Dialog title={title} onClose={onClose} footer={footer}>
      <div className="stack">
        {children}
        {typeToConfirm && (
          <div className="field">
            <label htmlFor={inputId}>{t("projects.removeTypeName", { name: typeToConfirm })}</label>
            <input
              id={inputId}
              className="input"
              value={typed}
              autoComplete="off"
              spellCheck={false}
              data-autofocus
              onChange={(event) => setTyped(event.target.value)}
            />
          </div>
        )}
      </div>
    </Dialog>
  );
}
