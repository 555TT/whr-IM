<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { homepageSkins, resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'
import { genderCodeToLabel, genderLabelToCode } from '../utils/gender'

const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const profile = reactive({
  nickname: '',
  gender: '女',
  signature: '',
  homepageSkin: 'aurora'
})

function syncProfile() {
  profile.nickname = authStore.user?.nickname || ''
  profile.gender = genderCodeToLabel(authStore.user?.gender ?? 0)
  profile.signature = authStore.user?.signature || ''
  profile.homepageSkin = authStore.user?.homepageSkin || 'aurora'
}

async function saveProfile() {
  loading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const { data } = await http.put('/users/me', {
      nickname: profile.nickname,
      gender: genderLabelToCode(profile.gender),
      signature: profile.signature,
      homepageSkin: profile.homepageSkin
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

function currentSkin() {
  return resolveHomepageSkin(profile.homepageSkin)
}

onMounted(syncProfile)
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="card apple-panel profile-shell" :class="currentSkin().surfaceClass">
      <div class="profile-header profile-hero">
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
      <div>
        <span class="apple-label">主页皮肤</span>
        <div class="skin-grid">
          <button
            v-for="skin in homepageSkins"
            :key="skin.key"
            type="button"
            class="skin-card"
            :class="[skin.previewClass, { active: profile.homepageSkin === skin.key }]"
            @click="profile.homepageSkin = skin.key"
          >
            <span class="skin-name">{{ skin.label }}</span>
          </button>
        </div>
      </div>
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
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
  color: #fff;
}

.profile-hero {
  color: #fff;
}

.profile-hero .muted,
.profile-hero .apple-label {
  color: rgba(255, 255, 255, 0.86);
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

.skin-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 14px;
}

.skin-card {
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  min-height: 92px;
  padding: 14px;
  color: #fff;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.2);
}

.skin-card.active {
  outline: 3px solid #fff;
  transform: translateY(-1px);
}

.skin-name {
  font-weight: 700;
  text-shadow: 0 2px 10px rgba(0, 0, 0, 0.22);
}

.skin-aurora {
  background: linear-gradient(135deg, #6dd5ed 0%, #2193b0 100%);
}

.skin-sunset {
  background: linear-gradient(135deg, #f97316 0%, #ec4899 100%);
}

.skin-galaxy {
  background: linear-gradient(135deg, #312e81 0%, #7c3aed 55%, #ec4899 100%);
}

.skin-mint {
  background: linear-gradient(135deg, #34d399 0%, #14b8a6 100%);
}

.skin-peach {
  background: linear-gradient(135deg, #fb7185 0%, #fdba74 100%);
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
