import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const devPort = Number(env.VITE_DEV_PORT || 5173);

  return {
    plugins: [vue()],
    server: {
      port: devPort,
      strictPort: true,
      hmr: {
        clientPort: devPort,
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
