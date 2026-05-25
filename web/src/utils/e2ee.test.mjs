import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildEncryptedMessageDisplay,
  buildPrivateKeyStorageKey,
  maskEncryptedMessage,
  normalizeEncryptedPayload,
  selectMessagePayloadForUser
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

test('selects sender payload for messages sent by current user', () => {
  assert.deepEqual(
    selectMessagePayloadForUser(
      {
        senderId: 7,
        senderCiphertext: 'sender-copy',
        senderAlgorithm: 'rsa-oaep-sha256',
        receiverCiphertext: 'receiver-copy',
        receiverAlgorithm: 'rsa-oaep-sha256'
      },
      7
    ),
    {
      ciphertext: 'sender-copy',
      algorithm: 'rsa-oaep-sha256'
    }
  )
})

test('selects receiver payload for messages received by current user', () => {
  assert.deepEqual(
    selectMessagePayloadForUser(
      {
        senderId: 7,
        senderCiphertext: 'sender-copy',
        senderAlgorithm: 'rsa-oaep-sha256',
        receiverCiphertext: 'receiver-copy',
        receiverAlgorithm: 'rsa-oaep-sha256'
      },
      9
    ),
    {
      ciphertext: 'receiver-copy',
      algorithm: 'rsa-oaep-sha256'
    }
  )
})
