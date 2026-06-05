<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'

interface UploadResponse {
  objectKey: string
  url: string
}

interface MomentAIAssistResponse {
  text: string
}

type MomentAIAssistMode = 'generate' | 'polish'

const authStore = useAuthStore()
const router = useRouter()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const content = ref('')
const uploadedImageKey = ref('')
const uploadedImageUrl = ref('')
const feedback = ref('')
const errorMessage = ref('')
const loading = ref(false)
const uploadingImage = ref(false)
const aiMode = ref<MomentAIAssistMode>('generate')
const aiTone = ref('自然')
const aiPrompt = ref('')
const aiLoading = ref(false)
const aiResult = ref('')

async function uploadImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  feedback.value = ''
  errorMessage.value = ''
  uploadingImage.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const { data } = await http.post<UploadResponse>('/uploads/images', formData)
    uploadedImageKey.value = data.objectKey
    uploadedImageUrl.value = data.url
    feedback.value = '图片已上传'
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    uploadingImage.value = false
    input.value = ''
  }
}

async function requestAIAssist() {
  const prompt = aiPrompt.value.trim()
  const currentContent = content.value.trim()
  if (aiMode.value === 'generate' && !prompt) {
    errorMessage.value = '请输入想法后再生成文案'
    feedback.value = ''
    return
  }
  if (aiMode.value === 'polish' && !currentContent) {
    errorMessage.value = '请先输入正文后再进行润色'
    feedback.value = ''
    return
  }

  const payload = {
    mode: aiMode.value,
    prompt: aiMode.value === 'generate' ? prompt : '',
    content: aiMode.value === 'polish' ? currentContent : '',
    tone: aiTone.value,
    hasImage: Boolean(uploadedImageKey.value)
  }
  aiLoading.value = true
  aiResult.value = ''
  feedback.value = ''
  errorMessage.value = ''
  try {
    const { data } = await http.post<MomentAIAssistResponse>('/moments/ai-assist', payload)
    aiResult.value = data.text
    feedback.value = 'AI 文案已生成'
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    aiLoading.value = false
  }
}

function applyAIResult() {
  if (!aiResult.value) return
  content.value = aiResult.value
  feedback.value = '已填入正文'
}

async function publishMoment() {
  if (!content.value.trim()) return
  feedback.value = ''
  errorMessage.value = ''
  loading.value = true
  try {
    await http.post('/moments', {
      content: content.value.trim(),
      imageKeys: uploadedImageKey.value ? [uploadedImageKey.value] : []
    })
    router.push({ path: '/moments', state: { momentPublishedMessage: '动态已发布' } })
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

function backToMoments() {
  router.push('/moments')
}
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="create-moment-layout">
      <div class="card apple-panel composer-card" :class="skin.surfaceClass">
        <div class="page-head">
          <div>
            <p class="apple-label">Create Moment</p>
            <h1>发布动态</h1>
            <p class="muted">保留完整发布能力：文本、图片、AI 配文与润色。</p>
          </div>
          <button class="apple-button secondary" type="button" @click="backToMoments">返回朋友圈</button>
        </div>
        <p v-if="feedback" class="status-text success">{{ feedback }}</p>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <textarea v-model="content" class="apple-textarea" placeholder="分享这一刻..." />
        <div class="ai-assistant card apple-panel" :class="skin.accentClass">
          <div class="ai-head">
            <div>
              <p class="apple-label">AI Moments</p>
              <strong>朋友圈助手</strong>
            </div>
            <span class="muted">只生成建议，不会自动发布</span>
          </div>
          <div class="ai-mode-row">
            <button class="apple-button secondary" type="button" :class="{ active: aiMode === 'generate' }" @click="aiMode = 'generate'">智能配文</button>
            <button class="apple-button secondary" type="button" :class="{ active: aiMode === 'polish' }" @click="aiMode = 'polish'">文案润色</button>
          </div>
          <label v-if="aiMode === 'generate'" class="ai-field">
            <span class="apple-label">这一刻想表达什么</span>
            <input v-model="aiPrompt" class="apple-input" placeholder="比如：周末和朋友露营，看日落很治愈" />
          </label>
          <p v-else class="muted ai-hint">将基于当前正文内容进行润色。</p>
          <label class="ai-field">
            <span class="apple-label">语气风格</span>
            <select v-model="aiTone" class="apple-input">
              <option value="自然">自然</option>
              <option value="幽默">幽默</option>
              <option value="文艺">文艺</option>
              <option value="简洁">简洁</option>
            </select>
          </label>
          <div class="ai-actions">
            <button class="apple-button secondary" type="button" :disabled="aiLoading" @click="requestAIAssist">{{ aiLoading ? '生成中...' : aiMode === 'polish' ? '开始润色' : '生成文案' }}</button>
            <button class="apple-button" type="button" :disabled="!aiResult" @click="applyAIResult">填入正文</button>
          </div>
          <div v-if="aiResult" class="ai-result">
            <p class="apple-label">AI 建议</p>
            <p>{{ aiResult }}</p>
          </div>
        </div>
        <div class="upload-field">
          <span class="apple-label">配图（可选）</span>
          <label class="upload-picker" :class="{ uploading: uploadingImage }">
            <input class="upload-input" type="file" accept="image/*" @change="uploadImage" />
            <template v-if="uploadedImageUrl">
              <img :src="uploadedImageUrl" alt="uploaded preview" class="upload-cover" />
              <div class="upload-overlay">更换图片</div>
            </template>
            <template v-else>
              <span class="upload-plus">+</span>
              <span class="upload-hint">上传图片</span>
            </template>
          </label>
        </div>
        <div v-if="uploadingImage" class="muted">图片上传中...</div>
        <div v-if="uploadedImageUrl" class="upload-tools">
          <button class="apple-button secondary" type="button" @click="uploadedImageKey = ''; uploadedImageUrl = ''">移除图片</button>
        </div>
        <button class="apple-button" :disabled="loading || uploadingImage" @click="publishMoment">{{ loading ? '发布中...' : '发布动态' }}</button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.create-moment-layout {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.composer-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
  color: #fff;
}

.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.composer-card .apple-label,
.composer-card .muted {
  color: rgba(255, 255, 255, 0.86);
}

.composer-card h1 {
  margin: 6px 0 0;
  font-size: 34px;
  letter-spacing: -0.03em;
}

.upload-field {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.upload-picker {
  position: relative;
  width: min(100%, 420px);
  min-height: 220px;
  border-radius: 28px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.16);
  border: 1px dashed rgba(255, 255, 255, 0.28);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
}

.upload-picker.uploading {
  opacity: 0.72;
}

.upload-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.upload-plus {
  font-size: 52px;
  line-height: 1;
}

.upload-hint {
  font-size: 14px;
}

.upload-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-overlay {
  position: absolute;
  inset: auto 0 0;
  padding: 16px;
  text-align: center;
  background: linear-gradient(180deg, transparent, rgba(15, 23, 42, 0.68));
}

.upload-tools {
  display: flex;
  gap: 10px;
}

.ai-assistant {
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 255, 0.96));
  color: #1d1d1f;
  border: 1px solid rgba(15, 23, 42, 0.08);
  box-shadow: 0 16px 28px rgba(15, 23, 42, 0.08);
}

.ai-assistant .apple-label {
  color: #4b5563;
}

.ai-assistant strong,
.ai-assistant span,
.ai-assistant p,
.ai-assistant label,
.ai-assistant .muted {
  color: #1f2937;
}

.ai-head,
.ai-actions,
.ai-mode-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.ai-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-field .apple-input,
.ai-field select {
  background: #fff;
  color: #1d1d1f;
  border: 1px solid rgba(15, 23, 42, 0.12);
}

.ai-hint {
  margin: 0;
  color: #4b5563;
}

.ai-result {
  padding: 16px 18px;
  border-radius: 22px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  border: 1px solid rgba(15, 23, 42, 0.08);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.45);
}

.ai-result p {
  margin: 6px 0 0;
  white-space: pre-wrap;
  color: #111827;
  line-height: 1.7;
}

.ai-mode-row .apple-button.active {
  border-color: transparent;
  background: var(--skin-accent, #0071e3);
  color: #fff;
}

.ai-actions .apple-button.secondary,
.ai-mode-row .apple-button.secondary {
  background: rgba(15, 23, 42, 0.06);
  color: #1f2937;
}

.ai-actions .apple-button[disabled] {
  opacity: 0.6;
}

@media (max-width: 760px) {
  .page-head {
    flex-direction: column;
  }
}
</style>
