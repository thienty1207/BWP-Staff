<svelte:options runes={true} />

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { getCurrentUser, logout } from '$lib/client/auth/api';
	import type { AuthenticatedUser } from '$lib/client/auth/model';
	import { AuthApiError } from '$lib/client/auth/model';

	type RootState = 'checking' | 'authenticated' | 'retry';

	let phase: RootState = $state('checking');
	let user: AuthenticatedUser | null = $state(null);
	let errorMessage = $state('');
	let logoutInFlight = $state(false);

	onMount(() => {
		void verifySession();
	});

	async function verifySession() {
		if (logoutInFlight) {
			return;
		}

		phase = 'checking';
		errorMessage = '';

		try {
			user = await getCurrentUser();
			phase = 'authenticated';
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
</script>

<svelte:head>
	<title>BWP SonaSea</title>
	<meta name="description" content="BWP SonaSea staff portal" />
</svelte:head>

<main class="page-shell">
	<div class="auth-layout">
		<header class="topbar">
			<div class="brand" aria-label="BWP SonaSea">
				<span class="brand-mark" aria-hidden="true">BWP</span>
				<span>BWP SonaSea</span>
			</div>
			<ThemeToggle />
		</header>

		{#if phase === 'checking'}
			<section class="status-panel" role="status" aria-live="polite">
				<p class="eyebrow">Secure workspace</p>
				<h1>Verifying your session</h1>
				<p class="muted-copy">Please wait while we confirm your access.</p>
			</section>
		{:else if phase === 'retry'}
			<section class="status-panel" aria-labelledby="session-error-title">
				<p class="eyebrow">Connection issue</p>
				<h1 id="session-error-title">Unable to verify session</h1>
				<p class="muted-copy">The service could not confirm your session.</p>
				<p class="error-message" role="alert" aria-live="assertive">{errorMessage}</p>
				<button class="secondary-button" type="button" onclick={() => void verifySession()}>Retry</button>
			</section>
		{:else if user}
			<section class="identity-card" aria-labelledby="identity-title">
				<div class="identity-heading">
					<p class="eyebrow">Authenticated workspace</p>
					<h1 id="identity-title">BWP SonaSea</h1>
					<p class="welcome">Signed in as <strong>{user.full_name}</strong></p>
				</div>

				<dl class="identity-list">
					<div>
						<dt>Department</dt>
						<dd>{user.department.name}</dd>
					</div>
					<div>
						<dt>Role</dt>
						<dd>{user.role}</dd>
					</div>
				</dl>

				<div class="card-actions">
					<button class="danger-button" type="button" disabled={logoutInFlight} onclick={() => void handleLogout()}>
						{logoutInFlight ? 'Signing out…' : 'Logout'}
					</button>
				</div>

				{#if errorMessage}
					<p class="error-message" role="alert" aria-live="assertive">{errorMessage}</p>
				{/if}
			</section>
		{/if}
	</div>
</main>
