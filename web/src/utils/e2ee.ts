const ALGORITHM = 'RSA-OAEP'
const HASH = 'SHA-256'
export const E2EE_MESSAGE_ALGORITHM = 'rsa-oaep-sha256'

export type EncryptedPayload = {
  ciphertext?: string
  algorithm?: string
}

export type DirectionalEncryptedMessage = {
  senderId: number
  senderCiphertext?: string
  senderAlgorithm?: string
  receiverCiphertext?: string
  receiverAlgorithm?: string
}

export function buildPrivateKeyStorageKey(userId: number) {
  return `e2ee-private-key:${userId}`
}

export function maskEncryptedMessage(value?: string | null) {
  return value ? value : '***（已加密）'
}

export function normalizeEncryptedPayload(payload?: EncryptedPayload) {
  return {
    ciphertext: payload?.ciphertext || '',
    algorithm: payload?.algorithm || ''
  }
}

export function buildEncryptedMessageDisplay(
  decryptedText: string | null | undefined,
  payload?: EncryptedPayload
) {
  return {
    content: maskEncryptedMessage(decryptedText),
    ...normalizeEncryptedPayload(payload)
  }
}

export function selectMessagePayloadForUser(message: DirectionalEncryptedMessage, currentUserId?: number | null) {
  if (currentUserId && message.senderId === currentUserId) {
    return normalizeEncryptedPayload({
      ciphertext: message.senderCiphertext,
      algorithm: message.senderAlgorithm
    })
  }

  return normalizeEncryptedPayload({
    ciphertext: message.receiverCiphertext,
    algorithm: message.receiverAlgorithm
  })
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  return btoa(String.fromCharCode(...new Uint8Array(buffer)))
}

function base64ToArrayBuffer(value: string) {
  const binary = atob(value)
  const bytes = new Uint8Array(binary.length)

  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }

  return bytes.buffer
}

export async function generateKeyPair() {
  return window.crypto.subtle.generateKey(
    {
      name: ALGORITHM,
      modulusLength: 2048,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: HASH
    },
    true,
    ['encrypt', 'decrypt']
  )
}

export async function exportPublicKey(publicKey: CryptoKey) {
  const buffer = await window.crypto.subtle.exportKey('spki', publicKey)
  return arrayBufferToBase64(buffer)
}

export async function exportPrivateKey(privateKey: CryptoKey) {
  const buffer = await window.crypto.subtle.exportKey('pkcs8', privateKey)
  return arrayBufferToBase64(buffer)
}

export async function importPublicKey(serializedKey: string) {
  return window.crypto.subtle.importKey(
    'spki',
    base64ToArrayBuffer(serializedKey),
    { name: ALGORITHM, hash: HASH },
    true,
    ['encrypt']
  )
}

export async function importPrivateKey(serializedKey: string) {
  return window.crypto.subtle.importKey(
    'pkcs8',
    base64ToArrayBuffer(serializedKey),
    { name: ALGORITHM, hash: HASH },
    true,
    ['decrypt']
  )
}

export async function encryptMessage(publicKey: CryptoKey, content: string) {
  const encoded = new TextEncoder().encode(content)
  const ciphertext = await window.crypto.subtle.encrypt({ name: ALGORITHM }, publicKey, encoded)

  return {
    ciphertext: arrayBufferToBase64(ciphertext),
    algorithm: E2EE_MESSAGE_ALGORITHM
  }
}

export async function decryptMessage(privateKey: CryptoKey, ciphertext: string) {
  const decrypted = await window.crypto.subtle.decrypt(
    { name: ALGORITHM },
    privateKey,
    base64ToArrayBuffer(ciphertext)
  )

  return new TextDecoder().decode(decrypted)
}

export function savePrivateKey(userId: number, serializedKey: string) {
  localStorage.setItem(buildPrivateKeyStorageKey(userId), serializedKey)
}

export function loadPrivateKey(userId: number) {
  return localStorage.getItem(buildPrivateKeyStorageKey(userId)) || ''
}
