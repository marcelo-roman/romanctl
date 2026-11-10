import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { info, testToken } from "../test/fixtures";
import { server } from "../test/server";
import { api, ApiError, tokenHeader } from "./client";

describe("api client", () => {
  it("sends the page token on every request", async () => {
    const seen: (string | null)[] = [];
    server.use(
      http.get("/api/info", ({ request }) => {
        seen.push(request.headers.get(tokenHeader));
        return HttpResponse.json(info);
      }),
      http.post("/api/jobs", ({ request }) => {
        seen.push(request.headers.get(tokenHeader));
        return HttpResponse.json({ id: "1" }, { status: 202 });
      }),
    );
    await api.info();
    await api.startJob({ action: "templates.update" });
    expect(seen).toEqual([testToken, testToken]);
  });

  it("surfaces the backend error field", async () => {
    server.use(http.post("/api/jobs", () => HttpResponse.json({ error: "valor inválido em \"name=-rf\"" }, { status: 400 })));
    const failure = api.startJob({ action: "project.up", name: "-rf" });
    await expect(failure).rejects.toBeInstanceOf(ApiError);
    await expect(failure).rejects.toThrow('valor inválido em "name=-rf"');
  });

  it("falls back to the HTTP status when the body has no error", async () => {
    server.use(http.get("/api/info", () => new HttpResponse(null, { status: 500 })));
    await expect(api.info()).rejects.toThrow("HTTP 500");
  });
});
