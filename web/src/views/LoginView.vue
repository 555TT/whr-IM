<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const errorMessage = ref('')
const loading = ref(false)

const form = reactive({
  username: typeof route.query.username === 'string' ? route.query.username : '',
  password: ''
})

watch(
  () => route.query.username,
  (value) => {
    form.username = typeof value === 'string' ? value : ''
  }
)

function validate() {
  if (form.username.length < 4 || form.username.length > 20) {
    return '用户名长度需在 4 到 20 位之间'
  }
  if (form.password.length < 6 || form.password.length > 20) {
    return '密码长度需在 6 到 20 位之间'
  }
  return ''
}

async function login() {
  errorMessage.value = validate()
  if (errorMessage.value) return
  loading.value = true
  try {
    const { data } = await http.post('/auth/login', {
      username: form.username,
      password: form.password
    })
    authStore.setSession(data.token, data.user)
    router.push('/chat')
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-shell auth-page">
    <div class="auth-hero">
      <div class="hero-copy">
        <p class="apple-label">Easy Chat</p>
        <h1 class="apple-title">简单、克制、专注的即时沟通体验。</h1>
        <p class="apple-subtitle">
          面向课程项目的 Web 端 IM，提供账号、好友、聊天与资料管理能力。
        </p>
      </div>
      <div class="card auth-card">
        <div class="auth-card-head">
          <h2>登录</h2>
          <p class="muted">使用你的账号继续进入聊天空间。</p>
        </div>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <input v-model="form.username" class="apple-input" placeholder="用户名（4-20 位）" />
        <input v-model="form.password" class="apple-input" type="password" placeholder="密码（6-20 位）" />
        <button class="apple-button" :disabled="loading" @click="login">登录</button>
        <p class="switch-text muted">
          还没有账号？
          <RouterLink to="/register">去注册</RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 24px;
}

.auth-hero {
  width: min(1120px, 100%);
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 28px;
  align-items: center;
}

.hero-copy {
  padding: 32px 12px 32px 8px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.auth-card {
  padding: 30px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.auth-card-head h2 {
  margin: 0 0 8px;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.auth-card-head p {
  margin: 0;
}

.switch-text {
  margin: 0;
}

.switch-text a {
  color: #0071e3;
  text-decoration: none;
  font-weight: 600;
}

@media (max-width: 900px) {
  .auth-hero {
    grid-template-columns: 1fr;
  }

  .hero-copy {
    padding: 8px 0;
  }
}
</style>
