<svelte:options runes={true} />

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { getCurrentUser, login } from '$lib/client/auth/api';
	import { AuthApiError } from '$lib/client/auth/model';

	type LoginState = 'checking' | 'form' | 'retry' | 'submitting';

	let phase: LoginState = $state('checking');
	let username = $state('');
	let password = $state('');
	let errorMessage = $state('');

	onMount(() => {
		void checkExistingSession();
	});

	async function checkExistingSession() {
		phase = 'checking';
		errorMessage = '';

		try {
			await getCurrentUser();
			await goto('/', { replaceState: true });
		} catch (error) {
			if (error instanceof AuthApiError && error.kind === 'unauthenticated') {
				phase = 'form';
				return;
			}

			phase = 'retry';
			errorMessage = 'Unable to verify session. Please try again.';
		}
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (phase === 'checking' || phase === 'submitting') {
			return;
		}

		const trimmedUsername = username.trim();
		username = trimmedUsername;
		if (!trimmedUsername || !password) {
			phase = 'form';
			errorMessage = 'Please enter your username and password.';
			return;
		}

		phase = 'submitting';
		errorMessage = '';

		try {
			await login(trimmedUsername, password);
			password = '';
			await goto('/');
		} catch (error) {
			phase = 'form';
			errorMessage = loginErrorMessage(error);
		}
	}

	function loginErrorMessage(error: unknown): string {
		if (error instanceof AuthApiError && error.kind === 'invalid_credentials') {
			return 'Invalid username or password.';
		}
		if (error instanceof AuthApiError && error.kind === 'invalid_input') {
			return 'Please enter a valid username and password.';
		}
		return 'The service is temporarily unavailable. Please try again.';
	}
</script>

<svelte:head>
	<title>BWP SonaSea — Login</title>
	<meta name="description" content="BWP SonaSea sign-in" />
</svelte:head>

<main class="page-shell login-shell">
	<div class="auth-layout">
		<header class="topbar">
			<div class="brand">
				<img class="brand-logo" src="/images/bwp-logo.png" alt="BWP SonaSea" />
			</div>
			<ThemeToggle />
		</header>

		{#if phase === 'checking'}
			<section class="status-panel" role="status" aria-live="polite">
				<h1>Checking your session</h1>
				<p class="muted-copy">Please wait while we confirm your access.</p>
			</section>
		{:else if phase === 'retry'}
			<section class="status-panel" aria-labelledby="login-check-error-title">
				<p class="eyebrow">Connection issue</p>
				<h1 id="login-check-error-title">Unable to verify session</h1>
				<p class="muted-copy">The service could not confirm your current session.</p>
				<p class="error-message" role="alert" aria-live="assertive">{errorMessage}</p>
				<button class="secondary-button" type="button" onclick={() => void checkExistingSession()}>Retry</button>
			</section>
		{:else}
			<section class="auth-card" aria-labelledby="login-title">
				<h1 id="login-title" class="login-title">BWP SonaSea Staff</h1>

				<form class="form-stack" onsubmit={handleSubmit} aria-busy={phase === 'submitting'}>
					<div class="field">
						<label for="username">Username</label>
						<input
							id="username"
							name="username"
							type="text"
							autocomplete="username"
							inputmode="text"
							bind:value={username}
							disabled={phase === 'submitting'}
							aria-describedby={errorMessage ? 'login-error' : undefined}
						/>
					</div>

					<div class="field">
						<label for="password">Password</label>
						<input
							id="password"
							name="password"
							type="password"
							autocomplete="current-password"
							bind:value={password}
							disabled={phase === 'submitting'}
							aria-describedby={errorMessage ? 'login-error' : undefined}
						/>
					</div>

					<button class="primary-button" type="submit" disabled={phase === 'submitting'}>
						{phase === 'submitting' ? 'Signing in…' : 'Sign in'}
					</button>
				</form>

				{#if errorMessage}
					<p id="login-error" class="error-message" role="alert" aria-live="assertive">{errorMessage}</p>
				{/if}
			</section>
		{/if}
	</div>
</main>
