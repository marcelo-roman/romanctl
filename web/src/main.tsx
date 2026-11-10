import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";
import { QueryClient } from "@tanstack/react-query";
import { App } from "./App";
import { Providers } from "./Providers";
import "./styles/app.css";

const client = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: true } },
});

const root = document.getElementById("root");
if (root) {
  createRoot(root).render(
    <StrictMode>
      <BrowserRouter>
        <Providers client={client}>
          <App />
        </Providers>
      </BrowserRouter>
    </StrictMode>,
  );
}
