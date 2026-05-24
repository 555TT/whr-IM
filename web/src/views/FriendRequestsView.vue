<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'

interface FriendRequestItem {
  id: number
  fromUserId: number
  fromUsername: string
  toUserId: number
  message: string
  status: string
}

const requests = ref<FriendRequestItem[]>([])
const toUsername = ref('')
const note = ref('')
const feedback = ref('')
const errorMessage = ref('')

async function loadRequests() {
  const { data } = await http.get('/friend-requests/incoming')
  requests.value = data.filter((item: FriendRequestItem) => item.status === 'pending')
}

async function sendRequest() {
  if (!toUsername.value.trim()) return
  feedback.value = ''
  errorMessage.value = ''
  try {
    await http.post('/friend-requests', { toUsername: toUsername.value.trim(), message: note.value })
    feedback.value = '申请已发送'
    toUsername.value = ''
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
  <div class="page-shell apple-page">
    <AppNav />
    <section class="requests-layout">
      <div class="card apple-panel request-form-card">
        <p class="apple-label">Add friend</p>
        <h1>发起好友申请</h1>
        <p class="muted">输入目标用户名，并附上一句简短说明。</p>
        <p v-if="feedback" class="status-text success">{{ feedback }}</p>
        <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
        <input v-model="toUsername" class="apple-input" placeholder="目标用户名" />
        <input v-model="note" class="apple-input" placeholder="申请附言" />
        <button class="apple-button" @click="sendRequest">发送申请</button>
      </div>

      <div class="card apple-panel request-list-card">
        <div class="list-head">
          <div>
            <p class="apple-label">Incoming</p>
            <h2>收到的好友申请</h2>
          </div>
        </div>
        <div v-if="requests.length === 0" class="empty-state">暂无收到的待处理好友申请</div>
        <div v-for="item in requests" :key="item.id" class="request-row">
          <div class="request-copy">
            <strong>来自 {{ item.fromUsername }}</strong>
            <p>{{ item.message || '无附言' }}</p>
          </div>
          <div class="request-actions">
            <button class="apple-button secondary" @click="accept(item.id)">同意</button>
            <button class="apple-button danger" @click="reject(item.id)">拒绝</button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.requests-layout {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 24px;
}

.request-form-card,
.request-list-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.request-form-card h1,
.request-list-card h2 {
  margin: 6px 0 0;
  font-size: 34px;
  letter-spacing: -0.03em;
}

.request-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 0;
  border-top: 1px solid rgba(29, 29, 31, 0.08);
}

.request-copy p {
  margin: 6px 0 0;
  color: #6e6e73;
}

.request-actions {
  display: flex;
  gap: 10px;
}

@media (max-width: 920px) {
  .requests-layout {
    grid-template-columns: 1fr;
  }

  .request-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
