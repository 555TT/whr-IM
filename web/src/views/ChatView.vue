<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

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
const sending = ref(false)
const messageListRef = ref<HTMLElement | null>(null)
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
  await scrollToBottom()
}

async function selectFriend(friendId: number) {
  currentFriendId.value = friendId
  await loadMessages()
}

async function sendMessage() {
  if (!currentFriendId.value || !draft.value.trim() || sending.value) return
  errorMessage.value = ''
  sending.value = true
  try {
    const { data } = await http.post('/messages', {
      receiverId: currentFriendId.value,
      content: draft.value.trim()
    })
    messages.value.push(data)
    draft.value = ''
    await scrollToBottom()
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    sending.value = false
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
  socket.onmessage = async (event) => {
    const payload = JSON.parse(event.data)
    if (payload.type === 'chat_message') {
      const chatMessage = payload.data as ChatMessage
      if (
        currentFriendId.value &&
        (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)
      ) {
        messages.value.push(chatMessage)
        await scrollToBottom()
      }
    }
  }
}

async function scrollToBottom() {
  await nextTick()
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

watch(messages, () => {
  scrollToBottom()
}, { deep: true })

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
  <div class="page-shell apple-page">
    <AppNav />
    <section class="chat-shell card">
      <aside class="sidebar">
        <div class="sidebar-top">
          <div>
            <p class="apple-label">Contacts</p>
            <h2>好友列表</h2>
          </div>
          <small class="muted">{{ socketConnected ? '在线同步中' : '等待连接' }}</small>
        </div>
        <div v-if="friends.length === 0" class="empty-state">暂无好友，请先在好友申请页添加好友。</div>
        <button
          v-for="friend in friends"
          :key="friend.friendId"
          class="friend-item"
          :class="{ active: currentFriendId === friend.friendId }"
          @click="selectFriend(friend.friendId)"
        >
          <div class="friend-avatar">{{ friend.nickname.slice(0, 1).toUpperCase() }}</div>
          <div class="friend-copy">
            <strong>{{ friend.nickname }}</strong>
            <small>ID: {{ friend.friendId }}</small>
            <span>{{ friend.signature || '这个人很懒，还没写签名。' }}</span>
          </div>
        </button>
      </aside>

      <section class="chat-panel">
        <div class="chat-top">
          <div>
            <p class="apple-label">Conversation</p>
            <h2>{{ currentFriend ? currentFriend.nickname : '聊天窗口' }}</h2>
            <small class="muted" v-if="authStore.user">当前身份：{{ authStore.user.nickname || authStore.user.username }}</small>
          </div>
          <button class="apple-button secondary" @click="loadFriends">刷新好友</button>
        </div>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <div v-if="!currentFriendId" class="empty-state">请选择一个好友开始聊天</div>
        <div v-else ref="messageListRef" class="messages">
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
          <input
            v-model="draft"
            class="apple-input"
            :disabled="!currentFriendId || sending"
            placeholder="输入消息"
            @keyup.enter="sendMessage"
          />
          <button class="apple-button" :disabled="!currentFriendId || sending" @click="sendMessage">
            {{ sending ? '发送中...' : '发送' }}
          </button>
        </div>
      </section>
    </section>
  </div>
</template>

<style scoped>
.chat-shell {
  display: grid;
  grid-template-columns: 340px 1fr;
  min-height: 760px;
  overflow: hidden;
}

.sidebar {
  padding: 28px 22px;
  border-right: 1px solid rgba(29, 29, 31, 0.08);
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: rgba(255, 255, 255, 0.52);
}

.sidebar-top,
.chat-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.sidebar-top h2,
.chat-top h2 {
  margin: 6px 0 0;
  font-size: 28px;
  letter-spacing: -0.03em;
}

.friend-item {
  display: grid;
  grid-template-columns: 48px 1fr;
  gap: 14px;
  align-items: center;
  padding: 14px;
  border: none;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.82);
  text-align: left;
  box-shadow: inset 0 0 0 1px rgba(29, 29, 31, 0.05);
}

.friend-item.active {
  background: rgba(0, 113, 227, 0.12);
  box-shadow: inset 0 0 0 1px rgba(0, 113, 227, 0.18);
}

.friend-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, #eef5ff 0%, #d9e8ff 100%);
  color: #0071e3;
  font-weight: 700;
}

.friend-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.friend-copy small,
.friend-copy span {
  color: #6e6e73;
}

.chat-panel {
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.messages {
  flex: 1;
  min-height: 420px;
  max-height: 520px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 8px 6px 8px 0;
}

.message-row {
  display: flex;
}

.message-row.mine {
  justify-content: flex-end;
}

.message-item {
  max-width: 68%;
  padding: 14px 16px;
  border-radius: 22px;
  background: rgba(245, 245, 247, 0.95);
  box-shadow: inset 0 0 0 1px rgba(29, 29, 31, 0.04);
}

.message-row.mine .message-item {
  background: linear-gradient(180deg, #d7ebff 0%, #c6e0ff 100%);
}

.message-item strong {
  display: block;
  margin-bottom: 6px;
}

.message-item p {
  margin: 0;
}

.composer {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
}

@media (max-width: 980px) {
  .chat-shell {
    grid-template-columns: 1fr;
  }

  .sidebar {
    border-right: none;
    border-bottom: 1px solid rgba(29, 29, 31, 0.08);
  }
}
</style>
