import {defineConfig} from "vite";

export default defineConfig({
    build: {
        outDir: "../backend/web",
        emptyOutDir: true
    },
    server: {
        port: 5173,
        host: true,
        proxy: {
            "/api": {
                target: "http://127.0.0.1:19528",
                changeOrigin: true,
                ws: true
            }
        }
    },
    preview: {
        port: 4173,
        host: true,
        proxy: {
            "/api": {
                target: "http://127.0.0.1:19528",
                changeOrigin: true,
                ws: true
            }
        }
    }
});