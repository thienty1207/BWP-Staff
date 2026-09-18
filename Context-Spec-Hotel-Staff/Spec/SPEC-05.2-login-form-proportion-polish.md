# SPEC-05.2 — Login Form Proportion & Header Polish

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Expected baseline: `main` at or after `55523ee6a6beb99873f45b38592189a05fb3d4e3`
>
> Status:
>
> - SPEC-05 ✅ CLOSED
> - SPEC-05.1 ✅ CLOSED
> - SPEC-05.2 = this focused UI follow-up
> - SPEC-06 remains unchanged and must **not** be implemented here

---

## 1. Goal

Refine the existing `/login` form so it looks **taller, narrower, cleaner, and more premium**, matching the user's latest visual reference.

This SPEC does **not** redesign the full login page.

Keep the existing:

- real background image;
- sharp/unblurred background;
- right-side desktop placement;
- dark-only login presentation;
- glass/blurred login panel;
- username/password auth;
- user icon;
- lock icon;
- show/hide password;
- Remember me;
- current backend auth flow.

The main change is the **form proportion and header composition**.

---

## 2. User-approved visual direction

Desktop intent:

```text
full-screen hotel background
                         tall narrow glass login panel
                         aligned on the far right
```

The panel should feel:

- narrower than SPEC-05.1;
- taller through vertical spacing;
- visually balanced;
- premium;
- not cramped;
- not excessively wide.

The user specifically wants the login form to look **thon hơn**:

```text
reduce width
increase vertical presence
```

---

## 3. Icon asset

Replace the decorative header icon from the visual reference with the user's real SVG asset:

```text
D:\Works\prject\BWP-SonaSea\img\icon\hotel-login-icon.svg
```

Copy/use it in frontend static assets, recommended:

```text
frontend/static/images/hotel-login-icon.svg
```

Runtime path:

```text
/images/hotel-login-icon.svg
```

Do not recreate it with CSS.

Do not substitute another icon.

Do not use a lotus icon.

Do not use the full BWP logo block on `/login`.

---

## 4. Scope

This SPEC is frontend-only unless a real blocker is discovered.

Likely files:

```text
frontend/src/routes/login/+page.svelte
frontend/src/lib/styles/app.css
frontend/static/images/hotel-login-icon.svg
frontend/tests/branding-assets.test.ts
```

Possibly add/update a focused login UI test if needed.

Do not change:

- backend auth;
- database;
- migrations;
- ticket list;
- SPEC-06;
- app-wide theme system;
- session behavior.

---

## 5. Background behavior stays unchanged

Keep the existing SPEC-05.1 background behavior:

```text
background image = sharp
background image = NOT blurred
form panel       = blurred/glass
```

Do not reintroduce background blur.

Keep the current real image:

```text
frontend/static/images/login-background.png
```

The background remains the visual showcase.

---

## 6. Desktop panel proportions

Make the login panel narrower than the current ~420px implementation.

Target desktop width:

```text
340px–360px
```

Recommended:

```text
350px
```

The panel should gain vertical presence through padding and spacing, not through an arbitrary fixed height.

Recommended visual target:

```text
width: ~350px
min-height: approximately 600–640px when content permits
padding-inline: 28–32px
padding-block: 34–40px
```

Important:

- avoid a rigid fixed height that can clip content;
- use natural content height plus generous vertical spacing;
- preserve responsive behavior.

---

## 7. Desktop placement

Keep the form aligned to the right side of the viewport.

Requirements:

- near the right edge;
- retain comfortable outer margin;
- do not place flush against browser edge;
- preserve most of the hotel background for viewing.

Do not center the form on wide desktop.

---

## 8. Header composition

The login form header must be **center aligned**.

Center all of:

- decorative icon;
- `BWP SONASEA STAFF` or current small kicker text if retained;
- decorative divider/line if used;
- `Sign in`;
- supporting copy.

Required hierarchy:

```text
      [hotel-login-icon.svg]

      BWP SONASEA STAFF

           Sign in

Access your BWP SonaSea account.
```

Exact spacing should be visually balanced.

Do not left-align the form heading.

---

## 9. Decorative header icon

Use:

```text
/images/hotel-login-icon.svg
```

Place it in the centered header composition.

The icon should be:

- small;
- decorative;
- visually aligned with the red accent system;
- not oversized;
- not treated as a full brand logo.

If the SVG already contains its own color, preserve it unless CSS masking is clearly required.

Do not modify the source SVG unnecessarily.

---

## 10. Form field alignment

Although the header is centered, the actual form controls should retain conventional alignment.

Keep field labels left-aligned:

```text
Username
[ input ]

Password
[ input ]

Remember me
```

This improves scanability and readability.

Required:

- Username label left;
- Password label left;
- Remember me left;
- input text left;
- submit button text centered.

---

## 11. Username field

Keep:

- user icon;
- username-only auth;
- accessible label;
- current autocomplete;
- current trim behavior.

Do not change authentication semantics.

---

## 12. Password field

Keep:

- lock icon;
- eye/eye-off toggle;
- accessible show/hide control;
- current-password autocomplete;
- current secure state handling.

Do not persist password.

---

## 13. Remember me

Keep the existing safe SPEC-05.1 behavior:

```text
localStorage key:
bwp-remembered-username
```

Remember username only.

Never store:

- raw password;
- session token;
- auth cookie value;
- JWT.

Do not change backend session TTL.

---

## 14. Submit button

Keep one primary:

```text
Sign in
```

Requirements:

- full panel content width;
- centered button label;
- comfortable height;
- premium red accent;
- loading state preserved;
- disabled while login request is in flight.

An arrow icon is optional only if it is lightweight and does not add a dependency.

Do not add a Forgot Password link.

Do not add SSO.

---

## 15. Vertical rhythm

The new look should feel taller through spacing.

Increase vertical spacing between:

- icon/kicker;
- heading;
- supporting copy;
- fields;
- Remember me;
- submit button.

Avoid excessive blank space.

The form must still fit common laptop heights without requiring awkward scrolling.

---

## 16. Footer decoration

Do **not** add new decorative footer copy such as:

```text
Great People Create Memorable Stays
```

unless that text already exists as an approved project requirement.

The visual reference is for proportion/layout, not permission to invent extra product copy.

Keep the form focused.

---

## 17. Responsive behavior

Desktop:

```text
>= 900px:
right-aligned
~350px panel
```

Tablet/mobile:

```text
< 900px:
center panel for usability
```

Small mobile:

```text
panel width = available viewport width minus safe horizontal margins
```

Do not force a 350px width if viewport is smaller.

Verify:

```text
360
390
430
768
1280
1920
```

No horizontal overflow.

---

## 18. Accessibility

Preserve:

- semantic labels;
- visible focus states;
- keyboard-operable password visibility toggle;
- contrast;
- loading state;
- error announcement;
- correct button type.

Decorative icon should use an empty alt or `aria-hidden` if appropriate.

---

## 19. Performance

This SPEC should not affect runtime performance materially.

Do not add:

- UI framework;
- icon library;
- animation library;
- large JS dependency.

Use the SVG as a static asset.

Keep CSS/Svelte simple.

---

## 20. No mock data

No mock/fake runtime data.

Do not add:

- fake auth user;
- fake login success;
- fake ticket;
- demo data.

Existing real auth flow remains mandatory.

---

## 21. Testing

Update focused tests so they assert the new approved UI contract.

Recommended checks:

- `/images/hotel-login-icon.svg` is referenced;
- old lotus/decorative placeholder is absent;
- login logo block remains absent;
- theme toggle remains absent;
- panel max width is approximately `350px` or equivalent narrow rule;
- header text alignment is centered;
- background remains unblurred;
- panel retains backdrop blur;
- Remember me remains username-only;
- no password/token storage introduced.

Do not add a large component testing framework.

---

## 22. Manual verification

Verify real `/login` with frontend/backend running.

Desktop:

- panel visibly narrower than SPEC-05.1;
- panel visually taller through spacing;
- right aligned;
- background remains clearly visible;
- header icon centered;
- header text centered;
- Sign in centered;
- supporting text centered;
- labels remain left aligned;
- inputs/buttons fit correctly;
- show/hide password works;
- Remember me works.

Responsive:

```text
360
390
430
768
1280
1920
```

Check:

- no clipping;
- no horizontal overflow;
- panel remains readable;
- no full BWP logo block;
- no theme selector.

---

## 23. Non-goals

Do not implement:

- SPEC-06;
- Create Ticket;
- ticket detail;
- chat;
- checklist;
- admin;
- backend auth changes;
- forgot password;
- new branding background;
- logo redesign;
- theme-system redesign.

---

## 24. Definition of Done

SPEC-05.2 is complete when:

- new SVG asset is copied to frontend static assets;
- new SVG is used in login header;
- login panel is materially narrower than SPEC-05.1;
- panel retains comfortable vertical presence;
- desktop panel remains right aligned;
- login header content is centered;
- `Sign in` is centered;
- supporting text is centered;
- Username/Password labels remain left aligned;
- user/lock/show-password controls still work;
- Remember me behavior is unchanged and safe;
- background remains sharp/unblurred;
- panel remains glass/blurred;
- no logo block;
- no login theme selector;
- responsive widths pass;
- no backend/database/migration changes;
- no mock data;
- frontend check/build/tests pass.

---

## 25. Suggested commit

```text
style: refine login form proportions for SPEC-05.2
```

---

## 26. Next phase

After SPEC-05.2 closes:

```text
SPEC-06 — Create Ticket / New Request + Targeted Hardening
```
