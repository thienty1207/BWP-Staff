# SPEC-04 — Login UI + Auth Integration

> **Project:** BWP SonaSea  
> **Phase:** First real frontend feature after SPEC-03 Authentication Backend  
> **Baseline:** `main` at or after `2eb58be666ccbad84c3e0037bbb26ec560948dd0`  
> **Authoritative context:** `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
>
> Read in order: `AGENTS.md` → `PROJECT_CONTEXT.md` → SPEC-01 → SPEC-02 → SPEC-03 → this SPEC → current code.
>
> This SPEC implements the **login UI and browser authentication integration only**. It does not implement Tickets, Report, Settings, chat, Staff Meal, Announcements, admin UI, or later product features.

---

## 1. Goal

Replace the untouched SvelteKit starter UI with the first real BWP SonaSea frontend flow:

```text
/login
  ↓
real username/password form
  ↓
POST /api/v1/auth/login
  ↓
HttpOnly session cookie
  ↓
authenticated /
  ↓
GET /api/v1/auth/me
  ↓
real persisted user identity
  ↓
POST /api/v1/auth/logout
```

The result must be responsive, minimal, dark/light themed, accessible, fast, and real-data-only.

---

## 2. Existing foundation

Frontend stack is locked:

```text
SvelteKit
Svelte 5
TypeScript
Bun
Vite
```

Backend auth contract already exists:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Authentication credential is the backend-owned HttpOnly `bwp_session` cookie.

Do not introduce JWT, Bearer auth, localStorage auth tokens, sessionStorage auth tokens, or JavaScript cookie parsing.

---

## 3. Current frontend baseline

At SPEC-04 start, the frontend is still essentially the SvelteKit starter:

```text
frontend/src/
├── app.d.ts
├── app.html
├── lib/
└── routes/
    ├── +layout.svelte
    └── +page.svelte
```

The visible starter copy such as `Welcome to SvelteKit` must be removed.

---

## 4. Same-origin `/api` contract

Frontend application code must call relative API paths:

```text
/api/v1/auth/login
/api/v1/auth/logout
/api/v1/auth/me
```

Do not scatter absolute backend URLs through Svelte components.

For local development, configure the existing Vite dev server to proxy:

```text
/api/* -> http://127.0.0.1:3000/api/*
```

The browser therefore talks to the SvelteKit origin while Vite forwards `/api` to the Go backend.

This keeps the host-only `SameSite=Lax` session cookie simple in development and preserves the intended production direction of one frontend-visible public site with `/api/*` routed to the backend.

Production reverse-proxy deployment itself is outside SPEC-04.

---

## 5. Environment rule

The permanent local environment rule remains:

```text
backend/.env
```

Do not create frontend env files for this SPEC:

```text
frontend/.env
frontend/.env.local
frontend/.env.development
.env.example
.env.test
```

The fixed `127.0.0.1:3000` target is local Vite dev-tool configuration only, not a production browser API address.

---

## 6. Required routes

Implement only:

```text
/login
/
```

Do not create fake future routes such as:

```text
/tickets
/report
/settings
/admin
```

`/` is a deliberately minimal authenticated entry screen until the Tickets SPEC.

---

## 7. Login page

`/login` must contain:

- BWP SonaSea text identity;
- optional `Staff Portal` subtitle;
- username field;
- password field;
- Sign in button;
- loading/submitting state;
- safe error area;
- dark/light theme control.

Do not add dead UI for:

- Sign up;
- Create account;
- Forgot password;
- Google login;
- OAuth;
- SSO;
- email login;
- Remember me.

Those capabilities do not exist.

---

## 8. Visual direction

### Dark

```text
black / charcoal
red accent
clean NestJS-like visual character
subtle depth/borders
minimal
```

### Light

```text
white / light neutral
blue accent
clean SARA-like operational visual character
minimal
```

Both themes must have identical behavior and layout.

Do not use stock hotel imagery or fabricate an official company logo. Text branding is sufficient unless an approved project asset already exists.

---

## 9. Responsive requirements

Verify at least:

```text
360px
390px
430px
768px
desktop
```

Requirements:

- no horizontal scroll;
- readable mobile padding;
- touch controls around 44px minimum height;
- focused desktop card width;
- visible keyboard focus;
- error messages must not break layout.

---

## 10. Accessibility

Use real form semantics:

- visible `<label>` for username;
- visible `<label>` for password;
- `autocomplete="username"`;
- `autocomplete="current-password"`;
- real submit button;
- Enter submits the form;
- loading/disabled state remains understandable;
- accessible error live/alert region;
- accessible theme toggle;
- visible focus states.

Placeholder-only labels are not acceptable.

---

## 11. Login input behavior

Username:

- trim surrounding whitespace;
- do not lowercase/uppercase automatically;
- do not treat email as a separate login identifier.

Password:

- do not trim;
- do not transform;
- never persist;
- never log.

Backend validation remains authoritative.

---

## 12. Login request

Send:

```http
POST /api/v1/auth/login
Content-Type: application/json
```

Body:

```json
{
  "username": "...",
  "password": "..."
}
```

Use:

```ts
credentials: 'include'
```

Do not add `Authorization`.

Prevent duplicate concurrent submissions.

---

## 13. Login result handling

### 200

- do not inspect or store the session token;
- clear transient login errors;
- navigate to `/`;
- `/` must independently call `/me`.

### 401 `invalid_credentials`

Show a generic message such as:

```text
Invalid username or password.
```

### 400

Show a concise invalid-input message.

### 500 / network failure

Show a generic retryable service/connectivity message.

Never display stack traces, DB errors, parser internals, or the password.

---

## 14. Already-authenticated `/login`

When `/login` mounts, call:

```text
GET /api/v1/auth/me
```

- `200` → replace/navigate to `/`
- `401` → show login form
- network/500 → do not pretend the user is logged out; show retryable service state

Avoid redirect loops.

---

## 15. Minimal authenticated `/`

`/` is **not** the Tickets page yet.

On mount:

```text
GET /api/v1/auth/me
```

Until auth is known, show a neutral loading state and do not flash protected identity.

### 200

Display only real `/me` identity, e.g.:

```text
BWP SonaSea
Signed in as <full_name>
<department name>
<role>
Logout
theme control
```

Do not show fake ticket counts, fake dashboard cards, mock announcements, or placeholder business data.

### 401

Replace/navigate to:

```text
/login
```

### 500 / network failure

Do **not** redirect to login.

Show:

```text
Unable to verify session
Retry
```

Backend outage is not equivalent to unauthenticated.

---

## 16. Reload behavior

After successful login, refreshing `/` must stay authenticated via:

```text
HttpOnly cookie + GET /me
```

Do not use localStorage as the auth source of truth.

The public user object may exist in memory while the page is alive, but a full reload must bootstrap from `/me`.

---

## 17. Logout

Authenticated root must provide a real Logout action.

Send:

```http
POST /api/v1/auth/logout
```

with:

```ts
credentials: 'include'
```

Prevent duplicate logout requests.

### 204

- clear only in-memory identity if used;
- replace/navigate to `/login`.

### 500 / network failure

- do not pretend logout succeeded;
- do not redirect;
- do not locally clear authenticated UI state;
- show retryable error;
- allow the user to press Logout again.

Do not attempt to delete `bwp_session` in JavaScript.

---

## 18. Back navigation after logout

After successful logout, Back/direct navigation to `/` must not reveal protected content.

`/` must bootstrap with `/me`; `401` sends the user to `/login`.

---

## 19. Frontend auth module

Keep auth networking out of page markup where practical.

Reasonable structure:

```text
frontend/src/lib/client/auth/
├── api.ts
└── model.ts
```

Focused functions:

```text
login(username, password)
getCurrentUser()
logout()
```

Do not build a generated SDK, generic interceptor framework, frontend repository/service architecture, or dependency injection for three endpoints.

---

## 20. Safe frontend user type

Mirror only the safe backend identity:

```ts
type AuthenticatedUser = {
  id: number;
  username: string;
  employee_code: string;
  full_name: string;
  role: string;
  department: {
    id: number;
    code: string;
    name: string;
  };
  avatar_url: string | null;
};
```

Do not define frontend fields for password hashes, session hashes, raw tokens, or revocation internals.

---

## 21. Backend error model

Backend safe errors use:

```json
{
  "error": {
    "code": "...",
    "message": "...",
    "request_id": "..."
  }
}
```

Parse this safely into a small typed representation if useful.

`request_id` may be retained for troubleshooting.

Do not expose unexpected implementation errors directly to users.

---

## 22. Theme foundation

Establish a small reusable CSS-variable theme foundation.

Conceptual variables:

```text
--bg
--surface
--surface-elevated
--text
--text-muted
--border
--accent
--accent-hover
--danger
--focus
```

A reasonable global stylesheet is:

```text
frontend/src/lib/styles/app.css
```

Do not build a giant design system.

---

## 23. Theme selection

Support:

```text
Light
Dark
```

Recommended rule:

1. saved explicit preference if present;
2. otherwise `prefers-color-scheme`;
3. manual toggle.

It is acceptable to persist **only** the visual theme preference to localStorage, e.g.:

```text
bwp-theme
```

Do not store session tokens, passwords, or auth source-of-truth data there.

Avoid an obvious theme flash if practical, but do not add a theme framework.

---

## 24. Svelte state

Follow the existing Svelte 5/runes-mode project.

Use straightforward local state.

Do not add a global state library.

A small shared in-memory auth helper is acceptable only if it materially simplifies the two real routes.

---

## 25. Client-side auth bootstrap

The HttpOnly cookie is browser-owned.

Resolve authentication by calling `/me`.

Do not inspect cookies in JavaScript.

Do not decode tokens.

Do not derive identity from browser storage.

---

## 26. SSR scope

Do not build a complex SSR cookie-forwarding/auth layer in SPEC-04.

Client-side `/me` bootstrap is sufficient for this phase.

Protected UI must remain neutral/loading until auth is known.

---

## 27. No mock data

Do not hard-code a runtime current user or fake successful login.

Manual verification must use the real Go backend and real PostgreSQL development data.

Static UI copy is not mock business data.

---

## 28. Backend scope

SPEC-03 is closed.

Do not change Go auth by default.

If a genuine frontend integration blocker is found:

1. identify it clearly;
2. make only the smallest SPEC-03-compatible fix if truly required;
3. add regression coverage;
4. report it.

Do not add new auth endpoints.

---

## 29. Database scope

SPEC-04 needs no schema change.

Do not modify:

```text
backend/migrations/
database schema
```

Do not create migration `0021`.

---

## 30. Vite proxy

Update the existing Vite configuration to proxy only `/api` to the local backend.

Conceptually:

```ts
server: {
  proxy: {
    '/api': {
      target: 'http://127.0.0.1:3000'
    }
  }
}
```

Preserve the existing SvelteKit plugin/adapter configuration.

---

## 31. Browser security

Never:

- read the HttpOnly session token;
- mirror it into another cookie;
- store auth secrets in localStorage/sessionStorage;
- log password/token/cookie values;
- put password/token in URL/query;
- add Bearer auth;
- change `SameSite=Lax` to `None` for convenience.

---

## 32. No auth polling

Do not poll `/me` continuously.

Call it only when needed, such as:

```text
/login bootstrap
/ bootstrap
full reload
explicit Retry
```

---

## 33. No fake remember-me

Do not add a Remember me checkbox.

Session lifetime is owned by the backend SPEC-03 TTL.

---

## 34. No role feature framework yet

Role may be displayed as identity information.

Do not build admin menus, permission routing, department authorization, or an RBAC framework in this SPEC.

---

## 35. Page titles

Use:

```text
/login -> BWP SonaSea — Login
/      -> BWP SonaSea
```

Remove visible Svelte starter branding.

---

## 36. Error UX

- Login errors near the form.
- Root session verification errors with Retry.
- Logout errors keep authenticated UI visible.
- Do not use browser `alert()` as the primary UX.

---

## 37. Loading UX

Implement clear lightweight states for:

```text
checking existing session
signing in
verifying authenticated root
logging out
```

No skeleton library is needed.

---

## 38. Form state

Do not persist password.

Do not automatically persist username/password into localStorage.

Standard browser password-manager autocomplete is allowed and expected.

---

## 39. Browser history

Recommended:

```text
unauthenticated /        -> replace /login
authenticated /login     -> replace /
successful login         -> /
successful logout        -> replace /login
```

Avoid auth loops.

---

## 40. Network error distinction

This is mandatory:

```text
401 != network/500
```

Only `401` means unauthenticated.

A backend outage during `/me` must show retry state, not redirect to login.

A logout failure must preserve authenticated UI and allow retry.

---

## 41. Response handling

Use TypeScript types and safe parsing.

At minimum:

- inspect HTTP status;
- parse expected JSON only where required;
- treat malformed/unexpected responses as generic service errors;
- do not crash the page.

Do not add a schema-validation framework only for these endpoints.

---

## 42. Expected frontend structure

A reasonable result:

```text
frontend/src/
├── lib/
│   ├── client/
│   │   └── auth/
│   │       ├── api.ts
│   │       └── model.ts
│   └── styles/
│       └── app.css
└── routes/
    ├── +layout.svelte
    ├── +page.svelte
    └── login/
        └── +page.svelte
```

Create only files with real behavior.

---

## 43. Verification commands

From `frontend/`:

```bash
bun install --frozen-lockfile
bun run check
bun run build
```

Do not claim lint/test commands that do not exist.

Do not add a large testing framework solely to manufacture a checklist.

If backend files are touched, run the appropriate full Go regression suite as well.

---

## 44. Required manual browser verification

Run the real system.

Backend:

```bash
go run ./cmd/server
```

Frontend:

```bash
bun run dev
```

Open:

```text
http://localhost:5173/login
```

Verify:

```text
wrong password
  -> generic error
  -> remains /login

correct credentials
  -> 200
  -> HttpOnly bwp_session exists
  -> navigate /

/
  -> /me 200
  -> real identity visible

refresh /
  -> still authenticated

Logout
  -> 204
  -> /login

Back/direct /
  -> /me 401
  -> /login
```

Do not paste raw cookie values into docs/reports.

---

## 45. Theme verification

Check both Dark and Light on:

```text
/login
/
```

If theme persistence is implemented, refresh and verify it remains.

Verify localStorage contains only the visual preference and no auth secret.

---

## 46. Responsive verification

Inspect:

```text
360
390
430
768
desktop
```

Confirm no overflow, usable inputs/buttons, stable errors, visible focus, and usable theme control.

---

## 47. Documentation

After implementation, update only actual-state documentation:

```text
PROJECT_CONTEXT.md
README.md if frontend/run instructions are stale
SPEC-04 file under Context-Spec-BWP-SonaSea/Spec/
```

Do not rewrite unrelated product requirements.

---

## 48. Strict non-goals

Do not implement:

```text
Tickets UI
Open/Closed tabs
Pending/Accepted
New Request
ticket chat
checklist
Report
Settings
Profile password change
Staff Meal
Announcements
Admin UI
notifications
WebSocket
uploads
mobile app
PWA
service worker
OAuth
SSO
JWT
RBAC
signup
forgot password
user CRUD
```

---

## 49. Diff review

Before finishing, inspect the diff for:

- mock users;
- fake auth success;
- auth tokens in localStorage/sessionStorage;
- hard-coded real credentials;
- frontend env files;
- absolute backend URLs scattered in app code;
- SameSite/CORS broadening;
- migration/schema changes;
- Tickets/Report/Admin scope creep;
- debug password/token logging;
- unnecessary UI frameworks;
- dead Forgot Password/Signup links.

Remove anything outside SPEC-04.

---

## 50. Definition of Done

SPEC-04 is complete only when:

- `/login` exists;
- starter Svelte welcome is gone;
- username/password real login works;
- `credentials: include` is used;
- no JS auth token storage exists;
- wrong credentials show generic error;
- double submit is prevented;
- successful login navigates `/`;
- authenticated `/login` redirects `/`;
- `/` verifies with `/me`;
- `/` redirects only on real `401`;
- backend/network failure shows Retry instead of fake logout;
- `/` displays only real `/me` identity;
- no fake dashboard/ticket data exists;
- reload persists auth via cookie + `/me`;
- logout calls real backend;
- logout success navigates `/login`;
- logout failure preserves authenticated UI and supports retry;
- Back after logout cannot reveal protected content;
- dark theme works;
- light theme works;
- theme preference does not store auth data;
- mobile/desktop layouts work;
- accessibility requirements are met;
- Vite `/api` proxy works;
- no frontend `.env` file exists;
- no migration/schema change exists;
- SPEC-03 is not weakened;
- `bun install --frozen-lockfile` passes;
- `bun run check` passes;
- `bun run build` passes;
- real browser login/reload/logout flow passes;
- dark/light and responsive manual checks pass;
- docs reflect actual state;
- no future-feature scope creep remains.

---

## 51. Next phase

After SPEC-04, the next phase can begin the actual authenticated staff interface.

Likely:

```text
SPEC-05 — Tickets Read/List Foundation + Main Authenticated Shell
```

Do not implement SPEC-05 in this phase.
