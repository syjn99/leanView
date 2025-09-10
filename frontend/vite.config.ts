import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig } from 'vite';

// https://vite.dev/config/
export default defineConfig({
    plugins: [react(), tailwindcss()],
    base: '/',
    build: {
        outDir: 'dist',
        sourcemap: false, // No source maps in production
        rollupOptions: {
            output: {
                // Add hash to filenames for cache busting
                entryFileNames: 'assets/[name].[hash].js',
                chunkFileNames: 'assets/[name].[hash].js',
                assetFileNames: 'assets/[name].[hash].[ext]',
            },
        },
    },
    resolve: { alias: { '@': path.resolve(__dirname, './src') } },
});
