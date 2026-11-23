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

    await user.type(screen.getByRole("searchbox", { name: "Search tools" }), "kube");
    expect(toolNames()).toEqual(["kubectl"]);
    expect(screen.getByText("1 of 4")).toBeInTheDocument();

    await user.clear(screen.getByRole("searchbox", { name: "Search tools" }));
    await user.click(screen.getByRole("button", { name: /Languages/ }));
    expect(toolNames()).toEqual(["node", "go"]);

    await user.click(screen.getByRole("checkbox", { name: "Installed only" }));
    expect(toolNames()).toEqual(["node"]);
  });

  it("flags updates and greys out unavailable tools", async () => {
    renderApp("/tools");
    const node = (await screen.findByText("Node.js")).closest("li");
    expect(node).not.toBeNull();
    expect(within(node as HTMLElement).getByText("update available")).toBeInTheDocument();
    const kubectl = screen.getByText("kubectl", { selector: ".strong" }).closest("li") as HTMLElement;
    expect(within(kubectl).queryByText("update available")).not.toBeInTheDocument();
    const helm = screen.getByText("Helm").closest("li") as HTMLElement;
    expect(helm).toHaveClass("tool-unavailable");
    expect(within(helm).getByText("only available on macOS")).toBeInTheDocument();
    expect(within(helm).queryByRole("button")).not.toBeInTheDocument();
  });

  it("installs a chosen version", async () => {
    const { user } = renderApp("/tools");
    const go = (await screen.findByText("Go")).closest("li") as HTMLElement;
    await user.click(within(go).getByRole("button", { name: "Choose version" }));
    const dialog = await screen.findByRole("dialog", { name: "Versions of Go" });
    await user.click(await within(dialog).findByRole("button", { name: "Install 1.0.0" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "tool.install", key: "go", version: "1.0.0" }]));
  });

  it("asks before uninstalling", async () => {
    const { user } = renderApp("/tools");
    const node = (await screen.findByText("Node.js")).closest("li") as HTMLElement;
    await user.click(within(node).getByRole("button", { name: "Uninstall" }));
    const dialog = await screen.findByRole("dialog", { name: "Uninstall Node.js?" });
    expect(postedJobs).toEqual([]);
    await user.click(within(dialog).getByRole("button", { name: "Uninstall" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "tool.uninstall", key: "node" }]));
  });
});
