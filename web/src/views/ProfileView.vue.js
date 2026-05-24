import { onMounted, reactive, ref } from 'vue';
import AppNav from '../components/AppNav.vue';
import { http } from '../api/http';
import { useAuthStore } from '../stores/auth';
const authStore = useAuthStore();
const loading = ref(false);
const message = ref('');
const errorMessage = ref('');
const profile = reactive({
    nickname: '',
    gender: 0,
    signature: ''
});
function syncProfile() {
    profile.nickname = authStore.user?.nickname || '';
    profile.gender = authStore.user?.gender || 0;
    profile.signature = authStore.user?.signature || '';
}
async function saveProfile() {
    loading.value = true;
    message.value = '';
    errorMessage.value = '';
    try {
        const { data } = await http.put('/users/me', profile);
        authStore.user = data;
        syncProfile();
        message.value = '保存成功';
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        loading.value = false;
    }
}
onMounted(syncProfile);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
if (__VLS_ctx.message) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "success-text" },
    });
    (__VLS_ctx.message);
}
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "error-text" },
    });
    (__VLS_ctx.errorMessage);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "昵称",
});
(__VLS_ctx.profile.nickname);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "number",
    placeholder: "性别：0/1/2",
});
(__VLS_ctx.profile.gender);
__VLS_asFunctionalElement(__VLS_intrinsicElements.textarea)({
    value: (__VLS_ctx.profile.signature),
    placeholder: "个性签名",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.saveProfile) },
    disabled: (__VLS_ctx.loading),
});
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['page-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['content-card']} */ ;
/** @type {__VLS_StyleScopedClasses['success-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error-text']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppNav: AppNav,
            loading: loading,
            message: message,
            errorMessage: errorMessage,
            profile: profile,
            saveProfile: saveProfile,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
