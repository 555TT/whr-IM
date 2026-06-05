import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import AppNav from '../components/AppNav.vue';
import CreateGroupModal from '../components/CreateGroupModal.vue';
import GroupInfoPanel from '../components/GroupInfoPanel.vue';
import { http } from '../api/http';
import { resolveHomepageSkin } from '../constants/homepageSkins';
import { useAuthStore } from '../stores/auth';
import { formatChatMessageTime } from '../utils/chat-time';
import { createChatSocket } from '../utils/websocket';
const authStore = useAuthStore();
const router = useRouter();
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin));
const friends = ref([]);
const groups = ref([]);
const messages = ref([]);
const conversationType = ref(null);
const currentFriendId = ref(null);
const currentGroupId = ref(null);
const currentGroupDetail = ref(null);
const sidebarTab = ref('friend');
const draft = ref('');
const errorMessage = ref('');
const socketConnected = ref(false);
const sending = ref(false);
const messageListRef = ref(null);
const showCreateGroup = ref(false);
const showGroupInfo = ref(false);
const selectionMode = ref(false);
const selectedMessageKeys = ref([]);
let socket = null;
const currentFriend = computed(() => friends.value.find((item) => item.friendId === currentFriendId.value) || null);
const totalConversationCount = computed(() => friends.value.length + groups.value.length);
const currentConversationHint = computed(() => {
    if (conversationType.value === 'friend')
        return currentFriend.value?.signature || '单聊会话已开启';
    if (conversationType.value === 'group')
        return currentGroupDetail.value ? `${currentGroupDetail.value.members.length} 位成员参与会话` : '群组会话已开启';
    return '选择联系人后可开始发送实时消息';
});
const canSendMessage = computed(() => !!conversationType.value && !selectionMode.value && !sending.value);
const composerStatusText = computed(() => {
    if (!conversationType.value)
        return '待选择会话';
    if (selectionMode.value)
        return '选择模式中';
    return '可发送';
});
const headerActionLabel = computed(() => {
    if (conversationType.value === 'group')
        return '群信息';
    if (conversationType.value === 'friend')
        return '返回消息中心';
    return sidebarTab.value === 'group' ? '刷新群聊' : '刷新好友';
});
async function handleHeaderAction() {
    errorMessage.value = '';
    if (conversationType.value === 'group') {
        showGroupInfo.value = true;
        return;
    }
    if (conversationType.value === 'friend') {
        backToList();
        return;
    }
    try {
        if (sidebarTab.value === 'group') {
            await loadGroups();
            return;
        }
        await loadFriends();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
// 当前在聊会话(群或好友)的展示标题
const conversationTitle = computed(() => {
    if (conversationType.value === 'friend')
        return currentFriend.value?.nickname || '聊天窗口';
    if (conversationType.value === 'group')
        return currentGroupDetail.value?.name || '群聊';
    return '聊天窗口';
});
// 群消息发送者显示名查找
function groupMemberDisplayName(senderId) {
    if (!currentGroupDetail.value)
        return String(senderId);
    const m = currentGroupDetail.value.members.find((mm) => mm.userId === senderId);
    return m ? m.nickname || m.username : String(senderId);
}
function buildRenderKey(message) {
    if (message.id) {
        return `${message.sourceType || 'friend'}-${message.id}`;
    }
    return [
        message.sourceType || 'friend',
        message.senderId,
        message.receiverId ?? 'na',
        message.groupId ?? 'na',
        message.createdAt ?? 'na',
        message.content
    ].join('-');
}
function toRenderMessage(message) {
    const renderMessage = {
        ...message,
        sourceType: 'friend'
    };
    return {
        ...renderMessage,
        renderKey: buildRenderKey(renderMessage)
    };
}
async function loadFriends() {
    const { data } = await http.get('/friends');
    friends.value = data;
}
async function loadGroups() {
    const { data } = await http.get('/groups');
    groups.value = data;
}
async function loadFriendMessages() {
    if (!currentFriendId.value) {
        messages.value = [];
        return;
    }
    const { data } = await http.get(`/normal-messages?friendId=${currentFriendId.value}`);
    messages.value = data.map(toRenderMessage);
    await scrollToBottom();
}
function toGroupRenderMessage(message) {
    const renderMessage = {
        ...message,
        sourceType: 'group'
    };
    return {
        ...renderMessage,
        renderKey: buildRenderKey(renderMessage)
    };
}
async function loadGroupMessages() {
    if (!currentGroupId.value) {
        messages.value = [];
        return;
    }
    const { data } = await http.get(`/groups/${currentGroupId.value}/normal-messages`);
    messages.value = data.map(toGroupRenderMessage);
    await scrollToBottom();
}
async function loadGroupDetail(groupId) {
    const { data } = await http.get(`/groups/${groupId}`);
    currentGroupDetail.value = data;
}
async function selectFriend(friendId) {
    clearSelection();
    conversationType.value = 'friend';
    currentFriendId.value = friendId;
    currentGroupId.value = null;
    currentGroupDetail.value = null;
    await loadFriendMessages();
}
async function selectGroup(groupId) {
    clearSelection();
    conversationType.value = 'group';
    currentGroupId.value = groupId;
    currentFriendId.value = null;
    try {
        await loadGroupDetail(groupId);
        await loadGroupMessages();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function sendMessage() {
    if (!canSendMessage.value || !draft.value.trim())
        return;
    errorMessage.value = '';
    if (conversationType.value === 'friend') {
        await sendFriendMessage();
    }
    else if (conversationType.value === 'group') {
        await sendGroupMessage();
    }
}
async function sendFriendMessage() {
    if (!currentFriendId.value)
        return;
    sending.value = true;
    try {
        const content = draft.value.trim();
        const { data } = await http.post('/normal-messages', {
            receiverId: currentFriendId.value,
            content
        });
        messages.value.push(toRenderMessage(data));
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
async function sendGroupMessage() {
    if (!currentGroupId.value)
        return;
    sending.value = true;
    try {
        const content = draft.value.trim();
        const { data } = await http.post(`/groups/${currentGroupId.value}/normal-messages`, {
            content
        });
        messages.value.push(toGroupRenderMessage(data));
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
async function handleCreateGroup(payload) {
    errorMessage.value = '';
    try {
        const { data } = await http.post('/groups', payload);
        showCreateGroup.value = false;
        await loadGroups();
        await selectGroup(data.id);
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function handleInviteMembers(memberIds) {
    if (!currentGroupId.value)
        return;
    errorMessage.value = '';
    try {
        const { data } = await http.post(`/groups/${currentGroupId.value}/members`, { memberIds });
        currentGroupDetail.value = data;
        showGroupInfo.value = false;
        await loadGroups();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function handleLeaveGroup() {
    if (!currentGroupId.value)
        return;
    if (!window.confirm('确认退出该群?退出后将不再收到该群的新消息。'))
        return;
    try {
        await http.delete(`/groups/${currentGroupId.value}/members/me`);
        showGroupInfo.value = false;
        conversationType.value = null;
        currentGroupId.value = null;
        currentGroupDetail.value = null;
        messages.value = [];
        await loadGroups();
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
    socket.onmessage = async (event) => {
        const payload = JSON.parse(event.data);
        if (payload.type === 'normal_chat_message') {
            const chatMessage = payload.data;
            if (conversationType.value === 'friend' &&
                currentFriendId.value &&
                (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)) {
                messages.value.push(toRenderMessage(chatMessage));
                await scrollToBottom();
            }
        }
        else if (payload.type === 'normal_group_message') {
            const groupMessage = payload.data;
            if (groupMessage.senderId === authStore.user?.id)
                return;
            if (conversationType.value === 'group' &&
                currentGroupId.value === groupMessage.groupId) {
                messages.value.push(toGroupRenderMessage(groupMessage));
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
function friendAvatarUrl(friend) {
    return friend.avatar || '';
}
function messageKey(message) {
    return message.renderKey;
}
function isSelected(message) {
    return selectedMessageKeys.value.includes(messageKey(message));
}
function openAIChat() {
    router.push('/ai-chat');
}
function openEncryptedChat() {
    router.push('/encrypted-chat');
}
function openFavorites() {
    router.push('/favorites');
}
function enterSelectionMode(message) {
    selectionMode.value = true;
    toggleSelection(message);
}
function toggleSelection(message) {
    const key = messageKey(message);
    if (selectedMessageKeys.value.includes(key)) {
        selectedMessageKeys.value = selectedMessageKeys.value.filter((item) => item !== key);
        if (selectedMessageKeys.value.length === 0) {
            selectionMode.value = false;
        }
        return;
    }
    selectedMessageKeys.value = [...selectedMessageKeys.value, key];
}
async function favoriteSelectedMessages() {
    if (selectedMessageKeys.value.length === 0)
        return;
    errorMessage.value = '';
    const selected = messages.value.filter((message) => isSelected(message) && message.id && message.sourceType);
    try {
        await http.post('/favorites', {
            items: selected.map((message) => ({ sourceType: message.sourceType, sourceMessageId: message.id, content: message.content }))
        });
        selectionMode.value = false;
        selectedMessageKeys.value = [];
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
function clearSelection() {
    selectionMode.value = false;
    selectedMessageKeys.value = [];
}
function backToList() {
    clearSelection();
    conversationType.value = null;
    currentFriendId.value = null;
    currentGroupId.value = null;
    currentGroupDetail.value = null;
    messages.value = [];
}
onMounted(async () => {
    try {
        await authStore.bootstrap();
    }
    catch (error) {
        errorMessage.value = error.message;
        return;
    }
    try {
        await loadFriends();
        await loadGroups();
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
/** @type {__VLS_StyleScopedClasses['chat-theme-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-banner']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-pill']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['encrypted-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['encrypted-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-top']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-top']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['tab-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['active']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['mine']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['mobile-show-chat']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-top']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-banner']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-banner']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top']} */ ;
/** @type {__VLS_StyleScopedClasses['back-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['messages']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-shell apple-page chat-theme-shell" },
    ...{ class: (__VLS_ctx.skin.surfaceClass) },
});
/** @type {[typeof AppNav, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(AppNav, new AppNav({}));
const __VLS_1 = __VLS_0({}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "chat-shell card" },
    ...{ class: ({ 'mobile-show-chat': __VLS_ctx.conversationType !== null }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.aside, __VLS_intrinsicElements.aside)({
    ...{ class: "sidebar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar-banner" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "sidebar-banner-copy" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "signal-pill" },
    ...{ class: ({ online: __VLS_ctx.socketConnected }) },
});
(__VLS_ctx.socketConnected ? '消息通道已连接' : '消息通道连接中');
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar-top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({
    ...{ class: "muted" },
});
(__VLS_ctx.totalConversationCount);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.openAIChat) },
    ...{ class: "ai-entry-card" },
    type: "button",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "ai-entry-arrow" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.openEncryptedChat) },
    ...{ class: "encrypted-entry-card" },
    type: "button",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "encrypted-entry-arrow" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.openFavorites) },
    ...{ class: "favorite-entry-card" },
    type: "button",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "favorite-entry-arrow" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar-tabs" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.sidebarTab = 'friend';
        } },
    ...{ class: "tab-btn" },
    ...{ class: ({ active: __VLS_ctx.sidebarTab === 'friend' }) },
});
(__VLS_ctx.friends.length);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.sidebarTab = 'group';
        } },
    ...{ class: "tab-btn" },
    ...{ class: ({ active: __VLS_ctx.sidebarTab === 'group' }) },
});
(__VLS_ctx.groups.length);
if (__VLS_ctx.sidebarTab === 'friend') {
    if (__VLS_ctx.friends.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "empty-state" },
        });
    }
    for (const [friend] of __VLS_getVForSourceType((__VLS_ctx.friends))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.sidebarTab === 'friend'))
                        return;
                    __VLS_ctx.selectFriend(friend.friendId);
                } },
            key: ('f' + friend.friendId),
            ...{ class: "friend-item" },
            ...{ class: ({ active: __VLS_ctx.conversationType === 'friend' && __VLS_ctx.currentFriendId === friend.friendId }) },
        });
        if (__VLS_ctx.friendAvatarUrl(friend)) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
                src: (friend.avatar),
                alt: "avatar",
                ...{ class: "friend-avatar avatar-image" },
            });
        }
        else {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "friend-avatar" },
            });
            (friend.nickname.slice(0, 1).toUpperCase());
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "friend-copy" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
        (friend.nickname);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (friend.signature || '这个人很懒，还没写签名。');
    }
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.sidebarTab === 'friend'))
                    return;
                __VLS_ctx.showCreateGroup = true;
            } },
        ...{ class: "apple-button create-group-btn" },
        type: "button",
    });
    if (__VLS_ctx.groups.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "empty-state" },
        });
    }
    for (const [group] of __VLS_getVForSourceType((__VLS_ctx.groups))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.sidebarTab === 'friend'))
                        return;
                    __VLS_ctx.selectGroup(group.id);
                } },
            key: ('g' + group.id),
            ...{ class: "friend-item" },
            ...{ class: ({ active: __VLS_ctx.conversationType === 'group' && __VLS_ctx.currentGroupId === group.id }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "friend-avatar group" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "friend-copy" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
        (group.name);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (group.memberCount);
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "chat-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chat-top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.backToList) },
    ...{ class: "back-btn" },
    type: "button",
    'aria-label': "返回会话列表",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chat-top-main" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
(__VLS_ctx.conversationType === 'group' ? 'Group Chat' : 'Direct Message');
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.conversationTitle);
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({
    ...{ class: "muted" },
});
(__VLS_ctx.currentConversationHint);
if (__VLS_ctx.authStore.user) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({
        ...{ class: "muted" },
    });
    (__VLS_ctx.authStore.user.nickname || __VLS_ctx.authStore.user.username);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.handleHeaderAction) },
    ...{ class: "apple-button secondary refresh-btn" },
    type: "button",
});
(__VLS_ctx.headerActionLabel);
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text error" },
    });
    (__VLS_ctx.errorMessage);
}
if (__VLS_ctx.selectionMode) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "selection-toolbar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.selectedMessageKeys.length);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "selection-actions" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.clearSelection) },
        ...{ class: "apple-button secondary" },
        type: "button",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.favoriteSelectedMessages) },
        ...{ class: "apple-button" },
        type: "button",
    });
}
if (!__VLS_ctx.conversationType) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state chat-empty-state" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-illustration" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ref: "messageListRef",
        ...{ class: "messages" },
    });
    /** @type {typeof __VLS_ctx.messageListRef} */ ;
    for (const [message] of __VLS_getVForSourceType((__VLS_ctx.messages))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onContextmenu: (...[$event]) => {
                    if (!!(!__VLS_ctx.conversationType))
                        return;
                    __VLS_ctx.enterSelectionMode(message);
                } },
            key: (message.renderKey),
            ...{ class: "message-row" },
            ...{ class: ({ mine: __VLS_ctx.isMine(message) }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!!(!__VLS_ctx.conversationType))
                        return;
                    __VLS_ctx.selectionMode ? __VLS_ctx.toggleSelection(message) : undefined;
                } },
            ...{ class: "message-item" },
            ...{ class: ({ selected: __VLS_ctx.isSelected(message) }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "message-meta" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
        (__VLS_ctx.isMine(message)
            ? '我'
            : __VLS_ctx.conversationType === 'group'
                ? __VLS_ctx.groupMemberDisplayName(message.senderId)
                : __VLS_ctx.currentFriend?.nickname || message.senderId);
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "composer-meta" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.composerStatusText);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onKeyup: (__VLS_ctx.sendMessage) },
    ...{ class: "apple-input" },
    disabled: (!__VLS_ctx.canSendMessage),
    placeholder: "输入消息，按回车发送",
});
(__VLS_ctx.draft);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.sendMessage) },
    ...{ class: "apple-button" },
    disabled: (!__VLS_ctx.canSendMessage),
});
(__VLS_ctx.sending ? '发送中...' : '发送');
/** @type {[typeof CreateGroupModal, ]} */ ;
// @ts-ignore
const __VLS_3 = __VLS_asFunctionalComponent(CreateGroupModal, new CreateGroupModal({
    ...{ 'onClose': {} },
    ...{ 'onSubmit': {} },
    visible: (__VLS_ctx.showCreateGroup),
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname }))),
}));
const __VLS_4 = __VLS_3({
    ...{ 'onClose': {} },
    ...{ 'onSubmit': {} },
    visible: (__VLS_ctx.showCreateGroup),
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname }))),
}, ...__VLS_functionalComponentArgsRest(__VLS_3));
let __VLS_6;
let __VLS_7;
let __VLS_8;
const __VLS_9 = {
    onClose: (...[$event]) => {
        __VLS_ctx.showCreateGroup = false;
    }
};
const __VLS_10 = {
    onSubmit: (__VLS_ctx.handleCreateGroup)
};
var __VLS_5;
/** @type {[typeof GroupInfoPanel, ]} */ ;
// @ts-ignore
const __VLS_11 = __VLS_asFunctionalComponent(GroupInfoPanel, new GroupInfoPanel({
    ...{ 'onClose': {} },
    ...{ 'onInvite': {} },
    ...{ 'onLeave': {} },
    visible: (__VLS_ctx.showGroupInfo && !!__VLS_ctx.currentGroupDetail),
    groupName: (__VLS_ctx.currentGroupDetail?.name || ''),
    ownerId: (__VLS_ctx.currentGroupDetail?.ownerId || 0),
    members: (__VLS_ctx.currentGroupDetail?.members || []),
    myUserId: (__VLS_ctx.authStore.user?.id ?? null),
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname }))),
}));
const __VLS_12 = __VLS_11({
    ...{ 'onClose': {} },
    ...{ 'onInvite': {} },
    ...{ 'onLeave': {} },
    visible: (__VLS_ctx.showGroupInfo && !!__VLS_ctx.currentGroupDetail),
    groupName: (__VLS_ctx.currentGroupDetail?.name || ''),
    ownerId: (__VLS_ctx.currentGroupDetail?.ownerId || 0),
    members: (__VLS_ctx.currentGroupDetail?.members || []),
    myUserId: (__VLS_ctx.authStore.user?.id ?? null),
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname }))),
}, ...__VLS_functionalComponentArgsRest(__VLS_11));
let __VLS_14;
let __VLS_15;
let __VLS_16;
const __VLS_17 = {
    onClose: (...[$event]) => {
        __VLS_ctx.showGroupInfo = false;
    }
};
const __VLS_18 = {
    onInvite: (__VLS_ctx.handleInviteMembers)
};
const __VLS_19 = {
    onLeave: (__VLS_ctx.handleLeaveGroup)
};
var __VLS_13;
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-page']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-theme-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-banner']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-banner-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-pill']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-top']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-entry-arrow']} */ ;
/** @type {__VLS_StyleScopedClasses['encrypted-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['encrypted-entry-arrow']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-entry-card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-entry-arrow']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['tab-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['tab-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['avatar-image']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['create-group-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['group']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top']} */ ;
/** @type {__VLS_StyleScopedClasses['back-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-top-main']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['refresh-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['selection-toolbar']} */ ;
/** @type {__VLS_StyleScopedClasses['selection-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-illustration']} */ ;
/** @type {__VLS_StyleScopedClasses['messages']} */ ;
/** @type {__VLS_StyleScopedClasses['message-row']} */ ;
/** @type {__VLS_StyleScopedClasses['message-item']} */ ;
/** @type {__VLS_StyleScopedClasses['message-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            CreateGroupModal: CreateGroupModal,
            GroupInfoPanel: GroupInfoPanel,
            formatChatMessageTime: formatChatMessageTime,
            authStore: authStore,
            skin: skin,
            friends: friends,
            groups: groups,
            messages: messages,
            conversationType: conversationType,
            currentFriendId: currentFriendId,
            currentGroupId: currentGroupId,
            currentGroupDetail: currentGroupDetail,
            sidebarTab: sidebarTab,
            draft: draft,
            errorMessage: errorMessage,
            socketConnected: socketConnected,
            sending: sending,
            messageListRef: messageListRef,
            showCreateGroup: showCreateGroup,
            showGroupInfo: showGroupInfo,
            selectionMode: selectionMode,
            selectedMessageKeys: selectedMessageKeys,
            currentFriend: currentFriend,
            totalConversationCount: totalConversationCount,
            currentConversationHint: currentConversationHint,
            canSendMessage: canSendMessage,
            composerStatusText: composerStatusText,
            headerActionLabel: headerActionLabel,
            handleHeaderAction: handleHeaderAction,
            conversationTitle: conversationTitle,
            groupMemberDisplayName: groupMemberDisplayName,
            selectFriend: selectFriend,
            selectGroup: selectGroup,
            sendMessage: sendMessage,
            handleCreateGroup: handleCreateGroup,
            handleInviteMembers: handleInviteMembers,
            handleLeaveGroup: handleLeaveGroup,
            isMine: isMine,
            friendAvatarUrl: friendAvatarUrl,
            isSelected: isSelected,
            openAIChat: openAIChat,
            openEncryptedChat: openEncryptedChat,
            openFavorites: openFavorites,
            enterSelectionMode: enterSelectionMode,
            toggleSelection: toggleSelection,
            favoriteSelectedMessages: favoriteSelectedMessages,
            clearSelection: clearSelection,
            backToList: backToList,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
