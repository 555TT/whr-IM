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
  decryptGroupMessage
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

interface RenderMessage {
  id?: number
  senderId: number
  receiverId?: number
  senderCiphertext?: string
  receiverCiphertext?: string
  groupId?: number
  sourceType?: 'friend' | 'group'
  content: string
  createdAt?: string
}

type ConversationType = 'friend' | 'group'

const authStore = useAuthStore()
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
const friends = ref<FriendItem[]>([])
const groups = ref<GroupListItem[]>([])
const groupDetails = ref<Record<number, GroupDetail>>({})
const messages = ref<RenderMessage[]>([])
const conversationType = ref<ConversationType | null>(null)
const currentFriendId = ref<number | null>(null)
const currentGroupId = ref<number | null>(null)
const currentGroupDetail = ref<GroupDetail | null>(null)
const sidebarTab = ref<'friend' | 'group'>('friend')
const draft = ref('')
const errorMessage = ref('')
const cryptoReady = ref(true)
const socketConnected = ref(false)
const sending = ref(false)
const privateKey = ref<CryptoKey | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const showCreateGroup = ref(false)
const showGroupInfo = ref(false)
let socket: WebSocket | null = null

function hasSupportedPublicKey(publicKey?: string, publicKeyAlgorithm?: string) {
  return Boolean(publicKey && publicKeyAlgorithm === E2EE_MESSAGE_ALGORITHM)
}

function isSecureGroupDetail(detail?: GroupDetail | null) {
  return Boolean(
    detail?.members.length && detail.members.every((member) => hasSupportedPublicKey(member.publicKey, member.publicKeyAlgorithm))
  )
}

const secureFriends = computed(() =>
  friends.value.filter((friend) => hasSupportedPublicKey(friend.publicKey, friend.publicKeyAlgorithm))
)

const secureGroups = computed(() =>
  groups.value.filter((group) => isSecureGroupDetail(groupDetails.value[group.id]))
)

const currentFriend = computed(() => secureFriends.value.find((item) => item.friendId === currentFriendId.value) || null)
const totalConversationCount = computed(() => secureFriends.value.length + secureGroups.value.length)
const encryptedAvailabilityText = computed(() => (cryptoReady.value ? '端到端加密已就绪' : '当前设备无法发送加密消息'))
const friendSecureCapabilityText = computed(() => {
  if (secureFriends.value.length > 0) return `${secureFriends.value.length} 位联系人已满足加密条件`
  if (friends.value.length === 0) return '暂无好友可供建立安全会话'
  return '现有好友尚未全部满足安全加密条件'
})
const groupSecureCapabilityText = computed(() => {
  if (secureGroups.value.length > 0) return `${secureGroups.value.length} 个群聊可安全收发`
  if (groups.value.length === 0) return '暂无群聊可供建立安全会话'
  return '现有群聊成员中仍存在未启用安全密钥的对象'
})
const currentConversationHint = computed(() => {
  if (conversationType.value === 'friend') return currentFriend.value?.signature || '当前是双人加密直聊'
  if (conversationType.value === 'group') return currentGroupDetail.value ? `${currentGroupDetail.value.members.length} 位成员参与加密群聊` : '当前是加密群组会话'
  return '选择联系人或群组后即可开始安全通信'
})
const conversationTitle = computed(() => {
  if (conversationType.value === 'friend') return currentFriend.value?.nickname || '加密直聊'
  if (conversationType.value === 'group') return currentGroupDetail.value?.name || '加密群聊'
  return 'Encrypted Channel'
})

function groupMemberDisplayName(senderId: number) {
  if (!currentGroupDetail.value) return String(senderId)
  const member = currentGroupDetail.value.members.find((item) => item.userId === senderId)
  return member ? member.nickname || member.username : String(senderId)
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
      sourceType: 'friend',
      ...buildEncryptedMessageDisplay('', payload)
    }
  }

  try {
    const content = await decryptMessage(privateKey.value, payload.ciphertext)
    return {
      ...message,
      sourceType: 'friend',
      ...buildEncryptedMessageDisplay(content, payload)
    }
  } catch {
    return {
      ...message,
      sourceType: 'friend',
      ...buildEncryptedMessageDisplay('', payload)
    }
  }
}

async function loadFriends() {
  const { data } = await http.get('/friends')
  friends.value = data
}

async function loadGroupDetail(groupId: number, options: { persist?: boolean } = {}) {
  const { data } = await http.get(`/groups/${groupId}`)
  const detail = data as GroupDetail
  if (options.persist !== false) {
    groupDetails.value = {
      ...groupDetails.value,
      [groupId]: detail
    }
  }
  return detail
}

async function loadGroups() {
  const { data } = await http.get('/groups')
  const loadedGroups = data as GroupListItem[]
  groups.value = loadedGroups
  const nextGroupIds = new Set(loadedGroups.map((group) => group.id))
  const retainedDetails = Object.fromEntries(
    Object.entries(groupDetails.value).filter(([groupId]) => nextGroupIds.has(Number(groupId)))
  ) as Record<number, GroupDetail>
  const detailEntries = await Promise.all(
    loadedGroups.map(async (group) => {
      try {
        const detail = await loadGroupDetail(group.id, { persist: false })
        return [group.id, detail] as const
      } catch {
        return [group.id, retainedDetails[group.id] ?? null] as const
      }
    })
  )
  groupDetails.value = {
    ...retainedDetails,
    ...Object.fromEntries(detailEntries.filter((entry): entry is readonly [number, GroupDetail] => Boolean(entry[1])))
  }
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
    sourceType: 'group',
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

async function selectFriend(friendId: number) {
  const friend = secureFriends.value.find((item) => item.friendId === friendId)
  if (!friend) {
    conversationType.value = null
    currentFriendId.value = null
    currentGroupId.value = null
    currentGroupDetail.value = null
    messages.value = []
    errorMessage.value = '该联系人尚未完成当前端到端加密协议配置，暂不可进入安全会话。'
    return
  }

  errorMessage.value = ''
  conversationType.value = 'friend'
  currentFriendId.value = friendId
  currentGroupId.value = null
  currentGroupDetail.value = null
  await loadFriendMessages()
}

async function selectGroup(groupId: number) {
  const group = secureGroups.value.find((item) => item.id === groupId)
  if (!group) {
    conversationType.value = null
    currentFriendId.value = null
    currentGroupId.value = null
    currentGroupDetail.value = null
    messages.value = []
    errorMessage.value = '该群聊存在未满足当前端到端加密协议的成员，暂不可进入安全会话。'
    return
  }

  const previousConversationType = conversationType.value
  const previousFriendId = currentFriendId.value
  const previousGroupId = currentGroupId.value
  const previousGroupDetail = currentGroupDetail.value
  const previousMessages = [...messages.value]

  errorMessage.value = ''
  conversationType.value = 'group'
  currentGroupId.value = groupId
  currentFriendId.value = null
  messages.value = []
  try {
    currentGroupDetail.value = await loadGroupDetail(groupId)
    await loadGroupMessages()
  } catch (error) {
    conversationType.value = previousConversationType
    currentFriendId.value = previousFriendId
    currentGroupId.value = previousGroupId
    currentGroupDetail.value = previousGroupDetail
    messages.value = previousMessages
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
  if (!friend || !hasSupportedPublicKey(friend.publicKey, friend.publicKeyAlgorithm)) {
    errorMessage.value = '当前目标未满足安全会话要求，请确认对方已启用受支持的端到端加密公钥。'
    return
  }

  const ownPublicKey = authStore.user?.publicKey
  const ownPublicKeyAlgorithm = authStore.user?.publicKeyAlgorithm
  if (!hasSupportedPublicKey(ownPublicKey, ownPublicKeyAlgorithm)) {
    errorMessage.value = '当前账号缺少受支持的端到端加密公钥，暂时无法发送安全消息。'
    return
  }

  const receiverPublicKeyValue = friend.publicKey
  if (!receiverPublicKeyValue) {
    errorMessage.value = '当前目标缺少可用的端到端加密公钥，暂时无法建立安全会话。'
    return
  }

  if (!ownPublicKey) {
    errorMessage.value = '当前账号缺少受支持的端到端加密公钥，暂时无法发送安全消息。'
    return
  }

  sending.value = true
  try {
    const receiverPublicKey = await importPublicKey(receiverPublicKeyValue)
    const senderPublicKey = await importPublicKey(ownPublicKey)
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

  sending.value = true
  try {
    const detail = await loadGroupDetail(currentGroupId.value)
    currentGroupDetail.value = detail
    if (!detail) throw new Error('群信息加载失败')
    if (!detail.members.length || !detail.members.every((member) => hasSupportedPublicKey(member.publicKey, member.publicKeyAlgorithm))) {
      throw new Error('当前群聊成员未全部满足安全会话要求，请在成员完成受支持密钥配置后重试。')
    }

    const members = detail.members.map((member) => ({
      userId: member.userId,
      publicKey: member.publicKey || '',
      publicKeyAlgorithm: member.publicKeyAlgorithm || ''
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
  const invalidMemberIds = payload.memberIds.filter(
    (memberId) => !secureFriends.value.some((friend) => friend.friendId === memberId)
  )
  if (invalidMemberIds.length > 0) {
    errorMessage.value = '所选成员中存在未满足安全会话要求的对象，请仅选择已完成受支持密钥配置的联系人。'
    return
  }

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
  const invalidMemberIds = memberIds.filter((memberId) => !secureFriends.value.some((friend) => friend.friendId === memberId))
  if (invalidMemberIds.length > 0) {
    errorMessage.value = '待邀请成员中存在未满足安全会话要求的对象，请先完成受支持密钥配置。'
    return
  }

  try {
    const { data } = await http.post(`/groups/${currentGroupId.value}/members`, { memberIds })
    const detail = data as GroupDetail
    currentGroupDetail.value = detail
    groupDetails.value = {
      ...groupDetails.value,
      [detail.id]: detail
    }
    showGroupInfo.value = false
    await loadGroups()
    if (currentGroupId.value && !secureGroups.value.some((group) => group.id === currentGroupId.value)) {
      conversationType.value = null
      currentGroupId.value = null
      currentGroupDetail.value = null
      messages.value = []
      errorMessage.value = '新增成员后，该群聊暂不满足当前安全会话要求。'
    }
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function handleLeaveGroup() {
  if (!currentGroupId.value) return
  if (!window.confirm('确认退出该加密群聊？退出后将不再收到该群的新消息。')) return
  try {
    const leavingGroupId = currentGroupId.value
    await http.delete(`/groups/${leavingGroupId}/members/me`)
    showGroupInfo.value = false
    conversationType.value = null
    currentGroupId.value = null
    currentGroupDetail.value = null
    messages.value = []
    if (leavingGroupId != null) {
      const { [leavingGroupId]: _removedGroup, ...remainingDetails } = groupDetails.value
      groupDetails.value = remainingDetails
    }
    await loadGroups()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

function isMine(message: { senderId: number }) {
  return message.senderId === authStore.user?.id
}

function messageKey(message: RenderMessage) {
  if (message.id != null) return `${message.sourceType || 'message'}-${message.id}`

  return [
    message.sourceType || 'message',
    message.groupId ?? 'direct',
    message.senderId,
    message.receiverId ?? 'broadcast',
    message.createdAt ?? 'no-time',
    message.senderCiphertext ?? '',
    message.receiverCiphertext ?? '',
    message.content
  ].join(':')
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
      if (groupMessage.senderId === authStore.user?.id) return
      if (conversationType.value === 'group' && currentGroupId.value === groupMessage.groupId) {
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

watch(
  messages,
  () => {
    scrollToBottom()
  },
  { deep: true }
)

watch([secureFriends, groupDetails], ([nextFriends, nextGroupDetails]) => {
  if (conversationType.value === 'friend' && currentFriendId.value && !nextFriends.some((friend) => friend.friendId === currentFriendId.value)) {
    conversationType.value = null
    currentFriendId.value = null
    currentGroupId.value = null
    currentGroupDetail.value = null
    messages.value = []
    errorMessage.value = '当前私聊目标已不再满足安全会话要求，会话已返回列表。'
  }

  if (conversationType.value === 'group' && currentGroupId.value) {
    const activeGroupDetail = nextGroupDetails[currentGroupId.value]
    if (activeGroupDetail && !isSecureGroupDetail(activeGroupDetail)) {
      conversationType.value = null
      currentFriendId.value = null
      currentGroupId.value = null
      currentGroupDetail.value = null
      messages.value = []
      errorMessage.value = '当前群聊成员配置已不再满足安全会话要求，会话已返回列表。'
    }
  }
})

function friendAvatarUrl(friend: FriendItem) {
  return friend.avatar || ''
}

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
  <div class="page-shell apple-page encrypted-page-shell" :class="skin.surfaceClass">
    <AppNav />

    <section class="encrypted-shell card" :class="{ 'mobile-show-chat': conversationType !== null }">
      <aside class="encrypted-sidebar">
        <div class="sidebar-hero">
          <div>
            <div class="encrypted-badge">E2EE</div>
            <p class="apple-label">Secure Messaging</p>
            <h2>加密通道</h2>
            <p class="sidebar-copy">仅保留端到端加密直聊与群聊能力。当前页面不会展示 AI 助手、收藏等普通消息入口。</p>
          </div>
          <span class="signal-pill" :class="{ online: socketConnected }">
            {{ socketConnected ? '安全链路在线' : '安全链路连接中' }}
          </span>
        </div>

        <div class="security-panel">
          <div>
            <p class="apple-label">Protection Status</p>
            <h3>{{ encryptedAvailabilityText }}</h3>
            <p class="security-caption">{{ sidebarTab === 'friend' ? friendSecureCapabilityText : groupSecureCapabilityText }}</p>
          </div>
          <small>{{ totalConversationCount }} 个加密会话可用</small>
        </div>

        <div class="sidebar-tabs">
          <button class="tab-btn" :class="{ active: sidebarTab === 'friend' }" @click="sidebarTab = 'friend'">
            加密私聊 ({{ secureFriends.length }})
          </button>
          <button class="tab-btn" :class="{ active: sidebarTab === 'group' }" @click="sidebarTab = 'group'">
            加密群聊 ({{ secureGroups.length }})
          </button>
        </div>

        <template v-if="sidebarTab === 'friend'">
          <div v-if="secureFriends.length === 0" class="empty-state">
            {{
              friends.length === 0
                ? '暂无可用联系人，请先建立好友关系并确保对方已启用受支持的加密公钥。'
                : '暂无满足安全会话条件的联系人。请确认好友已上传与当前协议匹配的端到端加密公钥。'
            }}
          </div>
          <button
            v-for="friend in secureFriends"
            :key="`f-${friend.friendId}`"
            class="conversation-item"
            :class="{ active: conversationType === 'friend' && currentFriendId === friend.friendId }"
            @click="selectFriend(friend.friendId)"
          >
            <img v-if="friendAvatarUrl(friend)" :src="friend.avatar" alt="avatar" class="conversation-avatar avatar-image" />
            <div v-else class="conversation-avatar">{{ friend.nickname.slice(0, 1).toUpperCase() }}</div>
            <div class="conversation-copy">
              <strong>{{ friend.nickname }}</strong>
              <span>{{ friend.signature || '对方暂未设置签名。' }}</span>
            </div>
          </button>
        </template>

        <template v-else>
          <button class="apple-button create-group-btn" type="button" @click="showCreateGroup = true">＋ 新建加密群聊</button>
          <div v-if="secureGroups.length === 0" class="empty-state">
            {{
              groups.length === 0
                ? '暂无加密群聊，点击上方按钮创建。'
                : '暂无满足安全会话条件的群聊。请确保群内所有成员都已配置受支持的端到端加密公钥。'
            }}
          </div>
          <button
            v-for="group in secureGroups"
            :key="`g-${group.id}`"
            class="conversation-item"
            :class="{ active: conversationType === 'group' && currentGroupId === group.id }"
            @click="selectGroup(group.id)"
          >
            <div class="conversation-avatar group">#</div>
            <div class="conversation-copy">
              <strong>{{ group.name }}</strong>
              <span>{{ group.memberCount }} 位成员</span>
            </div>
          </button>
        </template>
      </aside>

      <section class="encrypted-main-panel">
        <div class="chat-top">
          <button class="back-btn" type="button" @click="backToList" aria-label="返回会话列表">‹</button>
          <div class="chat-top-main">
            <p class="apple-label">{{ conversationType === 'group' ? 'Encrypted Group' : 'Encrypted Direct Message' }}</p>
            <h2>{{ conversationTitle }}</h2>
            <small class="muted">{{ currentConversationHint }}</small>
            <small v-if="authStore.user" class="muted">当前身份：{{ authStore.user.nickname || authStore.user.username }}</small>
          </div>
          <button v-if="conversationType === 'group'" class="apple-button secondary refresh-btn" @click="showGroupInfo = true">群信息</button>
          <button v-else class="apple-button secondary refresh-btn" @click="loadFriends">刷新联系人</button>
        </div>

        <div class="hero-note">
          <p class="apple-label">Encrypted Workspace</p>
          <p>所有消息内容均沿用现有端到端加密收发流程。该页面聚焦密钥、会话与成员安全，不承载普通消息中心功能。</p>
        </div>

        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <p v-if="!cryptoReady && !errorMessage" class="status-text error">当前环境不支持端到端加密，仅可查看已有加密会话。</p>

        <div v-if="!conversationType" class="empty-state chat-empty-state">
          <div class="empty-illustration">SEC</div>
          <strong>{{ totalConversationCount > 0 ? '进入加密工作区' : '暂无可用安全会话' }}</strong>
          <span>
            {{
              totalConversationCount > 0
                ? '从左侧选择加密私聊或群聊，开始发送仅限目标成员可解密的消息。'
                : '当前没有满足端到端加密条件的联系人或群聊。请先完成受支持公钥配置后再进入安全会话。'
            }}
          </span>
        </div>

        <div v-else ref="messageListRef" class="messages">
          <div v-for="message in messages" :key="messageKey(message)" class="message-row" :class="{ mine: isMine(message) }">
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
          <div class="composer-meta">
            <span>加密状态</span>
            <strong>{{ cryptoReady ? '已开启' : '不可用' }}</strong>
          </div>
          <input
            v-model="draft"
            class="apple-input"
            :disabled="!conversationType || sending || !cryptoReady"
            :placeholder="cryptoReady ? '输入消息，按回车发送加密内容' : '当前环境不支持发送加密消息'"
            @keyup.enter="sendMessage"
          />
          <button class="apple-button" :disabled="!conversationType || sending || !cryptoReady" @click="sendMessage">
            {{ sending ? '发送中...' : '发送' }}
          </button>
        </div>
      </section>
    </section>

    <CreateGroupModal
      :visible="showCreateGroup"
      :friends="secureFriends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))"
      @close="showCreateGroup = false"
      @submit="handleCreateGroup"
    />

    <GroupInfoPanel
      :visible="showGroupInfo && !!currentGroupDetail"
      :group-name="currentGroupDetail?.name || ''"
      :owner-id="currentGroupDetail?.ownerId || 0"
      :members="currentGroupDetail?.members || []"
      :my-user-id="authStore.user?.id ?? null"
      :friends="secureFriends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))"
      @close="showGroupInfo = false"
      @invite="handleInviteMembers"
      @leave="handleLeaveGroup"
    />
  </div>
</template>

<style scoped>
.encrypted-page-shell {
  position: relative;
  color: #e7f6ee;
}

.encrypted-page-shell::before {
  content: '';
  position: fixed;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(18, 102, 73, 0.26), transparent 34%),
    radial-gradient(circle at top right, rgba(55, 125, 91, 0.18), transparent 28%),
    radial-gradient(circle at bottom right, rgba(10, 41, 28, 0.34), transparent 40%),
    linear-gradient(180deg, rgba(2, 8, 6, 0.98), rgba(6, 14, 11, 1));
  pointer-events: none;
  z-index: 0;
}

.encrypted-shell {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 340px 1fr;
  min-height: 760px;
  overflow: hidden;
  background: rgba(4, 11, 8, 0.72);
  border: 1px solid rgba(110, 231, 183, 0.09);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.42);
}

.encrypted-sidebar {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 28px 22px;
  border-right: 1px solid rgba(110, 231, 183, 0.08);
  background: linear-gradient(180deg, rgba(8, 19, 14, 0.97), rgba(6, 13, 10, 0.95));
}

.sidebar-hero,
.security-panel,
.hero-note,
.chat-empty-state,
.conversation-item,
.composer-meta,
.message-item,
.tab-btn,
.create-group-btn {
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
}

.sidebar-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 20px;
  border-radius: 24px;
  background: linear-gradient(135deg, rgba(18, 52, 35, 0.9), rgba(7, 17, 13, 0.94));
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.1);
}

.encrypted-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 62px;
  padding: 8px 12px;
  border-radius: 999px;
  margin-bottom: 18px;
  background: rgba(34, 197, 94, 0.14);
  color: #86efac;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
}

.sidebar-hero h2 {
  margin: 8px 0 10px;
  color: #f0fdf4;
  font-size: 32px;
  letter-spacing: -0.03em;
}

.sidebar-copy {
  margin: 0;
  color: rgba(220, 252, 231, 0.78);
  line-height: 1.7;
}

.signal-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(8, 20, 14, 0.92);
  color: #96a69d;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.08);
}

.signal-pill.online {
  color: #86efac;
}

.security-panel {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
  border-radius: 22px;
  background: rgba(10, 20, 15, 0.86);
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.08);
}

.security-panel h3 {
  margin: 6px 0 0;
  color: #f0fdf4;
  font-size: 20px;
  letter-spacing: -0.02em;
}

.security-caption {
  margin: 8px 0 0;
  color: rgba(220, 252, 231, 0.68);
  font-size: 13px;
  line-height: 1.5;
}

.security-panel small {
  color: rgba(220, 252, 231, 0.64);
}

.sidebar-tabs,
.chat-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.sidebar-tabs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.tab-btn {
  padding: 10px 12px;
  border: none;
  border-radius: 14px;
  background: rgba(10, 20, 15, 0.82);
  color: rgba(231, 246, 238, 0.84);
  font-weight: 600;
  cursor: pointer;
  font-size: 13px;
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.06);
}

.tab-btn.active {
  background: rgba(20, 83, 45, 0.52);
  color: #bbf7d0;
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.18);
}

.create-group-btn {
  width: 100%;
  margin-bottom: 4px;
  font-size: 13px;
  padding: 10px 14px;
}

.conversation-item {
  display: grid;
  grid-template-columns: 48px 1fr;
  gap: 14px;
  align-items: center;
  padding: 14px;
  border: none;
  border-radius: 22px;
  background: rgba(9, 18, 13, 0.86);
  text-align: left;
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.05);
  color: inherit;
}

.conversation-item.active {
  background: rgba(20, 83, 45, 0.56);
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.16);
}

.conversation-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, rgba(19, 78, 54, 0.95) 0%, rgba(12, 58, 39, 0.94) 100%);
  color: #bbf7d0;
  font-weight: 700;
}

.avatar-image {
  object-fit: cover;
  display: block;
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.28);
}

.conversation-avatar.group {
  background: linear-gradient(180deg, rgba(82, 53, 9, 0.94) 0%, rgba(54, 35, 8, 0.95) 100%);
  color: #fde68a;
}

.conversation-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conversation-copy strong {
  color: #f0fdf4;
}

.conversation-copy span {
  color: rgba(220, 252, 231, 0.64);
}

.encrypted-main-panel {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 28px;
  background: linear-gradient(180deg, rgba(3, 9, 7, 0.82), rgba(6, 13, 10, 0.94));
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
  margin: 0;
  line-height: 1.1;
  font-size: 30px;
  color: #f0fdf4;
  letter-spacing: -0.03em;
}

.chat-top-main small {
  display: block;
  margin: 0;
  line-height: 1.4;
}

.chat-top-main small:first-of-type {
  color: rgba(220, 252, 231, 0.74);
  font-size: 13px;
  font-weight: 500;
}

.chat-top-main small:last-of-type {
  color: rgba(220, 252, 231, 0.58);
  font-size: 12px;
}

.hero-note {
  padding: 18px 20px;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(11, 29, 20, 0.92), rgba(6, 16, 12, 0.95));
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.08);
}

.hero-note p:last-child {
  margin: 8px 0 0;
  color: rgba(220, 252, 231, 0.78);
  line-height: 1.7;
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
  background: linear-gradient(180deg, rgba(8, 20, 14, 0.92), rgba(5, 13, 9, 0.96));
  text-align: center;
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.08);
}

.empty-illustration {
  width: 72px;
  height: 72px;
  border-radius: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(20, 83, 45, 0.42);
  color: #bbf7d0;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.chat-empty-state strong {
  font-size: 20px;
  color: #f0fdf4;
}

.chat-empty-state span,
.empty-state {
  color: rgba(220, 252, 231, 0.64);
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
  background: rgba(10, 20, 15, 0.94);
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.05);
}

.message-row.mine .message-item {
  background: linear-gradient(180deg, rgba(20, 83, 45, 0.9) 0%, rgba(16, 66, 36, 0.94) 100%);
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
  color: #f0fdf4;
}

.message-meta small,
.message-item p {
  color: rgba(220, 252, 231, 0.78);
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
  background: rgba(9, 18, 13, 0.92);
  box-shadow: inset 0 0 0 1px rgba(110, 231, 183, 0.06);
  text-align: left;
}

.composer-meta span {
  color: rgba(220, 252, 231, 0.6);
  font-size: 12px;
  line-height: 1.3;
}

.composer-meta strong {
  margin-top: 2px;
  color: #f0fdf4;
  font-size: 14px;
  line-height: 1.3;
}

.composer .apple-input {
  min-width: 0;
  background: rgba(9, 18, 13, 0.92);
  color: #f0fdf4;
  border: 1px solid rgba(110, 231, 183, 0.08);
}

.composer .apple-input::placeholder {
  color: rgba(220, 252, 231, 0.42);
}

.composer .apple-button {
  white-space: nowrap;
}

.back-btn {
  display: none;
}

@media (max-width: 980px) {
  .encrypted-shell {
    grid-template-columns: 1fr;
  }

  .encrypted-sidebar {
    border-right: none;
    border-bottom: 1px solid rgba(110, 231, 183, 0.08);
  }
}

@media (max-width: 768px) {
  .encrypted-shell {
    min-height: 0;
    display: block;
    overflow: visible;
  }

  .encrypted-sidebar {
    padding: 18px 14px;
    border-bottom: none;
  }

  .encrypted-main-panel {
    display: none;
    padding: 16px 14px calc(16px + env(safe-area-inset-bottom));
    min-height: calc(100vh - 140px);
  }

  .encrypted-shell.mobile-show-chat .encrypted-sidebar {
    display: none;
  }

  .encrypted-shell.mobile-show-chat .encrypted-main-panel {
    display: flex;
  }

  .sidebar-hero {
    padding: 16px;
  }

  .sidebar-hero h2,
  .chat-top-main h2 {
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
    background: rgba(255, 255, 255, 0.08);
    color: #f0fdf4;
    font-size: 22px;
    line-height: 1;
    cursor: pointer;
    flex-shrink: 0;
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

  .conversation-item {
    padding: 12px;
    border-radius: 16px;
  }

  .composer {
    grid-template-columns: 1fr auto;
    gap: 8px;
    position: sticky;
    bottom: 0;
    background: rgba(5, 12, 8, 0.92);
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
