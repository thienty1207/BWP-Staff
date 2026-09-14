<svelte:options runes={true} />

<script lang="ts">
	import type { TicketDetailState } from '$lib/client/tickets/detail-state';
	import type { TicketIdentity, TicketSummary } from '$lib/client/tickets/model';

	type TicketTimestamp = {
		date: string;
		time: string;
	};

	type Props = {
		state: TicketDetailState;
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

	function departmentLabel(department: TicketSummary['department']): string {
		return department.code ? `${department.name} (${department.code})` : department.name;
	}

	function locationLabel(location: TicketSummary['location']): string {
		if (!location) {
			return '—';
		}
		return location.code ? `${location.name} (${location.code})` : location.name;
	}

	function assignmentLabel(values: Array<{ name?: string; full_name?: string }>): string {
		if (values.length === 0) {
			return '—';
		}
		return values.map((value) => value.name ?? value.full_name ?? '—').join(', ');
	}

	function descriptionLabel(description: string | null): string {
		return description?.trim() || '—';
	}

	function timestampDate(value: string | null): string {
		return formatTimestamp(value)?.date ?? '—';
	}

	function timestampTime(value: string | null): string | null {
		return formatTimestamp(value)?.time ?? null;
	}
</script>

<section class="ticket-detail" aria-labelledby="ticket-detail-heading">
	<header class="ticket-detail-header">
		<div>
			<p class="eyebrow">Tickets</p>
			<h2 id="ticket-detail-heading">Ticket Detail</h2>
		</div>
		<div class="ticket-detail-actions">
			<button class="secondary-button ticket-detail-close" type="button" onclick={onClose}>Close</button>
			<button class="secondary-button ticket-detail-back" type="button" onclick={onBack}>Back to Tickets</button>
		</div>
	</header>

	{#if state.status === 'loading'}
		<div class="ticket-detail-status" role="status" aria-live="polite">
			<p>Loading ticket…</p>
		</div>
	{:else if state.status === 'error' || state.status === 'not_found'}
		<div class="ticket-detail-status" role="alert" aria-live="assertive">
			<p>{state.status === 'not_found' ? 'Ticket not found.' : state.errorMessage}</p>
			{#if state.status === 'error'}
				<button class="secondary-button" type="button" onclick={onRetry}>Retry</button>
			{/if}
		</div>
	{:else if state.ticket}
		<div class="ticket-detail-content">
			<section class="ticket-detail-section" aria-labelledby="ticket-request-heading">
				<div class="ticket-detail-section-heading">
					<div>
						<p class="eyebrow">Request</p>
						<h3 id="ticket-request-heading">{state.ticket.title}</h3>
					</div>
					<span class:closed={state.ticket.status === 'closed'} class="status-badge">{statusLabel(state.ticket.status)}</span>
				</div>
				<div class="ticket-detail-title-row">
					<span class="ticket-detail-label">Title</span>
					<h4 class:ticket-title-priority={state.ticket.priority} class="ticket-detail-title">{state.ticket.title}</h4>
				</div>
				<div class="ticket-detail-field">
					<span class="ticket-detail-label">Description</span>
					<p class="ticket-detail-description">{descriptionLabel(state.ticket.description)}</p>
				</div>
			</section>

			<section class="ticket-detail-section" aria-labelledby="ticket-request-information-heading">
				<h3 id="ticket-request-information-heading">Request information</h3>
				<dl class="ticket-detail-list">
					<div><dt>Request Department</dt><dd>{departmentLabel(state.ticket.department)}</dd></div>
					<div><dt>Location</dt><dd>{locationLabel(state.ticket.location)}</dd></div>
					<div><dt>Requester</dt><dd>{identityLabel(state.ticket.requester)}</dd></div>
					<div><dt>Owner</dt><dd>{identityLabel(state.ticket.accepted_by)}</dd></div>
				</dl>
			</section>

			<section class="ticket-detail-section" aria-labelledby="ticket-assignment-heading">
				<h3 id="ticket-assignment-heading">Assignment</h3>
				<dl class="ticket-detail-list">
					<div><dt>Assigned Departments</dt><dd>{assignmentLabel(state.ticket.assigned_departments)}</dd></div>
					<div><dt>Assigned Users</dt><dd>{assignmentLabel(state.ticket.assigned_users)}</dd></div>
				</dl>
			</section>

			<section class="ticket-detail-section" aria-labelledby="ticket-timing-heading">
				<h3 id="ticket-timing-heading">Timing</h3>
				<dl class="ticket-detail-list">
					<div>
						<dt>Created</dt>
						<dd>
							<time class="ticket-timestamp" datetime={state.ticket.created_at}>
								<span class="ticket-timestamp-date">{timestampDate(state.ticket.created_at)}</span>
								{#if timestampTime(state.ticket.created_at)}<span class="ticket-timestamp-time">{timestampTime(state.ticket.created_at)}</span>{/if}
							</time>
						</dd>
					</div>
					<div>
						<dt>Due</dt>
						<dd>
							{#if state.ticket.due_at}
								<time class="ticket-timestamp" datetime={state.ticket.due_at}>
									<span class="ticket-timestamp-date">{timestampDate(state.ticket.due_at)}</span>
									{#if timestampTime(state.ticket.due_at)}<span class="ticket-timestamp-time">{timestampTime(state.ticket.due_at)}</span>{/if}
								</time>
							{:else}—{/if}
						</dd>
					</div>
					<div>
						<dt>Accepted</dt>
						<dd>
							{#if state.ticket.accepted_at}
								<time class="ticket-timestamp" datetime={state.ticket.accepted_at}>
									<span class="ticket-timestamp-date">{timestampDate(state.ticket.accepted_at)}</span>
									{#if timestampTime(state.ticket.accepted_at)}<span class="ticket-timestamp-time">{timestampTime(state.ticket.accepted_at)}</span>{/if}
								</time>
							{:else}—{/if}
						</dd>
					</div>
					<div>
						<dt>Closed</dt>
						<dd>
							{#if state.ticket.closed_at}
								<time class="ticket-timestamp" datetime={state.ticket.closed_at}>
									<span class="ticket-timestamp-date">{timestampDate(state.ticket.closed_at)}</span>
									{#if timestampTime(state.ticket.closed_at)}<span class="ticket-timestamp-time">{timestampTime(state.ticket.closed_at)}</span>{/if}
								</time>
							{:else}—{/if}
						</dd>
					</div>
					<div>
						<dt>Updated</dt>
						<dd>
							<time class="ticket-timestamp" datetime={state.ticket.updated_at}>
								<span class="ticket-timestamp-date">{timestampDate(state.ticket.updated_at)}</span>
								{#if timestampTime(state.ticket.updated_at)}<span class="ticket-timestamp-time">{timestampTime(state.ticket.updated_at)}</span>{/if}
							</time>
						</dd>
					</div>
				</dl>
			</section>
		</div>
	{/if}
</section>
