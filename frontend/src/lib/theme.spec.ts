import { describe, it, expect, beforeEach } from 'vitest';
import { applyTheme, loadThemePreference, resolveDark } from './theme';
import type { ClassList, Storage } from './theme';

// Minimal in-memory stubs — no DOM required
function makeClassList(): ClassList & { classes: Set<string> } {
	const classes = new Set<string>();
	return {
		classes,
		add: (cls) => classes.add(cls),
		remove: (cls) => classes.delete(cls)
	};
}

function makeStorage(
	initial: Record<string, string> = {}
): Storage & { data: Record<string, string> } {
	const data = { ...initial };
	return {
		data,
		getItem: (key) => data[key] ?? null,
		setItem: (key, value) => {
			data[key] = value;
		}
	};
}

describe('applyTheme', () => {
	let classList: ReturnType<typeof makeClassList>;
	let storage: ReturnType<typeof makeStorage>;

	beforeEach(() => {
		classList = makeClassList();
		storage = makeStorage();
	});

	it('adds "dark" class for the dark preference', () => {
		applyTheme('dark', false, classList, storage);
		expect(classList.classes.has('dark')).toBe(true);
	});

	it('removes "dark" class for the light preference', () => {
		classList.add('dark'); // start in dark mode
		applyTheme('light', true, classList, storage);
		expect(classList.classes.has('dark')).toBe(false);
	});

	it('follows the OS for the system preference', () => {
		applyTheme('system', true, classList, storage);
		expect(classList.classes.has('dark')).toBe(true);
		applyTheme('system', false, classList, storage);
		expect(classList.classes.has('dark')).toBe(false);
	});

	it('saves the preference to storage', () => {
		applyTheme('system', true, classList, storage);
		expect(storage.data['theme']).toBe('system');
		applyTheme('light', true, classList, storage);
		expect(storage.data['theme']).toBe('light');
	});
});

describe('loadThemePreference', () => {
	it('returns the saved light or dark preference', () => {
		expect(loadThemePreference(makeStorage({ theme: 'dark' }))).toBe('dark');
		expect(loadThemePreference(makeStorage({ theme: 'light' }))).toBe('light');
	});

	it('defaults to system when nothing or something unknown is saved', () => {
		expect(loadThemePreference(makeStorage())).toBe('system');
		expect(loadThemePreference(makeStorage({ theme: 'sepia' }))).toBe('system');
	});
});

describe('resolveDark', () => {
	it('explicit preferences ignore the OS', () => {
		expect(resolveDark('dark', false)).toBe(true);
		expect(resolveDark('light', true)).toBe(false);
	});

	it('system follows the OS', () => {
		expect(resolveDark('system', true)).toBe(true);
		expect(resolveDark('system', false)).toBe(false);
	});
});
