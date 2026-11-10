import type { FormEvent } from "react";
import type { Project } from "../../api/types";
import { Badge } from "../../components/Badge";
import { Dialog } from "../../components/Dialog";
import { TextField } from "../../components/Field";
import { useI18n } from "../../i18n/context";
import { useStartJob } from "../../jobs/useStartJob";
import { useGenerateForm } from "./useGenerateForm";

function GeneratorOptions({ project, value, onChange }: { project: Project; value: string; onChange: (kind: string) => void }) {
  const { t } = useI18n();
  return (
    <fieldset className="field">
      <legend>{t("projects.generator")}</legend>
      <div className="option-list">
        {project.generators.map((generator) => (
          <label key={generator.kind} className={generator.kind === value ? "option option-selected" : "option"}>
            <input type="radio" name="generator" value={generator.kind} checked={generator.kind === value} onChange={() => onChange(generator.kind)} />
            <span className="option-body">
              <span className="row">
                <code className="mono strong">{generator.kind}</code>
                {generator.needsIn && <Badge tone="info">--in</Badge>}
              </span>
              <span className="muted small">{generator.description}</span>
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}

export function GenerateDialog({ project, onClose }: { project: Project; onClose: () => void }) {
  const { t } = useI18n();
  const form = useGenerateForm(project);
  const start = useStartJob({ onStarted: onClose });
  const formId = `generate-${project.name}`;
  const show = (key: typeof form.errors.entity) => (form.touched && key ? t(key) : null);

  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    form.touch();
    if (!form.valid) return;
    start.mutate(form.body());
  };

  const footer = (
    <>
      <button type="button" className="button" onClick={onClose}>
        {t("common.cancel")}
      </button>
      <button type="submit" form={formId} className="button button-primary" disabled={start.isPending}>
        {t("projects.generate")}
      </button>
    </>
  );

  return (
    <Dialog title={t("projects.generateTitle", { name: project.name })} onClose={onClose} footer={footer} size="lg">
      <form id={formId} className="stack" onSubmit={onSubmit} noValidate>
        <GeneratorOptions project={project} value={form.fields.kind} onChange={form.setters.setKind} />
        <TextField
          label={t("projects.entity")}
          hint={t("projects.entityHint")}
          value={form.fields.entity}
          onValueChange={form.setters.setEntity}
          error={show(form.errors.entity)}
          autoComplete="off"
          spellCheck={false}
          data-autofocus
        />
        {form.needsIn && (
          <TextField
            label={t("projects.in")}
            hint={t("projects.inHint")}
            value={form.fields.inside}
            onValueChange={form.setters.setInside}
            error={show(form.errors.in)}
            autoComplete="off"
            spellCheck={false}
          />
        )}
        <TextField
          label={t("projects.plural")}
          value={form.fields.plural}
          onValueChange={form.setters.setPlural}
          error={show(form.errors.plural)}
          autoComplete="off"
          spellCheck={false}
        />
      </form>
    </Dialog>
  );
}
