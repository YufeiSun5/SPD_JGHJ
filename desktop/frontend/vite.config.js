import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  
  // 🔧 关键配置1: 使用相对路径，Wails必需
  base: './',
  
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@wailsjs': resolve(__dirname, 'wailsjs')
    }
  },
  
  build: {
    // 🔧 关键配置2: 输出到dist目录（与frontend同级）
    outDir: '../dist',
    emptyOutDir: false,
    
    // 🔧 关键配置3: 不要代码分割，避免动态import问题
    rollupOptions: {
      output: {
        manualChunks: undefined,
        // 简化文件名，便于调试
        entryFileNames: 'assets/[name].js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]'
      }
    },
    
    // 关闭source map，减小体积
    sourcemap: false,
    
    // 目标环境
    target: 'esnext',
    minify: 'esbuild' // 使用esbuild压缩，更快
  },
  
  server: {
    port: 34115,
    strictPort: true
  }
})

