import { useState } from "react";
import { useServices } from "../../api/queries";
import type { Service } from "../../api/types";
import { PageHeader } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { ServiceCard } from "./ServiceCard";
import { ServiceLogsDialog } from "./ServiceLogsDialog";

export function ServicesPage() {
  const { t } = useI18n();
  const services = useServices();
  const [logsFor, setLogsFor] = useState<Service | null>(null);
  return (
    <>
      <PageHeader title={t("services.title")} />
      <QueryState query={services}>
        {(list) => (
          <div className="card-grid">
            {list.map((service) => (
              <ServiceCard key={service.key} service={service} onLogs={() => setLogsFor(service)} />
            ))}
          </div>
        )}
      </QueryState>
      {logsFor && <ServiceLogsDialog service={logsFor} onClose={() => setLogsFor(null)} />}
    </>
  );
}
