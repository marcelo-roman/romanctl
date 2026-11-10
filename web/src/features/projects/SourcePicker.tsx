import type { Catalog, Template } from "../../api/types";
import { Badge } from "../../components/Badge";
import { useI18n } from "../../i18n/context";
import { groupTemplatesByStack } from "../../lib/catalog";

interface SourcePickerProps {
  catalog: Catalog;
  selectedKey: string;
  onSimple: (type: string) => void;
  onTemplate: (template: Template) => void;
}

function TemplateCard({ template, selected, onSelect }: { template: Template; selected: boolean; onSelect: () => void }) {
  const { t } = useI18n();
  const services = template.services ?? [];
  const requires = template.requires ?? [];
  return (
    <button type="button" className={selected ? "card card-selectable card-selected" : "card card-selectable"} aria-pressed={selected} onClick={onSelect}>
      <span className="card-title mono">
        {template.stack}/{template.id}
      </span>
      <span className="tags">
        {template.arch.map((tag) => (
          <Badge key={tag} tone="info">
            {tag}
          </Badge>
        ))}
      </span>
      <span className="card-text">{template.description}</span>
      {services.length > 0 && <span className="muted small">{t("projects.services", { list: services.join(", ") })}</span>}
      {requires.length > 0 && <span className="muted small">{t("wizard.requires", { list: requires.join(", ") })}</span>}
    </button>
  );
}

export function SourcePicker({ catalog, selectedKey, onSimple, onTemplate }: SourcePickerProps) {
  const { t } = useI18n();
  const groups = groupTemplatesByStack(catalog);
  return (
    <div className="stack">
      <h3 className="subheading">{t("wizard.simple")}</h3>
      <div className="card-grid card-grid-sm">
        {catalog.simpleTypes.map((simple) => {
          const selected = selectedKey === `simple:${simple.type}`;
          return (
            <button
              key={simple.type}
              type="button"
              className={selected ? "card card-selectable card-selected" : "card card-selectable"}
              aria-pressed={selected}
              onClick={() => onSimple(simple.type)}
            >
              <span className="card-title mono">{simple.type}</span>
              <span className="card-text">{simple.description}</span>
            </button>
          );
        })}
      </div>
      <h3 className="subheading">{t("wizard.templates")}</h3>
      {groups.map((group) => (
        <div key={group.stack} className="stack-group">
          <h4 className="stack-name mono">{group.stack}</h4>
          <div className="card-grid">
            {group.templates.map((template) => (
              <TemplateCard
                key={template.id}
                template={template}
                selected={selectedKey === `template:${template.stack}/${template.id}`}
                onSelect={() => onTemplate(template)}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
