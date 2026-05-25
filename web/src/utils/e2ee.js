const ALGORITHM = 'RSA-OAEP';
const HASH = 'SHA-256';
export const E2EE_MESSAGE_ALGORITHM = 'rsa-oaep-sha256';
// 浏览器只在 secure context（HTTPS 或 localhost）下暴露 window.crypto.subtle，
// 否则取到的会是 undefined。手机通过 http://电脑IP:端口 访问是典型的踩坑场景。
// 在任何加解密入口被调用之前先抛友好错误，避免出现 "Cannot read properties of
// undefined (reading 'generateKey')" 这种泄漏到 UI 的程序员级报错。
function ensureSubtleAvailable() {
    if (typeof window === 'undefined')
        return;
    const subtle = window.crypto?.subtle;
    if (subtle)
        return;
    if (window.isSecureContext === false) {
        throw new Error('当前页面不是安全上下文（非 HTTPS / 非 localhost），浏览器禁用了加密 API。' +
            '请使用 https:// 或 localhost 访问。');
    }
    throw new Error('当前浏览器不支持 Web Crypto API，请更换为较新的浏览器。');
}
export function buildPrivateKeyStorageKey(userId) {
    return `e2ee-private-key:${userId}`;
}
export function maskEncryptedMessage(value) {
    return value ? value : '***（已加密）';
}
export function normalizeEncryptedPayload(payload) {
    return {
        ciphertext: payload?.ciphertext || '',
        algorithm: payload?.algorithm || ''
    };
}
export function buildEncryptedMessageDisplay(decryptedText, payload) {
    return {
        content: maskEncryptedMessage(decryptedText),
        ...normalizeEncryptedPayload(payload)
    };
}
export function selectMessagePayloadForUser(message, currentUserId) {
    if (currentUserId && message.senderId === currentUserId) {
        return normalizeEncryptedPayload({
            ciphertext: message.senderCiphertext,
            algorithm: message.senderAlgorithm
        });
    }
    return normalizeEncryptedPayload({
        ciphertext: message.receiverCiphertext,
        algorithm: message.receiverAlgorithm
    });
}
function arrayBufferToBase64(buffer) {
    return btoa(String.fromCharCode(...new Uint8Array(buffer)));
}
function base64ToArrayBuffer(value) {
    const binary = atob(value);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes.buffer;
}
export async function generateKeyPair() {
    ensureSubtleAvailable();
    return window.crypto.subtle.generateKey({
        name: ALGORITHM,
        modulusLength: 2048,
        publicExponent: new Uint8Array([1, 0, 1]),
        hash: HASH
    }, true, ['encrypt', 'decrypt']);
}
export async function exportPublicKey(publicKey) {
    const buffer = await window.crypto.subtle.exportKey('spki', publicKey);
    return arrayBufferToBase64(buffer);
}
export async function exportPrivateKey(privateKey) {
    const buffer = await window.crypto.subtle.exportKey('pkcs8', privateKey);
    return arrayBufferToBase64(buffer);
}
export async function importPublicKey(serializedKey) {
    ensureSubtleAvailable();
    return window.crypto.subtle.importKey('spki', base64ToArrayBuffer(serializedKey), { name: ALGORITHM, hash: HASH }, true, ['encrypt']);
}
export async function importPrivateKey(serializedKey) {
    ensureSubtleAvailable();
    return window.crypto.subtle.importKey('pkcs8', base64ToArrayBuffer(serializedKey), { name: ALGORITHM, hash: HASH }, true, ['decrypt']);
}
export async function encryptMessage(publicKey, content) {
    const encoded = new TextEncoder().encode(content);
    const ciphertext = await window.crypto.subtle.encrypt({ name: ALGORITHM }, publicKey, encoded);
    return {
        ciphertext: arrayBufferToBase64(ciphertext),
        algorithm: E2EE_MESSAGE_ALGORITHM
    };
}
export async function decryptMessage(privateKey, ciphertext) {
    const decrypted = await window.crypto.subtle.decrypt({ name: ALGORITHM }, privateKey, base64ToArrayBuffer(ciphertext));
    return new TextDecoder().decode(decrypted);
}
export function savePrivateKey(userId, serializedKey) {
    localStorage.setItem(buildPrivateKeyStorageKey(userId), serializedKey);
}
export function loadPrivateKey(userId) {
    return localStorage.getItem(buildPrivateKeyStorageKey(userId)) || '';
}
