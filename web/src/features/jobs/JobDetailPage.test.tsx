import { screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { job } from "../../test/fixtures";
import { renderApp } from "../../test/render";
import { server } from "../../test/server";

describe("job detail", () => {
  it("polls while running, shows progress lines and then the result", async () => {
    let calls = 0;
    server.use(
      http.get("/api/jobs/:id", () => {
        calls++;
        if (calls === 1) return HttpResponse.json(job({ id: "3", lines: ["Baixando o template…"] }));
        return HttpResponse.json(
          job({
            id: "3",
            status: "succeeded",
            lines: ["Baixando o template…", "Pronto"],
            finished: "2026-11-10T12:00:04Z",
            result: { ok: true, path: "/tmp/proj/loja" },
          }),
        );
      }),
    );
    renderApp("/jobs/3");
    expect(await screen.findByText("romanctl up loja -d")).toBeInTheDocument();
    expect(screen.getByText("Baixando o template…")).toBeInTheDocument();
    expect(screen.getByText("em andamento")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/Pronto/)).toBeInTheDocument(), { timeout: 3000 });
    expect(screen.getByText("concluída")).toBeInTheDocument();
    expect(screen.getByText(/"path": "\/tmp\/proj\/loja"/)).toBeInTheDocument();
    expect(calls).toBeGreaterThanOrEqual(2);
  });

  it("highlights the error of a failed job", async () => {
    server.use(http.get("/api/jobs/:id", () => HttpResponse.json(job({ id: "4", status: "failed", error: "porta ocupada", finished: "2026-11-10T12:00:01Z" }))));
    renderApp("/jobs/4");
    expect(await screen.findByRole("alert")).toHaveTextContent("porta ocupada");
    expect(screen.getByText("falhou")).toBeInTheDocument();
  });

  it("shows a toast and refreshes projects when a started job finishes", async () => {
    let projectCalls = 0;
    let jobCalls = 0;
    server.use(
      http.post("/api/jobs", () => HttpResponse.json(job({ id: "7", title: "Subir loja" }), { status: 202 })),
      http.get("/api/jobs/:id", () => {
        jobCalls++;
        return HttpResponse.json(job({ id: "7", title: "Subir loja", status: jobCalls > 1 ? "succeeded" : "running" }));
      }),
      http.get("/api/projects", () => {
        projectCalls++;
        return HttpResponse.json([baseProject]);
      }),
    );
    const { user } = renderApp("/projects");
    await user.click(await screen.findByRole("button", { name: "Subir" }));
    expect(await screen.findByText("Concluída: Subir loja", {}, { timeout: 3000 })).toBeInTheDocument();
    await waitFor(() => expect(projectCalls).toBeGreaterThanOrEqual(2));
  });
});

const baseProject = {
  name: "loja",
  path: "/tmp/proj/loja",
  type: "go",
  modified: "2026-11-10T12:00:00Z",
  url: "",
  target: "",
  generators: [],
};
