# SPEC-03 — Authentication Backend

> **Project:** BWP SonaSea  
> **Phase:** Authentication backend after SPEC-01 Database Foundation and SPEC-02 Backend Foundation  
> **Baseline:** `main` at or after `2d7cb5670b8b2b1b3c0679ae1505db08693736c9`
> **Authoritative context:** `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
>
> Every coding agent must read, in this order:
>
> 1. `AGENTS.md`
> 2. `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
> 3. `Context-Spec-BWP-SonaSea/Spec/SPEC-01-database-foundation.md`
> 4. `Context-Spec-BWP-SonaSea/Spec/SPEC-02-backend-foundation.md`
> 5. this SPEC
> 6. current repository implementation
>
> This SPEC implements **backend authentication only**.
>
> It does not implement the login UI, ticket APIs, chat, admin CRUD, password
> change UI, or any other later feature.

---

# 1. Goal

Implement real username/password authentication for BWP SonaSea using the
existing PostgreSQL schema and the HTTP foundation completed in SPEC-02.

Required endpoints:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Authentication model:

```text
username + password
        ↓
Argon2id verification
        ↓
cryptographically random session token
        ↓
raw token returned only as HttpOnly cookie
        ↓
SHA-256 hash stored in PostgreSQL auth_sessions
        ↓
authenticated requests resolve cookie -> session -> active user
```

The result must be:

```text
simple
explicit
secure
database-backed
easy to trace
compatible with later SvelteKit integration
```

No JWT architecture is required.

---

# 2. Locked Authentication Direction

The project authentication direction is already decided.

Use:

```text
username/password login
server-side PostgreSQL sessions
HttpOnly cookie
admin-provisioned accounts
```

Do not use:

```text
email login
public signup
forgot-password email flow
magic links
OAuth
SSO
JWT access tokens
JWT refresh tokens
Firebase Auth
browser localStorage tokens
browser sessionStorage tokens
```

Email remains optional profile/contact data only.

The login identifier is:

```text
username
```

only.

---

# 3. Existing Foundation

SPEC-01 already provides:

```text
users
auth_sessions
departments
user_preferences
```

Relevant user fields already include:

```text
id
username
employee_code
password_hash
full_name
department_id
role
avatar_url
is_active
last_login_at
```

Relevant session fields already include:

```text
id
session_token_hash
user_id
expires_at
last_used_at
user_agent
ip_address
created_at
revoked_at
```

SPEC-01 also provides Argon2id password hashing support.

SPEC-02 already provides:

```text
AppState
pgxpool
/api/v1 foundation
central safe JSON errors
request ID
request logging
panic recovery
credentialed CORS
/health
/ready
startup lifecycle
graceful shutdown
```

SPEC-03 must build on those foundations.

Do not replace them.

---

# 4. No Database Migration

The current schema already supports this authentication implementation.

SPEC-03 must not add or rewrite migrations unless a real blocking schema defect
is discovered.

Do not modify:

```text
0001..0017
0019
0020
```

Migration version `0018` remains permanently retired.

Do not create `0021` merely for authentication code.

If a genuine blocking schema problem is found:

```text
stop
report the exact blocker
do not silently change the schema
```

---

# 5. Required Backend Structure

Create the first real client feature package only now, because authentication
behavior is being implemented.

Preferred shape:

```text
backend/
├── app/
├── client/
│   └── auth/
│       ├── handler.go
│       ├── service.go
│       ├── repository.go
│       ├── model.go
│       └── ...tests
├── config/
├── shared/
│   ├── database.go
│   ├── security/
│   │   ├── password.go
│   │   ├── password_test.go
│   │   ├── session.go
│   │   └── session_test.go
│   └── ...
└── cmd/
```

This is guidance, not a requirement to create exactly five auth files.

Keep files responsibility-based.

Do not create empty future packages such as:

```text
tickets
chat
checklist
notifications
reports
settings
```

during SPEC-03.

---

# 6. Layering

Use the project default:

```text
HTTP
 ↓
Handler
 ↓
Service
 ↓
Repository
 ↓
pgx/pgxpool
 ↓
PostgreSQL
```

## Handler

Owns:

```text
JSON parsing
HTTP input validation
cookie read/write/delete
request metadata extraction
HTTP response
```

## Service

Owns:

```text
credential verification
disabled-user decision
session token creation
session expiry calculation
authentication workflow
logout workflow
```

## Repository

Owns:

```text
user lookup SQL
session insert SQL
session lookup SQL
session revoke SQL
last_login_at update
transaction boundaries involving database writes
```

Do not put password-verification logic directly in SQL.

Do not put raw SQL in Fiber handlers.

---

# 7. Central HTTP Error Reuse

SPEC-03 must continue using the safe central error response behavior from
SPEC-02.

The first feature package must not duplicate JSON error writing in every
handler.

If the current `app.AppError` location creates an import cycle when the auth
package is wired into the app, perform the **smallest concrete refactor** that
makes the central typed HTTP error reusable.

A reasonable solution is a small package such as:

```text
backend/shared/httperror/
```

containing only the reusable safe HTTP error type.

Rules:

- This refactor is allowed because there are now two real consumers:
  - app foundation
  - auth feature
- Do not create a generic enterprise error framework.
- Do not introduce error-code registries, reflection, factories, or interfaces.
- Preserve the SPEC-02 response shape.

External error response remains:

```json
{
  "error": {
    "code": "invalid_credentials",
    "message": "Invalid username or password",
    "request_id": "..."
  }
}
```

Internal database/security errors must still be logged server-side and returned
as safe generic `500`.

---

# 8. Authentication Routes

Register exactly:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Do not add:

```text
/register
/signup
/forgot-password
/reset-password
/refresh
/token
/oauth
/sso
/users
/change-password
```

in this SPEC.

---

# 9. Login Request

Endpoint:

```http
POST /api/v1/auth/login
Content-Type: application/json
```

Request:

```json
{
  "username": "hothienty",
  "password": "..."
}
```

Rules:

```text
username required
password required
```

For username:

- trim surrounding whitespace;
- reject empty username;
- maximum length must respect the database contract (`VARCHAR(50)`);
- do not silently lowercase or uppercase it;
- do not accept email as an alternate identifier.

For password:

- preserve the password exactly as entered;
- do not trim password whitespace;
- reject empty password;
- apply a reasonable upper input bound to prevent absurd request sizes;
- do not enforce account password-complexity rules during login.

A maximum login password input size around `1024` bytes is sufficient.

Do not log the request body.

Do not log the password.

---

# 10. Invalid JSON / Invalid Input

Malformed JSON or structurally invalid login input must return a safe client
error.

Preferred response:

```http
400 Bad Request
```

Example:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Invalid request",
    "request_id": "..."
  }
}
```

Do not include parser internals in the response.

Do not echo password input.

---

# 11. Login User Lookup

Lookup by exact persisted username.

Conceptually:

```sql
SELECT
    id,
    username,
    employee_code,
    password_hash,
    full_name,
    department_id,
    role,
    avatar_url,
    is_active
FROM users
WHERE username = $1
LIMIT 1;
```

The exact query may join the department if that avoids an unnecessary second
round trip.

Use explicit columns.

Do not use `SELECT *`.

Do not query by email.

---

# 12. Password Verification

Use the existing:

```go
security.VerifyPassword(...)
```

Argon2id implementation.

Do not add bcrypt.

Do not add plaintext fallback.

Do not compare password hashes directly as strings.

Do not re-hash the supplied password and compare hash strings because Argon2id
uses salts.

Flow:

```text
lookup username
     ↓
obtain stored Argon2id PHC hash
     ↓
VerifyPassword(raw password, stored hash)
```

---

# 13. Unknown User Timing Discipline

Do not immediately return before performing an Argon2id verification when a
well-formed username does not exist.

Use a fixed valid **dummy Argon2id PHC hash** that is not secret and exists only
to make the unknown-user password path perform comparable expensive password
verification work.

Rules:

- the dummy value is not an account password;
- it contains no credential;
- it may be a package constant;
- use the same existing Argon2id verifier;
- do not generate a new expensive dummy hash per failed login request.

This is intended to reduce obvious username-enumeration timing differences.

Do not claim perfect constant-time behavior across the entire network/database
stack.

---

# 14. Invalid Credentials Response

For these cases:

```text
username not found
wrong password
inactive/disabled user
```

return the same external authentication failure:

```http
401 Unauthorized
```

```json
{
  "error": {
    "code": "invalid_credentials",
    "message": "Invalid username or password",
    "request_id": "..."
  }
}
```

Do not reveal:

```text
"username does not exist"
"password wrong"
"account disabled"
```

to the unauthenticated client.

Server-side logs also must not contain the raw password.

---

# 15. Disabled Users

A user with:

```text
is_active = FALSE
```

must not successfully log in.

If an already-authenticated user's account becomes inactive, that user's
existing session must stop authenticating on the next authenticated request.

This means session validation must include the current persisted user
`is_active` state.

Do not rely only on the user state captured when the session was created.

---

# 16. Session Token Generation

On successful login generate a cryptographically random opaque session token.

Required minimum entropy:

```text
256 bits
```

Recommended implementation:

```text
32 random bytes from crypto/rand
        ↓
base64.RawURLEncoding
```

Expected raw token is opaque.

It is not:

```text
a JWT
a user ID
a timestamp
a reversible serialized struct
```

Do not use:

```text
math/rand
UUID v1
predictable counters
```

for authentication session secrets.

---

# 17. Session Token Storage

The browser receives the **raw token only in the HttpOnly cookie**.

PostgreSQL stores only a one-way hash.

Recommended session token hash:

```text
SHA-256(raw session token)
```

Store a stable text encoding such as lowercase hex.

Flow:

```text
raw token
   ├──> HttpOnly cookie
   └──> SHA-256
          ↓
      auth_sessions.session_token_hash
```

Never store the raw session token in:

```text
PostgreSQL
logs
JSON response
audit logs
request ID
URL/query string
```

No extra `SESSION_SECRET` is required for this opaque random-token design.

Do not invent one.

---

# 18. Session Security Helper

Add small explicit helpers under the existing security package where useful.

Conceptually:

```go
GenerateSessionToken() (rawToken string, tokenHash string, err error)
HashSessionToken(rawToken string) string
```

Exact signatures may vary.

Tests must verify:

- generated raw token is non-empty;
- hash is non-empty;
- raw token and stored hash differ;
- hashing the same raw token is deterministic;
- independently generated raw tokens differ;
- helper uses cryptographically secure randomness;
- raw token is URL/cookie-safe.

Do not expose the raw token through errors.

---

# 19. Session Creation

After successful password verification and active-user validation:

```text
generate raw session token
hash token
calculate expires_at
insert auth_sessions row
update users.last_login_at
commit
set cookie
return authenticated user
```

The session insert and `last_login_at` update should be one short database
transaction so login write state is internally consistent.

If the transaction fails:

```text
do not set a valid auth cookie
return safe 500
```

Do not retry login writes indefinitely.

---

# 20. Multiple Sessions

Allow multiple active sessions for one user.

Examples:

```text
desktop browser
phone browser
another authorized workstation
```

A new login must not automatically revoke all other sessions.

Do not implement:

```text
single-session-only policy
logout-all-devices
session management UI
```

in SPEC-03.

---

# 21. Session Lifetime

Add one legitimate runtime configuration value:

```text
AUTH_SESSION_TTL_HOURS
```

Default:

```text
12
```

Validation:

```text
minimum: 1 hour
maximum: 720 hours (30 days)
```

This is a real runtime auth setting, not a test-only variable.

For local development, if the variable is explicitly needed in configuration,
use the existing real:

```text
backend/.env
```

Do not create:

```text
.env.example
.env.test
AUTH_SESSION_TEST_TTL
```

Config unit tests may temporarily override the official
`AUTH_SESSION_TTL_HOURS` with `t.Setenv`.

Session expiry is fixed from login:

```text
created_at
    ↓
expires_at = login time + configured TTL
```

Do not implement rolling/sliding session renewal in SPEC-03.

---

# 22. Auth Cookie Contract

Use one cookie name:

```text
bwp_session
```

Cookie properties:

```text
HttpOnly = true
Path = /
SameSite = Lax
Domain = not explicitly set (host-only)
```

Secure behavior:

```text
APP_ENV=development -> Secure=false
non-development     -> Secure=true
```

Do not add another env variable only for cookie Secure mode.

Cookie expiry/max-age should match the server-side session expiry as closely as
practical.

The raw token must not be returned in JSON.

---

# 23. SameSite Deployment Rule

Current auth design intentionally uses:

```text
SameSite=Lax
```

Production frontend and backend should therefore be deployed as the same
site, for example:

```text
https://staff.example.com
https://api.example.com
```

or behind one public site.

Do not silently change to:

```text
SameSite=None
```

just to support an unrelated cross-site deployment.

If a genuinely cross-site frontend/backend deployment is required later, that
must be an explicit security decision together with CSRF protections.

Do not weaken cookie policy in SPEC-03.

---

# 24. Login Response

Successful login:

```http
200 OK
Set-Cookie: bwp_session=...
Content-Type: application/json
```

Response:

```json
{
  "user": {
    "id": 1,
    "username": "hothienty",
    "employee_code": "IT-ADMIN-001",
    "full_name": "Ho Thien Ty",
    "role": "admin",
    "department": {
      "id": 1,
      "code": "IT",
      "name": "IT Department"
    },
    "avatar_url": null
  }
}
```

Do not return:

```text
password_hash
session_token_hash
raw session token
is_active internals that the UI does not need
database row internals
```

The exact JSON field naming must remain stable and use normal snake_case JSON
where shown.

---

# 25. User Identity Model

Define one explicit authenticated-user/identity model reusable by:

```text
login response
GET /auth/me
authenticated request context
future authorization checks
```

It should contain only useful identity information.

Recommended fields:

```text
user id
username
employee code
full name
role
department id/code/name
avatar URL
```

The authenticated request context may additionally hold internal session
information such as:

```text
session id
session expiry
```

without exposing it in the public user JSON.

Do not store `password_hash` in request context.

---

# 26. Authenticated Session Lookup

For an incoming raw cookie token:

```text
read cookie
 ↓
validate basic token size/shape
 ↓
SHA-256 hash
 ↓
query auth_sessions by session_token_hash
 ↓
join current users
 ↓
verify:
    revoked_at IS NULL
    expires_at > NOW()
    user is_active = TRUE
 ↓
authenticated principal
```

Use one efficient SQL query where practical.

Required session validity conditions:

```text
matching hash
not revoked
not expired
active user
```

Do not load every session for the user.

Do not scan by user ID.

Use the indexed token hash.

---

# 27. Missing / Invalid Session

These cases are unauthenticated:

```text
missing cookie
malformed cookie token
unknown token hash
expired session
revoked session
inactive user
```

Protected authentication endpoints should return:

```http
401 Unauthorized
```

Example:

```json
{
  "error": {
    "code": "unauthenticated",
    "message": "Authentication required",
    "request_id": "..."
  }
}
```

Do not distinguish externally between:

```text
expired
revoked
unknown
disabled
```

unless a later product requirement explicitly needs it.

---

# 28. Authentication Middleware

Implement one small authentication middleware/helper that can be reused by
later backend specs.

Conceptually:

```text
RequireAuth
```

Responsibilities:

```text
read bwp_session cookie
resolve valid persisted session
resolve active persisted user
store authenticated principal in Fiber request-local context
continue request
```

Provide one small typed helper to read the current principal from the request
context.

Do not implement role/permission middleware yet.

Do not create:

```text
RequireAdmin
RequireDepartment
RBAC engine
permission matrix
```

until an endpoint actually needs those rules.

---

# 29. Fiber Request Context

The authenticated principal is request-scoped only.

Do not retain Fiber request-local state after the request ends.

Do not store the authenticated user in global mutable memory.

Do not use an in-memory session map.

PostgreSQL remains the session source of truth.

---

# 30. GET /api/v1/auth/me

Route:

```http
GET /api/v1/auth/me
```

Must use the authentication middleware.

Successful response:

```http
200 OK
```

```json
{
  "user": {
    "id": 1,
    "username": "hothienty",
    "employee_code": "IT-ADMIN-001",
    "full_name": "Ho Thien Ty",
    "role": "admin",
    "department": {
      "id": 1,
      "code": "IT",
      "name": "IT Department"
    },
    "avatar_url": null
  }
}
```

`/me` must read the current persisted session/user state.

It must not trust user information supplied by the browser.

Missing/invalid session:

```http
401
```

---

# 31. POST /api/v1/auth/logout

Route:

```http
POST /api/v1/auth/logout
```

Logout should be idempotent.

Behavior:

```text
read bwp_session cookie if present
       ↓
hash raw token if structurally acceptable
       ↓
revoke matching active session if it exists
       ↓
expire/delete browser cookie
       ↓
204 No Content
```

If the cookie is:

```text
missing
invalid
already revoked
expired
unknown
```

still clear the browser cookie and return:

```http
204 No Content
```

Do not expose whether the token existed.

---

# 32. Session Revocation

Logout must persist revocation:

```text
revoked_at = NOW()
```

Do not physically delete the auth session row on normal logout.

Reason:

```text
explicit revocation state
future auditing/debugging support
historical consistency
```

After logout, the same raw cookie token must no longer authenticate.

---

# 33. Cookie Deletion

Logout must overwrite/expire the same cookie contract.

Use:

```text
Name = bwp_session
Path = /
HttpOnly = true
SameSite = Lax
Secure based on APP_ENV
expired time / negative MaxAge as appropriate
```

The deletion cookie must match the original path/name/security behavior closely
enough for the browser to remove it.

Do not return the old token.

---

# 34. last_login_at

On successful login:

```text
users.last_login_at = current login time
```

Do not update `last_login_at` on failed login.

Do not use `last_login_at` for session validity.

---

# 35. last_used_at

`auth_sessions.last_used_at` already exists in the database.

SPEC-03 does not require a write on every authenticated API request.

Avoid turning every read request into a PostgreSQL write solely to maintain
`last_used_at`.

It may remain NULL or be initialized during session creation.

Do not implement a complicated touch/throttling subsystem in SPEC-03.

A later requirement may define low-frequency session activity updates if they
become operationally useful.

---

# 36. User-Agent and IP Metadata

When creating a session, store useful request metadata in the existing columns:

```text
user_agent
ip_address
```

Rules:

- extract User-Agent in the handler;
- apply a reasonable upper bound before persistence;
- use the direct client IP information currently available to Fiber;
- do not blindly trust arbitrary forwarded IP headers;
- do not add reverse-proxy trust configuration in SPEC-03.

If IP parsing fails, storing NULL is preferable to failing a valid login.

Do not log cookies to capture this metadata.

---

# 37. Request ID

All auth routes continue using the SPEC-02 server-generated request ID.

Do not trust incoming client `X-Request-ID`.

Authentication errors must contain the same generated request ID as:

```text
X-Request-ID response header
request log
central error body
```

Do not change the request-ID middleware behavior completed in SPEC-02.

---

# 38. Logging

Allowed auth logging may include:

```text
request_id
route
status
internal database error where safe
user ID after successful authenticated resolution if operationally useful
```

Never log:

```text
password
password_hash
raw session token
session cookie
session_token_hash
full Cookie header
Authorization header
login JSON request body
```

Avoid logging unknown usernames on failed login unless a concrete operational
need is later established.

Do not leak credentials through error wrapping that is sent to clients.

---

# 39. No Fake Auth Data

Authentication must use real PostgreSQL state.

Do not implement:

```go
if username == "hothienty" && password == "123" {
    return true
}
```

Do not create in-memory users.

Do not hard-code the real development admin password.

The development admin remains provisioned through the existing explicit
development seed workflow.

---

# 40. Development Admin

The existing development admin may be used for local manual verification.

Username:

```text
hothienty
```

Its password must continue to come only from the real local development
configuration.

Do not copy the password into:

```text
Go source
SPEC files
README
tests
Git history
```

---

# 41. Permanent Environment Rule

The project's environment rule remains unchanged.

Local development uses exactly:

```text
backend/.env
```

Never create:

```text
.env.example
.env.test
.env.testing
.env.local
DATABASE_TEST_URL
TEST_DATABASE_URL
TEST_DB_URL
APP_ENV=test
```

only for testing.

Database-backed tests use:

```text
DATABASE_URL
```

and temporary isolated PostgreSQL schemas.

The new:

```text
AUTH_SESSION_TTL_HOURS
```

is permitted because it is a real runtime authentication setting used by the
application.

Do not add test-only auth env variables.

---

# 42. PostgreSQL Test Isolation

Authentication integration tests must use the real local PostgreSQL instance
configured by:

```text
DATABASE_URL
```

and an isolated temporary schema.

Pattern:

```text
DATABASE_URL
   ↓
connect to development PostgreSQL
   ↓
CREATE SCHEMA spec03_<unique>
   ↓
search_path = spec03_<unique>,public
   ↓
run migrations
   ↓
insert explicit test department/user/session rows
   ↓
run authentication tests
   ↓
DROP only spec03_<unique>
```

Never:

```text
drop the database
truncate public
delete normal development users
reset the development schema
```

Do not create `DATABASE_TEST_URL`.

---

# 43. Auth Test Passwords

Automated tests may use fixed **test-only passwords inside test source** for
isolated temporary users.

That is test fixture data, not production configuration.

Examples may be clearly named:

```text
test-only-correct-password
test-only-wrong-password
```

Rules:

- never use the real development admin password;
- create the isolated test user's Argon2id hash during test setup;
- test data must live only in the temporary test schema.

This does not violate the project's no-mock production-data rule.

---

# 44. Login Integration Tests

Required integration coverage:

1. correct username/password returns `200`;
2. login sets `bwp_session`;
3. login cookie is HttpOnly;
4. login cookie Path is `/`;
5. login cookie SameSite is Lax;
6. development cookie Secure behavior is correct;
7. non-development cookie Secure behavior is correct at the cookie helper level;
8. successful login inserts a real `auth_sessions` row;
9. database stores token hash, not raw cookie token;
10. session expiry reflects configured TTL;
11. successful login updates `users.last_login_at`;
12. wrong password returns `401`;
13. unknown username returns the same public `401` contract;
14. inactive user returns the same public `401` contract;
15. failed login creates no session;
16. failed login does not update `last_login_at`;
17. email value is not treated as an alternate login identifier;
18. login response contains only safe public user fields;
19. request ID in auth errors matches the server response request ID.

Do not assert the random session token's exact value.

---

# 45. Multiple Session Test

Verify:

```text
same active user
login once -> session A
login again -> session B
```

Both sessions should remain independently valid until:

```text
their expiry
their individual revocation
user deactivation
```

Do not accidentally revoke session A when session B is created.

---

# 46. /me Tests

Required:

```text
valid session -> 200 + correct current user
missing cookie -> 401
unknown token -> 401
expired session -> 401
revoked session -> 401
inactive user -> 401
```

Also verify:

- database password hash is never returned;
- session token/hash is never returned;
- user identity comes from PostgreSQL, not from browser input.

---

# 47. Logout Tests

Required:

```text
valid session logout -> 204
matching auth_sessions row gets revoked_at
cookie is expired/deleted
same token after logout -> /me returns 401

missing cookie logout -> 204
unknown cookie logout -> 204
already revoked cookie logout -> 204
```

Logout must be idempotent.

---

# 48. Token Security Tests

Verify:

```text
raw token is not equal to token hash
raw token does not appear in auth_sessions.session_token_hash
two generated tokens differ
same raw token hashes to same value
different raw tokens hash differently
```

Do not print raw tokens in test failure output unless absolutely necessary.

Prefer failure messages that do not echo secret material.

---

# 49. Generic Credential Failure Tests

Ensure external response does not reveal whether:

```text
username exists
password is wrong
user is inactive
```

All three should expose the same public:

```text
status
error code
message
```

Do not test exact execution-time equality.

Do test that the unknown-user code path still performs the dummy Argon2id
verification behavior at the service/unit boundary where feasible without
fragile timing assertions.

---

# 50. Central Error Regression

Existing SPEC-02 guarantees must continue to work.

Auth implementation must not regress:

```text
server-generated request ID
panic logging
safe 500 errors
CORS
/health
/ready
graceful shutdown
```

Do not duplicate middleware inside the auth feature.

---

# 51. CORS and Cookies

Do not change current SPEC-02 CORS policy unless a concrete auth integration
test proves a required cookie behavior cannot work.

Current direction remains:

```text
single configured FRONTEND_ORIGIN
credentials enabled
no wildcard origin
X-Request-ID exposed
```

Do not add `Authorization` just because authentication is being implemented.

This auth design uses the session cookie, not Bearer tokens.

---

# 52. CSRF Scope

SPEC-03 uses:

```text
HttpOnly
Secure in non-development
SameSite=Lax
explicit credentialed CORS origin
```

Do not invent a half-designed cross-site CSRF token subsystem in this SPEC.

Do not change SameSite to None.

If later deployment or API requirements require genuinely cross-site
credentialed requests, CSRF design must be revisited explicitly before that
deployment model is allowed.

---

# 53. Rate Limiting Scope

Do not add an ad-hoc in-memory login rate limiter merely to say rate limiting
exists.

A per-process in-memory limiter would be incomplete for future multi-instance
deployment and would introduce shared mutable state.

Login abuse/rate limiting is a legitimate production-hardening concern, but it
is outside this SPEC unless the project later adopts a concrete persistent or
edge-level strategy.

Do not add Redis in SPEC-03.

Do not claim the login endpoint is brute-force-proof.

---

# 54. Session Cleanup Scope

Do not implement a background cleanup worker for expired/revoked sessions in
SPEC-03.

Validity queries already reject:

```text
expired sessions
revoked sessions
```

Historical cleanup can be added later if table growth justifies it.

Do not delete expired sessions during every `/me` request.

---

# 55. Password Rehash Scope

Do not build password-change or password-rehash migration logic in SPEC-03.

Existing Argon2id hashes are valid.

If password parameters change in a later security-hardening phase, a
`NeedsRehash` policy may be added then.

Do not silently mutate password hashes during ordinary `/me`.

---

# 56. No User CRUD

SPEC-03 authenticates existing users.

It does not implement:

```text
create user API
edit user API
disable user API
delete user API
reset password API
admin user listing
```

The admin-provisioned account direction remains.

Those operations belong to later admin functionality.

---

# 57. No Frontend Work

Do not change SvelteKit in SPEC-03.

No:

```text
login page
form
browser fetch helper
route guard
frontend auth store
sidebar user display
```

Those belong to the next integration/UI SPEC.

SPEC-03 must be fully testable at the backend HTTP/API level.

---

# 58. Route Registration

Keep route wiring explicit.

A reasonable shape is:

```go
api := server.Group("/api/v1")
auth.RegisterRoutes(api, state.DB, settings)
```

or an equally simple equivalent.

Do not create a generic route registrar framework.

Do not use reflection.

Do not scan packages dynamically.

Ensure no import cycle is introduced.

---

# 59. AppState

Keep:

```go
type AppState struct {
    DB *pgxpool.Pool
}
```

unless a concrete implementation requirement genuinely needs another shared
long-lived dependency.

Do not place:

```text
current user
session map
password hasher state
cookie token
```

in `AppState`.

Configuration values can be passed explicitly during route construction.

---

# 60. SQL Discipline

Use parameterized explicit SQL.

No ORM.

No query builder unless already required elsewhere.

Use:

```text
explicit column lists
indexed session-token lookup
short login transaction
one session validation query where practical
```

Avoid:

```text
N+1 department query if a join is simpler
SELECT *
loading all user sessions
unbounded scans
```

---

# 61. Session Lookup Performance

The hot authenticated request path should conceptually be:

```text
cookie
 ↓
SHA-256 in Go
 ↓
indexed equality lookup on session_token_hash
 ↓
join one user + department
```

Do not query sessions by:

```text
raw token
user ID first
LIKE
sequential token scans
```

Do not add Redis before measurement.

---

# 62. Transactions

Use a short transaction where atomicity is real.

Successful login write transaction should cover:

```text
INSERT auth_sessions
UPDATE users.last_login_at
```

Do not keep the transaction open while running Argon2id password verification.

Correct order:

```text
DB user lookup
 ↓
Argon2id verify outside transaction
 ↓
generate session token
 ↓
short DB write transaction
```

Do not hold a DB connection/transaction during expensive password hashing.

---

# 63. Context and Timeouts

Use standard `context.Context`.

Database calls must remain bounded by request cancellation and existing backend
timeout discipline.

Do not use `context.Background()` for normal request database work when the
request context is available.

Do not create a global HTTP timeout middleware.

Do not retain Fiber request state beyond handler lifetime.

---

# 64. Security Failure Behavior

For any unexpected internal authentication error:

```text
log internally with request ID
return safe 500
```

Do not expose:

```text
SQL text
PostgreSQL host
password hash
token hash
cookie value
stack trace
Argon2 internals
```

Known auth failures remain safe 4xx responses.

---

# 65. Cookie Secret Handling

The raw session cookie is a bearer credential.

Treat it as a secret.

Never include it in:

```text
fmt.Printf
log.Printf
error strings
panic messages
JSON
URL
database plaintext
test snapshots
Baron proof text
```

If a test needs to inspect the cookie, keep it only in test memory and avoid
printing its exact value on failure.

---

# 66. Documentation Updates

After implementation, update documentation only to reflect actual state.

Expected updates:

```text
PROJECT_CONTEXT.md current repository snapshot
README.md only if backend run/auth behavior documentation is now stale
```

Do not rewrite unrelated product requirements.

Do not change deferred Report/Admin UI decisions.

The SPEC itself should live at:

```text
Context-Spec-BWP-SonaSea/Spec/SPEC-03-authentication-backend.md
```

---

# 67. Recommended Implementation Order

Use this order unless repository inspection reveals a concrete dependency:

```text
1. Read context + SPEC-01 + SPEC-02 + SPEC-03
2. Inspect current app/config/security/schema/tests
3. Add AUTH_SESSION_TTL_HOURS config
4. Add session token generation/hash helpers
5. Add auth models
6. Add auth repository
7. Add auth service
8. Add login handler
9. Add authenticated session lookup middleware
10. Add /me
11. Add logout/revocation
12. Wire /api/v1/auth routes
13. Resolve central AppError import boundary minimally if required
14. Add focused unit tests
15. Add real PostgreSQL integration tests with temporary schema
16. Run all SPEC-01/SPEC-02 regression tests
17. Review security-sensitive diff
18. Update current repository snapshot/docs
```

Do not start frontend login work.

---

# 68. Verification Commands

Run from:

```text
backend/
```

Required:

```bash
gofmt -d .
go mod tidy
go vet ./...
go test ./... -count=1
go build ./...
```

Where supported:

```bash
go test -race ./... -count=1
```

Run real PostgreSQL integration tests using:

```text
DATABASE_URL from backend/.env
temporary isolated schema
```

Do not report a command as passed unless it actually ran.

---

# 69. Manual API Verification

Use the real local development backend/database.

Before verification, explicitly seed development data if required using the
existing seed command.

Do not make normal server startup auto-seed.

Start:

```bash
go run ./cmd/server
```

## Login

Send:

```http
POST /api/v1/auth/login
Content-Type: application/json
```

with the real local development admin credentials supplied through local
configuration.

Verify:

```text
200
safe user JSON
Set-Cookie bwp_session
no raw token in JSON
```

Do not paste the real password into committed scripts/docs.

## Me

Send the returned cookie:

```http
GET /api/v1/auth/me
```

Verify:

```text
200
same persisted user identity
```

## Logout

Send:

```http
POST /api/v1/auth/logout
```

with the same cookie.

Verify:

```text
204
cookie expired
session revoked
```

Then:

```http
GET /api/v1/auth/me
```

with the old token must return:

```text
401
```

---

# 70. Regression Verification

SPEC-03 must not break:

```text
GET /health
GET /ready
404 central error JSON
server-generated X-Request-ID
panic recovery/logging
configured CORS
graceful shutdown
database migration runner
development seed isolation
SPEC-01 PostgreSQL constraints/indexes
```

Do not weaken existing tests to force the new code to pass.

---

# 71. No Schema / Future Feature Scope Creep

Before finalizing, inspect the diff.

There should be no implementation of:

```text
ticket endpoints
Accept/Assign/Close
chat
WebSocket
checklists
notifications
Staff Meal
announcements
Report
Settings
Admin CRUD
frontend login UI
Redis
queues
object storage
```

There should be no migration/schema diff.

---

# 72. Definition of Done

SPEC-03 is complete only when all are true:

- Login exists at `POST /api/v1/auth/login`.
- Logout exists at `POST /api/v1/auth/logout`.
- Me exists at `GET /api/v1/auth/me`.
- Username is the only login identifier.
- Password verification uses existing Argon2id.
- Unknown/wrong/inactive credentials expose the same safe `401`.
- Unknown-user flow performs dummy Argon2id verification.
- Successful login requires `is_active = TRUE`.
- Raw session token uses cryptographically secure randomness.
- Raw session token has at least 256 bits of entropy.
- Browser receives raw token only through `bwp_session`.
- PostgreSQL stores only SHA-256 token hash.
- Raw token never appears in JSON/logs/database plaintext.
- Session row has a bounded expiry.
- `AUTH_SESSION_TTL_HOURS` is validated.
- No test-only auth environment variable exists.
- Cookie is HttpOnly.
- Cookie is SameSite=Lax.
- Cookie Path is `/`.
- Cookie is Secure outside development.
- Cookie is host-only.
- Multiple sessions per user are allowed.
- Login updates `last_login_at`.
- Session lookup rejects expired sessions.
- Session lookup rejects revoked sessions.
- Session lookup rejects inactive users.
- `/me` uses persisted session/user state.
- `/me` returns safe identity only.
- Logout persists `revoked_at`.
- Logout clears the browser cookie.
- Logout is idempotent.
- Revoked token cannot authenticate afterward.
- Authentication middleware is reusable by future feature specs.
- No role/RBAC framework was added.
- No JWT was added.
- No session map/in-memory user store was added.
- No database migration was added.
- No login UI/frontend work was added.
- PostgreSQL tests use `DATABASE_URL`.
- PostgreSQL tests isolate with temporary schemas.
- `.env.example` and `.env.test` were not created.
- Existing SPEC-01 tests still pass.
- Existing SPEC-02 tests still pass.
- `gofmt -d .` passes.
- `go mod tidy` passes.
- `go vet ./...` passes.
- `go test ./... -count=1` passes.
- `go build ./...` passes.
- race tests pass where supported.
- manual login -> me -> logout -> me failure flow is verified.
- documentation reflects the implemented repository state.
- final diff contains no future-feature scope creep.

---

# 73. Handoff to Next SPEC

After SPEC-03 is complete, backend authentication is ready for the next phase:

```text
SPEC-04 — Login UI + Auth Integration
```

That later SPEC may implement:

```text
SvelteKit login page
dark/light login UI
credentials: include
frontend /me bootstrap
authenticated route behavior
logout UI
frontend auth state
```

Do not implement those items in SPEC-03.

---

# 74. Final Implementation Rule

The desired authentication system is:

```text
POST /api/v1/auth/login
        ↓
username lookup
        ↓
Argon2id verification
        ↓
active user check
        ↓
32 random bytes
        ↓
raw token ───────────────> HttpOnly bwp_session cookie
        ↓
SHA-256
        ↓
auth_sessions
        ↓
fixed expiry


Authenticated request
        ↓
bwp_session
        ↓
SHA-256
        ↓
indexed auth_sessions lookup
        ↓
not revoked
not expired
active user
        ↓
authenticated principal


POST /api/v1/auth/logout
        ↓
hash cookie token
        ↓
revoked_at = NOW()
        ↓
expire cookie
        ↓
204
```

Keep it boring.

Keep it explicit.

Keep the raw token secret.

Use PostgreSQL as the source of truth.

Do not add JWT.

Do not create fake auth data.

Do not create test-only env files or test-only database variables.

Do not build future features early.
