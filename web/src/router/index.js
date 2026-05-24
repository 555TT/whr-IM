import { createRouter, createWebHistory } from 'vue-router';
import ChatView from '../views/ChatView.vue';
import FriendRequestsView from '../views/FriendRequestsView.vue';
import LoginView from '../views/LoginView.vue';
import ProfileView from '../views/ProfileView.vue';
import { useAuthStore } from '../stores/auth';
const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/', redirect: '/login' },
        { path: '/login', component: LoginView, meta: { guestOnly: true } },
        { path: '/chat', component: ChatView, meta: { requiresAuth: true } },
        { path: '/friend-requests', component: FriendRequestsView, meta: { requiresAuth: true } },
        { path: '/profile', component: ProfileView, meta: { requiresAuth: true } }
    ]
});
router.beforeEach(async (to) => {
    const authStore = useAuthStore();
    if (!authStore.bootstrapped) {
        await authStore.bootstrap();
    }
    if (to.meta.requiresAuth && !authStore.isLoggedIn) {
        return '/login';
    }
    if (to.meta.guestOnly && authStore.isLoggedIn) {
        return '/chat';
    }
    return true;
});
export default router;
