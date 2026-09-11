# SPEC-04.1 — Login Branding, Blurred Background & Visual Regression Polish

> **Project:** BWP SonaSea  
> **Phase:** Small visual follow-up to completed SPEC-04  
> **Baseline:** `main` at or after `e87b67785f54199921284fa3b5af78c33d9f0243`  
> **Authoritative context:** `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
>
> SPEC-04 is already complete. SPEC-04.1 only improves the existing authentication visuals and performs a focused browser regression check before SPEC-05.
>
> Read in order:
>
> 1. `AGENTS.md`
> 2. `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
> 3. `Context-Spec-BWP-SonaSea/Spec/SPEC-03-authentication-backend.md`
> 4. `Context-Spec-BWP-SonaSea/Spec/SPEC-04-login-ui-auth-integration.md`
> 5. this SPEC
> 6. current frontend implementation

---

# 1. Goal

Improve the completed authentication UI by:

1. replacing the temporary text/lettermark branding with the supplied BWP logo image;
2. adding the supplied login background image;
3. rendering that background with a visible blur effect;
4. preserving the existing dark/light theme behavior;
5. preserving all SPEC-04 authentication behavior exactly;
6. verifying that no uncaught application exception appears during login/logout navigation.

Target visual structure:

```text
full-screen background image
        ↓
blurred background layer
        ↓
theme-aware overlay
        ↓
sharp login card + real BWP logo
```

This is a visual polish task, not a new authentication feature.

---

# 2. Source Assets

Use exactly these local files supplied by the user:

```text
D:\Works\prject\BWP-SonaSea\img\login-background\background-1.jpg
D:\Works\prject\BWP-SonaSea\img\Logo\BWP-logo.png
```

Before modifying code:

- verify both files exist;
- verify they are readable image files;
- do not silently substitute another image;
- do not download a replacement;
- do not generate a replacement;
- do not fabricate another BWP logo.

If either source file is missing:

```text
STOP
report the exact missing path
do not continue with a substitute
```

---

# 3. Asset Placement

Do not hard-code local Windows filesystem paths into runtime frontend code.

Copy the supplied assets to:

```text
frontend/static/images/login-background.jpg
frontend/static/images/bwp-logo.png
```

Runtime frontend references must be:

```text
/images/login-background.jpg
/images/bwp-logo.png
```

The original source files under `img/` must remain untouched.

Do not commit duplicate/temp copies.

---

# 4. Scope

Expected implementation scope:

```text
frontend/static/images/login-background.jpg
frontend/static/images/bwp-logo.png
frontend/src/routes/login/+page.svelte
frontend/src/routes/+page.svelte
frontend/src/lib/styles/app.css
frontend/src/routes/+layout.svelte                     # only if genuinely needed
Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md
Context-Spec-BWP-SonaSea/Spec/SPEC-04.1-login-branding-background.md
frontend/README.md                                     # only if current docs become stale
```

Exact file count may be smaller.

Do not touch unrelated files.

---

# 5. Strict Non-Goals

Do NOT modify or implement:

```text
backend authentication
login API contract
logout API contract
/me API contract
cookies
session logic
database
migrations
CORS
Vite API proxy behavior
JWT
Bearer Authorization
localStorage auth
sessionStorage auth
Tickets
Report
Settings
Profile password change
Admin UI
chat
Staff Meal
Announcements
WebSocket
uploads
SPEC-05 features
```

Do not redesign completed auth behavior.

---

# 6. Login Background

The `/login` page must use:

```text
/images/login-background.jpg
```

Required behavior:

```text
full viewport
background cover
centered appropriately
responsive
no tiling
no distortion
no horizontal overflow
```

The background must remain visually recognizable, but it must **not** appear as a sharp photographic backdrop.

It must be deliberately softened with blur.

---

# 7. Blur Implementation

Blur only the background layer.

Do NOT blur:

```text
login card
logo
text
username input
password input
buttons
theme toggle
error messages
```

Preferred structure:

```text
login shell
├── blurred background layer / pseudo-element
├── theme-aware overlay
└── sharp foreground content
```

A reasonable visual starting range:

```text
blur: 8px–18px
scale: 1.03–1.08
```

The scale should prevent visible empty/transparent edges caused by the blur.

The exact final blur amount should be visually tuned.

Do not use JavaScript, canvas, or a third-party blur library.

---

# 8. Theme-Aware Overlay

The blurred image must be combined with an overlay so foreground contrast remains strong.

## Dark theme

Use a translucent:

```text
black / charcoal
optional extremely subtle red tint
```

Requirements:

- background image remains visible;
- form remains easy to read;
- red accent remains coherent;
- overlay must not reduce the page back to a flat black screen.

## Light theme

Use a translucent:

```text
white / soft neutral
optional subtle blue tint
```

Requirements:

- image remains visible but subdued;
- form boundaries remain readable;
- blue accent remains coherent;
- page must not become washed out.

---

# 9. Login Card Treatment

Keep the existing login card structure and functionality.

Allowed visual tuning:

```text
surface opacity
border
shadow
subtle glass effect
contrast
```

Do not:

```text
rebuild the page into a different layout
add a marketing hero
add stock copy
add animation that delays login
```

The card itself must remain sharp.

---

# 10. Real BWP Logo

Replace the temporary authentication brand treatment:

```text
[BWP letter mark] + BWP SonaSea text
```

with:

```text
/images/bwp-logo.png
```

Use a real `<img>` element.

Recommended:

```html
<img src="/images/bwp-logo.png" alt="BWP SonaSea" />
```

Use:

```text
constrained height
width: auto
object-fit: contain
```

Do not stretch or distort the logo.

If the image already contains the visible brand name, do not duplicate the same visible text immediately beside it.

---

# 11. Logo Placement

Use the real BWP logo anywhere the completed SPEC-04 auth UI still shows the temporary auth brand treatment.

At minimum:

```text
/login topbar/header
/ authenticated root topbar/header
```

Do not build future Tickets/Admin navigation.

---

# 12. Responsive Logo Requirements

Verify at:

```text
360px
390px
430px
768px
1280px / desktop
```

Requirements:

- no clipping;
- no squashing;
- no excessive width;
- no horizontal overflow;
- no collision with theme toggle;
- image remains legible.

The logo may be slightly smaller on narrow screens.

---

# 13. Dark/Light Compatibility

Both themes use:

```text
same BWP logo image
same login background image
same authentication behavior
```

Only CSS variables/overlay treatment differ.

Do not create unnecessary duplicate dark/light image files.

---

# 14. Performance

Do not add:

```text
image CDN
canvas blur
runtime image processing
third-party blur package
animation package
```

CSS is sufficient.

If the source background image is unusually large, reasonable optimization is allowed only if:

- visual quality remains good;
- original source remains untouched;
- no heavy runtime processing is introduced.

Do not visibly degrade the BWP logo.

---

# 15. Accessibility

Logo must have useful alt text.

The login background is decorative.

Do not expose the background as meaningful screen-reader content.

Existing SPEC-04 form accessibility must not regress:

```text
labels
autocomplete
keyboard submit
focus visibility
error live region
theme toggle accessibility
```

---

# 16. Authentication Behavior Must Remain Unchanged

The following completed SPEC-04 behavior must continue working:

```text
/login checks /me
authenticated /login -> /
wrong password -> generic safe error
correct login -> /
/ checks /me
reload / stays authenticated
logout 204 -> /login
logout failure stays authenticated and retryable
401 != backend/network failure
theme preference uses bwp-theme only
```

No auth logic rewrite is required.

---

# 17. Expected `/me 401` After Logout

After successful logout, `/login` performs its normal session bootstrap:

```text
GET /api/v1/auth/me
```

The backend may return:

```text
401 Unauthorized
```

because the session was correctly revoked.

This is **expected behavior**.

Browser DevTools may display this request in red.

Do NOT change correct auth behavior solely to remove this expected HTTP 401 from DevTools.

Correct frontend behavior is:

```text
/me 401
↓
map to unauthenticated
↓
show login form
```

The expected 401 is not itself an application bug.

---

# 18. Uncaught Browser Exception Regression Check

During manual browser verification, there must be no uncaught exception owned by BWP SonaSea application code.

A previously observed example was:

```text
Uncaught TypeError:
Cannot read properties of undefined (reading 'startTime')
```

with a stack similar to:

```text
VM...
<anonymous>
```

rather than a normal application source file.

If such an error appears:

1. inspect the stack/source;
2. verify whether it originates from:
   - BWP SonaSea frontend source; or
   - Web Preview / DevTools / browser extension / injected script;
3. verify again in a normal browser session without preview/tool injection where practical;
4. do not modify application code unless evidence shows the application owns the exception.

Search application source for the referenced symbol when useful.

Do not make speculative code changes to suppress an injected-tool error.

---

# 19. Theme Behavior

Keep the existing theme preference contract:

```text
localStorage key: bwp-theme
```

Do not add a second theme mechanism.

Do not store:

```text
auth token
session token
password
current authenticated user source-of-truth
image selection
```

in localStorage.

---

# 20. No Runtime Filesystem Paths

After implementation, runtime frontend source must not contain:

```text
D:\Works\prject\...
file://...
```

Those paths are source-copy instructions only.

Runtime references must be web paths under:

```text
/images/...
```

---

# 21. Manual Visual Verification

Run the real app.

Verify `/login` in both themes.

## Dark

- background image visible;
- background visibly blurred;
- dark overlay preserves readability;
- login card remains sharp;
- real BWP logo is visible;
- logo is not distorted;
- red accent remains coherent.

## Light

- same background image;
- background remains visibly blurred;
- light overlay preserves readability;
- login card remains sharp;
- real BWP logo is visible;
- logo is not distorted;
- blue accent remains coherent.

---

# 22. Responsive Verification

Inspect:

```text
360
390
430
768
1280 / desktop
```

Verify:

- background covers full viewport;
- no visible blur-edge gap;
- logo does not collide with theme toggle;
- form remains centered and usable;
- no horizontal scroll;
- card remains readable.

---

# 23. Asset Verification

Expected committed assets:

```text
frontend/static/images/login-background.jpg
frontend/static/images/bwp-logo.png
```

Do not accidentally commit:

```text
duplicate image copies
temporary exports
.psd
.ai
Windows shortcut files
unrelated image folders
```

---

# 24. Frontend Verification Commands

From:

```text
frontend/
```

run:

```bash
bun install --frozen-lockfile
bun run check
bun run build
bun test tests/auth-api.test.ts
```

Do not claim PASS unless actually executed.

---

# 25. Manual Auth Regression

Re-run the real browser auth flow:

```text
wrong password
-> generic login error

correct credentials
-> login success
-> /

refresh /
-> remains authenticated

logout
-> /login

/login bootstrap
-> /me may return expected 401
-> login form remains usable
```

Verify there is no uncaught application exception.

Do not expose raw password or cookie values.

---

# 26. Diff Review

Before commit, inspect the final diff for:

- backend changes;
- migration/schema changes;
- auth logic changes;
- hard-coded Windows asset paths;
- unblurred background;
- accidentally blurred login card;
- duplicated visible branding;
- broken light theme;
- broken dark theme;
- new frontend env files;
- speculative workaround for expected 401;
- speculative workaround for injected `startTime` error;
- SPEC-05 scope creep;
- duplicate/temp image files.

Remove anything outside SPEC-04.1.

---

# 27. Definition of Done

SPEC-04.1 is complete only when:

- background source file is copied into frontend static assets;
- BWP logo source file is copied into frontend static assets;
- `/login` uses the new background;
- background is visibly blurred;
- only the background is blurred;
- no blur-edge gaps are visible;
- dark overlay remains readable;
- light overlay remains readable;
- login card remains sharp;
- temporary brand treatment is replaced by the real logo image;
- real logo appears on `/login`;
- real logo appears on authenticated `/`;
- logo is responsive and undistorted;
- no local Windows runtime path exists;
- no backend/migration/database change exists;
- SPEC-04 auth behavior still works;
- expected `/me 401` after logout remains correctly handled;
- no uncaught application-owned browser exception remains;
- theme toggle still works;
- mobile widths still work;
- `bun run check` passes;
- `bun run build` passes;
- existing auth tests pass;
- manual dark/light verification is complete;
- manual auth regression is complete;
- no SPEC-05 scope creep exists.

---

# 28. Next Phase

After SPEC-04.1 is closed:

```text
SPEC-05
```

may begin the authenticated application shell and Tickets read/list work.

Do not implement SPEC-05 in this task.
