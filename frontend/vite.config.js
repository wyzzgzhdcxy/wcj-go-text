import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  optimizeDeps: {
    include: ['vue', 'vue-router', 'element-plus', '@element-plus/icons-vue']
  },
  build: {
    minify: 'esbuild',
    target: 'es2015',
    cssCodeSplit: true
  }
})
