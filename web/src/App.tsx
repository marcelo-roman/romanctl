import { Route, Routes } from "react-router";
import { Layout } from "./layout/Layout";
import { DashboardPage } from "./features/dashboard/DashboardPage";
import { ProjectsPage } from "./features/projects/ProjectsPage";
import { NewProjectPage } from "./features/projects/NewProjectPage";
import { ServicesPage } from "./features/services/ServicesPage";
import { ToolsPage } from "./features/tools/ToolsPage";
import { TemplatesPage } from "./features/templates/TemplatesPage";
import { JobsPage } from "./features/jobs/JobsPage";
import { JobDetailPage } from "./features/jobs/JobDetailPage";
import { NotFoundPage } from "./features/NotFoundPage";

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<DashboardPage />} />
        <Route path="projects" element={<ProjectsPage />} />
        <Route path="projects/new" element={<NewProjectPage />} />
        <Route path="services" element={<ServicesPage />} />
        <Route path="tools" element={<ToolsPage />} />
        <Route path="templates" element={<TemplatesPage />} />
        <Route path="jobs" element={<JobsPage />} />
        <Route path="jobs/:id" element={<JobDetailPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
