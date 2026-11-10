import { render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes, useLocation } from "react-router";
import { App } from "../App";
import { Providers } from "../Providers";

export function LocationProbe() {
  const location = useLocation();
  return <output data-testid="location">{location.pathname}</output>;
}

export function renderApp(path = "/") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const user = userEvent.setup();
  const result = render(
    <MemoryRouter initialEntries={[path]}>
      <Providers client={client}>
        <App />
        <Routes>
          <Route path="*" element={<LocationProbe />} />
        </Routes>
      </Providers>
    </MemoryRouter>,
  );
  return { ...result, user, client };
}
