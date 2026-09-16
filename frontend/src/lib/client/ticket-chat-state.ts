import type { TicketSummary } from './tickets/model';

export type TicketChatStatus = 'closed' | 'loading' | 'ready' | 'error' | 'not_found';

export type TicketChatState = {
	status: TicketChatStatus;
	selectedTicketID: number | null;
	ticket: TicketSummary | null;
	errorMessage: string;
};

export type TicketChatRequest = {
	ticketID: number;
	sequence: number;
};

function closedState(): TicketChatState {
	return {
		status: 'closed',
		selectedTicketID: null,
		ticket: null,
		errorMessage: ''
	};
}

export class TicketChatStateMachine {
	private sequence = 0;
	private currentState: TicketChatState = closedState();

	get state(): TicketChatState {
		return this.currentState;
	}

	begin(ticketID: number): TicketChatRequest {
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

	succeed(request: TicketChatRequest, ticket: TicketSummary): boolean {
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

	updateSelected(ticketID: number, ticket: TicketSummary): boolean {
		if (this.currentState.status === 'closed' || this.currentState.selectedTicketID !== ticketID) {
			return false;
		}
		this.currentState = {
			status: 'ready',
			selectedTicketID: ticketID,
			ticket,
			errorMessage: ''
		};
		return true;
	}

	fail(request: TicketChatRequest, status: 'error' | 'not_found', errorMessage: string): boolean {
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

	isCurrent(request: TicketChatRequest): boolean {
		return (
			request.sequence === this.sequence &&
			this.currentState.status === 'loading' &&
			this.currentState.selectedTicketID === request.ticketID
		);
	}
}
