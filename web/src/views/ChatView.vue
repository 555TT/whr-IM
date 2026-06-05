<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import CreateGroupModal from '../components/CreateGroupModal.vue'
import GroupInfoPanel from '../components/GroupInfoPanel.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'
import { formatChatMessageTime } from '../utils/chat-time'
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
  createdAt?: string
}

interface GroupListItem {
  id: number
  name: string
  ownerId: number
  memberCount: number
}

interface GroupMember {
  userId: number
  username: string
  nickname: string
  avatar: string
}

interface GroupDetail {
  id: number
  name: string
  ownerId: number
  members: GroupMember[]
}

interface GroupMessage {
  id: number
  groupId: number
  senderId: number
  content: string
  createdAt?: string
}

// 渲染层统一形态:文本 + 时间 + 发送者(单聊/群聊都满足)
interface RenderMessage {
  renderKey: string
  id?: number
  senderId: number
  receiverId?: number
  groupId?: number
  sourceType?: 'friend' | 'group'
  content: string
  createdAt?: string
}

type ConversationType = 'friend' | 'group'

const authStore = useAuthStore()
const router = useRouter()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const friends = ref<FriendItem[]>([])
const groups = ref<GroupListItem[]>([])
const messages = ref<RenderMessage[]>([])
const conversationType = ref<ConversationType | null>(null)
const currentFriendId = ref<number | null>(null)
const currentGroupId = ref<number | null>(null)
const currentGroupDetail = ref<GroupDetail | null>(null)
const sidebarTab = ref<'friend' | 'group'>('friend')
const draft = ref('')
const errorMessage = ref('')
const socketConnected = ref(false)
const sending = ref(false)
const messageListRef = ref<HTMLElement | null>(null)
const showCreateGroup = ref(false)
const showGroupInfo = ref(false)
const selectionMode = ref(false)
const selectedMessageKeys = ref<string[]>([])
let socket: WebSocket | null = null

const currentFriend = computed(() => friends.value.find((item) => item.friendId === currentFriendId.value) || null)
const totalConversationCount = computed(() => friends.value.length + groups.value.length)
const currentConversationHint = computed(() => {
  if (conversationType.value === 'friend') return currentFriend.value?.signature || '单聊会话已开启'
  if (conversationType.value === 'group') return currentGroupDetail.value ? `${currentGroupDetail.value.members.length} 位成员参与会话` : '群组会话已开启'
  return '选择联系人后可开始发送实时消息'
})
const canSendMessage = computed(() => !!conversationType.value && !selectionMode.value && !sending.value)
const composerStatusText = computed(() => {
  if (!conversationType.value) return '待选择会话'
  if (selectionMode.value) return '选择模式中'
  return '可发送'
})
const headerActionLabel = computed(() => {
  if (conversationType.value === 'group') return '群信息'
  if (conversationType.value === 'friend') return '返回消息中心'
  return sidebarTab.value === 'group' ? '刷新群聊' : '刷新好友'
})

async function handleHeaderAction() {
  errorMessage.value = ''
  if (conversationType.value === 'group') {
    showGroupInfo.value = true
    return
  }
  if (conversationType.value === 'friend') {
    backToList()
    return
  }
  try {
    if (sidebarTab.value === 'group') {
      await loadGroups()
      return
    }
    await loadFriends()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

// 当前在聊会话(群或好友)的展示标题
const conversationTitle = computed(() => {
  if (conversationType.value === 'friend') return currentFriend.value?.nickname || '聊天窗口'
  if (conversationType.value === 'group') return currentGroupDetail.value?.name || '群聊'
  return '聊天窗口'
})

// 群消息发送者显示名查找
function groupMemberDisplayName(senderId: number) {
  if (!currentGroupDetail.value) return String(senderId)
  const m = currentGroupDetail.value.members.find((mm) => mm.userId === senderId)
  return m ? m.nickname || m.username : String(senderId)
}

function buildRenderKey(message: Omit<RenderMessage, 'renderKey'>) {
  if (message.id) {
    return `${message.sourceType || 'friend'}-${message.id}`
  }

  return [
    message.sourceType || 'friend',
    message.senderId,
    message.receiverId ?? 'na',
    message.groupId ?? 'na',
    message.createdAt ?? 'na',
    message.content
  ].join('-')
}

function toRenderMessage(message: ChatMessage): RenderMessage {
  const renderMessage = {
    ...message,
    sourceType: 'friend' as const
  }

  return {
    ...renderMessage,
    renderKey: buildRenderKey(renderMessage)
  }
}

async function loadFriends() {
  const { data } = await http.get('/friends')
  friends.value = data
}

async function loadGroups() {
  const { data } = await http.get('/groups')
  groups.value = data as GroupListItem[]
}

async function loadFriendMessages() {
  if (!currentFriendId.value) {
    messages.value = []
    return
  }
  const { data } = await http.get(`/normal-messages?friendId=${currentFriendId.value}`)
  messages.value = (data as ChatMessage[]).map(toRenderMessage)
  await scrollToBottom()
}

function toGroupRenderMessage(message: GroupMessage): RenderMessage {
  const renderMessage = {
    ...message,
    sourceType: 'group' as const
  }

  return {
    ...renderMessage,
    renderKey: buildRenderKey(renderMessage)
  }
}

async function loadGroupMessages() {
  if (!currentGroupId.value) {
    messages.value = []
    return
  }
  const { data } = await http.get(`/groups/${currentGroupId.value}/normal-messages`)
  messages.value = (data as GroupMessage[]).map(toGroupRenderMessage)
  await scrollToBottom()
}

async function loadGroupDetail(groupId: number) {
  const { data } = await http.get(`/groups/${groupId}`)
  currentGroupDetail.value = data as GroupDetail
}

async function selectFriend(friendId: number) {
  clearSelection()
  conversationType.value = 'friend'
  currentFriendId.value = friendId
  currentGroupId.value = null
  currentGroupDetail.value = null
  await loadFriendMessages()
}

async function selectGroup(groupId: number) {
  clearSelection()
  conversationType.value = 'group'
  currentGroupId.value = groupId
  currentFriendId.value = null
  try {
    await loadGroupDetail(groupId)
    await loadGroupMessages()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function sendMessage() {
  if (!canSendMessage.value || !draft.value.trim()) return
  errorMessage.value = ''

  if (conversationType.value === 'friend') {
    await sendFriendMessage()
  } else if (conversationType.value === 'group') {
    await sendGroupMessage()
  }
}

async function sendFriendMessage() {
  if (!currentFriendId.value) return

  sending.value = true
  try {
    const content = draft.value.trim()
    const { data } = await http.post('/normal-messages', {
      receiverId: currentFriendId.value,
      content
    })
    messages.value.push(toRenderMessage(data as ChatMessage))
    draft.value = ''
    await scrollToBottom()
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    sending.value = false
  }
}

async function sendGroupMessage() {
  if (!currentGroupId.value) return

  sending.value = true
  try {
    const content = draft.value.trim()
    const { data } = await http.post(`/groups/${currentGroupId.value}/normal-messages`, {
      content
    })
    messages.value.push(toGroupRenderMessage(data as GroupMessage))
    draft.value = ''
    await scrollToBottom()
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    sending.value = false
  }
}

async function handleCreateGroup(payload: { name: string; memberIds: number[] }) {
  errorMessage.value = ''
  try {
    const { data } = await http.post('/groups', payload)
    showCreateGroup.value = false
    await loadGroups()
    await selectGroup((data as GroupDetail).id)
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function handleInviteMembers(memberIds: number[]) {
  if (!currentGroupId.value) return
  errorMessage.value = ''
  try {
    const { data } = await http.post(`/groups/${currentGroupId.value}/members`, { memberIds })
    currentGroupDetail.value = data as GroupDetail
    showGroupInfo.value = false
    await loadGroups()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function handleLeaveGroup() {
  if (!currentGroupId.value) return
  if (!window.confirm('确认退出该群?退出后将不再收到该群的新消息。')) return
  try {
    await http.delete(`/groups/${currentGroupId.value}/members/me`)
    showGroupInfo.value = false
    conversationType.value = null
    currentGroupId.value = null
    currentGroupDetail.value = null
    messages.value = []
    await loadGroups()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

function isMine(message: { senderId: number }) {
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
    if (payload.type === 'normal_chat_message') {
      const chatMessage = payload.data as ChatMessage
      if (
        conversationType.value === 'friend' &&
        currentFriendId.value &&
        (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)
      ) {
        messages.value.push(toRenderMessage(chatMessage))
        await scrollToBottom()
      }
    } else if (payload.type === 'normal_group_message') {
      const groupMessage = payload.data as GroupMessage
      if (groupMessage.senderId === authStore.user?.id) return
      if (
        conversationType.value === 'group' &&
        currentGroupId.value === groupMessage.groupId
      ) {
        messages.value.push(toGroupRenderMessage(groupMessage))
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

function friendAvatarUrl(friend: FriendItem) {
  return friend.avatar || ''
}

function messageKey(message: RenderMessage) {
  return message.renderKey
}

function isSelected(message: RenderMessage) {
  return selectedMessageKeys.value.includes(messageKey(message))
}

function openAIChat() {
  router.push('/ai-chat')
}

function openEncryptedChat() {
  router.push('/encrypted-chat')
}

function openFavorites() {
  router.push('/favorites')
}

function enterSelectionMode(message: RenderMessage) {
  selectionMode.value = true
  toggleSelection(message)
}

function toggleSelection(message: RenderMessage) {
  const key = messageKey(message)
  if (selectedMessageKeys.value.includes(key)) {
    selectedMessageKeys.value = selectedMessageKeys.value.filter((item) => item !== key)
    if (selectedMessageKeys.value.length === 0) {
      selectionMode.value = false
    }
    return
  }
  selectedMessageKeys.value = [...selectedMessageKeys.value, key]
}

async function favoriteSelectedMessages() {
  if (selectedMessageKeys.value.length === 0) return
  errorMessage.value = ''
  const selected = messages.value.filter((message) => isSelected(message) && message.id && message.sourceType)
  try {
    await http.post('/favorites', {
      items: selected.map((message) => ({ sourceType: message.sourceType, sourceMessageId: message.id, content: message.content }))
    })
    selectionMode.value = false
    selectedMessageKeys.value = []
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

function clearSelection() {
  selectionMode.value = false
  selectedMessageKeys.value = []
}

function backToList() {
  clearSelection()
  conversationType.value = null
  currentFriendId.value = null
  currentGroupId.value = null
  currentGroupDetail.value = null
  messages.value = []
}

onMounted(async () => {
  try {
    await authStore.bootstrap()
  } catch (error) {
    errorMessage.value = (error as Error).message
    return
  }

  try {
    await loadFriends()
    await loadGroups()
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
  <div class="page-shell apple-page chat-theme-shell" :class="skin.surfaceClass">
    <AppNav />
    <section
      class="chat-shell card"
      :class="{ 'mobile-show-chat': conversationType !== null }"
    >
      <aside class="sidebar">
        <div class="sidebar-banner">
          <div>
            <p class="apple-label">IM Dashboard</p>
            <h2>消息中心</h2>
            <p class="sidebar-banner-copy">好友、群组和消息会话都集中在这里。</p>
          </div>
          <span class="signal-pill" :class="{ online: socketConnected }">
            {{ socketConnected ? '消息通道已连接' : '消息通道连接中' }}
          </span>
        </div>
        <div class="sidebar-top">
          <div>
            <p class="apple-label">Conversations</p>
            <h3>全部会话</h3>
          </div>
          <small class="muted">{{ totalConversationCount }} 个会话</small>
        </div>
        <button class="ai-entry-card" type="button" @click="openAIChat">
          <div>
            <p class="apple-label">AI Assistant</p>
            <strong>AI 助手</strong>
            <span>和智能助手聊一聊</span>
          </div>
          <span class="ai-entry-arrow">→</span>
        </button>
        <button class="encrypted-entry-card" type="button" @click="openEncryptedChat">
          <div>
            <p class="apple-label">Encrypted Channel</p>
            <strong>加密通道</strong>
            <span>进入端到端加密消息页</span>
          </div>
          <span class="encrypted-entry-arrow">⇢</span>
        </button>
        <button class="favorite-entry-card" type="button" @click="openFavorites">
          <div>
            <p class="apple-label">My Favorites</p>
            <strong>我的收藏</strong>
            <span>统一查看已收藏消息</span>
          </div>
          <span class="favorite-entry-arrow">★</span>
        </button>
        <div class="sidebar-tabs">
          <button
            class="tab-btn"
            :class="{ active: sidebarTab === 'friend' }"
            @click="sidebarTab = 'friend'"
          >好友 ({{ friends.length }})</button>
          <button
            class="tab-btn"
            :class="{ active: sidebarTab === 'group' }"
            @click="sidebarTab = 'group'"
          >群聊 ({{ groups.length }})</button>
        </div>

        <template v-if="sidebarTab === 'friend'">
          <div v-if="friends.length === 0" class="empty-state">暂无好友，请先在好友申请页添加好友。</div>
          <button
            v-for="friend in friends"
            :key="'f' + friend.friendId"
            class="friend-item"
            :class="{ active: conversationType === 'friend' && currentFriendId === friend.friendId }"
            @click="selectFriend(friend.friendId)"
          >
            <img v-if="friendAvatarUrl(friend)" :src="friend.avatar" alt="avatar" class="friend-avatar avatar-image" />
            <div v-else class="friend-avatar">{{ friend.nickname.slice(0, 1).toUpperCase() }}</div>
            <div class="friend-copy">
              <strong>{{ friend.nickname }}</strong>
              <span>{{ friend.signature || '这个人很懒，还没写签名。' }}</span>
            </div>
          </button>
        </template>

        <template v-else>
          <button class="apple-button create-group-btn" type="button" @click="showCreateGroup = true">＋ 新建群聊</button>
          <div v-if="groups.length === 0" class="empty-state">暂无群聊,点击上方按钮创建。</div>
          <button
            v-for="group in groups"
            :key="'g' + group.id"
            class="friend-item"
            :class="{ active: conversationType === 'group' && currentGroupId === group.id }"
            @click="selectGroup(group.id)"
          >
            <div class="friend-avatar group">#</div>
            <div class="friend-copy">
              <strong>{{ group.name }}</strong>
              <span>{{ group.memberCount }} 位成员</span>
            </div>
          </button>
        </template>
      </aside>

      <section class="chat-panel">
        <div class="chat-top">
          <button class="back-btn" type="button" @click="backToList" aria-label="返回会话列表">‹</button>
          <div class="chat-top-main">
            <p class="apple-label">{{ conversationType === 'group' ? 'Group Chat' : 'Direct Message' }}</p>
            <h2>{{ conversationTitle }}</h2>
            <small class="muted">{{ currentConversationHint }}</small>
            <small class="muted" v-if="authStore.user">当前身份：{{ authStore.user.nickname || authStore.user.username }}</small>
          </div>
          <button
            class="apple-button secondary refresh-btn"
            type="button"
            @click="handleHeaderAction"
          >
            {{ headerActionLabel }}
          </button>
        </div>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <div v-if="selectionMode" class="selection-toolbar">
          <span>已选 {{ selectedMessageKeys.length }} 条</span>
          <div class="selection-actions">
            <button class="apple-button secondary" type="button" @click="clearSelection">取消选择</button>
            <button class="apple-button" type="button" @click="favoriteSelectedMessages">收藏</button>
          </div>
        </div>
        <div v-if="!conversationType" class="empty-state chat-empty-state">
          <div class="empty-illustration">💬</div>
          <strong>这里是 IM 实时会话区</strong>
          <span>从左侧选择好友或群聊，立即开始收发消息。</span>
        </div>
        <div v-else ref="messageListRef" class="messages">
          <div
            v-for="message in messages"
            :key="message.renderKey"
            class="message-row"
            :class="{ mine: isMine(message) }"
            @contextmenu.prevent="enterSelectionMode(message)"
          >
            <div class="message-item" :class="{ selected: isSelected(message) }" @click="selectionMode ? toggleSelection(message) : undefined">
              <div class="message-meta">
                <strong>
                  {{
                    isMine(message)
                      ? '我'
                      : conversationType === 'group'
                      ? groupMemberDisplayName(message.senderId)
                      : currentFriend?.nickname || message.senderId
                  }}
                </strong>
                <small v-if="formatChatMessageTime(message.createdAt)">{{ formatChatMessageTime(message.createdAt) }}</small>
              </div>
              <p>{{ message.content }}</p>
            </div>
          </div>
        </div>
        <div class="composer">
          <div class="composer-meta">
            <span>普通消息</span>
            <strong>{{ composerStatusText }}</strong>
          </div>
          <input
            v-model="draft"
            class="apple-input"
            :disabled="!canSendMessage"
            placeholder="输入消息，按回车发送"
            @keyup.enter="sendMessage"
          />
          <button
            class="apple-button"
            :disabled="!canSendMessage"
            @click="sendMessage"
          >
            {{ sending ? '发送中...' : '发送' }}
          </button>
        </div>
      </section>
    </section>

    <CreateGroupModal
      :visible="showCreateGroup"
      :friends="friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname }))"
      @close="showCreateGroup = false"
      @submit="handleCreateGroup"
    />

    <GroupInfoPanel
      :visible="showGroupInfo && !!currentGroupDetail"
      :group-name="currentGroupDetail?.name || ''"
      :owner-id="currentGroupDetail?.ownerId || 0"
      :members="currentGroupDetail?.members || []"
      :my-user-id="authStore.user?.id ?? null"
      :friends="friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname }))"
      @close="showGroupInfo = false"
      @invite="handleInviteMembers"
      @leave="handleLeaveGroup"
    />
  </div>
</template>

<style scoped>
.chat-theme-shell {
  position: relative;
}

.chat-theme-shell::before {
  content: '';
  position: fixed;
  inset: 0;
  background: radial-gradient(circle at top right, var(--skin-accent-soft, rgba(0, 113, 227, 0.1)), transparent 45%);
  pointer-events: none;
  z-index: 0;
}

.chat-shell {
  position: relative;
  z-index: 1;
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

.sidebar-banner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 18px;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(41, 151, 255, 0.16), rgba(52, 211, 153, 0.16));
  box-shadow: inset 0 0 0 1px rgba(41, 151, 255, 0.12);
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

.signal-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.8);
  color: #5f6368;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}

.signal-pill.online {
  color: #0f9d58;
}

.ai-entry-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px;
  border: none;
  border-radius: 22px;
  cursor: pointer;
  text-align: left;
  background: linear-gradient(135deg, rgba(124, 58, 237, 0.16), rgba(59, 130, 246, 0.16));
  box-shadow: inset 0 0 0 1px rgba(124, 58, 237, 0.12);
}

.ai-entry-card strong {
  display: block;
  margin-top: 4px;
  color: #1d1d1f;
  font-size: 18px;
}

.ai-entry-card span {
  display: block;
  margin-top: 4px;
  color: #425466;
  font-size: 13px;
}

.ai-entry-arrow {
  font-size: 24px;
  font-weight: 700;
  color: #7c3aed;
}

.encrypted-entry-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px;
  border: none;
  border-radius: 22px;
  cursor: pointer;
  text-align: left;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.16), rgba(59, 130, 246, 0.16));
  box-shadow: inset 0 0 0 1px rgba(16, 185, 129, 0.12);
}

.encrypted-entry-card strong {
  display: block;
  margin-top: 4px;
  color: #1d1d1f;
  font-size: 18px;
}

.encrypted-entry-card span {
  display: block;
  margin-top: 4px;
  color: #425466;
  font-size: 13px;
}

.encrypted-entry-arrow {
  font-size: 22px;
  font-weight: 700;
  color: #0f9d58;
}

.favorite-entry-card {
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
  box-shadow: inset 0 0 0 1px rgba(255, 184, 0, 0.12);
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
  font-weight: 700;
  color: #ff8c00;
}

.selection-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 20px;
  background: rgba(255, 248, 230, 0.96);
  border-radius: 18px;
}

.selection-actions {
  display: flex;
  gap: 10px;
}

.sidebar-top,
.chat-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.sidebar-top h3,
.chat-top h2 {
  margin: 6px 0 0;
  font-size: 28px;
  letter-spacing: -0.03em;
}

.sidebar-top h3 {
  font-size: 22px;
}

.chat-top-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 0;
}

.chat-top-main h2,
.chat-top-main small {
  max-width: 100%;
  overflow-wrap: anywhere;
}

.chat-top-main h2 {
  margin-top: 0;
  line-height: 1.1;
}

.chat-top-main small {
  display: block;
  margin: 0;
  line-height: 1.4;
}

.chat-top-main small:first-of-type {
  color: #425466;
  font-size: 13px;
  font-weight: 500;
}

.chat-top-main small:last-of-type {
  color: #6e6e73;
  font-size: 12px;
}

.chat-empty-state {
  min-height: 320px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px;
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(245, 248, 255, 0.92));
  text-align: center;
}

.empty-illustration {
  width: 72px;
  height: 72px;
  border-radius: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(41, 151, 255, 0.12);
  font-size: 34px;
}

.chat-empty-state strong {
  font-size: 20px;
}

.chat-empty-state span {
  color: #6e6e73;
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

.avatar-image {
  object-fit: cover;
  display: block;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.14);
}

.friend-avatar.group {
  background: linear-gradient(180deg, #fff5d6 0%, #ffe7a8 100%);
  color: #b25e00;
}

.sidebar-tabs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  margin-bottom: 4px;
}

.tab-btn {
  padding: 8px 10px;
  border: none;
  border-radius: 12px;
  background: rgba(245, 245, 247, 0.6);
  color: #1d1d1f;
  font-weight: 500;
  cursor: pointer;
  font-size: 13px;
}

.tab-btn.active {
  background: rgba(0, 113, 227, 0.12);
  color: #0071e3;
  box-shadow: inset 0 0 0 1px rgba(0, 113, 227, 0.18);
}

.create-group-btn {
  width: 100%;
  margin-bottom: 4px;
  font-size: 13px;
  padding: 10px 14px;
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

.message-item.selected {
  box-shadow: inset 0 0 0 2px rgba(255, 140, 0, 0.45), 0 12px 28px rgba(15, 23, 42, 0.08);
}

.message-row.mine .message-item {
  background: linear-gradient(180deg, #d7ebff 0%, #c6e0ff 100%);
}

.message-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 6px;
}

.message-item strong {
  display: block;
}

.message-meta small {
  color: #6e6e73;
}

.message-item p {
  margin: 0;
}

.composer {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 12px;
  align-items: center;
}

.composer-meta {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 92px;
  min-height: 56px;
  padding: 10px 12px;
  border-radius: 18px;
  background: rgba(245, 245, 247, 0.92);
  box-shadow: inset 0 0 0 1px rgba(29, 29, 31, 0.05);
  text-align: left;
}

.composer-meta span {
  color: #6e6e73;
  font-size: 12px;
  line-height: 1.3;
}

.composer-meta strong {
  margin-top: 2px;
  color: #0f172a;
  font-size: 14px;
  line-height: 1.3;
}

.composer .apple-input {
  min-width: 0;
}

.composer .apple-button {
  white-space: nowrap;
}

/* 桌面端默认隐藏移动端独有的返回按钮 */
.back-btn {
  display: none;
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

/* ≤768px：手机端走"列表/对话切换"模式 */
@media (max-width: 768px) {
  .chat-shell {
    min-height: 0;
    display: block; /* 不再 grid，由可见性切换决定显示哪一栏 */
    overflow: visible;
  }

  .sidebar {
    padding: 18px 14px;
    border-bottom: none;
  }

  .chat-panel {
    display: none;
    padding: 16px 14px calc(16px + env(safe-area-inset-bottom));
    min-height: calc(100vh - 140px);
  }

  /* 选中好友后：隐藏列表、显示对话 */
  .chat-shell.mobile-show-chat .sidebar {
    display: none;
  }
  .chat-shell.mobile-show-chat .chat-panel {
    display: flex;
  }

  .sidebar-top h3,
  .chat-top h2 {
    font-size: 22px;
    margin-top: 2px;
  }

  .sidebar-banner {
    padding: 14px;
  }

  .sidebar-banner h2 {
    font-size: 24px;
  }

  .chat-top {
    align-items: center;
    gap: 10px;
  }

  .back-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border: none;
    border-radius: 50%;
    background: rgba(29, 29, 31, 0.06);
    color: #1d1d1f;
    font-size: 22px;
    line-height: 1;
    cursor: pointer;
    flex-shrink: 0;
  }

  .chat-top-main {
    flex: 1;
    min-width: 0;
  }

  .chat-empty-state {
    min-height: 260px;
  }

  .refresh-btn {
    padding: 8px 12px;
    font-size: 13px;
  }

  .messages {
    min-height: 0;
    max-height: none;
    flex: 1;
  }

  .message-item {
    max-width: 82%;
    padding: 12px 14px;
    border-radius: 18px;
  }

  .friend-item {
    padding: 12px;
    border-radius: 16px;
  }

  .composer {
    grid-template-columns: 1fr auto;
    gap: 8px;
    position: sticky;
    bottom: 0;
    background: rgba(255, 255, 255, 0.85);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    padding-top: 8px;
    margin: 0 -14px -16px;
    padding-left: 14px;
    padding-right: 14px;
    padding-bottom: calc(8px + env(safe-area-inset-bottom));
  }

  .composer-meta {
    grid-column: 1 / -1;
    min-width: 0;
  }

  .composer .apple-input {
    padding: 12px 14px;
    border-radius: 14px;
  }

  .composer .apple-button {
    padding: 12px 16px;
  }
}
</style>
