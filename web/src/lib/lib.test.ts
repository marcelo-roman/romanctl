import { describe, expect, it } from "vitest";
import { catalog } from "../test/fixtures";
import { formatBytes, formatDuration } from "./format";
import { buildCreateBody, jobCommand, relatedQueryKeys } from "./jobs";
import { validateEntity, validateProjectName } from "./projectName";
import { isNewer, shortVersion } from "./version";

describe("version", () => {
  it("compares versions embedded in tool output", () => {
    expect(isNewer("v24.2.0", "v24.1.0")).toBe(true);
    expect(isNewer("1.31.0", "1.31.0")).toBe(false);
    expect(isNewer("1.10.0", "1.9.9")).toBe(true);
    expect(isNewer("go1.27.3", "go: downloading go1.27.3 (linux/amd64)")).toBe(false);
    expect(isNewer(undefined, "1.0")).toBe(false);
    expect(shortVersion("Docker version 29.8.0, build 88096ef")).toBe("29.8.0");
  });
});

describe("project name", () => {
  it("follows the backend rules", () => {
    expect(validateProjectName("loja", true)).toBeNull();
    expect(validateProjectName("1loja", false)).toBeNull();
    expect(validateProjectName("1loja", true)).toBe("validation.nameLetter");
    expect(validateProjectName("Loja", false)).toBe("validation.namePattern");
    expect(validateProjectName("-rf", false)).toBe("validation.namePattern");
    expect(validateProjectName("", false)).toBe("validation.nameRequired");
    expect(validateEntity("Product", true)).toBeNull();
    expect(validateEntity("--force", true)).toBe("validation.entityPattern");
    expect(validateEntity("", false)).toBeNull();
  });
});

describe("jobs", () => {
  it("builds create bodies and commands", () => {
    const template = catalog.templates[1];
    if (!template) throw new Error("fixture");
    expect(buildCreateBody({ source: { kind: "template", template }, name: "loja", services: [], git: true })).toEqual({
      action: "project.create",
      name: "loja",
      stack: "go",
      arch: ["ddd", "hexagonal"],
      with: [],
      git: true,
    });
    expect(jobCommand({ args: ["new", "loja", "--stack", "go"] })).toBe("romanctl new loja --stack go");
    expect(relatedQueryKeys("tool.install")).toContainEqual(["tools"]);
    expect(relatedQueryKeys("service.stop")).toContainEqual(["services"]);
  });
});

describe("format", () => {
  it("formats bytes and durations", () => {
    expect(formatBytes(1536)).toBe("1.5 KB");
    expect(formatBytes(0)).toBe("0 B");
    expect(formatDuration("2026-01-01T00:00:00Z", "2026-01-01T00:01:05Z")).toBe("1m 5s");
  });
});
