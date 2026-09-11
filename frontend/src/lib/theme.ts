export type Theme = 'light' | 'dark';

const themeStorageKey = 'bwp-theme';

export function getInitialTheme(): Theme {
	if (typeof window === 'undefined') {
		return 'light';
	}

	try {
		const savedTheme = window.localStorage.getItem(themeStorageKey);
		if (savedTheme === 'light' || savedTheme === 'dark') {
			return savedTheme;
		}
	} catch {
		// Fall back to the system preference when storage is unavailable.
	}

	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export function applyTheme(theme: Theme): void {
	if (typeof document !== 'undefined') {
		document.documentElement.dataset.theme = theme;
	}
}

export function saveTheme(theme: Theme): void {
	try {
		window.localStorage.setItem(themeStorageKey, theme);
	} catch {
		// Theme preference is optional and must not block the application.
	}
}
