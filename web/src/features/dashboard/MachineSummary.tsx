import { useSystem, systemRefreshMs } from "../../api/queries";
import type { Battery, SystemSnapshot } from "../../api/types";
import { QueryState } from "../../components/QueryState";
import { Section } from "../../components/PageHeader";
import { useI18n } from "../../i18n/context";
import type { MessageKey } from "../../i18n/messages";
import { formatBytes, formatPercent } from "../../lib/format";
import { cpuPercent, diskUsage, level, memoryUsage } from "../../lib/system";

interface MeterProps {
  label: string;
  percent: number;
  detail: string;
  invert?: boolean;
}

function Meter({ label, percent, detail, invert }: MeterProps) {
  const clamped = Math.max(0, Math.min(100, percent));
  const tone = level(invert ? 100 - clamped : clamped);
  return (
    <div className="meter">
      <div className="meter-head">
        <span className="meter-label">{label}</span>
        <span className="meter-value">{formatPercent(clamped)}</span>
      </div>
      <div className="meter-track" role="meter" aria-label={label} aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(clamped)}>
        <div className={`meter-fill meter-${tone}`} style={{ width: `${clamped}%` }} />
      </div>
      <span className="meter-detail muted">{detail}</span>
    </div>
  );
}

const batteryStatusKeys: Record<string, MessageKey> = {
  charging: "dashboard.batteryStatus.charging",
  discharging: "dashboard.batteryStatus.discharging",
  full: "dashboard.batteryStatus.full",
};

function BatteryMeter({ battery }: { battery: Battery }) {
  const { t } = useI18n();
  const status = t(batteryStatusKeys[battery.status] ?? "dashboard.batteryStatus.unknown");
  const hours = battery.hoursLeft ? ` · ${t("dashboard.hoursLeft", { hours: battery.hoursLeft.toFixed(1) })}` : "";
  return <Meter label={t("dashboard.battery")} percent={battery.percent} invert detail={`${status}${hours}`} />;
}

function Meters({ snapshot }: { snapshot: SystemSnapshot }) {
  const { t } = useI18n();
  const memory = memoryUsage(snapshot);
  const disk = diskUsage(snapshot);
  return (
    <div className="meters">
      <Meter label={t("dashboard.cpu")} percent={cpuPercent(snapshot)} detail={t("dashboard.cores", { count: snapshot.cores })} />
      <Meter label={t("dashboard.memory")} percent={memory.percent} detail={t("dashboard.usedOf", { used: formatBytes(memory.used), total: formatBytes(memory.total) })} />
      <Meter label={`${t("dashboard.disk")} ${snapshot.disk.path}`} percent={disk.percent} detail={t("dashboard.usedOf", { used: formatBytes(disk.used), total: formatBytes(disk.total) })} />
      {snapshot.battery && <BatteryMeter battery={snapshot.battery} />}
    </div>
  );
}

function TopProcesses({ snapshot }: { snapshot: SystemSnapshot }) {
  const { t } = useI18n();
  const top = snapshot.top.slice(0, 5);
  if (top.length === 0) return null;
  return (
    <details className="top-processes">
      <summary>{t("dashboard.topProcesses")}</summary>
      <table className="table table-compact">
        <tbody>
          {top.map((proc) => (
            <tr key={proc.pid}>
              <td className="mono">{proc.pid}</td>
              <td>{proc.name}</td>
              <td className="num">{formatPercent(proc.cpuPercentOfOneCore)}</td>
              <td className="num">{formatBytes(proc.rssBytes)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </details>
  );
}

export function MachineSummary() {
  const { t } = useI18n();
  const system = useSystem();
  const refresh = <span className="muted small">{t("dashboard.refreshEvery", { seconds: systemRefreshMs / 1000 })}</span>;
  return (
    <Section title={t("dashboard.machine")} actions={refresh}>
      <QueryState query={system}>
        {(snapshot) => (
          <>
            <Meters snapshot={snapshot} />
            <TopProcesses snapshot={snapshot} />
          </>
        )}
      </QueryState>
    </Section>
  );
}
