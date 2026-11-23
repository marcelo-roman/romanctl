import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderApp } from "../../test/render";
import { postedJobs } from "../../test/server";

describe("projects page", () => {
  it("lists projects with their status and url", async () => {
    renderApp("/projects");
    const store = await screen.findByRole("row", { name: /store/ });
    expect(within(store).getByText("go/hexagonal-ddd")).toBeInTheDocument();
    expect(within(store).getByText("down")).toBeInTheDocument();
    const site = screen.getByRole("row", { name: /site/ });
    expect(within(site).getByText("up")).toBeInTheDocument();
    expect(within(site).getByRole("link", { name: /localhost:5173/ })).toHaveAttribute("href", "http://localhost:5173");
  });

  it("posts a project.up job for a stopped project", async () => {
    const { user } = renderApp("/projects");
    const store = await screen.findByRole("row", { name: /store/ });
    await user.click(within(store).getByRole("button", { name: "Up" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.up", name: "store" }]));
  });

  it("posts project.down for a running project", async () => {
    const { user } = renderApp("/projects");
    const site = await screen.findByRole("row", { name: /site/ });
    await user.click(within(site).getByRole("button", { name: "Down" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.down", name: "site" }]));
  });

  it("requires typing the name before removing", async () => {
    const { user } = renderApp("/projects");
    const store = await screen.findByRole("row", { name: /store/ });
    await user.click(within(store).getByRole("button", { name: "Remove" }));
    const dialog = await screen.findByRole("dialog", { name: "Remove store" });
    const confirm = within(dialog).getByRole("button", { name: "Remove" });
    expect(confirm).toBeDisabled();
    await user.type(within(dialog).getByLabelText("Type store to confirm"), "store");
    await user.click(confirm);
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.remove", name: "store" }]));
  });

  it("generates code with the chosen generator and entity", async () => {
    const { user } = renderApp("/projects");
    const store = await screen.findByRole("row", { name: /store/ });
    await user.click(within(store).getByRole("button", { name: "Generate code" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Generate code" }));
    expect(within(dialog).getByText("Required.")).toBeInTheDocument();
    await user.type(within(dialog).getByLabelText("Entity name"), "Product");
    await user.click(within(dialog).getByRole("button", { name: "Generate code" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.generate", name: "store", kind: "resource", entity: "Product" }]));
  });

  it("closes dialogs with Escape", async () => {
    const { user } = renderApp("/projects");
    const store = await screen.findByRole("row", { name: /store/ });
    await user.click(within(store).getByRole("button", { name: "Logs" }));
    expect(await screen.findByText("listening on :8090")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
