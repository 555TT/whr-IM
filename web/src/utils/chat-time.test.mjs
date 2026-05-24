import test from 'node:test'
import assert from 'node:assert/strict'

import { formatChatMessageTime } from './chat-time.js'

test('formats same-day message time as HH:mm', () => {
  const now = new Date('2026-05-24T22:30:00+08:00')
  const sameDay = '2026-05-24T09:05:00+08:00'

  assert.equal(formatChatMessageTime(sameDay, now), '09:05')
})

test('formats cross-day message time as MM-DD HH:mm', () => {
  const now = new Date('2026-05-24T22:30:00+08:00')
  const previousDay = '2026-05-23T09:05:00+08:00'

  assert.equal(formatChatMessageTime(previousDay, now), '05-23 09:05')
})
