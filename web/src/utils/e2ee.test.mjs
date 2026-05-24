import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildEncryptedMessageDisplay,
  buildPrivateKeyStorageKey,
  maskEncryptedMessage,
  normalizeEncryptedPayload
} from './e2ee.js'

test('builds per-user private key storage keys', () => {
  assert.equal(buildPrivateKeyStorageKey(7), 'e2ee-private-key:7')
})

test('masks encrypted messages when decrypted text is unavailable', () => {
  assert.equal(maskEncryptedMessage(''), '***（已加密）')
  assert.equal(maskEncryptedMessage(null), '***（已加密）')
  assert.equal(maskEncryptedMessage('hello'), 'hello')
})

test('normalizes encrypted payloads for ui rendering', () => {
  assert.deepEqual(
    normalizeEncryptedPayload({ ciphertext: 'abc', algorithm: 'rsa-oaep-sha256' }),
    { ciphertext: 'abc', algorithm: 'rsa-oaep-sha256' }
  )

  assert.deepEqual(normalizeEncryptedPayload(), {
    ciphertext: '',
    algorithm: ''
  })
})

test('builds encrypted message display with content field for rendering while preserving payload metadata', () => {
  assert.deepEqual(
    buildEncryptedMessageDisplay('', { ciphertext: 'cipher', algorithm: 'rsa-oaep-sha256' }),
    {
      content: '***（已加密）',
      ciphertext: 'cipher',
      algorithm: 'rsa-oaep-sha256'
    }
  )
})
