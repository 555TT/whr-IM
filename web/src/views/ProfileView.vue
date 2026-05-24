<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { genderCodeToLabel, genderLabelToCode } from '../utils/gender'

const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const profile = reactive({
  nickname: '',
  gender: '女',
  signature: ''
})

function syncProfile() {
  profile.nickname = authStore.user?.nickname || ''
  profile.gender = genderCodeToLabel(authStore.user?.gender ?? 0)
  profile.signature = authStore.user?.signature || ''
}

async function saveProfile() {
  loading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const { data } = await http.put('/users/me', {
      nickname: profile.nickname,
      gender: genderLabelToCode(profile.gender),
      signature: profile.signature
    })
    authStore.user = data
    syncProfile()
    message.value = '资料已更新'
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(syncProfile)
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="card apple-panel profile-shell">
      <div class="profile-header">
        <div>
          <p class="apple-label">Profile</p>
          <h1>个人资料</h1>
          <p class="muted">头像为系统默认头像，不可修改。你可以调整昵称、性别和个性签名。</p>
        </div>
      </div>
      <p v-if="message" class="status-text success">{{ message }}</p>
      <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
      <div class="profile-grid">
        <label>
          <span class="apple-label">昵称</span>
          <input v-model="profile.nickname" class="apple-input" placeholder="昵称" />
        </label>
        <label>
          <span class="apple-label">性别</span>
          <select v-model="profile.gender" class="apple-input">
            <option value="女">女</option>
            <option value="男">男</option>
          </select>
        </label>
      </div>
      <label>
        <span class="apple-label">个性签名</span>
        <textarea v-model="profile.signature" class="apple-textarea" placeholder="写一句介绍自己的话" />
      </label>
      <div class="profile-actions">
        <button class="apple-button" :disabled="loading" @click="saveProfile">保存更改</button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.profile-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.profile-header h1 {
  margin: 8px 0 10px;
  font-size: 38px;
  letter-spacing: -0.03em;
}

.profile-header p {
  margin: 0;
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.profile-actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }

  .profile-actions {
    justify-content: stretch;
  }
}
</style>
