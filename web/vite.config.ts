/// <reference types="vitest/config" />
import { defineConfig, loadEnv, type Plugin } from "vite";
import react from "@vitejs/plugin-react";

const tokenPattern = /<meta name="romanctl-token" content="([^"]+)">/;

async function backendToken(target: string, fallback: string): Promise<string> {
  if (fallback) return fallback;
  try {
    const response = await fetch(`${target}/`);
    const match = tokenPattern.exec(await response.text());
    return match?.[1] ?? "";
  } catch {
    return "";
  }
}

function devToken(target: string, fallback: string): Plugin {
  return {
    name: "romanctl-dev-token",
    apply: "serve",
    async transformIndexHtml(html) {
      const token = await backendToken(target, fallback);
      if (!token) return html;
      return html.replace("</head>", `<meta name="romanctl-token" content="${token}"></head>`);
    },
  };
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const target = env.ROMANCTL_WEBUI_URL || "http://localhost:7777";
  return {
    base: "/",
    plugins: [react(), devToken(target, env.VITE_ROMANCTL_TOKEN ?? "")],
    build: {
      outDir: "../internal/webui/dist",
      emptyOutDir: true,
      assetsDir: "assets",
    },
    server: {
      proxy: {
        "/api": {
          target,
          changeOrigin: true,
          configure: (proxy) => {
            proxy.on("proxyReq", (request) => request.removeHeader("origin"));
          },
        },
      },
    },
    test: {
      environment: "jsdom",
      setupFiles: ["./src/test/setup.ts"],
      css: false,
    },
  };
});
