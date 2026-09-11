import { AuthApiError, type AuthenticatedUser } from './model';

const loginPath = '/api/v1/auth/login';
const logoutPath = '/api/v1/auth/logout';
const currentUserPath = '/api/v1/auth/me';

export async function login(username: string, password: string): Promise<AuthenticatedUser> {
	let response: Response;
	try {
		response = await fetch(loginPath, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ username: username.trim(), password })
		});
	} catch {
		throw retryableError();
	}

	if (response.status === 401) {
		throw new AuthApiError('invalid_credentials', 'Invalid username or password.', response.status);
	}
	if (response.status === 400) {
		throw new AuthApiError('invalid_input', 'Please enter a valid username and password.', response.status);
	}
	if (response.status !== 200) {
		throw retryableError(response.status);
	}

	return readUser(response);
}

export async function getCurrentUser(): Promise<AuthenticatedUser> {
	let response: Response;
	try {
		response = await fetch(currentUserPath, {
			method: 'GET',
			credentials: 'include'
		});
	} catch {
		throw retryableError();
	}

	if (response.status === 401) {
		throw new AuthApiError('unauthenticated', 'Unauthenticated.', response.status);
	}
	if (response.status !== 200) {
		throw retryableError(response.status);
	}

	return readUser(response);
}

export async function logout(): Promise<void> {
	let response: Response;
	try {
		response = await fetch(logoutPath, {
			method: 'POST',
			credentials: 'include'
		});
	} catch {
		throw retryableError();
	}

	if (response.status !== 204) {
		throw retryableError(response.status);
	}
}

async function readUser(response: Response): Promise<AuthenticatedUser> {
	let payload: unknown;
	try {
		payload = await response.json();
	} catch {
		throw retryableError(response.status);
	}

	const user = parseUser(payload);
	if (!user) {
		throw retryableError(response.status);
	}
	return user;
}

function parseUser(payload: unknown): AuthenticatedUser | null {
	if (!isRecord(payload) || !isRecord(payload.user)) {
		return null;
	}

	const user = payload.user;
	if (!isRecord(user.department)) {
		return null;
	}

	const department = user.department;
	const avatarURL = user.avatar_url;
	if (
		!isSafeInteger(user.id) ||
		typeof user.username !== 'string' ||
		typeof user.employee_code !== 'string' ||
		typeof user.full_name !== 'string' ||
		typeof user.role !== 'string' ||
		!isSafeInteger(department.id) ||
		typeof department.code !== 'string' ||
		typeof department.name !== 'string' ||
		(avatarURL !== null && typeof avatarURL !== 'string')
	) {
		return null;
	}

	return {
		id: user.id,
		username: user.username,
		employee_code: user.employee_code,
		full_name: user.full_name,
		role: user.role,
		department: {
			id: department.id,
			code: department.code,
			name: department.name
		},
		avatar_url: avatarURL
	};
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function isSafeInteger(value: unknown): value is number {
	return typeof value === 'number' && Number.isSafeInteger(value);
}

function retryableError(status?: number): AuthApiError {
	return new AuthApiError('retryable', 'The authentication service is temporarily unavailable.', status);
}
