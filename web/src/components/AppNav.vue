<template>
  <nav class="nav card">
    <router-link to="/chat">聊天</router-link>
    <router-link to="/friend-requests">好友申请</router-link>
    <router-link to="/profile">个人资料</router-link>
  </nav>
</template>

<style scoped>
.nav {
  display: flex;
  gap: 16px;
  padding: 16px 20px;
  margin-bottom: 20px;
}

.nav a {
  color: #2563eb;
  text-decoration: none;
  font-weight: 600;
}
</style>
