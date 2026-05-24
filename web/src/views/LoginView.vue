<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const errorMessage = ref('')
const successMessage = ref('')
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

async function register() {
  errorMessage.value = ''
  successMessage.value = ''
  loading.value = true
  try {
    const { data } = await http.post('/auth/register', form)
    successMessage.value = `注册成功：${data.user.username}`
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

async function login() {
  errorMessage.value = ''
  successMessage.value = ''
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
  <div class="page-shell login-page">
    <div class="card login-card">
      <h1>Easy Chat</h1>
      <p v-if="successMessage" class="success-text">{{ successMessage }}</p>
      <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      <input v-model="form.username" placeholder="用户名" />
      <input v-model="form.password" type="password" placeholder="密码" />
      <input v-model="form.confirmPassword" type="password" placeholder="确认密码" />
      <div class="actions">
        <button :disabled="loading" @click="register">注册</button>
        <button :disabled="loading" @click="login">登录</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

input {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
}

.actions {
  display: flex;
  gap: 12px;
}

button {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: white;
  cursor: pointer;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-text {
  color: #dc2626;
}

.success-text {
  color: #16a34a;
}
</style>
