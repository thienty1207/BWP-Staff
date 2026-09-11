import { join } from 'node:path';
import { expect, test } from 'bun:test';

const frontendRoot = join(import.meta.dir, '..');

test('SPEC-04.1 public branding assets are available', async () => {
	const background = Bun.file(join(frontendRoot, 'static/images/login-background.jpg'));
	const logo = Bun.file(join(frontendRoot, 'static/images/bwp-logo.png'));

	expect(await background.exists()).toBe(true);
	expect(background.size).toBeGreaterThan(0);
	expect(await logo.exists()).toBe(true);
	expect(logo.size).toBeGreaterThan(0);
});

test('login page uses public asset paths instead of local filesystem paths', async () => {
	const loginPage = await Bun.file(join(frontendRoot, 'src/routes/login/+page.svelte')).text();
	const stylesheet = await Bun.file(join(frontendRoot, 'src/lib/styles/app.css')).text();

	expect(loginPage).toContain('/images/bwp-logo.png');
	expect(loginPage).toContain('BWP SonaSea Staff');
	expect(loginPage).not.toContain('Staff Portal');
	expect(loginPage).not.toContain('Use your admin-provisioned staff account to continue.');
	expect(stylesheet).toContain("url('/images/login-background.jpg')");
	expect(stylesheet).toContain('filter: blur(2px)');
	expect(stylesheet).toContain('white-space: nowrap');
	expect(loginPage).not.toContain('D:\\Works\\');
	expect(stylesheet).not.toContain('D:\\Works\\');
	expect(loginPage).not.toContain('file://');
	expect(stylesheet).not.toContain('file://');
});
