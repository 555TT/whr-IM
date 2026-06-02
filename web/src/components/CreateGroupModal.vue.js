import { computed, ref, watch } from 'vue';
const props = defineProps();
const emit = defineEmits();
const groupName = ref('');
const selected = ref(new Set());
const submitting = ref(false);
const errorMessage = ref('');
watch(() => props.visible, (v) => {
    if (v) {
        groupName.value = '';
        selected.value = new Set();
        submitting.value = false;
        errorMessage.value = '';
    }
});
const selectableFriends = computed(() => props.friends);
function toggle(id) {
    const next = new Set(selected.value);
    if (next.has(id))
        next.delete(id);
    else
        next.add(id);
    selected.value = next;
}
async function handleSubmit() {
    errorMessage.value = '';
    const name = groupName.value.trim();
    if (!name) {
        errorMessage.value = '请输入群名称';
        return;
    }
    if (selected.value.size === 0) {
        errorMessage.value = '至少选择一位好友加入群聊';
        return;
    }
    // 校验所选好友都已上传公钥(没公钥就没法加密)
    const missing = props.friends.filter((f) => selected.value.has(f.friendId) && !f.publicKey);
    if (missing.length > 0) {
        errorMessage.value = `成员 ${missing.map((m) => m.nickname).join('、')} 尚未生成密钥,请稍后再试`;
        return;
    }
    submitting.value = true;
    try {
        emit('submit', { name, memberIds: Array.from(selected.value) });
    }
    finally {
        submitting.value = false;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['modal-field']} */ ;
/** @type {__VLS_StyleScopedClasses['member-item']} */ ;
// CSS variable injection 
// CSS variable injection end 
if (__VLS_ctx.visible) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.visible))
                    return;
                __VLS_ctx.emit('close');
            } },
        ...{ class: "modal-mask" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "modal-card card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "modal-top" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "apple-label" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
        ...{ class: "modal-field" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
        ...{ class: "apple-input" },
        placeholder: "请输入群名(50 字以内)",
        maxlength: "50",
    });
    (__VLS_ctx.groupName);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "modal-field" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "member-list" },
    });
    if (__VLS_ctx.selectableFriends.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "empty-state small" },
        });
    }
    for (const [friend] of __VLS_getVForSourceType((__VLS_ctx.selectableFriends))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.visible))
                        return;
                    __VLS_ctx.toggle(friend.friendId);
                } },
            key: (friend.friendId),
            type: "button",
            ...{ class: "member-item" },
            ...{ class: ({ active: __VLS_ctx.selected.has(friend.friendId) }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "member-avatar" },
        });
        (friend.nickname.slice(0, 1).toUpperCase());
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "member-name" },
        });
        (friend.nickname);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "member-tick" },
        });
        (__VLS_ctx.selected.has(friend.friendId) ? '✓' : '');
    }
    if (__VLS_ctx.errorMessage) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "status-text error" },
        });
        (__VLS_ctx.errorMessage);
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "modal-actions" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.visible))
                    return;
                __VLS_ctx.emit('close');
            } },
        ...{ class: "apple-button secondary" },
        type: "button",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.handleSubmit) },
        ...{ class: "apple-button" },
        type: "button",
        disabled: (__VLS_ctx.submitting),
    });
    (__VLS_ctx.submitting ? '创建中...' : '创建群聊');
}
/** @type {__VLS_StyleScopedClasses['modal-mask']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-top']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-field']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-field']} */ ;
/** @type {__VLS_StyleScopedClasses['member-list']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['small']} */ ;
/** @type {__VLS_StyleScopedClasses['member-item']} */ ;
/** @type {__VLS_StyleScopedClasses['member-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['member-name']} */ ;
/** @type {__VLS_StyleScopedClasses['member-tick']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            emit: emit,
            groupName: groupName,
            selected: selected,
            submitting: submitting,
            errorMessage: errorMessage,
            selectableFriends: selectableFriends,
            toggle: toggle,
            handleSubmit: handleSubmit,
        };
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
