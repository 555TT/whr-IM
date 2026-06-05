# Chat / Encrypted Chat Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the current unified message center into a normal chat page at `/chat` and a dedicated encrypted chat page at `/encrypted-chat`, while preserving existing encrypted history and moving AI/favorites to the normal page only.

**Architecture:** Keep the current encrypted message stack as-is and move it into a new `EncryptedChatView.vue`. Add a parallel non-encrypted friend/group message stack on the backend and point `ChatView.vue` at that new normal-message API. Reuse conversation list and message rendering patterns where useful, but do not force normal and encrypted business logic back into one component.

**Tech Stack:** Vue 3 + TypeScript + Vue Router + Pinia + Axios + WebSocket on the frontend; Go + Gin + GORM + existing websocket hub on the backend.

---

## File Structure Map

### Existing files to modify
- `web/src/router/index.ts` — register `/encrypted-chat` route.
- `web/src/views/ChatView.vue` — convert from encrypted chat page into normal message center.
- `web/src/utils/websocket.ts` — keep shared socket creation, no protocol changes here unless event naming helper is added.
- `server/internal/router/router.go` — register new normal message endpoints while keeping encrypted endpoints untouched.
- `server/internal/router/message_test.go` — add router tests for normal direct messages.
- `server/internal/router/group_test.go` — add router tests for normal group messages.
- `server/internal/router/websocket_test.go` — extend websocket coverage for new normal message events.
- `server/internal/router/auth_user_test.go` or related existing test helpers — only if constructor wiring changes require shared setup updates.
- `server/internal/router/router.go:79-94` and `server/internal/router/router.go:96-235` — constructor and route graph expansion.
- `server/internal/repository/group_gorm.go` / `server/internal/repository/group_memory.go` if extra group membership helpers are needed for normal group message authorization.

### New frontend files to create
- `web/src/views/EncryptedChatView.vue` — dedicated encrypted chat page using current E2EE logic and deep dark visual theme.
- `web/src/components/chat/ConversationSidebar.vue` — optional display-only sidebar component shared by normal/encrypted pages.
- `web/src/components/chat/MessageList.vue` — optional display-only message list component.
- `web/src/components/chat/MessageComposer.vue` — optional display-only composer shell.
- `web/src/api/normalChat.ts` — optional API wrapper for normal direct/group chat endpoints.

### New backend files to create
- `server/internal/model/normal_message.go` — non-encrypted direct message model.
- `server/internal/model/normal_group_message.go` — non-encrypted group message model.
- `server/internal/repository/normal_message_gorm.go` — GORM repo for normal direct messages.
- `server/internal/repository/normal_group_message_gorm.go` — GORM repo for normal group messages.
- `server/internal/service/normal_message.go` — business logic for normal direct chat.
- `server/internal/service/normal_group_message.go` — business logic for normal group chat.
- `server/internal/handler/normal_message.go` — HTTP handlers for normal direct chat.
- `server/internal/handler/normal_group_message.go` — HTTP handlers for normal group chat.

### Existing backend files to read carefully before implementation
- `server/internal/model/message.go` — encrypted direct message shape.
- `server/internal/model/group.go` — encrypted group message shape and membership models.
- `server/internal/service/message.go` — current encrypted direct message logic and websocket event naming.
- `server/internal/service/group_message.go` — current encrypted group message logic and membership validation.
- `server/internal/handler/message.go` — current encrypted direct message handler shape.
- `server/internal/handler/group_message.go` — current encrypted group message handler shape.
- `server/internal/router/router.go` — current registration points.
- `web/src/views/ChatView.vue` — current mixed page to extract encrypted behavior from.

---

### Task 1: Add failing backend tests for normal direct messages

**Files:**
- Modify: `server/internal/router/message_test.go`
- Read: `server/internal/router/router.go:159-175`
- Read: `server/internal/handler/message.go:20-60`

- [ ] **Step 1: Write the failing test**

Add a new router test that proves `/api/normal-messages` exists, persists plaintext content, and is isolated from encrypted `/api/messages`.

```go
func TestNormalDirectMessageSendAndHistoryFlow(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken, 2)

	sendReq := httptest.NewRequest(http.MethodPost, "/api/normal-messages", bytes.NewReader([]byte(`{"receiverId":2,"content":"hello bob"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusCreated {
		t.Fatalf("expected normal direct send 201, got %d with body %s", sendW.Code, sendW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/normal-messages?friendId=2", nil)
	listReq.Header.Set("Authorization", "Bearer "+aliceToken)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected normal direct list 200, got %d with body %s", listW.Code, listW.Body.String())
	}

	var listResp []struct {
		SenderID   uint64 `json:"senderId"`
		ReceiverID uint64 `json:"receiverId"`
		Content    string `json:"content"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("expected valid normal direct list json, got error: %v", err)
	}
	if len(listResp) != 1 || listResp[0].Content != "hello bob" {
		t.Fatalf("expected one plaintext normal direct message, got %+v", listResp)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run TestNormalDirectMessageSendAndHistoryFlow -v`

Expected: FAIL with 404 route-not-found or missing symbol errors because `/api/normal-messages` is not implemented yet.

- [ ] **Step 3: Write a second failing authorization test**

Add a second test in the same file to prove non-friends cannot use the normal direct chat endpoint.

```go
func TestNormalDirectMessageRejectsNonFriendUsers(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")
	_ = registerAndLogin(t, r, "bobby")

	sendReq := httptest.NewRequest(http.MethodPost, "/api/normal-messages", bytes.NewReader([]byte(`{"receiverId":2,"content":"hello"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusBadRequest {
		t.Fatalf("expected non-friend normal direct send 400, got %d with body %s", sendW.Code, sendW.Body.String())
	}
}
```

- [ ] **Step 4: Run the focused tests to verify they fail**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run 'TestNormalDirectMessage(SendAndHistoryFlow|RejectsNonFriendUsers)' -v`

Expected: FAIL for the new normal-message route coverage.

- [ ] **Step 5: Commit the red tests**

```bash
git add server/internal/router/message_test.go
git commit -m "test: add normal direct chat router coverage"
```

---

### Task 2: Implement normal direct message backend

**Files:**
- Create: `server/internal/model/normal_message.go`
- Create: `server/internal/repository/normal_message_gorm.go`
- Create: `server/internal/service/normal_message.go`
- Create: `server/internal/handler/normal_message.go`
- Modify: `server/internal/router/router.go`
- Test: `server/internal/router/message_test.go`

- [ ] **Step 1: Add the normal direct message model**

Create `server/internal/model/normal_message.go`.

```go
package model

import "time"

type NormalMessage struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	SenderID   uint64    `gorm:"not null;index" json:"senderId"`
	ReceiverID uint64    `gorm:"not null;index" json:"receiverId"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (NormalMessage) TableName() string {
	return "normal_messages"
}
```

- [ ] **Step 2: Add the GORM repository**

Create `server/internal/repository/normal_message_gorm.go`.

```go
package repository

import (
	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type NormalMessageRepository interface {
	Create(message *model.NormalMessage) error
	ListConversation(userID uint64, friendID uint64) ([]model.NormalMessage, error)
}

type GormNormalMessageRepository struct {
	db *gorm.DB
}

func NewGormNormalMessageRepository(db *gorm.DB) (*GormNormalMessageRepository, error) {
	if err := db.AutoMigrate(&model.NormalMessage{}); err != nil {
		return nil, err
	}
	return &GormNormalMessageRepository{db: db}, nil
}

func (r *GormNormalMessageRepository) Create(message *model.NormalMessage) error {
	return r.db.Create(message).Error
}

func (r *GormNormalMessageRepository) ListConversation(userID uint64, friendID uint64) ([]model.NormalMessage, error) {
	var messages []model.NormalMessage
	err := r.db.
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, friendID, friendID, userID).
		Order("created_at asc, id asc").
		Find(&messages).Error
	return messages, err
}
```

- [ ] **Step 3: Add the service**

Create `server/internal/service/normal_message.go`.

```go
package service

import (
	"fmt"
	"strings"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/ws"
)

type NormalMessageService struct {
	messageRepo repository.NormalMessageRepository
	friendRepo  repository.FriendRepository
	hub         *ws.Hub
}

type CreateNormalMessageInput struct {
	ReceiverID uint64
	Content    string
}

func NewNormalMessageService(messageRepo repository.NormalMessageRepository, friendRepo repository.FriendRepository, hub *ws.Hub) *NormalMessageService {
	return &NormalMessageService{messageRepo: messageRepo, friendRepo: friendRepo, hub: hub}
}

func (s *NormalMessageService) Create(userID uint64, input CreateNormalMessageInput) (*model.NormalMessage, error) {
	friends, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	isFriend := false
	for _, friend := range friends {
		if friend.FriendID == input.ReceiverID {
			isFriend = true
			break
		}
	}
	if !isFriend {
		return nil, fmt.Errorf("non-friend users cannot chat")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	message := &model.NormalMessage{
		SenderID:   userID,
		ReceiverID: input.ReceiverID,
		Content:    content,
	}
	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}
	if s.hub != nil {
		_ = s.hub.Send(input.ReceiverID, "normal_chat_message", message)
	}
	return message, nil
}

func (s *NormalMessageService) ListConversation(userID uint64, friendID uint64) ([]model.NormalMessage, error) {
	return s.messageRepo.ListConversation(userID, friendID)
}
```

- [ ] **Step 4: Add the handler**

Create `server/internal/handler/normal_message.go`.

```go
package handler

import (
	"net/http"
	"strconv"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type NormalMessageHandler struct {
	messageService *service.NormalMessageService
}

type createNormalMessageRequest struct {
	ReceiverID uint64 `json:"receiverId"`
	Content    string `json:"content"`
}

func NewNormalMessageHandler(messageService *service.NormalMessageService) *NormalMessageHandler {
	return &NormalMessageHandler{messageService: messageService}
}

func (h *NormalMessageHandler) Create(c *gin.Context) {
	var req createNormalMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	message, err := h.messageService.Create(c.MustGet("userID").(uint64), service.CreateNormalMessageInput{
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *NormalMessageHandler) List(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Query("friendId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid friendId"})
		return
	}
	messages, err := h.messageService.ListConversation(c.MustGet("userID").(uint64), friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}
```

- [ ] **Step 5: Wire the repository and routes**

Modify `server/internal/router/router.go` in three places.

Add repository construction in `New`:

```go
normalMessageRepo, err := repository.NewGormNormalMessageRepository(db)
if err != nil {
	log.Fatal(err)
}
```

Expand `NewWithRepositories` and `newEngine` signatures:

```go
	normalMessageRepo repository.NormalMessageRepository,
```

Register routes inside the `friendRepo != nil` block:

```go
if normalMessageRepo != nil {
	normalMessageService := service.NewNormalMessageService(normalMessageRepo, friendRepo, hub)
	normalMessageHandler := handler.NewNormalMessageHandler(normalMessageService)
	authed.POST("/normal-messages", normalMessageHandler.Create)
	authed.GET("/normal-messages", normalMessageHandler.List)
}
```

Keep the existing encrypted `/messages` routes unchanged.

- [ ] **Step 6: Run the focused tests to verify they pass**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run 'TestNormalDirectMessage(SendAndHistoryFlow|RejectsNonFriendUsers)' -v`

Expected: PASS.

- [ ] **Step 7: Commit the green implementation**

```bash
git add server/internal/model/normal_message.go server/internal/repository/normal_message_gorm.go server/internal/service/normal_message.go server/internal/handler/normal_message.go server/internal/router/router.go server/internal/router/message_test.go
git commit -m "feat: add normal direct chat backend"
```

---

### Task 3: Add failing backend tests for normal group messages

**Files:**
- Modify: `server/internal/router/group_test.go`
- Read: `server/internal/router/router.go:188-209`
- Read: `server/internal/model/group.go:27-50`

- [ ] **Step 1: Write the failing normal group send/list test**

Add a test to `server/internal/router/group_test.go`.

```go
func TestNormalGroupMessageSendAndHistoryFlow(t *testing.T) {
	r := newTestRouter(t)
	ownerToken := registerAndLogin(t, r, "alice")
	memberToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, ownerToken, memberToken, 2)
	groupID := createGroup(t, r, ownerToken, "Normal Group", []int{2})
	_ = memberToken

	sendReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%d/normal-messages", groupID), bytes.NewReader([]byte(`{"content":"plain hello group"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+ownerToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusCreated {
		t.Fatalf("expected normal group send 201, got %d with body %s", sendW.Code, sendW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/groups/%d/normal-messages", groupID), nil)
	listReq.Header.Set("Authorization", "Bearer "+ownerToken)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected normal group list 200, got %d with body %s", listW.Code, listW.Body.String())
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run TestNormalGroupMessageSendAndHistoryFlow -v`

Expected: FAIL with missing route or missing implementation.

- [ ] **Step 3: Add a failing membership test**

Add another test to ensure non-members cannot send normal group messages.

```go
func TestNormalGroupMessageRejectsNonMember(t *testing.T) {
	r := newTestRouter(t)
	ownerToken := registerAndLogin(t, r, "alice")
	outsiderToken := registerAndLogin(t, r, "bobby")
	groupID := createGroup(t, r, ownerToken, "Normal Group", []int{})

	sendReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%d/normal-messages", groupID), bytes.NewReader([]byte(`{"content":"should fail"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+outsiderToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusBadRequest {
		t.Fatalf("expected normal group non-member send 400, got %d with body %s", sendW.Code, sendW.Body.String())
	}
}
```

- [ ] **Step 4: Run the focused tests to verify they fail**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run 'TestNormalGroupMessage(SendAndHistoryFlow|RejectsNonMember)' -v`

Expected: FAIL.

- [ ] **Step 5: Commit the red tests**

```bash
git add server/internal/router/group_test.go
git commit -m "test: add normal group chat router coverage"
```

---

### Task 4: Implement normal group message backend

**Files:**
- Create: `server/internal/model/normal_group_message.go`
- Create: `server/internal/repository/normal_group_message_gorm.go`
- Create: `server/internal/service/normal_group_message.go`
- Create: `server/internal/handler/normal_group_message.go`
- Modify: `server/internal/router/router.go`
- Test: `server/internal/router/group_test.go`

- [ ] **Step 1: Add the normal group message model**

Create `server/internal/model/normal_group_message.go`.

```go
package model

import "time"

type NormalGroupMessage struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	GroupID   uint64    `gorm:"not null;index:idx_normal_group_messages_group_created_at,priority:1" json:"groupId"`
	SenderID  uint64    `gorm:"not null;index" json:"senderId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"index:idx_normal_group_messages_group_created_at,priority:2" json:"createdAt"`
}

func (NormalGroupMessage) TableName() string {
	return "normal_group_messages"
}
```

- [ ] **Step 2: Add the repository**

Create `server/internal/repository/normal_group_message_gorm.go`.

```go
package repository

import (
	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type NormalGroupMessageRepository interface {
	Create(message *model.NormalGroupMessage) error
	List(groupID uint64) ([]model.NormalGroupMessage, error)
}

type GormNormalGroupMessageRepository struct {
	db *gorm.DB
}

func NewGormNormalGroupMessageRepository(db *gorm.DB) (*GormNormalGroupMessageRepository, error) {
	if err := db.AutoMigrate(&model.NormalGroupMessage{}); err != nil {
		return nil, err
	}
	return &GormNormalGroupMessageRepository{db: db}, nil
}

func (r *GormNormalGroupMessageRepository) Create(message *model.NormalGroupMessage) error {
	return r.db.Create(message).Error
}

func (r *GormNormalGroupMessageRepository) List(groupID uint64) ([]model.NormalGroupMessage, error) {
	var messages []model.NormalGroupMessage
	err := r.db.Where("group_id = ?", groupID).Order("created_at asc, id asc").Find(&messages).Error
	return messages, err
}
```

- [ ] **Step 3: Add the service**

Create `server/internal/service/normal_group_message.go`.

```go
package service

import (
	"fmt"
	"strings"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/ws"
)

type NormalGroupMessageService struct {
	messageRepo repository.NormalGroupMessageRepository
	groupRepo   repository.GroupRepository
	hub         *ws.Hub
}

type CreateNormalGroupMessageInput struct {
	GroupID  uint64
	Content  string
}

func NewNormalGroupMessageService(messageRepo repository.NormalGroupMessageRepository, groupRepo repository.GroupRepository, hub *ws.Hub) *NormalGroupMessageService {
	return &NormalGroupMessageService{messageRepo: messageRepo, groupRepo: groupRepo, hub: hub}
}

func (s *NormalGroupMessageService) Create(userID uint64, input CreateNormalGroupMessageInput) (*model.NormalGroupMessage, error) {
	detail, err := s.groupRepo.FindByID(input.GroupID)
	if err != nil {
		return nil, err
	}
	isMember := false
	for _, member := range detail.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, fmt.Errorf("only group members can send messages")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	message := &model.NormalGroupMessage{
		GroupID:  input.GroupID,
		SenderID: userID,
		Content:  content,
	}
	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}
	if s.hub != nil {
		for _, member := range detail.Members {
			if member.UserID == userID {
				continue
			}
			_ = s.hub.Send(member.UserID, "normal_group_message", message)
		}
	}
	return message, nil
}

func (s *NormalGroupMessageService) List(userID uint64, groupID uint64) ([]model.NormalGroupMessage, error) {
	detail, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}
	for _, member := range detail.Members {
		if member.UserID == userID {
			return s.messageRepo.List(groupID)
		}
	}
	return nil, fmt.Errorf("only group members can view messages")
}
```

- [ ] **Step 4: Add the handler**

Create `server/internal/handler/normal_group_message.go`.

```go
package handler

import (
	"net/http"
	"strconv"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type NormalGroupMessageHandler struct {
	messageService *service.NormalGroupMessageService
}

type createNormalGroupMessageRequest struct {
	Content string `json:"content"`
}

func NewNormalGroupMessageHandler(messageService *service.NormalGroupMessageService) *NormalGroupMessageHandler {
	return &NormalGroupMessageHandler{messageService: messageService}
}

func (h *NormalGroupMessageHandler) Create(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	var req createNormalGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	message, err := h.messageService.Create(c.MustGet("userID").(uint64), service.CreateNormalGroupMessageInput{
		GroupID: groupID,
		Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *NormalGroupMessageHandler) List(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	messages, err := h.messageService.List(c.MustGet("userID").(uint64), groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}
```

- [ ] **Step 5: Register the normal group endpoints and repositories**

Modify `server/internal/router/router.go`.

Add repository construction in `New`:

```go
normalGroupMessageRepo, err := repository.NewGormNormalGroupMessageRepository(db)
if err != nil {
	log.Fatal(err)
}
```

Expand `NewWithRepositories` and `newEngine` signatures:

```go
	normalGroupMessageRepo repository.NormalGroupMessageRepository,
```

Register routes in the `/groups` block:

```go
if groupRepo != nil && normalGroupMessageRepo != nil {
	normalGroupMessageService := service.NewNormalGroupMessageService(normalGroupMessageRepo, groupRepo, hub)
	normalGroupMessageHandler := handler.NewNormalGroupMessageHandler(normalGroupMessageService)
	groups.POST("/:id/normal-messages", normalGroupMessageHandler.Create)
	groups.GET("/:id/normal-messages", normalGroupMessageHandler.List)
}
```

Keep encrypted `/groups/:id/messages` unchanged.

- [ ] **Step 6: Run the focused tests to verify they pass**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run 'TestNormalGroupMessage(SendAndHistoryFlow|RejectsNonMember)' -v`

Expected: PASS.

- [ ] **Step 7: Commit the implementation**

```bash
git add server/internal/model/normal_group_message.go server/internal/repository/normal_group_message_gorm.go server/internal/service/normal_group_message.go server/internal/handler/normal_group_message.go server/internal/router/router.go server/internal/router/group_test.go
git commit -m "feat: add normal group chat backend"
```

---

### Task 5: Add websocket regression coverage for normal chat events

**Files:**
- Modify: `server/internal/router/websocket_test.go`
- Read: `server/internal/service/message.go:31-74`
- Read: `server/internal/service/group_message.go`
- Read: `server/internal/ws/...` existing hub behavior

- [ ] **Step 1: Write the failing normal direct websocket test**

Add a websocket test that asserts the event type is `normal_chat_message` when a normal direct message is sent.

```go
func TestNormalDirectMessageBroadcastsNormalChatEvent(t *testing.T) {
	// Mirror the existing websocket test setup, but send POST /api/normal-messages
	// and assert envelope.Type == "normal_chat_message".
}
```

Use the same envelope assertions as the existing `chat_message` test, but check for plaintext `content` instead of ciphertext fields.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run TestNormalDirectMessageBroadcastsNormalChatEvent -v`

Expected: FAIL because the new event path is not yet fully covered or named incorrectly.

- [ ] **Step 3: Write the failing normal group websocket test**

```go
func TestNormalGroupMessageBroadcastsNormalGroupEvent(t *testing.T) {
	// Mirror existing group websocket coverage and assert envelope.Type == "normal_group_message".
}
```

- [ ] **Step 4: Run the focused websocket tests to verify they fail**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run 'TestNormal(DirectMessageBroadcastsNormalChatEvent|GroupMessageBroadcastsNormalGroupEvent)' -v`

Expected: FAIL.

- [ ] **Step 5: Make the minimal implementation adjustments**

If Task 2 and Task 4 already emit the new event names, keep this step limited to fixing mismatched payload or constructor wiring in the websocket test harness. Do not rename encrypted `chat_message` or `group_message` events.

Use these event names only:

```go
"normal_chat_message"
"normal_group_message"
```

- [ ] **Step 6: Run the focused websocket tests to verify they pass**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -run 'TestNormal(DirectMessageBroadcastsNormalChatEvent|GroupMessageBroadcastsNormalGroupEvent)' -v`

Expected: PASS.

- [ ] **Step 7: Commit the websocket coverage**

```bash
git add server/internal/router/websocket_test.go server/internal/service/normal_message.go server/internal/service/normal_group_message.go

git commit -m "test: cover normal chat websocket events"
```

---

### Task 6: Add frontend route and encrypted page shell

**Files:**
- Modify: `web/src/router/index.ts`
- Create: `web/src/views/EncryptedChatView.vue`
- Read: `web/src/views/ChatView.vue`
- Test: `web/src/views/EncryptedChatView.vue` via build

- [ ] **Step 1: Add the failing route import and route registration**

Modify `web/src/router/index.ts`.

```ts
import EncryptedChatView from '../views/EncryptedChatView.vue'
```

Add the route:

```ts
{ path: '/encrypted-chat', component: EncryptedChatView, meta: { requiresAuth: true } },
```

Create a temporary minimal page that proves routing and keeps the app compiling.

```vue
<script setup lang="ts">
import AppNav from '../components/AppNav.vue'
</script>

<template>
  <div class="page-shell apple-page encrypted-page-shell">
    <AppNav />
    <section class="card encrypted-placeholder-panel">
      <p class="apple-label">Encrypted Channel</p>
      <h1>加密通道</h1>
      <p>这里将承接现有端到端加密聊天能力。</p>
    </section>
  </div>
</template>

<style scoped>
.encrypted-page-shell {
  background: radial-gradient(circle at top right, rgba(34, 197, 94, 0.14), transparent 35%), #07110d;
  min-height: 100vh;
}

.encrypted-placeholder-panel {
  color: #d7ffe8;
  background: rgba(6, 20, 14, 0.88);
  box-shadow: inset 0 0 0 1px rgba(74, 222, 128, 0.16);
}
</style>
```

- [ ] **Step 2: Run the build to verify the route shell passes**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

Expected: PASS.

- [ ] **Step 3: Commit the route shell**

```bash
git add web/src/router/index.ts web/src/views/EncryptedChatView.vue

git commit -m "feat: add encrypted chat route shell"
```

---

### Task 7: Move the current encrypted chat logic into `EncryptedChatView.vue`

**Files:**
- Modify: `web/src/views/EncryptedChatView.vue`
- Read: `web/src/views/ChatView.vue:1-1359`
- Test: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

- [ ] **Step 1: Copy the current encrypted business logic into the new page**

Start from the current `web/src/views/ChatView.vue` encrypted logic, including these exact responsibilities:
- `ensureOwnKeyPair`
- `toRenderMessage`
- `toGroupRenderMessage`
- encrypted `loadFriendMessages`
- encrypted `loadGroupMessages`
- `sendFriendMessage`
- `sendGroupMessage`
- websocket handling for `chat_message` and `group_message`
- group creation / group info / invite / leave flow if those flows should remain available in the encrypted page

At this stage, do not refactor behavior. First preserve it in the new file.

- [ ] **Step 2: Remove non-encrypted-only UI from the encrypted page**

Delete these UI blocks from the encrypted page:
- AI 助手入口卡片
- 我的收藏入口卡片
- 收藏多选工具栏
- 普通消息中心文案

Delete these methods if they become unused:

```ts
openAIChat()
openFavorites()
favoriteSelectedMessages()
enterSelectionMode()
toggleSelection()
clearSelection()
```

- [ ] **Step 3: Apply the dedicated encrypted visual theme**

Update the page styles to a darker professional theme. Use this shell direction in the final file:

```css
.encrypted-theme-shell {
  position: relative;
  min-height: 100vh;
  background:
    radial-gradient(circle at top right, rgba(74, 222, 128, 0.12), transparent 28%),
    radial-gradient(circle at bottom left, rgba(45, 212, 191, 0.08), transparent 24%),
    linear-gradient(180deg, #030807 0%, #07110d 52%, #020504 100%);
}

.encrypted-shell {
  background: rgba(6, 18, 14, 0.86);
  border: 1px solid rgba(74, 222, 128, 0.14);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.45);
}

.encrypted-shell .sidebar,
.encrypted-shell .chat-panel,
.encrypted-shell .message-item,
.encrypted-shell .composer {
  color: #d9fbe8;
}
```

Keep the experience readable and restrained; do not add flashy animations.

- [ ] **Step 4: Update top-of-page copy to encrypted semantics**

Use labels like:

```vue
<p class="apple-label">Encrypted Channel</p>
<h2>加密通道</h2>
<p class="sidebar-banner-copy">仅展示具备端到端加密能力的安全会话。</p>
```

Use chat-top hints like:

```vue
<p class="apple-label">Secure Session</p>
<small class="muted">End-to-End Protected</small>
```

- [ ] **Step 5: Run the build to verify the extracted encrypted page passes**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

Expected: PASS.

- [ ] **Step 6: Commit the encrypted page extraction**

```bash
git add web/src/views/EncryptedChatView.vue

git commit -m "feat: move encrypted chat into dedicated page"
```

---

### Task 8: Filter encrypted page conversations to only secure-capable targets

**Files:**
- Modify: `web/src/views/EncryptedChatView.vue`
- Read: `web/src/views/ChatView.vue` friend/group types
- Test: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

- [ ] **Step 1: Add computed filters for secure-capable friends and groups**

In `EncryptedChatView.vue`, add friend filtering:

```ts
const encryptedFriends = computed(() =>
  friends.value.filter(
    (friend) => Boolean(friend.publicKey) && friend.publicKeyAlgorithm === E2EE_MESSAGE_ALGORITHM
  )
)
```

Add group filtering based on current group detail membership capabilities. Use a helper that only returns true if every member has a supported public key.

```ts
function groupSupportsEncryptedSession(detail: GroupDetail) {
  return detail.members.every(
    (member) => Boolean(member.publicKey) && member.publicKeyAlgorithm === E2EE_MESSAGE_ALGORITHM
  )
}
```

- [ ] **Step 2: Use the filtered lists in the template**

Replace sidebar loops so encrypted page uses:

```vue
v-for="friend in encryptedFriends"
```

and a filtered group list rather than `groups` directly.

If a user currently has no secure-capable conversations, show:

```vue
<div class="empty-state">暂无可用的加密会话，请先确保好友或群成员已开启加密能力。</div>
```

- [ ] **Step 3: Guard entry to unsupported conversations**

Even though unsupported targets are hidden, keep the runtime checks in send flow and list loading. Use these exact errors:

```ts
errorMessage.value = '对方未开启加密能力'
errorMessage.value = '该群存在未具备加密能力的成员'
errorMessage.value = '当前环境不支持安全会话'
```

- [ ] **Step 4: Run the build to verify the filtered page passes**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit the secure filtering**

```bash
git add web/src/views/EncryptedChatView.vue

git commit -m "feat: filter encrypted chat to secure-capable targets"
```

---

### Task 9: Convert `ChatView.vue` into the normal message center

**Files:**
- Modify: `web/src/views/ChatView.vue`
- Optional create: `web/src/api/normalChat.ts`
- Test: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

- [ ] **Step 1: Replace encrypted direct/group message interfaces with normal ones**

Define normal message shapes near the top of `ChatView.vue`:

```ts
interface NormalDirectMessage {
  id?: number
  senderId: number
  receiverId: number
  content: string
  createdAt?: string
}

interface NormalGroupMessage {
  id?: number
  groupId: number
  senderId: number
  content: string
  createdAt?: string
}
```

Update `RenderMessage` so it no longer carries encrypted ciphertext fields in the normal page.

```ts
interface RenderMessage {
  id?: number
  senderId: number
  receiverId?: number
  groupId?: number
  sourceType?: 'friend' | 'group'
  content: string
  createdAt?: string
}
```

- [ ] **Step 2: Remove encrypted-only imports and initialization**

Delete these imports from `ChatView.vue`:

```ts
buildEncryptedMessageDisplay,
decryptMessage,
E2EE_MESSAGE_ALGORITHM,
encryptMessage,
exportPrivateKey,
exportPublicKey,
generateKeyPair,
importPrivateKey,
importPublicKey,
loadPrivateKey,
savePrivateKey,
selectMessagePayloadForUser,
buildGroupMessageEnvelope,
decryptGroupMessage,
GROUP_CONTENT_ALGORITHM
```

Delete encrypted-only state and lifecycle calls:

```ts
const cryptoReady = ref(true)
const privateKey = ref<CryptoKey | null>(null)
await ensureOwnKeyPair()
```

The normal page must not initialize keys or block on crypto availability.

- [ ] **Step 3: Point message load/send methods at the new normal endpoints**

Replace normal direct flow with:

```ts
async function loadFriendMessages() {
  if (!currentFriendId.value) {
    messages.value = []
    return
  }
  const { data } = await http.get(`/normal-messages?friendId=${currentFriendId.value}`)
  messages.value = (data as NormalDirectMessage[]).map((message) => ({
    ...message,
    sourceType: 'friend'
  }))
  await scrollToBottom()
}
```

Replace normal group flow with:

```ts
async function loadGroupMessages() {
  if (!currentGroupId.value) {
    messages.value = []
    return
  }
  const { data } = await http.get(`/groups/${currentGroupId.value}/normal-messages`)
  messages.value = (data as NormalGroupMessage[]).map((message) => ({
    ...message,
    sourceType: 'group'
  }))
  await scrollToBottom()
}
```

Replace friend send:

```ts
async function sendFriendMessage() {
  if (!currentFriendId.value) return
  sending.value = true
  try {
    const content = draft.value.trim()
    const { data } = await http.post('/normal-messages', {
      receiverId: currentFriendId.value,
      content
    })
    messages.value.push({ ...(data as NormalDirectMessage), sourceType: 'friend' })
    draft.value = ''
    await scrollToBottom()
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    sending.value = false
  }
}
```

Replace group send:

```ts
async function sendGroupMessage() {
  if (!currentGroupId.value) return
  sending.value = true
  try {
    const content = draft.value.trim()
    const { data } = await http.post(`/groups/${currentGroupId.value}/normal-messages`, {
      content
    })
    messages.value.push({ ...(data as NormalGroupMessage), sourceType: 'group' })
    draft.value = ''
    await scrollToBottom()
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    sending.value = false
  }
}
```

- [ ] **Step 4: Add the encrypted channel entry card to the normal page**

Add this method:

```ts
function openEncryptedChat() {
  router.push('/encrypted-chat')
}
```

Add an entry card near AI/Favorites:

```vue
<button class="encrypted-entry-card" type="button" @click="openEncryptedChat">
  <div>
    <p class="apple-label">Encrypted Channel</p>
    <strong>加密通道</strong>
    <span>进入独立安全会话页面</span>
  </div>
  <span class="encrypted-entry-arrow">⌁</span>
</button>
```

- [ ] **Step 5: Remove encrypted state copy from the normal page UI**

Delete this block from the normal page composer:

```vue
<div class="composer-meta">
  <span>加密通道</span>
  <strong>{{ cryptoReady ? '已开启' : '不可用' }}</strong>
</div>
```

Replace the composer with a normal message input and button only.

Use placeholder text:

```vue
placeholder="输入消息，按回车发送"
```

- [ ] **Step 6: Update websocket handling for normal event types**

In `connectSocket`, consume the new normal event names only on the normal page:

```ts
if (payload.type === 'normal_chat_message') {
  const chatMessage = payload.data as NormalDirectMessage
  if (
    conversationType.value === 'friend' &&
    currentFriendId.value &&
    (chatMessage.senderId === currentFriendId.value || chatMessage.receiverId === currentFriendId.value)
  ) {
    messages.value.push({ ...chatMessage, sourceType: 'friend' })
    await scrollToBottom()
  }
} else if (payload.type === 'normal_group_message') {
  const groupMessage = payload.data as NormalGroupMessage
  if (groupMessage.senderId === authStore.user?.id) return
  if (conversationType.value === 'group' && currentGroupId.value === groupMessage.groupId) {
    messages.value.push({ ...groupMessage, sourceType: 'group' })
    await scrollToBottom()
  }
}
```

Do not consume encrypted `chat_message` or `group_message` in the normal page anymore.

- [ ] **Step 7: Keep AI and favorites only in the normal page**

Retain these functions and their entry cards in `ChatView.vue`:

```ts
openAIChat()
openFavorites()
```

Keep multi-select favorites if desired for the normal page. If the current favorite backend assumes friend/group/ai source types only, keep `sourceType: 'friend' | 'group'` unchanged in the normal page render messages.

- [ ] **Step 8: Run the build to verify the normal page passes**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

Expected: PASS.

- [ ] **Step 9: Commit the normal page conversion**

```bash
git add web/src/views/ChatView.vue web/src/router/index.ts

git commit -m "feat: convert message center to normal chat"
```

---

### Task 10: Final integration verification

**Files:**
- Verify: `web/src/views/ChatView.vue`
- Verify: `web/src/views/EncryptedChatView.vue`
- Verify: `server/internal/router/router.go`
- Verify: `server/internal/router/message_test.go`
- Verify: `server/internal/router/group_test.go`
- Verify: `server/internal/router/websocket_test.go`

- [ ] **Step 1: Run backend router tests for normal + encrypted coexistence**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./internal/router -v`

Expected: PASS with both old encrypted tests and new normal chat tests green.

- [ ] **Step 2: Run frontend production build**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`

Expected: PASS.

- [ ] **Step 3: Manually verify the required acceptance flows**

Use this checklist:

```md
- [ ] `/chat` shows AI 助手、我的收藏、加密通道入口
- [ ] `/chat` 点击好友进入普通聊天
- [ ] `/chat` 点击群聊进入普通群聊
- [ ] `/chat` 页面没有“加密通道已开启”主提示
- [ ] `/encrypted-chat` 不显示 AI 助手或我的收藏
- [ ] `/encrypted-chat` 仅显示具备加密能力的好友/群聊
- [ ] `/encrypted-chat` 能查看现有加密历史
- [ ] `/encrypted-chat` 保持深色、专业、沉稳风格
- [ ] 移动端列表/会话切换仍可用
```

- [ ] **Step 4: Commit the final integration checkpoint**

```bash
git add web/src/views/ChatView.vue web/src/views/EncryptedChatView.vue web/src/router/index.ts server/internal/router/router.go server/internal/router/message_test.go server/internal/router/group_test.go server/internal/router/websocket_test.go server/internal/model/normal_message.go server/internal/model/normal_group_message.go server/internal/repository/normal_message_gorm.go server/internal/repository/normal_group_message_gorm.go server/internal/service/normal_message.go server/internal/service/normal_group_message.go server/internal/handler/normal_message.go server/internal/handler/normal_group_message.go

git commit -m "feat: split normal and encrypted chat flows"
```

---

## Self-Review

### Spec coverage check
- `/chat` normal page responsibility: covered in Task 9 and Task 10.
- `/encrypted-chat` dedicated page: covered in Task 6, Task 7, Task 8.
- Existing encrypted history retained: covered in Task 7.
- Normal direct/group backend split: covered in Task 1 through Task 4.
- Websocket event separation: covered in Task 5 and Task 9.
- Encrypted-capable filtering: covered in Task 8.
- Dark professional encrypted style: covered in Task 7.
- AI/favorites only on normal page: covered in Task 7 and Task 9.

### Placeholder scan
- No `TODO` / `TBD` placeholders remain.
- All code-writing steps include concrete code blocks.
- All verification steps include exact commands.

### Type consistency check
- Normal direct events and routes use `NormalMessage`, `/normal-messages`, `normal_chat_message` consistently.
- Normal group events and routes use `NormalGroupMessage`, `/groups/:id/normal-messages`, `normal_group_message` consistently.
- Encrypted routes remain `/messages` and `/groups/:id/messages` consistently.
