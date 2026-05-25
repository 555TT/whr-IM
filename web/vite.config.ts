import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import basicSsl from '@vitejs/plugin-basic-ssl'

// HTTPS 仅在测试手机时才需要（浏览器只在 secure context 下暴露 window.crypto.subtle，
// 通过电脑 IP 访问的手机属于非 secure context）。
// 默认关闭：HTTPS 页面访问 HTTP 接口会被 Safari 等浏览器以"混合内容"拦截，
// 直接表现为登录"请求失败"。
//
// 调试手机时启用方式（任选其一）：
//   1) VITE_HTTPS=true npm run dev
//   2) 设置 VITE_API_BASE_URL=/api, VITE_WS_BASE_URL=/ws 让 Vite 反代后端
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const useHttps = env.VITE_HTTPS === 'true'

  return {
    plugins: [vue(), ...(useHttps ? [basicSsl()] : [])],
    server: {
      port: 5173,
      host: '0.0.0.0',
      https: useHttps,
      // 当且仅当前端用相对路径 /api、/ws 时，下面的代理才生效
      proxy: {
        '/api': {
          target: 'http://127.0.0.1:8080',
          changeOrigin: true
        },
        '/ws': {
          target: 'ws://127.0.0.1:8080',
          ws: true,
          changeOrigin: true
        }
      }
    }
  }
})
