# SPEC-05.1 — Login UI Visual Refresh

> Project: BWP SonaSea
> Baseline: `main` at or after `99b19407e50d36813156a6c94653ac1fe8105fc1`
> This is a **small follow-up UI SPEC** after SPEC-05.
> SPEC-06 remains unchanged and is **not** implemented here.

---

## 1. Read order

Read in this order:

1. `AGENTS.md`
2. `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
3. `SPEC-01`
4. `SPEC-02`
5. `SPEC-03`
6. `SPEC-04`
7. `SPEC-04.1`
8. `SPEC-05`
9. this SPEC
10. current repository implementation

Current explicit user instruction overrides older assumptions when there is no real conflict with locked project rules.

---

## 2. Goal

Refresh the `/login` page so it matches the newer visual direction provided by the user.

This SPEC changes only the **login page UI/UX presentation**.

Implement:

- full-screen login background image
- **no blur on the background image**
- login form as a floating blurred/glass panel
- login form visually placed on the **right side** on desktop
- username icon
- password/lock icon
- show/hide password control
- remember-me checkbox
- dark-only login presentation
- removal of login-page logo and login-page theme toggle

Do **not** redesign the authenticated Tickets shell.
Do **not** start SPEC-06 here.

---

## 3. Design source provided by user

The user supplied two visual references:

1. a login form style reference
2. a background image reference

The intended result is:

```text
login page = full-screen background image + right-side login form panel
```

The background is meant to be clearly visible and admired.
Therefore, unlike SPEC-04.1, this page should **not blur the background itself**.

---

## 4. Assets

Use the real local asset the user specified:

```text
D:\Works\prject\BWP-SonaSea\img\login-background\background-1.png
```

Image properties provided by user:

```text
1920 x 1080
```

Copy/use it in frontend static assets, for example:

```text
frontend/static/images/login-background.png
```

Do not replace it with another image.
Do not use placeholder art.
Do not use CSS-only gradient as the main background.

Important:

- the login page must not show the old standalone BWP logo block;
- the login page must not show the Light/Dark theme selection button;
- only the background image and the login form should dominate the page.

---

## 5. Scope boundary

This SPEC is **frontend-only** unless a concrete blocker is discovered.

Allowed scope:

- `frontend/src/routes/login/+page.svelte`
- shared login-page styles
- tiny login UI helpers if needed
- lightweight frontend tests related to the login page
- SPEC documentation update

Not in scope:

- backend auth contract changes
- session model changes
- database changes
- migrations
- tickets shell redesign
- SPEC-06 feature work

---

## 6. Auth behavior must stay the same

Do not change the existing backend auth contract from SPEC-03.

Keep:

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- HttpOnly cookie auth
- username-only login

Do not introduce:

- email login
- JWT
- OAuth
- localStorage auth
- sessionStorage auth
- reading cookies from JS

The current login flow must continue to work exactly as before.

---

## 7. Login page structure

The `/login` page should contain only two main visual layers:

1. the full-screen background image
2. the login form panel

Do **not** show on `/login`:

- top logo block
- theme toggle button
- extra authenticated-shell navigation
- fake marketing cards
- footer clutter
- dashboard content

The page should feel cinematic and focused.

---

## 8. Background behavior

Background requirements:

- full viewport coverage
- `background-size: cover`
- `background-position: center`
- no blur applied to the background image
- no previous pseudo-element background blur technique
- a subtle dark overlay is allowed for readability
- background must remain visibly sharp enough to appreciate the hotel image

Allowed:

- soft vignette or dark overlay for readability

Not allowed:

- strong blur on the page background
- heavy opaque overlay that hides the image

---

## 9. Login form panel behavior

The login form panel becomes the blurred element.

Required visual behavior:

- dark glass / translucent panel
- `backdrop-filter: blur(...)`
- rounded corners
- subtle border
- subtle shadow
- readable contrast over the image

The form itself should feel premium and modern.

Important distinction:

```text
Background = not blurred
Form panel = blurred / glassmorphism
```

---

## 10. Desktop layout

On desktop widths, place the login form at the **right side** of the screen so most of the background remains visible.

Intent:

```text
left side  -> background showcase
right side -> login form
```

Guidance:

- form should be visually aligned near the right edge
- keep reasonable outer spacing/padding
- do not stick the form flush against the browser edge
- maintain comfortable reading width

A reasonable desktop form width is approximately:

```text
360px – 440px
```

Do not center the form on large desktop unless responsive width makes that unavoidable.

---

## 11. Tablet and mobile layout

Responsive behavior is still required.

On narrower screens:

- prioritize usability and readability
- the form may become centered or nearly centered if needed
- background remains full-screen
- form width should adapt without horizontal overflow

Do not force an unusable far-right layout on narrow mobile screens.

Verify at least:

```text
360
390
430
768
1280
```

No horizontal overflow.

---

## 12. Login form content

The form should include:

- a heading, e.g. `Sign in`
- a short supporting line similar to the user reference, e.g. access your BWP SonaSea account
- username input
- password input
- remember-me checkbox
- submit button

Do not include:

- separate BWP logo block above the form
- theme selector
- forgot-password flow
- public signup link
- SSO button
- email field

Keep the form concise.

---

## 13. Username input

Requirements:

- label or accessible labeling
- placeholder appropriate for username
- leading user icon
- current validation behavior preserved
- trim username the same way current app already does
- use appropriate `autocomplete`, e.g. `username`

Do not change login identifier semantics.

Login remains:

```text
username + password
```

---

## 14. Password input

Requirements:

- label or accessible labeling
- leading lock icon
- show/hide password toggle button
- accessible toggle with clear `aria-label`
- keep current password submission behavior unchanged
- preserve secure handling of password in frontend state
- use `autocomplete="current-password"`

The user explicitly wants show-password behavior.

A trailing eye / eye-off icon control is appropriate.

---

## 15. Remember-me behavior

The user wants a remember-me feature.

Security constraints still apply.

### Required rule

Do **not** store the raw password in:

- localStorage
- sessionStorage
- plain cookies
- custom browser storage
- source code

### Safe interpretation for this SPEC

Implement a **Remember me** checkbox that safely supports convenience without storing plaintext password.

Allowed behavior for SPEC-05.1:

- remember the username only in local storage when the user opts in; or
- remember only the checkbox state and rely on browser autofill for password; or
- another equally safe approach that does **not** store the raw password.

Minimum requirement:

- visible Remember me control exists
- behavior is functional in a safe way
- no plaintext password persistence is introduced

If remembering username is implemented:

- when checked and login succeeds, store the username only
- when unchecked, clear stored remembered username
- preload remembered username on future visits

Do not change backend session TTL because of this checkbox in SPEC-05.1.
That belongs to a later explicit auth/session feature if ever needed.

---

## 16. Theme behavior on `/login`

For this SPEC, `/login` is **dark by default and dark-only in presentation**.

Rules:

- do not show login-page theme toggle
- do not show Light/Dark switch on `/login`
- login page should render using the dark visual system

Important:

This does **not** require deleting the existing app-wide theme system.

Authenticated pages may continue to use the existing theme behavior.
The route-specific login page simply should not expose theme switching UI.

---

## 17. Authenticated routes must remain unchanged

Do not redesign `/` or the Tickets shell in this SPEC.

Keep existing authenticated behavior from SPEC-05.

Changes must be limited to the login experience unless a small shared style extraction is genuinely needed.

---

## 18. Accessibility

Maintain or improve accessibility:

- semantic form controls
- keyboard-accessible show/hide password button
- visible focus styles
- labels or accessible names
- button disabled/loading state during submit
- no contrast regression

Do not sacrifice usability for aesthetics.

---

## 19. No mock data rule still applies

Even though this is a UI refresh, the permanent project rule still applies.

Do not add:

- fake auth users
- fake login success
- fake tickets
- sample dashboard cards
- demo data for the login page

Keep the real backend-authenticated flow.

---

## 20. Dependencies

Do not add a large UI or icon library for two or three icons unless there is a very strong reason.

Prefer:

- inline SVG
- tiny existing assets
- simple current stack approach

This is a small UI refinement.

---

## 21. Testing expectations

At minimum, preserve and update relevant frontend tests.

Suggested coverage:

- login page renders without logo block/theme toggle
- background asset path exists
- remember-me safe behavior if implemented
- password visibility toggle works or at least its state logic is covered
- existing auth API tests remain green

Do not add a massive testing framework just for this.

---

## 22. Manual verification

Manually verify with real backend and frontend running:

1. open `/login`
2. background image is visible and not blurred
3. no login-page logo block
4. no Light/Dark theme button on `/login`
5. form appears on right side on desktop
6. form itself has blurred/glass look
7. username field shows user icon
8. password field shows lock icon
9. show/hide password works
10. remember-me control is visible and functional in the safe defined way
11. login still succeeds with real backend
12. invalid password still shows existing safe error behavior
13. authenticated `/` still works after login
14. logout still works
15. no horizontal overflow on mobile widths

---

## 23. Files likely to change

Likely frontend files:

```text
frontend/src/routes/login/+page.svelte
frontend/src/lib/styles/app.css
frontend/tests/...
frontend/static/images/login-background.png
```

Possibly documentation:

```text
Context-Spec-BWP-SonaSea/Spec/SPEC-05.1-login-ui-visual-refresh.md
Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md
frontend/README.md
```

Only update docs if they become stale.

---

## 24. Verification commands

From `frontend/` run:

```bash
bun install --frozen-lockfile
bun run check
bun run build
bun test
```

Only report commands actually executed.

---

## 25. Non-goals

This SPEC does **not** implement:

- SPEC-06 ticket creation
- new request modal
- departments API
- locations API
- ticket actions
- chat
- checklist
- report
- settings redesign
- admin UI
- forgot-password feature
- email login
- new backend remember-session semantics

---

## 26. Definition of Done

This SPEC is complete only when all of the following are true:

- `/login` uses the real provided background image
- background is not blurred
- login form panel is blurred/glass-styled
- login page no longer shows the old logo block
- login page no longer shows the Light/Dark theme button
- desktop login form is right-aligned in a premium layout
- responsive layout works at 360/390/430/768/1280
- username input includes icon
- password input includes icon
- show/hide password works
- remember-me control exists and works in a safe manner
- raw password is not persisted in browser storage
- auth flow from SPEC-03/04 remains working
- authenticated shell is not unintentionally redesigned
- frontend checks/build/tests pass
- no mock data was introduced
- no backend/database/migration changes were introduced without a true blocker

---

## 27. Next phase

After this SPEC, continue with:

```text
SPEC-06 — Create Ticket / New Request
```

SPEC-06 remains unchanged by this document.
