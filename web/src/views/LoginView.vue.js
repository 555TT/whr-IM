import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { http } from '../api/http';
import { useAuthStore } from '../stores/auth';
const router = useRouter();
const authStore = useAuthStore();
const errorMessage = ref('');
const successMessage = ref('');
const loading = ref(false);
const form = reactive({
    username: '',
    password: '',
    confirmPassword: ''
});
async function register() {
    errorMessage.value = '';
    successMessage.value = '';
    loading.value = true;
    try {
        const { data } = await http.post('/auth/register', form);
        successMessage.value = `注册成功：${data.user.username}`;
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        loading.value = false;
    }
}
async function login() {
    errorMessage.value = '';
    successMessage.value = '';
    loading.value = true;
    try {
        const { data } = await http.post('/auth/login', {
            username: form.username,
            password: form.password
        });
        authStore.setSession(data.token, data.user);
        router.push('/chat');
    }
    catch (error) {
        errorMessage.value = error.message;
    }
    finally {
        loading.value = false;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-shell login-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "card login-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
if (__VLS_ctx.successMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "success-text" },
    });
    (__VLS_ctx.successMessage);
}
if (__VLS_ctx.errorMessage) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "error-text" },
    });
    (__VLS_ctx.errorMessage);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "用户名",
});
(__VLS_ctx.form.username);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "password",
    placeholder: "密码",
});
(__VLS_ctx.form.password);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "password",
    placeholder: "确认密码",
});
(__VLS_ctx.form.confirmPassword);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "actions" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.register) },
    disabled: (__VLS_ctx.loading),
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.login) },
    disabled: (__VLS_ctx.loading),
});
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['login-page']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card']} */ ;
/** @type {__VLS_StyleScopedClasses['success-text']} */ ;
/** @type {__VLS_StyleScopedClasses['error-text']} */ ;
/** @type {__VLS_StyleScopedClasses['actions']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            errorMessage: errorMessage,
            successMessage: successMessage,
            loading: loading,
            form: form,
            register: register,
            login: login,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
