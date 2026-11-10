import { RefreshCw } from "lucide-react";
import { useServiceLogs } from "../../api/queries";
import type { Service } from "../../api/types";
import { Dialog } from "../../components/Dialog";
import { LogViewer } from "../../components/LogViewer";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";

export function ServiceLogsDialog({ service, onClose }: { service: Service; onClose: () => void }) {
  const { t } = useI18n();
  const logs = useServiceLogs(service.key, service.running);
  const title = t("services.logsTitle", { name: service.name });
  const footer = (
    <button type="button" className="button" disabled={logs.isFetching} onClick={() => void logs.refetch()}>
      <RefreshCw size={13} className={logs.isFetching ? "spin" : undefined} aria-hidden="true" />
      {t("common.refresh")}
    </button>
  );
  return (
    <Dialog title={title} onClose={onClose} size="lg" footer={footer}>
      <QueryState query={logs}>{(data) => <LogViewer lines={data.lines} emptyText={t("services.logsEmpty")} label={title} />}</QueryState>
    </Dialog>
  );
}
