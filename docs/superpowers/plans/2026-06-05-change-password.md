# Change Password Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a logged-in change-password flow with a dedicated security page, old-password verification, password-hash update, and forced re-login after success.

**Architecture:** The backend will add a new authenticated `PUT /api/users/me/password` endpoint in the existing auth handler/service/repository chain. The frontend will add a dedicated `/security` view linked from the profile page, submit the password-change form to the new API, then clear the current session and redirect back to login with a one-time success message.

**Tech Stack:** Go, Gin, GORM, bcrypt, SQLite router tests, Vue 3, TypeScript, Pinia, Vue Router, Axios, Vite

---

## File structure map

### Backend files
- Modify: `server/internal/repository/user_memory.go` — extend `UserRepository` with password-hash update support and implement it for in-memory users
- Modify: `server/internal/repository/user_gorm.go` — implement password-hash persistence for GORM-backed users
- Modify: `server/internal/service/auth.go` — add change-password input, validation, bcrypt old-password verification, and password-hash update orchestration
- Modify: `server/internal/handler/auth_user.go` — bind password-change JSON and map service errors to HTTP responses
- Modify: `server/internal/router/router.go` — register `PUT /api/users/me/password`
- Test: `server/internal/router/auth_user_test.go` — add success/failure coverage plus old-password/new-password login assertions

### Frontend files
- Modify: `web/src/router/index.ts` — register authenticated `/security` route
- Modify: `web/src/views/ProfileView.vue` — add a navigation entry to the dedicated security page
- Create: `web/src/views/SecurityView.vue` — render password form, validate input, call API, clear session, redirect to login with success message
- Modify: `web/src/views/LoginView.vue` — read one-time `message` query text and display it after redirect

---

### Task 1: Add backend tests for password change behavior

**Files:**
- Modify: `server/internal/router/auth_user_test.go`
- Test: `server/internal/router/auth_user_test.go`

- [ ] **Step 1: Write the failing success-path test**

Add this test to `server/internal/router/auth_user_test.go` near the existing auth handler tests:

```go
func TestUserCanChangeOwnPasswordAndMustUseNewPasswordAfterward(t *testing.T) {
	r := newTestRouter(t)
	token := registerAndLogin(t, r, "alice")

	changeBody := []byte(`{"oldPassword":"secret123","newPassword":"newsecret456","confirmNewPassword":"newsecret456"}`)
	changeReq := httptest.NewRequest(http.MethodPut, "/api/users/me/password", bytes.NewReader(changeBody))
	changeReq.Header.Set("Content-Type", "application/json")
	changeReq.Header.Set("Authorization", "Bearer "+token)
	changeW := httptest.NewRecorder()

	r.ServeHTTP(changeW, changeReq)

	if changeW.Code != http.StatusOK {
		t.Fatalf("expected password update status 200, got %d with body %s", changeW.Code, changeW.Body.String())
	}

	var changeResp struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(changeW.Body.Bytes(), &changeResp); err != nil {
		t.Fatalf("expected valid password update response json, got error: %v", err)
	}
	if changeResp.Message != "password updated" {
		t.Fatalf("expected success message password updated, got %q", changeResp.Message)
	}

	oldLoginBody := []byte(`{"username":"alice","password":"secret123"}`)
	oldLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(oldLoginBody))
	oldLoginReq.Header.Set("Content-Type", "application/json")
	oldLoginW := httptest.NewRecorder()
	 r.ServeHTTP(oldLoginW, oldLoginReq)

	if oldLoginW.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password login status 401, got %d with body %s", oldLoginW.Code, oldLoginW.Body.String())
	}

	newLoginBody := []byte(`{"username":"alice","password":"newsecret456"}`)
	newLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(newLoginBody))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginW := httptest.NewRecorder()
	r.ServeHTTP(newLoginW, newLoginReq)

	if newLoginW.Code != http.StatusOK {
		t.Fatalf("expected new password login status 200, got %d with body %s", newLoginW.Code, newLoginW.Body.String())
	}
}
```

- [ ] **Step 2: Write the failing validation tests**

Add this test below the success case:

```go
func TestUserCannotChangePasswordWithInvalidPayload(t *testing.T) {
	r := newTestRouter(t)
	token := registerAndLogin(t, r, "bobby")

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "wrong old password",
			body: `{"oldPassword":"wrong123","newPassword":"newsecret456","confirmNewPassword":"newsecret456"}`,
		},
		{
			name: "mismatched confirm password",
			body: `{"oldPassword":"secret123","newPassword":"newsecret456","confirmNewPassword":"different456"}`,
		},
		{
			name: "short new password",
			body: `{"oldPassword":"secret123","newPassword":"12345","confirmNewPassword":"12345"}`,
		},
		{
			name: "same old and new password",
			body: `{"oldPassword":"secret123","newPassword":"secret123","confirmNewPassword":"secret123"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/users/me/password", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected password update status 400, got %d with body %s", w.Code, w.Body.String())
			}
		})
	}
}
```

- [ ] **Step 3: Run the targeted router tests to verify they fail**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run 'TestUserCanChangeOwnPasswordAndMustUseNewPasswordAfterward|TestUserCannotChangePasswordWithInvalidPayload'
```

Expected: FAIL because `PUT /api/users/me/password` is not registered yet.

- [ ] **Step 4: Commit the failing-test checkpoint**

```bash
cd /Users/zyb/bishe/whr-im && git add server/internal/router/auth_user_test.go && git commit -m "test: add password change router coverage"
```

### Task 2: Implement backend repository, service, handler, and route

**Files:**
- Modify: `server/internal/repository/user_memory.go:13-19,103-117`
- Modify: `server/internal/repository/user_gorm.go:74-84`
- Modify: `server/internal/service/auth.go:18-28,57-61,195-217`
- Modify: `server/internal/handler/auth_user.go:46-49,151-173`
- Modify: `server/internal/router/router.go:148-152`
- Test: `server/internal/router/auth_user_test.go`

- [ ] **Step 1: Extend the repository interface with password-hash update support**

In `server/internal/repository/user_memory.go`, update the interface and add the in-memory method:

```go
type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint64) (*model.User, error)
	UpdateProfile(userID uint64, nickname string, gender int, signature string, avatar string, homepageSkin string, avatarAccessory string, titleBadge string, homepageBackground string, homepageLayout string) (*model.User, error)
	UpdatePublicKey(userID uint64, publicKey string, algorithm string) (*model.User, error)
	UpdatePasswordHash(userID uint64, passwordHash string) (*model.User, error)
}

func (r *InMemoryUserRepository) UpdatePasswordHash(userID uint64, passwordHash string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	user.PasswordHash = passwordHash

	copyUser := *user
	return &copyUser, nil
}
```

- [ ] **Step 2: Implement the GORM password-hash update**

In `server/internal/repository/user_gorm.go`, add:

```go
func (r *GormUserRepository) UpdatePasswordHash(userID uint64, passwordHash string) (*model.User, error) {
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", passwordHash)
	if result.Error != nil {
		return nil, result.Error
	}
	return r.FindByID(userID)
}
```

- [ ] **Step 3: Add service input types and change-password rules**

In `server/internal/service/auth.go`, add a new error and input type near the existing auth types:

```go
var ErrInvalidPasswordChange = errors.New("invalid password change")

type ChangePasswordInput struct {
	OldPassword        string
	NewPassword        string
	ConfirmNewPassword string
}
```

Then add the service method below `UpdatePublicKey`:

```go
func (s *AuthService) ChangePassword(userID uint64, input ChangePasswordInput) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	if input.OldPassword == "" {
		return fmt.Errorf("%w: old password is required", ErrInvalidPasswordChange)
	}
	if len(input.NewPassword) < minPasswordLength || len(input.NewPassword) > maxPasswordLength {
		return fmt.Errorf("%w: password length must be between 6 and 20", ErrInvalidPasswordChange)
	}
	if input.NewPassword != input.ConfirmNewPassword {
		return fmt.Errorf("%w: passwords do not match", ErrInvalidPasswordChange)
	}
	if input.OldPassword == input.NewPassword {
		return fmt.Errorf("%w: new password must be different from old password", ErrInvalidPasswordChange)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.OldPassword)); err != nil {
		return fmt.Errorf("%w: old password is incorrect", ErrInvalidPasswordChange)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.repo.UpdatePasswordHash(userID, string(hash))
	return err
}
```

- [ ] **Step 4: Add handler request binding and HTTP mapping**

In `server/internal/handler/auth_user.go`, add the request type:

```go
type changePasswordRequest struct {
	OldPassword        string `json:"oldPassword"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
}
```

Then add the handler method below `UpdateMyPublicKey`:

```go
func (h *AuthUserHandler) UpdateMyPassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	userID := c.MustGet("userID").(uint64)
	if err := h.authService.ChangePassword(userID, service.ChangePasswordInput{
		OldPassword:        req.OldPassword,
		NewPassword:        req.NewPassword,
		ConfirmNewPassword: req.ConfirmNewPassword,
	}); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, service.ErrInvalidPasswordChange) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}
```

- [ ] **Step 5: Wire the authenticated route**

In `server/internal/router/router.go`, add the route beside the existing `/users/me` endpoints:

```go
		// PUT /api/users/me/password: 修改当前登录用户密码，成功后前端需重新登录。
		authed.PUT("/users/me/password", authHandler.UpdateMyPassword)
```

- [ ] **Step 6: Run the targeted password-change tests**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router -run 'TestUserCanChangeOwnPasswordAndMustUseNewPasswordAfterward|TestUserCannotChangePasswordWithInvalidPayload'
```

Expected: PASS.

- [ ] **Step 7: Run the full auth router test file to guard regressions**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router
```

Expected: PASS.

- [ ] **Step 8: Commit the backend implementation**

```bash
cd /Users/zyb/bishe/whr-im && git add server/internal/repository/user_memory.go server/internal/repository/user_gorm.go server/internal/service/auth.go server/internal/handler/auth_user.go server/internal/router/router.go server/internal/router/auth_user_test.go && git commit -m "feat: add authenticated password change api"
```

### Task 3: Add the dedicated security page and login redirect message

**Files:**
- Modify: `web/src/router/index.ts:1-35`
- Modify: `web/src/views/ProfileView.vue:1-153`
- Create: `web/src/views/SecurityView.vue`
- Modify: `web/src/views/LoginView.vue:1-87`

- [ ] **Step 1: Add the authenticated `/security` route**

In `web/src/router/index.ts`, import the new view and register the route:

```ts
import SecurityView from '../views/SecurityView.vue'
```

```ts
    { path: '/profile', component: ProfileView, meta: { requiresAuth: true } },
    { path: '/security', component: SecurityView, meta: { requiresAuth: true } },
    { path: '/decorations', component: DecorationsView, meta: { requiresAuth: true } },
```

- [ ] **Step 2: Add a security entry to the profile page**

In `web/src/views/ProfileView.vue`, import `RouterLink` and add a secondary action next to the save button:

```ts
import { RouterLink } from 'vue-router'
```

```vue
<div class="profile-actions">
  <RouterLink class="apple-button secondary-button" to="/security">账号安全</RouterLink>
  <button class="apple-button" :disabled="loading" @click="saveProfile">保存更改</button>
</div>
```

Add the matching scoped style:

```css
.secondary-button {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
}
```

- [ ] **Step 3: Create the security view with validation and forced re-login flow**

Create `web/src/views/SecurityView.vue` with this content:

```vue
<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppNav from '../components/AppNav.vue'
import { http } from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmNewPassword: ''
})

function validate() {
  if (!form.oldPassword) {
    return '请输入旧密码'
  }
  if (form.newPassword.length < 6 || form.newPassword.length > 20) {
    return '密码长度需在 6 到 20 位之间'
  }
  if (form.newPassword !== form.confirmNewPassword) {
    return '两次输入的新密码不一致'
  }
  if (form.oldPassword === form.newPassword) {
    return '新密码不能与旧密码相同'
  }
  return ''
}

async function changePassword() {
  errorMessage.value = ''
  message.value = ''
  const validationMessage = validate()
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }

  loading.value = true
  try {
    await http.put('/users/me/password', form)
    authStore.clearSession()
    await router.push({
      path: '/login',
      query: { message: '密码修改成功，请重新登录' }
    })
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-shell apple-page">
    <AppNav />
    <section class="card apple-panel security-shell">
      <div class="security-header">
        <p class="apple-label">Security</p>
        <h1>账号安全</h1>
        <p class="muted">修改密码后将立即退出当前登录状态，请使用新密码重新登录。</p>
      </div>
      <p v-if="message" class="status-text success">{{ message }}</p>
      <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
      <label>
        <span class="apple-label">旧密码</span>
        <input v-model="form.oldPassword" class="apple-input" type="password" placeholder="请输入旧密码" />
      </label>
      <label>
        <span class="apple-label">新密码</span>
        <input v-model="form.newPassword" class="apple-input" type="password" placeholder="请输入新密码（6-20 位）" />
      </label>
      <label>
        <span class="apple-label">确认新密码</span>
        <input v-model="form.confirmNewPassword" class="apple-input" type="password" placeholder="请再次输入新密码" />
      </label>
      <div class="security-actions">
        <button class="apple-button" :disabled="loading" @click="changePassword">确认修改</button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.security-shell {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.security-header h1 {
  margin: 8px 0 10px;
  font-size: 38px;
  letter-spacing: -0.03em;
}

.security-header p {
  margin: 0;
}

label {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.security-actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .security-actions {
    justify-content: stretch;
  }
}
</style>
```

- [ ] **Step 4: Display the one-time success message on the login page**

In `web/src/views/LoginView.vue`, switch to using the route query and initialize the login error area from `message`:

```ts
import { reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
```

```ts
const route = useRoute()
const errorMessage = ref(typeof route.query.message === 'string' ? route.query.message : '')
```

Keep the existing username watcher, then add this second watcher:

```ts
watch(
  () => route.query.message,
  (value) => {
    errorMessage.value = typeof value === 'string' ? value : ''
  }
)
```

This keeps the current single status area and avoids adding another component for one-time redirect feedback.

- [ ] **Step 5: Run the frontend build to verify the new route and view compile**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && npm run build
```

Expected: PASS with the Vite production bundle output.

- [ ] **Step 6: Commit the frontend flow**

```bash
cd /Users/zyb/bishe/whr-im && git add web/src/router/index.ts web/src/views/ProfileView.vue web/src/views/SecurityView.vue web/src/views/LoginView.vue && git commit -m "feat: add change password security page"
```

### Task 4: Verify the integrated flow end-to-end

**Files:**
- Test: `server/internal/router/auth_user_test.go`
- Test: `web/src/views/SecurityView.vue`

- [ ] **Step 1: Re-run the backend router suite**

Run:
```bash
cd /Users/zyb/bishe/whr-im/server && go test ./internal/router
```

Expected: PASS.

- [ ] **Step 2: Re-run the frontend build**

Run:
```bash
cd /Users/zyb/bishe/whr-im/web && npm run build
```

Expected: PASS.

- [ ] **Step 3: Sanity-check the full user flow manually**

Run the app locally and verify this exact path:

```text
1. Register/login as a test user with password secret123.
2. Open /profile and confirm the “账号安全” entry is visible.
3. Open /security.
4. Submit wrong old password and confirm the page stays put with an error.
5. Submit oldPassword=secret123, newPassword=newsecret456, confirmNewPassword=newsecret456.
6. Confirm the app redirects to /login and shows “密码修改成功，请重新登录”.
7. Confirm login with secret123 now fails.
8. Confirm login with newsecret456 succeeds.
```

- [ ] **Step 4: Commit the verification checkpoint if any manual-fix changes were needed**

```bash
cd /Users/zyb/bishe/whr-im && git status --short
```

Expected: no output if no extra fixes were needed. If fixes were required during verification, stage only those files and create a final commit describing the correction.
