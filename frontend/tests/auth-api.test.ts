import { afterEach, expect, test } from 'bun:test';
import { getCurrentUser, login, logout } from '../src/lib/client/auth/api';

const originalFetch = globalThis.fetch;

const user = {
	id: 1,
	username: 'test-user',
	employee_code: 'TEST-001',
	full_name: 'Test User',
	role: 'staff',
	department: { id: 1, code: 'TEST', name: 'Test Department' },
	avatar_url: null
};

function jsonResponse(body: unknown, status: number): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
}

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('login sends a relative request with trimmed username and untouched password', async () => {
	let requestInput: RequestInfo | URL | undefined;
	let requestInit: RequestInit | undefined;
	globalThis.fetch = async (input, init) => {
		requestInput = input;
		requestInit = init;
		return jsonResponse({ user }, 200);
	};

	const result = await login('  test-user  ', ' pass word ');

	expect(result).toEqual(user);
	expect(requestInput).toBe('/api/v1/auth/login');
	expect(requestInit?.credentials).toBe('include');
	expect(requestInit?.method).toBe('POST');
	expect(requestInit?.headers).toEqual({ 'Content-Type': 'application/json' });
	expect(JSON.parse(String(requestInit?.body))).toEqual({
		username: 'test-user',
		password: ' pass word '
	});
});

test('login maps invalid credentials to a generic typed failure', async () => {
	globalThis.fetch = async () => jsonResponse({ error: { code: 'invalid_credentials' } }, 401);

	try {
		await login('hothienty', 'wrong');
		throw new Error('login unexpectedly succeeded');
	} catch (error) {
		expect(error).toMatchObject({ kind: 'invalid_credentials', status: 401 });
	}
});

test('current-user rejects only authentication failures as unauthenticated', async () => {
	globalThis.fetch = async () => jsonResponse({ error: { code: 'unauthenticated' } }, 401);

	try {
		await getCurrentUser();
		throw new Error('current-user unexpectedly succeeded');
	} catch (error) {
		expect(error).toMatchObject({ kind: 'unauthenticated', status: 401 });
	}
});

test('logout maps service failure to a retryable typed failure', async () => {
	globalThis.fetch = async () => jsonResponse({ error: { code: 'internal_server_error' } }, 500);

	try {
		await logout();
		throw new Error('logout unexpectedly succeeded');
	} catch (error) {
		expect(error).toMatchObject({ kind: 'retryable', status: 500 });
	}
});
