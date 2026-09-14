import { afterEach, expect, test } from 'bun:test';
import { getTicket } from '../src/lib/client/tickets/api';
import { TicketDetailStateMachine } from '../src/lib/client/tickets/detail-state';
import type { TicketSummary } from '../src/lib/client/tickets/model';

const originalFetch = globalThis.fetch;

const ticket: TicketSummary = {
	id: 101,
	title: 'Air conditioner request',
	description: 'The complete detail description.',
	status: 'accepted',
	priority: true,
	due_at: '2026-09-14T10:00:00Z',
	created_at: '2026-09-14T08:00:00Z',
	updated_at: '2026-09-14T08:30:00Z',
	requester: { id: 10, full_name: 'Requester', department_code: 'REC' },
	department: { id: 1, code: 'HK', name: 'Housekeeping' },
	location: { id: 2, code: 'ROOM-1', name: 'Room 1' },
	accepted_by: { id: 20, full_name: 'Owner', department_code: 'HK' },
	accepted_at: '2026-09-14T08:20:00Z',
	assigned_departments: [{ id: 3, code: 'FO', name: 'Front Office' }],
	assigned_users: [{ id: 30, full_name: 'Assignee', department_code: 'FO' }],
	closed_at: null
};

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('getTicket sends one authenticated detail GET and parses the ticket envelope strictly', async () => {
	let requestInput: RequestInfo | URL | undefined;
	let requestInit: RequestInit | undefined;
	globalThis.fetch = async (input, init) => {
		requestInput = input;
		requestInit = init;
		return new Response(JSON.stringify({ ticket }), { status: 200, headers: { 'Content-Type': 'application/json' } });
	};

	const result = await getTicket(ticket.id);

	expect(result).toEqual(ticket);
	expect(String(requestInput)).toBe('/api/v1/tickets/101');
	expect(requestInit?.method).toBe('GET');
	expect(requestInit?.credentials).toBe('include');
});

test('getTicket maps validation, auth, not-found, and server failures safely', async () => {
	globalThis.fetch = async () => new Response(JSON.stringify({ error: { code: 'invalid_request' } }), { status: 400 });
	await expect(getTicket(101)).rejects.toMatchObject({ kind: 'invalid_input', status: 400 });

	globalThis.fetch = async () => new Response('{}', { status: 401 });
	await expect(getTicket(101)).rejects.toMatchObject({ kind: 'unauthenticated', status: 401 });

	globalThis.fetch = async () => new Response(JSON.stringify({ error: { code: 'ticket_not_found', message: 'do not render backend text' } }), { status: 404 });
	await expect(getTicket(101)).rejects.toMatchObject({ kind: 'not_found', status: 404, code: 'ticket_not_found' });

	globalThis.fetch = async () => new Response('{}', { status: 503 });
	await expect(getTicket(101)).rejects.toMatchObject({ kind: 'retryable', status: 503 });
});

test('getTicket treats malformed successful payloads as retryable', async () => {
	const malformedPayloads = [
		{},
		{ ticket: { ...ticket, description: 42 } },
		{ ticket: { ...ticket, assigned_users: null } },
		{ ticket: { ...ticket, accepted_at: 'not-a-timestamp' } }
	];

	for (const payload of malformedPayloads) {
		globalThis.fetch = async () => new Response(JSON.stringify(payload), { status: 200 });
		await expect(getTicket(101)).rejects.toMatchObject({ kind: 'retryable', status: 200 });
	}
});

test('getTicket maps network failures to retryable', async () => {
	globalThis.fetch = async () => {
		throw new Error('offline');
	};

	await expect(getTicket(101)).rejects.toMatchObject({ kind: 'retryable' });
});

test('detail selection clears stale content and latest selection wins', () => {
	const machine = new TicketDetailStateMachine();
	const ticketA = { ...ticket, id: 1, title: 'Ticket A' };
	const ticketB = { ...ticket, id: 2, title: 'Ticket B' };

	const requestA = machine.begin(ticketA.id);
	expect(machine.state).toMatchObject({ status: 'loading', selectedTicketID: 1, ticket: null });
	expect(machine.succeed(requestA, ticketA)).toBe(true);

	const requestB = machine.begin(ticketB.id);
	expect(machine.state).toMatchObject({ status: 'loading', selectedTicketID: 2, ticket: null });
	expect(machine.succeed(requestB, ticketB)).toBe(true);
	expect(machine.succeed(requestA, ticketA)).toBe(false);
	expect(machine.state).toMatchObject({ status: 'ready', selectedTicketID: 2, ticket: ticketB });
});

test('detail close invalidates in-flight responses without reopening the view', () => {
	const machine = new TicketDetailStateMachine();
	const request = machine.begin(ticket.id);

	machine.close();

	expect(machine.succeed(request, ticket)).toBe(false);
	expect(machine.state).toEqual({ status: 'closed', selectedTicketID: null, ticket: null, errorMessage: '' });
});

test('Tickets page keeps detail interactions separate from list controls and exposes accessible detail states', async () => {
	const page = await Bun.file(new URL('../src/routes/+page.svelte', import.meta.url)).text();
	const detail = await Bun.file(new URL('../src/lib/components/TicketDetail.svelte', import.meta.url)).text();
	const styles = await Bun.file(new URL('../src/lib/styles/app.css', import.meta.url)).text();

	expect(page).toContain('getTicket');
	expect(page).toContain('TicketDetailStateMachine');
	expect(page).toContain('ticketRequestInFlight');
	expect(page).toContain('detailState');
	expect(page).toContain('openNewRequest');
	expect(page).toContain('switchView');
	expect(page).toContain('loadMore');
	expect(detail).toContain('Back to Tickets');
	expect(detail).toContain('role="status"');
	expect(detail).toContain('role="alert"');
	expect(detail).toContain('Description');
	expect(detail).toContain('Assigned Departments');
	expect(detail).toContain('Assigned Users');
	expect(detail).toContain("description?.trim() || '—'");
	expect(detail).toContain("values.length === 0");
	expect(detail).toContain('ticket-title-priority');
	expect(detail).not.toContain('{@html');
	expect(styles).toContain('.tickets-workspace');
	expect(styles).toContain('grid-template-columns: minmax(0, 1fr) clamp(20rem, 28vw, 26rem)');
	expect(styles).toContain('.ticket-detail');
	expect(styles).toContain('.ticket-detail-back');
	expect(styles).toContain('border: 1px solid var(--danger)');
	expect(styles).toContain('white-space: pre-wrap');
	expect(styles).toContain('overflow-wrap: anywhere');

	const switchViewStart = page.indexOf('function switchView');
	const switchViewEnd = page.indexOf('\n\t}\n', switchViewStart);
	expect(page.slice(switchViewStart, switchViewEnd)).toContain('closeTicketDetail();');

	const openNewRequestStart = page.indexOf('function openNewRequest');
	const openNewRequestEnd = page.indexOf('\n\t}\n', openNewRequestStart);
	expect(page.slice(openNewRequestStart, openNewRequestEnd)).toContain('closeTicketDetail();');

	const retryDetailStart = page.indexOf('function retryTicketDetail');
	const retryDetailEnd = page.indexOf('\n\t}\n', retryDetailStart);
	const retryDetailBody = page.slice(retryDetailStart, retryDetailEnd);
	expect(retryDetailBody).toContain('loadTicketDetail(request)');
	expect(retryDetailBody).not.toContain('loadFirstPage');
	expect(page).toContain('onBack={closeTicketDetail}');
});
