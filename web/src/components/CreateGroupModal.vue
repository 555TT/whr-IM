<script setup lang="ts">
import { computed, ref, watch } from 'vue'

interface FriendOption {
  friendId: number
  nickname: string
  publicKey?: string
}

const props = defineProps<{
  visible: boolean
  friends: FriendOption[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', payload: { name: string; memberIds: number[] }): void
}>()

const groupName = ref('')
const selected = ref<Set<number>>(new Set())
const submitting = ref(false)
const errorMessage = ref('')

watch(
  () => props.visible,
  (v) => {
    if (v) {
      groupName.value = ''
      selected.value = new Set()
      submitting.value = false
      errorMessage.value = ''
    }
  }
)

const selectableFriends = computed(() => props.friends)

function toggle(id: number) {
  const next = new Set(selected.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selected.value = next
}

async function handleSubmit() {
  errorMessage.value = ''
  const name = groupName.value.trim()
  if (!name) {
    errorMessage.value = '请输入群名称'
    return
  }
  if (selected.value.size === 0) {
    errorMessage.value = '至少选择一位好友加入群聊'
    return
  }
  // 校验所选好友都已上传公钥(没公钥就没法加密)
  const missing = props.friends.filter((f) => selected.value.has(f.friendId) && !f.publicKey)
  if (missing.length > 0) {
    errorMessage.value = `成员 ${missing.map((m) => m.nickname).join('、')} 尚未生成密钥,请稍后再试`
    return
  }
  submitting.value = true
  try {
    emit('submit', { name, memberIds: Array.from(selected.value) })
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div v-if="visible" class="modal-mask" @click.self="emit('close')">
    <div class="modal-card card">
      <div class="modal-top">
        <p class="apple-label">New Group</p>
        <h3>新建群聊</h3>
      </div>
      <label class="modal-field">
        <span>群名称</span>
        <input v-model="groupName" class="apple-input" placeholder="请输入群名(50 字以内)" maxlength="50" />
      </label>
      <div class="modal-field">
        <span>选择群成员</span>
        <div class="member-list">
          <div v-if="selectableFriends.length === 0" class="empty-state small">没有好友可邀请,请先去好友页添加好友。</div>
          <button
            v-for="friend in selectableFriends"
            :key="friend.friendId"
            type="button"
            class="member-item"
            :class="{ active: selected.has(friend.friendId) }"
            @click="toggle(friend.friendId)"
          >
            <span class="member-avatar">{{ friend.nickname.slice(0, 1).toUpperCase() }}</span>
            <span class="member-name">{{ friend.nickname }}</span>
            <span class="member-tick">{{ selected.has(friend.friendId) ? '✓' : '' }}</span>
          </button>
        </div>
      </div>
      <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
      <div class="modal-actions">
        <button class="apple-button secondary" type="button" @click="emit('close')">取消</button>
        <button class="apple-button" type="button" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? '创建中...' : '创建群聊' }}
        </button>
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
  padding: 26px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.modal-top h3 {
  margin: 4px 0 0;
  font-size: 22px;
  letter-spacing: -0.02em;
}
.modal-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
  color: #1d1d1f;
}
.modal-field span {
  font-weight: 500;
  color: #1d1d1f;
}
.member-list {
  max-height: 260px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 4px;
  border-radius: 14px;
  background: rgba(245, 245, 247, 0.6);
}
.member-item {
  display: grid;
  grid-template-columns: 32px 1fr 24px;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 10px;
  border: none;
  background: rgba(255, 255, 255, 0.9);
  cursor: pointer;
  text-align: left;
  font-size: 14px;
  box-shadow: inset 0 0 0 1px rgba(29, 29, 31, 0.04);
}
.member-item.active {
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
.member-tick {
  color: #0071e3;
  font-weight: 700;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
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
