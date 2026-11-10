import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll, beforeEach } from "vitest";
import { testToken } from "./fixtures";
import { postedJobs, server } from "./server";

beforeAll(() => server.listen({ onUnhandledFrame: "error" }));

beforeEach(() => {
  const meta = document.createElement("meta");
  meta.name = "romanctl-token";
  meta.content = testToken;
  document.head.appendChild(meta);
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
  postedJobs.length = 0;
  document.head.querySelectorAll('meta[name="romanctl-token"]').forEach((meta) => meta.remove());
  window.localStorage.clear();
});

afterAll(() => server.close());
