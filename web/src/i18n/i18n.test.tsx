import { screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { info } from "../test/fixtures";
import { renderApp } from "../test/render";
import { server } from "../test/server";
import { translate } from "./context";
import { dictionaries } from "./messages";

describe("i18n", () => {
  it("follows a Portuguese backend", async () => {
    server.use(http.get("/api/info", () => HttpResponse.json({ ...info, lang: "pt" })));
    renderApp("/projects");
    expect(await screen.findByRole("link", { name: "Projetos" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Projetos", level: 1 })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /^Subir$/ })).toBeInTheDocument();
    expect(document.documentElement.lang).toBe("pt-BR");
  });

  it("defaults to English and switches to Portuguese from the header", async () => {
    const { user } = renderApp("/");
    expect(await screen.findByRole("heading", { name: "Dashboard", level: 1 })).toBeInTheDocument();
    expect(document.documentElement.lang).toBe("en");
    await user.selectOptions(screen.getByLabelText("Language"), "pt");
    await waitFor(() => expect(screen.getByRole("heading", { name: "Painel", level: 1 })).toBeInTheDocument());
    expect(screen.getByRole("link", { name: "Ferramentas" })).toBeInTheDocument();
  });

  it("interpolates variables and has every key in both languages", () => {
    expect(translate("en", "tools.count", { shown: 2, total: 5 })).toBe("2 of 5");
    expect(translate("pt", "tools.count", { shown: 2, total: 5 })).toBe("2 de 5");
    const enKeys = Object.keys(dictionaries.en).sort();
    expect(Object.keys(dictionaries.pt).sort()).toEqual(enKeys);
    for (const key of enKeys) expect(dictionaries.pt[key as keyof typeof dictionaries.pt]).not.toBe("");
  });
});
