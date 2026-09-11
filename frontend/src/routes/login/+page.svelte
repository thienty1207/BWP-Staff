<svelte:options runes={true} />

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getCurrentUser, login } from '$lib/client/auth/api';
	import { AuthApiError } from '$lib/client/auth/model';

	type LoginState = 'checking' | 'form' | 'retry' | 'submitting';

	const rememberedUsernameKey = 'bwp-remembered-username';

	let phase: LoginState = $state('checking');
	let username = $state('');
	let password = $state('');
	let rememberMe = $state(false);
	let showPassword = $state(false);
	let errorMessage = $state('');

	onMount(() => {
		loadRememberedUsername();
		void checkExistingSession();
	});

	function loadRememberedUsername() {
		try {
			const rememberedUsername = window.localStorage.getItem(rememberedUsernameKey);
			if (rememberedUsername) {
				username = rememberedUsername;
				rememberMe = true;
			}
		} catch {
			// Remembering a username is optional and must not block login.
		}
	}

	function clearRememberedUsername() {
		try {
			window.localStorage.removeItem(rememberedUsernameKey);
		} catch {
			// Remembering a username is optional and must not block login.
		}
	}

	function persistRememberedUsername(value: string) {
		try {
			if (rememberMe) {
				window.localStorage.setItem(rememberedUsernameKey, value);
				return;
			}

			window.localStorage.removeItem(rememberedUsernameKey);
		} catch {
			// Remembering a username is optional and must not block login.
		}
	}

	function handleRememberMeChange(event: Event) {
		rememberMe = (event.currentTarget as HTMLInputElement).checked;
		if (!rememberMe) {
			clearRememberedUsername();
		}
	}

	function togglePasswordVisibility() {
		showPassword = !showPassword;
	}

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
			persistRememberedUsername(trimmedUsername);
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
	<div class="login-content">
		{#if phase === 'checking'}
			<section class="login-status" role="status" aria-live="polite">
				<h1>Checking your session</h1>
				<p class="muted-copy">Please wait while we confirm your access.</p>
			</section>
		{:else if phase === 'retry'}
			<section class="login-status" aria-labelledby="login-check-error-title">
				<p class="login-kicker">BWP SONASEA STAFF</p>
				<h1 id="login-check-error-title">Unable to verify session</h1>
				<p class="muted-copy">The service could not confirm your current session.</p>
				<p class="error-message" role="alert" aria-live="assertive">{errorMessage}</p>
				<button class="secondary-button" type="button" onclick={() => void checkExistingSession()}>Retry</button>
			</section>
		{:else}
			<section class="login-panel" aria-labelledby="login-title">
				<div class="login-heading">
					<p class="login-kicker">BWP SONASEA STAFF</p>
					<h1 id="login-title">Sign in</h1>
					<p class="login-supporting">Access your BWP SonaSea account.</p>
				</div>

				<form class="login-form" onsubmit={handleSubmit} aria-busy={phase === 'submitting'}>
					<div class="login-field">
						<label for="username">Username</label>
						<div class="input-shell">
							<span class="input-icon" aria-hidden="true">
								<svg viewBox="0 0 24 24" focusable="false">
									<circle cx="12" cy="8" r="3.25"></circle>
									<path d="M5 20c.8-3.1 3.1-4.75 7-4.75s6.2 1.65 7 4.75"></path>
								</svg>
							</span>
							<input
								id="username"
								name="username"
								type="text"
								autocomplete="username"
								inputmode="text"
								placeholder="Enter your username"
								bind:value={username}
								disabled={phase === 'submitting'}
								aria-describedby={errorMessage ? 'login-error' : undefined}
							/>
						</div>
					</div>

					<div class="login-field">
						<label for="password">Password</label>
						<div class="input-shell">
							<span class="input-icon" aria-hidden="true">
								<svg viewBox="0 0 24 24" focusable="false">
									<rect x="5.5" y="10" width="13" height="10" rx="2"></rect>
									<path d="M8.5 10V7.5a3.5 3.5 0 0 1 7 0V10"></path>
								</svg>
							</span>
							<input
								id="password"
								name="password"
								type={showPassword ? 'text' : 'password'}
								autocomplete="current-password"
								placeholder="Enter your password"
								bind:value={password}
								disabled={phase === 'submitting'}
								aria-describedby={errorMessage ? 'login-error' : undefined}
							/>
							<button
								class="password-toggle"
								type="button"
								aria-label={showPassword ? 'Hide password' : 'Show password'}
								aria-pressed={showPassword}
								onclick={togglePasswordVisibility}
							>
								{#if showPassword}
									<svg viewBox="0 0 24 24" focusable="false" aria-hidden="true">
										<path d="M3 3l18 18"></path>
										<path d="M10.6 10.6a2 2 0 0 0 2.8 2.8"></path>
										<path d="M9.9 4.3A10.9 10.9 0 0 1 12 4c5 0 8.9 4 10 8a11.5 11.5 0 0 1-3.1 4.7"></path>
										<path d="M6.2 6.2C3.9 7.7 2.5 10 2 12c1.1 4 5 8 10 8 1.2 0 2.3-.2 3.3-.5"></path>
									</svg>
								{:else}
									<svg viewBox="0 0 24 24" focusable="false" aria-hidden="true">
										<path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12Z"></path>
										<circle cx="12" cy="12" r="2.5"></circle>
									</svg>
								{/if}
							</button>
						</div>
					</div>

					<div class="remember-row">
						<label class="remember-control">
							<input
								type="checkbox"
								name="remember_me"
								checked={rememberMe}
								disabled={phase === 'submitting'}
								onchange={handleRememberMeChange}
							/>
							<span>Remember me</span>
						</label>
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
