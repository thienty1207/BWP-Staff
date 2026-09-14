import type { TicketSummary } from './model';

export type TicketDetailStatus = 'closed' | 'loading' | 'ready' | 'error' | 'not_found';

export type TicketDetailState = {
	status: TicketDetailStatus;
	selectedTicketID: number | null;
	ticket: TicketSummary | null;
	errorMessage: string;
};

export type TicketDetailRequest = {
	ticketID: number;
	sequence: number;
};

function closedState(): TicketDetailState {
	return {
		status: 'closed',
		selectedTicketID: null,
		ticket: null,
		errorMessage: ''
	};
}

export class TicketDetailStateMachine {
	private sequence = 0;
	private currentState: TicketDetailState = closedState();

	get state(): TicketDetailState {
		return this.currentState;
	}

	begin(ticketID: number): TicketDetailRequest {
		const request = { ticketID, sequence: ++this.sequence };
		this.currentState = {
			status: 'loading',
			selectedTicketID: ticketID,
			ticket: null,
			errorMessage: ''
		};
		return request;
	}

	close(): void {
		this.sequence += 1;
		this.currentState = closedState();
	}

	succeed(request: TicketDetailRequest, ticket: TicketSummary): boolean {
		if (!this.isCurrent(request)) {
			return false;
		}
		this.currentState = {
			status: 'ready',
			selectedTicketID: request.ticketID,
			ticket,
			errorMessage: ''
		};
		return true;
	}

	fail(request: TicketDetailRequest, status: 'error' | 'not_found', errorMessage: string): boolean {
		if (!this.isCurrent(request)) {
			return false;
		}
		this.currentState = {
			status,
			selectedTicketID: request.ticketID,
			ticket: null,
			errorMessage
		};
		return true;
	}

	isCurrent(request: TicketDetailRequest): boolean {
		return (
			request.sequence === this.sequence &&
			this.currentState.status === 'loading' &&
			this.currentState.selectedTicketID === request.ticketID
		);
	}
}
