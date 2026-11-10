import { createContext, useContext } from "react";

export type ToastKind = "info" | "success" | "error";

export interface Toast {
  id: number;
  kind: ToastKind;
  message: string;
  link?: { to: string; label: string };
}

export interface ToastApi {
  push: (toast: Omit<Toast, "id">) => void;
  dismiss: (id: number) => void;
}

export const ToastContext = createContext<ToastApi>({ push: () => undefined, dismiss: () => undefined });

export function useToast(): ToastApi {
  return useContext(ToastContext);
}
