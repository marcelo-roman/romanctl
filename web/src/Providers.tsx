import type { ReactNode } from "react";
import { QueryClientProvider, type QueryClient } from "@tanstack/react-query";
import { I18nProvider } from "./i18n/I18nProvider";
import { ToastProvider } from "./feedback/ToastProvider";
import { JobTrackerProvider } from "./jobs/JobTrackerProvider";

export function Providers({ client, children }: { client: QueryClient; children: ReactNode }) {
  return (
    <QueryClientProvider client={client}>
      <I18nProvider>
        <ToastProvider>
          <JobTrackerProvider>{children}</JobTrackerProvider>
        </ToastProvider>
      </I18nProvider>
    </QueryClientProvider>
  );
}
