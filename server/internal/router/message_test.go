package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessageHistoryReturnsConversationInTimeOrder(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createMessage(t, r, aliceToken, `{"receiverId":2,"ciphertext":"hello bob encrypted","algorithm":"rsa-oaep-sha256"}`)
	createMessage(t, r, bobToken, `{"receiverId":1,"ciphertext":"hi alice encrypted","algorithm":"rsa-oaep-sha256"}`)

	historyReq := httptest.NewRequest(http.MethodGet, "/api/messages?friendId=2", nil)
	historyReq.Header.Set("Authorization", "Bearer "+aliceToken)
	historyW := httptest.NewRecorder()
	r.ServeHTTP(historyW, historyReq)

	if historyW.Code != http.StatusOK {
		t.Fatalf("expected history status 200, got %d with body %s", historyW.Code, historyW.Body.String())
	}

	var historyResp []struct {
		SenderID   uint64 `json:"senderId"`
		ReceiverID uint64 `json:"receiverId"`
		Ciphertext string `json:"ciphertext"`
		Algorithm  string `json:"algorithm"`
		CreatedAt  string `json:"createdAt"`
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
	if historyResp[0].Ciphertext != "hello bob encrypted" || historyResp[1].Ciphertext != "hi alice encrypted" {
		t.Fatalf("expected ciphertext messages in order, got %#v", historyResp)
	}
	if historyResp[0].Algorithm != "rsa-oaep-sha256" || historyResp[1].Algorithm != "rsa-oaep-sha256" {
		t.Fatalf("expected message algorithm rsa-oaep-sha256, got %#v", historyResp)
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

	messageReq := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader([]byte(`{"receiverId":2,"ciphertext":"hello bob encrypted","algorithm":"sealed-box"}`)))
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
