import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import electron from 'vite-plugin-electron'
import renderer from 'vite-plugin-electron-renderer'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vue(),
    electron([
      {
        entry: 'electron/main.ts',
        vite: {
          build: {
            rollupOptions: {
              external: ['ssh2', 'cpu-features'],
            },
          },
        },
      },
      { entry: 'electron/preload.ts', onstart(args) { args.reload() } },
    ]),
    renderer(),
  ],
  resolve: {
    alias: { '@': resolve(__dirname, 'src') },
  },
})
