# E2EE Message Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace plaintext chat message storage with true end-to-end encrypted ciphertext storage, browser-local private keys, server-stored public keys, and ciphertext-only message transport.

**Architecture:** The backend will store user public keys and ciphertext-only messages, expose an authenticated public-key update endpoint, and return ciphertext in both history and WebSocket flows. The frontend will generate and persist a browser-local keypair, upload the public key, encrypt outgoing messages with the friend’s public key, decrypt incoming/history messages locally, and fall back to `***（已加密）` when decryption is unavailable.

**Tech Stack:** Go, Gin, GORM, MySQL/SQLite tests, Gorilla WebSocket, Vue 3, TypeScript, Pinia, Web Crypto API, node:test, Vite

---

## File structure map

### Backend files
- Modify: `init.sql` — rebuild schema for public-key fields and ciphertext message storage
- Modify: `server/internal/model/user.go` — add public-key fields to the user model
- Modify: `server/internal/model/message.go` — replace plaintext content fields with ciphertext fields
- Modify: `server/internal/repository/user_memory.go` — extend the `UserRepository` interface with public-key update support
- Modify: `server/internal/repository/user_gorm.go` — implement repository method to update public key
- Modify: `server/internal/repository/message_gorm.go` — persist/read ciphertext message model
- Modify: `server/internal/service/auth.go` — expose public-key update behavior through auth service and validate public-key payloads
- Modify: `server/internal/service/friend.go` — include public key in friend list DTO
- Modify: `server/internal/service/message.go` — reject plaintext flow and create ciphertext messages only
- Modify: `server/internal/handler/auth_user.go` — add `PUT /api/users/me/public-key`
- Modify: `server/internal/handler/message.go` — bind ciphertext request payloads
- Modify: `server/internal/router/router.go` — wire the new authenticated route
- Test: `server/internal/router/auth_user_test.go` — add public-key endpoint coverage
- Test: `server/internal/router/friend_test.go` — assert friend list includes public key fields
- Test: `server/internal/router/message_test.go` — assert ciphertext-only message history
- Test: `server/internal/router/websocket_test.go` — assert ciphertext-only WebSocket delivery

### Frontend files
- Create: `web/src/utils/e2ee.ts` — key generation, import/export, encrypt/decrypt, browser-local storage helpers
- Create: `web/src/utils/e2ee.js` — JS mirror for node:test
- Create: `web/src/utils/e2ee.test.mjs` — node:test coverage for fallback-safe crypto helpers that do not require DOM rendering
- Modify: `web/src/stores/auth.ts` — include public key fields in current user typing if returned by backend
- Modify: `web/src/views/ChatView.vue` — upload key if missing, block send when friend has no public key or local private key is missing, encrypt before send, decrypt when rendering
- Optional modify if needed during implementation: `web/src/api/http.ts` only if request/response handling needs adjustment; do not touch unless required

---

### Task 1: Rebuild schema for public keys and ciphertext messages

**Files:**
- Modify: `init.sql:1-65`
- Test: `server/internal/router/auth_user_test.go`
- Test: `server/internal/router/message_test.go`

- [ ] **Step 1: Write the failing backend assertions that describe the new schema-backed JSON shape**

Add a new auth test case in `server/internal/router/auth_user_test.go` that updates the current user public key and expects the profile response to contain `publicKey` and `publicKeyAlgorithm`.

```go
type updatePublicKeyBody struct {
    PublicKey string `json:"publicKey"`
    Algorithm string `json:"algorithm"`
}

type profileResponse struct {
    ID                 uint64 `json:"id"`
    Username           string `json:"username"`
    PublicKey          string `json:"publicKey"`
    PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
}
```

Also add a message-history assertion in `server/internal/router/message_test.go` that expects `ciphertext` and `algorithm` fields instead of `content`.

```go
var historyResp []struct {
    SenderID   uint64 `json:"senderId"`
    ReceiverID uint64 `json:"receiverId"`
    Ciphertext string `json:"ciphertext"`
    Algorithm  string `json:"algorithm"`
    CreatedAt  string `json:"createdAt"`
}
```

- [ ] **Step 2: Run targeted tests to verify they fail for the right reason**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run 'TestUserCanUpdateOwnPublicKey|TestMessageHistoryReturnsConversationInTimeOrder'
```

Expected: FAIL because the route/model/JSON fields do not exist yet, or because `content` is still returned instead of `ciphertext`.

- [ ] **Step 3: Update `init.sql` to the new fresh schema**

Replace the user/message table definitions with ciphertext/public-key fields.

```sql
CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    username VARCHAR(50) NOT NULL COMMENT '用户名，唯一',
    password_hash VARCHAR(255) NOT NULL COMMENT '加密后的密码',
    nickname VARCHAR(50) NOT NULL COMMENT '昵称',
    avatar VARCHAR(255) NOT NULL DEFAULT 'https://api.dicebear.com/7.x/initials/svg?seed=default-user' COMMENT '系统默认头像地址，不允许用户修改',
    gender TINYINT NOT NULL DEFAULT 0 COMMENT '性别：0-女，1-男',
    signature VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个性签名',
    public_key TEXT NOT NULL DEFAULT '' COMMENT '用户公钥',
    public_key_algorithm VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用户公钥算法',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    sender_id BIGINT UNSIGNED NOT NULL COMMENT '发送者 ID',
    receiver_id BIGINT UNSIGNED NOT NULL COMMENT '接收者 ID',
    ciphertext TEXT NOT NULL COMMENT '消息密文',
    algorithm VARCHAR(50) NOT NULL COMMENT '消息加密算法',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    PRIMARY KEY (id),
    KEY idx_messages_sender_receiver_created_at (sender_id, receiver_id, created_at),
    KEY idx_messages_receiver_sender_created_at (receiver_id, sender_id, created_at),
    CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_messages_receiver FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';
```

- [ ] **Step 4: Update Go models to match the schema**

Edit `server/internal/model/user.go`:

```go
type User struct {
    ID                 uint64 `gorm:"primaryKey" json:"id"`
    Username           string `gorm:"uniqueIndex;size:50;not null" json:"username"`
    PasswordHash       string `gorm:"size:255;not null" json:"-"`
    Nickname           string `gorm:"size:50;not null" json:"nickname"`
    Avatar             string `gorm:"size:255;not null" json:"avatar"`
    Gender             int    `gorm:"not null;default:0" json:"gender"`
    Signature          string `gorm:"size:255;not null;default:''" json:"signature"`
    PublicKey          string `gorm:"type:text;not null;default:''" json:"publicKey"`
    PublicKeyAlgorithm string `gorm:"size:50;not null;default:''" json:"publicKeyAlgorithm"`
}
```

Edit `server/internal/model/message.go`:

```go
type Message struct {
    ID         uint64    `gorm:"primaryKey" json:"id"`
    SenderID   uint64    `gorm:"not null;index" json:"senderId"`
    ReceiverID uint64    `gorm:"not null;index" json:"receiverId"`
    Ciphertext string    `gorm:"type:text;not null" json:"ciphertext"`
    Algorithm  string    `gorm:"size:50;not null" json:"algorithm"`
    CreatedAt  time.Time `json:"createdAt"`
}
```

- [ ] **Step 5: Run the same targeted tests to verify the model/schema layer is now aligned**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run 'TestUserCanUpdateOwnPublicKey|TestMessageHistoryReturnsConversationInTimeOrder'
```

Expected: still FAIL, but now for missing repository/service/handler behavior rather than missing fields.

- [ ] **Step 6: Commit the schema/model groundwork**

```bash
cd /Users/zyb/bishe/whr-im && git add init.sql server/internal/model/user.go server/internal/model/message.go server/internal/router/auth_user_test.go server/internal/router/message_test.go && git commit -m "feat: prepare e2ee schema models"
```

### Task 2: Add public-key persistence and authenticated update API

**Files:**
- Modify: `server/internal/repository/user_memory.go:13-18`
- Modify: `server/internal/repository/user_gorm.go:12-76`
- Modify: `server/internal/service/auth.go:24-137`
- Modify: `server/internal/handler/auth_user.go:12-111`
- Modify: `server/internal/router/router.go:61-84`
- Test: `server/internal/router/auth_user_test.go`

- [ ] **Step 1: Write the failing test for updating the current user public key**

Add `TestUserCanUpdateOwnPublicKey` to `server/internal/router/auth_user_test.go`.

```go
func TestUserCanUpdateOwnPublicKey(t *testing.T) {
    r := newTestRouter(t)
    token := registerAndLogin(t, r, "alice")

    req := httptest.NewRequest(http.MethodPut, "/api/users/me/public-key", bytes.NewReader([]byte(`{"publicKey":"alice-public-key","algorithm":"rsa-oaep-sha256"}`)))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
    }

    var resp struct {
        PublicKey          string `json:"publicKey"`
        PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
    }
    if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
        t.Fatalf("expected valid json, got error: %v", err)
    }
    if resp.PublicKey != "alice-public-key" || resp.PublicKeyAlgorithm != "rsa-oaep-sha256" {
        t.Fatalf("expected saved public key fields, got %#v", resp)
    }
}
```

- [ ] **Step 2: Run the auth router test to verify it fails**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run TestUserCanUpdateOwnPublicKey
```

Expected: FAIL with 404 or missing method/field behavior.

- [ ] **Step 3: Add repository support for updating public keys**

First extend the interface in `server/internal/repository/user_memory.go`:

```go
type UserRepository interface {
    Create(user *model.User) error
    FindByUsername(username string) (*model.User, error)
    FindByID(id uint64) (*model.User, error)
    UpdateProfile(userID uint64, nickname string, gender int, signature string) (*model.User, error)
    UpdatePublicKey(userID uint64, publicKey string, algorithm string) (*model.User, error)
}
```

Add the in-memory implementation in the same file:

```go
func (r *InMemoryUserRepository) UpdatePublicKey(userID uint64, publicKey string, algorithm string) (*model.User, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    user, ok := r.users[userID]
    if !ok {
        return nil, ErrUserNotFound
    }

    user.PublicKey = publicKey
    user.PublicKeyAlgorithm = algorithm

    copyUser := *user
    return &copyUser, nil
}
```

Then add the GORM implementation body in `server/internal/repository/user_gorm.go`:

```go
func (r *GormUserRepository) UpdatePublicKey(userID uint64, publicKey string, algorithm string) (*model.User, error) {
    updates := map[string]interface{}{
        "public_key":           publicKey,
        "public_key_algorithm": algorithm,
    }
    result := r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates)
    if result.Error != nil {
        return nil, result.Error
    }
    if result.RowsAffected == 0 {
        return nil, ErrUserNotFound
    }
    return r.FindByID(userID)
}
```

- [ ] **Step 4: Add auth service method for public-key update**

In `server/internal/service/auth.go`, add:

```go
type UpdatePublicKeyInput struct {
    PublicKey string
    Algorithm string
}

func (s *AuthService) UpdatePublicKey(userID uint64, input UpdatePublicKeyInput) (*model.User, error) {
    if input.PublicKey == "" {
        return nil, fmt.Errorf("public key is required")
    }
    if input.Algorithm == "" {
        return nil, fmt.Errorf("algorithm is required")
    }
    return s.repo.UpdatePublicKey(userID, input.PublicKey, input.Algorithm)
}
```

- [ ] **Step 5: Add the authenticated handler and route**

In `server/internal/handler/auth_user.go`, add request type and handler:

```go
type updatePublicKeyRequest struct {
    PublicKey string `json:"publicKey"`
    Algorithm string `json:"algorithm"`
}

func (h *AuthUserHandler) UpdateMyPublicKey(c *gin.Context) {
    var req updatePublicKeyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
        return
    }

    user, err := h.authService.UpdatePublicKey(c.MustGet("userID").(uint64), service.UpdatePublicKeyInput{
        PublicKey: req.PublicKey,
        Algorithm: req.Algorithm,
    })
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    c.JSON(http.StatusOK, user)
}
```

Wire it in `server/internal/router/router.go`:

```go
authed.PUT("/users/me/public-key", authHandler.UpdateMyPublicKey)
```

- [ ] **Step 6: Run the test to verify it passes**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run TestUserCanUpdateOwnPublicKey
```

Expected: PASS.

- [ ] **Step 7: Commit the public-key API task**

```bash
cd /Users/zyb/bishe/whr-im && git add server/internal/repository/user_memory.go server/internal/repository/user_gorm.go server/internal/service/auth.go server/internal/handler/auth_user.go server/internal/router/router.go server/internal/router/auth_user_test.go && git commit -m "feat: add public key update api"
```

### Task 3: Expose friend public keys and enforce ciphertext-only message creation

**Files:**
- Modify: `server/internal/service/friend.go:24-107`
- Modify: `server/internal/service/message.go:11-53`
- Modify: `server/internal/handler/message.go:12-51`
- Modify: `server/internal/router/message_test.go:11-87`
- Modify: `server/internal/router/friend_test.go`
- Modify: `server/internal/router/websocket_test.go:15-139`

- [ ] **Step 1: Write failing tests for friend-list public keys and ciphertext-only messaging**

In `server/internal/router/friend_test.go`, extend the friend list response struct to include:

```go
PublicKey          string `json:"publicKey"`
PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
```

Assert that after Bob uploads a public key, Alice’s friend list includes Bob’s key fields.

In `server/internal/router/message_test.go`, change message creation bodies to:

```go
createMessage(t, r, aliceToken, `{"receiverId":2,"ciphertext":"cipher-1","algorithm":"rsa-oaep-sha256"}`)
createMessage(t, r, bobToken, `{"receiverId":1,"ciphertext":"cipher-2","algorithm":"rsa-oaep-sha256"}`)
```

and assert:

```go
if historyResp[0].Ciphertext != "cipher-1" || historyResp[1].Ciphertext != "cipher-2" {
    t.Fatalf("expected ciphertext messages in order, got %#v", historyResp)
}
```

In `server/internal/router/websocket_test.go`, send ciphertext and assert delivered `ciphertext`/`algorithm` instead of `content`.

- [ ] **Step 2: Run the three router tests to verify they fail**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run 'TestFriendRequestAcceptAndFriendsListFlow|TestMessageHistoryReturnsConversationInTimeOrder|TestWebSocketDeliversChatMessageToOnlineFriend'
```

Expected: FAIL because friend DTO lacks public key fields and message handlers/services still expect plaintext `content`.

- [ ] **Step 3: Extend the friend DTO with public-key fields**

In `server/internal/service/friend.go`:

```go
type FriendListItem struct {
    UserID              uint64 `json:"userId"`
    FriendID            uint64 `json:"friendId"`
    Nickname            string `json:"nickname"`
    Avatar              string `json:"avatar"`
    Signature           string `json:"signature"`
    PublicKey           string `json:"publicKey"`
    PublicKeyAlgorithm  string `json:"publicKeyAlgorithm"`
}
```

When appending items:

```go
items = append(items, FriendListItem{
    UserID:             friend.UserID,
    FriendID:           friend.FriendID,
    Nickname:           profile.Nickname,
    Avatar:             profile.Avatar,
    Signature:          profile.Signature,
    PublicKey:          profile.PublicKey,
    PublicKeyAlgorithm: profile.PublicKeyAlgorithm,
})
```

- [ ] **Step 4: Switch the message handler/service to ciphertext input**

In `server/internal/handler/message.go`:

```go
type createMessageRequest struct {
    ReceiverID uint64 `json:"receiverId"`
    Ciphertext string `json:"ciphertext"`
    Algorithm  string `json:"algorithm"`
}
```

```go
message, err := h.messageService.Create(c.MustGet("userID").(uint64), service.CreateMessageInput{
    ReceiverID: req.ReceiverID,
    Ciphertext: req.Ciphertext,
    Algorithm:  req.Algorithm,
})
```

In `server/internal/service/message.go`:

```go
type CreateMessageInput struct {
    ReceiverID uint64
    Ciphertext string
    Algorithm  string
}
```

Validate both required fields and create the model using ciphertext:

```go
if input.Ciphertext == "" {
    return nil, fmt.Errorf("ciphertext is required")
}
if input.Algorithm == "" {
    return nil, fmt.Errorf("algorithm is required")
}
message := &model.Message{
    SenderID:   userID,
    ReceiverID: input.ReceiverID,
    Ciphertext: input.Ciphertext,
    Algorithm:  input.Algorithm,
}
```

- [ ] **Step 5: Run the targeted router tests to verify they pass**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run 'TestFriendRequestAcceptAndFriendsListFlow|TestMessageHistoryReturnsConversationInTimeOrder|TestWebSocketDeliversChatMessageToOnlineFriend'
```

Expected: PASS.

- [ ] **Step 6: Commit the ciphertext-only backend behavior**

```bash
cd /Users/zyb/bishe/whr-im && git add server/internal/service/friend.go server/internal/service/message.go server/internal/handler/message.go server/internal/router/friend_test.go server/internal/router/message_test.go server/internal/router/websocket_test.go && git commit -m "feat: store and deliver ciphertext messages"
```

### Task 4: Add browser-local E2EE utilities with test coverage

**Files:**
- Create: `web/src/utils/e2ee.ts`
- Create: `web/src/utils/e2ee.js`
- Create: `web/src/utils/e2ee.test.mjs`

- [ ] **Step 1: Write the failing frontend utility tests**

Create `web/src/utils/e2ee.test.mjs` with tests that cover storage-key naming and fallback behavior that can run in Node.

```js
import test from 'node:test'
import assert from 'node:assert/strict'

import {
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
})

test('normalizes encrypted payloads for ui rendering', () => {
  assert.deepEqual(
    normalizeEncryptedPayload({ ciphertext: 'abc', algorithm: 'rsa-oaep-sha256' }),
    { ciphertext: 'abc', algorithm: 'rsa-oaep-sha256' }
  )
})
```

- [ ] **Step 2: Run the node test to verify it fails**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && node --test src/utils/e2ee.test.mjs
```

Expected: FAIL because `e2ee.js` does not exist yet.

- [ ] **Step 3: Add minimal shared helper implementations**

Create `web/src/utils/e2ee.js`:

```js
export function buildPrivateKeyStorageKey(userId) {
  return `e2ee-private-key:${userId}`
}

export function maskEncryptedMessage(value) {
  return value ? value : '***（已加密）'
}

export function normalizeEncryptedPayload(payload) {
  return {
    ciphertext: payload?.ciphertext || '',
    algorithm: payload?.algorithm || ''
  }
}
```

Create `web/src/utils/e2ee.ts` with the same helpers plus browser-only Web Crypto functions:

```ts
const ALGORITHM = 'RSA-OAEP'
const HASH = 'SHA-256'
const DEFAULT_MESSAGE_ALGORITHM = 'rsa-oaep-sha256'

export function buildPrivateKeyStorageKey(userId: number) {
  return `e2ee-private-key:${userId}`
}

export function maskEncryptedMessage(value?: string | null) {
  return value ? value : '***（已加密）'
}

export function normalizeEncryptedPayload(payload?: { ciphertext?: string; algorithm?: string }) {
  return {
    ciphertext: payload?.ciphertext || '',
    algorithm: payload?.algorithm || ''
  }
}
```

Also include browser functions in `e2ee.ts`:

```ts
function arrayBufferToBase64(buffer: ArrayBuffer) {
  return btoa(String.fromCharCode(...new Uint8Array(buffer)))
}

function base64ToArrayBuffer(value: string) {
  const binary = atob(value)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i)
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
  return window.crypto.subtle.importKey('spki', base64ToArrayBuffer(serializedKey), { name: ALGORITHM, hash: HASH }, true, ['encrypt'])
}

export async function importPrivateKey(serializedKey: string) {
  return window.crypto.subtle.importKey('pkcs8', base64ToArrayBuffer(serializedKey), { name: ALGORITHM, hash: HASH }, true, ['decrypt'])
}

export async function encryptMessage(publicKey: CryptoKey, content: string) {
  const encoded = new TextEncoder().encode(content)
  const ciphertext = await window.crypto.subtle.encrypt({ name: ALGORITHM }, publicKey, encoded)
  return { ciphertext: arrayBufferToBase64(ciphertext), algorithm: DEFAULT_MESSAGE_ALGORITHM }
}

export async function decryptMessage(privateKey: CryptoKey, ciphertext: string) {
  const decrypted = await window.crypto.subtle.decrypt({ name: ALGORITHM }, privateKey, base64ToArrayBuffer(ciphertext))
  return new TextDecoder().decode(decrypted)
}

export function savePrivateKey(userId: number, serializedKey: string) {
  localStorage.setItem(buildPrivateKeyStorageKey(userId), serializedKey)
}

export function loadPrivateKey(userId: number) {
  return localStorage.getItem(buildPrivateKeyStorageKey(userId)) || ''
}
```

- [ ] **Step 4: Run the node test to verify it passes**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && node --test src/utils/e2ee.test.mjs
```

Expected: PASS.

- [ ] **Step 5: Commit the utility layer**

```bash
cd /Users/zyb/bishe/whr-im && git add web/src/utils/e2ee.ts web/src/utils/e2ee.js web/src/utils/e2ee.test.mjs && git commit -m "feat: add frontend e2ee utilities"
```

### Task 5: Integrate key generation, upload, encryption, decryption, and masked fallback in chat UI

**Files:**
- Modify: `web/src/stores/auth.ts:5-52`
- Modify: `web/src/views/ChatView.vue:1-364`
- Test: `web/src/utils/e2ee.test.mjs`
- Test: `web/src/utils/chat-time.test.mjs`

- [ ] **Step 1: Add the failing UI-adjacent helper test for masked rendering**

Extend `web/src/utils/e2ee.test.mjs`:

```js
test('masks encrypted messages when decrypted body is missing but leaves ciphertext metadata untouched', () => {
  const payload = normalizeEncryptedPayload({ ciphertext: 'cipher', algorithm: 'rsa-oaep-sha256' })
  assert.equal(maskEncryptedMessage(''), '***（已加密）')
  assert.equal(payload.ciphertext, 'cipher')
  assert.equal(payload.algorithm, 'rsa-oaep-sha256')
})
```

- [ ] **Step 2: Run the helper tests to verify they still pass before UI integration**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && node --test src/utils/e2ee.test.mjs src/utils/chat-time.test.mjs
```

Expected: PASS. This guards the fallback helpers before wiring them into Vue.

- [ ] **Step 3: Update auth typing to carry public key fields**

In `web/src/stores/auth.ts`:

```ts
export interface CurrentUser {
  id?: number
  username: string
  nickname: string
  avatar: string
  gender?: number
  signature?: string
  publicKey?: string
  publicKeyAlgorithm?: string
}
```

- [ ] **Step 4: Refactor `ChatView.vue` to work with encrypted payloads**

Note before editing: keep `RenderMessage.content` as a purely UI-derived field. Do not send it back to the backend, and do not reintroduce plaintext `content` into any request payload.

Replace the local interfaces with:

```ts
interface FriendItem {
  userId: number
  friendId: number
  nickname: string
  avatar: string
  signature: string
  publicKey?: string
  publicKeyAlgorithm?: string
}

interface ChatMessage {
  id?: number
  senderId: number
  receiverId: number
  ciphertext: string
  algorithm: string
  createdAt?: string
}

interface RenderMessage extends ChatMessage {
  content: string
}
```

Add imports from `../utils/e2ee`:

```ts
import {
  decryptMessage,
  encryptMessage,
  exportPrivateKey,
  exportPublicKey,
  generateKeyPair,
  importPrivateKey,
  importPublicKey,
  loadPrivateKey,
  maskEncryptedMessage,
  savePrivateKey
} from '../utils/e2ee'
```

Replace message state and add private-key state:

```ts
const messages = ref<RenderMessage[]>([])
const privateKey = ref<CryptoKey | null>(null)
```

Add helper to ensure the current user has a local keypair and uploaded public key:

```ts
async function ensureOwnKeyPair() {
  if (!authStore.user?.id) return

  const storedPrivateKey = loadPrivateKey(authStore.user.id)
  if (storedPrivateKey) {
    privateKey.value = await importPrivateKey(storedPrivateKey)
    return
  }

  const keyPair = await generateKeyPair()
  const serializedPublicKey = await exportPublicKey(keyPair.publicKey)
  const serializedPrivateKey = await exportPrivateKey(keyPair.privateKey)
  savePrivateKey(authStore.user.id, serializedPrivateKey)
  privateKey.value = keyPair.privateKey
  await http.put('/users/me/public-key', {
    publicKey: serializedPublicKey,
    algorithm: 'rsa-oaep-sha256'
  })
  authStore.user = {
    ...authStore.user,
    publicKey: serializedPublicKey,
    publicKeyAlgorithm: 'rsa-oaep-sha256'
  }
}
```

Add a decode helper for history/WebSocket payloads:

```ts
async function toRenderMessage(message: ChatMessage): Promise<RenderMessage> {
  if (!privateKey.value) {
    return { ...message, content: maskEncryptedMessage('') }
  }

  try {
    const content = await decryptMessage(privateKey.value, message.ciphertext)
    return { ...message, content }
  } catch {
    return { ...message, content: maskEncryptedMessage('') }
  }
}
```

Change `loadMessages` to decrypt results before assignment:

```ts
const { data } = await http.get(`/messages?friendId=${currentFriendId.value}`)
messages.value = await Promise.all((data as ChatMessage[]).map(toRenderMessage))
```

Change `sendMessage` to block on missing friend public key or local private key and send ciphertext only:

```ts
if (!privateKey.value) {
  errorMessage.value = '当前设备没有可用私钥'
  return
}
const friend = currentFriend.value
if (!friend?.publicKey) {
  errorMessage.value = '对方未启用端到端加密消息'
  return
}
const publicKey = await importPublicKey(friend.publicKey)
const encrypted = await encryptMessage(publicKey, draft.value.trim())
const { data } = await http.post('/messages', {
  receiverId: currentFriendId.value,
  ciphertext: encrypted.ciphertext,
  algorithm: encrypted.algorithm
})
messages.value.push(await toRenderMessage(data as ChatMessage))
```

Change WebSocket handling to decrypt the incoming ciphertext:

```ts
const chatMessage = payload.data as ChatMessage
...
messages.value.push(await toRenderMessage(chatMessage))
```

Call `ensureOwnKeyPair()` during mount after `authStore.bootstrap()` and before `loadFriends()`.

- [ ] **Step 5: Keep the existing message template but render decrypted-or-masked `content`**

The existing template line should remain:

```vue
<p>{{ message.content }}</p>
```

because `RenderMessage.content` is now derived locally from ciphertext.

- [ ] **Step 6: Run frontend tests and build**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && node --test src/utils/e2ee.test.mjs src/utils/chat-time.test.mjs && npm run build
```

Expected: PASS for both node tests and Vite build.

- [ ] **Step 7: Commit the frontend integration**

```bash
cd /Users/zyb/bishe/whr-im && git add web/src/stores/auth.ts web/src/views/ChatView.vue web/src/utils/e2ee.ts web/src/utils/e2ee.js web/src/utils/e2ee.test.mjs && git commit -m "feat: enable chat e2ee in frontend"
```

### Task 6: Run end-to-end verification for ciphertext storage and masked fallback

**Files:**
- Verify: `init.sql`
- Verify: `server/internal/router/auth_user_test.go`
- Verify: `server/internal/router/friend_test.go`
- Verify: `server/internal/router/message_test.go`
- Verify: `server/internal/router/websocket_test.go`
- Verify: `web/src/views/ChatView.vue`

- [ ] **Step 1: Run the full backend router regression suite**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router
```

Expected: PASS.

- [ ] **Step 2: Run the frontend utility tests and build again from a clean checkpoint**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && node --test src/utils/gender.test.mjs src/utils/chat-time.test.mjs src/utils/e2ee.test.mjs && npm run build
```

Expected: PASS.

- [ ] **Step 3: Recreate the database from `init.sql`**

Run with your local MySQL setup, for example:
```bash
mysql -u root -p < /Users/zyb/bishe/whr-im/init.sql
```

Expected: tables are recreated with `users.public_key`, `users.public_key_algorithm`, `messages.ciphertext`, and `messages.algorithm`.

- [ ] **Step 4: Manual end-to-end verification with two users**

Check this sequence:
1. Register/login A and B
2. Ensure both users visit the app once so each uploads a public key
3. Make A and B friends
4. Send A → B message
5. Confirm database row in `messages` stores ciphertext instead of plaintext
6. Confirm B can read the decrypted message in chat UI
7. Clear B browser local storage for the private key
8. Reload chat history and confirm the same message body shows `***（已加密）`
9. Confirm that message time still shows using the existing formatter

- [ ] **Step 5: Commit any final test-only adjustments if needed**

```bash
cd /Users/zyb/bishe/whr-im && git status
```

Expected: clean working tree. If not clean because of deliberate test or fixture adjustments, stage only those files and create a final commit.

---

## Self-review checklist
- Spec coverage: included schema rebuild, public-key storage, ciphertext-only transport/storage, local decrypt/masked fallback, and end-to-end verification.
- Placeholder scan: removed TBD/TODO language; every task has concrete files, code, and commands.
- Type consistency: plan consistently uses `publicKey`, `publicKeyAlgorithm`, `ciphertext`, and `algorithm` across backend and frontend.
