<svelte:options runes={true} />

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import NewRequestDialog from '$lib/components/NewRequestDialog.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { getCurrentUser, logout } from '$lib/client/auth/api';
	import { listTickets } from '$lib/client/tickets/api';
	import { TicketApiError, type TicketListPage, type TicketSummary, type TicketView } from '$lib/client/tickets/model';
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
			newRequestOpen = true;
		}
	}

	function closeNewRequest() {
		newRequestOpen = false;
	}

	async function handleNewRequestCreated() {
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

	function assignmentLabel(ticket: TicketSummary): string {
		const departments = ticket.assigned_departments.map((department) => department.name);
		const users = ticket.assigned_users.map((assignedUser) => assignedUser.full_name);
		const assignments = [...departments, ...users];
		return assignments.length > 0 ? assignments.join(', ') : 'Unassigned';
	}

	function statusLabel(status: TicketSummary['status']): string {
		return status[0].toUpperCase() + status.slice(1);
	}

	function formatDate(value: string | null): string {
		if (!value) {
			return '—';
		}
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) {
			return '—';
		}
		return new Intl.DateTimeFormat(undefined, {
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(date);
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

		<main class="app-main">
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
								<th scope="col">ID</th>
								<th scope="col">Request</th>
								<th scope="col">Department</th>
								<th scope="col">Location</th>
								<th scope="col">Requester</th>
								<th scope="col">Assignment</th>
								<th scope="col">Status</th>
								<th scope="col">Created</th>
								<th scope="col">Due</th>
							</tr>
						</thead>
						<tbody>
							{#each tickets as ticket (ticket.id)}
								<tr>
									<td class="ticket-id">#{ticket.id}</td>
									<td>
										<div class="request-cell">
											<strong>{ticket.title}</strong>
											{#if ticket.priority}<span class="priority-badge">Priority</span>{/if}
										</div>
									</td>
									<td>{ticket.department.name}</td>
									<td>{ticket.location?.name ?? '—'}</td>
									<td>{ticket.requester.full_name}</td>
									<td>{assignmentLabel(ticket)}</td>
									<td><span class:closed={ticket.status === 'closed'} class="status-badge">{statusLabel(ticket.status)}</span></td>
									<td>{formatDate(ticket.created_at)}</td>
									<td>{formatDate(ticket.due_at)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				<div class="mobile-ticket-cards">
					{#each tickets as ticket (ticket.id)}
						<article class="ticket-card">
							<div class="ticket-card-heading">
								<span class="ticket-id">#{ticket.id}</span>
								<span class:closed={ticket.status === 'closed'} class="status-badge">{statusLabel(ticket.status)}</span>
							</div>
							<h2>{ticket.title}</h2>
							{#if ticket.priority}<span class="priority-badge">Priority</span>{/if}
							<dl class="ticket-card-details">
								<div><dt>Department</dt><dd>{ticket.department.name}</dd></div>
								<div><dt>Location</dt><dd>{ticket.location?.name ?? '—'}</dd></div>
								<div><dt>Requester</dt><dd>{ticket.requester.full_name}</dd></div>
								<div><dt>Assignment</dt><dd>{assignmentLabel(ticket)}</dd></div>
								<div><dt>Created</dt><dd>{formatDate(ticket.created_at)}</dd></div>
								<div><dt>Due</dt><dd>{formatDate(ticket.due_at)}</dd></div>
							</dl>
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
