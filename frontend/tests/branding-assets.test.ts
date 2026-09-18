import { join } from 'node:path';
import { expect, test } from 'bun:test';

const frontendRoot = join(import.meta.dir, '..');

test('Hotel Staff login assets are available', async () => {
	const background = Bun.file(join(frontendRoot, 'static/images/login-background.jpg'));
	const loginIcon = Bun.file(join(frontendRoot, 'static/images/hotel-login-icon.svg'));

	expect(await background.exists()).toBe(true);
	expect(background.size).toBeGreaterThan(0);
	expect(await loginIcon.exists()).toBe(true);
	expect(loginIcon.size).toBeGreaterThan(0);
});

test('login page uses public asset paths instead of local filesystem paths', async () => {
	const loginPage = await Bun.file(join(frontendRoot, 'src/routes/login/+page.svelte')).text();
	const shellPage = await Bun.file(join(frontendRoot, 'src/routes/+page.svelte')).text();
	const stylesheet = await Bun.file(join(frontendRoot, 'src/lib/styles/app.css')).text();

	expect(loginPage).toContain('Sign in');
	expect(loginPage).toContain('Hotel Staff — Login');
	expect(loginPage).toContain('Access your Hotel Staff account');
	expect(loginPage).toContain('/images/hotel-login-icon.svg');
	expect(loginPage).toContain('Remember me');
	expect(loginPage).toContain('Show password');
	expect(loginPage).toContain('Hide password');
	expect(loginPage).toContain('hotel-staff-remembered-username');
	expect(loginPage).toContain('autocomplete="username"');
	expect(loginPage).toContain('autocomplete="current-password"');
	expect(loginPage).not.toMatch(/BWP|SonaSea|Best Western/);
	expect(shellPage).toContain('Hotel Staff');
	expect(shellPage).not.toMatch(/BWP|SonaSea|Best Western|\/images\/bwp-logo\.png/);
	expect(loginPage.toLowerCase()).not.toContain('lotus');
	expect(loginPage).not.toContain('ThemeToggle');
	expect(loginPage).not.toContain('Staff Portal');
	expect(loginPage).not.toContain('Use your admin-provisioned staff account to continue.');
	expect(stylesheet).toContain("url('/images/login-background.jpg')");
	expect(stylesheet).toContain('backdrop-filter: blur(18px)');
	expect(stylesheet).not.toContain('filter: blur(2px)');
	expect(stylesheet).toContain('justify-content: flex-end');
	expect(stylesheet).toContain('width: min(100%, 350px);');
	expect(stylesheet).toMatch(/\.login-heading\s*\{[\s\S]*text-align:\s*center;/);
	expect(stylesheet).toMatch(/\.login-field label\s*\{[\s\S]*text-align:\s*left;/);
	expect(loginPage).not.toContain('sessionStorage');
	expect(loginPage).not.toContain('document.cookie');
	expect(loginPage).not.toContain('localStorage.setItem(rememberedUsernameKey, password)');
	expect(loginPage).not.toContain('D:\\Works\\');
	expect(stylesheet).not.toContain('D:\\Works\\');
	expect(loginPage).not.toContain('file://');
	expect(stylesheet).not.toContain('file://');
});

test('runtime public assets exclude retired BWP files and the Svelte starter favicon', async () => {
	const bwpLogo = Bun.file(join(frontendRoot, 'static/images/bwp-logo.png'));
	const retiredLoginBackground = Bun.file(join(frontendRoot, 'static/images/login-background.png'));
	const starterFavicon = Bun.file(join(frontendRoot, 'src/lib/assets/favicon.svg'));
	const layout = await Bun.file(join(frontendRoot, 'src/routes/+layout.svelte')).text();

	expect(await bwpLogo.exists()).toBe(false);
	expect(await retiredLoginBackground.exists()).toBe(false);
	expect(await starterFavicon.exists()).toBe(false);
	expect(layout).toContain('<link rel="icon" href="/images/hotel-login-icon.svg" />');
	expect(layout).not.toContain('$lib/assets/favicon.svg');
	expect(layout).not.toContain('favicon.svg');
});

test('frontend crawler and environment-file policies protect the internal runtime surface', async () => {
	const robots = await Bun.file(join(frontendRoot, 'static/robots.txt')).text();
	const ignoreRules = await Bun.file(join(frontendRoot, '.gitignore')).text();

	expect(robots.trim()).toBe('User-agent: *\nDisallow: /');
	expect(ignoreRules).toMatch(/^\.env$/m);
	expect(ignoreRules).toMatch(/^\.env\.\*$/m);
	expect(ignoreRules).not.toContain('!.env.example');
	expect(ignoreRules).not.toContain('!.env.test');
});
