<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const displayName = computed(() => authStore.user?.nickname || authStore.user?.username || '未登录')

function logout() {
  authStore.clearSession()
  router.push('/login')
}
</script>

<template>
  <div class="nav-wrap">
    <nav class="nav card">
      <div class="brand">
        <span class="brand-dot"></span>
        <span>Easy Chat</span>
      </div>
      <div class="nav-links">
        <router-link to="/chat">聊天</router-link>
        <router-link to="/friend-requests">好友申请</router-link>
        <router-link to="/profile">个人资料</router-link>
      </div>
      <div class="nav-user">
        <span class="nav-user-name">{{ displayName }}</span>
        <button class="logout-btn" @click="logout">退出</button>
      </div>
    </nav>
  </div>
</template>

<style scoped>
.nav-wrap {
  margin-bottom: 28px;
}

.nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 22px;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.brand-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: linear-gradient(180deg, #47b1ff 0%, #0071e3 100%);
  box-shadow: 0 0 12px rgba(0, 113, 227, 0.35);
}

.nav-links {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.nav a {
  padding: 10px 14px;
  border-radius: 999px;
  color: #6e6e73;
  text-decoration: none;
  font-weight: 600;
  transition: background 0.2s ease, color 0.2s ease;
}

.nav a.router-link-active {
  background: rgba(0, 113, 227, 0.1);
  color: #0071e3;
}

.nav a:hover {
  background: rgba(29, 29, 31, 0.05);
  color: #1d1d1f;
}

.nav-user {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav-user-name {
  color: #6e6e73;
  font-size: 14px;
}

.logout-btn {
  border: none;
  border-radius: 999px;
  padding: 10px 14px;
  background: rgba(29, 29, 31, 0.06);
  color: #1d1d1f;
  cursor: pointer;
}

@media (max-width: 768px) {
  .nav-wrap {
    margin-bottom: 16px;
  }
  .nav {
    flex-wrap: wrap;
    gap: 10px;
    padding: 12px 14px;
  }
  .brand {
    font-size: 15px;
  }
  .nav-links {
    order: 3;
    width: 100%;
    justify-content: space-between;
    gap: 6px;
  }
  .nav a {
    padding: 8px 10px;
    font-size: 14px;
    flex: 1;
    text-align: center;
  }
  .nav-user-name {
    font-size: 12px;
    max-width: 90px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .logout-btn {
    padding: 8px 12px;
    font-size: 13px;
  }
}
</style>
