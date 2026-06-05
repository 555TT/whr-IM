import { computed, onMounted, reactive, ref } from 'vue';
import AppNav from '../components/AppNav.vue';
import AvatarCropper from '../components/AvatarCropper.vue';
import { http } from '../api/http';
import { resolveHomepageSkin } from '../constants/homepageSkins';
import { useAuthStore } from '../stores/auth';
import { genderCodeToLabel, genderLabelToCode } from '../utils/gender';
const maxAvatarSize = 2 * 1024 * 1024;
const authStore = useAuthStore();
const loading = ref(false);
const message = ref('');
const errorMessage = ref('');
const avatarPreviewUrl = ref('');
const avatarFile = ref(null);
const cropperVisible = ref(false);
const cropperImageUrl = ref('');
const profile = reactive({
    nickname: '',
    gender: '女',
    signature: ''
});
const displayAvatar = computed(() => avatarPreviewUrl.value || authStore.user?.avatar || '');
function syncProfile() {
    profile.nickname = authStore.user?.nickname || '';
    profile.gender = genderCodeToLabel(authStore.user?.gender ?? 0);
    profile.signature = authStore.user?.signature || '';
    avatarPreviewUrl.value = '';
    avatarFile.value = null;
}
function pickAvatar(event) {
    const input = event.target;
    const file = input.files?.[0];
    if (!file)
        return;
    errorMessage.value = '';
    message.value = '';
    if (!file.type.startsWith('image/')) {
        errorMessage.value = '请选择图片文件';
        input.value = '';
        return;
    }
    if (file.size > maxAvatarSize) {
        errorMessage.value = '头像图片不能超过 2MB';
        input.value = '';
        return;
    }
    cropperImageUrl.value = URL.createObjectURL(file);
    cropperVisible.value = true;
    input.value = '';
}
function applyCroppedAvatar(payload) {
    avatarFile.value = payload.file;
    avatarPreviewUrl.value = payload.previewUrl;
    cropperVisible.value = false;
    cropperImageUrl.value = '';
}
async function saveProfile() {
    loading.value = true;
    message.value = '';
    errorMessage.value = '';
    try {
        let avatar = authStore.user?.avatar || '';
        if (avatarFile.value) {
            const formData = new FormData();
            formData.append('file', avatarFile.value);
            const { data: uploadData } = await http.post('/uploads/images', formData);
            avatar = uploadData.url;
        }
        const { data } = await http.put('/users/me', {
            nickname: profile.nickname,
            gender: genderLabelToCode(profile.gender),
            signature: profile.signature,
            avatar,
            homepageSkin: authStore.user?.homepageSkin || 'aurora',
            avatarAccessory: authStore.user?.avatarAccessory || 'none',
            titleBadge: authStore.user?.titleBadge || 'none',
            homepageBackground: authStore.user?.homepageBackground || 'plain',
            homepageLayout: authStore.user?.homepageLayout || 'classic'
        });
        authStore.user = data;
        syncProfile();
        message.value = '资料已更新';
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        loading.value = false;
    }
}
function currentSkin() {
    return resolveHomepageSkin(authStore.user?.homepageSkin);
}
onMounted(syncProfile);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['profile-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-header']} */ ;
/** @type {__VLS_StyleScopedClasses['skin-card']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-header-main']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-actions']} */ ;
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
    ...{ class: "card apple-panel profile-shell" },
    ...{ class: (__VLS_ctx.currentSkin().surfaceClass) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "profile-header profile-hero" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "profile-header-main" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
    ...{ class: "avatar-picker" },
    ...{ class: (__VLS_ctx.currentSkin().accentClass) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onChange: (__VLS_ctx.pickAvatar) },
    ...{ class: "avatar-input" },
    type: "file",
    accept: "image/*",
});
if (__VLS_ctx.displayAvatar) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.displayAvatar),
        alt: "avatar",
        ...{ class: "profile-avatar" },
    });
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "avatar-overlay" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "muted" },
});
if (__VLS_ctx.message) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text success" },
    });
    (__VLS_ctx.message);
}
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "status-text error" },
    });
    (__VLS_ctx.errorMessage);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "profile-grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ class: "apple-input" },
    placeholder: "昵称",
});
(__VLS_ctx.profile.nickname);
__VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.select, __VLS_intrinsicElements.select)({
    value: (__VLS_ctx.profile.gender),
    ...{ class: "apple-input" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
    value: "女",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
    value: "男",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "apple-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.textarea)({
    value: (__VLS_ctx.profile.signature),
    ...{ class: "apple-textarea" },
    placeholder: "写一句介绍自己的话",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "profile-actions" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.saveProfile) },
    ...{ class: "apple-button" },
    disabled: (__VLS_ctx.loading),
});
/** @type {[typeof AvatarCropper, ]} */ ;
// @ts-ignore
const __VLS_3 = __VLS_asFunctionalComponent(AvatarCropper, new AvatarCropper({
    ...{ 'onClose': {} },
    ...{ 'onConfirm': {} },
    visible: (__VLS_ctx.cropperVisible),
    imageUrl: (__VLS_ctx.cropperImageUrl),
}));
const __VLS_4 = __VLS_3({
    ...{ 'onClose': {} },
    ...{ 'onConfirm': {} },
    visible: (__VLS_ctx.cropperVisible),
    imageUrl: (__VLS_ctx.cropperImageUrl),
}, ...__VLS_functionalComponentArgsRest(__VLS_3));
let __VLS_6;
let __VLS_7;
let __VLS_8;
const __VLS_9 = {
    onClose: (...[$event]) => {
        __VLS_ctx.cropperVisible = false;
    }
};
const __VLS_10 = {
    onConfirm: (__VLS_ctx.applyCroppedAvatar)
};
var __VLS_5;
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-page']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-header']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-header-main']} */ ;
/** @type {__VLS_StyleScopedClasses['avatar-picker']} */ ;
/** @type {__VLS_StyleScopedClasses['avatar-input']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['avatar-overlay']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['success']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-input']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-label']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-textarea']} */ ;
/** @type {__VLS_StyleScopedClasses['profile-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['apple-button']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            AvatarCropper: AvatarCropper,
            loading: loading,
            message: message,
            errorMessage: errorMessage,
            cropperVisible: cropperVisible,
            cropperImageUrl: cropperImageUrl,
            profile: profile,
            displayAvatar: displayAvatar,
            pickAvatar: pickAvatar,
            applyCroppedAvatar: applyCroppedAvatar,
            saveProfile: saveProfile,
            currentSkin: currentSkin,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
