import { useState } from "react";
import { Download, Loader2, Play, RotateCw, ScrollText, Square } from "lucide-react";
import type { Service } from "../../api/types";
import { Badge, StatusDot } from "../../components/Badge";
import { ConfirmDialog } from "../../components/ConfirmDialog";
import { CopyButton } from "../../components/CopyButton";
import { useI18n } from "../../i18n/context";
import { serviceJob, useRunningJob } from "../../jobs/useRunningJob";
import { useStartJob } from "../../jobs/useStartJob";

type ServiceAction = "service.install" | "service.start" | "service.stop" | "service.restart";
type EngineAction = Extract<ServiceAction, "service.stop" | "service.restart">;

function ServiceState({ service, busy }: { service: Service; busy: boolean }) {
  const { t } = useI18n();
  if (busy) return <Loader2 size={14} className="spin" aria-label={t("job.status.running")} />;
  if (service.unsupported) return <Badge tone="warn" title={service.unsupported}>{t("status.unsupported")}</Badge>;
  if (!service.installed) return <Badge>{t("status.notInstalled")}</Badge>;
  return <StatusDot on={service.running} label={service.running ? t("status.running") : t("status.stopped")} />;
}

function EngineConfirm({ service, action, onConfirm, onClose }: { service: Service; action: EngineAction; onConfirm: () => void; onClose: () => void }) {
  const { t } = useI18n();
  const stopping = action === "service.stop";
  return (
    <ConfirmDialog
      title={t(stopping ? "services.stopEngineTitle" : "services.restartEngineTitle", { name: service.name })}
      confirmLabel={t(stopping ? "services.stop" : "services.restart")}
      danger
      onConfirm={onConfirm}
      onClose={onClose}
    >
      <p>{t(stopping ? "services.stopEngineBody" : "services.restartEngineBody")}</p>
    </ConfirmDialog>
  );
}

export function ServiceCard({ service, onLogs }: { service: Service; onLogs: () => void }) {
  const { t } = useI18n();
  const start = useStartJob();
  const running = useRunningJob(serviceJob(service.key));
  const [pending, setPending] = useState<EngineAction | null>(null);
  const busy = start.isPending || Boolean(running);
  const key = service.key;

  const run = (action: ServiceAction) => start.mutate({ action, key });
  const guarded = (action: EngineAction) => (service.engine ? setPending(action) : run(action));

  return (
    <article className={service.unsupported ? "card card-disabled" : "card"}>
      <div className="row-between">
        <h2 className="card-title">
          {service.name}
          {service.engine && <Badge tone="info">{t("services.engine")}</Badge>}
        </h2>
        <ServiceState service={service} busy={busy} />
      </div>
      {service.unsupported && <p className="muted small">{service.unsupported}</p>}
      {service.wanted && !service.running && <p className="small warn-text">{t("services.wanted")}</p>}
      {service.url && (
        <div className="row url-row">
          <code className="mono small truncate" title={service.url}>
            {service.url}
          </code>
          <CopyButton value={service.url} />
        </div>
      )}
      <div className="row-actions">
        {!service.installed && !service.unsupported && (
          <button type="button" className="button button-sm" disabled={busy} onClick={() => run("service.install")}>
            <Download size={13} aria-hidden="true" />
            {t("services.install")}
          </button>
        )}
        {service.installed && !service.running && (
          <button type="button" className="button button-sm" disabled={busy} onClick={() => run("service.start")}>
            <Play size={13} aria-hidden="true" />
            {t("services.start")}
          </button>
        )}
        {service.installed && service.running && (
          <>
            <button type="button" className="button button-sm" disabled={busy} onClick={() => guarded("service.stop")}>
              <Square size={13} aria-hidden="true" />
              {t("services.stop")}
            </button>
            <button type="button" className="button button-sm" disabled={busy} onClick={() => guarded("service.restart")}>
              <RotateCw size={13} aria-hidden="true" />
              {t("services.restart")}
            </button>
          </>
        )}
        {service.installed && (
          <button type="button" className="button button-sm" onClick={onLogs}>
            <ScrollText size={13} aria-hidden="true" />
            {t("common.logs")}
          </button>
        )}
      </div>
      {pending && (
        <EngineConfirm
          service={service}
          action={pending}
          onClose={() => setPending(null)}
          onConfirm={() => {
            setPending(null);
            run(pending);
          }}
        />
      )}
    </article>
  );
}
