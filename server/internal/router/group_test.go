package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
