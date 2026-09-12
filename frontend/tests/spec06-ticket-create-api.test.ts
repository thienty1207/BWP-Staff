import { afterEach, expect, test } from 'bun:test';
import { createTicket } from '../src/lib/client/tickets/api';
import { TicketApiError } from '../src/lib/client/tickets/model';

const originalFetch = globalThis.fetch;

const ticket = {
	id: 102,
	title: 'Lobby TV is not displaying content',
	status: 'pending',
	priority: false,
	due_at: '2026-09-12T02:30:00Z',
	created_at: '2026-09-12T02:00:00Z',
	updated_at: '2026-09-12T02:00:00Z',
	requester: { id: 10, full_name: 'Requester' },
	department: { id: 2, code: 'IT', name: 'IT Department' },
	location: null,
	accepted_by: null,
	accepted_at: null,
	assigned_departments: [],
	assigned_users: [],
	closed_at: null
};

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('createTicket sends one relative authenticated POST and parses a 201 ticket', async () => {
	let requestInput: RequestInfo | URL | undefined;
	let requestInit: RequestInit | undefined;
	globalThis.fetch = async (input, init) => {
		requestInput = input;
		requestInit = init;
		return new Response(JSON.stringify({ ticket }), { status: 201, headers: { 'Content-Type': 'application/json' } });
	};

	const request = {
		department_id: 2,
		location_id: null,
		title: 'Lobby TV is not displaying content',
		description: 'Optional details',
		priority: false,
		due_at: '2026-09-12T09:30:00+07:00'
	};

	await expect(createTicket(request)).resolves.toEqual(ticket);
	expect(String(requestInput)).toBe('/api/v1/tickets');
	expect(requestInit?.method).toBe('POST');
	expect(requestInit?.credentials).toBe('include');
	expect(requestInit?.headers).toEqual({ 'Content-Type': 'application/json' });
	expect(requestInit?.body).toBe(JSON.stringify(request));
});

test('createTicket maps 400, 401, 500 and network failures without retrying', async () => {
	let badRequestCalls = 0;
	globalThis.fetch = async () => {
		badRequestCalls += 1;
		return new Response(JSON.stringify({ error: { code: 'department_unavailable' } }), { status: 400 });
	};
	await expect(createTicket({ department_id: 2, location_id: null, title: 'Request', description: null, priority: false, due_at: null })).rejects.toMatchObject({ kind: 'invalid_input', status: 400 });

	globalThis.fetch = async () => new Response('{}', { status: 401 });
	await expect(createTicket({ department_id: 2, location_id: null, title: 'Request', description: null, priority: false, due_at: null })).rejects.toMatchObject({ kind: 'unauthenticated', status: 401 });

	globalThis.fetch = async () => new Response('{}', { status: 500 });
	await expect(createTicket({ department_id: 2, location_id: null, title: 'Request', description: null, priority: false, due_at: null })).rejects.toMatchObject({ kind: 'retryable', status: 500 });

	let networkCalls = 0;
	globalThis.fetch = async () => {
		networkCalls += 1;
		throw new Error('network');
	};
	await expect(createTicket({ department_id: 2, location_id: null, title: 'Request', description: null, priority: false, due_at: null })).rejects.toMatchObject({ kind: 'retryable' });
	expect(badRequestCalls).toBe(1);
	expect(networkCalls).toBe(1);
});

test('createTicket preserves known safe 400 backend error codes', async () => {
	const request = { department_id: 2, location_id: null, title: 'Request', description: null, priority: false, due_at: null };
	for (const code of ['department_unavailable', 'location_unavailable', 'invalid_request']) {
		globalThis.fetch = async () => new Response(JSON.stringify({ error: { code } }), { status: 400 });
		await expect(createTicket(request)).rejects.toMatchObject({ kind: 'invalid_input', status: 400, code });
	}
});

test('createTicket safely falls back for malformed or unknown 400 errors', async () => {
	const request = { department_id: 2, location_id: null, title: 'Request', description: null, priority: false, due_at: null };
	const responses = [
		new Response('{not-json', { status: 400 }),
		new Response(JSON.stringify({ error: { code: 'unexpected_backend_code', message: 'do not show this' } }), { status: 400 }),
		new Response(JSON.stringify({ error: { message: 'do not show this either' } }), { status: 400 })
	];

	for (const response of responses) {
		globalThis.fetch = async () => response;
		let caught: unknown;
		try {
			await createTicket(request);
		} catch (error) {
			caught = error;
		}
		expect(caught).toBeInstanceOf(TicketApiError);
		expect(caught).toMatchObject({ kind: 'invalid_input', status: 400 });
		expect((caught as TicketApiError).code).toBeUndefined();
		expect((caught as Error).message).not.toContain('do not show this');
	}
});
