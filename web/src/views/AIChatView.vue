<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'

interface AIChatMessage {
  id: number
  userId: number
  role: 'user' | 'assistant'
  content: string
  createdAt: string
}

const authStore = useAuthStore()
const router = useRouter()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const messages = ref<AIChatMessage[]>([])
const draft = ref('')
const loading = ref(false)
const sending = ref(false)
const selectionMode = ref(false)
const selectedMessageIds = ref<number[]>([])
const errorMessage = ref('')

async function loadMessages() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await http.get('/ai-chat/messages')
    messages.value = data
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}

function openFavorites() {
  router.push('/favorites')
}

function isSelected(message: AIChatMessage) {
  return selectedMessageIds.value.includes(message.id)
}

function enterSelectionMode(message: AIChatMessage) {
  selectionMode.value = true
  toggleSelection(message)
}

function toggleSelection(message: AIChatMessage) {
  if (selectedMessageIds.value.includes(message.id)) {
    selectedMessageIds.value = selectedMessageIds.value.filter((id) => id !== message.id)
    if (selectedMessageIds.value.length === 0) {
      selectionMode.value = false
    }
    return
  }
  selectedMessageIds.value = [...selectedMessageIds.value, message.id]
}

async function favoriteSelectedMessages() {
  if (selectedMessageIds.value.length === 0) return
  errorMessage.value = ''
  try {
    await http.post('/favorites', {
      items: selectedMessageIds.value.map((id) => {
        const message = messages.value.find((item) => item.id === id)
        return { sourceType: 'ai', sourceMessageId: id, content: message?.content || '' }
      })
    })
    selectionMode.value = false
    selectedMessageIds.value = []
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

function clearSelection() {
  selectionMode.value = false
  selectedMessageIds.value = []
}

async function sendMessage() {
  if (!draft.value.trim() || sending.value) return
  sending.value = true
  errorMessage.value = ''
  try {
    const { data } = await http.post('/ai-chat/messages', { content: draft.value.trim() })
    messages.value.push(...(data as AIChatMessage[]))
    draft.value = ''
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    sending.value = false
  }
}

onMounted(loadMessages)
</script>

<template>
  <div class="page-shell apple-page chat-theme-shell" :class="skin.surfaceClass">
    <AppNav />
    <section class="chat-shell card ai-chat-shell">
      <aside class="sidebar ai-sidebar">
        <div class="sidebar-banner ai-banner">
          <div>
            <p class="apple-label">AI Assistant</p>
            <h2>AI 聊天</h2>
            <p class="sidebar-banner-copy">和智能助手进行独立对话，历史消息会自动保存。</p>
          </div>
        </div>
        <button class="favorite-entry-card" type="button" @click="openFavorites">
          <div>
            <p class="apple-label">My Favorites</p>
            <strong>我的收藏</strong>
            <span>统一查看已收藏消息</span>
          </div>
          <span class="favorite-entry-arrow">★</span>
        </button>
      </aside>

      <section class="chat-panel">
        <div class="chat-top">
          <div class="chat-top-main">
            <p class="apple-label">AI Conversation</p>
            <h2>智能助手</h2>
            <small class="muted">可直接提问、续聊，消息会保存在当前账号下。</small>
          </div>
        </div>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <div v-if="selectionMode" class="selection-toolbar">
          <span>已选 {{ selectedMessageIds.length }} 条</span>
          <div class="selection-actions">
            <button class="apple-button secondary" type="button" @click="clearSelection">取消选择</button>
            <button class="apple-button" type="button" @click="favoriteSelectedMessages">收藏</button>
          </div>
        </div>
        <div v-if="loading" class="empty-state chat-empty-state">
          <div class="empty-illustration">⌛</div>
          <strong>AI 会话加载中</strong>
        </div>
        <div v-else-if="messages.length === 0" class="empty-state chat-empty-state">
          <div class="empty-illustration">🤖</div>
          <strong>开始你的第一条 AI 对话</strong>
          <span>你可以让 AI 帮你写文案、回答问题或整理思路。</span>
        </div>
        <div v-else class="messages">
          <div
            v-for="message in messages"
            :key="message.id"
            class="message-row"
            :class="{ mine: message.role === 'user' }"
            @contextmenu.prevent="enterSelectionMode(message)"
          >
            <div class="message-item" :class="{ selected: isSelected(message) }" @click="selectionMode ? toggleSelection(message) : undefined">
              <div class="message-meta">
                <strong>{{ message.role === 'user' ? '我' : 'AI 助手' }}</strong>
                <small>{{ message.createdAt }}</small>
              </div>
              <p>{{ message.content }}</p>
            </div>
          </div>
        </div>
        <div class="composer ai-composer">
          <input
            v-model="draft"
            class="apple-input"
            :disabled="selectionMode || sending"
            placeholder="输入你的问题，按回车发送"
            @keyup.enter="sendMessage"
          />
          <button class="apple-button" :disabled="sending" @click="sendMessage">
            {{ sending ? '发送中...' : '发送' }}
          </button>
        </div>
      </section>
    </section>
  </div>
</template>

<style scoped>
.chat-theme-shell {
  position: relative;
}

.chat-shell {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 320px 1fr;
  min-height: 760px;
  overflow: hidden;
}

.ai-sidebar {
  padding: 28px 22px;
  border-right: 1px solid rgba(29, 29, 31, 0.08);
  background: rgba(255, 255, 255, 0.52);
}

.ai-banner {
  padding: 20px;
  border-radius: 24px;
  background: linear-gradient(135deg, rgba(124, 58, 237, 0.16), rgba(59, 130, 246, 0.16));
}

.sidebar-banner h2 {
  margin: 6px 0 6px;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.sidebar-banner-copy {
  margin: 0;
  color: #425466;
  font-size: 14px;
}

.chat-panel {
  display: flex;
  flex-direction: column;
  min-height: 760px;
}

.chat-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 28px 28px 18px;
  border-bottom: 1px solid rgba(29, 29, 31, 0.08);
}

.messages {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px 28px;
  overflow-y: auto;
}

.message-row {
  display: flex;
}

.message-row.mine {
  justify-content: flex-end;
}

.message-item {
  max-width: min(640px, 85%);
  padding: 16px 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.08);
}

.message-row.mine .message-item {
  background: linear-gradient(135deg, #0071e3, #34c759);
  color: #fff;
}

.message-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.message-meta small {
  color: #6e6e73;
}

.message-row.mine .message-meta small {
  color: rgba(255, 255, 255, 0.78);
}

.message-item p {
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.7;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px;
  color: #6e6e73;
}

.empty-illustration {
  font-size: 48px;
}

.ai-composer {
  display: flex;
  gap: 12px;
  padding: 18px 28px 28px;
  border-top: 1px solid rgba(29, 29, 31, 0.08);
}

.ai-composer .apple-input {
  flex: 1;
}

.favorite-entry-card {
  margin-top: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px;
  border: none;
  border-radius: 22px;
  cursor: pointer;
  text-align: left;
  background: linear-gradient(135deg, rgba(255, 184, 0, 0.16), rgba(255, 99, 71, 0.16));
}

.favorite-entry-card strong {
  display: block;
  margin-top: 4px;
  color: #1d1d1f;
  font-size: 18px;
}

.favorite-entry-card span {
  display: block;
  margin-top: 4px;
  color: #425466;
  font-size: 13px;
}

.favorite-entry-arrow {
  font-size: 22px;
  color: #ff8c00;
}

.selection-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 28px;
  background: rgba(255, 248, 230, 0.96);
  border-bottom: 1px solid rgba(255, 184, 0, 0.18);
}

.selection-actions {
  display: flex;
  gap: 10px;
}

.message-item.selected {
  box-shadow: inset 0 0 0 2px rgba(255, 140, 0, 0.45), 0 12px 28px rgba(15, 23, 42, 0.08);
}

@media (max-width: 920px) {
  .chat-shell {
    grid-template-columns: 1fr;
  }

  .ai-sidebar {
    border-right: none;
    border-bottom: 1px solid rgba(29, 29, 31, 0.08);
  }
}
</style>
