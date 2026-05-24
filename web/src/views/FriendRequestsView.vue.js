import { onMounted, ref } from 'vue';
import AppNav from '../components/AppNav.vue';
import { http } from '../api/http';
const requests = ref([]);
const toUserId = ref(null);
const note = ref('');
const feedback = ref('');
const errorMessage = ref('');
async function loadRequests() {
    const { data } = await http.get('/friend-requests/incoming');
    requests.value = data.filter((item) => item.status === 'pending');
}
async function sendRequest() {
    if (!toUserId.value)
        return;
    feedback.value = '';
    errorMessage.value = '';
    try {
        await http.post('/friend-requests', { toUserId: toUserId.value, message: note.value });
        feedback.value = '申请已发送';
        toUserId.value = null;
        note.value = '';
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function accept(id) {
    feedback.value = '';
    errorMessage.value = '';
    try {
        await http.put(`/friend-requests/${id}/accept`);
        feedback.value = '已同意好友申请，请前往聊天页查看好友列表。';
        await loadRequests();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function reject(id) {
    feedback.value = '';
    errorMessage.value = '';
    try {
        await http.put(`/friend-requests/${id}/reject`);
        feedback.value = '已拒绝好友申请';
        await loadRequests();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
onMounted(loadRequests);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['send-row']} */ ;
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
    ...{ class: "card content-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
if (__VLS_ctx.feedback) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "success-text" },
    });
    (__VLS_ctx.feedback);
}
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "error-text" },
    });
    (__VLS_ctx.errorMessage);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "send-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "number",
    placeholder: "目标用户 ID",
});
(__VLS_ctx.toUserId);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "申请附言",
});
(__VLS_ctx.note);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.sendRequest) },
});
if (__VLS_ctx.requests.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
}
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.requests))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (item.id),
        ...{ class: "request-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (item.fromUserId);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
    (item.message || '无附言');
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
    (item.status);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "request-actions" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.accept(item.id);
            } },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.reject(item.id);
            } },
        ...{ class: "ghost" },
    });
}
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['page-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['content-card']} */ ;
/** @type {__VLS_StyleScopedClasses['success-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error-text']} */ ;
/** @type {__VLS_StyleScopedClasses['send-row']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['request-row']} */ ;
/** @type {__VLS_StyleScopedClasses['request-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['ghost']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            requests: requests,
            toUserId: toUserId,
            note: note,
            feedback: feedback,
            errorMessage: errorMessage,
            sendRequest: sendRequest,
            accept: accept,
            reject: reject,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
