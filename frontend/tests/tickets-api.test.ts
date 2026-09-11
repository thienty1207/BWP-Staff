import { afterEach, expect, test } from 'bun:test';
import { listTickets } from '../src/lib/client/tickets/api';

const originalFetch = globalThis.fetch;

const page = {
	tickets: [
		{
			id: 101,
			title: 'Air conditioner request',
			status: 'pending',
			priority: true,
			due_at: null,
			created_at: '2026-09-11T10:00:00Z',
			updated_at: '2026-09-11T10:00:00Z',
			requester: { id: 10, full_name: 'Requester' },
			department: { id: 1, code: 'IT', name: 'IT Department' },
			location: null,
			accepted_by: null,
			accepted_at: null,
			assigned_departments: [],
			assigned_users: [],
			closed_at: null
		}
	],
	page: { has_more: true, next_before_created_at: '2026-09-11T10:00:00Z', next_before_id: 101 }
};

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('ticket list sends the selected view through the relative authenticated endpoint', async () => {
	let requestInput: RequestInfo | URL | undefined;
	let requestInit: RequestInit | undefined;
	globalThis.fetch = async (input, init) => {
		requestInput = input;
		requestInit = init;
		return new Response(JSON.stringify(page), { status: 200, headers: { 'Content-Type': 'application/json' } });
	};

	const result = await listTickets('open', { limit: 50 });

	expect(result).toEqual(page);
	expect(String(requestInput)).toBe('/api/v1/tickets?view=open&limit=50');
	expect(requestInit?.method).toBe('GET');
	expect(requestInit?.credentials).toBe('include');
});

test('ticket list sends both cursor fields for keyset pagination', async () => {
	let requestInput: RequestInfo | URL | undefined;
	globalThis.fetch = async (input) => {
		requestInput = input;
		return new Response(JSON.stringify({ tickets: [], page: { has_more: false, next_before_created_at: null, next_before_id: null } }), { status: 200 });
	};

	await listTickets('closed', {
		limit: 1,
		cursor: { before_created_at: '2026-09-11T10:00:00Z', before_id: 101 }
	});

	expect(String(requestInput)).toBe(
		'/api/v1/tickets?view=closed&limit=1&before_created_at=2026-09-11T10%3A00%3A00Z&before_id=101'
	);
});

test('ticket list keeps authentication and retryable failures typed', async () => {
	globalThis.fetch = async () => new Response(JSON.stringify({ error: { code: 'unauthenticated' } }), { status: 401 });
	try {
		await listTickets('open');
		throw new Error('unauthenticated ticket list unexpectedly succeeded');
	} catch (error) {
		expect(error).toMatchObject({ kind: 'unauthenticated', status: 401 });
	}

	globalThis.fetch = async () => new Response(JSON.stringify({ error: { code: 'internal_server_error' } }), { status: 500 });
	try {
		await listTickets('open');
		throw new Error('failed ticket list unexpectedly succeeded');
	} catch (error) {
		expect(error).toMatchObject({ kind: 'retryable', status: 500 });
	}
});
