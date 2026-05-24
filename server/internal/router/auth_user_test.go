package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whr-im/server/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegisterLoginAndProfileFlow(t *testing.T) {
	r := newTestRouter(t)

	registerBody := []byte(`{"username":"alice","password":"secret123","confirmPassword":"secret123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()

	r.ServeHTTP(registerW, registerReq)

	if registerW.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d with body %s", registerW.Code, registerW.Body.String())
	}

	var registerResp struct {
		User struct {
			ID       uint64  `json:"id"`
			Username string `json:"username"`
			Nickname string `json:"nickname"`
			Avatar   string `json:"avatar"`
		} `json:"user"`
	}
	if err := json.Unmarshal(registerW.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("expected valid register response json, got error: %v", err)
	}

	if registerResp.User.Username != "alice" {
		t.Fatalf("expected username alice, got %q", registerResp.User.Username)
	}
	if registerResp.User.Nickname != "alice" {
		t.Fatalf("expected default nickname alice, got %q", registerResp.User.Nickname)
	}
	if registerResp.User.Avatar == "" {
		t.Fatal("expected default avatar to be assigned")
	}

	loginBody := []byte(`{"username":"alice","password":"secret123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d with body %s", loginW.Code, loginW.Body.String())
	}

	var loginResp struct {
		Token string `json:"token"`
		User  struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("expected valid login response json, got error: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("expected jwt token to be returned")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	meW := httptest.NewRecorder()

	r.ServeHTTP(meW, meReq)

	if meW.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d with body %s", meW.Code, meW.Body.String())
	}

	var meResp struct {
		Username  string `json:"username"`
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(meW.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("expected valid me response json, got error: %v", err)
	}
	if meResp.Username != "alice" {
		t.Fatalf("expected me username alice, got %q", meResp.Username)
	}

	updateBody := []byte(`{"nickname":"Alice","gender":2,"signature":"hello im"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	updateW := httptest.NewRecorder()

	r.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d with body %s", updateW.Code, updateW.Body.String())
	}

	var updateResp struct {
		Nickname  string `json:"nickname"`
		Gender    int    `json:"gender"`
		Signature string `json:"signature"`
		Avatar    string `json:"avatar"`
	}
	if err := json.Unmarshal(updateW.Body.Bytes(), &updateResp); err != nil {
		t.Fatalf("expected valid update response json, got error: %v", err)
	}
	if updateResp.Nickname != "Alice" {
		t.Fatalf("expected updated nickname Alice, got %q", updateResp.Nickname)
	}
	if updateResp.Gender != 2 {
		t.Fatalf("expected updated gender 2, got %d", updateResp.Gender)
	}
	if updateResp.Signature != "hello im" {
		t.Fatalf("expected updated signature, got %q", updateResp.Signature)
	}
	if updateResp.Avatar == "" {
		t.Fatal("expected avatar to remain populated")
	}
}

func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	r := newTestRouter(t)
	body := []byte(`{"username":"alice","password":"secret123","confirmPassword":"secret123"}`)

	firstReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	firstReq.Header.Set("Content-Type", "application/json")
	firstW := httptest.NewRecorder()
	r.ServeHTTP(firstW, firstReq)
	if firstW.Code != http.StatusCreated {
		t.Fatalf("expected first register status 201, got %d", firstW.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	secondReq.Header.Set("Content-Type", "application/json")
	secondW := httptest.NewRecorder()
	r.ServeHTTP(secondW, secondReq)

	if secondW.Code != http.StatusConflict {
		t.Fatalf("expected duplicate register status 409, got %d with body %s", secondW.Code, secondW.Body.String())
	}
}

func newTestRouter(t *testing.T) http.Handler {
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
	messageRepo, err := repository.NewGormMessageRepository(db)
	if err != nil {
		t.Fatalf("failed to create gorm message repository: %v", err)
	}

	return NewWithRepositories(userRepo, friendRepo, messageRepo)
}
