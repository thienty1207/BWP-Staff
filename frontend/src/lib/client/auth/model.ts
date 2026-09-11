export type AuthenticatedUser = {
	id: number;
	username: string;
	employee_code: string;
	full_name: string;
	role: string;
	department: {
		id: number;
		code: string;
		name: string;
	};
	avatar_url: string | null;
};

export type AuthApiErrorKind =
	| 'invalid_credentials'
	| 'invalid_input'
	| 'unauthenticated'
	| 'retryable';

export class AuthApiError extends Error {
	readonly kind: AuthApiErrorKind;
	readonly status?: number;

	constructor(kind: AuthApiErrorKind, message: string, status?: number) {
		super(message);
		this.name = 'AuthApiError';
		this.kind = kind;
		this.status = status;
	}
}
