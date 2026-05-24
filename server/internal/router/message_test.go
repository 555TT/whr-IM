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

	createMessage(t, r, aliceToken, `{"receiverId":2,"content":"hello bob"}`)
	createMessage(t, r, bobToken, `{"receiverId":1,"content":"hi alice"}`)

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
		Content    string `json:"content"`
		MsgType    string `json:"msgType"`
		CreatedAt  string `json:"createdAt"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("expected valid history json, got error: %v", err)
	}
	if len(historyResp) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(historyResp))
	}
	if historyResp[0].Content != "hello bob" || historyResp[1].Content != "hi alice" {
		t.Fatalf("expected messages in order, got %#v", historyResp)
	}
	if historyResp[0].MsgType != "text" {
		t.Fatalf("expected default msgType text, got %q", historyResp[0].MsgType)
	}
	if historyResp[0].CreatedAt == "" || historyResp[1].CreatedAt == "" {
		t.Fatalf("expected non-empty createdAt values, got %#v", historyResp)
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
