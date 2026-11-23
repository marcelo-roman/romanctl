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
        if (calls === 1) return HttpResponse.json(job({ id: "3", lines: ["Downloading the template…"] }));
        return HttpResponse.json(
          job({
            id: "3",
            status: "succeeded",
            lines: ["Downloading the template…", "Done"],
            finished: "2026-11-10T12:00:04Z",
            result: { ok: true, path: "/tmp/proj/store" },
          }),
        );
      }),
    );
    renderApp("/jobs/3");
    expect(await screen.findByText("romanctl up store -d")).toBeInTheDocument();
    expect(screen.getByText("Downloading the template…")).toBeInTheDocument();
    expect(screen.getByText("running")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/Done/)).toBeInTheDocument(), { timeout: 3000 });
    expect(screen.getByText("succeeded")).toBeInTheDocument();
    expect(screen.getByText(/"path": "\/tmp\/proj\/store"/)).toBeInTheDocument();
    expect(calls).toBeGreaterThanOrEqual(2);
  });

  it("highlights the error of a failed job", async () => {
    server.use(http.get("/api/jobs/:id", () => HttpResponse.json(job({ id: "4", status: "failed", error: "port in use", finished: "2026-11-10T12:00:01Z" }))));
    renderApp("/jobs/4");
    expect(await screen.findByRole("alert")).toHaveTextContent("port in use");
    expect(screen.getByText("failed")).toBeInTheDocument();
  });

  it("shows a toast and refreshes projects when a started job finishes", async () => {
    let projectCalls = 0;
    let jobCalls = 0;
    server.use(
      http.post("/api/jobs", () => HttpResponse.json(job({ id: "7", title: "Start store" }), { status: 202 })),
      http.get("/api/jobs/:id", () => {
        jobCalls++;
        return HttpResponse.json(job({ id: "7", title: "Start store", status: jobCalls > 1 ? "succeeded" : "running" }));
      }),
      http.get("/api/projects", () => {
        projectCalls++;
        return HttpResponse.json([baseProject]);
      }),
    );
    const { user } = renderApp("/projects");
    await user.click(await screen.findByRole("button", { name: "Up" }));
    expect(await screen.findByText("Finished: Start store", {}, { timeout: 3000 })).toBeInTheDocument();
    await waitFor(() => expect(projectCalls).toBeGreaterThanOrEqual(2));
  });
});

const baseProject = {
  name: "store",
  path: "/tmp/proj/store",
  type: "go",
  modified: "2026-11-10T12:00:00Z",
  url: "",
  target: "",
  generators: [],
};
