<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import { resolveHomepageSkin } from '../constants/homepageSkins'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const displayName = computed(() => authStore.user?.nickname || authStore.user?.username || '未登录')
const navSkin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))

function logout() {
  authStore.clearSession()
  router.push('/login')
}
</script>

<template>
  <div class="nav-wrap">
    <nav class="nav card" :class="[navSkin.surfaceClass, navSkin.accentClass]">
      <div class="brand">
        <span class="brand-badge">IM</span>
        <div class="brand-copy">
          <strong>Easy Chat IM</strong>
          <small>即时通讯 / 实时会话 / 群组消息</small>
        </div>
      </div>
      <div class="nav-links">
        <router-link to="/chat">消息中心</router-link>
        <router-link to="/moments">朋友圈</router-link>
        <router-link to="/contacts">通讯录</router-link>
        <router-link to="/profile">我的名片</router-link>
        <router-link to="/decorations">个性装扮</router-link>
      </div>
      <div class="nav-user">
        <span class="nav-user-status">● 已登录</span>
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
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
  color: #fff;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brand-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
  height: 42px;
  padding: 0 12px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.18);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.22);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.brand-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.brand-copy strong {
  font-size: 17px;
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.brand-copy small {
  color: rgba(255, 255, 255, 0.78);
  font-size: 12px;
  white-space: nowrap;
}

.nav-links {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.nav a {
  padding: 10px 14px;
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.86);
  text-decoration: none;
  font-weight: 600;
  transition: background 0.2s ease, color 0.2s ease;
}

.nav a.router-link-active {
  background: var(--skin-accent-soft, rgba(0, 113, 227, 0.1));
  color: #fff;
}

.nav a:hover {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

.nav-user {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.nav-user-status {
  color: #92ffb0;
  font-size: 12px;
  font-weight: 700;
}

.nav-user-name {
  color: rgba(255, 255, 255, 0.86);
  font-size: 14px;
}

.logout-btn {
  border: none;
  border-radius: 999px;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
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
    width: 100%;
  }

  .brand-copy strong {
    font-size: 15px;
  }

  .brand-copy small {
    font-size: 11px;
    white-space: normal;
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
