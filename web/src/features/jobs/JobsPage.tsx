import { useJobs } from "../../api/queries";
import { PageHeader } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { JobList } from "./JobList";

export function JobsPage() {
  const { t } = useI18n();
  const jobs = useJobs();
  return (
    <>
      <PageHeader title={t("jobs.title")} />
      <section className="panel">
        <QueryState query={jobs}>{(list) => <JobList jobs={list} emptyText={t("jobs.empty")} />}</QueryState>
      </section>
    </>
  );
}
