<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const profile = reactive({
  nickname: '',
  gender: 0,
  signature: ''
})

function syncProfile() {
  profile.nickname = authStore.user?.nickname || ''
  profile.gender = authStore.user?.gender || 0
  profile.signature = authStore.user?.signature || ''
}

async function saveProfile() {
  loading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const { data } = await http.put('/users/me', profile)
    authStore.user = data
    syncProfile()
    message.value = '保存成功'
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(syncProfile)
</script>

<template>
  <div class="page-shell page-layout">
    <AppNav />
    <div class="card content-card">
      <h2>个人资料</h2>
      <p>头像为系统默认头像，不可修改。</p>
      <p v-if="message" class="success-text">{{ message }}</p>
      <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      <input v-model="profile.nickname" placeholder="昵称" />
      <input v-model.number="profile.gender" type="number" placeholder="性别：0/1/2" />
      <textarea v-model="profile.signature" placeholder="个性签名" />
      <button :disabled="loading" @click="saveProfile">保存</button>
    </div>
  </div>
</template>

<style scoped>
.page-layout {
  padding: 24px;
}

.content-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

input,
textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
}

textarea {
  min-height: 120px;
}

button {
  width: 160px;
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
