import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import AppNav from '../components/AppNav.vue';
import CreateGroupModal from '../components/CreateGroupModal.vue';
import GroupInfoPanel from '../components/GroupInfoPanel.vue';
import { http } from '../api/http';
import { resolveHomepageSkin } from '../constants/homepageSkins';
import { useAuthStore } from '../stores/auth';
import { formatChatMessageTime } from '../utils/chat-time';
import { buildEncryptedMessageDisplay, decryptMessage, E2EE_MESSAGE_ALGORITHM, encryptMessage, exportPrivateKey, exportPublicKey, generateKeyPair, importPrivateKey, importPublicKey, loadPrivateKey, savePrivateKey, selectMessagePayloadForUser } from '../utils/e2ee';
import { buildGroupMessageEnvelope, decryptGroupMessage } from '../utils/group-e2ee';
import { createChatSocket } from '../utils/websocket';
const authStore = useAuthStore();
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
const cryptoReady = ref(true); // 端到端加密是否就绪，未就绪时禁用发送
const socketConnected = ref(false);
const sending = ref(false);
const privateKey = ref(null);
const messageListRef = ref(null);
const showCreateGroup = ref(false);
const showGroupInfo = ref(false);
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
async function uploadOwnPublicKey(serializedPublicKey) {
    await http.put('/users/me/public-key', {
        publicKey: serializedPublicKey,
        algorithm: E2EE_MESSAGE_ALGORITHM
    });
    authStore.user = authStore.user
        ? {
            ...authStore.user,
            publicKey: serializedPublicKey,
            publicKeyAlgorithm: E2EE_MESSAGE_ALGORITHM
        }
        : null;
}
async function exportPublicKeyFromPrivateKey(serializedPrivateKey) {
    const importedPrivateKey = await importPrivateKey(serializedPrivateKey);
    const jwk = await window.crypto.subtle.exportKey('jwk', importedPrivateKey);
    const publicKey = await window.crypto.subtle.importKey('jwk', {
        kty: jwk.kty,
        n: jwk.n,
        e: jwk.e,
        alg: jwk.alg,
        ext: true,
        key_ops: ['encrypt']
    }, { name: 'RSA-OAEP', hash: 'SHA-256' }, true, ['encrypt']);
    return exportPublicKey(publicKey);
}
async function ensureOwnKeyPair() {
    if (!authStore.user?.id)
        return;
    const storedPrivateKey = loadPrivateKey(authStore.user.id);
    if (storedPrivateKey) {
        privateKey.value = await importPrivateKey(storedPrivateKey);
        if (authStore.user.publicKey && authStore.user.publicKeyAlgorithm === E2EE_MESSAGE_ALGORITHM) {
            return;
        }
        const serializedPublicKey = await exportPublicKeyFromPrivateKey(storedPrivateKey);
        await uploadOwnPublicKey(serializedPublicKey);
        return;
    }
    const keyPair = await generateKeyPair();
    const serializedPublicKey = await exportPublicKey(keyPair.publicKey);
    const serializedPrivateKey = await exportPrivateKey(keyPair.privateKey);
    savePrivateKey(authStore.user.id, serializedPrivateKey);
    privateKey.value = keyPair.privateKey;
    await uploadOwnPublicKey(serializedPublicKey);
}
async function toRenderMessage(message) {
    const payload = selectMessagePayloadForUser(message, authStore.user?.id);
    if (!privateKey.value) {
        return {
            ...message,
            ...buildEncryptedMessageDisplay('', payload)
        };
    }
    try {
        const content = await decryptMessage(privateKey.value, payload.ciphertext);
        return {
            ...message,
            ...buildEncryptedMessageDisplay(content, payload)
        };
    }
    catch {
        return {
            ...message,
            ...buildEncryptedMessageDisplay('', payload)
        };
    }
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
    const { data } = await http.get(`/messages?friendId=${currentFriendId.value}`);
    messages.value = await Promise.all(data.map(toRenderMessage));
    await scrollToBottom();
}
async function toGroupRenderMessage(message) {
    const base = {
        id: message.id,
        senderId: message.senderId,
        groupId: message.groupId,
        content: '***（已加密）',
        createdAt: message.createdAt
    };
    if (!privateKey.value)
        return base;
    try {
        base.content = await decryptGroupMessage(privateKey.value, message.keyCiphertext, message.contentCiphertext, message.contentIv);
    }
    catch {
        // 保持加密占位
    }
    return base;
}
async function loadGroupMessages() {
    if (!currentGroupId.value) {
        messages.value = [];
        return;
    }
    const { data } = await http.get(`/groups/${currentGroupId.value}/messages`);
    messages.value = await Promise.all(data.map(toGroupRenderMessage));
    await scrollToBottom();
}
async function loadGroupDetail(groupId) {
    const { data } = await http.get(`/groups/${groupId}`);
    currentGroupDetail.value = data;
}
async function selectFriend(friendId) {
    conversationType.value = 'friend';
    currentFriendId.value = friendId;
    currentGroupId.value = null;
    currentGroupDetail.value = null;
    await loadFriendMessages();
}
async function selectGroup(groupId) {
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
    if (!draft.value.trim() || sending.value)
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
    if (!privateKey.value) {
        errorMessage.value = '当前设备没有可用私钥';
        return;
    }
    const friend = currentFriend.value;
    if (!friend?.publicKey) {
        errorMessage.value = '对方未启用端到端加密消息';
        return;
    }
    if (!authStore.user?.publicKey) {
        errorMessage.value = '当前账号公钥不可用';
        return;
    }
    sending.value = true;
    try {
        const receiverPublicKey = await importPublicKey(friend.publicKey);
        const senderPublicKey = await importPublicKey(authStore.user.publicKey);
        const content = draft.value.trim();
        const receiverEncrypted = await encryptMessage(receiverPublicKey, content);
        const senderEncrypted = await encryptMessage(senderPublicKey, content);
        const { data } = await http.post('/messages', {
            receiverId: currentFriendId.value,
            senderCiphertext: senderEncrypted.ciphertext,
            senderAlgorithm: senderEncrypted.algorithm,
            receiverCiphertext: receiverEncrypted.ciphertext,
            receiverAlgorithm: receiverEncrypted.algorithm
        });
        messages.value.push(await toRenderMessage(data));
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
    if (!privateKey.value) {
        errorMessage.value = '当前设备没有可用私钥';
        return;
    }
    // 发消息前刷新一次成员列表,避免缺漏新成员的密钥包裹
    sending.value = true;
    try {
        await loadGroupDetail(currentGroupId.value);
        const detail = currentGroupDetail.value;
        if (!detail)
            throw new Error('群信息加载失败');
        const members = detail.members.map((m) => ({
            userId: m.userId,
            publicKey: m.publicKey || '',
            publicKeyAlgorithm: m.publicKeyAlgorithm || ''
        }));
        const content = draft.value.trim();
        const envelope = await buildGroupMessageEnvelope(content, members);
        const { data } = await http.post(`/groups/${currentGroupId.value}/messages`, {
            contentCiphertext: envelope.contentCiphertext,
            contentIv: envelope.contentIv,
            contentAlgorithm: envelope.contentAlgorithm,
            memberKeys: envelope.memberKeys
        });
        messages.value.push(await toGroupRenderMessage(data));
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
        if (payload.type === 'chat_message') {
            const chatMessage = payload.data;
            if (conversationType.value === 'friend' &&
                currentFriendId.value &&
                (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)) {
                messages.value.push(await toRenderMessage(chatMessage));
                await scrollToBottom();
            }
        }
        else if (payload.type === 'group_message') {
            const groupMessage = payload.data;
            // 自己发的消息已经在 sendGroupMessage 里 push 过,避免重复
            if (groupMessage.senderId === authStore.user?.id)
                return;
            if (conversationType.value === 'group' &&
                currentGroupId.value === groupMessage.groupId) {
                messages.value.push(await toGroupRenderMessage(groupMessage));
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
function backToList() {
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
    // 加密初始化失败不应阻断好友列表与历史消息的展示，因此独立 try
    try {
        await ensureOwnKeyPair();
    }
    catch (error) {
        cryptoReady.value = false;
        errorMessage.value = error.message;
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
if (__VLS_ctx.conversationType === 'group') {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.conversationType === 'group'))
                    return;
                __VLS_ctx.showGroupInfo = true;
            } },
        ...{ class: "apple-button secondary refresh-btn" },
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.loadFriends) },
        ...{ class: "apple-button secondary refresh-btn" },
    });
}
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text error" },
    });
    (__VLS_ctx.errorMessage);
}
if (!__VLS_ctx.cryptoReady && !__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text error" },
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
(__VLS_ctx.cryptoReady ? '已开启' : '不可用');
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onKeyup: (__VLS_ctx.sendMessage) },
    ...{ class: "apple-input" },
    disabled: (!__VLS_ctx.conversationType || __VLS_ctx.sending || !__VLS_ctx.cryptoReady),
    placeholder: (__VLS_ctx.cryptoReady ? '输入消息，按回车发送' : '当前环境不支持发送加密消息'),
});
(__VLS_ctx.draft);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.sendMessage) },
    ...{ class: "apple-button" },
    disabled: (!__VLS_ctx.conversationType || __VLS_ctx.sending || !__VLS_ctx.cryptoReady),
});
(__VLS_ctx.sending ? '发送中...' : '发送');
/** @type {[typeof CreateGroupModal, ]} */ ;
// @ts-ignore
const __VLS_3 = __VLS_asFunctionalComponent(CreateGroupModal, new CreateGroupModal({
    ...{ 'onClose': {} },
    ...{ 'onSubmit': {} },
    visible: (__VLS_ctx.showCreateGroup),
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))),
}));
const __VLS_4 = __VLS_3({
    ...{ 'onClose': {} },
    ...{ 'onSubmit': {} },
    visible: (__VLS_ctx.showCreateGroup),
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))),
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
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))),
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
    friends: (__VLS_ctx.friends.map((f) => ({ friendId: f.friendId, nickname: f.nickname, publicKey: f.publicKey }))),
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
/** @type {__VLS_StyleScopedClasses['sidebar-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['tab-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['tab-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['friend-item']} */ ;
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
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['refresh-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
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
            cryptoReady: cryptoReady,
            socketConnected: socketConnected,
            sending: sending,
            messageListRef: messageListRef,
            showCreateGroup: showCreateGroup,
            showGroupInfo: showGroupInfo,
            currentFriend: currentFriend,
            totalConversationCount: totalConversationCount,
            currentConversationHint: currentConversationHint,
            conversationTitle: conversationTitle,
            groupMemberDisplayName: groupMemberDisplayName,
            loadFriends: loadFriends,
            selectFriend: selectFriend,
            selectGroup: selectGroup,
            sendMessage: sendMessage,
            handleCreateGroup: handleCreateGroup,
            handleInviteMembers: handleInviteMembers,
            handleLeaveGroup: handleLeaveGroup,
            isMine: isMine,
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
