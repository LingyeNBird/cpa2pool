import { defineConfig } from 'vite-plus';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';
import { viteSingleFile } from 'vite-plugin-singlefile';

export default defineConfig({
  root: 'web',
  plugins: [vue(), tailwindcss(), viteSingleFile()],
  build: { outDir: '../internal/console/dist', emptyOutDir: true },
  server: { host: '127.0.0.1', port: 5173, strictPort: true, proxy: { '/v0': 'http://127.0.0.1:18317' } },
  fmt: { printWidth: 100, singleQuote: true },
  lint: { ignorePatterns: ['.local/**', 'internal/**'] },
});
