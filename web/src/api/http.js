import axios from 'axios';
const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api';
export const http = axios.create({
    baseURL: apiBaseURL
});
http.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});
http.interceptors.response.use((response) => response, (error) => {
    const message = error.response?.data?.message || '请求失败，请稍后重试';
    return Promise.reject(new Error(message));
});
