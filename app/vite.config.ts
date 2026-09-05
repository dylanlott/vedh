import { fileURLToPath, URL } from 'node:url';
import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '..', '');
  const devPort = Number(env.VITE_DEV_PORT || 5173);
  const apiPort = Number(env.PORT || 8080);

  return {
    envDir: '..',
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
        },
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      include: ['__tests__/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
      exclude: ['e2e/**', 'dist/**', 'node_modules/**'],
    },
    server: {
      port: devPort,
      // Allow automatic fallback (e.g. 5174) when the preferred port is busy.
      strictPort: false,
      watch: {
        usePolling: true,
        interval: 100,
      },
      proxy: {
        '/graphql': {
          target: `http://localhost:${apiPort}`,
          changeOrigin: true,
          ws: true,
          secure: false,
        },
      },
    },
    define: {
      __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
    },
  };
});
