import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import AppNav from '../components/AppNav.vue';
import { http } from '../api/http';
import { resolveHomepageSkin } from '../constants/homepageSkins';
import { useAuthStore } from '../stores/auth';
const authStore = useAuthStore();
const router = useRouter();
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin));
const moments = ref([]);
const feedback = ref(typeof history.state?.momentPublishedMessage === 'string' ? history.state.momentPublishedMessage : '');
const errorMessage = ref('');
const commentDrafts = ref({});
async function loadMoments() {
    try {
        const { data } = await http.get('/moments');
        moments.value = data;
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function toggleLike(item) {
    errorMessage.value = '';
    feedback.value = '';
    try {
        if (item.likedByMe) {
            await http.delete(`/moments/${item.id}/likes/me`);
        }
        else {
            await http.post(`/moments/${item.id}/likes`);
        }
        await loadMoments();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function submitComment(item) {
    const content = commentDrafts.value[item.id]?.trim();
    if (!content)
        return;
    errorMessage.value = '';
    feedback.value = '';
    try {
        await http.post(`/moments/${item.id}/comments`, { content });
        commentDrafts.value[item.id] = '';
        await loadMoments();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function deleteMoment(item) {
    errorMessage.value = '';
    feedback.value = '';
    try {
        await http.delete(`/moments/${item.id}`);
        feedback.value = '动态已删除';
        await loadMoments();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
function openHomepage(userId) {
    router.push(`/users/${userId}`);
}
function openCreateMoment() {
    router.push('/moments/create');
}
onMounted(async () => {
    await loadMoments();
    if (feedback.value) {
        history.replaceState({}, document.title);
    }
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['moments-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-card']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['clickable-name']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-head']} */ ;
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
    ...{ class: "moments-layout" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "card apple-panel moments-hero" },
    ...{ class: (__VLS_ctx.skin.surfaceClass) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "moments-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "muted" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.openCreateMoment) },
    ...{ class: "apple-button secondary" },
    type: "button",
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "feed-column" },
});
if (__VLS_ctx.moments.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card apple-panel empty-state-card" },
    });
}
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.moments))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.article, __VLS_intrinsicElements.article)({
        key: (item.id),
        ...{ class: "card apple-panel moment-card" },
        ...{ class: (__VLS_ctx.skin.accentClass) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "moment-head" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.openHomepage(item.userId);
            } },
        src: (item.avatar),
        alt: "avatar",
        ...{ class: "avatar clickable-avatar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.openHomepage(item.userId);
            } },
        ...{ class: "clickable-name" },
    });
    (item.nickname);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "muted moment-time" },
    });
    (item.createdAt);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "moment-content" },
    });
    (item.content);
    if (item.images.length) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "moment-images" },
        });
        for (const [src] of __VLS_getVForSourceType((item.images))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
                key: (src),
                src: (src),
                alt: "moment image",
                ...{ class: "moment-image" },
            });
        }
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "moment-actions" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.toggleLike(item);
            } },
        ...{ class: "apple-button secondary" },
        type: "button",
    });
    (item.likedByMe ? '取消点赞' : '点赞');
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "muted" },
    });
    (item.likeCount);
    if (__VLS_ctx.authStore.user?.id === item.userId) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.authStore.user?.id === item.userId))
                        return;
                    __VLS_ctx.deleteMoment(item);
                } },
            ...{ class: "apple-button danger" },
            type: "button",
        });
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "comment-composer" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
        ...{ class: "apple-input" },
        placeholder: "写下评论...",
    });
    (__VLS_ctx.commentDrafts[item.id]);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.submitComment(item);
            } },
        ...{ class: "apple-button secondary" },
        type: "button",
    });
    if (item.comments.length) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "comment-list" },
        });
        for (const [comment] of __VLS_getVForSourceType((item.comments))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                key: (comment.id),
                ...{ class: "comment-item" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
            (comment.nickname);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "comment-separator" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (comment.content);
        }
    }
}
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-page']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-head']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['success']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['feed-column']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-card']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-head']} */ ;
/** @type {__VLS_StyleScopedClasses['avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['clickable-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['clickable-name']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-time']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-content']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-images']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-image']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['danger']} */ ;
/** @type {__VLS_StyleScopedClasses['comment-composer']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['comment-list']} */ ;
/** @type {__VLS_StyleScopedClasses['comment-item']} */ ;
/** @type {__VLS_StyleScopedClasses['comment-separator']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            authStore: authStore,
            skin: skin,
            moments: moments,
            feedback: feedback,
            errorMessage: errorMessage,
            commentDrafts: commentDrafts,
            toggleLike: toggleLike,
            submitComment: submitComment,
            deleteMoment: deleteMoment,
            openHomepage: openHomepage,
            openCreateMoment: openCreateMoment,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
