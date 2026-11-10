import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import type { Job, JobRequest } from "../api/types";
import * as fixtures from "./fixtures";

export const postedJobs: JobRequest[] = [];

export const handlers = [
  http.get("/api/info", () => HttpResponse.json(fixtures.info)),
  http.get("/api/system", () => HttpResponse.json(fixtures.system)),
  http.get("/api/projects", () => HttpResponse.json(fixtures.projects)),
  http.get("/api/projects/:name/logs", () => HttpResponse.json({ lines: ["listening on :8090"], running: true, path: "/tmp/log" })),
  http.get("/api/catalog", () => HttpResponse.json(fixtures.catalog)),
  http.get("/api/services", () => HttpResponse.json(fixtures.services)),
  http.get("/api/services/:key/logs", () => HttpResponse.json({ lines: [] })),
  http.get("/api/tools", () => HttpResponse.json(fixtures.tools)),
  http.get("/api/tools/:key/versions", () => HttpResponse.json({ versions: [{ id: "1.0.0", label: "1.0.0" }] })),
  http.get("/api/jobs", () => HttpResponse.json([] as Job[])),
  http.get("/api/jobs/:id", ({ params }) => HttpResponse.json(fixtures.job({ id: String(params.id), status: "succeeded", finished: "2026-11-10T12:00:05Z" }))),
  http.post("/api/jobs", async ({ request }) => {
    const body = (await request.json()) as JobRequest;
    postedJobs.push(body);
    return HttpResponse.json(fixtures.job({ id: "7", action: body.action, title: `job ${body.action}` }), { status: 202 });
  }),
];

export const server = setupServer(...handlers);
