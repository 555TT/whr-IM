package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocketDeliversChatMessageToOnlineFriend(t *testing.T) {
	r := newTestRouter(t)
	server := httptest.NewServer(r)
	defer server.Close()

	aliceToken := registerAndLoginHTTP(t, server.URL, "alice")
	bobToken := registerAndLoginHTTP(t, server.URL, "bobby")
	makeFriendsHTTP(t, server.URL, aliceToken, bobToken)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=" + bobToken
	bobConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected websocket connect success, got error: %v", err)
	}
	defer bobConn.Close()

	createMessageHTTP(t, server.URL, aliceToken, `{"receiverId":2,"senderCiphertext":"hello ws sender copy","senderAlgorithm":"rsa-oaep-sha256","receiverCiphertext":"hello ws encrypted","receiverAlgorithm":"rsa-oaep-sha256"}`)

	if err := bobConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("expected set read deadline success, got error: %v", err)
	}
	_, message, err := bobConn.ReadMessage()
	if err != nil {
		t.Fatalf("expected websocket message, got error: %v", err)
	}

	var envelope struct {
		Type string `json:"type"`
		Data struct {
			SenderID           uint64 `json:"senderId"`
			ReceiverID         uint64 `json:"receiverId"`
			SenderCiphertext   string `json:"senderCiphertext"`
			SenderAlgorithm    string `json:"senderAlgorithm"`
			ReceiverCiphertext string `json:"receiverCiphertext"`
			ReceiverAlgorithm  string `json:"receiverAlgorithm"`
			CreatedAt          string `json:"createdAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("expected valid websocket json, got error: %v", err)
	}
	if bytes.Contains(message, []byte("\"content\"")) {
		t.Fatalf("expected websocket payload without plaintext content field, got %s", string(message))
	}
	if envelope.Type != "chat_message" {
		t.Fatalf("expected chat_message, got %q", envelope.Type)
	}
	if envelope.Data.SenderCiphertext != "hello ws sender copy" {
		t.Fatalf("expected delivered sender ciphertext hello ws sender copy, got %q", envelope.Data.SenderCiphertext)
	}
	if envelope.Data.ReceiverCiphertext != "hello ws encrypted" {
		t.Fatalf("expected delivered receiver ciphertext hello ws encrypted, got %q", envelope.Data.ReceiverCiphertext)
	}
	if envelope.Data.SenderAlgorithm != "rsa-oaep-sha256" {
		t.Fatalf("expected delivered sender algorithm rsa-oaep-sha256, got %q", envelope.Data.SenderAlgorithm)
	}
	if envelope.Data.ReceiverAlgorithm != "rsa-oaep-sha256" {
		t.Fatalf("expected delivered receiver algorithm rsa-oaep-sha256, got %q", envelope.Data.ReceiverAlgorithm)
	}
	if envelope.Data.CreatedAt == "" {
		t.Fatalf("expected delivered createdAt, got empty payload %#v", envelope.Data)
	}
}

func TestNormalDirectMessageBroadcastsNormalChatEvent(t *testing.T) {
	r := newTestRouter(t)
	server := httptest.NewServer(r)
	defer server.Close()

	aliceToken, aliceID := registerAndLoginHTTPWithID(t, server.URL, "alice")
	bobToken, bobID := registerAndLoginHTTPWithID(t, server.URL, "bobby")
	makeFriendsHTTP(t, server.URL, aliceToken, bobToken)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=" + bobToken
	bobConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected websocket connect success, got error: %v", err)
	}
	defer bobConn.Close()

	createNormalMessageHTTP(t, server.URL, aliceToken, bobID, "hello normal ws")

	message := readWebSocketEventByType(t, bobConn, "normal_chat_message")

	var envelope struct {
		Type string `json:"type"`
		Data struct {
			SenderID   uint64 `json:"senderId"`
			ReceiverID uint64 `json:"receiverId"`
			Content    string `json:"content"`
			CreatedAt  string `json:"createdAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("expected valid websocket json, got error: %v", err)
	}
	if envelope.Type != "normal_chat_message" {
		t.Fatalf("expected normal_chat_message, got %q in payload %s", envelope.Type, string(message))
	}
	if bytes.Contains(message, []byte("Ciphertext")) {
		t.Fatalf("expected websocket payload without ciphertext fields, got %s", string(message))
	}
	if envelope.Data.SenderID != aliceID {
		t.Fatalf("expected senderId %d, got %d", aliceID, envelope.Data.SenderID)
	}
	if envelope.Data.ReceiverID != bobID {
		t.Fatalf("expected receiverId %d, got %d", bobID, envelope.Data.ReceiverID)
	}
	if envelope.Data.Content != "hello normal ws" {
		t.Fatalf("expected plaintext content hello normal ws, got %q", envelope.Data.Content)
	}
	if envelope.Data.CreatedAt == "" {
		t.Fatalf("expected delivered createdAt, got empty payload %#v", envelope.Data)
	}
}

func TestNormalGroupMessageBroadcastsNormalGroupEvent(t *testing.T) {
	r := newTestRouter(t)
	server := httptest.NewServer(r)
	defer server.Close()

	aliceToken, aliceID := registerAndLoginHTTPWithID(t, server.URL, "alice")
	bobToken, bobID := registerAndLoginHTTPWithID(t, server.URL, "bobby")
	makeFriendsHTTP(t, server.URL, aliceToken, bobToken)

	groupID := createGroupHTTP(t, server.URL, aliceToken, "普通群", []uint64{bobID})

	aliceWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=" + aliceToken
	aliceConn, _, err := websocket.DefaultDialer.Dial(aliceWSURL, nil)
	if err != nil {
		t.Fatalf("expected websocket connect success for alice, got error: %v", err)
	}
	defer aliceConn.Close()

	bobWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=" + bobToken
	bobConn, _, err := websocket.DefaultDialer.Dial(bobWSURL, nil)
	if err != nil {
		t.Fatalf("expected websocket connect success for bob, got error: %v", err)
	}
	defer bobConn.Close()

	createNormalGroupMessageHTTP(t, server.URL, aliceToken, groupID, "hello normal group")

	assertNormalGroupMessageEnvelope(t, aliceConn, groupID, aliceID, "hello normal group")
	assertNormalGroupMessageEnvelope(t, bobConn, groupID, aliceID, "hello normal group")
}

func registerAndLoginHTTP(t *testing.T, baseURL string, username string) string {
	t.Helper()

	token, _ := registerAndLoginHTTPWithID(t, baseURL, username)
	return token
}

func registerAndLoginHTTPWithID(t *testing.T, baseURL string, username string) (string, uint64) {
	t.Helper()

	registerBody := []byte(`{"username":"` + username + `","password":"secret123","confirmPassword":"secret123"}`)
	registerReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp, err := http.DefaultClient.Do(registerReq)
	if err != nil {
		t.Fatalf("expected register http request success, got error: %v", err)
	}
	defer registerResp.Body.Close()
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected register status 201 for %s, got %d", username, registerResp.StatusCode)
	}

	var registerData struct {
		User struct {
			ID uint64 `json:"id"`
		} `json:"user"`
	}
	if err := json.NewDecoder(registerResp.Body).Decode(&registerData); err != nil {
		t.Fatalf("expected valid register response json, got error: %v", err)
	}
	if registerData.User.ID == 0 {
		t.Fatalf("expected non-zero user id for %s", username)
	}

	loginBody := []byte(`{"username":"` + username + `","password":"secret123"}`)
	loginReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := http.DefaultClient.Do(loginReq)
	if err != nil {
		t.Fatalf("expected login http request success, got error: %v", err)
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected login status 200 for %s, got %d", username, loginResp.StatusCode)
	}

	var loginData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&loginData); err != nil {
		t.Fatalf("expected valid login response json, got error: %v", err)
	}
	if loginData.Token == "" {
		t.Fatalf("expected non-empty token for %s", username)
	}
	return loginData.Token, registerData.User.ID
}

func makeFriendsHTTP(t *testing.T, baseURL string, aliceToken string, bobToken string) {
	t.Helper()

	requestBody := []byte(`{"toUsername":"bobby","message":"add me"}`)
	requestReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/friend-requests", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestReq.Header.Set("Authorization", "Bearer "+aliceToken)
	requestResp, err := http.DefaultClient.Do(requestReq)
	if err != nil {
		t.Fatalf("expected friend request http success, got error: %v", err)
	}
	defer requestResp.Body.Close()
	if requestResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected friend request status 201, got %d", requestResp.StatusCode)
	}

	acceptReq, _ := http.NewRequest(http.MethodPut, baseURL+"/api/friend-requests/1/accept", nil)
	acceptReq.Header.Set("Authorization", "Bearer "+bobToken)
	acceptResp, err := http.DefaultClient.Do(acceptReq)
	if err != nil {
		t.Fatalf("expected accept request http success, got error: %v", err)
	}
	defer acceptResp.Body.Close()
	if acceptResp.StatusCode != http.StatusOK {
		t.Fatalf("expected accept request status 200, got %d", acceptResp.StatusCode)
	}
}

func createMessageHTTP(t *testing.T, baseURL string, token string, body string) {
	t.Helper()

	messageReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/messages", bytes.NewReader([]byte(body)))
	messageReq.Header.Set("Content-Type", "application/json")
	messageReq.Header.Set("Authorization", "Bearer "+token)
	messageResp, err := http.DefaultClient.Do(messageReq)
	if err != nil {
		t.Fatalf("expected create message http success, got error: %v", err)
	}
	defer messageResp.Body.Close()
	if messageResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create message status 201, got %d", messageResp.StatusCode)
	}
}

func createNormalMessageHTTP(t *testing.T, baseURL string, token string, receiverID uint64, content string) {
	t.Helper()

	body := []byte(fmt.Sprintf(`{"receiverId":%d,"content":%q}`, receiverID, content))
	messageReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/normal-messages", bytes.NewReader(body))
	messageReq.Header.Set("Content-Type", "application/json")
	messageReq.Header.Set("Authorization", "Bearer "+token)
	messageResp, err := http.DefaultClient.Do(messageReq)
	if err != nil {
		t.Fatalf("expected create normal message http success, got error: %v", err)
	}
	defer messageResp.Body.Close()
	if messageResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create normal message status 201, got %d", messageResp.StatusCode)
	}
}

func createGroupHTTP(t *testing.T, baseURL string, token string, name string, memberIDs []uint64) uint64 {
	t.Helper()

	memberParts := make([]string, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		memberParts = append(memberParts, fmt.Sprintf("%d", memberID))
	}
	body := []byte(fmt.Sprintf(`{"name":%q,"memberIds":[%s]}`, name, strings.Join(memberParts, ",")))
	groupReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/groups", bytes.NewReader(body))
	groupReq.Header.Set("Content-Type", "application/json")
	groupReq.Header.Set("Authorization", "Bearer "+token)
	groupResp, err := http.DefaultClient.Do(groupReq)
	if err != nil {
		t.Fatalf("expected create group http success, got error: %v", err)
	}
	defer groupResp.Body.Close()
	if groupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create group status 201, got %d", groupResp.StatusCode)
	}

	var response struct {
		ID uint64 `json:"id"`
	}
	if err := json.NewDecoder(groupResp.Body).Decode(&response); err != nil {
		t.Fatalf("expected create group json, got error: %v", err)
	}
	if response.ID == 0 {
		t.Fatalf("expected non-zero group id")
	}
	return response.ID
}

func createNormalGroupMessageHTTP(t *testing.T, baseURL string, token string, groupID uint64, content string) {
	t.Helper()

	body := []byte(fmt.Sprintf(`{"content":%q}`, content))
	messageReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/groups/"+fmt.Sprintf("%d", groupID)+"/normal-messages", bytes.NewReader(body))
	messageReq.Header.Set("Content-Type", "application/json")
	messageReq.Header.Set("Authorization", "Bearer "+token)
	messageResp, err := http.DefaultClient.Do(messageReq)
	if err != nil {
		t.Fatalf("expected create normal group message http success, got error: %v", err)
	}
	defer messageResp.Body.Close()
	if messageResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create normal group message status 201, got %d", messageResp.StatusCode)
	}
}

func readWebSocketEventByType(t *testing.T, conn *websocket.Conn, wantType string) []byte {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := conn.SetReadDeadline(deadline); err != nil {
			t.Fatalf("expected set read deadline success, got error: %v", err)
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("expected websocket %s event before timeout, got error: %v", wantType, err)
		}

		var envelope struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &envelope); err != nil {
			t.Fatalf("expected valid websocket json while waiting for %s, got error: %v", wantType, err)
		}
		if envelope.Type == wantType {
			return message
		}
	}
}

func assertNormalGroupMessageEnvelope(t *testing.T, conn *websocket.Conn, groupID uint64, senderID uint64, content string) {
	t.Helper()

	message := readWebSocketEventByType(t, conn, "normal_group_message")

	var envelope struct {
		Type string `json:"type"`
		Data struct {
			GroupID   uint64 `json:"groupId"`
			SenderID  uint64 `json:"senderId"`
			Content   string `json:"content"`
			CreatedAt string `json:"createdAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("expected valid websocket json, got error: %v", err)
	}
	if envelope.Type != "normal_group_message" {
		t.Fatalf("expected normal_group_message, got %q in payload %s", envelope.Type, string(message))
	}
	if bytes.Contains(message, []byte("Ciphertext")) {
		t.Fatalf("expected websocket payload without ciphertext fields, got %s", string(message))
	}
	if envelope.Data.GroupID != groupID {
		t.Fatalf("expected groupId %d, got %d", groupID, envelope.Data.GroupID)
	}
	if envelope.Data.SenderID != senderID {
		t.Fatalf("expected senderId %d, got %d", senderID, envelope.Data.SenderID)
	}
	if envelope.Data.Content != content {
		t.Fatalf("expected plaintext content %q, got %q", content, envelope.Data.Content)
	}
	if envelope.Data.CreatedAt == "" {
		t.Fatalf("expected delivered createdAt, got empty payload %#v", envelope.Data)
	}
}
