import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver({ importStyle: 'sass' })],
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:12566',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            // Element Plus 按组件拆分
            if (id.includes('element-plus/es/components/table')) return 'el-table'
            if (id.includes('element-plus/es/components/form')) return 'el-form'
            if (id.includes('element-plus/es/components/dialog')) return 'el-dialog'
            if (id.includes('element-plus/es/components/pagination')) return 'el-pagination'
            if (id.includes('element-plus/es/components/date-picker')) return 'el-date-picker'
            if (id.includes('element-plus/es/components/select')) return 'el-select'
            if (id.includes('element-plus/es/components/menu')) return 'el-menu'
            if (id.includes('element-plus/es/components/dropdown')) return 'el-dropdown'
            // Element Plus 核心和其他组件
            if (id.includes('element-plus')) return 'element-plus-core'
            if (id.includes('@element-plus/icons-vue')) return 'element-icons'
            // Vue 核心
            if (id.includes('vue') || id.includes('vue-router') || id.includes('pinia')) return 'vue-vendor'
          }
        }
      }
    }
  }
})
