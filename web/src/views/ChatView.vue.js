import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import AppNav from '../components/AppNav.vue';
import { http } from '../api/http';
import { useAuthStore } from '../stores/auth';
import { createChatSocket } from '../utils/websocket';
const authStore = useAuthStore();
const friends = ref([]);
const messages = ref([]);
const currentFriendId = ref(null);
const draft = ref('');
const errorMessage = ref('');
const socketConnected = ref(false);
let socket = null;
const currentFriend = computed(() => friends.value.find((item) => item.friendId === currentFriendId.value) || null);
async function loadFriends() {
    const { data } = await http.get('/friends');
    friends.value = data;
    if (!currentFriendId.value && friends.value.length > 0) {
        currentFriendId.value = friends.value[0].friendId;
        await loadMessages();
    }
    if (currentFriendId.value && !friends.value.some((item) => item.friendId === currentFriendId.value)) {
        currentFriendId.value = friends.value[0]?.friendId || null;
        await loadMessages();
    }
}
async function loadMessages() {
    if (!currentFriendId.value) {
        messages.value = [];
        return;
    }
    const { data } = await http.get(`/messages?friendId=${currentFriendId.value}`);
    messages.value = data;
}
async function selectFriend(friendId) {
    currentFriendId.value = friendId;
    await loadMessages();
}
async function sendMessage() {
    if (!currentFriendId.value || !draft.value.trim())
        return;
    errorMessage.value = '';
    try {
        const { data } = await http.post('/messages', {
            receiverId: currentFriendId.value,
            content: draft.value.trim()
        });
        messages.value.push(data);
        draft.value = '';
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
function isMine(message) {
    return message.senderId === authStore.user?.id;
}
function connectSocket() {
    if (!authStore.token)
        return;
    socket = createChatSocket(authStore.token);
    socket.onopen = () => {
        socketConnected.value = true;
    };
    socket.onclose = () => {
        socketConnected.value = false;
    };
    socket.onerror = () => {
        errorMessage.value = 'WebSocket 连接失败';
    };
    socket.onmessage = (event) => {
        const payload = JSON.parse(event.data);
        if (payload.type === 'chat_message') {
            const chatMessage = payload.data;
            if (currentFriendId.value &&
                (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)) {
                messages.value.push(chatMessage);
            }
        }
    };
}
onMounted(async () => {
    try {
        await authStore.bootstrap();
        await loadFriends();
        connectSocket();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
});
onBeforeUnmount(() => {
    socket?.close();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['mine']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-shell page-layout" },
});
/** @type {[typeof AppNav, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(AppNav, new AppNav({}));
const __VLS_1 = __VLS_0({}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chat-layout" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.aside, __VLS_intrinsicElements.aside)({
    ...{ class: "card sidebar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
(__VLS_ctx.socketConnected ? 'WS 已连接' : 'WS 未连接');
if (__VLS_ctx.friends.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
}
for (const [friend] of __VLS_getVForSourceType((__VLS_ctx.friends))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.selectFriend(friend.friendId);
            } },
        key: (friend.friendId),
        ...{ class: "friend-item" },
        ...{ class: ({ active: __VLS_ctx.currentFriendId === friend.friendId }) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (friend.nickname);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
    (friend.friendId);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (friend.signature || '这个人很懒，还没写签名。');
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "card chat-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chat-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
if (__VLS_ctx.currentFriend) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
    (__VLS_ctx.currentFriend.nickname);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.loadFriends) },
    ...{ class: "refresh-btn" },
});
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "error-text" },
    });
    (__VLS_ctx.errorMessage);
}
if (!__VLS_ctx.currentFriendId) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "messages" },
    });
    for (const [message, index] of __VLS_getVForSourceType((__VLS_ctx.messages))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (index),
            ...{ class: "message-row" },
            ...{ class: ({ mine: __VLS_ctx.isMine(message) }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "message-item" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
        (__VLS_ctx.isMine(message) ? '我' : __VLS_ctx.currentFriend?.nickname || message.senderId);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
        (message.content);
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "composer" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onKeyup: (__VLS_ctx.sendMessage) },
    disabled: (!__VLS_ctx.currentFriendId),
    placeholder: "输入消息",
});
(__VLS_ctx.draft);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.sendMessage) },
    disabled: (!__VLS_ctx.currentFriendId),
});
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['page-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-header']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-header']} */ ;
/** @type {__VLS_StyleScopedClasses['refresh-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['error-text']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['messages']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            friends: friends,
            messages: messages,
            currentFriendId: currentFriendId,
            draft: draft,
            errorMessage: errorMessage,
            socketConnected: socketConnected,
            currentFriend: currentFriend,
            loadFriends: loadFriends,
            selectFriend: selectFriend,
            sendMessage: sendMessage,
            isMine: isMine,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
