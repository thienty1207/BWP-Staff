<svelte:options runes={true} />

<script lang="ts">
	import { onMount } from 'svelte';
	import { applyTheme, getInitialTheme, saveTheme, type Theme } from '$lib/theme';

	let theme = $state<Theme>('light');

	onMount(() => {
		theme = getInitialTheme();
		applyTheme(theme);
	});

	function toggleTheme() {
		theme = theme === 'dark' ? 'light' : 'dark';
		applyTheme(theme);
		saveTheme(theme);
	}
</script>

<button
	type="button"
	class="theme-toggle"
	aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
	aria-pressed={theme === 'dark'}
	onclick={toggleTheme}
>
	<span aria-hidden="true">{theme === 'dark' ? '☼' : '☾'}</span>
	<span>{theme === 'dark' ? 'Light' : 'Dark'}</span>
</button>
