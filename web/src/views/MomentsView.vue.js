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
const content = ref('');
const uploadedImageKey = ref('');
const uploadedImageUrl = ref('');
const feedback = ref('');
const errorMessage = ref('');
const loading = ref(false);
const uploadingImage = ref(false);
const commentDrafts = ref({});
const aiMode = ref('generate');
const aiTone = ref('自然');
const aiPrompt = ref('');
const aiLoading = ref(false);
const aiResult = ref('');
async function loadMoments() {
    try {
        const { data } = await http.get('/moments');
        moments.value = data;
    }
    catch (error) {
        errorMessage.value = error.message;
    }
}
async function uploadImage(event) {
    const input = event.target;
    const file = input.files?.[0];
    if (!file)
        return;
    feedback.value = '';
    errorMessage.value = '';
    uploadingImage.value = true;
    try {
        const formData = new FormData();
        formData.append('file', file);
        const { data } = await http.post('/uploads/images', formData);
        uploadedImageKey.value = data.objectKey;
        uploadedImageUrl.value = data.url;
        feedback.value = '图片已上传';
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        uploadingImage.value = false;
        input.value = '';
    }
}
async function requestAIAssist() {
    const prompt = aiPrompt.value.trim();
    const currentContent = content.value.trim();
    if (aiMode.value === 'generate' && !prompt) {
        errorMessage.value = '请输入想法后再生成文案';
        feedback.value = '';
        return;
    }
    if (aiMode.value === 'polish' && !currentContent) {
        errorMessage.value = '请先输入正文后再进行润色';
        feedback.value = '';
        return;
    }
    const payload = {
        mode: aiMode.value,
        prompt: aiMode.value === 'generate' ? prompt : '',
        content: aiMode.value === 'polish' ? currentContent : '',
        tone: aiTone.value,
        hasImage: Boolean(uploadedImageKey.value)
    };
    aiLoading.value = true;
    aiResult.value = '';
    feedback.value = '';
    errorMessage.value = '';
    try {
        const { data } = await http.post('/moments/ai-assist', payload);
        aiResult.value = data.text;
        feedback.value = 'AI 文案已生成';
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        aiLoading.value = false;
    }
}
function applyAIResult() {
    if (!aiResult.value)
        return;
    content.value = aiResult.value;
    feedback.value = '已填入正文';
}
async function publishMoment() {
    if (!content.value.trim())
        return;
    feedback.value = '';
    errorMessage.value = '';
    loading.value = true;
    try {
        await http.post('/moments', {
            content: content.value.trim(),
            imageKeys: uploadedImageKey.value ? [uploadedImageKey.value] : []
        });
        content.value = '';
        uploadedImageKey.value = '';
        uploadedImageUrl.value = '';
        aiPrompt.value = '';
        aiResult.value = '';
        feedback.value = '动态已发布';
        await loadMoments();
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        loading.value = false;
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
onMounted(loadMoments);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['composer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['moment-card']} */ ;
/** @type {__VLS_StyleScopedClasses['composer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-result']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-mode-row']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-picker']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-picker']} */ ;
/** @type {__VLS_StyleScopedClasses['clickable-name']} */ ;
/** @type {__VLS_StyleScopedClasses['moments-layout']} */ ;
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
    ...{ class: "card apple-panel composer-card" },
    ...{ class: (__VLS_ctx.skin.surfaceClass) },
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.textarea)({
    value: (__VLS_ctx.content),
    ...{ class: "apple-textarea" },
    placeholder: "分享这一刻...",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ai-assistant card apple-panel" },
    ...{ class: (__VLS_ctx.skin.accentClass) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ai-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "muted" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ai-mode-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.aiMode = 'generate';
        } },
    ...{ class: "apple-button secondary" },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.aiMode === 'generate' }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.aiMode = 'polish';
        } },
    ...{ class: "apple-button secondary" },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.aiMode === 'polish' }) },
});
if (__VLS_ctx.aiMode === 'generate') {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
        ...{ class: "ai-field" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "apple-label" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
        ...{ class: "apple-input" },
        placeholder: "比如：周末和朋友露营，看日落很治愈",
    });
    (__VLS_ctx.aiPrompt);
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "muted ai-hint" },
    });
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
    ...{ class: "ai-field" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.select, __VLS_intrinsicElements.select)({
    value: (__VLS_ctx.aiTone),
    ...{ class: "apple-input" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
    value: "自然",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
    value: "幽默",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
    value: "文艺",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
    value: "简洁",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ai-actions" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.requestAIAssist) },
    ...{ class: "apple-button secondary" },
    type: "button",
    disabled: (__VLS_ctx.aiLoading),
});
(__VLS_ctx.aiLoading ? '生成中...' : __VLS_ctx.aiMode === 'polish' ? '开始润色' : '生成文案');
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.applyAIResult) },
    ...{ class: "apple-button" },
    type: "button",
    disabled: (!__VLS_ctx.aiResult),
});
if (__VLS_ctx.aiResult) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "ai-result" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "apple-label" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
    (__VLS_ctx.aiResult);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-field" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
    ...{ class: "upload-picker" },
    ...{ class: ({ uploading: __VLS_ctx.uploadingImage }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onChange: (__VLS_ctx.uploadImage) },
    ...{ class: "upload-input" },
    type: "file",
    accept: "image/*",
});
if (__VLS_ctx.uploadedImageUrl) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.uploadedImageUrl),
        alt: "uploaded preview",
        ...{ class: "upload-cover" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "upload-overlay" },
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "upload-plus" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "upload-hint" },
    });
}
if (__VLS_ctx.uploadingImage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "muted" },
    });
}
if (__VLS_ctx.uploadedImageUrl) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "upload-tools" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.uploadedImageUrl))
                    return;
                __VLS_ctx.uploadedImageKey = '';
                __VLS_ctx.uploadedImageUrl = '';
            } },
        ...{ class: "apple-button secondary" },
        type: "button",
    });
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.publishMoment) },
    ...{ class: "apple-button" },
    disabled: (__VLS_ctx.loading || __VLS_ctx.uploadingImage),
});
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
/** @type {__VLS_StyleScopedClasses['composer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['success']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-textarea']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-assistant']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-head']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-mode-row']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-field']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-field']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['ai-result']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-field']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-picker']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-input']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-cover']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-overlay']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-plus']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-tools']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
/** @type {__VLS_StyleScopedClasses['secondary']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
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
            content: content,
            uploadedImageKey: uploadedImageKey,
            uploadedImageUrl: uploadedImageUrl,
            feedback: feedback,
            errorMessage: errorMessage,
            loading: loading,
            uploadingImage: uploadingImage,
            commentDrafts: commentDrafts,
            aiMode: aiMode,
            aiTone: aiTone,
            aiPrompt: aiPrompt,
            aiLoading: aiLoading,
            aiResult: aiResult,
            uploadImage: uploadImage,
            requestAIAssist: requestAIAssist,
            applyAIResult: applyAIResult,
            publishMoment: publishMoment,
            toggleLike: toggleLike,
            submitComment: submitComment,
            deleteMoment: deleteMoment,
            openHomepage: openHomepage,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
