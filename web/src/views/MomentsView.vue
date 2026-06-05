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

const authStore = useAuthStore()
const router = useRouter()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const moments = ref<MomentItem[]>([])
const feedback = ref(typeof history.state?.momentPublishedMessage === 'string' ? history.state.momentPublishedMessage : '')
const errorMessage = ref('')
const commentDrafts = ref<Record<number, string>>({})

async function loadMoments() {
  try {
    const { data } = await http.get('/moments')
    moments.value = data
  } catch (error) {
    errorMessage.value = (error as Error).message
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

function openCreateMoment() {
  router.push('/moments/create')
}

onMounted(async () => {
  await loadMoments()
  if (feedback.value) {
    history.replaceState({}, document.title)
  }
})
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="moments-layout">
      <div class="card apple-panel moments-hero" :class="skin.surfaceClass">
        <div class="moments-head">
          <div>
            <p class="apple-label">Moments</p>
            <h1>朋友圈</h1>
            <p class="muted">浏览好友动态，点赞、评论，或删除自己发布的内容。</p>
          </div>
          <button class="apple-button secondary" type="button" @click="openCreateMoment">发布动态</button>
        </div>
        <p v-if="feedback" class="status-text success">{{ feedback }}</p>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
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
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.moments-hero,
.moment-card,
.empty-state-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.moments-hero {
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
  color: #fff;
}

.moments-hero .apple-label,
.moments-hero .muted {
  color: rgba(255, 255, 255, 0.86);
}

.moment-card {
  box-shadow: 0 12px 28px var(--skin-accent-soft, rgba(15, 23, 42, 0.08));
}

.moments-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.moments-hero h1 {
  margin: 6px 0 0;
  font-size: 34px;
  letter-spacing: -0.03em;
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
  .moments-head {
    flex-direction: column;
  }
}
</style>
