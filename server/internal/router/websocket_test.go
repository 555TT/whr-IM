package router

import (
	"bytes"
	"encoding/json"
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
	bobToken := registerAndLoginHTTP(t, server.URL, "bob")
	makeFriendsHTTP(t, server.URL, aliceToken, bobToken)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=" + bobToken
	bobConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected websocket connect success, got error: %v", err)
	}
	defer bobConn.Close()

	createMessageHTTP(t, server.URL, aliceToken, `{"receiverId":2,"content":"hello ws"}`)

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
			SenderID   uint64 `json:"senderId"`
			ReceiverID uint64 `json:"receiverId"`
			Content    string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("expected valid websocket json, got error: %v", err)
	}
	if envelope.Type != "chat_message" {
		t.Fatalf("expected chat_message, got %q", envelope.Type)
	}
	if envelope.Data.Content != "hello ws" {
		t.Fatalf("expected delivered content hello ws, got %q", envelope.Data.Content)
	}
}

func registerAndLoginHTTP(t *testing.T, baseURL string, username string) string {
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

	loginBody := []byte(`{"username":"` + username + `","password":"secret123"}`)
	loginReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := http.DefaultClient.Do(loginReq)
	if err != nil {
		t.Fatalf("expected login http request success, got error: %v", err)
	}
	defer loginResp.Body.Close()

	var loginData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&loginData); err != nil {
		t.Fatalf("expected valid login response json, got error: %v", err)
	}
	return loginData.Token
}

func makeFriendsHTTP(t *testing.T, baseURL string, aliceToken string, bobToken string) {
	t.Helper()

	requestBody := []byte(`{"toUserId":2,"message":"add me"}`)
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
