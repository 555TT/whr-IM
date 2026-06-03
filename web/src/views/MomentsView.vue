<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'

interface MomentCommentItem {
  id: number
  userId: number
  nickname: string
  content: string
  createdAt: string
}

interface MomentItem {
  id: number
  userId: number
  nickname: string
  avatar: string
  content: string
  images: string[]
  likeCount: number
  likedByMe: boolean
  comments: MomentCommentItem[]
  createdAt: string
}

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
const moments = ref<MomentItem[]>([])
const content = ref('')
const uploadedImageKey = ref('')
const uploadedImageUrl = ref('')
const feedback = ref('')
const errorMessage = ref('')
const loading = ref(false)
const uploadingImage = ref(false)
const commentDrafts = ref<Record<number, string>>({})
const aiMode = ref<MomentAIAssistMode>('generate')
const aiTone = ref('自然')
const aiPrompt = ref('')
const aiLoading = ref(false)
const aiResult = ref('')

async function loadMoments() {
  try {
    const { data } = await http.get('/moments')
    moments.value = data
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

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
    content.value = ''
    uploadedImageKey.value = ''
    uploadedImageUrl.value = ''
    aiPrompt.value = ''
    aiResult.value = ''
    feedback.value = '动态已发布'
    await loadMoments()
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

async function toggleLike(item: MomentItem) {
  errorMessage.value = ''
  feedback.value = ''
  try {
    if (item.likedByMe) {
      await http.delete(`/moments/${item.id}/likes/me`)
    } else {
      await http.post(`/moments/${item.id}/likes`)
    }
    await loadMoments()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function submitComment(item: MomentItem) {
  const content = commentDrafts.value[item.id]?.trim()
  if (!content) return
  errorMessage.value = ''
  feedback.value = ''
  try {
    await http.post(`/moments/${item.id}/comments`, { content })
    commentDrafts.value[item.id] = ''
    await loadMoments()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function deleteMoment(item: MomentItem) {
  errorMessage.value = ''
  feedback.value = ''
  try {
    await http.delete(`/moments/${item.id}`)
    feedback.value = '动态已删除'
    await loadMoments()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

function openHomepage(userId: number) {
  router.push(`/users/${userId}`)
}

onMounted(loadMoments)
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="moments-layout">
      <div class="card apple-panel composer-card" :class="skin.surfaceClass">
        <p class="apple-label">Moments</p>
        <h1>朋友圈</h1>
        <p class="muted">可选择一张图片后再发布动态。</p>
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
        <button class="apple-button" :disabled="loading || uploadingImage" @click="publishMoment">发布动态</button>
      </div>

      <div class="feed-column">
        <div v-if="moments.length === 0" class="card apple-panel empty-state-card">暂无动态</div>
        <article v-for="item in moments" :key="item.id" class="card apple-panel moment-card" :class="skin.accentClass">
          <div class="moment-head">
            <img :src="item.avatar" alt="avatar" class="avatar clickable-avatar" @click="openHomepage(item.userId)" />
            <div>
              <strong class="clickable-name" @click="openHomepage(item.userId)">{{ item.nickname }}</strong>
              <p class="muted moment-time">{{ item.createdAt }}</p>
            </div>
          </div>
          <p class="moment-content">{{ item.content }}</p>
          <div v-if="item.images.length" class="moment-images">
            <img v-for="src in item.images" :key="src" :src="src" alt="moment image" class="moment-image" />
          </div>
          <div class="moment-actions">
            <button class="apple-button secondary" type="button" @click="toggleLike(item)">
              {{ item.likedByMe ? '取消点赞' : '点赞' }}
            </button>
            <span class="muted">{{ item.likeCount }} 人点赞</span>
            <button
              v-if="authStore.user?.id === item.userId"
              class="apple-button danger"
              type="button"
              @click="deleteMoment(item)"
            >
              删除动态
            </button>
          </div>
          <div class="comment-composer">
            <input
              v-model="commentDrafts[item.id]"
              class="apple-input"
              placeholder="写下评论..."
            />
            <button class="apple-button secondary" type="button" @click="submitComment(item)">评论</button>
          </div>
          <div v-if="item.comments.length" class="comment-list">
            <div v-for="comment in item.comments" :key="comment.id" class="comment-item">
              <strong>{{ comment.nickname }}</strong>
              <span class="comment-separator">：</span>
              <span>{{ comment.content }}</span>
            </div>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.moments-layout {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 24px;
}

.composer-card,
.moment-card,
.empty-state-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.composer-card {
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
  color: #fff;
}

.composer-card .apple-label,
.composer-card .muted {
  color: rgba(255, 255, 255, 0.86);
}

.moment-card {
  box-shadow: 0 12px 28px var(--skin-accent-soft, rgba(15, 23, 42, 0.08));
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

.ai-assistant {
  gap: 12px;
  background: rgba(255, 255, 255, 0.92);
  color: #1d1d1f;
  border: 1px solid var(--skin-accent-soft, rgba(0, 113, 227, 0.16));
}

.ai-assistant .apple-label {
  color: #6e6e73;
}

.ai-assistant .muted,
.ai-assistant strong,
.ai-assistant span,
.ai-assistant p,
.ai-assistant label {
  color: #1d1d1f;
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

.ai-hint {
  margin: 0;
}

.ai-result {
  padding: 14px 16px;
  border-radius: 16px;
  background: var(--skin-accent-soft, rgba(0, 113, 227, 0.08));
}

.ai-result p {
  margin: 0;
  white-space: pre-wrap;
}

.ai-mode-row .apple-button.active {
  border-color: transparent;
  background: var(--skin-accent, #0071e3);
  color: #fff;
}

.ai-actions .apple-button[disabled] {
  opacity: 0.6;
}

.upload-picker {
  position: relative;
  width: 132px;
  height: 132px;
  border-radius: 24px;
  border: 1px dashed rgba(0, 113, 227, 0.24);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(245, 247, 250, 0.98) 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.upload-picker:hover {
  transform: translateY(-1px);
  border-color: rgba(0, 113, 227, 0.45);
  box-shadow: 0 14px 28px rgba(0, 113, 227, 0.12);
}

.upload-picker.uploading {
  opacity: 0.72;
  pointer-events: none;
}

.upload-input {
  display: none;
}

.upload-plus {
  font-size: 46px;
  line-height: 1;
  color: #0071e3;
  font-weight: 300;
}

.upload-hint {
  font-size: 13px;
  color: #6e6e73;
}

.upload-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-overlay {
  position: absolute;
  inset: auto 0 0 0;
  padding: 10px 12px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(29, 29, 31, 0.68) 100%);
  color: #fff;
  font-size: 12px;
  text-align: center;
}

.upload-tools {
  display: flex;
  gap: 12px;
}

.feed-column {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.moment-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
}

.clickable-avatar {
  cursor: pointer;
}

.clickable-name {
  cursor: pointer;
}

.clickable-name:hover {
  color: #0071e3;
}

.moment-time {
  margin: 4px 0 0;
  font-size: 13px;
}

.moment-content {
  margin: 0;
  white-space: pre-wrap;
}

.moment-images {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
}

.moment-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.comment-composer {
  display: flex;
  gap: 12px;
  align-items: center;
}

.comment-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
  border-radius: 14px;
  background: rgba(29, 29, 31, 0.04);
}

.comment-item {
  font-size: 14px;
  line-height: 1.5;
}

.comment-separator {
  color: #6e6e73;
}

.moment-image {
  width: 100%;
  border-radius: 16px;
  object-fit: cover;
  background: rgba(0, 0, 0, 0.04);
}

@media (max-width: 920px) {
  .moments-layout {
    grid-template-columns: 1fr;
  }
}
</style>
