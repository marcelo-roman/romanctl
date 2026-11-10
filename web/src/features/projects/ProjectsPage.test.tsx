import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderApp } from "../../test/render";
import { postedJobs } from "../../test/server";

describe("projects page", () => {
  it("lists projects with their status and url", async () => {
    renderApp("/projects");
    const loja = await screen.findByRole("row", { name: /loja/ });
    expect(within(loja).getByText("go/hexagonal-ddd")).toBeInTheDocument();
    expect(within(loja).getByText("parado")).toBeInTheDocument();
    const site = screen.getByRole("row", { name: /site/ });
    expect(within(site).getByText("no ar")).toBeInTheDocument();
    expect(within(site).getByRole("link", { name: /localhost:5173/ })).toHaveAttribute("href", "http://localhost:5173");
  });

  it("posts a project.up job for a stopped project", async () => {
    const { user } = renderApp("/projects");
    const loja = await screen.findByRole("row", { name: /loja/ });
    await user.click(within(loja).getByRole("button", { name: "Subir" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.up", name: "loja" }]));
  });

  it("posts project.down for a running project", async () => {
    const { user } = renderApp("/projects");
    const site = await screen.findByRole("row", { name: /site/ });
    await user.click(within(site).getByRole("button", { name: "Derrubar" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.down", name: "site" }]));
  });

  it("requires typing the name before removing", async () => {
    const { user } = renderApp("/projects");
    const loja = await screen.findByRole("row", { name: /loja/ });
    await user.click(within(loja).getByRole("button", { name: "Remover" }));
    const dialog = await screen.findByRole("dialog", { name: "Remover loja" });
    const confirm = within(dialog).getByRole("button", { name: "Remover" });
    expect(confirm).toBeDisabled();
    await user.type(within(dialog).getByLabelText("Digite loja para confirmar"), "loja");
    await user.click(confirm);
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.remove", name: "loja" }]));
  });

  it("generates code with the chosen generator and entity", async () => {
    const { user } = renderApp("/projects");
    const loja = await screen.findByRole("row", { name: /loja/ });
    await user.click(within(loja).getByRole("button", { name: "Gerar código" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Gerar código" }));
    expect(within(dialog).getByText("Obrigatório.")).toBeInTheDocument();
    await user.type(within(dialog).getByLabelText("Nome da entidade"), "Product");
    await user.click(within(dialog).getByRole("button", { name: "Gerar código" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.generate", name: "loja", kind: "resource", entity: "Product" }]));
  });

  it("closes dialogs with Escape", async () => {
    const { user } = renderApp("/projects");
    const loja = await screen.findByRole("row", { name: /loja/ });
    await user.click(within(loja).getByRole("button", { name: "Logs" }));
    expect(await screen.findByText("listening on :8090")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
