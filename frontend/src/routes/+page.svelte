<svelte:options runes={true} />

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import NewRequestDialog from '$lib/components/NewRequestDialog.svelte';
	import TicketChat from '$lib/components/TicketChat.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { getCurrentUser, logout } from '$lib/client/auth/api';
	import { acceptTicket, getTicket, listTickets } from '$lib/client/tickets/api';
	import { TicketChatStateMachine, type TicketChatRequest, type TicketChatState } from '$lib/client/ticket-chat-state';
	import { TicketApiError, type TicketIdentity, type TicketListPage, type TicketSummary, type TicketView } from '$lib/client/tickets/model';
	import type { AuthenticatedUser } from '$lib/client/auth/model';
	import { AuthApiError } from '$lib/client/auth/model';

	type RootState = 'checking' | 'authenticated' | 'retry';
	type TicketState = 'loading' | 'ready' | 'error';

	const pageSize = 50;
	const emptyPage: TicketListPage = {
		has_more: false,
		next_before_created_at: null,
		next_before_id: null
	};

	let phase: RootState = $state('checking');
	let user: AuthenticatedUser | null = $state(null);
	let activeView: TicketView = $state('open');
	let ticketState: TicketState = $state('loading');
	let tickets: TicketSummary[] = $state([]);
	let page: TicketListPage = $state(emptyPage);
	let errorMessage = $state('');
	let ticketErrorMessage = $state('');
	let loadMoreErrorMessage = $state('');
	let logoutInFlight = $state(false);
	let ticketRequestInFlight = $state(false);
	let drawerOpen = $state(false);
	let newRequestOpen = $state(false);
	let ticketRequestSequence = 0;
	let acceptInFlightIDs: number[] = $state([]);
	let acceptErrorTicketID: number | null = $state(null);
	let acceptErrorMessage = $state('');
	let acceptRetryableTicketID: number | null = $state(null);
	let chatViewGeneration = 0;
	const ticketChatMachine = new TicketChatStateMachine();
	let chatState: TicketChatState = $state(ticketChatMachine.state);
	let selectedAcceptErrorMessage = $derived(
		acceptErrorTicketID === chatState.selectedTicketID ? acceptErrorMessage : ''
	);
	let selectedAcceptRetryable = $derived(
		acceptRetryableTicketID !== null && acceptRetryableTicketID === chatState.selectedTicketID
	);

	onMount(() => {
		void verifySession();
	});

	async function verifySession() {
		if (logoutInFlight) {
			return;
		}

		phase = 'checking';
		user = null;
		errorMessage = '';
		ticketErrorMessage = '';
		ticketState = 'loading';
		tickets = [];
		page = emptyPage;
		closeTicketChat();

		try {
			user = await getCurrentUser();
			phase = 'authenticated';
			await loadFirstPage('open');
		} catch (error) {
			user = null;
			if (error instanceof AuthApiError && error.kind === 'unauthenticated') {
				await goto('/login', { replaceState: true });
				return;
			}

			phase = 'retry';
			errorMessage = 'Unable to verify session. Please try again.';
		}
	}

	async function loadFirstPage(view: TicketView) {
		if (ticketRequestInFlight) {
			return;
		}

		const requestSequence = ++ticketRequestSequence;
		ticketRequestInFlight = true;
		ticketState = 'loading';
		ticketErrorMessage = '';
		loadMoreErrorMessage = '';
		tickets = [];
		page = emptyPage;

		try {
			const result = await listTickets(view, { limit: pageSize });
			if (requestSequence !== ticketRequestSequence) {
				return;
			}
			tickets = result.tickets;
			page = result.page;
			ticketState = 'ready';
		} catch (error) {
			if (requestSequence !== ticketRequestSequence) {
				return;
			}
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				await goto('/login', { replaceState: true });
				return;
			}
			ticketState = 'error';
			ticketErrorMessage = 'Unable to load tickets. Please try again.';
		} finally {
			if (requestSequence === ticketRequestSequence) {
				ticketRequestInFlight = false;
			}
		}
	}

	async function loadMore() {
		if (
			ticketRequestInFlight ||
			!page.has_more ||
			page.next_before_created_at === null ||
			page.next_before_id === null
		) {
			return;
		}

		const requestSequence = ++ticketRequestSequence;
		ticketRequestInFlight = true;

		try {
			const result = await listTickets(activeView, {
				limit: pageSize,
				cursor: {
					before_created_at: page.next_before_created_at,
					before_id: page.next_before_id
				}
			});
			if (requestSequence !== ticketRequestSequence) {
				return;
			}
			tickets = [...tickets, ...result.tickets];
			page = result.page;
			loadMoreErrorMessage = '';
			ticketState = 'ready';
		} catch (error) {
			if (requestSequence !== ticketRequestSequence) {
				return;
			}
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				await goto('/login', { replaceState: true });
				return;
			}
			loadMoreErrorMessage = 'Unable to load more tickets. Please try again.';
		} finally {
			if (requestSequence === ticketRequestSequence) {
				ticketRequestInFlight = false;
			}
		}
	}

	function switchView(view: TicketView) {
		if (view === activeView || ticketRequestInFlight) {
			return;
		}
		closeTicketChat();
		activeView = view;
		drawerOpen = false;
		void loadFirstPage(view);
	}

	function retryTicketLoad() {
		void loadFirstPage(activeView);
	}

	function retryLoadMore() {
		void loadMore();
	}

	function openNewRequest() {
		if (!ticketRequestInFlight) {
			closeTicketChat();
			newRequestOpen = true;
		}
	}

	function closeNewRequest() {
		newRequestOpen = false;
	}

	async function handleNewRequestCreated() {
		closeTicketChat();
		activeView = 'open';
		drawerOpen = false;
		await loadFirstPage('open');
	}

	async function handleLogout() {
		if (logoutInFlight || !user) {
			return;
		}

		logoutInFlight = true;
		errorMessage = '';
		closeTicketChat();

		try {
			await logout();
			user = null;
			await goto('/login', { replaceState: true });
		} catch {
			logoutInFlight = false;
			errorMessage = 'Unable to sign out right now. Please try again.';
		}
	}

	function closeDrawer() {
		drawerOpen = false;
	}

	function syncTicketChatState() {
		chatState = ticketChatMachine.state;
	}

	type AcceptMutationRequest = {
		ticketID: number;
		chatGeneration: number;
	};

	function isAcceptInFlight(ticketID: number | null): boolean {
		return ticketID !== null && acceptInFlightIDs.includes(ticketID);
	}

	function setAcceptInFlight(ticketID: number, inFlight: boolean) {
		if (inFlight) {
			if (!acceptInFlightIDs.includes(ticketID)) {
				acceptInFlightIDs = [...acceptInFlightIDs, ticketID];
			}
			return;
		}
		acceptInFlightIDs = acceptInFlightIDs.filter((value) => value !== ticketID);
	}

	function isCurrentAcceptChatContext(request: AcceptMutationRequest): boolean {
		return (
			request.chatGeneration === chatViewGeneration &&
			chatState.status !== 'closed' &&
			chatState.selectedTicketID === request.ticketID
		);
	}

	function patchTicketInList(updatedTicket: TicketSummary) {
		tickets = tickets.map((ticket) => (ticket.id === updatedTicket.id ? updatedTicket : ticket));
	}

	function setAcceptError(ticketID: number, message: string, retryable: boolean) {
		acceptErrorTicketID = ticketID;
		acceptErrorMessage = message;
		acceptRetryableTicketID = retryable ? ticketID : null;
	}

	function openTicketChat(ticketID: number) {
		chatViewGeneration += 1;
		acceptErrorTicketID = null;
		acceptErrorMessage = '';
		acceptRetryableTicketID = null;
		const request = ticketChatMachine.begin(ticketID);
		syncTicketChatState();
		void loadTicketChat(request);
	}

	function openTicketChatFromTitle(event: MouseEvent, ticketID: number) {
		event.stopPropagation();
		openTicketChat(ticketID);
	}

	function openTicketChatFromRowKey(event: KeyboardEvent, ticketID: number) {
		if (event.target !== event.currentTarget) {
			return;
		}
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			openTicketChat(ticketID);
		}
	}

	async function loadTicketChat(request: TicketChatRequest) {
		try {
			const ticket = await getTicket(request.ticketID);
			if (ticketChatMachine.succeed(request, ticket)) {
				syncTicketChatState();
			}
		} catch (error) {
			if (!ticketChatMachine.isCurrent(request)) {
				return;
			}
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				ticketChatMachine.close();
				syncTicketChatState();
				await goto('/login', { replaceState: true });
				return;
			}
			if (error instanceof TicketApiError && error.kind === 'not_found') {
				ticketChatMachine.fail(request, 'not_found', 'Ticket not found.');
				syncTicketChatState();
				return;
			}
			ticketChatMachine.fail(request, 'error', 'Unable to load ticket. Please try again.');
			syncTicketChatState();
		}
	}

	function closeTicketChat() {
		chatViewGeneration += 1;
		acceptErrorTicketID = null;
		acceptErrorMessage = '';
		acceptRetryableTicketID = null;
		ticketChatMachine.close();
		syncTicketChatState();
	}

	function retryTicketChat() {
		if (chatState.selectedTicketID === null || chatState.status === 'loading') {
			return;
		}
		chatViewGeneration += 1;
		const request = ticketChatMachine.begin(chatState.selectedTicketID);
		syncTicketChatState();
		void loadTicketChat(request);
	}

	async function refreshTicketAfterAcceptConflict(request: AcceptMutationRequest) {
		try {
			const refreshedTicket = await getTicket(request.ticketID);
			patchTicketInList(refreshedTicket);
			if (isCurrentAcceptChatContext(request) && ticketChatMachine.updateSelected(request.ticketID, refreshedTicket)) {
				syncTicketChatState();
			}
		} catch (error) {
			if (
				isCurrentAcceptChatContext(request) &&
				error instanceof TicketApiError &&
				error.kind === 'unauthenticated'
			) {
				closeTicketChat();
				await goto('/login', { replaceState: true });
			}
		}
	}

	async function handleAcceptTicket() {
		const ticketID = chatState.selectedTicketID;
		const currentTicket = chatState.ticket;
		if (
			ticketID === null ||
			currentTicket === null ||
			currentTicket.status !== 'pending' ||
			isAcceptInFlight(ticketID)
		) {
			return;
		}

		const request: AcceptMutationRequest = { ticketID, chatGeneration: chatViewGeneration };
		setAcceptInFlight(ticketID, true);
		setAcceptError(ticketID, '', false);

		try {
			const updatedTicket = await acceptTicket(ticketID);
			patchTicketInList(updatedTicket);
			if (isCurrentAcceptChatContext(request) && ticketChatMachine.updateSelected(ticketID, updatedTicket)) {
				syncTicketChatState();
				setAcceptError(ticketID, '', false);
			}
		} catch (error) {
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				if (isCurrentAcceptChatContext(request)) {
					closeTicketChat();
					await goto('/login', { replaceState: true });
				}
				return;
			}

			if (error instanceof TicketApiError && (error.kind === 'already_accepted' || error.kind === 'closed')) {
				if (isCurrentAcceptChatContext(request)) {
					setAcceptError(
						ticketID,
						error.kind === 'already_accepted'
							? 'This ticket was already accepted by another staff member.'
							: 'Closed tickets cannot be accepted.',
						false
					);
				}
				await refreshTicketAfterAcceptConflict(request);
				return;
			}

			if (!isCurrentAcceptChatContext(request)) {
				return;
			}
			if (error instanceof TicketApiError && error.kind === 'not_found') {
				setAcceptError(ticketID, 'Ticket not found. It may have been removed.', false);
				return;
			}
			setAcceptError(ticketID, 'Unable to accept this ticket. Please try again.', true);
		} finally {
			setAcceptInFlight(ticketID, false);
		}
	}

	function initials(fullName: string): string {
		const parts = fullName.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) {
			return '?';
		}
		return parts
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase() ?? '')
			.join('');
	}

	function identityLabel(identity: TicketIdentity): string {
		return identity.department_code ? `${identity.full_name} (${identity.department_code})` : identity.full_name;
	}

	function descriptionLabel(description: string | null): string {
		return description?.trim() || '—';
	}

	function statusLabel(status: TicketSummary['status']): string {
		return status[0].toUpperCase() + status.slice(1);
	}

	type TicketTimestamp = {
		date: string;
		time: string;
	};

	function formatTimestamp(value: string | null): TicketTimestamp | null {
		if (!value) {
			return null;
		}
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) {
			return null;
		}
		const hours = date.getHours();
		return {
			date: `${String(date.getDate()).padStart(2, '0')}-${date.toLocaleString('en-US', { month: 'short' })}-${date.getFullYear()}`,
			time: `${String(hours % 12 || 12).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')} ${hours >= 12 ? 'PM' : 'AM'}`
		};
	}

	function timestampDate(value: string | null): string {
		return formatTimestamp(value)?.date ?? '—';
	}

	function timestampTime(value: string | null): string | null {
		return formatTimestamp(value)?.time ?? null;
	}
</script>

<svelte:head>
	<title>Tickets | BWP SonaSea</title>
	<meta name="description" content="BWP SonaSea staff tickets" />
</svelte:head>

{#if phase === 'checking'}
	<main class="page-shell">
		<section class="status-panel" role="status" aria-live="polite">
			<p class="eyebrow">Secure workspace</p>
			<h1>Verifying your session</h1>
			<p class="muted-copy">Please wait while we confirm your access.</p>
		</section>
	</main>
{:else if phase === 'retry'}
	<main class="page-shell">
		<section class="status-panel" aria-labelledby="session-error-title">
			<p class="eyebrow">Connection issue</p>
			<h1 id="session-error-title">Unable to verify session</h1>
			<p class="muted-copy">The service could not confirm your session.</p>
			<p class="error-message" role="alert" aria-live="assertive">{errorMessage}</p>
			<button class="secondary-button" type="button" onclick={() => void verifySession()}>Retry</button>
		</section>
	</main>
{:else if user}
	<div class="app-shell">
		{#if drawerOpen}
			<button class="drawer-backdrop" type="button" aria-label="Close navigation" onclick={closeDrawer}></button>
		{/if}

		<aside id="primary-navigation" class:drawer-open={drawerOpen} class="app-sidebar" aria-label="Primary navigation">
			<div class="sidebar-brand">
				<img class="sidebar-logo" src="/images/bwp-logo.png" alt="BWP SonaSea" />
			</div>

			<nav class="sidebar-nav" aria-label="Staff portal sections">
				<button class="nav-item active" type="button" aria-current="page" onclick={() => closeDrawer()}>
					<span aria-hidden="true">▦</span>
					Tickets
				</button>
				<button class="nav-item" type="button" disabled>
					<span aria-hidden="true">▤</span>
					Report
				</button>
				<button class="nav-item" type="button" disabled>
					<span aria-hidden="true">⚙</span>
					Settings
				</button>
			</nav>

			<div class="sidebar-footer">
				<div class="sidebar-profile">
					<div class="profile-avatar">
						{#if user.avatar_url}
							<img src={user.avatar_url} alt="" />
						{:else}
							<span aria-hidden="true">{initials(user.full_name)}</span>
						{/if}
					</div>
					<div class="profile-copy">
						<strong>{user.full_name}</strong>
						<span>{user.department.name}</span>
					</div>
				</div>
				<div class="sidebar-actions">
					<ThemeToggle />
					<button class="danger-button sidebar-logout" type="button" disabled={logoutInFlight} onclick={() => void handleLogout()}>
						{logoutInFlight ? 'Signing out…' : 'Logout'}
					</button>
				</div>
			</div>
		</aside>

		<main class:chat-view={chatState.status !== 'closed'} class="app-main">
			<header class="mobile-header">
				<button
					class="menu-button"
					type="button"
					aria-controls="primary-navigation"
					aria-expanded={drawerOpen}
					onclick={() => (drawerOpen = true)}
				>
					<span aria-hidden="true">☰</span>
					<span>Tickets</span>
				</button>
			</header>
			{#if errorMessage}
				<p class="error-message shell-error" role="alert" aria-live="assertive">{errorMessage}</p>
			{/if}

			<div class="tickets-page-header">
				<div>
					<p class="eyebrow">Staff workspace</p>
					<h1>Tickets</h1>
				</div>
				<button class="primary-button new-request-button" type="button" disabled={ticketRequestInFlight} onclick={openNewRequest}>New Request</button>
			</div>

			<div class="ticket-tabs" role="tablist" aria-label="Ticket status">
				<button
					class:active={activeView === 'open'}
					class="tab-button"
					type="button"
					role="tab"
					aria-selected={activeView === 'open'}
					disabled={ticketRequestInFlight}
					onclick={() => switchView('open')}
				>
					Open
				</button>
				<button
					class:active={activeView === 'closed'}
					class="tab-button"
					type="button"
					role="tab"
					aria-selected={activeView === 'closed'}
					disabled={ticketRequestInFlight}
					onclick={() => switchView('closed')}
				>
					Closed
				</button>
			</div>

			{#if ticketState === 'loading' && tickets.length === 0}
				<section class="ticket-status-panel" role="status" aria-live="polite">
					<p>Loading tickets…</p>
				</section>
			{:else if ticketState === 'error' && tickets.length === 0}
				<section class="ticket-status-panel" aria-labelledby="ticket-error-title">
					<h2 id="ticket-error-title">Unable to load tickets</h2>
					<p class="muted-copy">The ticket service could not be reached.</p>
					<p class="error-message" role="alert" aria-live="assertive">{ticketErrorMessage}</p>
					<button class="secondary-button" type="button" onclick={retryTicketLoad}>Retry</button>
				</section>
		{:else if tickets.length === 0}
			<section class="ticket-status-panel empty-ticket-panel" role="status" aria-live="polite">
				<p>{activeView === 'open' ? 'No open tickets.' : 'No closed tickets.'}</p>
			</section>
		{:else}
			<div class:chat-open={chatState.status !== 'closed'} class="tickets-workspace">
				<section class="ticket-list-region" aria-label="Ticket list">
					{#if loadMoreErrorMessage}
						<div class="inline-ticket-error" role="alert" aria-live="assertive">
							<span>{loadMoreErrorMessage}</span>
							<button class="secondary-button" type="button" disabled={ticketRequestInFlight} onclick={retryLoadMore}>
								{ticketRequestInFlight ? 'Retrying…' : 'Retry'}
							</button>
						</div>
					{/if}

					<div class="desktop-ticket-table">
						<table>
							<thead>
								<tr>
									<th scope="col">Requester</th>
									<th scope="col">Location</th>
									<th scope="col">Title</th>
									<th scope="col">Description</th>
									<th scope="col">Status</th>
									<th scope="col">Owner</th>
									<th scope="col">Created On</th>
									<th scope="col">Due Date</th>
								</tr>
							</thead>
							<tbody>
								{#each tickets as ticket (ticket.id)}
									<tr
										class="ticket-row-clickable"
										class:ticket-row-selected={chatState.selectedTicketID === ticket.id}
										onclick={() => openTicketChat(ticket.id)}
										onkeydown={(event) => openTicketChatFromRowKey(event, ticket.id)}
									>
										<td>{identityLabel(ticket.requester)}</td>
										<td>{ticket.location?.name ?? '—'}</td>
										<td class="ticket-title-cell">
											<button
												class="ticket-title-button"
												type="button"
												onclick={(event) => openTicketChatFromTitle(event, ticket.id)}
												onkeydown={(event) => event.stopPropagation()}
											>
												<span class:ticket-title-priority={ticket.priority}>{ticket.title}</span>
											</button>
										</td>
										<td><span class="ticket-description">{descriptionLabel(ticket.description)}</span></td>
										<td>
											<span
												class:accepted={ticket.status === 'accepted'}
												class:closed={ticket.status === 'closed'}
												class="status-badge"
											>
												{statusLabel(ticket.status)}
											</span>
										</td>
										<td>{ticket.accepted_by ? identityLabel(ticket.accepted_by) : '—'}</td>
										<td>
											<time class="ticket-timestamp" datetime={ticket.created_at}>
												<span class="ticket-timestamp-date">{timestampDate(ticket.created_at)}</span>
												{#if timestampTime(ticket.created_at)}<span class="ticket-timestamp-time">{timestampTime(ticket.created_at)}</span>{/if}
											</time>
										</td>
										<td>
											<div class="ticket-due-cell">
												<time class="ticket-timestamp" datetime={ticket.due_at ?? undefined}>
													<span class="ticket-timestamp-date">{timestampDate(ticket.due_at)}</span>
													{#if timestampTime(ticket.due_at)}<span class="ticket-timestamp-time">{timestampTime(ticket.due_at)}</span>{/if}
												</time>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

						<div class="mobile-ticket-cards">
							{#each tickets as ticket (ticket.id)}
								<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
								<article
									class="ticket-card ticket-card-clickable"
									class:ticket-card-selected={chatState.selectedTicketID === ticket.id}
									onclick={() => openTicketChat(ticket.id)}
									onkeydown={(event) => openTicketChatFromRowKey(event, ticket.id)}
								>
									<div class="ticket-card-heading">
										<div class="ticket-card-title">
											<span class="ticket-card-icon" aria-hidden="true">◈</span>
											<h2>
												<button
													class="ticket-title-button"
													type="button"
													onclick={(event) => openTicketChatFromTitle(event, ticket.id)}
													onkeydown={(event) => event.stopPropagation()}
												>
													<span class:ticket-title-priority={ticket.priority}>{ticket.title}</span>
												</button>
											</h2>
										</div>
										<span
											class:accepted={ticket.status === 'accepted'}
											class:closed={ticket.status === 'closed'}
											class="status-badge"
										>
											{statusLabel(ticket.status)}
										</span>
									</div>
									<div class="ticket-card-meta">
										<div><span>Location</span><strong>{ticket.location?.name ?? '—'}</strong></div>
										{#if ticket.accepted_by}<div><span>Owner</span><strong>{identityLabel(ticket.accepted_by)}</strong></div>{/if}
										<div><span>Requester</span><strong>{identityLabel(ticket.requester)}</strong></div>
										<div>
											<span>Created</span>
											<strong class="ticket-timestamp">
												<span class="ticket-timestamp-date">{timestampDate(ticket.created_at)}</span>
												{#if timestampTime(ticket.created_at)}<span class="ticket-timestamp-time">{timestampTime(ticket.created_at)}</span>{/if}
											</strong>
										</div>
									</div>
									{#if ticket.description?.trim()}<p class="ticket-card-description">{ticket.description.trim()}</p>{/if}
								</article>
							{/each}
						</div>

					{#if page.has_more && !loadMoreErrorMessage}
						<div class="load-more-row">
							<button class="secondary-button" type="button" disabled={ticketRequestInFlight} onclick={() => void loadMore()}>
								{ticketRequestInFlight ? 'Loading…' : 'Load more'}
							</button>
						</div>
					{/if}
				</section>

				{#if chatState.status !== 'closed'}
					<TicketChat
						state={chatState}
						formatTimestamp={formatTimestamp}
						onClose={closeTicketChat}
						onBack={closeTicketChat}
						onRetry={retryTicketChat}
						onAccept={handleAcceptTicket}
						onRetryAccept={handleAcceptTicket}
						acceptInFlight={isAcceptInFlight(chatState.selectedTicketID)}
						acceptErrorMessage={selectedAcceptErrorMessage}
						acceptRetryable={selectedAcceptRetryable}
					/>
				{/if}
			</div>
		{/if}
		</main>

		<NewRequestDialog
			open={newRequestOpen}
			disabled={ticketRequestInFlight}
			onClose={closeNewRequest}
			onCreated={handleNewRequestCreated}
		/>
	</div>
{/if}
