<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { http } from '../api/http'

const router = useRouter()
const errorMessage = ref('')
const successMessage = ref('')
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

function validate() {
  if (form.username.length < 4 || form.username.length > 20) {
    return '用户名长度需在 4 到 20 位之间'
  }
  if (form.password.length < 6 || form.password.length > 20) {
    return '密码长度需在 6 到 20 位之间'
  }
  if (form.password !== form.confirmPassword) {
    return '两次输入的密码不一致'
  }
  return ''
}

async function register() {
  errorMessage.value = ''
  successMessage.value = ''
  const validationMessage = validate()
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }
  loading.value = true
  try {
    const { data } = await http.post('/auth/register', form)
    successMessage.value = `注册成功：${data.user.username}，正在跳转登录页。`
    const username = form.username
    form.username = ''
    form.password = ''
    form.confirmPassword = ''
    setTimeout(() => {
      router.push({ path: '/login', query: { username } })
    }, 800)
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
        <p class="apple-label">Create account</p>
        <h1 class="apple-title">注册后立即进入 IM 通讯空间。</h1>
        <p class="apple-subtitle">
          完成注册后即可使用消息会话、好友通讯录、群组沟通与个人主页等即时通讯功能。
        </p>
        <div class="hero-highlights">
          <span>消息收发</span>
          <span>好友添加</span>
          <span>群组沟通</span>
          <span>个人名片</span>
        </div>
      </div>
      <div class="card auth-card">
        <div class="auth-card-head">
          <h2>注册</h2>
          <p class="muted">填写基础信息，快速创建新账号。</p>
        </div>
        <p v-if="successMessage" class="status-text success">{{ successMessage }}</p>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <input v-model="form.username" class="apple-input" placeholder="用户名（4-20 位）" />
        <input v-model="form.password" class="apple-input" type="password" placeholder="密码（6-20 位）" />
        <input v-model="form.confirmPassword" class="apple-input" type="password" placeholder="确认密码" />
        <button class="apple-button" :disabled="loading" @click="register">注册账号</button>
        <p class="switch-text muted">
          已有账号？
          <RouterLink to="/login">返回登录</RouterLink>
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

.hero-highlights {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-highlights span {
  padding: 10px 14px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: inset 0 0 0 1px rgba(29, 29, 31, 0.06);
  color: #0f172a;
  font-size: 13px;
  font-weight: 600;
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
