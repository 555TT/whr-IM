# Global Homepage Skin Expansion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand the existing homepage skin feature so the selected preset skin affects the navigation bar, profile page background/header, moments page cards and primary accent, friend requests page accent, and chat page background.

**Architecture:** Keep `homepageSkin` as the single persisted user preference and derive all page styling from a shared frontend skin token map. Move repeated preset definitions out of individual views into one reusable module, then apply skin classes/CSS variables in each target page without changing login/register or global input semantics. Preserve danger button coloring and avoid touching `server/config/config.yaml`.

**Tech Stack:** Vue 3, TypeScript, Pinia, Vue Router, scoped CSS, existing Go backend (no new API or schema change for this scope)

---

## File structure

**Create:**
- `web/src/constants/homepageSkins.ts` — shared preset skin definitions and helpers for CSS class naming / lookup.

**Modify:**
- `web/src/stores/auth.ts` — ensure current user skin field remains typed and available to all views/components.
- `web/src/components/AppNav.vue` — apply current user skin class to nav shell and active-link accent.
- `web/src/views/ProfileView.vue` — reuse shared skin constants and apply current user skin to page shell / header block.
- `web/src/views/MomentsView.vue` — apply current user skin to composer card, feed cards, and primary accent hooks.
- `web/src/views/FriendRequestsView.vue` — apply current user skin to top form card and action emphasis.
- `web/src/views/ChatView.vue` — apply current user skin to chat page background atmosphere only, without reworking message bubble behavior.
- `web/src/views/UserHomepageView.vue` — switch to shared skin constants so homepage and the rest of the app use one source of truth.
- `web/src/styles/main.css` — add shared CSS custom properties / utility classes for skin accent variables if needed.

**Test / verify:**
- `cd web && npm run build`
- manual UI verification across nav/profile/moments/friend requests/chat/homepage

---

### Task 1: Centralize homepage skin presets

**Files:**
- Create: `web/src/constants/homepageSkins.ts`
- Modify: `web/src/views/ProfileView.vue:1-203`
- Modify: `web/src/views/UserHomepageView.vue:1-260`

- [ ] **Step 1: Write the failing build expectation**

The current code duplicates preset skin definitions in multiple views. The intended API is a shared constant module used by both profile and homepage pages:

```ts
export type HomepageSkinKey = 'aurora' | 'sunset' | 'galaxy' | 'mint' | 'peach'

export interface HomepageSkinDefinition {
  key: HomepageSkinKey
  label: string
  previewClass: string
  surfaceClass: string
  accentClass: string
}

export const homepageSkins: HomepageSkinDefinition[] = [
  { key: 'aurora', label: '极光', previewClass: 'skin-aurora', surfaceClass: 'theme-aurora', accentClass: 'theme-accent-aurora' },
  { key: 'sunset', label: '日落', previewClass: 'skin-sunset', surfaceClass: 'theme-sunset', accentClass: 'theme-accent-sunset' },
  { key: 'galaxy', label: '星河', previewClass: 'skin-galaxy', surfaceClass: 'theme-galaxy', accentClass: 'theme-accent-galaxy' },
  { key: 'mint', label: '薄荷', previewClass: 'skin-mint', surfaceClass: 'theme-mint', accentClass: 'theme-accent-mint' },
  { key: 'peach', label: '蜜桃', previewClass: 'skin-peach', surfaceClass: 'theme-peach', accentClass: 'theme-accent-peach' }
]
```

- [ ] **Step 2: Create the shared skin module**

```ts
// web/src/constants/homepageSkins.ts
export type HomepageSkinKey = 'aurora' | 'sunset' | 'galaxy' | 'mint' | 'peach'

export interface HomepageSkinDefinition {
  key: HomepageSkinKey
  label: string
  previewClass: string
  surfaceClass: string
  accentClass: string
}

export const homepageSkins: HomepageSkinDefinition[] = [
  { key: 'aurora', label: '极光', previewClass: 'skin-aurora', surfaceClass: 'theme-aurora', accentClass: 'theme-accent-aurora' },
  { key: 'sunset', label: '日落', previewClass: 'skin-sunset', surfaceClass: 'theme-sunset', accentClass: 'theme-accent-sunset' },
  { key: 'galaxy', label: '星河', previewClass: 'skin-galaxy', surfaceClass: 'theme-galaxy', accentClass: 'theme-accent-galaxy' },
  { key: 'mint', label: '薄荷', previewClass: 'skin-mint', surfaceClass: 'theme-mint', accentClass: 'theme-accent-mint' },
  { key: 'peach', label: '蜜桃', previewClass: 'skin-peach', surfaceClass: 'theme-peach', accentClass: 'theme-accent-peach' }
]

export function resolveHomepageSkin(key?: string): HomepageSkinDefinition {
  return homepageSkins.find((item) => item.key === key) || homepageSkins[0]
}
```

- [ ] **Step 3: Refactor profile page to reuse shared presets**

Replace local `homepageSkins` with:

```ts
import { homepageSkins, resolveHomepageSkin } from '../constants/homepageSkins'
```

Add computed class hook:

```ts
function currentSkin() {
  return resolveHomepageSkin(profile.homepageSkin)
}
```

Apply it to the main section:

```vue
<section class="card apple-panel profile-shell" :class="currentSkin().surfaceClass">
```

- [ ] **Step 4: Refactor user homepage to reuse shared presets**

Replace local map with:

```ts
import { resolveHomepageSkin } from '../constants/homepageSkins'

function homepageSkinClass() {
  return resolveHomepageSkin(profile.value?.homepageSkin).surfaceClass
}
```

- [ ] **Step 5: Run build to verify refactor compiles**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`
Expected: PASS, Vite build succeeds with no TypeScript errors.

- [ ] **Step 6: Commit**

```bash
git add web/src/constants/homepageSkins.ts web/src/views/ProfileView.vue web/src/views/UserHomepageView.vue
git commit -m "refactor: centralize homepage skin presets"
```

---

### Task 2: Apply skin to shared navigation and profile page background

**Files:**
- Modify: `web/src/components/AppNav.vue:1-149`
- Modify: `web/src/views/ProfileView.vue:59-202`
- Modify: `web/src/styles/main.css:1-180`
- Modify: `web/src/stores/auth.ts:1-53`

- [ ] **Step 1: Write the failing design expectation**

The selected skin should affect nav and profile styling via shared user state:

```ts
const navSkin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
```

And in template:

```vue
<nav class="nav card" :class="[navSkin.surfaceClass, navSkin.accentClass]">
```

- [ ] **Step 2: Add global skin utility classes and CSS variables**

Append to `web/src/styles/main.css`:

```css
.theme-aurora {
  --skin-accent: #2193b0;
  --skin-accent-soft: rgba(33, 147, 176, 0.14);
  --skin-surface: linear-gradient(135deg, #6dd5ed 0%, #2193b0 100%);
}

.theme-sunset {
  --skin-accent: #ec4899;
  --skin-accent-soft: rgba(236, 72, 153, 0.14);
  --skin-surface: linear-gradient(135deg, #f97316 0%, #ec4899 100%);
}

.theme-galaxy {
  --skin-accent: #7c3aed;
  --skin-accent-soft: rgba(124, 58, 237, 0.16);
  --skin-surface: linear-gradient(135deg, #312e81 0%, #7c3aed 55%, #ec4899 100%);
}

.theme-mint {
  --skin-accent: #14b8a6;
  --skin-accent-soft: rgba(20, 184, 166, 0.14);
  --skin-surface: linear-gradient(135deg, #34d399 0%, #14b8a6 100%);
}

.theme-peach {
  --skin-accent: #fb7185;
  --skin-accent-soft: rgba(251, 113, 133, 0.14);
  --skin-surface: linear-gradient(135deg, #fb7185 0%, #fdba74 100%);
}
```

- [ ] **Step 3: Apply skin in AppNav**

Use shared resolver:

```ts
import { computed } from 'vue'
import { resolveHomepageSkin } from '../constants/homepageSkins'

const navSkin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
```

And template:

```vue
<nav class="nav card" :class="[navSkin.surfaceClass, navSkin.accentClass]">
```

Then update styles to use CSS variables:

```css
.nav {
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
}

.nav a.router-link-active {
  background: var(--skin-accent-soft, rgba(0, 113, 227, 0.1));
  color: var(--skin-accent, #0071e3);
}
```

- [ ] **Step 4: Expand profile page background/header styling**

In `ProfileView.vue`, apply skin class to shell and update header to use the surface gradient as a hero block:

```vue
<section class="card apple-panel profile-shell" :class="currentSkin().surfaceClass">
  <div class="profile-header profile-hero">
```

Update CSS:

```css
.profile-shell {
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
}

.profile-hero {
  color: #fff;
}

.profile-hero .muted,
.profile-hero .apple-label {
  color: rgba(255, 255, 255, 0.86);
}
```

- [ ] **Step 5: Run build to verify nav/profile theme compiles**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add web/src/components/AppNav.vue web/src/views/ProfileView.vue web/src/styles/main.css web/src/stores/auth.ts
git commit -m "feat: apply homepage skin to nav and profile"
```

---

### Task 3: Extend skin to moments and friend requests

**Files:**
- Modify: `web/src/views/MomentsView.vue:1-396`
- Modify: `web/src/views/FriendRequestsView.vue:1-156`
- Modify: `web/src/styles/main.css:1-260`

- [ ] **Step 1: Write the failing design expectation**

Moments and friend request surfaces should consume the current user’s chosen skin:

```ts
const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
```

Moments template examples:

```vue
<div class="card apple-panel composer-card" :class="skin.surfaceClass">
<article v-for="item in moments" :key="item.id" class="card apple-panel moment-card" :class="skin.accentClass">
```

Friend requests examples:

```vue
<div class="card apple-panel request-form-card" :class="skin.surfaceClass">
<div class="card apple-panel request-list-card" :class="skin.accentClass">
```

- [ ] **Step 2: Apply theme hooks to MomentsView**

Add imports and computed resolver, then apply classes to:
- composer card surface
- moment cards accent frame
- primary action color hooks via CSS variable

Update scoped CSS to read:

```css
.composer-card {
  background: var(--skin-surface, rgba(255, 255, 255, 0.78));
  color: #fff;
}

.moment-card {
  box-shadow: 0 12px 28px var(--skin-accent-soft, rgba(15, 23, 42, 0.08));
}
```

Keep delete/danger buttons unchanged.

- [ ] **Step 3: Apply theme hooks to FriendRequestsView**

Add auth store + resolver, then:
- theme request form card with full surface gradient
- theme request list card header / accept emphasis with skin accent

Example import block:

```ts
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { resolveHomepageSkin } from '../constants/homepageSkins'
```

- [ ] **Step 4: Run build to verify social pages compile**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/views/MomentsView.vue web/src/views/FriendRequestsView.vue web/src/styles/main.css
git commit -m "feat: extend homepage skin to social pages"
```

---

### Task 4: Extend skin to chat background atmosphere

**Files:**
- Modify: `web/src/views/ChatView.vue`
- Modify: `web/src/styles/main.css`
- Test: manual verification in chat page

- [ ] **Step 1: Write the failing design expectation**

Chat should gain background atmosphere only, not bubble redesign:

```vue
<div class="page-shell apple-page chat-theme-shell" :class="skin.surfaceClass">
```

- [ ] **Step 2: Apply current user skin to ChatView container**

Import shared resolver and auth store if not already present:

```ts
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { resolveHomepageSkin } from '../constants/homepageSkins'

const skin = computed(() => resolveHomepageSkin(authStore.user?.homepageSkin))
```

Then apply the class on the outer shell or main content region.

- [ ] **Step 3: Add background-only CSS treatment**

Use variables to tint the chat page without restyling bubbles:

```css
.chat-theme-shell {
  position: relative;
}

.chat-theme-shell::before {
  content: '';
  position: fixed;
  inset: 0;
  background: radial-gradient(circle at top right, var(--skin-accent-soft, rgba(0, 113, 227, 0.1)), transparent 45%);
  pointer-events: none;
  z-index: 0;
}
```

Ensure actual chat content stays above it with `position: relative; z-index: 1;` on the main page container.

- [ ] **Step 4: Run build to verify chat theme compiles**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/views/ChatView.vue web/src/styles/main.css
git commit -m "feat: add homepage skin chat background"
```

---

### Task 5: Full verification

**Files:**
- Verify only

- [ ] **Step 1: Run backend tests**

Run: `cd "/Users/zyb/bishe/whr-im/server" && go test ./...`
Expected: PASS with 0 failures.

- [ ] **Step 2: Run frontend build**

Run: `cd "/Users/zyb/bishe/whr-im/web" && npm run build`
Expected: PASS.

- [ ] **Step 3: Manual verification checklist**

Check these scenarios in the browser:

```text
1. 在个人资料页切换皮肤并保存，刷新后仍保持
2. 导航栏高亮色和背景随皮肤变化
3. 个人资料页头部/背景随皮肤变化
4. 朋友圈发布卡片与动态卡片出现对应皮肤氛围
5. 好友申请页顶部卡片和按钮强调色随皮肤变化
6. 聊天页背景有皮肤氛围，但聊天气泡本身未被重做
7. 用户主页仍保持与当前皮肤一致
8. server/config/config.yaml 未参与本次提交
```

- [ ] **Step 4: Commit verification-safe finishing changes**

```bash
git status --short
git add web/src/constants/homepageSkins.ts web/src/components/AppNav.vue web/src/views/ProfileView.vue web/src/views/MomentsView.vue web/src/views/FriendRequestsView.vue web/src/views/UserHomepageView.vue web/src/views/ChatView.vue web/src/styles/main.css web/src/stores/auth.ts server/internal/model/user.go server/internal/service/auth.go server/internal/handler/auth_user.go server/internal/repository/user_gorm.go server/internal/repository/user_memory.go server/internal/router/auth_user_test.go server/internal/router/friend_test.go server/internal/service/moment_private_bucket_test.go init.sql
git commit -m "feat: expand homepage skin across social surfaces"
```

---

## Self-review
- Spec coverage: nav, profile, moments, friend requests, and chat background are all covered. Login/register/global inputs are intentionally excluded.
- Placeholder scan: no TODO/TBD placeholders remain.
- Type consistency: use `homepageSkin` everywhere in backend JSON, store state, and frontend rendering; shared resolver prevents drift between pages.

Plan complete and saved to `docs/superpowers/plans/2026-06-03-global-homepage-skin-expansion.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?