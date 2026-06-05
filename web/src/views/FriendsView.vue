<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'

interface FriendItem {
  userId: number
  friendId: number
  nickname: string
  avatar: string
  signature: string
}

const authStore = useAuthStore()
const router = useRouter()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const friends = ref<FriendItem[]>([])
const loading = ref(false)
const errorMessage = ref('')
const feedback = ref(typeof history.state?.friendDeletedMessage === 'string' ? history.state.friendDeletedMessage : '')

async function loadFriends() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await http.get('/friends')
    friends.value = data
  } catch (error) {
    friends.value = []
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

function openHomepage(friendId: number) {
  router.push(`/users/${friendId}`)
}

function initialOf(name: string) {
  return name.slice(0, 1).toUpperCase()
}

onMounted(async () => {
  await loadFriends()
  if (feedback.value) {
    history.replaceState({}, document.title)
  }
})
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="friends-layout">
      <div class="card apple-panel friends-hero" :class="skin.surfaceClass">
        <p class="apple-label">Friends</p>
        <h1>我的好友</h1>
        <p class="muted">查看所有已添加好友，点击即可进入对方主页。</p>
        <p v-if="feedback" class="status-text success">{{ feedback }}</p>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
      </div>

      <div v-if="loading" class="card apple-panel empty-state-card">好友列表加载中...</div>
      <div v-else-if="friends.length === 0" class="card apple-panel empty-state-card">你还没有添加好友，先去通讯录申请页添加吧。</div>
      <div v-else class="friends-grid">
        <button
          v-for="friend in friends"
          :key="friend.friendId"
          class="card apple-panel friend-card"
          type="button"
          @click="openHomepage(friend.friendId)"
        >
          <div class="friend-avatar">{{ initialOf(friend.nickname) }}</div>
          <div class="friend-copy">
            <strong>{{ friend.nickname }}</strong>
            <p>{{ friend.signature || '这个人很低调，还没有留下签名。' }}</p>
          </div>
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.friends-layout {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.friends-hero,
.empty-state-card,
.friend-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.friends-hero {
  color: #fff;
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
}

.friends-hero .apple-label,
.friends-hero .muted {
  color: rgba(255, 255, 255, 0.86);
}

.friends-hero h1 {
  margin: 6px 0 0;
  font-size: 34px;
  letter-spacing: -0.03em;
}

.friends-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 18px;
}

.friend-card {
  border: none;
  text-align: left;
  cursor: pointer;
  background: #fff;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.friend-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.12);
}

.friend-avatar {
  width: 62px;
  height: 62px;
  border-radius: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 113, 227, 0.12);
  color: #0071e3;
  font-size: 26px;
  font-weight: 700;
}

.friend-copy {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.friend-copy strong {
  font-size: 20px;
  color: #1d1d1f;
}

.friend-copy p {
  margin: 0;
  color: #6e6e73;
  line-height: 1.6;
}
</style>
