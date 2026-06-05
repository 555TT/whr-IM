<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmNewPassword: ''
})

function validate() {
  if (!form.oldPassword) {
    return '请输入旧密码'
  }
  if (form.newPassword.length < 6 || form.newPassword.length > 20) {
    return '密码长度需在 6 到 20 位之间'
  }
  if (form.newPassword !== form.confirmNewPassword) {
    return '两次输入的新密码不一致'
  }
  if (form.oldPassword === form.newPassword) {
    return '新密码不能与旧密码相同'
  }
  return ''
}

async function changePassword() {
  errorMessage.value = ''
  message.value = ''
  const validationMessage = validate()
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }

  loading.value = true
  try {
    await http.put('/users/me/password', form)
    authStore.clearSession()
    await router.push({
      path: '/login',
      query: { message: '密码修改成功，请重新登录' }
    })
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="card apple-panel security-shell">
      <div class="security-header">
        <p class="apple-label">Security</p>
        <h1>账号安全</h1>
        <p class="muted">修改密码后将立即退出当前登录状态，请使用新密码重新登录。</p>
      </div>
      <p v-if="message" class="status-text success">{{ message }}</p>
      <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
      <label>
        <span class="apple-label">旧密码</span>
        <input v-model="form.oldPassword" class="apple-input" type="password" placeholder="请输入旧密码" />
      </label>
      <label>
        <span class="apple-label">新密码</span>
        <input v-model="form.newPassword" class="apple-input" type="password" placeholder="请输入新密码（6-20 位）" />
      </label>
      <label>
        <span class="apple-label">确认新密码</span>
        <input v-model="form.confirmNewPassword" class="apple-input" type="password" placeholder="请再次输入新密码" />
      </label>
      <div class="security-actions">
        <button class="apple-button" :disabled="loading" @click="changePassword">确认修改</button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.security-shell {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.security-header h1 {
  margin: 8px 0 10px;
  font-size: 38px;
  letter-spacing: -0.03em;
}

.security-header p {
  margin: 0;
}

label {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.security-actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .security-actions {
    justify-content: stretch;
  }
}
</style>
