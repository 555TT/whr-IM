<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import AppNav from '../components/AppNav.vue'
import CreateGroupModal from '../components/CreateGroupModal.vue'
import GroupInfoPanel from '../components/GroupInfoPanel.vue'
import { http } from '../api/http'
import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'
import { formatChatMessageTime } from '../utils/chat-time'
import {
  buildEncryptedMessageDisplay,
  decryptMessage,
  E2EE_MESSAGE_ALGORITHM,
  encryptMessage,
  exportPrivateKey,
  exportPublicKey,
  generateKeyPair,
  importPrivateKey,
  importPublicKey,
  loadPrivateKey,
  savePrivateKey,
  selectMessagePayloadForUser
} from '../utils/e2ee'
import {
  buildGroupMessageEnvelope,
  decryptGroupMessage,
  GROUP_CONTENT_ALGORITHM
} from '../utils/group-e2ee'
import { createChatSocket } from '../utils/websocket'

interface FriendItem {
  userId: number
  friendId: number
  nickname: string
  avatar: string
  signature: string
  publicKey?: string
  publicKeyAlgorithm?: string
}

interface ChatMessage {
  id?: number
  senderId: number
  receiverId: number
  senderCiphertext: string
  senderAlgorithm: string
  receiverCiphertext: string
  receiverAlgorithm: string
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
  publicKey?: string
  publicKeyAlgorithm?: string
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
  contentCiphertext: string
  contentIv: string
  contentAlgorithm: string
  keyCiphertext: string
  keyAlgorithm: string
  createdAt?: string
}

// 渲染层统一形态:文本 + 时间 + 发送者(单聊/群聊都满足)
interface RenderMessage {
  id?: number
  senderId: number
  // 单聊字段(可选,仅单聊用)
  receiverId?: number
  senderCiphertext?: string
  receiverCiphertext?: string
  // 群聊字段(可选)
  groupId?: number
  // 渲染数据
  content: string
  createdAt?: string
}

type ConversationType = 'friend' | 'group'

const authStore = useAuthStore()
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
const cryptoReady = ref(true) // 端到端加密是否就绪，未就绪时禁用发送
const socketConnected = ref(false)
const sending = ref(false)
const privateKey = ref<CryptoKey | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const showCreateGroup = ref(false)
const showGroupInfo = ref(false)
let socket: WebSocket | null = null

const currentFriend = computed(() => friends.value.find((item) => item.friendId === currentFriendId.value) || null)

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

async function uploadOwnPublicKey(serializedPublicKey: string) {
  await http.put('/users/me/public-key', {
    publicKey: serializedPublicKey,
    algorithm: E2EE_MESSAGE_ALGORITHM
  })

  authStore.user = authStore.user
    ? {
        ...authStore.user,
        publicKey: serializedPublicKey,
        publicKeyAlgorithm: E2EE_MESSAGE_ALGORITHM
      }
    : null
}

async function exportPublicKeyFromPrivateKey(serializedPrivateKey: string) {
  const importedPrivateKey = await importPrivateKey(serializedPrivateKey)
  const jwk = await window.crypto.subtle.exportKey('jwk', importedPrivateKey)
  const publicKey = await window.crypto.subtle.importKey(
    'jwk',
    {
      kty: jwk.kty,
      n: jwk.n,
      e: jwk.e,
      alg: jwk.alg,
      ext: true,
      key_ops: ['encrypt']
    },
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    true,
    ['encrypt']
  )

  return exportPublicKey(publicKey)
}

async function ensureOwnKeyPair() {
  if (!authStore.user?.id) return

  const storedPrivateKey = loadPrivateKey(authStore.user.id)
  if (storedPrivateKey) {
    privateKey.value = await importPrivateKey(storedPrivateKey)

    if (authStore.user.publicKey && authStore.user.publicKeyAlgorithm === E2EE_MESSAGE_ALGORITHM) {
      return
    }

    const serializedPublicKey = await exportPublicKeyFromPrivateKey(storedPrivateKey)
    await uploadOwnPublicKey(serializedPublicKey)
    return
  }

  const keyPair = await generateKeyPair()
  const serializedPublicKey = await exportPublicKey(keyPair.publicKey)
  const serializedPrivateKey = await exportPrivateKey(keyPair.privateKey)

  savePrivateKey(authStore.user.id, serializedPrivateKey)
  privateKey.value = keyPair.privateKey

  await uploadOwnPublicKey(serializedPublicKey)
}

async function toRenderMessage(message: ChatMessage): Promise<RenderMessage> {
  const payload = selectMessagePayloadForUser(message, authStore.user?.id)

  if (!privateKey.value) {
    return {
      ...message,
      ...buildEncryptedMessageDisplay('', payload)
    }
  }

  try {
    const content = await decryptMessage(privateKey.value, payload.ciphertext)
    return {
      ...message,
      ...buildEncryptedMessageDisplay(content, payload)
    }
  } catch {
    return {
      ...message,
      ...buildEncryptedMessageDisplay('', payload)
    }
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
  const { data } = await http.get(`/messages?friendId=${currentFriendId.value}`)
  messages.value = await Promise.all((data as ChatMessage[]).map(toRenderMessage))
  await scrollToBottom()
}

async function toGroupRenderMessage(message: GroupMessage): Promise<RenderMessage> {
  const base: RenderMessage = {
    id: message.id,
    senderId: message.senderId,
    groupId: message.groupId,
    content: '***（已加密）',
    createdAt: message.createdAt
  }
  if (!privateKey.value) return base
  try {
    base.content = await decryptGroupMessage(
      privateKey.value,
      message.keyCiphertext,
      message.contentCiphertext,
      message.contentIv
    )
  } catch {
    // 保持加密占位
  }
  return base
}

async function loadGroupMessages() {
  if (!currentGroupId.value) {
    messages.value = []
    return
  }
  const { data } = await http.get(`/groups/${currentGroupId.value}/messages`)
  messages.value = await Promise.all((data as GroupMessage[]).map(toGroupRenderMessage))
  await scrollToBottom()
}

async function loadGroupDetail(groupId: number) {
  const { data } = await http.get(`/groups/${groupId}`)
  currentGroupDetail.value = data as GroupDetail
}

async function selectFriend(friendId: number) {
  conversationType.value = 'friend'
  currentFriendId.value = friendId
  currentGroupId.value = null
  currentGroupDetail.value = null
  await loadFriendMessages()
}

async function selectGroup(groupId: number) {
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
  if (!draft.value.trim() || sending.value) return
  errorMessage.value = ''

  if (conversationType.value === 'friend') {
    await sendFriendMessage()
  } else if (conversationType.value === 'group') {
    await sendGroupMessage()
  }
}

async function sendFriendMessage() {
  if (!currentFriendId.value) return

  if (!privateKey.value) {
    errorMessage.value = '当前设备没有可用私钥'
    return
  }

  const friend = currentFriend.value
  if (!friend?.publicKey) {
    errorMessage.value = '对方未启用端到端加密消息'
    return
  }

  if (!authStore.user?.publicKey) {
    errorMessage.value = '当前账号公钥不可用'
    return
  }

  sending.value = true
  try {
    const receiverPublicKey = await importPublicKey(friend.publicKey)
    const senderPublicKey = await importPublicKey(authStore.user.publicKey)
    const content = draft.value.trim()
    const receiverEncrypted = await encryptMessage(receiverPublicKey, content)
    const senderEncrypted = await encryptMessage(senderPublicKey, content)
    const { data } = await http.post('/messages', {
      receiverId: currentFriendId.value,
      senderCiphertext: senderEncrypted.ciphertext,
      senderAlgorithm: senderEncrypted.algorithm,
      receiverCiphertext: receiverEncrypted.ciphertext,
      receiverAlgorithm: receiverEncrypted.algorithm
    })
    messages.value.push(await toRenderMessage(data as ChatMessage))
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
  if (!privateKey.value) {
    errorMessage.value = '当前设备没有可用私钥'
    return
  }

  // 发消息前刷新一次成员列表,避免缺漏新成员的密钥包裹
  sending.value = true
  try {
    await loadGroupDetail(currentGroupId.value)
    const detail = currentGroupDetail.value
    if (!detail) throw new Error('群信息加载失败')

    const members = detail.members.map((m) => ({
      userId: m.userId,
      publicKey: m.publicKey || '',
      publicKeyAlgorithm: m.publicKeyAlgorithm || ''
    }))
    const content = draft.value.trim()
    const envelope = await buildGroupMessageEnvelope(content, members)
    const { data } = await http.post(`/groups/${currentGroupId.value}/messages`, {
      contentCiphertext: envelope.contentCiphertext,
      contentIv: envelope.contentIv,
      contentAlgorithm: envelope.contentAlgorithm,
      memberKeys: envelope.memberKeys
    })
    messages.value.push(await toGroupRenderMessage(data as GroupMessage))
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
    if (payload.type === 'chat_message') {
      const chatMessage = payload.data as ChatMessage
      if (
        conversationType.value === 'friend' &&
        currentFriendId.value &&
        (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)
      ) {
        messages.value.push(await toRenderMessage(chatMessage))
        await scrollToBottom()
      }
    } else if (payload.type === 'group_message') {
      const groupMessage = payload.data as GroupMessage
      // 自己发的消息已经在 sendGroupMessage 里 push 过,避免重复
      if (groupMessage.senderId === authStore.user?.id) return
      if (
        conversationType.value === 'group' &&
        currentGroupId.value === groupMessage.groupId
      ) {
        messages.value.push(await toGroupRenderMessage(groupMessage))
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

function backToList() {
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

  // 加密初始化失败不应阻断好友列表与历史消息的展示，因此独立 try
  try {
    await ensureOwnKeyPair()
  } catch (error) {
    cryptoReady.value = false
    errorMessage.value = (error as Error).message
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
        <div class="sidebar-top">
          <div>
            <p class="apple-label">Conversations</p>
            <h2>会话</h2>
          </div>
          <small class="muted">{{ socketConnected ? '在线同步中' : '等待连接' }}</small>
        </div>
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
            <div class="friend-avatar">{{ friend.nickname.slice(0, 1).toUpperCase() }}</div>
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
            <p class="apple-label">{{ conversationType === 'group' ? 'Group' : 'Conversation' }}</p>
            <h2>{{ conversationTitle }}</h2>
            <small class="muted" v-if="authStore.user">当前身份：{{ authStore.user.nickname || authStore.user.username }}</small>
          </div>
          <button v-if="conversationType === 'group'" class="apple-button secondary refresh-btn" @click="showGroupInfo = true">群信息</button>
          <button v-else class="apple-button secondary refresh-btn" @click="loadFriends">刷新好友</button>
        </div>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <p v-if="!cryptoReady && !errorMessage" class="status-text error">
          当前环境不支持端到端加密，仅可查看已有会话。
        </p>
        <div v-if="!conversationType" class="empty-state">请选择一个好友或群聊开始聊天</div>
        <div v-else ref="messageListRef" class="messages">
          <div
            v-for="(message, index) in messages"
            :key="index"
            class="message-row"
            :class="{ mine: isMine(message) }"
          >
            <div class="message-item">
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
          <input
            v-model="draft"
            class="apple-input"
            :disabled="!conversationType || sending || !cryptoReady"
            :placeholder="cryptoReady ? '输入消息' : '当前环境不支持发送加密消息'"
            @keyup.enter="sendMessage"
          />
          <button
            class="apple-button"
            :disabled="!conversationType || sending || !cryptoReady"
            @click="sendMessage"
          >
            {{ sending ? '发送中...' : '发送' }}
          </button>
        </div>
      </section>
    </section>

    <CreateGroupModal
      :visible="showCreateGroup"
      :friends="friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))"
      @close="showCreateGroup = false"
      @submit="handleCreateGroup"
    />

    <GroupInfoPanel
      :visible="showGroupInfo && !!currentGroupDetail"
      :group-name="currentGroupDetail?.name || ''"
      :owner-id="currentGroupDetail?.ownerId || 0"
      :members="currentGroupDetail?.members || []"
      :my-user-id="authStore.user?.id ?? null"
      :friends="friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))"
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
  grid-template-columns: 1fr auto;
  gap: 12px;
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

  .sidebar-top h2,
  .chat-top h2 {
    font-size: 22px;
    margin-top: 2px;
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

  .composer .apple-input {
    padding: 12px 14px;
    border-radius: 14px;
  }

  .composer .apple-button {
    padding: 12px 16px;
  }
}
</style>
