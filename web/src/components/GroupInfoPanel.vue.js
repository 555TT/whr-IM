import { computed, ref, watch } from 'vue';
const props = defineProps();
const emit = defineEmits();
const showInvite = ref(false);
const selected = ref(new Set());
const errorMessage = ref('');
const submitting = ref(false);
watch(() => props.visible, (v) => {
    if (v) {
        showInvite.value = false;
        selected.value = new Set();
        errorMessage.value = '';
        submitting.value = false;
    }
});
const memberIdSet = computed(() => new Set(props.members.map((m) => m.userId)));
const inviteCandidates = computed(() => props.friends.filter((f) => !memberIdSet.value.has(f.friendId)));
function toggle(id) {
    const next = new Set(selected.value);
    if (next.has(id))
        next.delete(id);
    else
        next.add(id);
    selected.value = next;
}
async function submitInvite() {
    errorMessage.value = '';
    if (selected.value.size === 0) {
        errorMessage.value = '请选择至少一位好友';
        return;
    }
    const missing = inviteCandidates.value.filter((f) => selected.value.has(f.friendId) && !f.publicKey);
    if (missing.length > 0) {
        errorMessage.value = `成员 ${missing.map((m) => m.nickname).join('、')} 尚未生成密钥,请稍后再试`;
        return;
    }
    submitting.value = true;
    try {
        emit('invite', Array.from(selected.value));
    }
    finally {
        submitting.value = false;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['invite-item']} */ ;
/** @type {__VLS_StyleScopedClasses['badge']} */ ;
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
    (__VLS_ctx.groupName);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-block" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "info-label" },
    });
    (__VLS_ctx.members.length);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
        ...{ class: "member-list" },
    });
    for (const [m] of __VLS_getVForSourceType((__VLS_ctx.members))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
            key: (m.userId),
            ...{ class: "member-row" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "member-avatar" },
        });
        ((m.nickname || m.username).slice(0, 1).toUpperCase());
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "member-name" },
        });
        (m.nickname || m.username);
        if (m.userId === __VLS_ctx.ownerId) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "badge" },
            });
        }
        if (__VLS_ctx.myUserId === m.userId) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "badge me" },
            });
        }
    }
    if (!__VLS_ctx.showInvite) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "modal-actions" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.visible))
                        return;
                    if (!(!__VLS_ctx.showInvite))
                        return;
                    __VLS_ctx.showInvite = true;
                } },
            ...{ class: "apple-button secondary" },
            type: "button",
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.visible))
                        return;
                    if (!(!__VLS_ctx.showInvite))
                        return;
                    __VLS_ctx.emit('leave');
                } },
            ...{ class: "apple-button danger" },
            type: "button",
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.visible))
                        return;
                    if (!(!__VLS_ctx.showInvite))
                        return;
                    __VLS_ctx.emit('close');
                } },
            ...{ class: "apple-button" },
            type: "button",
        });
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "invite-block" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "info-label" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "invite-list" },
        });
        if (__VLS_ctx.inviteCandidates.length === 0) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "empty-state small" },
            });
        }
        for (const [f] of __VLS_getVForSourceType((__VLS_ctx.inviteCandidates))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
                ...{ onClick: (...[$event]) => {
                        if (!(__VLS_ctx.visible))
                            return;
                        if (!!(!__VLS_ctx.showInvite))
                            return;
                        __VLS_ctx.toggle(f.friendId);
                    } },
                key: (f.friendId),
                type: "button",
                ...{ class: "invite-item" },
                ...{ class: ({ active: __VLS_ctx.selected.has(f.friendId) }) },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "member-avatar" },
            });
            (f.nickname.slice(0, 1).toUpperCase());
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "member-name" },
            });
            (f.nickname);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "member-tick" },
            });
            (__VLS_ctx.selected.has(f.friendId) ? '✓' : '');
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
                    if (!!(!__VLS_ctx.showInvite))
                        return;
                    __VLS_ctx.showInvite = false;
                } },
            ...{ class: "apple-button secondary" },
            type: "button",
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (__VLS_ctx.submitInvite) },
            ...{ class: "apple-button" },
            type: "button",
            disabled: (__VLS_ctx.submitting),
        });
        (__VLS_ctx.submitting ? '邀请中...' : '确认邀请');
    }
}
/** @type {__VLS_StyleScopedClasses['modal-mask']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-top']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['info-block']} */ ;
/** @type {__VLS_StyleScopedClasses['info-label']} */ ;
/** @type {__VLS_StyleScopedClasses['member-list']} */ ;
/** @type {__VLS_StyleScopedClasses['member-row']} */ ;
/** @type {__VLS_StyleScopedClasses['member-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['member-name']} */ ;
/** @type {__VLS_StyleScopedClasses['badge']} */ ;
/** @type {__VLS_StyleScopedClasses['badge']} */ ;
/** @type {__VLS_StyleScopedClasses['me']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['danger']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['invite-block']} */ ;
/** @type {__VLS_StyleScopedClasses['info-label']} */ ;
/** @type {__VLS_StyleScopedClasses['invite-list']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['small']} */ ;
/** @type {__VLS_StyleScopedClasses['invite-item']} */ ;
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
            showInvite: showInvite,
            selected: selected,
            errorMessage: errorMessage,
            submitting: submitting,
            inviteCandidates: inviteCandidates,
            toggle: toggle,
            submitInvite: submitInvite,
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
