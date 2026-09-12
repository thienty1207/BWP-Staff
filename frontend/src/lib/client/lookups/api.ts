import { LookupApiError, type LookupDepartment, type LookupLocation } from './model';

const departmentsPath = '/api/v1/departments';
const locationsPath = '/api/v1/locations';

export async function getDepartments(): Promise<LookupDepartment[]> {
	const payload = await getLookupPayload(departmentsPath);
	if (!isRecord(payload) || !Array.isArray(payload.departments)) {
		throw retryableError(200);
	}

	const departments = payload.departments.map(parseDepartment);
	if (departments.some((department) => department === null)) {
		throw retryableError(200);
	}
	return departments as LookupDepartment[];
}

export async function getLocations(): Promise<LookupLocation[]> {
	const payload = await getLookupPayload(locationsPath);
	if (!isRecord(payload) || !Array.isArray(payload.locations)) {
		throw retryableError(200);
	}

	const locations = payload.locations.map(parseLocation);
	if (locations.some((location) => location === null)) {
		throw retryableError(200);
	}
	return locations as LookupLocation[];
}

async function getLookupPayload(path: string): Promise<unknown> {
	let response: Response;
	try {
		response = await fetch(path, { method: 'GET', credentials: 'include' });
	} catch {
		throw retryableError();
	}

	if (response.status === 401) {
		throw new LookupApiError('unauthenticated', 'Unauthenticated.', response.status);
	}
	if (response.status !== 200) {
		throw retryableError(response.status);
	}

	try {
		return await response.json();
	} catch {
		throw retryableError(response.status);
	}
}

function parseDepartment(value: unknown): LookupDepartment | null {
	if (!isRecord(value) || !isPositiveSafeInteger(value.id) || typeof value.code !== 'string' || typeof value.name !== 'string') {
		return null;
	}
	return { id: value.id, code: value.code, name: value.name };
}

function parseLocation(value: unknown): LookupLocation | null {
	if (!isRecord(value) || !isPositiveSafeInteger(value.id) || typeof value.name !== 'string') {
		return null;
	}
	if (value.code !== null && typeof value.code !== 'string') {
		return null;
	}
	return { id: value.id, code: value.code, name: value.name };
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function isPositiveSafeInteger(value: unknown): value is number {
	return typeof value === 'number' && Number.isSafeInteger(value) && value > 0;
}

function retryableError(status?: number): LookupApiError {
	return new LookupApiError('retryable', 'The lookup service is temporarily unavailable.', status);
}
