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
	expect(dialog).toContain('department_unavailable');
	expect(dialog).toContain('location_unavailable');
	expect(dialog).toContain('The selected department is no longer available. Please select another department.');
	expect(dialog).toContain('The selected location is no longer available. Please select another location.');
	expect(dialog).toContain('Please review the request details and try again.');
	expect(page).toContain('New Request');
	expect(page).not.toContain('Refresh');
});
