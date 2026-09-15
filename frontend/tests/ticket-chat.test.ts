import { afterEach, expect, test } from 'bun:test';
import { getTicket } from '../src/lib/client/tickets/api';
import { TicketChatStateMachine } from '../src/lib/client/ticket-chat-state';
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

test('chat selection clears stale content and latest selection wins', () => {
	const machine = new TicketChatStateMachine();
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

test('chat close invalidates in-flight responses without reopening the view', () => {
	const machine = new TicketChatStateMachine();
	const request = machine.begin(ticket.id);

	machine.close();

	expect(machine.succeed(request, ticket)).toBe(false);
	expect(machine.state).toEqual({ status: 'closed', selectedTicketID: null, ticket: null, errorMessage: '' });
});

test('Tickets page keeps chat interactions separate from list controls and exposes accessible chat states', async () => {
	const page = await Bun.file(new URL('../src/routes/+page.svelte', import.meta.url)).text();
	const chat = await Bun.file(new URL('../src/lib/components/TicketChat.svelte', import.meta.url)).text();
	const styles = await Bun.file(new URL('../src/lib/styles/app.css', import.meta.url)).text();

	expect(page).toContain('getTicket');
	expect(page).toContain('TicketChatStateMachine');
	expect(page).toContain('ticketRequestInFlight');
	expect(page).toContain('chatState');
	expect(page).toContain('openNewRequest');
	expect(page).toContain('switchView');
	expect(page).toContain('loadMore');
	expect(page).toContain('TicketChat');
	expect(page).not.toContain('TicketDetail');
	expect(chat).toContain('Back to Tickets');
	expect(chat).toContain('aria-labelledby="ticket-chat-heading"');
	expect(chat).toContain('role="status"');
	expect(chat).toContain('role="alert"');
	expect(chat).toContain('Chat');
	expect(chat).toContain('Chats');
	expect(chat).toContain('Checklist');
	expect(chat).toContain('aria-selected="true"');
	expect(chat).toContain('aria-disabled="true"');
	expect(chat).toContain('has created a new request');
	expect(chat).toContain('Accepted by');
	expect(chat).toContain('Type a message');
	expect(chat).toContain('disabled');
	expect(chat).not.toContain('Ticket Detail');
	expect(chat).not.toContain('Request information');
	expect(chat).not.toContain('Assigned Departments');
	expect(chat).not.toContain('Assigned Users');
	expect(chat).not.toContain('Due');
	expect(chat).not.toContain('Updated');
	expect(chat).not.toContain('Closed');
	expect(chat).not.toContain('{@html');
	expect(chat).not.toContain('fetch(');
	expect(chat).not.toContain('onsubmit');
	expect(chat).not.toContain('POST');
	expect(styles).toContain('.tickets-workspace');
	expect(styles).toContain('.tickets-workspace.chat-open');
	expect(styles).toContain('grid-template-columns: minmax(0, 1fr) clamp(19rem, 22vw, 22rem)');
	expect(styles).toContain('.ticket-chat');
	expect(styles).toContain('.ticket-chat-back');
	expect(styles).toContain('.ticket-chat-conversation');
	expect(styles).toContain('overflow-y: auto');
	expect(styles).toContain('border: 1px solid var(--danger)');

	const switchViewStart = page.indexOf('function switchView');
	const switchViewEnd = page.indexOf('\n\t}\n', switchViewStart);
	expect(page.slice(switchViewStart, switchViewEnd)).toContain('closeTicketChat();');

	const openNewRequestStart = page.indexOf('function openNewRequest');
	const openNewRequestEnd = page.indexOf('\n\t}\n', openNewRequestStart);
	expect(page.slice(openNewRequestStart, openNewRequestEnd)).toContain('closeTicketChat();');

	const retryChatStart = page.indexOf('function retryTicketChat');
	const retryChatEnd = page.indexOf('\n\t}\n', retryChatStart);
	const retryChatBody = page.slice(retryChatStart, retryChatEnd);
	expect(retryChatBody).toContain('loadTicketChat(request)');
	expect(retryChatBody).not.toContain('loadFirstPage');
	expect(page).toContain('onBack={closeTicketChat}');
});

test('ticket rows and cards open Chat while the title remains a single semantic activation', async () => {
	const page = await Bun.file(new URL('../src/routes/+page.svelte', import.meta.url)).text();

	const tableRows = page.slice(page.indexOf('<tbody>'), page.indexOf('</tbody>'));
	const mobileCards = page.slice(page.indexOf('<div class="mobile-ticket-cards">'), page.indexOf('{#if page.has_more'));

	expect(tableRows).toContain('class="ticket-row-clickable"');
	expect(tableRows).toContain('class:ticket-row-selected={chatState.selectedTicketID === ticket.id}');
	expect(tableRows).toContain('onclick={() => openTicketChat(ticket.id)}');
	expect(mobileCards).toContain('class="ticket-card ticket-card-clickable"');
	expect(mobileCards).toContain('class:ticket-card-selected={chatState.selectedTicketID === ticket.id}');
	expect(mobileCards).toContain('onclick={() => openTicketChat(ticket.id)}');

	expect(page).toContain('class="ticket-title-button"');
	expect(page).toContain('function openTicketChatFromTitle(event: MouseEvent, ticketID: number)');
	expect(page).toContain('event.stopPropagation();');
	expect(page).toMatch(/openTicketChatFromTitle\(event, ticket\.id\)/);

	const titleActivations = page.match(/openTicketChatFromTitle\(event, ticket\.id\)/g) ?? [];
	expect(titleActivations).toHaveLength(2);
});

test('Chat presentation uses compact Sara-aligned semantics and scoped disabled cursors', async () => {
	const chat = await Bun.file(new URL('../src/lib/components/TicketChat.svelte', import.meta.url)).text();
	const styles = await Bun.file(new URL('../src/lib/styles/app.css', import.meta.url)).text();

	expect(chat).not.toContain('<p class="eyebrow">Tickets</p>');
	expect(chat).toContain('<h2 id="ticket-chat-heading">Chat</h2>');
	expect(chat).toContain('ticket-chat-summary-meta');
	expect(chat).toContain('by {identityLabel(state.ticket.accepted_by)}');
	expect(chat).toContain('class:accepted={state.ticket.status === \'accepted\'}');
	expect(chat).toContain('class:closed={state.ticket.status === \'closed\'}');

	const summaryStart = chat.indexOf('<div class="ticket-chat-summary">');
	const summaryEnd = chat.indexOf('{:else if state.status === \'loading\'}', summaryStart);
	const summaryMarkup = chat.slice(summaryStart, summaryEnd);
	expect(summaryMarkup.match(/{state\.ticket\.title}/g) ?? []).toHaveLength(1);
	expect(chat).not.toContain('ticket-chat-summary-list');

	expect(styles).toContain('.status-badge.accepted');
	expect(styles).toContain('.ticket-row-clickable');
	expect(styles).toContain('.ticket-row-selected');
	expect(styles).toContain('.ticket-card-clickable');
	expect(styles).toContain('.ticket-card-selected');
	expect(styles).toContain('.ticket-chat-tab:disabled');
	expect(styles).toContain('.ticket-chat-icon-button:disabled');
	expect(styles).toContain('.ticket-chat-actions button:disabled');
	expect(styles).toContain('cursor: default;');
	expect(styles).toContain('button:disabled {\n\tcursor: wait;');
});

test('canonical SPEC makes whole-ticket pointer activation mandatory', async () => {
	const spec = await Bun.file(
		new URL('../../Context-Spec-BWP-SonaSea/Spec/SPEC-07-ticket-chat-shell-conversation-foundation.md', import.meta.url)
	).text();

	expect(spec).toContain('pointer click/tap anywhere on the ticket row/card MUST open Chat');
	expect(spec).toContain('whole row/card pointer target');
	expect(spec).toContain('one user activation = one GET request');
	expect(spec).toContain('selected ticket has visible but subtle selection feedback');
	expect(spec).not.toContain('row/card pointer click also opens Chat');
});

test('Ticket Chat keeps the summary compact and uses real persisted activity only', async () => {
	const chat = await Bun.file(new URL('../src/lib/components/TicketChat.svelte', import.meta.url)).text();

	expect(chat).toContain('<h3 class="ticket-chat-title"');
	expect(chat).toContain('class:ticket-title-priority={state.ticket.priority}');
	expect(chat).toContain('location?.name || \'—\'');
	expect(chat).not.toContain('location.code');
	expect(chat).toContain('state.ticket.accepted_by && state.ticket.accepted_at');
	expect(chat).not.toContain('assigned_departments');
	expect(chat).not.toContain('assigned_users');
	expect(chat).not.toContain('closed_at');
	expect(chat).not.toContain('updated_at');
	expect(chat).not.toContain('due_at');
});

test('Ticket Chat does not expose operational mutations in the shell', async () => {
	const chat = await Bun.file(new URL('../src/lib/components/TicketChat.svelte', import.meta.url)).text();

	expect(chat).toContain('aria-label="Attach file"');
	expect(chat).toContain('aria-label="Voice message"');
	expect(chat).toContain('aria-label="More message options"');
	expect(chat).toContain('>Accept</button>');
	expect(chat).toContain('>Assign</button>');
	expect(chat).toContain('>Close</button>');
	expect(chat).not.toContain('onclick={onAccept}');
	expect(chat).not.toContain('onclick={onAssign}');
	expect(chat).not.toContain('onclick={onCloseTicket}');
});

test('New Request keeps create-ticket error codes narrowly scoped', async () => {
	const newRequest = await Bun.file(new URL('../src/lib/components/NewRequestDialog.svelte', import.meta.url)).text();

	expect(newRequest).toContain('type CreateTicketErrorCode');
	expect(newRequest).toContain('function createTicketErrorMessage(code?: CreateTicketErrorCode)');
	expect(newRequest).not.toContain('TicketApiErrorCode');
});
