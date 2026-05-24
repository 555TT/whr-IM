import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import AppNav from '../components/AppNav.vue';
import { http } from '../api/http';
import { useAuthStore } from '../stores/auth';
import { formatChatMessageTime } from '../utils/chat-time';
import { createChatSocket } from '../utils/websocket';
const authStore = useAuthStore();
const friends = ref([]);
const messages = ref([]);
const currentFriendId = ref(null);
const draft = ref('');
const errorMessage = ref('');
const socketConnected = ref(false);
const sending = ref(false);
const messageListRef = ref(null);
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
    await scrollToBottom();
}
async function selectFriend(friendId) {
    currentFriendId.value = friendId;
    await loadMessages();
}
async function sendMessage() {
    if (!currentFriendId.value || !draft.value.trim() || sending.value)
        return;
    errorMessage.value = '';
    sending.value = true;
    try {
        const { data } = await http.post('/messages', {
            receiverId: currentFriendId.value,
            content: draft.value.trim()
        });
        messages.value.push(data);
        draft.value = '';
        await scrollToBottom();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        sending.value = false;
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
    socket.onmessage = async (event) => {
        const payload = JSON.parse(event.data);
        if (payload.type === 'chat_message') {
            const chatMessage = payload.data;
            if (currentFriendId.value &&
                (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)) {
                messages.value.push(chatMessage);
                await scrollToBottom();
            }
        }
    };
}
async function scrollToBottom() {
    await nextTick();
    if (messageListRef.value) {
        messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
    }
}
watch(messages, () => {
    scrollToBottom();
}, { deep: true });
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
/** @type {__VLS_StyleScopedClasses['sidebar-top']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['mine']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-shell apple-page" },
});
/** @type {[typeof AppNav, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(AppNav, new AppNav({}));
const __VLS_1 = __VLS_0({}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "chat-shell card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.aside, __VLS_intrinsicElements.aside)({
    ...{ class: "sidebar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar-top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({
    ...{ class: "muted" },
});
(__VLS_ctx.socketConnected ? '在线同步中' : '等待连接');
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
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "friend-avatar" },
    });
    (friend.nickname.slice(0, 1).toUpperCase());
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "friend-copy" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (friend.nickname);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (friend.signature || '这个人很懒，还没写签名。');
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "chat-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chat-top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.currentFriend ? __VLS_ctx.currentFriend.nickname : '聊天窗口');
if (__VLS_ctx.authStore.user) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({
        ...{ class: "muted" },
    });
    (__VLS_ctx.authStore.user.nickname || __VLS_ctx.authStore.user.username);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.loadFriends) },
    ...{ class: "apple-button secondary" },
});
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text error" },
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
        ref: "messageListRef",
        ...{ class: "messages" },
    });
    /** @type {typeof __VLS_ctx.messageListRef} */ ;
    for (const [message, index] of __VLS_getVForSourceType((__VLS_ctx.messages))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (index),
            ...{ class: "message-row" },
            ...{ class: ({ mine: __VLS_ctx.isMine(message) }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "message-item" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "message-meta" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
        (__VLS_ctx.isMine(message) ? '我' : __VLS_ctx.currentFriend?.nickname || message.senderId);
        if (__VLS_ctx.formatChatMessageTime(message.createdAt)) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
            (__VLS_ctx.formatChatMessageTime(message.createdAt));
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
        (message.content);
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "composer" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onKeyup: (__VLS_ctx.sendMessage) },
    ...{ class: "apple-input" },
    disabled: (!__VLS_ctx.currentFriendId || __VLS_ctx.sending),
    placeholder: "输入消息",
});
(__VLS_ctx.draft);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.sendMessage) },
    ...{ class: "apple-button" },
    disabled: (!__VLS_ctx.currentFriendId || __VLS_ctx.sending),
});
(__VLS_ctx.sending ? '发送中...' : '发送');
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-page']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-top']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['messages']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            formatChatMessageTime: formatChatMessageTime,
            authStore: authStore,
            friends: friends,
            messages: messages,
            currentFriendId: currentFriendId,
            draft: draft,
            errorMessage: errorMessage,
            socketConnected: socketConnected,
            sending: sending,
            messageListRef: messageListRef,
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
