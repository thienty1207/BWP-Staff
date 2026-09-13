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

test('Location results render names only while preserving database IDs', async () => {
	const dialogPath = new URL('../src/lib/components/NewRequestDialog.svelte', import.meta.url);
	const dialog = await Bun.file(dialogPath).text();

	expect(dialog).toContain('{location.name}');
	expect(dialog).toContain('locationID = String(location.id)');
	expect(dialog).not.toContain('96BWV-ROOM-1024');
	expect(dialog).not.toContain('96BWV-AREA-017');
	expect(dialog).toContain('<option value={String(department.id)}>{department.name}</option>');
});

test('Location picker is searchable, bounded, and preserves clear-to-null semantics', async () => {
	const dialogPath = new URL('../src/lib/components/NewRequestDialog.svelte', import.meta.url);
	const stylesPath = new URL('../src/lib/styles/app.css', import.meta.url);
	const dialog = await Bun.file(dialogPath).text();
	const styles = await Bun.file(stylesPath).text();

	expect(dialog).toContain('role="combobox"');
	expect(dialog).toContain('role="listbox"');
	expect(dialog).toContain('Searching…');
	expect(dialog).toContain('No locations found.');
	expect(dialog).toContain('Unable to load locations. Retry.');
	expect(dialog).toContain('setTimeout');
	expect(dialog).toContain('}, 180);');
	expect(dialog).toContain('locationRequestSequence');
	expect(dialog).toContain('requestSequence !== locationRequestSequence');
	expect(dialog).toContain('locationID = \'\'');
	expect(dialog).toContain('No location');
	expect(dialog).not.toContain('location.code');
	expect(dialog).not.toContain('BWP-AREA-');
	expect(dialog).not.toContain('96BWV-');
	expect(styles).toContain('new-request-location-results');
	expect(styles).toContain('overflow-y: auto');
	expect(styles).toContain('max-height: min(12rem, 34svh)');
});

test('Location picker exposes the full list and blocks stale keyboard selection while loading', async () => {
	const dialogPath = new URL('../src/lib/components/NewRequestDialog.svelte', import.meta.url);
	const dialog = await Bun.file(dialogPath).text();

	expect(dialog).toContain('getLocations()');
	expect(dialog).toContain('getLocations({ q: query.trim() })');
	expect(dialog).not.toContain('getLocations({ limit: locationSearchLimit })');
	expect(dialog).not.toContain('getLocations({ q: query.trim(), limit: locationSearchLimit })');

	const scheduleStart = dialog.indexOf('function scheduleLocationSearch');
	const searchStart = dialog.indexOf('async function searchLocations');
	const schedule = dialog.slice(scheduleStart, searchStart);
	expect(schedule).toContain('highlightedLocationIndex = -1');

	const keydownStart = dialog.indexOf('function handleLocationKeydown');
	const cancelStart = dialog.indexOf('function handleCancel');
	const keydown = dialog.slice(keydownStart, cancelStart);
	expect(keydown).toContain('if (locationSearchLoading)');
});
