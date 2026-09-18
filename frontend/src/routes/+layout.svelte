<script lang="ts">
	import { onMount } from 'svelte';
	import { applyTheme, getInitialTheme } from '$lib/theme';
	import '$lib/styles/app.css';

	let { children } = $props();

	onMount(() => {
		applyTheme(getInitialTheme());
	});
</script>

<svelte:head>
	<link rel="icon" href="/images/hotel-login-icon.svg" />
</svelte:head>

{@render children()}
