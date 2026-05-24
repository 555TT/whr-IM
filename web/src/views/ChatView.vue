<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { createChatSocket } from '../utils/websocket'

interface FriendItem {
  userId: number
  friendId: number
  nickname: string
  avatar: string
  signature: string
}

interface ChatMessage {
  id?: number
  senderId: number
  receiverId: number
  content: string
  msgType?: string
}

const authStore = useAuthStore()
const friends = ref<FriendItem[]>([])
const messages = ref<ChatMessage[]>([])
const currentFriendId = ref<number | null>(null)
const draft = ref('')
const errorMessage = ref('')
const socketConnected = ref(false)
let socket: WebSocket | null = null

const currentFriend = computed(() => friends.value.find((item) => item.friendId === currentFriendId.value) || null)

async function loadFriends() {
  const { data } = await http.get('/friends')
  friends.value = data
  if (!currentFriendId.value && friends.value.length > 0) {
    currentFriendId.value = friends.value[0].friendId
    await loadMessages()
  }
  if (currentFriendId.value && !friends.value.some((item) => item.friendId === currentFriendId.value)) {
    currentFriendId.value = friends.value[0]?.friendId || null
    await loadMessages()
  }
}

async function loadMessages() {
  if (!currentFriendId.value) {
    messages.value = []
    return
  }
  const { data } = await http.get(`/messages?friendId=${currentFriendId.value}`)
  messages.value = data
}

async function selectFriend(friendId: number) {
  currentFriendId.value = friendId
  await loadMessages()
}

async function sendMessage() {
  if (!currentFriendId.value || !draft.value.trim()) return
  errorMessage.value = ''
  try {
    const { data } = await http.post('/messages', {
      receiverId: currentFriendId.value,
      content: draft.value.trim()
    })
    messages.value.push(data)
    draft.value = ''
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

function isMine(message: ChatMessage) {
  return message.senderId === authStore.user?.id
}

function connectSocket() {
  if (!authStore.token) return
  socket = createChatSocket(authStore.token)
  socket.onopen = () => {
    socketConnected.value = true
  }
  socket.onclose = () => {
    socketConnected.value = false
  }
  socket.onerror = () => {
    errorMessage.value = 'WebSocket 连接失败'
  }
  socket.onmessage = (event) => {
    const payload = JSON.parse(event.data)
    if (payload.type === 'chat_message') {
      const chatMessage = payload.data as ChatMessage
      if (
        currentFriendId.value &&
        (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)
      ) {
        messages.value.push(chatMessage)
      }
    }
  }
}

onMounted(async () => {
  try {
    await authStore.bootstrap()
    await loadFriends()
    connectSocket()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
})

onBeforeUnmount(() => {
  socket?.close()
})
</script>

<template>
  <div class="page-shell page-layout">
    <AppNav />
    <div class="chat-layout">
      <aside class="card sidebar">
        <div class="sidebar-header">
          <h3>好友列表</h3>
          <small>{{ socketConnected ? 'WS 已连接' : 'WS 未连接' }}</small>
        </div>
        <div v-if="friends.length === 0" class="empty-state">暂无好友，请先在好友申请页添加好友。</div>
        <button
          v-for="friend in friends"
          :key="friend.friendId"
          class="friend-item"
          :class="{ active: currentFriendId === friend.friendId }"
          @click="selectFriend(friend.friendId)"
        >
          <strong>{{ friend.nickname }}</strong>
          <small>ID: {{ friend.friendId }}</small>
          <span>{{ friend.signature || '这个人很懒，还没写签名。' }}</span>
        </button>
      </aside>
      <section class="card chat-panel">
        <div class="chat-header">
          <div>
            <h3>聊天窗口</h3>
            <small v-if="currentFriend">当前聊天：{{ currentFriend.nickname }}</small>
          </div>
          <button class="refresh-btn" @click="loadFriends">刷新好友</button>
        </div>
        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
        <div v-if="!currentFriendId" class="empty-state">请选择一个好友开始聊天</div>
        <div v-else class="messages">
          <div
            v-for="(message, index) in messages"
            :key="index"
            class="message-row"
            :class="{ mine: isMine(message) }"
          >
            <div class="message-item">
              <strong>{{ isMine(message) ? '我' : currentFriend?.nickname || message.senderId }}</strong>
              <p>{{ message.content }}</p>
            </div>
          </div>
        </div>
        <div class="composer">
          <input v-model="draft" :disabled="!currentFriendId" placeholder="输入消息" @keyup.enter="sendMessage" />
          <button :disabled="!currentFriendId" @click="sendMessage">发送</button>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.page-layout {
  padding: 24px;
}

.chat-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 20px;
}

.sidebar,
.chat-panel {
  padding: 20px;
}

.sidebar {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sidebar-header,
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.friend-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 12px;
  border: 1px solid #dbe4f0;
  border-radius: 8px;
  background: #f8fbff;
  color: #1f2937;
}

.friend-item.active {
  border-color: #2563eb;
  background: #eff6ff;
}

.messages {
  min-height: 360px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 16px 0;
}

.message-row {
  display: flex;
}

.message-row.mine {
  justify-content: flex-end;
}

.message-item {
  max-width: 70%;
  padding: 12px;
  border-radius: 8px;
  background: #f3f4f6;
}

.message-row.mine .message-item {
  background: #dbeafe;
}

.composer {
  display: flex;
  gap: 12px;
}

input {
  flex: 1;
  padding: 12px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
}

button {
  padding: 12px 16px;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: white;
  cursor: pointer;
}

.refresh-btn {
  background: #4b5563;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.empty-state {
  color: #6b7280;
}

.error-text {
  color: #dc2626;
}
</style>
