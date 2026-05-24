import test from 'node:test'
import assert from 'node:assert/strict'

import { genderCodeToLabel, genderLabelToCode } from './gender.js'

test('maps stored codes to Chinese gender labels', () => {
  assert.equal(genderCodeToLabel(0), '女')
  assert.equal(genderCodeToLabel(1), '男')
})

test('maps displayed Chinese gender labels back to stored codes', () => {
  assert.equal(genderLabelToCode('女'), 0)
  assert.equal(genderLabelToCode('男'), 1)
})
