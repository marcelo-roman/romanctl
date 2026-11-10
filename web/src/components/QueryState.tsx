import type { UseQueryResult } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useI18n } from "../i18n/context";

export function Loading({ text }: { text?: string }) {
  const { t } = useI18n();
  return (
    <p className="muted loading" role="status">
      {text ?? t("common.loading")}
    </p>
  );
}

export function ErrorMessage({ error }: { error: Error }) {
  const { t } = useI18n();
  return (
    <p className="error-box" role="alert">
      {t("errors.load", { message: error.message })}
    </p>
  );
}

interface QueryStateProps<T> {
  query: UseQueryResult<T, Error>;
  loadingText?: string;
  children: (data: T) => ReactNode;
}

export function QueryState<T>({ query, loadingText, children }: QueryStateProps<T>) {
  if (query.data !== undefined) return <>{children(query.data)}</>;
  if (query.error) return <ErrorMessage error={query.error} />;
  return <Loading text={loadingText} />;
}
