<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'

interface FavoriteItem {
  id: number
  sourceType: 'friend' | 'group' | 'ai'
  sourceMessageId: number
  conversationId: number
  senderId: number
  senderName: string
  content: string
  messageCreatedAt: string
  favoritedAt: string
}

const authStore = useAuthStore()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const favorites = ref<FavoriteItem[]>([])
const loading = ref(false)
const removingId = ref<number | null>(null)
const errorMessage = ref('')

function sourceLabel(type: FavoriteItem['sourceType']) {
  if (type === 'friend') return '单聊'
  if (type === 'group') return '群聊'
  return 'AI'
}

async function loadFavorites() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await http.get('/favorites')
    favorites.value = data
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

async function removeFavorite(id: number) {
  removingId.value = id
  errorMessage.value = ''
  try {
    await http.delete(`/favorites/${id}`)
    favorites.value = favorites.value.filter((item) => item.id !== id)
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    removingId.value = null
  }
}

onMounted(loadFavorites)
</script>

<template>
  <div class="page-shell apple-page favorites-page" :class="skin.surfaceClass">
    <AppNav />
    <section class="favorites-shell">
      <div class="card apple-panel favorites-hero">
        <p class="apple-label">My Favorites</p>
        <h1>我的收藏</h1>
        <p class="muted">统一查看单聊、群聊和 AI 聊天中收藏的消息。</p>
      </div>

      <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
      <div v-if="loading" class="card apple-panel empty-state-card">收藏列表加载中...</div>
      <div v-else-if="favorites.length === 0" class="card apple-panel empty-state-card">还没有收藏消息，去消息中心试试右键收藏吧。</div>
      <div v-else class="favorite-list">
        <article v-for="item in favorites" :key="item.id" class="card apple-panel favorite-card">
          <div class="favorite-head">
            <div>
              <span class="favorite-tag">{{ sourceLabel(item.sourceType) }}</span>
              <strong>{{ item.senderName }}</strong>
            </div>
            <button class="apple-button danger" :disabled="removingId === item.id" @click="removeFavorite(item.id)">
              {{ removingId === item.id ? '取消中...' : '取消收藏' }}
            </button>
          </div>
          <p class="favorite-content">{{ item.content }}</p>
          <div class="favorite-meta muted">
            <span>原消息时间：{{ item.messageCreatedAt }}</span>
            <span>收藏时间：{{ item.favoritedAt }}</span>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.favorites-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.favorites-hero,
.favorite-card,
.empty-state-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.favorites-hero {
  color: #fff;
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
}

.favorites-hero .apple-label,
.favorites-hero .muted {
  color: rgba(255, 255, 255, 0.86);
}

.favorite-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.favorite-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.favorite-tag {
  display: inline-block;
  margin-right: 10px;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(0, 113, 227, 0.12);
  color: #0071e3;
  font-size: 12px;
  font-weight: 700;
}

.favorite-content {
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.7;
  color: #1d1d1f;
}

.favorite-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 13px;
}
</style>
