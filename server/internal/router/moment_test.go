package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMomentsAIAssistGenerateSuccess(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")

	body := []byte(`{"mode":"generate","prompt":"写一条周末露营朋友圈","tone":"自然","hasImage":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moments/ai-assist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected ai assist generate status 200, got %d with body %s", w.Code, w.Body.String())
	}

	var resp struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid ai assist generate json, got error: %v", err)
	}
	if resp.Text == "" {
		t.Fatal("expected generated text")
	}
}

func TestMomentsAIAssistPolishSuccess(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")

	body := []byte(`{"mode":"polish","content":"今天和朋友吃饭很开心","tone":"文艺","hasImage":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moments/ai-assist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected ai assist polish status 200, got %d with body %s", w.Code, w.Body.String())
	}

	var resp struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid ai assist polish json, got error: %v", err)
	}
	if resp.Text == "" {
		t.Fatal("expected polished text")
	}
}

func TestMomentsAIAssistRejectsInvalidRequest(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")

	body := []byte(`{"mode":"generate","prompt":"","tone":"自然","hasImage":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moments/ai-assist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected ai assist invalid status 400, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestMomentsAIAssistRequiresAuthentication(t *testing.T) {
	r := newTestRouter(t)

	body := []byte(`{"mode":"generate","prompt":"写一条周末露营朋友圈","tone":"自然","hasImage":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moments/ai-assist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected ai assist unauth status 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestCreateMomentRejectsWhitespaceOnlyContent(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")

	createBody := []byte(`{"content":"   ","imageKeys":[]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/moments", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()

	r.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusBadRequest {
		t.Fatalf("expected whitespace-only create moment status 400, got %d with body %s", createW.Code, createW.Body.String())
	}
}

func TestCreateCommentRejectsWhitespaceOnlyContent(t *testing.T) {
	r := newTestRouter(t)
	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	makeFriends(t, r, aliceToken, bobToken)

	createBody := []byte(`{"content":"第一条朋友圈","imageKeys":[]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/moments", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create moment status 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	commentBody := []byte(`{"content":"   "}`)
	commentReq := httptest.NewRequest(http.MethodPost, "/api/moments/1/comments", bytes.NewReader(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+bobToken)
	commentW := httptest.NewRecorder()
	r.ServeHTTP(commentW, commentReq)

	if commentW.Code != http.StatusBadRequest {
		t.Fatalf("expected whitespace-only comment status 400, got %d with body %s", commentW.Code, commentW.Body.String())
	}
}

func TestUserCanCreateAndListFriendVisibleMoments(t *testing.T) {
	r := newTestRouter(t)

	aliceToken := registerAndLogin(t, r, "alice")
	bobToken := registerAndLogin(t, r, "bobby")
	carolToken := registerAndLogin(t, r, "carol")
	makeFriends(t, r, aliceToken, bobToken)

	createBody := []byte(`{"content":"第一条朋友圈","imageKeys":["uploads/users/1/a.png"]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/moments", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceToken)
	createW := httptest.NewRecorder()

	r.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("expected create moment status 201, got %d with body %s", createW.Code, createW.Body.String())
	}

	var created struct {
		ID      uint64   `json:"id"`
		Content string   `json:"content"`
		Images  []string `json:"images"`
		UserID  uint64   `json:"userId"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("expected valid moment create response json, got error: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created moment id")
	}
	if created.Content != "第一条朋友圈" {
		t.Fatalf("expected created moment content, got %q", created.Content)
	}
	if created.UserID != 1 {
		t.Fatalf("expected created userId 1, got %d", created.UserID)
	}
	if len(created.Images) != 1 {
		t.Fatalf("expected one image, got %d", len(created.Images))
	}

	likeReq := httptest.NewRequest(http.MethodPost, "/api/moments/1/likes", nil)
	likeReq.Header.Set("Authorization", "Bearer "+bobToken)
	likeW := httptest.NewRecorder()
	r.ServeHTTP(likeW, likeReq)
	if likeW.Code != http.StatusCreated {
		t.Fatalf("expected like status 201, got %d with body %s", likeW.Code, likeW.Body.String())
	}

	commentBody := []byte(`{"content":"这条动态不错"}`)
	commentReq := httptest.NewRequest(http.MethodPost, "/api/moments/1/comments", bytes.NewReader(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+bobToken)
	commentW := httptest.NewRecorder()
	r.ServeHTTP(commentW, commentReq)
	if commentW.Code != http.StatusCreated {
		t.Fatalf("expected comment status 201, got %d with body %s", commentW.Code, commentW.Body.String())
	}

	bobListReq := httptest.NewRequest(http.MethodGet, "/api/moments", nil)
	bobListReq.Header.Set("Authorization", "Bearer "+bobToken)
	bobListW := httptest.NewRecorder()
	r.ServeHTTP(bobListW, bobListReq)

	if bobListW.Code != http.StatusOK {
		t.Fatalf("expected bob list status 200, got %d with body %s", bobListW.Code, bobListW.Body.String())
	}

	var bobMoments []struct {
		ID        uint64   `json:"id"`
		UserID    uint64   `json:"userId"`
		Nickname  string   `json:"nickname"`
		Content   string   `json:"content"`
		Images    []string `json:"images"`
		LikeCount int      `json:"likeCount"`
		LikedByMe bool     `json:"likedByMe"`
		Comments  []struct {
			UserID   uint64 `json:"userId"`
			Nickname string `json:"nickname"`
			Content  string `json:"content"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(bobListW.Body.Bytes(), &bobMoments); err != nil {
		t.Fatalf("expected valid bob moment list json, got error: %v", err)
	}
	if len(bobMoments) != 1 {
		t.Fatalf("expected bob to see 1 moment, got %d", len(bobMoments))
	}
	if bobMoments[0].UserID != 1 {
		t.Fatalf("expected bob to see alice's moment, got userId %d", bobMoments[0].UserID)
	}
	if bobMoments[0].Nickname != "alice" {
		t.Fatalf("expected nickname alice, got %q", bobMoments[0].Nickname)
	}
	if bobMoments[0].LikeCount != 1 {
		t.Fatalf("expected likeCount 1, got %d", bobMoments[0].LikeCount)
	}
	if !bobMoments[0].LikedByMe {
		t.Fatal("expected bob likedByMe true after like")
	}
	if len(bobMoments[0].Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(bobMoments[0].Comments))
	}
	if bobMoments[0].Comments[0].UserID != 2 {
		t.Fatalf("expected comment userId 2, got %d", bobMoments[0].Comments[0].UserID)
	}
	if bobMoments[0].Comments[0].Nickname != "bobby" {
		t.Fatalf("expected comment nickname bobby, got %q", bobMoments[0].Comments[0].Nickname)
	}
	if bobMoments[0].Comments[0].Content != "这条动态不错" {
		t.Fatalf("expected comment content, got %q", bobMoments[0].Comments[0].Content)
	}

	unlikeReq := httptest.NewRequest(http.MethodDelete, "/api/moments/1/likes/me", nil)
	unlikeReq.Header.Set("Authorization", "Bearer "+bobToken)
	unlikeW := httptest.NewRecorder()
	r.ServeHTTP(unlikeW, unlikeReq)
	if unlikeW.Code != http.StatusNoContent {
		t.Fatalf("expected unlike status 204, got %d with body %s", unlikeW.Code, unlikeW.Body.String())
	}

	bobDeleteReq := httptest.NewRequest(http.MethodDelete, "/api/moments/1", nil)
	bobDeleteReq.Header.Set("Authorization", "Bearer "+bobToken)
	bobDeleteW := httptest.NewRecorder()
	r.ServeHTTP(bobDeleteW, bobDeleteReq)
	if bobDeleteW.Code != http.StatusBadRequest {
		t.Fatalf("expected non-author delete status 400, got %d with body %s", bobDeleteW.Code, bobDeleteW.Body.String())
	}

	aliceDeleteReq := httptest.NewRequest(http.MethodDelete, "/api/moments/1", nil)
	aliceDeleteReq.Header.Set("Authorization", "Bearer "+aliceToken)
	aliceDeleteW := httptest.NewRecorder()
	r.ServeHTTP(aliceDeleteW, aliceDeleteReq)
	if aliceDeleteW.Code != http.StatusNoContent {
		t.Fatalf("expected author delete status 204, got %d with body %s", aliceDeleteW.Code, aliceDeleteW.Body.String())
	}

	bobListAgainReq := httptest.NewRequest(http.MethodGet, "/api/moments", nil)
	bobListAgainReq.Header.Set("Authorization", "Bearer "+bobToken)
	bobListAgainW := httptest.NewRecorder()
	r.ServeHTTP(bobListAgainW, bobListAgainReq)
	if bobListAgainW.Code != http.StatusOK {
		t.Fatalf("expected bob second list status 200, got %d with body %s", bobListAgainW.Code, bobListAgainW.Body.String())
	}
	var bobMomentsAgain []struct {
		LikeCount int  `json:"likeCount"`
		LikedByMe bool `json:"likedByMe"`
	}
	if err := json.Unmarshal(bobListAgainW.Body.Bytes(), &bobMomentsAgain); err != nil {
		t.Fatalf("expected valid bob second list json, got error: %v", err)
	}
	if len(bobMomentsAgain) != 0 {
		t.Fatalf("expected bob second list size 0 after author delete, got %d", len(bobMomentsAgain))
	}

	carolListReq := httptest.NewRequest(http.MethodGet, "/api/moments", nil)
	carolListReq.Header.Set("Authorization", "Bearer "+carolToken)
	carolListW := httptest.NewRecorder()
	r.ServeHTTP(carolListW, carolListReq)

	if carolListW.Code != http.StatusOK {
		t.Fatalf("expected carol list status 200, got %d with body %s", carolListW.Code, carolListW.Body.String())
	}

	var carolMoments []struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(carolListW.Body.Bytes(), &carolMoments); err != nil {
		t.Fatalf("expected valid carol moment list json, got error: %v", err)
	}
	if len(carolMoments) != 0 {
		t.Fatalf("expected carol to see 0 moments, got %d", len(carolMoments))
	}
}
