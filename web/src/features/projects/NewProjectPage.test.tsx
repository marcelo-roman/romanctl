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
    await user.type(screen.getByLabelText("Nome do projeto"), "loja");
    await user.click(screen.getByRole("button", { name: "Criar projeto" }));
    await waitFor(() =>
      expect(postedJobs).toEqual([
        { action: "project.create", name: "loja", stack: "go", arch: ["ddd", "hexagonal"], with: ["postgres"], git: true },
      ]),
    );
    await waitFor(() => expect(screen.getByTestId("location")).toHaveTextContent("/jobs/7"));
  });

  it("builds the simple-type body without git when unchecked", async () => {
    const { user } = renderApp("/projects/new");
    await user.click(await screen.findByRole("button", { name: /^node/ }));
    await user.type(screen.getByLabelText("Nome do projeto"), "1api");
    await user.click(screen.getByRole("checkbox", { name: "Iniciar um repositório git" }));
    await user.click(screen.getByRole("button", { name: "Criar projeto" }));
    await waitFor(() => expect(postedJobs).toEqual([{ action: "project.create", name: "1api", type: "node", git: false }]));
  });

  it("rejects invalid names and template names that start with a digit", async () => {
    const { user } = renderApp("/projects/new");
    await user.click(await screen.findByRole("button", { name: /go\/hexagonal-ddd/ }));
    const name = screen.getByLabelText("Nome do projeto");
    await user.type(name, "Loja");
    await user.click(screen.getByRole("button", { name: "Criar projeto" }));
    expect(screen.getByText(/Use letras minúsculas/)).toBeInTheDocument();
    await user.clear(name);
    await user.type(name, "1loja");
    expect(screen.getByText("Para templates o nome precisa começar com uma letra.")).toBeInTheDocument();
    expect(postedJobs).toEqual([]);
  });
});
