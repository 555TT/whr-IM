<script setup lang="ts">
import { computed, ref, watch } from 'vue'

interface MemberItem {
  userId: number
  username: string
  nickname: string
  publicKey?: string
}

interface FriendOption {
  friendId: number
  nickname: string
  publicKey?: string
}

const props = defineProps<{
  visible: boolean
  groupName: string
  ownerId: number
  members: MemberItem[]
  myUserId: number | null
  friends: FriendOption[] // 我的好友(用于邀请)
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'invite', memberIds: number[]): void
  (e: 'leave'): void
}>()

const showInvite = ref(false)
const selected = ref<Set<number>>(new Set())
const errorMessage = ref('')
const submitting = ref(false)

watch(
  () => props.visible,
  (v) => {
    if (v) {
      showInvite.value = false
      selected.value = new Set()
      errorMessage.value = ''
      submitting.value = false
    }
  }
)

const memberIdSet = computed(() => new Set(props.members.map((m) => m.userId)))

const inviteCandidates = computed(() =>
  props.friends.filter((f) => !memberIdSet.value.has(f.friendId))
)

function toggle(id: number) {
  const next = new Set(selected.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selected.value = next
}

async function submitInvite() {
  errorMessage.value = ''
  if (selected.value.size === 0) {
    errorMessage.value = '请选择至少一位好友'
    return
  }
  const missing = inviteCandidates.value.filter(
    (f) => selected.value.has(f.friendId) && !f.publicKey
  )
  if (missing.length > 0) {
    errorMessage.value = `成员 ${missing.map((m) => m.nickname).join('、')} 尚未生成密钥,请稍后再试`
    return
  }
  submitting.value = true
  try {
    emit('invite', Array.from(selected.value))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div v-if="visible" class="modal-mask" @click.self="emit('close')">
    <div class="modal-card card">
      <div class="modal-top">
        <p class="apple-label">Group Info</p>
        <h3>{{ groupName }}</h3>
      </div>

      <div class="info-block">
        <p class="info-label">群成员 ({{ members.length }})</p>
        <ul class="member-list">
          <li v-for="m in members" :key="m.userId" class="member-row">
            <span class="member-avatar">{{ (m.nickname || m.username).slice(0, 1).toUpperCase() }}</span>
            <span class="member-name">{{ m.nickname || m.username }}</span>
            <span v-if="m.userId === ownerId" class="badge">群主</span>
            <span v-if="myUserId === m.userId" class="badge me">我</span>
          </li>
        </ul>
      </div>

      <div v-if="!showInvite" class="modal-actions">
        <button class="apple-button secondary" type="button" @click="showInvite = true">邀请好友</button>
        <button class="apple-button danger" type="button" @click="emit('leave')">退出群聊</button>
        <button class="apple-button" type="button" @click="emit('close')">关闭</button>
      </div>

      <div v-else class="invite-block">
        <p class="info-label">选择要邀请的好友</p>
        <div class="invite-list">
          <div v-if="inviteCandidates.length === 0" class="empty-state small">所有好友都已经在群里了。</div>
          <button
            v-for="f in inviteCandidates"
            :key="f.friendId"
            type="button"
            class="invite-item"
            :class="{ active: selected.has(f.friendId) }"
            @click="toggle(f.friendId)"
          >
            <span class="member-avatar">{{ f.nickname.slice(0, 1).toUpperCase() }}</span>
            <span class="member-name">{{ f.nickname }}</span>
            <span class="member-tick">{{ selected.has(f.friendId) ? '✓' : '' }}</span>
          </button>
        </div>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <div class="modal-actions">
          <button class="apple-button secondary" type="button" @click="showInvite = false">取消</button>
          <button class="apple-button" type="button" :disabled="submitting" @click="submitInvite">
            {{ submitting ? '邀请中...' : '确认邀请' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(29, 29, 31, 0.32);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.modal-card {
  width: min(420px, calc(100vw - 32px));
  max-height: calc(100vh - 64px);
  overflow-y: auto;
  padding: 26px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.modal-top h3 {
  margin: 4px 0 0;
  font-size: 22px;
  letter-spacing: -0.02em;
}
.info-label {
  margin: 0 0 8px;
  font-weight: 500;
  font-size: 14px;
}
.member-list,
.invite-list {
  list-style: none;
  margin: 0;
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: rgba(245, 245, 247, 0.6);
  border-radius: 14px;
  max-height: 260px;
  overflow-y: auto;
}
.member-row,
.invite-item {
  display: grid;
  grid-template-columns: 32px 1fr auto;
  gap: 10px;
  align-items: center;
  padding: 8px 12px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  box-shadow: inset 0 0 0 1px rgba(29, 29, 31, 0.04);
  border: none;
  text-align: left;
}
.invite-item.active {
  background: rgba(0, 113, 227, 0.12);
  box-shadow: inset 0 0 0 1px rgba(0, 113, 227, 0.18);
}
.member-avatar {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: linear-gradient(180deg, #eef5ff 0%, #d9e8ff 100%);
  color: #0071e3;
  font-weight: 700;
}
.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 8px;
  background: rgba(0, 113, 227, 0.12);
  color: #0071e3;
}
.badge.me {
  background: rgba(29, 29, 31, 0.06);
  color: #6e6e73;
}
.member-tick {
  color: #0071e3;
  font-weight: 700;
}
.modal-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}
.apple-button.danger {
  background: rgba(215, 0, 21, 0.1);
  color: #d70015;
  box-shadow: inset 0 0 0 1px rgba(215, 0, 21, 0.18);
}
.empty-state.small {
  padding: 12px;
  font-size: 13px;
}
.status-text.error {
  color: #d70015;
  font-size: 13px;
  margin: 0;
}
</style>
