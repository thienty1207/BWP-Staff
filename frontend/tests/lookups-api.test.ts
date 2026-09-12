import { afterEach, expect, test } from 'bun:test';
import { getDepartments, getLocations } from '../src/lib/client/lookups/api';

const originalFetch = globalThis.fetch;

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('lookup APIs use relative authenticated endpoints and preserve nullable location codes', async () => {
	const requests: string[] = [];
	globalThis.fetch = async (input, init) => {
		requests.push(`${String(input)}:${init?.credentials}`);
		if (String(input) === '/api/v1/departments') {
			return new Response(JSON.stringify({ departments: [{ id: 2, code: 'IT', name: 'IT Department' }] }), { status: 200 });
		}
		return new Response(JSON.stringify({ locations: [{ id: 15, code: null, name: 'Lobby' }] }), { status: 200 });
	};

	await expect(getDepartments()).resolves.toEqual([{ id: 2, code: 'IT', name: 'IT Department' }]);
	await expect(getLocations()).resolves.toEqual([{ id: 15, code: null, name: 'Lobby' }]);
	expect(requests).toEqual(['/api/v1/departments:include', '/api/v1/locations:include']);
});

test('lookup APIs distinguish unauthenticated and retryable/malformed responses', async () => {
	globalThis.fetch = async () => new Response('{}', { status: 401 });
	await expect(getDepartments()).rejects.toMatchObject({ kind: 'unauthenticated', status: 401 });

	globalThis.fetch = async () => new Response('{}', { status: 503 });
	await expect(getLocations()).rejects.toMatchObject({ kind: 'retryable', status: 503 });

	globalThis.fetch = async () => new Response(JSON.stringify({ locations: [{ id: 0, code: 'BAD', name: 'Bad' }] }), { status: 200 });
	await expect(getLocations()).rejects.toMatchObject({ kind: 'retryable', status: 200 });
});

test('RFC3339 ticket timestamps require an explicit timezone-shaped value', async () => {
	const { isRFC3339Timestamp } = await import('../src/lib/client/tickets/api');

	expect(isRFC3339Timestamp('2026-09-12T09:30:00+07:00')).toBe(true);
	expect(isRFC3339Timestamp('2026-09-12T02:30:00.123Z')).toBe(true);
	expect(isRFC3339Timestamp('September 12, 2026 09:30:00')).toBe(false);
	expect(isRFC3339Timestamp('2026-09-12')).toBe(false);
	expect(isRFC3339Timestamp('2026-09-12T09:30:00')).toBe(false);
	expect(isRFC3339Timestamp('2026-02-30T09:30:00Z')).toBe(false);
});
