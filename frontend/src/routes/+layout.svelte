<script lang="ts">
	import { onMount } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import { applyTheme, getInitialTheme } from '$lib/theme';
	import '$lib/styles/app.css';

	let { children } = $props();

	onMount(() => {
		applyTheme(getInitialTheme());
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
