import { Link } from "react-router";
import { useI18n } from "../i18n/context";

export function NotFoundPage() {
  const { t } = useI18n();
  return (
    <div className="stack">
      <p>{t("errors.notFound")}</p>
      <Link className="link" to="/">
        {t("nav.dashboard")}
      </Link>
    </div>
  );
}
