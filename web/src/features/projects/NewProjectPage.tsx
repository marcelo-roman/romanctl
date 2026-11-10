import type { FormEvent } from "react";
import { Link } from "react-router";
import { ArrowLeft } from "lucide-react";
import { useCatalog } from "../../api/queries";
import { TextField } from "../../components/Field";
import { PageHeader, Section } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { useStartJob } from "../../jobs/useStartJob";
import { SourcePicker } from "./SourcePicker";
import { useNewProjectForm } from "./useNewProjectForm";

type Form = ReturnType<typeof useNewProjectForm>;

function sourceLabel(form: Form): string {
  if (!form.source) return "";
  if (form.source.kind === "simple") return form.source.type;
  return `${form.source.template.stack}/${form.source.template.id}`;
}

function ServiceChoices({ form }: { form: Form }) {
  const { t } = useI18n();
  if (form.source?.kind !== "template") return null;
  const available = form.source.template.services ?? [];
  return (
    <fieldset className="field">
      <legend>{t("wizard.services")}</legend>
      {available.length === 0 && <span className="muted small">{t("wizard.noServices")}</span>}
      <div className="row wrap">
        {available.map((service) => (
          <label key={service} className="checkbox">
            <input type="checkbox" checked={form.services.includes(service)} onChange={() => form.toggleService(service)} />
            {service}
          </label>
        ))}
      </div>
    </fieldset>
  );
}

function DetailsForm({ form }: { form: Form }) {
  const { t } = useI18n();
  const start = useStartJob({ navigateToJob: true });

  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    form.touch();
    const body = form.body();
    if (!form.valid || !body) return;
    start.mutate(body);
  };

  if (!form.source) return <p className="muted">{t("wizard.chooseFirst")}</p>;
  return (
    <form className="stack" onSubmit={onSubmit} noValidate>
      <p className="small">{t("wizard.selected", { name: sourceLabel(form) })}</p>
      <TextField
        label={t("wizard.name")}
        hint={t("wizard.nameHint")}
        value={form.name}
        onValueChange={form.setName}
        error={form.touched && form.nameError ? t(form.nameError) : null}
        autoComplete="off"
        spellCheck={false}
        maxLength={63}
      />
      <ServiceChoices form={form} />
      <label className="checkbox">
        <input type="checkbox" checked={form.git} onChange={(event) => form.setGit(event.target.checked)} />
        {t("wizard.git")}
      </label>
      <div>
        <button type="submit" className="button button-primary" disabled={start.isPending}>
          {t("wizard.create")}
        </button>
      </div>
    </form>
  );
}

export function NewProjectPage() {
  const { t } = useI18n();
  const catalog = useCatalog();
  const form = useNewProjectForm();
  return (
    <>
      <Link className="link small back-link" to="/projects">
        <ArrowLeft size={14} aria-hidden="true" />
        {t("projects.title")}
      </Link>
      <PageHeader title={t("wizard.title")} />
      <div className="wizard">
        <Section title={t("wizard.stepSource")}>
          <QueryState query={catalog}>
            {(data) => <SourcePicker catalog={data} selectedKey={form.selectedKey} onSimple={form.chooseSimple} onTemplate={form.chooseTemplate} />}
          </QueryState>
        </Section>
        <Section title={t("wizard.stepDetails")}>
          <DetailsForm form={form} />
        </Section>
      </div>
    </>
  );
}
