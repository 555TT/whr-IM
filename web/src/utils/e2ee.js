const ALGORITHM = 'RSA-OAEP';
const HASH = 'SHA-256';
export const E2EE_MESSAGE_ALGORITHM = 'rsa-oaep-sha256';
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
    return window.crypto.subtle.importKey('spki', base64ToArrayBuffer(serializedKey), { name: ALGORITHM, hash: HASH }, true, ['encrypt']);
}
export async function importPrivateKey(serializedKey) {
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
