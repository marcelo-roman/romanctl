import { useCallback, useMemo, useRef, useState, type ReactNode } from "react";
import { Link } from "react-router";
import { CheckCircle2, Info, X, XCircle } from "lucide-react";
import { useI18n } from "../i18n/context";
import { ToastContext, type Toast, type ToastApi, type ToastKind } from "./toastContext";

const toastLifetimeMs = 8000;

const icons: Record<ToastKind, typeof Info> = { info: Info, success: CheckCircle2, error: XCircle };

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: (id: number) => void }) {
  const { t } = useI18n();
  const Icon = icons[toast.kind];
  return (
    <div className={`toast toast-${toast.kind}`} role={toast.kind === "error" ? "alert" : "status"}>
      <Icon size={16} aria-hidden="true" />
      <span className="toast-message">{toast.message}</span>
      {toast.link && (
        <Link className="toast-link" to={toast.link.to} onClick={() => onDismiss(toast.id)}>
          {toast.link.label}
        </Link>
      )}
      <button type="button" className="icon-button" aria-label={t("common.close")} onClick={() => onDismiss(toast.id)}>
        <X size={14} aria-hidden="true" />
      </button>
    </div>
  );
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(1);

  const dismiss = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const push = useCallback(
    (toast: Omit<Toast, "id">) => {
      const id = nextId.current++;
      setToasts((current) => [...current.slice(-4), { ...toast, id }]);
      window.setTimeout(() => dismiss(id), toastLifetimeMs);
    },
    [dismiss],
  );

  const value = useMemo<ToastApi>(() => ({ push, dismiss }), [push, dismiss]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="toasts" aria-live="polite">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onDismiss={dismiss} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}
