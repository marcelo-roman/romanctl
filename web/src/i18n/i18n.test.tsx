import { screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { info } from "../test/fixtures";
import { renderApp } from "../test/render";
import { server } from "../test/server";
import { translate } from "./context";
import { dictionaries } from "./messages";

describe("i18n", () => {
  it("follows the backend language", async () => {
    server.use(http.get("/api/info", () => HttpResponse.json({ ...info, lang: "en" })));
    renderApp("/projects");
    expect(await screen.findByRole("link", { name: "Projects" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Projects", level: 1 })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /^Up$/ })).toBeInTheDocument();
    expect(document.documentElement.lang).toBe("en");
  });

  it("defaults to Portuguese and switches to English from the header", async () => {
    const { user } = renderApp("/");
    expect(await screen.findByRole("heading", { name: "Painel", level: 1 })).toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText("Idioma"), "en");
    await waitFor(() => expect(screen.getByRole("heading", { name: "Dashboard", level: 1 })).toBeInTheDocument());
    expect(screen.getByRole("link", { name: "Tools" })).toBeInTheDocument();
  });

  it("interpolates variables and has every key in both languages", () => {
    expect(translate("en", "tools.count", { shown: 2, total: 5 })).toBe("2 of 5");
    expect(translate("pt", "tools.count", { shown: 2, total: 5 })).toBe("2 de 5");
    const ptKeys = Object.keys(dictionaries.pt).sort();
    expect(Object.keys(dictionaries.en).sort()).toEqual(ptKeys);
    for (const key of ptKeys) expect(dictionaries.en[key as keyof typeof dictionaries.en]).not.toBe("");
  });
});
