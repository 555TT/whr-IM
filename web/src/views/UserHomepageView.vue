<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'

interface PublicProfile {
  id: number
  nickname: string
  avatar: string
  signature: string
}

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

const route = useRoute()
const authStore = useAuthStore()
const profile = ref<PublicProfile | null>(null)
const moments = ref<MomentItem[]>([])
const loading = ref(false)
const errorMessage = ref('')
const commentDrafts = ref<Record<number, string>>({})

function currentUserId() {
  return Number(route.params.id)
}

async function loadHomepage() {
  loading.value = true
  errorMessage.value = ''
  commentDrafts.value = {}
  try {
    const userId = currentUserId()
    const [{ data: profileData }, { data: momentsData }] = await Promise.all([
      http.get(`/users/${userId}/profile`),
      http.get(`/users/${userId}/moments`)
    ])
    profile.value = profileData
    moments.value = momentsData
  } catch (error) {
    profile.value = null
    moments.value = []
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

async function toggleLike(item: MomentItem) {
  errorMessage.value = ''
  try {
    if (item.likedByMe) {
      await http.delete(`/moments/${item.id}/likes/me`)
    } else {
      await http.post(`/moments/${item.id}/likes`)
    }
    await loadHomepage()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function submitComment(item: MomentItem) {
  const content = commentDrafts.value[item.id]?.trim()
  if (!content) return
  errorMessage.value = ''
  try {
    await http.post(`/moments/${item.id}/comments`, { content })
    commentDrafts.value[item.id] = ''
    await loadHomepage()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function deleteMoment(item: MomentItem) {
  errorMessage.value = ''
  try {
    await http.delete(`/moments/${item.id}`)
    await loadHomepage()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

watch(() => route.params.id, loadHomepage)
onMounted(loadHomepage)
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="homepage-layout">
      <div class="card apple-panel homepage-profile" v-if="profile">
        <div class="profile-top">
          <img :src="profile.avatar" alt="avatar" class="homepage-avatar" />
          <div>
            <p class="apple-label">Homepage</p>
            <h1>{{ profile.nickname }}</h1>
            <p class="muted">{{ profile.signature || '这个人很低调，还没有留下签名。' }}</p>
          </div>
        </div>
      </div>

      <div class="feed-column">
        <div v-if="loading" class="card apple-panel empty-state-card">主页加载中...</div>
        <div v-else-if="errorMessage" class="card apple-panel empty-state-card error-state">{{ errorMessage }}</div>
        <div v-else-if="moments.length === 0" class="card apple-panel empty-state-card">
          {{ authStore.user?.id === profile?.id ? '你还没有发布动态' : '对方还没有发布动态' }}
        </div>
        <article v-for="item in moments" :key="item.id" class="card apple-panel moment-card">
          <div class="moment-head">
            <img :src="item.avatar" alt="avatar" class="avatar" />
            <div>
              <strong>{{ item.nickname }}</strong>
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
.homepage-layout {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.homepage-profile,
.moment-card,
.empty-state-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.profile-top {
  display: flex;
  align-items: center;
  gap: 20px;
}

.homepage-avatar {
  width: 96px;
  height: 96px;
  border-radius: 28px;
  object-fit: cover;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.14);
}

.profile-top h1 {
  margin: 6px 0 8px;
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
  gap: 14px;
}

.avatar {
  width: 52px;
  height: 52px;
  border-radius: 18px;
  object-fit: cover;
}

.moment-time {
  margin: 4px 0 0;
}

.moment-content {
  margin: 0;
  color: #1d1d1f;
  line-height: 1.7;
}

.moment-images {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.moment-image {
  width: min(100%, 240px);
  border-radius: 22px;
  object-fit: cover;
  box-shadow: 0 14px 26px rgba(15, 23, 42, 0.12);
}

.moment-actions,
.comment-composer {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.comment-composer .apple-input {
  flex: 1;
  min-width: 220px;
}

.comment-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.comment-item {
  padding: 10px 14px;
  border-radius: 16px;
  background: rgba(0, 113, 227, 0.06);
  color: #1d1d1f;
}

.comment-separator {
  color: #6e6e73;
}

.error-state {
  color: #d93025;
}

@media (max-width: 760px) {
  .profile-top {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
