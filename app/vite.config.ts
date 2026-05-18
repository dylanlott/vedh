import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const devPort = Number(env.VITE_DEV_PORT || 5173);

  return {
    plugins: [vue()],
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
        },
      },
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
          target: 'http://localhost:8080',
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
