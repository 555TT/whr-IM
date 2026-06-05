<script setup lang="ts">
import { computed, reactive, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { homepageSkins, resolveHomepageSkin } from '../constants/homepageSkins'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { genderLabelToCode } from '../utils/gender'

const avatarAccessories = [
  { key: 'none', label: '无挂件' },
  { key: 'star-ring', label: '星环' },
  { key: 'flower-crown', label: '花环' },
  { key: 'spark-frame', label: '光点框' }
]

const titleBadges = [
  { key: 'none', label: '无称号' },
  { key: 'day-dreamer', label: '摸鱼达人' },
  { key: 'night-chatter', label: '夜聊星人' },
  { key: 'always-online', label: '今日在线' }
]

const homepageBackgrounds = [
  { key: 'plain', label: '简约白' },
  { key: 'soft-glow', label: '柔光渐变' },
  { key: 'starry', label: '星夜' },
  { key: 'mint-fog', label: '薄荷雾面' }
]

const homepageLayouts = [
  { key: 'classic', label: '经典卡片' },
  { key: 'poster', label: '居中海报' },
  { key: 'split', label: '信息分栏' }
]

const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const decoration = reactive({
  homepageSkin: authStore.user?.homepageSkin || 'aurora',
  avatarAccessory: authStore.user?.avatarAccessory || 'none',
  titleBadge: authStore.user?.titleBadge || 'none',
  homepageBackground: authStore.user?.homepageBackground || 'plain',
  homepageLayout: authStore.user?.homepageLayout || 'classic'
})

const previewSkin = computed(() => resolveHomepageSkin(decoration.homepageSkin))
const currentBadge = computed(() => titleBadges.find((item) => item.key === decoration.titleBadge)?.label || '无称号')

async function saveDecorations() {
  if (!authStore.user) return
  loading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const { data } = await http.put('/users/me', {
      nickname: authStore.user.nickname,
      gender: genderLabelToCode(authStore.user.gender === 1 ? '男' : '女'),
      signature: authStore.user.signature || '',
      homepageSkin: decoration.homepageSkin,
      avatarAccessory: decoration.avatarAccessory,
      titleBadge: decoration.titleBadge,
      homepageBackground: decoration.homepageBackground,
      homepageLayout: decoration.homepageLayout
    })
    authStore.user = data
    message.value = '个性装扮已保存'
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
    <section class="decorations-layout">
      <div class="card apple-panel preview-card" :class="previewSkin.surfaceClass">
        <p class="apple-label">Preview</p>
        <h1>个性装扮</h1>
        <div class="preview-profile" :class="decoration.homepageLayout">
          <div class="preview-avatar-wrap" :class="decoration.avatarAccessory">
            <img v-if="authStore.user?.avatar" :src="authStore.user.avatar" alt="avatar" class="preview-avatar-image" />
            <div v-else class="preview-avatar">头像</div>
          </div>
          <div>
            <div class="preview-title-row">
              <strong>{{ authStore.user?.nickname || authStore.user?.username }}</strong>
              <span class="title-badge">{{ currentBadge }}</span>
            </div>
            <p class="muted">{{ authStore.user?.signature || '这里展示你的个性签名。' }}</p>
            <small>背景：{{ homepageBackgrounds.find((item) => item.key === decoration.homepageBackground)?.label }}</small>
            <small>布局：{{ homepageLayouts.find((item) => item.key === decoration.homepageLayout)?.label }}</small>
          </div>
        </div>
      </div>

      <p v-if="message" class="status-text success">{{ message }}</p>
      <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>

      <div class="card apple-panel section-card">
        <p class="apple-label">主页皮肤</p>
        <div class="option-grid skin-grid">
          <button v-for="skin in homepageSkins" :key="skin.key" type="button" class="skin-card" :class="[skin.previewClass, { active: decoration.homepageSkin === skin.key }]" @click="decoration.homepageSkin = skin.key">
            <span class="skin-name">{{ skin.label }}</span>
          </button>
        </div>
      </div>

      <div class="card apple-panel section-card">
        <p class="apple-label">头像挂件</p>
        <div class="option-grid">
          <button v-for="item in avatarAccessories" :key="item.key" type="button" class="option-card" :class="{ active: decoration.avatarAccessory === item.key }" @click="decoration.avatarAccessory = item.key">{{ item.label }}</button>
        </div>
      </div>

      <div class="card apple-panel section-card">
        <p class="apple-label">称号</p>
        <div class="option-grid">
          <button v-for="item in titleBadges" :key="item.key" type="button" class="option-card" :class="{ active: decoration.titleBadge === item.key }" @click="decoration.titleBadge = item.key">{{ item.label }}</button>
        </div>
      </div>

      <div class="card apple-panel section-card">
        <p class="apple-label">主页背景预设</p>
        <div class="option-grid">
          <button v-for="item in homepageBackgrounds" :key="item.key" type="button" class="option-card" :class="{ active: decoration.homepageBackground === item.key }" @click="decoration.homepageBackground = item.key">{{ item.label }}</button>
        </div>
      </div>

      <div class="card apple-panel section-card">
        <p class="apple-label">主页布局预设</p>
        <div class="option-grid">
          <button v-for="item in homepageLayouts" :key="item.key" type="button" class="option-card" :class="{ active: decoration.homepageLayout === item.key }" @click="decoration.homepageLayout = item.key">{{ item.label }}</button>
        </div>
      </div>

      <div class="actions">
        <button class="apple-button" :disabled="loading" @click="saveDecorations">{{ loading ? '保存中...' : '保存装扮' }}</button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.decorations-layout {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.preview-card,
.section-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.preview-card {
  color: #fff;
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
}

.preview-profile {
  display: flex;
  gap: 18px;
  align-items: center;
}

.preview-profile.poster {
  flex-direction: column;
  align-items: flex-start;
}

.preview-profile.split {
  justify-content: space-between;
}

.preview-avatar-wrap,
.preview-avatar,
.preview-avatar-image {
  width: 96px;
  height: 96px;
  border-radius: 28px;
}

.preview-avatar-wrap {
  position: relative;
  display: inline-flex;
}

.preview-avatar {
  background: rgba(255, 255, 255, 0.22);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.preview-avatar-image {
  object-fit: cover;
  display: block;
}

.preview-avatar-wrap.star-ring {
  box-shadow: 0 0 0 4px rgba(255, 215, 0, 0.5);
}

.preview-avatar-wrap.flower-crown {
  box-shadow: 0 -6px 0 0 rgba(244, 114, 182, 0.45);
}

.preview-avatar-wrap.spark-frame {
  box-shadow: 0 0 0 4px rgba(255, 255, 255, 0.38), 0 0 18px rgba(255, 255, 255, 0.3);
}

.preview-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.title-badge {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.18);
  font-size: 12px;
  font-weight: 700;
}

.option-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
}

.skin-grid {
  margin-top: 4px;
}

.skin-card,
.option-card {
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 18px;
  min-height: 76px;
  padding: 14px;
  cursor: pointer;
  text-align: left;
  background: #fff;
}

.skin-card.active,
.option-card.active {
  outline: 2px solid #0071e3;
  transform: translateY(-1px);
}

.skin-name {
  font-weight: 700;
  color: #fff;
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

.actions {
  display: flex;
  justify-content: flex-end;
}
</style>
