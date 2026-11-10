import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderApp } from "../../test/render";
import { postedJobs } from "../../test/server";

function toolNames(): string[] {
  return screen.getAllByRole("listitem").filter((item) => item.classList.contains("tool-row")).map((item) => item.querySelector("code")?.textContent ?? "");
}

describe("tools page", () => {
  it("filters by search text and by category", async () => {
    const { user } = renderApp("/tools");
    await screen.findByText("Node.js");
    expect(toolNames()).toEqual(["node", "go", "kubectl", "helm"]);

    await user.type(screen.getByRole("searchbox", { name: "Buscar ferramenta" }), "kube");
    expect(toolNames()).toEqual(["kubectl"]);
    expect(screen.getByText("1 de 4")).toBeInTheDocument();

    await user.clear(screen.getByRole("searchbox", { name: "Buscar ferramenta" }));
    await user.click(screen.getByRole("button", { name: /Linguagens/ }));
    expect(toolNames()).toEqual(["node", "go"]);

    await user.click(screen.getByRole("checkbox", { name: "Só instaladas" }));
    expect(toolNames()).toEqual(["node"]);
  });

  it("flags updates and greys out unavailable tools", async () => {
    renderApp("/tools");
    const node = (await screen.findByText("Node.js")).closest("li");
    expect(node).not.toBeNull();
    expect(within(node as HTMLElement).getByText("atualização disponível")).toBeInTheDocument();
    const kubectl = screen.getByText("kubectl", { selector: ".strong" }).closest("li") as HTMLElement;
    expect(within(kubectl).queryByText("atualização disponível")).not.toBeInTheDocument();
    const helm = screen.getByText("Helm").closest("li") as HTMLElement;
    expect(helm).toHaveClass("tool-unavailable");
    expect(within(helm).getByText("só existe para macOS")).toBeInTheDocument();
    expect(within(helm).queryByRole("button")).not.toBeInTheDocument();
  });

  it("installs a chosen version", async () => {
    const { user } = renderApp("/tools");
    const go = (await screen.findByText("Go")).closest("li") as HTMLElement;
    await user.click(within(go).getByRole("button", { name: "Escolher versão" }));
    const dialog = await screen.findByRole("dialog", { name: "Versões de Go" });
    await user.click(await within(dialog).findByRole("button", { name: "Instalar 1.0.0" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "tool.install", key: "go", version: "1.0.0" }]));
  });

  it("asks before uninstalling", async () => {
    const { user } = renderApp("/tools");
    const node = (await screen.findByText("Node.js")).closest("li") as HTMLElement;
    await user.click(within(node).getByRole("button", { name: "Remover" }));
    const dialog = await screen.findByRole("dialog", { name: "Remover Node.js?" });
    expect(postedJobs).toEqual([]);
    await user.click(within(dialog).getByRole("button", { name: "Remover" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "tool.uninstall", key: "node" }]));
  });
});
