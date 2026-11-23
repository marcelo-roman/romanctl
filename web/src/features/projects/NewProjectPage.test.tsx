import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderApp } from "../../test/render";
import { postedJobs } from "../../test/server";

describe("new project wizard", () => {
  it("builds the template body with the chosen services and opens the job", async () => {
    const { user } = renderApp("/projects/new");
    await user.click(await screen.findByRole("button", { name: /go\/hexagonal-ddd/ }));
    expect(screen.getByRole("checkbox", { name: "postgres" })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: "redis" })).toBeChecked();
    await user.click(screen.getByRole("checkbox", { name: "redis" }));
    await user.type(screen.getByLabelText("Project name"), "store");
    await user.click(screen.getByRole("button", { name: "Create project" }));
    await waitFor(() =>
      expect(postedJobs).toEqual([
        { action: "project.create", name: "store", stack: "go", arch: ["ddd", "hexagonal"], with: ["postgres"], git: true },
      ]),
    );
    await waitFor(() => expect(screen.getByTestId("location")).toHaveTextContent("/jobs/7"));
  });

  it("builds the simple-type body without git when unchecked", async () => {
    const { user } = renderApp("/projects/new");
    await user.click(await screen.findByRole("button", { name: /^node/ }));
    await user.type(screen.getByLabelText("Project name"), "1api");
    await user.click(screen.getByRole("checkbox", { name: "Initialize a git repository" }));
    await user.click(screen.getByRole("button", { name: "Create project" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.create", name: "1api", type: "node", git: false }]));
  });

  it("rejects invalid names and template names that start with a digit", async () => {
    const { user } = renderApp("/projects/new");
    await user.click(await screen.findByRole("button", { name: /go\/hexagonal-ddd/ }));
    const name = screen.getByLabelText("Project name");
    await user.type(name, "Store");
    await user.click(screen.getByRole("button", { name: "Create project" }));
    expect(screen.getByText(/Use lowercase letters/)).toBeInTheDocument();
    await user.clear(name);
    await user.type(name, "1store");
    expect(screen.getByText("Template projects must start with a letter.")).toBeInTheDocument();
    expect(postedJobs).toEqual([]);
  });
});
