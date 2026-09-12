import { expect, test } from 'bun:test';

test('New Request remains a focused component and the Tickets page exposes its action', async () => {
	const dialogPath = new URL('../src/lib/components/NewRequestDialog.svelte', import.meta.url);
	const pagePath = new URL('../src/routes/+page.svelte', import.meta.url);
	const dialog = await Bun.file(dialogPath).text();
	const page = await Bun.file(pagePath).text();

	expect(dialog).toContain('Create request');
	expect(dialog).toContain('datetime-local');
	expect(dialog).toContain('getDepartments');
	expect(dialog).toContain('getLocations');
	expect(page).toContain('New Request');
	expect(page).not.toContain('Refresh');
});
