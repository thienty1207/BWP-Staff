<svelte:options runes={true} />

<script lang="ts">
	import type { TicketChatState } from '$lib/client/ticket-chat-state';
	import type { TicketIdentity, TicketSummary } from '$lib/client/tickets/model';

	type TicketTimestamp = {
		date: string;
		time: string;
	};

	type Props = {
		state: TicketChatState;
		formatTimestamp: (value: string | null) => TicketTimestamp | null;
		onClose: () => void;
		onBack: () => void;
		onRetry: () => void;
	};

	let { state, formatTimestamp, onClose, onBack, onRetry }: Props = $props();

	function identityLabel(identity: TicketIdentity | null): string {
		if (!identity) {
			return '—';
		}
		return identity.department_code ? `${identity.full_name} (${identity.department_code})` : identity.full_name;
	}

	function statusLabel(status: TicketSummary['status']): string {
		return status[0].toUpperCase() + status.slice(1);
	}

	function locationLabel(location: TicketSummary['location']): string {
		return location?.name || '—';
	}

	function timestampDate(value: string | null): string {
		return formatTimestamp(value)?.date ?? '—';
	}

	function timestampTime(value: string | null): string | null {
		return formatTimestamp(value)?.time ?? null;
	}

	function activityTimestamp(value: string | null): string {
		const timestamp = formatTimestamp(value);
		return timestamp ? `${timestamp.date} ${timestamp.time}` : '—';
	}
</script>

<section class="ticket-chat" aria-labelledby="ticket-chat-heading">
	<header class="ticket-chat-header">
		<div>
			<h2 id="ticket-chat-heading">Chat</h2>
		</div>
		<div class="ticket-chat-header-actions">
			<button class="secondary-button ticket-chat-back" type="button" onclick={onBack}>Back to Tickets</button>
			<button class="secondary-button ticket-chat-close" type="button" aria-label="Close Chat" onclick={onClose}>×</button>
		</div>
	</header>

	{#if state.ticket}
		<div class="ticket-chat-summary">
			<div class="ticket-chat-summary-heading">
				<div class="ticket-chat-title-row">
					<span class="ticket-chat-icon" aria-hidden="true">◈</span>
					<h3 class="ticket-chat-title" class:ticket-title-priority={state.ticket.priority}>{state.ticket.title}</h3>
				</div>
				<span
					class:accepted={state.ticket.status === 'accepted'}
					class:closed={state.ticket.status === 'closed'}
					class="status-badge"
				>
					{statusLabel(state.ticket.status)}
				</span>
			</div>

			<div class="ticket-chat-summary-meta">
				<div class="ticket-chat-summary-line">
					<span class="ticket-chat-summary-label">Requester</span>
					<strong class="ticket-chat-summary-value">{identityLabel(state.ticket.requester)}</strong>
					{#if state.ticket.accepted_by}<span class="ticket-chat-summary-owner">by {identityLabel(state.ticket.accepted_by)}</span>{/if}
				</div>
				<div class="ticket-chat-summary-line">
					<span class="ticket-chat-summary-label">Location</span>
					<strong class="ticket-chat-summary-value">{locationLabel(state.ticket.location)}</strong>
					<time class="ticket-chat-summary-created" datetime={state.ticket.created_at}>
						<span class="ticket-chat-summary-created-label">Created</span>
						<span>{timestampDate(state.ticket.created_at)}</span>
						{#if timestampTime(state.ticket.created_at)}<span>{timestampTime(state.ticket.created_at)}</span>{/if}
					</time>
				</div>
			</div>
		</div>
	{:else if state.status === 'loading'}
		<div class="ticket-chat-summary ticket-chat-summary-loading" role="status" aria-live="polite">
			<p>Loading selected ticket…</p>
		</div>
	{/if}

	<div class="ticket-chat-tabs" role="tablist" aria-label="Ticket conversation sections">
		<button class="ticket-chat-tab active" type="button" role="tab" aria-selected="true">Chats</button>
		<button class="ticket-chat-tab" type="button" role="tab" aria-selected="false" aria-disabled="true" disabled>Checklist</button>
	</div>

	<section class="ticket-chat-conversation" role="log" aria-label="Ticket conversation" aria-live="polite" aria-busy={state.status === 'loading'}>
		{#if state.status === 'loading'}
			<p class="ticket-chat-status" role="status">Loading ticket…</p>
		{:else if state.status === 'error' || state.status === 'not_found'}
			<div class="ticket-chat-status" role="alert" aria-live="assertive">
				<p>{state.status === 'not_found' ? 'Ticket not found.' : state.errorMessage}</p>
				{#if state.status === 'error'}
					<button class="secondary-button" type="button" onclick={onRetry}>Retry</button>
				{/if}
			</div>
		{:else if state.ticket}
			<article class="ticket-chat-event ticket-chat-event-created">
				<div class="ticket-chat-event-heading">
					<span class="ticket-chat-event-icon" aria-hidden="true">◈</span>
					<div>
						<strong>{identityLabel(state.ticket.requester)}</strong>
						<p>has created a new request</p>
					</div>
					<time datetime={state.ticket.created_at}>{activityTimestamp(state.ticket.created_at)}</time>
				</div>
				<dl class="ticket-chat-event-details">
					<div><dt>Location</dt><dd>{locationLabel(state.ticket.location)}</dd></div>
					<div><dt>Title</dt><dd>{state.ticket.title}</dd></div>
				</dl>
			</article>

			{#if state.ticket.accepted_by && state.ticket.accepted_at}
				<article class="ticket-chat-event ticket-chat-event-accepted">
					<div class="ticket-chat-event-heading">
						<span class="ticket-chat-event-icon" aria-hidden="true">✓</span>
						<strong>Accepted by {identityLabel(state.ticket.accepted_by)}</strong>
						<time datetime={state.ticket.accepted_at}>{activityTimestamp(state.ticket.accepted_at)}</time>
					</div>
				</article>
			{/if}
		{/if}
	</section>

	<div class="ticket-chat-composer" aria-disabled="true">
		<button class="ticket-chat-icon-button" type="button" aria-label="Attach file" aria-disabled="true" disabled>📎</button>
		<button class="ticket-chat-icon-button" type="button" aria-label="Voice message" aria-disabled="true" disabled>🎤</button>
		<span class="ticket-chat-composer-input" role="textbox" aria-readonly="true" aria-disabled="true">Type a message</span>
		<button class="ticket-chat-icon-button" type="button" aria-label="More message options" aria-disabled="true" disabled>⋮</button>
	</div>

	<div class="ticket-chat-actions" aria-label="Ticket actions">
		<button class="secondary-button" type="button" aria-disabled="true" disabled>Accept</button>
		<button class="secondary-button" type="button" aria-disabled="true" disabled>Assign</button>
		<button class="secondary-button" type="button" aria-disabled="true" disabled>Close</button>
	</div>
</section>
