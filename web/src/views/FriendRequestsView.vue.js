import { onMounted, ref } from 'vue';
import AppNav from '../components/AppNav.vue';
import { http } from '../api/http';
const requests = ref([]);
const toUsername = ref('');
const note = ref('');
const feedback = ref('');
const errorMessage = ref('');
async function loadRequests() {
    const { data } = await http.get('/friend-requests/incoming');
    requests.value = data.filter((item) => item.status === 'pending');
}
async function sendRequest() {
    if (!toUsername.value.trim())
        return;
    feedback.value = '';
    errorMessage.value = '';
    try {
        await http.post('/friend-requests', { toUsername: toUsername.value.trim(), message: note.value });
        feedback.value = '申请已发送';
        toUsername.value = '';
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
/** @type {__VLS_StyleScopedClasses['request-form-card']} */ ;
/** @type {__VLS_StyleScopedClasses['request-list-card']} */ ;
/** @type {__VLS_StyleScopedClasses['requests-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['request-row']} */ ;
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
    ...{ class: "requests-layout" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "card apple-panel request-form-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "muted" },
});
if (__VLS_ctx.feedback) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text success" },
    });
    (__VLS_ctx.feedback);
}
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text error" },
    });
    (__VLS_ctx.errorMessage);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ class: "apple-input" },
    placeholder: "目标用户名",
});
(__VLS_ctx.toUsername);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ class: "apple-input" },
    placeholder: "申请附言",
});
(__VLS_ctx.note);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.sendRequest) },
    ...{ class: "apple-button" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "card apple-panel request-list-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "list-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
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
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "request-copy" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (item.fromUserId);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
    (item.message || '无附言');
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "request-actions" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.accept(item.id);
            } },
        ...{ class: "apple-button secondary" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.reject(item.id);
            } },
        ...{ class: "apple-button danger" },
    });
}
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-page']} */ ;
/** @type {__VLS_StyleScopedClasses['requests-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['request-form-card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['success']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['request-list-card']} */ ;
/** @type {__VLS_StyleScopedClasses['list-head']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['request-row']} */ ;
/** @type {__VLS_StyleScopedClasses['request-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['request-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['danger']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            requests: requests,
            toUsername: toUsername,
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
