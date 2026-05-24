<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'

interface FriendRequestItem {
  id: number
  fromUserId: number
  toUserId: number
  message: string
  status: string
}

const requests = ref<FriendRequestItem[]>([])
const toUserId = ref<number | null>(null)
const note = ref('')
const feedback = ref('')
const errorMessage = ref('')

async function loadRequests() {
  const { data } = await http.get('/friend-requests/incoming')
  requests.value = data.filter((item: FriendRequestItem) => item.status === 'pending')
}

async function sendRequest() {
  if (!toUserId.value) return
  feedback.value = ''
  errorMessage.value = ''
  try {
    await http.post('/friend-requests', { toUserId: toUserId.value, message: note.value })
    feedback.value = '申请已发送'
    toUserId.value = null
    note.value = ''
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function accept(id: number) {
  feedback.value = ''
  errorMessage.value = ''
  try {
    await http.put(`/friend-requests/${id}/accept`)
    feedback.value = '已同意好友申请，请前往聊天页查看好友列表。'
    await loadRequests()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

async function reject(id: number) {
  feedback.value = ''
  errorMessage.value = ''
  try {
    await http.put(`/friend-requests/${id}/reject`)
    feedback.value = '已拒绝好友申请'
    await loadRequests()
  } catch (error) {
    errorMessage.value = (error as Error).message
  }
}

onMounted(loadRequests)
</script>

<template>
  <div class="page-shell page-layout">
    <AppNav />
    <div class="card content-card">
      <h2>好友申请</h2>
      <p v-if="feedback" class="success-text">{{ feedback }}</p>
      <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      <div class="send-row">
        <input v-model.number="toUserId" type="number" placeholder="目标用户 ID" />
        <input v-model="note" placeholder="申请附言" />
        <button @click="sendRequest">发送申请</button>
      </div>
      <div v-if="requests.length === 0" class="empty-state">暂无收到的待处理好友申请</div>
      <div v-for="item in requests" :key="item.id" class="request-row">
        <div>
          <strong>来自用户 {{ item.fromUserId }}</strong>
          <p>{{ item.message || '无附言' }}</p>
          <small>{{ item.status }}</small>
        </div>
        <div class="request-actions">
          <button @click="accept(item.id)">同意</button>
          <button class="ghost" @click="reject(item.id)">拒绝</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-layout {
  padding: 24px;
}

.content-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.send-row,
.request-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.send-row input {
  flex: 1;
}

input {
  padding: 12px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
}

button {
  padding: 10px 14px;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: white;
  cursor: pointer;
}

.ghost {
  background: #ef4444;
}

.request-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.empty-state {
  color: #6b7280;
}

.error-text {
  color: #dc2626;
}

.success-text {
  color: #16a34a;
}
</style>
