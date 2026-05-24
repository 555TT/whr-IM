package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFriendRequestAcceptAndFriendsListFlow(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bob")

	requestBody := []byte(`{"toUserId":2,"message":"add me"}`)
	requestReq := httptest.NewRequest(http.MethodPost, "/api/friend-requests", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestReq.Header.Set("Authorization", "Bearer "+aliceToken)
	requestW := httptest.NewRecorder()
	r.ServeHTTP(requestW, requestReq)

	if requestW.Code != http.StatusCreated {
		t.Fatalf("expected friend request status 201, got %d with body %s", requestW.Code, requestW.Body.String())
	}

	incomingReq := httptest.NewRequest(http.MethodGet, "/api/friend-requests/incoming", nil)
	incomingReq.Header.Set("Authorization", "Bearer "+bobToken)
	incomingW := httptest.NewRecorder()
	r.ServeHTTP(incomingW, incomingReq)

	if incomingW.Code != http.StatusOK {
		t.Fatalf("expected incoming requests status 200, got %d with body %s", incomingW.Code, incomingW.Body.String())
	}

	var incomingResp []struct {
		ID         uint64  `json:"id"`
		FromUserID uint64  `json:"fromUserId"`
		ToUserID   uint64  `json:"toUserId"`
		Message    string `json:"message"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(incomingW.Body.Bytes(), &incomingResp); err != nil {
		t.Fatalf("expected valid incoming list json, got error: %v", err)
	}
	if len(incomingResp) != 1 {
		t.Fatalf("expected 1 incoming request, got %d", len(incomingResp))
	}
	if incomingResp[0].Status != "pending" {
		t.Fatalf("expected pending status, got %q", incomingResp[0].Status)
	}

	acceptReq := httptest.NewRequest(http.MethodPut, "/api/friend-requests/1/accept", nil)
	acceptReq.Header.Set("Authorization", "Bearer "+bobToken)
	acceptW := httptest.NewRecorder()
	r.ServeHTTP(acceptW, acceptReq)

	if acceptW.Code != http.StatusOK {
		t.Fatalf("expected accept status 200, got %d with body %s", acceptW.Code, acceptW.Body.String())
	}

	friendsReq := httptest.NewRequest(http.MethodGet, "/api/friends", nil)
	friendsReq.Header.Set("Authorization", "Bearer "+aliceToken)
	friendsW := httptest.NewRecorder()
	r.ServeHTTP(friendsW, friendsReq)

	if friendsW.Code != http.StatusOK {
		t.Fatalf("expected friends status 200, got %d with body %s", friendsW.Code, friendsW.Body.String())
	}

	var friendsResp []struct {
		UserID    uint64  `json:"userId"`
		FriendID  uint64  `json:"friendId"`
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(friendsW.Body.Bytes(), &friendsResp); err != nil {
		t.Fatalf("expected valid friends list json, got error: %v", err)
	}
	if len(friendsResp) != 1 {
		t.Fatalf("expected 1 friend, got %d", len(friendsResp))
	}
	if friendsResp[0].FriendID != 2 {
		t.Fatalf("expected friend id 2, got %d", friendsResp[0].FriendID)
	}
}

func TestFriendRequestCanBeRejected(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bob")

	requestBody := []byte(`{"toUserId":2,"message":"add me"}`)
	requestReq := httptest.NewRequest(http.MethodPost, "/api/friend-requests", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestReq.Header.Set("Authorization", "Bearer "+aliceToken)
	requestW := httptest.NewRecorder()
	r.ServeHTTP(requestW, requestReq)
	if requestW.Code != http.StatusCreated {
		t.Fatalf("expected friend request status 201, got %d", requestW.Code)
	}

	rejectReq := httptest.NewRequest(http.MethodPut, "/api/friend-requests/1/reject", nil)
	rejectReq.Header.Set("Authorization", "Bearer "+bobToken)
	rejectW := httptest.NewRecorder()
	r.ServeHTTP(rejectW, rejectReq)

	if rejectW.Code != http.StatusOK {
		t.Fatalf("expected reject status 200, got %d with body %s", rejectW.Code, rejectW.Body.String())
	}

	friendsReq := httptest.NewRequest(http.MethodGet, "/api/friends", nil)
	friendsReq.Header.Set("Authorization", "Bearer "+aliceToken)
	friendsW := httptest.NewRecorder()
	r.ServeHTTP(friendsW, friendsReq)
	if friendsW.Code != http.StatusOK {
		t.Fatalf("expected friends status 200, got %d", friendsW.Code)
	}

	var friendsResp []map[string]any
	if err := json.Unmarshal(friendsW.Body.Bytes(), &friendsResp); err != nil {
		t.Fatalf("expected valid friends response json, got error: %v", err)
	}
	if len(friendsResp) != 0 {
		t.Fatalf("expected no friends after rejection, got %d", len(friendsResp))
	}
}

func registerAndLogin(t *testing.T, r http.Handler, username string) string {
	t.Helper()

	registerBody := []byte(`{"username":"` + username + `","password":"secret123","confirmPassword":"secret123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	r.ServeHTTP(registerW, registerReq)
	if registerW.Code != http.StatusCreated {
		t.Fatalf("expected register status 201 for %s, got %d with body %s", username, registerW.Code, registerW.Body.String())
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
	return loginResp.Token
}
