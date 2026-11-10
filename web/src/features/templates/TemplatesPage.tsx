import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { useCatalog, useInfo } from "../../api/queries";
import type { Catalog, Generator, Info, Template } from "../../api/types";
import { Badge } from "../../components/Badge";
import { PageHeader, Section } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { useRunningJob } from "../../jobs/useRunningJob";
import { useStartJob } from "../../jobs/useStartJob";
import { groupTemplatesByStack, templateGenerators } from "../../lib/catalog";

function SourceStatus({ info }: { info: Info }) {
  const { t } = useI18n();
  const start = useStartJob();
  const running = useRunningJob((job) => job.action === "templates.update");
  const busy = start.isPending || Boolean(running);
  return (
    <div className="stack">
      {!info.templates.installed && <p className="warn-text">{t("templates.notInstalled")}</p>}
      <dl className="meta-grid">
        <dt>{t("templates.installed")}</dt>
        <dd>{info.templates.installed ? t("common.yes") : t("common.no")}</dd>
        <dt>{t("templates.ref")}</dt>
        <dd className="mono">{info.templates.ref || t("templates.refUnknown")}</dd>
        <dt>{t("templates.path")}</dt>
        <dd className="mono break">{info.templates.path}</dd>
      </dl>
      <div>
        <button type="button" className="button button-primary" disabled={busy} onClick={() => start.mutate({ action: "templates.update" })}>
          <RefreshCw size={14} className={busy ? "spin" : undefined} aria-hidden="true" />
          {t("templates.update")}
        </button>
      </div>
    </div>
  );
}

function GeneratorList({ generators }: { generators: Generator[] }) {
  const { t } = useI18n();
  if (generators.length === 0) return <p className="muted small">{t("templates.noGenerators")}</p>;
  return (
    <ul className="generator-list">
      {generators.map((generator) => (
        <li key={generator.kind}>
          <div className="row">
            <code className="mono strong">{generator.kind}</code>
            {generator.needsIn && <Badge tone="info">{t("templates.needsIn")}</Badge>}
            {generator.aliases && generator.aliases.length > 0 && (
              <span className="muted small">{t("templates.aliases", { list: generator.aliases.join(", ") })}</span>
            )}
          </div>
          <span className="small">{generator.description}</span>
        </li>
      ))}
    </ul>
  );
}

function TemplateEntry({ catalog, template }: { catalog: Catalog; template: Template }) {
  const { t } = useI18n();
  const services = template.services ?? [];
  return (
    <details className="template-entry">
      <summary>
        <span className="mono strong">{template.id}</span>
        <span className="tags">
          {template.arch.map((tag) => (
            <Badge key={tag} tone="info">
              {tag}
            </Badge>
          ))}
        </span>
      </summary>
      <div className="stack template-body">
        <p>{template.description}</p>
        {services.length > 0 && <p className="small">{t("projects.services", { list: services.join(", ") })}</p>}
        {template.url && <p className="small mono">{t("templates.port", { url: template.url })}</p>}
        <h4 className="subheading">{t("templates.generators")}</h4>
        <GeneratorList generators={templateGenerators(catalog, template)} />
      </div>
    </details>
  );
}

function CatalogBrowser({ catalog }: { catalog: Catalog }) {
  const { t } = useI18n();
  const groups = groupTemplatesByStack(catalog);
  const [stack, setStack] = useState(groups[0]?.stack ?? "");
  const current = groups.find((group) => group.stack === stack);
  return (
    <div className="browser">
      <div className="browser-nav" role="group" aria-label={t("templates.stacks")}>
        {groups.map((group) => (
          <button
            key={group.stack}
            type="button"
            className={group.stack === stack ? "chip chip-active" : "chip"}
            aria-pressed={group.stack === stack}
            onClick={() => setStack(group.stack)}
          >
            {group.stack}
            <span className="chip-count">{group.templates.length}</span>
          </button>
        ))}
      </div>
      <div className="stack">
        {current?.templates.map((template) => (
          <TemplateEntry key={template.id} catalog={catalog} template={template} />
        ))}
      </div>
      <div className="stack">
        <h3 className="subheading">{t("templates.simpleTypes")}</h3>
        <ul className="plain-list">
          {catalog.simpleTypes.map((simple) => (
            <li key={simple.type} className="row">
              <code className="mono strong">{simple.type}</code>
              <span className="muted small">{simple.description}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

export function TemplatesPage() {
  const { t } = useI18n();
  const info = useInfo();
  const catalog = useCatalog();
  return (
    <>
      <PageHeader title={t("templates.title")} />
      <div className="stack">
        <Section title={t("templates.source")}>
          <QueryState query={info}>{(data) => <SourceStatus info={data} />}</QueryState>
        </Section>
        <Section title={t("templates.catalog")}>
          <QueryState query={catalog}>{(data) => <CatalogBrowser catalog={data} />}</QueryState>
        </Section>
      </div>
    </>
  );
}
