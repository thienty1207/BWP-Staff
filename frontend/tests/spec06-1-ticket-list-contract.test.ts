import { expect, test } from 'bun:test';

const pagePath = new URL('../src/routes/+page.svelte', import.meta.url);
const stylesPath = new URL('../src/lib/styles/app.css', import.meta.url);

test('desktop ticket columns follow the approved business hierarchy', async () => {
	const page = await Bun.file(pagePath).text();
	const columns = [...page.matchAll(/<th scope="col">([^<]+)<\/th>/g)].map((match) => match[1].trim());

	expect(columns).toEqual(['Requester', 'Location', 'Title', 'Description', 'Status', 'Owner', 'Created On', 'Due Date']);
	expect(page).not.toContain('ticket-id');
	expect(page).not.toContain('Assignment');
	expect(page).not.toContain('<th scope="col">Department</th>');
	expect(page).not.toContain('<th scope="col">Action</th>');
});

test('ticket list renders real descriptions and accepted_by owner labels', async () => {
	const page = await Bun.file(pagePath).text();
	const styles = await Bun.file(stylesPath).text();

	expect(page).toContain('ticket.description');
	expect(page).toContain('ticket.accepted_by');
	expect(page).toContain('identityLabel');
	expect(page).toContain('ticket-description');
	expect(styles).toContain('-webkit-line-clamp: 2');

	const dueCellIndex = page.indexOf('ticket-due-cell');
	const priorityIndex = page.indexOf('ticket.priority', dueCellIndex);
	expect(dueCellIndex).toBeGreaterThan(-1);
	expect(priorityIndex).toBeGreaterThan(dueCellIndex);
});

test('mobile ticket list uses a compact summary without legacy stacked fields', async () => {
	const page = await Bun.file(pagePath).text();

	expect(page).toContain('ticket-card-meta');
	expect(page).toContain('ticket-card-description');
	expect(page).not.toContain('ticket-card-details');
	expect(page).not.toContain('<dt>Department</dt>');
	expect(page).not.toContain('<dt>Assignment</dt>');
	expect(page).not.toContain('Ticket Detail');
	expect(page).not.toContain('href="/tickets');
});
