// 群聊混合加密(AES-GCM 加密内容 + RSA-OAEP 包裹 AES 密钥)
//
// 设计:
// - 发送者每发一条消息生成一次性 AES-256-GCM 会话密钥与 12 字节 IV
// - 用 AES 密钥对明文加密,得到一份共享密文(group_messages.content_ciphertext)
// - 然后对每个群成员公钥用 RSA-OAEP 加密 AES 密钥的 raw 字节,得到 N 份小密钥密文
// - 服务端只看到密文与各成员的密钥包裹,无法获得明文
// - 接收者用本地私钥先解出 AES 密钥,再解出明文
//
// 与单聊的区别:单聊每条消息存两份完整 RSA 密文(senderCt + receiverCt),
// 群聊每条消息只存一份共享 AES 密文 + N 份很小的 RSA 密钥密文,数据库不会随群规模膨胀。
import { base64ToBytes, bytesToBase64, E2EE_MESSAGE_ALGORITHM, encryptRawWithPublicKey, decryptRawWithPrivateKey, importPublicKey } from './e2ee';
export const GROUP_CONTENT_ALGORITHM = 'aes-gcm-256';
const AES_KEY_LENGTH = 256;
const AES_IV_BYTES = 12;
function ensureSubtleAvailable() {
    if (typeof window === 'undefined')
        return;
    const subtle = window.crypto?.subtle;
    if (subtle)
        return;
    if (window.isSecureContext === false) {
        throw new Error('当前页面不是安全上下文,浏览器禁用了加密 API。请使用 https:// 或 localhost 访问。');
    }
    throw new Error('当前浏览器不支持 Web Crypto API,请更换为较新的浏览器。');
}
async function generateAesKey() {
    ensureSubtleAvailable();
    return window.crypto.subtle.generateKey({ name: 'AES-GCM', length: AES_KEY_LENGTH }, true, ['encrypt', 'decrypt']);
}
async function exportAesKeyRaw(key) {
    return window.crypto.subtle.exportKey('raw', key);
}
async function importAesKeyRaw(raw) {
    return window.crypto.subtle.importKey('raw', raw, { name: 'AES-GCM' }, false, ['decrypt']);
}
/**
 * 为整群构建一条群消息的完整密文信封。
 * @param plaintext 待发送的明文
 * @param members  当前群所有成员(包括自己)的公钥列表;memberKeys 必须覆盖所有成员
 */
export async function buildGroupMessageEnvelope(plaintext, members) {
    ensureSubtleAvailable();
    if (!members.length) {
        throw new Error('群成员列表为空,无法发送消息');
    }
    for (const m of members) {
        if (!m.publicKey) {
            throw new Error(`成员 ${m.userId} 还没有上传公钥,暂时无法发送加密消息`);
        }
    }
    const aesKey = await generateAesKey();
    const aesRaw = await exportAesKeyRaw(aesKey);
    const iv = window.crypto.getRandomValues(new Uint8Array(AES_IV_BYTES));
    const contentBuffer = await window.crypto.subtle.encrypt({ name: 'AES-GCM', iv }, aesKey, new TextEncoder().encode(plaintext));
    const memberKeys = [];
    for (const m of members) {
        const memberPubKey = await importPublicKey(m.publicKey);
        const wrapped = await encryptRawWithPublicKey(memberPubKey, aesRaw);
        memberKeys.push({
            userId: m.userId,
            keyCiphertext: wrapped,
            keyAlgorithm: E2EE_MESSAGE_ALGORITHM
        });
    }
    return {
        contentCiphertext: bytesToBase64(contentBuffer),
        contentIv: bytesToBase64(iv.buffer),
        contentAlgorithm: GROUP_CONTENT_ALGORITHM,
        memberKeys
    };
}
/**
 * 用本地私钥解密一条群消息(只需要自己那份 keyCiphertext)。
 */
export async function decryptGroupMessage(myPrivateKey, keyCiphertext, contentCiphertext, contentIv) {
    ensureSubtleAvailable();
    const aesRaw = await decryptRawWithPrivateKey(myPrivateKey, keyCiphertext);
    const aesKey = await importAesKeyRaw(aesRaw);
    const plain = await window.crypto.subtle.decrypt({ name: 'AES-GCM', iv: new Uint8Array(base64ToBytes(contentIv)) }, aesKey, base64ToBytes(contentCiphertext));
    return new TextDecoder().decode(plain);
}
