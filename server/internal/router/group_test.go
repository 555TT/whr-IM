package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"whr-im/server/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 验证:创建群时,如果传入的成员不是创建者好友,接口应当 400 拒绝。
func TestCreateGroupRejectsNonFriendMembers(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	_ = registerAndLogin(t, r, "bobby") // bob 已注册但不是 alice 的好友
	// alice 没有任何好友,memberIds 包含 bob 的 id 应被拒绝
	body := []byte(`{"name":"无好友群","memberIds":[2]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("not your friend")) {
		t.Fatalf("expected non-friend error, got body %s", w.Body.String())
	}
}

// 验证:发群消息时,memberKeys 必须完整覆盖当前所有群成员;少一份/多一份/缺哪个用户都拒绝。
func TestSendGroupMessageRejectsIncompleteMemberKeys(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	// alice 创建群,把 bob 加入
	createBody := []byte(`{"name":"群1","memberIds":[2]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}
	var detail struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &detail); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}

	// 群成员是 alice(1) 和 bob(2);只给 alice 一份 key,缺 bob
	missingBody := []byte(`{
        "contentCiphertext":"ct",
        "contentIv":"iv",
        "contentAlgorithm":"aes-gcm-256",
        "memberKeys":[{"userId":1,"keyCiphertext":"k1","keyAlgorithm":"rsa-oaep-sha256"}]
    }`)
	missingReq := httptest.NewRequest(http.MethodPost, "/api/groups/1/messages", bytes.NewReader(missingBody))
	missingReq.Header.Set("Content-Type", "application/json")
	missingReq.Header.Set("Authorization", "Bearer "+aliceToken)
	missingW := httptest.NewRecorder()
	r.ServeHTTP(missingW, missingReq)

	if missingW.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for incomplete memberKeys, got %d with body %s", missingW.Code, missingW.Body.String())
	}
	if !bytes.Contains(missingW.Body.Bytes(), []byte("memberKeys must cover")) {
		t.Fatalf("expected memberKeys cover-all error, got body %s", missingW.Body.String())
	}

	// 多一份不属于群的 user 也拒绝
	extraBody := []byte(`{
        "contentCiphertext":"ct",
        "contentIv":"iv",
        "contentAlgorithm":"aes-gcm-256",
        "memberKeys":[
          {"userId":1,"keyCiphertext":"k1","keyAlgorithm":"rsa-oaep-sha256"},
          {"userId":2,"keyCiphertext":"k2","keyAlgorithm":"rsa-oaep-sha256"},
          {"userId":3,"keyCiphertext":"k3","keyAlgorithm":"rsa-oaep-sha256"}
        ]
    }`)
	extraReq := httptest.NewRequest(http.MethodPost, "/api/groups/1/messages", bytes.NewReader(extraBody))
	extraReq.Header.Set("Content-Type", "application/json")
	extraReq.Header.Set("Authorization", "Bearer "+aliceToken)
	extraW := httptest.NewRecorder()
	r.ServeHTTP(extraW, extraReq)

	if extraW.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for extra memberKeys, got %d with body %s", extraW.Code, extraW.Body.String())
	}

	// 完整 memberKeys 应当成功
	okBody := []byte(`{
        "contentCiphertext":"ct",
        "contentIv":"iv",
        "contentAlgorithm":"aes-gcm-256",
        "memberKeys":[
          {"userId":1,"keyCiphertext":"k1","keyAlgorithm":"rsa-oaep-sha256"},
          {"userId":2,"keyCiphertext":"k2","keyAlgorithm":"rsa-oaep-sha256"}
        ]
    }`)
	okReq := httptest.NewRequest(http.MethodPost, "/api/groups/1/messages", bytes.NewReader(okBody))
	okReq.Header.Set("Content-Type", "application/json")
	okReq.Header.Set("Authorization", "Bearer "+aliceToken)
	okW := httptest.NewRecorder()
	r.ServeHTTP(okW, okReq)

	if okW.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for complete memberKeys, got %d with body %s", okW.Code, okW.Body.String())
	}

	// bob 拉历史消息,只能看到自己那份 key
	listReq := httptest.NewRequest(http.MethodGet, "/api/groups/1/messages", nil)
	listReq.Header.Set("Authorization", "Bearer "+bobToken)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d with body %s", listW.Code, listW.Body.String())
	}
	var listed []struct {
		KeyCiphertext string `json:"keyCiphertext"`
		SenderID      uint64 `json:"senderId"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listed); err != nil {
		t.Fatalf("expected list json, got error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 message visible to bob, got %d", len(listed))
	}
	if listed[0].KeyCiphertext != "k2" {
		t.Fatalf("expected bob's keyCiphertext k2, got %q", listed[0].KeyCiphertext)
	}
	if listed[0].SenderID != 1 {
		t.Fatalf("expected senderId 1, got %d", listed[0].SenderID)
	}
}

// 验证:非群成员不能发消息也不能拉历史。
func TestNonMemberCannotAccessGroupMessages(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)
	carolToken := registerAndLogin(t, r, "carol")

	// alice 创建群只含自己和 bob
	createBody := []byte(`{"name":"群1","memberIds":[2]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	// carol 拉历史 → 拒绝
	listReq := httptest.NewRequest(http.MethodGet, "/api/groups/1/messages", nil)
	listReq.Header.Set("Authorization", "Bearer "+carolToken)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusBadRequest {
		t.Fatalf("expected non-member list status 400, got %d with body %s", listW.Code, listW.Body.String())
	}
}

func TestNormalGroupMessageSendAndHistoryFlow(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, aliceID := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createBody := []byte(fmt.Sprintf(`{"name":"普通群","memberIds":[%d]}`, bobID))
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	var createdGroup struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}
	if createdGroup.ID == 0 {
		t.Fatalf("expected non-zero group id, got %d", createdGroup.ID)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", bytes.NewReader([]byte(`{"content":"hello group"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusCreated {
		t.Fatalf("expected normal group message create status 201, got %d with body %s", sendW.Code, sendW.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", nil)
	historyReq.Header.Set("Authorization", "Bearer "+bobToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)
	if historyW.Code != http.StatusOK {
		t.Fatalf("expected normal group history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		SenderID  uint64 `json:"senderId"`
		GroupID   uint64 `json:"groupId"`
		Content   string `json:"content"`
		CreatedAt string `json:"createdAt"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid normal group history json, got error: %v", err)
	}
	if len(historyResp) != 1 {
		t.Fatalf("expected 1 normal group message, got %d", len(historyResp))
	}
	if historyResp[0].SenderID != aliceID || historyResp[0].GroupID != createdGroup.ID {
		t.Fatalf("expected normal group message from sender %d in group %d, got %#v", aliceID, createdGroup.ID, historyResp[0])
	}
	if historyResp[0].Content != "hello group" {
		t.Fatalf("expected plaintext content in normal group history response, got %#v", historyResp[0])
	}
	if historyResp[0].CreatedAt == "" {
		t.Fatalf("expected non-empty createdAt, got %#v", historyResp[0])
	}

	encryptedHistoryReq := httptest.NewRequest(http.MethodGet, "/api/groups/"+uint64Param(createdGroup.ID)+"/messages", nil)
	encryptedHistoryReq.Header.Set("Authorization", "Bearer "+bobToken)
	encryptedHistoryW := httptest.NewRecorder()
	r.ServeHTTP(encryptedHistoryW, encryptedHistoryReq)
	if encryptedHistoryW.Code != http.StatusOK {
		t.Fatalf("expected encrypted group history status 200, got %d with body %s", encryptedHistoryW.Code, encryptedHistoryW.Body.String())
	}

	var encryptedHistoryResp []struct {
		SenderID          uint64 `json:"senderId"`
		GroupID           uint64 `json:"groupId"`
		ContentCiphertext string `json:"contentCiphertext"`
		CreatedAt         string `json:"createdAt"`
	}
	if err := json.Unmarshal(encryptedHistoryW.Body.Bytes(), &encryptedHistoryResp); err != nil {
		t.Fatalf("expected valid encrypted group history json, got error: %v", err)
	}
	if len(encryptedHistoryResp) != 0 {
		t.Fatalf("expected encrypted group history to stay isolated from normal group messages, got %#v", encryptedHistoryResp)
	}
}

func TestNormalGroupMessageRejectsWhitespaceOnlyContent(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, _ := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createBody := []byte(fmt.Sprintf(`{"name":"普通群","memberIds":[%d]}`, bobID))
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	var createdGroup struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", bytes.NewReader([]byte(`{"content":"   \n\t  "}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusBadRequest {
		t.Fatalf("expected whitespace-only normal group message status 400, got %d with body %s", sendW.Code, sendW.Body.String())
	}
	if !bytes.Contains(sendW.Body.Bytes(), []byte("content is required")) {
		t.Fatalf("expected content required error, got body %s", sendW.Body.String())
	}
}

func TestNormalGroupMessageHistoryReturnsAscendingOrder(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, aliceID := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createBody := []byte(fmt.Sprintf(`{"name":"普通群","memberIds":[%d]}`, bobID))
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	var createdGroup struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}

	firstReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", bytes.NewReader([]byte(`{"content":"first message"}`)))
	firstReq.Header.Set("Content-Type", "application/json")
	firstReq.Header.Set("Authorization", "Bearer "+aliceToken)
	firstW := httptest.NewRecorder()
	r.ServeHTTP(firstW, firstReq)
	if firstW.Code != http.StatusCreated {
		t.Fatalf("expected first normal group message create status 201, got %d with body %s", firstW.Code, firstW.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", bytes.NewReader([]byte(`{"content":"second message"}`)))
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.Header.Set("Authorization", "Bearer "+bobToken)
	secondW := httptest.NewRecorder()
	r.ServeHTTP(secondW, secondReq)
	if secondW.Code != http.StatusCreated {
		t.Fatalf("expected second normal group message create status 201, got %d with body %s", secondW.Code, secondW.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)
	if historyW.Code != http.StatusOK {
		t.Fatalf("expected normal group history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		SenderID uint64 `json:"senderId"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid normal group history json, got error: %v", err)
	}
	if len(historyResp) != 2 {
		t.Fatalf("expected 2 normal group messages, got %d", len(historyResp))
	}
	if historyResp[0].SenderID != aliceID || historyResp[0].Content != "first message" {
		t.Fatalf("expected first history item to be alice's first message, got %#v", historyResp[0])
	}
	if historyResp[1].SenderID != bobID || historyResp[1].Content != "second message" {
		t.Fatalf("expected second history item to be bob's second message, got %#v", historyResp[1])
	}
}

func TestNormalGroupMessageRejectsNonMember(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, _ := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)
	carolToken, _ := registerAndLoginWithID(t, r, "carol")

	createBody := []byte(fmt.Sprintf(`{"name":"普通群","memberIds":[%d]}`, bobID))
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	var createdGroup struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}
	if createdGroup.ID == 0 {
		t.Fatalf("expected non-zero group id, got %d", createdGroup.ID)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", bytes.NewReader([]byte(`{"content":"intrude"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+carolToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusForbidden {
		t.Fatalf("expected non-member normal group message status 403, got %d with body %s", sendW.Code, sendW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", nil)
	listReq.Header.Set("Authorization", "Bearer "+carolToken)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusForbidden {
		t.Fatalf("expected non-member normal group history status 403, got %d with body %s", listW.Code, listW.Body.String())
	}
}

func TestNormalGroupRoutesDoNotDependOnEncryptedGroupMessageRepo(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test database: %v", err)
	}

	userRepo, err := repository.NewGormUserRepository(db)
	if err != nil {
		t.Fatalf("failed to create gorm user repository: %v", err)
	}
	friendRepo, err := repository.NewGormFriendRepository(db)
	if err != nil {
		t.Fatalf("failed to create gorm friend repository: %v", err)
	}
	groupRepo, err := repository.NewGormGroupRepository(db)
	if err != nil {
		t.Fatalf("failed to create gorm group repository: %v", err)
	}
	normalGroupMessageRepo, err := repository.NewGormNormalGroupMessageRepository(db)
	if err != nil {
		t.Fatalf("failed to create gorm normal group message repository: %v", err)
	}

	r := NewWithRepositories(userRepo, friendRepo, nil, nil, groupRepo, nil, normalGroupMessageRepo, nil, nil, nil)

	aliceToken, _ := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createBody := []byte(fmt.Sprintf(`{"name":"普通群","memberIds":[%d]}`, bobID))
	createReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	var createdGroup struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", bytes.NewReader([]byte(`{"content":"hello without encrypted repo"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusCreated {
		t.Fatalf("expected normal group message create status 201 without encrypted repo, got %d with body %s", sendW.Code, sendW.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/groups/"+uint64Param(createdGroup.ID)+"/normal-messages", nil)
	historyReq.Header.Set("Authorization", "Bearer "+bobToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)
	if historyW.Code != http.StatusOK {
		t.Fatalf("expected normal group history status 200 without encrypted repo, got %d with body %s", historyW.Code, historyW.Body.String())
	}
}
