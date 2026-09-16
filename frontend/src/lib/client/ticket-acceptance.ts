import { acceptTicket, getTicket } from './tickets/api';
import { TicketApiError, type TicketSummary, type TicketView } from './tickets/model';

export type AcceptOrigin = 'row' | 'chat';

export type AcceptViewContext = {
	listSequence: number;
	listView: TicketView;
	chatGeneration: number;
	chatOpen: boolean;
	selectedTicketID: number | null;
};

export type AcceptMutationRequest = AcceptViewContext & {
	ticketID: number;
	origin: AcceptOrigin;
};

export type TicketAcceptanceApi = {
	acceptTicket: (ticketID: number) => Promise<TicketSummary>;
	getTicket: (ticketID: number) => Promise<TicketSummary>;
};

export type TicketAcceptanceCallbacks = {
	getCurrentTicket: (ticketID: number, origin: AcceptOrigin) => TicketSummary | null;
	getView: () => AcceptViewContext;
	patchList: (ticket: TicketSummary) => void;
	updateChat: (ticket: TicketSummary) => void;
	setError: (origin: AcceptOrigin, ticketID: number, message: string, retryable: boolean) => void;
	clearError: (origin: AcceptOrigin, ticketID: number) => void;
	onUnauthenticated: (request: AcceptMutationRequest) => void | Promise<void>;
	onInFlightChange: (ticketID: number, inFlight: boolean) => void;
};

export class TicketAcceptanceController {
	private readonly api: TicketAcceptanceApi;
	private readonly inFlightTicketIDs = new Set<number>();

	constructor(api: Partial<TicketAcceptanceApi> = {}) {
		this.api = {
			acceptTicket: api.acceptTicket ?? acceptTicket,
			getTicket: api.getTicket ?? getTicket
		};
	}

	isInFlight(ticketID: number): boolean {
		return this.inFlightTicketIDs.has(ticketID);
	}

	async acceptTicketByID(ticketID: number, origin: AcceptOrigin, callbacks: TicketAcceptanceCallbacks): Promise<void> {
		if (this.inFlightTicketIDs.has(ticketID)) {
			return;
		}

		const currentTicket = callbacks.getCurrentTicket(ticketID, origin);
		if (!currentTicket || currentTicket.id !== ticketID || currentTicket.status !== 'pending') {
			return;
		}

		const request: AcceptMutationRequest = {
			...callbacks.getView(),
			ticketID,
			origin
		};
		this.inFlightTicketIDs.add(ticketID);
		callbacks.onInFlightChange(ticketID, true);
		callbacks.clearError(origin, ticketID);

		try {
			const updatedTicket = await this.api.acceptTicket(ticketID);
			this.applyTicketResult(request, updatedTicket, callbacks);
		} catch (error) {
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				await callbacks.onUnauthenticated(request);
				return;
			}

			if (error instanceof TicketApiError && (error.kind === 'already_accepted' || error.kind === 'closed')) {
				if (this.isOriginCurrent(request, callbacks.getView())) {
					callbacks.setError(
						origin,
						ticketID,
						error.kind === 'already_accepted'
							? 'This ticket was already accepted by another staff member.'
							: 'This ticket is already closed.',
						false
					);
				}
				await this.refreshAfterConflict(request, callbacks);
				return;
			}

			if (this.isOriginCurrent(request, callbacks.getView())) {
				callbacks.setError(
					origin,
					ticketID,
					error instanceof TicketApiError && error.kind === 'not_found'
						? 'Ticket not found. It may have been removed.'
						: 'Unable to accept this ticket. Please try again.',
					error instanceof TicketApiError && error.kind !== 'not_found'
				);
			}
		} finally {
			this.inFlightTicketIDs.delete(ticketID);
			callbacks.onInFlightChange(ticketID, false);
		}
	}

	private applyTicketResult(
		request: AcceptMutationRequest,
		updatedTicket: TicketSummary,
		callbacks: TicketAcceptanceCallbacks
	): void {
		const view = callbacks.getView();
		if (this.isListCurrent(request, view)) {
			callbacks.patchList(updatedTicket);
		}
		if (this.isChatCurrent(request, view)) {
			callbacks.updateChat(updatedTicket);
		}
		if (this.isOriginCurrent(request, view)) {
			callbacks.clearError(request.origin, request.ticketID);
		}
	}

	private async refreshAfterConflict(request: AcceptMutationRequest, callbacks: TicketAcceptanceCallbacks): Promise<void> {
		try {
			const refreshedTicket = await this.api.getTicket(request.ticketID);
			this.applyTicketResult(request, refreshedTicket, callbacks);
		} catch (error) {
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				await callbacks.onUnauthenticated(request);
				return;
			}
			if (this.isOriginCurrent(request, callbacks.getView())) {
				callbacks.setError(request.origin, request.ticketID, 'Unable to refresh this ticket. Please try again.', true);
			}
		}
	}

	private isOriginCurrent(request: AcceptMutationRequest, view: AcceptViewContext): boolean {
		return request.origin === 'row' ? this.isListCurrent(request, view) : this.isChatCurrent(request, view);
	}

	private isListCurrent(request: AcceptMutationRequest, view: AcceptViewContext): boolean {
		return request.listSequence === view.listSequence && request.listView === view.listView;
	}

	private isChatCurrent(request: AcceptMutationRequest, view: AcceptViewContext): boolean {
		return (
			request.chatGeneration === view.chatGeneration &&
			view.chatOpen &&
			view.selectedTicketID === request.ticketID
		);
	}
}
