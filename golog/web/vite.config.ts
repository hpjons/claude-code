import { defineConfig } from "vite";

// `npm run dev` serves the app with live reload; API calls go to a running `golog -no-browser`.
export default defineConfig({
  build: { outDir: "dist", emptyOutDir: true, assetsInlineLimit: 0 },
  server: { proxy: { "/api": "http://127.0.0.1:7341" } },
});
