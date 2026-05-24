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
			ID                 uint64 `json:"id"`
			Username           string `json:"username"`
			Nickname           string `json:"nickname"`
			Avatar             string `json:"avatar"`
			PublicKey          string `json:"publicKey"`
			PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
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
	if registerResp.User.PublicKey != "" {
		t.Fatalf("expected empty publicKey on register, got %q", registerResp.User.PublicKey)
	}
	if registerResp.User.PublicKeyAlgorithm != "" {
		t.Fatalf("expected empty publicKeyAlgorithm on register, got %q", registerResp.User.PublicKeyAlgorithm)
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
			Username           string `json:"username"`
			PublicKey          string `json:"publicKey"`
			PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("expected valid login response json, got error: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("expected jwt token to be returned")
	}
	if loginResp.User.PublicKey != "" {
		t.Fatalf("expected empty publicKey on login, got %q", loginResp.User.PublicKey)
	}
	if loginResp.User.PublicKeyAlgorithm != "" {
		t.Fatalf("expected empty publicKeyAlgorithm on login, got %q", loginResp.User.PublicKeyAlgorithm)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	meW := httptest.NewRecorder()

	r.ServeHTTP(meW, meReq)

	if meW.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d with body %s", meW.Code, meW.Body.String())
	}

	var meResp struct {
		Username           string `json:"username"`
		Nickname           string `json:"nickname"`
		Avatar             string `json:"avatar"`
		Signature          string `json:"signature"`
		PublicKey          string `json:"publicKey"`
		PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
	}
	if err := json.Unmarshal(meW.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("expected valid me response json, got error: %v", err)
	}
	if meResp.Username != "alice" {
		t.Fatalf("expected me username alice, got %q", meResp.Username)
	}
	if meResp.PublicKey != "" {
		t.Fatalf("expected empty publicKey in profile, got %q", meResp.PublicKey)
	}
	if meResp.PublicKeyAlgorithm != "" {
		t.Fatalf("expected empty publicKeyAlgorithm in profile, got %q", meResp.PublicKeyAlgorithm)
	}

	updateFemaleBody := []byte(`{"nickname":"Alice","gender":0,"signature":"hello im"}`)
	updateFemaleReq := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewReader(updateFemaleBody))
	updateFemaleReq.Header.Set("Content-Type", "application/json")
	updateFemaleReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	updateFemaleW := httptest.NewRecorder()

	r.ServeHTTP(updateFemaleW, updateFemaleReq)

	if updateFemaleW.Code != http.StatusOK {
		t.Fatalf("expected female update status 200, got %d with body %s", updateFemaleW.Code, updateFemaleW.Body.String())
	}

	var updateFemaleResp struct {
		Nickname           string `json:"nickname"`
		Gender             int    `json:"gender"`
		Signature          string `json:"signature"`
		Avatar             string `json:"avatar"`
		PublicKey          string `json:"publicKey"`
		PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
	}
	if err := json.Unmarshal(updateFemaleW.Body.Bytes(), &updateFemaleResp); err != nil {
		t.Fatalf("expected valid female update response json, got error: %v", err)
	}
	if updateFemaleResp.Nickname != "Alice" {
		t.Fatalf("expected updated nickname Alice, got %q", updateFemaleResp.Nickname)
	}
	if updateFemaleResp.Gender != 0 {
		t.Fatalf("expected updated gender 0 for female, got %d", updateFemaleResp.Gender)
	}
	if updateFemaleResp.Signature != "hello im" {
		t.Fatalf("expected updated signature, got %q", updateFemaleResp.Signature)
	}
	if updateFemaleResp.Avatar == "" {
		t.Fatal("expected avatar to remain populated")
	}
	if updateFemaleResp.PublicKey != "" {
		t.Fatalf("expected publicKey to remain empty after profile update, got %q", updateFemaleResp.PublicKey)
	}
	if updateFemaleResp.PublicKeyAlgorithm != "" {
		t.Fatalf("expected publicKeyAlgorithm to remain empty after profile update, got %q", updateFemaleResp.PublicKeyAlgorithm)
	}

	updateMaleBody := []byte(`{"nickname":"Alice","gender":1,"signature":"hello im"}`)
	updateMaleReq := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewReader(updateMaleBody))
	updateMaleReq.Header.Set("Content-Type", "application/json")
	updateMaleReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	updateMaleW := httptest.NewRecorder()

	r.ServeHTTP(updateMaleW, updateMaleReq)

	if updateMaleW.Code != http.StatusOK {
		t.Fatalf("expected male update status 200, got %d with body %s", updateMaleW.Code, updateMaleW.Body.String())
	}

	var updateMaleResp struct {
		Gender             int    `json:"gender"`
		PublicKey          string `json:"publicKey"`
		PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
	}
	if err := json.Unmarshal(updateMaleW.Body.Bytes(), &updateMaleResp); err != nil {
		t.Fatalf("expected valid male update response json, got error: %v", err)
	}
	if updateMaleResp.Gender != 1 {
		t.Fatalf("expected updated gender 1 for male, got %d", updateMaleResp.Gender)
	}
	if updateMaleResp.PublicKey != "" {
		t.Fatalf("expected publicKey to remain empty after second profile update, got %q", updateMaleResp.PublicKey)
	}
	if updateMaleResp.PublicKeyAlgorithm != "" {
		t.Fatalf("expected publicKeyAlgorithm to remain empty after second profile update, got %q", updateMaleResp.PublicKeyAlgorithm)
	}
}

func TestUserCanUpdateOwnPublicKey(t *testing.T) {
	r := newTestRouter(t)
	token := registerAndLogin(t, r, "alice")

	updateBody := []byte(`{"publicKey":"alice-public-key","algorithm":"rsa-oaep-sha256"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/users/me/public-key", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateW := httptest.NewRecorder()

	r.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("expected public key update status 200, got %d with body %s", updateW.Code, updateW.Body.String())
	}

	var updateResp struct {
		PublicKey          string `json:"publicKey"`
		PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
	}
	if err := json.Unmarshal(updateW.Body.Bytes(), &updateResp); err != nil {
		t.Fatalf("expected valid public key update response json, got error: %v", err)
	}
	if updateResp.PublicKey != "alice-public-key" {
		t.Fatalf("expected updated publicKey, got %q", updateResp.PublicKey)
	}
	if updateResp.PublicKeyAlgorithm != "rsa-oaep-sha256" {
		t.Fatalf("expected updated publicKeyAlgorithm rsa-oaep-sha256, got %q", updateResp.PublicKeyAlgorithm)
	}
}

func TestUserCannotUpdatePublicKeyWithInvalidPayload(t *testing.T) {
	r := newTestRouter(t)
	token := registerAndLogin(t, r, "bobby")

	t.Run("empty public key returns 400", func(t *testing.T) {
		updateBody := []byte(`{"publicKey":"","algorithm":"rsa-oaep-sha256"}`)
		updateReq := httptest.NewRequest(http.MethodPut, "/api/users/me/public-key", bytes.NewReader(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateW := httptest.NewRecorder()

		r.ServeHTTP(updateW, updateReq)

		if updateW.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for empty publicKey, got %d with body %s", updateW.Code, updateW.Body.String())
		}
	})

	t.Run("empty algorithm returns 400", func(t *testing.T) {
		updateBody := []byte(`{"publicKey":"bob-public-key","algorithm":""}`)
		updateReq := httptest.NewRequest(http.MethodPut, "/api/users/me/public-key", bytes.NewReader(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateW := httptest.NewRecorder()

		r.ServeHTTP(updateW, updateReq)

		if updateW.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for empty algorithm, got %d with body %s", updateW.Code, updateW.Body.String())
		}
	})

	t.Run("unsupported algorithm returns 400", func(t *testing.T) {
		updateBody := []byte(`{"publicKey":"bob-public-key","algorithm":"x25519"}`)
		updateReq := httptest.NewRequest(http.MethodPut, "/api/users/me/public-key", bytes.NewReader(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateW := httptest.NewRecorder()

		r.ServeHTTP(updateW, updateReq)

		if updateW.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for unsupported algorithm, got %d with body %s", updateW.Code, updateW.Body.String())
		}
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		updateBody := []byte(`not-json`)
		updateReq := httptest.NewRequest(http.MethodPut, "/api/users/me/public-key", bytes.NewReader(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateW := httptest.NewRecorder()

		r.ServeHTTP(updateW, updateReq)

		if updateW.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for invalid JSON, got %d with body %s", updateW.Code, updateW.Body.String())
		}
	})
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

func TestRegisterRejectsInvalidUsernameAndPasswordLength(t *testing.T) {
	r := newTestRouter(t)

	shortUsernameBody := []byte(`{"username":"abc","password":"secret123","confirmPassword":"secret123"}`)
	shortUsernameReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(shortUsernameBody))
	shortUsernameReq.Header.Set("Content-Type", "application/json")
	shortUsernameW := httptest.NewRecorder()
	r.ServeHTTP(shortUsernameW, shortUsernameReq)
	if shortUsernameW.Code != http.StatusBadRequest {
		t.Fatalf("expected short username status 400, got %d with body %s", shortUsernameW.Code, shortUsernameW.Body.String())
	}
	if !bytes.Contains(shortUsernameW.Body.Bytes(), []byte("username length must be between 4 and 20")) {
		t.Fatalf("expected username length error, got %s", shortUsernameW.Body.String())
	}

	shortPasswordBody := []byte(`{"username":"validuser","password":"12345","confirmPassword":"12345"}`)
	shortPasswordReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(shortPasswordBody))
	shortPasswordReq.Header.Set("Content-Type", "application/json")
	shortPasswordW := httptest.NewRecorder()
	r.ServeHTTP(shortPasswordW, shortPasswordReq)
	if shortPasswordW.Code != http.StatusBadRequest {
		t.Fatalf("expected short password status 400, got %d with body %s", shortPasswordW.Code, shortPasswordW.Body.String())
	}
	if !bytes.Contains(shortPasswordW.Body.Bytes(), []byte("password length must be between 6 and 20")) {
		t.Fatalf("expected password length error, got %s", shortPasswordW.Body.String())
	}
}

func TestLoginRejectsInvalidUsernameAndPasswordLength(t *testing.T) {
	r := newTestRouter(t)

	shortUsernameBody := []byte(`{"username":"abc","password":"secret123"}`)
	shortUsernameReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(shortUsernameBody))
	shortUsernameReq.Header.Set("Content-Type", "application/json")
	shortUsernameW := httptest.NewRecorder()
	r.ServeHTTP(shortUsernameW, shortUsernameReq)
	if shortUsernameW.Code != http.StatusBadRequest {
		t.Fatalf("expected short username login status 400, got %d with body %s", shortUsernameW.Code, shortUsernameW.Body.String())
	}
	if !bytes.Contains(shortUsernameW.Body.Bytes(), []byte("username length must be between 4 and 20")) {
		t.Fatalf("expected username length error, got %s", shortUsernameW.Body.String())
	}

	shortPasswordBody := []byte(`{"username":"validuser","password":"12345"}`)
	shortPasswordReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(shortPasswordBody))
	shortPasswordReq.Header.Set("Content-Type", "application/json")
	shortPasswordW := httptest.NewRecorder()
	r.ServeHTTP(shortPasswordW, shortPasswordReq)
	if shortPasswordW.Code != http.StatusBadRequest {
		t.Fatalf("expected short password login status 400, got %d with body %s", shortPasswordW.Code, shortPasswordW.Body.String())
	}
	if !bytes.Contains(shortPasswordW.Body.Bytes(), []byte("password length must be between 6 and 20")) {
		t.Fatalf("expected password length error, got %s", shortPasswordW.Body.String())
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
