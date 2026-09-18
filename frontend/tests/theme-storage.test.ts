import { expect, test } from 'bun:test';
import { getInitialTheme, saveTheme } from '../src/lib/theme';

test('persists and restores the theme using the Hotel Staff storage key', () => {
	const values = new Map<string, string>();
	const hadWindow = Reflect.has(globalThis, 'window');
	const previousWindow = Reflect.get(globalThis, 'window');
	const windowStub = {
		localStorage: {
			getItem: (key: string) => values.get(key) ?? null,
			setItem: (key: string, value: string) => values.set(key, value)
		},
		matchMedia: () => ({ matches: false })
	};

	Object.defineProperty(globalThis, 'window', {
		configurable: true,
		value: windowStub
	});

	try {
		saveTheme('dark');

		expect(values.get('hotel-staff-theme')).toBe('dark');
		expect(getInitialTheme()).toBe('dark');
		expect([...values.keys()]).toEqual(['hotel-staff-theme']);
	} finally {
		if (hadWindow) {
			Object.defineProperty(globalThis, 'window', {
				configurable: true,
				value: previousWindow
			});
		} else {
			Reflect.deleteProperty(globalThis, 'window');
		}
	}
});
