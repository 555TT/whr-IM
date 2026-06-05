<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import AvatarCropper from '../components/AvatarCropper.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'
import { genderCodeToLabel, genderLabelToCode } from '../utils/gender'

interface UploadResponse {
  objectKey: string
  url: string
}

const maxAvatarSize = 2 * 1024 * 1024

const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const avatarPreviewUrl = ref('')
const avatarFile = ref<File | null>(null)
const cropperVisible = ref(false)
const cropperImageUrl = ref('')
const profile = reactive({
  nickname: '',
  gender: '女',
  signature: ''
})

const displayAvatar = computed(() => avatarPreviewUrl.value || authStore.user?.avatar || '')

function syncProfile() {
  profile.nickname = authStore.user?.nickname || ''
  profile.gender = genderCodeToLabel(authStore.user?.gender ?? 0)
  profile.signature = authStore.user?.signature || ''
  avatarPreviewUrl.value = ''
  avatarFile.value = null
}

function pickAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  errorMessage.value = ''
  message.value = ''
  if (!file.type.startsWith('image/')) {
    errorMessage.value = '请选择图片文件'
    input.value = ''
    return
  }
  if (file.size > maxAvatarSize) {
    errorMessage.value = '头像图片不能超过 2MB'
    input.value = ''
    return
  }
  cropperImageUrl.value = URL.createObjectURL(file)
  cropperVisible.value = true
  input.value = ''
}

function applyCroppedAvatar(payload: { file: File; previewUrl: string }) {
  avatarFile.value = payload.file
  avatarPreviewUrl.value = payload.previewUrl
  cropperVisible.value = false
  cropperImageUrl.value = ''
}

async function saveProfile() {
  loading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    let avatar = authStore.user?.avatar || ''
    if (avatarFile.value) {
      const formData = new FormData()
      formData.append('file', avatarFile.value)
      const { data: uploadData } = await http.post<UploadResponse>('/uploads/images', formData)
      avatar = uploadData.url
    }
    const { data } = await http.put('/users/me', {
      nickname: profile.nickname,
      gender: genderLabelToCode(profile.gender),
      signature: profile.signature,
      avatar,
      homepageSkin: authStore.user?.homepageSkin || 'aurora',
      avatarAccessory: authStore.user?.avatarAccessory || 'none',
      titleBadge: authStore.user?.titleBadge || 'none',
      homepageBackground: authStore.user?.homepageBackground || 'plain',
      homepageLayout: authStore.user?.homepageLayout || 'classic'
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
  return resolveHomepageSkin(authStore.user?.homepageSkin)
}

onMounted(syncProfile)
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="card apple-panel profile-shell" :class="currentSkin().surfaceClass">
      <div class="profile-header profile-hero">
        <div class="profile-header-main">
          <label class="avatar-picker" :class="currentSkin().accentClass">
            <input class="avatar-input" type="file" accept="image/*" @change="pickAvatar" />
            <img v-if="displayAvatar" :src="displayAvatar" alt="avatar" class="profile-avatar" />
            <div class="avatar-overlay">点击更换头像</div>
          </label>
          <div>
            <p class="apple-label">Profile</p>
            <h1>个人资料</h1>
            <p class="muted">这里只保留基础资料编辑：昵称、性别、个性签名和头像。个性装扮请前往独立装扮中心。</p>
          </div>
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
    <AvatarCropper :visible="cropperVisible" :image-url="cropperImageUrl" @close="cropperVisible = false" @confirm="applyCroppedAvatar" />
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

.profile-header-main {
  display: flex;
  align-items: center;
  gap: 20px;
}

.avatar-picker {
  position: relative;
  width: 120px;
  height: 120px;
  border-radius: 32px;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.18);
}

.avatar-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  z-index: 2;
}

.profile-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.avatar-overlay {
  position: absolute;
  inset: auto 0 0;
  padding: 12px 10px;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0), rgba(15, 23, 42, 0.72));
  color: #fff;
  font-size: 12px;
  text-align: center;
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
  .profile-header-main {
    flex-direction: column;
    align-items: flex-start;
  }

  .profile-grid {
    grid-template-columns: 1fr;
  }

  .profile-actions {
    justify-content: stretch;
  }
}
</style>
