import { Link } from "react-router";
import { useJobs, useProjects, useServices } from "../../api/queries";
import { ExternalLink } from "../../components/ExternalLink";
import { PageHeader, Section } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { isProjectUp, projectUrl } from "../../lib/catalog";
import { JobList } from "../jobs/JobList";
import { MachineSummary } from "./MachineSummary";

function ProjectsUp() {
  const { t } = useI18n();
  const projects = useProjects();
  return (
    <Section title={t("dashboard.projectsUp")}>
      <QueryState query={projects}>
        {(list) => {
          const up = list.filter(isProjectUp);
          if (up.length === 0) return <p className="muted">{t("dashboard.noProjectsUp")}</p>;
          return (
            <ul className="plain-list">
              {up.map((project) => (
                <li key={project.name} className="row-between">
                  <Link className="link strong" to="/projects">
                    {project.name}
                  </Link>
                  <ExternalLink url={projectUrl(project)} />
                </li>
              ))}
            </ul>
          );
        }}
      </QueryState>
    </Section>
  );
}

function ServicesRunning() {
  const { t } = useI18n();
  const services = useServices();
  return (
    <Section title={t("dashboard.servicesRunning")}>
      <QueryState query={services}>
        {(list) => {
          const running = list.filter((service) => service.running);
          if (running.length === 0) return <p className="muted">{t("dashboard.noServicesRunning")}</p>;
          return (
            <ul className="plain-list">
              {running.map((service) => (
                <li key={service.key} className="row-between">
                  <span className="strong">{service.name}</span>
                  <code className="mono small truncate">{service.url}</code>
                </li>
              ))}
            </ul>
          );
        }}
      </QueryState>
    </Section>
  );
}

function RecentJobs() {
  const { t } = useI18n();
  const jobs = useJobs();
  const seeAll = (
    <Link className="link small" to="/jobs">
      {t("dashboard.seeAll")}
    </Link>
  );
  return (
    <Section title={t("dashboard.recentJobs")} actions={seeAll}>
      <QueryState query={jobs}>{(list) => <JobList jobs={list.slice(0, 5)} emptyText={t("dashboard.noJobs")} />}</QueryState>
    </Section>
  );
}

export function DashboardPage() {
  const { t } = useI18n();
  return (
    <>
      <PageHeader title={t("dashboard.title")} />
      <div className="grid">
        <MachineSummary />
        <ProjectsUp />
        <ServicesRunning />
        <RecentJobs />
      </div>
    </>
  );
}
