package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessageHistoryReturnsConversationInTimeOrder(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createMessage(t, r, aliceToken, `{"receiverId":2,"senderCiphertext":"hello alice copy","senderAlgorithm":"rsa-oaep-sha256","receiverCiphertext":"hello bob encrypted","receiverAlgorithm":"rsa-oaep-sha256"}`)
	createMessage(t, r, bobToken, `{"receiverId":1,"senderCiphertext":"hi bob copy","senderAlgorithm":"rsa-oaep-sha256","receiverCiphertext":"hi alice encrypted","receiverAlgorithm":"rsa-oaep-sha256"}`)

	historyReq := httptest.NewRequest(http.MethodGet, "/api/messages?friendId=2", nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)

	if historyW.Code != http.StatusOK {
		t.Fatalf("expected history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		SenderID           uint64 `json:"senderId"`
		ReceiverID         uint64 `json:"receiverId"`
		SenderCiphertext   string `json:"senderCiphertext"`
		SenderAlgorithm    string `json:"senderAlgorithm"`
		ReceiverCiphertext string `json:"receiverCiphertext"`
		ReceiverAlgorithm  string `json:"receiverAlgorithm"`
		CreatedAt          string `json:"createdAt"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid history json, got error: %v", err)
	}
	if bytes.Contains(historyW.Body.Bytes(), []byte("\"content\"")) {
		t.Fatalf("expected history response without plaintext content field, got body %s", historyW.Body.String())
	}
	if len(historyResp) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(historyResp))
	}
	if historyResp[0].SenderCiphertext != "hello alice copy" || historyResp[1].SenderCiphertext != "hi bob copy" {
		t.Fatalf("expected sender ciphertexts in order, got %#v", historyResp)
	}
	if historyResp[0].ReceiverCiphertext != "hello bob encrypted" || historyResp[1].ReceiverCiphertext != "hi alice encrypted" {
		t.Fatalf("expected receiver ciphertexts in order, got %#v", historyResp)
	}
	if historyResp[0].SenderAlgorithm != "rsa-oaep-sha256" || historyResp[1].SenderAlgorithm != "rsa-oaep-sha256" {
		t.Fatalf("expected sender algorithm rsa-oaep-sha256, got %#v", historyResp)
	}
	if historyResp[0].ReceiverAlgorithm != "rsa-oaep-sha256" || historyResp[1].ReceiverAlgorithm != "rsa-oaep-sha256" {
		t.Fatalf("expected receiver algorithm rsa-oaep-sha256, got %#v", historyResp)
	}
	if historyResp[0].CreatedAt == "" || historyResp[1].CreatedAt == "" {
		t.Fatalf("expected non-empty createdAt values, got %#v", historyResp)
	}
}

func TestMessageRejectsUnsupportedAlgorithm(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	messageReq := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader([]byte(`{"receiverId":2,"senderCiphertext":"hello alice copy","senderAlgorithm":"rsa-oaep-sha256","receiverCiphertext":"hello bob encrypted","receiverAlgorithm":"sealed-box"}`)))
	messageReq.Header.Set("Content-Type", "application/json")
	messageReq.Header.Set("Authorization", "Bearer "+aliceToken)
	messageW := httptest.NewRecorder()
	r.ServeHTTP(messageW, messageReq)

	if messageW.Code != http.StatusBadRequest {
		t.Fatalf("expected create message status 400, got %d with body %s", messageW.Code, messageW.Body.String())
	}
	if !bytes.Contains(messageW.Body.Bytes(), []byte("unsupported algorithm")) {
		t.Fatalf("expected unsupported algorithm error, got body %s", messageW.Body.String())
	}
}

func TestNormalDirectMessageSendAndHistoryFlow(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, aliceID := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	sendNormalMessage(t, r, aliceToken, bobID, "hello bob")

	historyReq := httptest.NewRequest(http.MethodGet, "/api/normal-messages?friendId="+uint64Param(bobID), nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)

	if historyW.Code != http.StatusOK {
		t.Fatalf("expected normal message history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		SenderID   uint64 `json:"senderId"`
		ReceiverID uint64 `json:"receiverId"`
		Content    string `json:"content"`
		CreatedAt  string `json:"createdAt"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid normal message history json, got error: %v", err)
	}
	if len(historyResp) != 1 {
		t.Fatalf("expected 1 normal message, got %d", len(historyResp))
	}
	if historyResp[0].SenderID != aliceID || historyResp[0].ReceiverID != bobID {
		t.Fatalf("expected normal message between sender %d and receiver %d, got %#v", aliceID, bobID, historyResp[0])
	}
	if historyResp[0].Content != "hello bob" {
		t.Fatalf("expected plaintext content in history response, got %#v", historyResp[0])
	}
	if historyResp[0].CreatedAt == "" {
		t.Fatalf("expected non-empty createdAt, got %#v", historyResp[0])
	}

	encryptedHistoryReq := httptest.NewRequest(http.MethodGet, "/api/messages?friendId="+uint64Param(bobID), nil)
	encryptedHistoryReq.Header.Set("Authorization", "Bearer "+aliceToken)
	encryptedHistoryW := httptest.NewRecorder()
	r.ServeHTTP(encryptedHistoryW, encryptedHistoryReq)

	if encryptedHistoryW.Code != http.StatusOK {
		t.Fatalf("expected encrypted message history status 200, got %d with body %s", encryptedHistoryW.Code, encryptedHistoryW.Body.String())
	}

	var encryptedHistoryResp []struct {
		SenderID           uint64 `json:"senderId"`
		ReceiverID         uint64 `json:"receiverId"`
		SenderCiphertext   string `json:"senderCiphertext"`
		SenderAlgorithm    string `json:"senderAlgorithm"`
		ReceiverCiphertext string `json:"receiverCiphertext"`
		ReceiverAlgorithm  string `json:"receiverAlgorithm"`
		CreatedAt          string `json:"createdAt"`
	}
	if err := json.Unmarshal(encryptedHistoryW.Body.Bytes(), &encryptedHistoryResp); err != nil {
		t.Fatalf("expected valid encrypted history json, got error: %v", err)
	}
	if len(encryptedHistoryResp) != 0 {
		t.Fatalf("expected encrypted message history to stay isolated from normal messages, got %#v", encryptedHistoryResp)
	}
}

func TestNormalDirectMessageRejectsWhitespaceOnlyContent(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, _ := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	sendReq := httptest.NewRequest(http.MethodPost, "/api/normal-messages", bytes.NewReader([]byte(`{"receiverId":`+uint64Param(bobID)+`,"content":"   "}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusBadRequest {
		t.Fatalf("expected whitespace-only normal message status 400, got %d with body %s", sendW.Code, sendW.Body.String())
	}
	if !bytes.Contains(sendW.Body.Bytes(), []byte("content is required")) {
		t.Fatalf("expected whitespace-only content validation error, got body %s", sendW.Body.String())
	}
}

func TestNormalDirectMessageHistoryReturnsBidirectionalConversationInTimeOrder(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, aliceID := registerAndLoginWithID(t, r, "alice")
	bobToken, bobID := registerAndLoginWithID(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	sendNormalMessage(t, r, aliceToken, bobID, "hello bob")
	sendNormalMessage(t, r, bobToken, aliceID, "hi alice")

	historyReq := httptest.NewRequest(http.MethodGet, "/api/normal-messages?friendId="+uint64Param(bobID), nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)

	if historyW.Code != http.StatusOK {
		t.Fatalf("expected bidirectional normal message history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		SenderID   uint64 `json:"senderId"`
		ReceiverID uint64 `json:"receiverId"`
		Content    string `json:"content"`
		CreatedAt  string `json:"createdAt"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid bidirectional normal message history json, got error: %v", err)
	}
	if len(historyResp) != 2 {
		t.Fatalf("expected 2 normal messages, got %d", len(historyResp))
	}
	if historyResp[0].SenderID != aliceID || historyResp[0].ReceiverID != bobID || historyResp[0].Content != "hello bob" {
		t.Fatalf("expected first normal message from alice to bob, got %#v", historyResp[0])
	}
	if historyResp[1].SenderID != bobID || historyResp[1].ReceiverID != aliceID || historyResp[1].Content != "hi alice" {
		t.Fatalf("expected second normal message from bob to alice, got %#v", historyResp[1])
	}
	if historyResp[0].CreatedAt == "" || historyResp[1].CreatedAt == "" {
		t.Fatalf("expected non-empty createdAt values, got %#v", historyResp)
	}
}

func TestNormalDirectMessageRejectsNonFriendUsers(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, _ := registerAndLoginWithID(t, r, "alice")
	_, bobID := registerAndLoginWithID(t, r, "bobby")

	sendReq := httptest.NewRequest(http.MethodPost, "/api/normal-messages", bytes.NewReader([]byte(`{"receiverId":`+uint64Param(bobID)+`,"content":"hello bob"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusBadRequest {
		t.Fatalf("expected non-friend normal message status 400, got %d with body %s", sendW.Code, sendW.Body.String())
	}
	if !bytes.Contains(sendW.Body.Bytes(), []byte("non-friend users cannot chat")) {
		t.Fatalf("expected non-friend normal message error, got body %s", sendW.Body.String())
	}
}

func TestNormalDirectMessageHistoryRejectsNonFriendUsers(t *testing.T) {
	r := newTestRouter(t)

	aliceToken, _ := registerAndLoginWithID(t, r, "alice")
	_, bobID := registerAndLoginWithID(t, r, "bobby")

	historyReq := httptest.NewRequest(http.MethodGet, "/api/normal-messages?friendId="+uint64Param(bobID), nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)

	if historyW.Code != http.StatusBadRequest {
		t.Fatalf("expected non-friend normal message history status 400, got %d with body %s", historyW.Code, historyW.Body.String())
	}
	if !bytes.Contains(historyW.Body.Bytes(), []byte("non-friend users cannot chat")) {
		t.Fatalf("expected non-friend normal message history error, got body %s", historyW.Body.String())
	}
}

func createMessage(t *testing.T, r http.Handler, token string, body string) {
	t.Helper()

	messageReq := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader([]byte(body)))
	messageReq.Header.Set("Content-Type", "application/json")
	messageReq.Header.Set("Authorization", "Bearer "+token)
	messageW := httptest.NewRecorder()
	r.ServeHTTP(messageW, messageReq)
	if messageW.Code != http.StatusCreated {
		t.Fatalf("expected create message status 201, got %d with body %s", messageW.Code, messageW.Body.String())
	}
}

func registerAndLoginWithID(t *testing.T, r http.Handler, username string) (string, uint64) {
	t.Helper()

	registerBody := []byte(`{"username":"` + username + `","password":"secret123","confirmPassword":"secret123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	r.ServeHTTP(registerW, registerReq)
	if registerW.Code != http.StatusCreated {
		t.Fatalf("expected register status 201 for %s, got %d with body %s", username, registerW.Code, registerW.Body.String())
	}

	var registerResp struct {
		User struct {
			ID uint64 `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(registerW.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("expected valid register json for %s, got error: %v", username, err)
	}
	if registerResp.User.ID == 0 {
		t.Fatalf("expected non-zero user id for %s", username)
	}

	loginBody := []byte(`{"username":"` + username + `","password":"secret123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("expected login status 200 for %s, got %d with body %s", username, loginW.Code, loginW.Body.String())
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("expected valid login json for %s, got error: %v", username, err)
	}
	if loginResp.Token == "" {
		t.Fatalf("expected token for %s", username)
	}

	return loginResp.Token, registerResp.User.ID
}

func sendNormalMessage(t *testing.T, r http.Handler, token string, receiverID uint64, content string) {
	t.Helper()

	body := []byte(fmt.Sprintf(`{"receiverId":%d,"content":%q}`, receiverID, content))
	sendReq := httptest.NewRequest(http.MethodPost, "/api/normal-messages", bytes.NewReader(body))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+token)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusCreated {
		t.Fatalf("expected normal message create status 201, got %d with body %s", sendW.Code, sendW.Body.String())
	}
}

func uint64Param(value uint64) string {
	return fmt.Sprintf("%d", value)
}

func TestAIChatSendAndHistoryFlow(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")

	sendReq := httptest.NewRequest(http.MethodPost, "/api/ai-chat/messages", bytes.NewReader([]byte(`{"content":"帮我写一条晚安消息"}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusCreated {
		t.Fatalf("expected ai chat create status 201, got %d with body %s", sendW.Code, sendW.Body.String())
	}

	var sendResp []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(sendW.Body.Bytes(), &sendResp); err != nil {
		t.Fatalf("expected valid ai chat create json, got error: %v", err)
	}
	if len(sendResp) != 2 {
		t.Fatalf("expected 2 ai chat messages in create response, got %d", len(sendResp))
	}
	if sendResp[0].Role != "user" || sendResp[0].Content != "帮我写一条晚安消息" {
		t.Fatalf("expected first ai chat message to be user input, got %#v", sendResp[0])
	}
	if sendResp[1].Role != "assistant" || sendResp[1].Content == "" {
		t.Fatalf("expected second ai chat message to be assistant reply, got %#v", sendResp[1])
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/ai-chat/messages", nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)

	if historyW.Code != http.StatusOK {
		t.Fatalf("expected ai chat history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid ai chat history json, got error: %v", err)
	}
	if len(historyResp) != 2 {
		t.Fatalf("expected 2 ai chat history messages, got %d", len(historyResp))
	}
	if historyResp[0].Role != "user" || historyResp[1].Role != "assistant" {
		t.Fatalf("expected ai chat history roles in order, got %#v", historyResp)
	}
}

func TestAIChatRejectsWhitespaceOnlyContent(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")

	sendReq := httptest.NewRequest(http.MethodPost, "/api/ai-chat/messages", bytes.NewReader([]byte(`{"content":"   "}`)))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+aliceToken)
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)

	if sendW.Code != http.StatusBadRequest {
		t.Fatalf("expected ai chat invalid status 400, got %d with body %s", sendW.Code, sendW.Body.String())
	}
}

func TestAIChatRequiresAuthentication(t *testing.T) {
	r := newTestRouter(t)

	historyReq := httptest.NewRequest(http.MethodGet, "/api/ai-chat/messages", nil)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)
	if historyW.Code != http.StatusUnauthorized {
		t.Fatalf("expected ai chat unauth status 401, got %d with body %s", historyW.Code, historyW.Body.String())
	}
}

func TestFavoritesCreateListAndDeleteFlow(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createMessage(t, r, aliceToken, `{"receiverId":2,"senderCiphertext":"hello alice copy","senderAlgorithm":"rsa-oaep-sha256","receiverCiphertext":"hello bob encrypted","receiverAlgorithm":"rsa-oaep-sha256"}`)

	createGroupReq := httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewReader([]byte(`{"name":"收藏群","memberIds":[2]}`)))
	createGroupReq.Header.Set("Content-Type", "application/json")
	createGroupReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createGroupW := httptest.NewRecorder()
	r.ServeHTTP(createGroupW, createGroupReq)
	if createGroupW.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d with body %s", createGroupW.Code, createGroupW.Body.String())
	}

	groupMessageReq := httptest.NewRequest(http.MethodPost, "/api/groups/1/messages", bytes.NewReader([]byte(`{
		"contentCiphertext":"group-ct",
		"contentIv":"group-iv",
		"contentAlgorithm":"aes-gcm-256",
		"memberKeys":[
		  {"userId":1,"keyCiphertext":"k1","keyAlgorithm":"rsa-oaep-sha256"},
		  {"userId":2,"keyCiphertext":"k2","keyAlgorithm":"rsa-oaep-sha256"}
		]
	}`)))
	groupMessageReq.Header.Set("Content-Type", "application/json")
	groupMessageReq.Header.Set("Authorization", "Bearer "+aliceToken)
	groupMessageW := httptest.NewRecorder()
	r.ServeHTTP(groupMessageW, groupMessageReq)
	if groupMessageW.Code != http.StatusCreated {
		t.Fatalf("expected create group message 201, got %d with body %s", groupMessageW.Code, groupMessageW.Body.String())
	}

	aiChatReq := httptest.NewRequest(http.MethodPost, "/api/ai-chat/messages", bytes.NewReader([]byte(`{"content":"收藏一条 AI 回复"}`)))
	aiChatReq.Header.Set("Content-Type", "application/json")
	aiChatReq.Header.Set("Authorization", "Bearer "+aliceToken)
	aiChatW := httptest.NewRecorder()
	r.ServeHTTP(aiChatW, aiChatReq)
	if aiChatW.Code != http.StatusCreated {
		t.Fatalf("expected create ai chat 201, got %d with body %s", aiChatW.Code, aiChatW.Body.String())
	}

	favoriteReq := httptest.NewRequest(http.MethodPost, "/api/favorites", bytes.NewReader([]byte(`{
		"items":[
		  {"sourceType":"friend","sourceMessageId":1},
		  {"sourceType":"group","sourceMessageId":1},
		  {"sourceType":"ai","sourceMessageId":2}
		]
	}`)))
	favoriteReq.Header.Set("Content-Type", "application/json")
	favoriteReq.Header.Set("Authorization", "Bearer "+aliceToken)
	favoriteW := httptest.NewRecorder()
	r.ServeHTTP(favoriteW, favoriteReq)
	if favoriteW.Code != http.StatusCreated {
		t.Fatalf("expected create favorites 201, got %d with body %s", favoriteW.Code, favoriteW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/favorites", nil)
	listReq.Header.Set("Authorization", "Bearer "+aliceToken)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected list favorites 200, got %d with body %s", listW.Code, listW.Body.String())
	}

	var listResp []struct {
		ID         uint64 `json:"id"`
		SourceType string `json:"sourceType"`
		Content    string `json:"content"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("expected valid favorites list json, got error: %v", err)
	}
	if len(listResp) != 3 {
		t.Fatalf("expected 3 favorites, got %d", len(listResp))
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/favorites/1", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+aliceToken)
	deleteW := httptest.NewRecorder()
	r.ServeHTTP(deleteW, deleteReq)
	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("expected delete favorite 204, got %d with body %s", deleteW.Code, deleteW.Body.String())
	}
}

func makeFriends(t *testing.T, r http.Handler, aliceToken string, bobToken string) {
	t.Helper()

	requestBody := []byte(`{"toUsername":"bobby","message":"add me"}`)
	requestReq := httptest.NewRequest(http.MethodPost, "/api/friend-requests", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestReq.Header.Set("Authorization", "Bearer "+aliceToken)
	requestW := httptest.NewRecorder()
	r.ServeHTTP(requestW, requestReq)
	if requestW.Code != http.StatusCreated {
		t.Fatalf("expected friend request status 201, got %d with body %s", requestW.Code, requestW.Body.String())
	}

	acceptReq := httptest.NewRequest(http.MethodPut, "/api/friend-requests/1/accept", nil)
	acceptReq.Header.Set("Authorization", "Bearer "+bobToken)
	acceptW := httptest.NewRecorder()
	r.ServeHTTP(acceptW, acceptReq)
	if acceptW.Code != http.StatusOK {
		t.Fatalf("expected accept status 200, got %d with body %s", acceptW.Code, acceptW.Body.String())
	}
}
