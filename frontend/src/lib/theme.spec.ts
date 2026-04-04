import { describe, it, expect, beforeEach } from 'vitest';
import { applyTheme, detectInitialTheme } from './theme';
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

function makeStorage(initial: Record<string, string> = {}): Storage & { data: Record<string, string> } {
	const data = { ...initial };
	return {
		data,
		getItem: (key) => data[key] ?? null,
		setItem: (key, value) => { data[key] = value; }
	};
}

describe('applyTheme', () => {
	let classList: ReturnType<typeof makeClassList>;
	let storage: ReturnType<typeof makeStorage>;

	beforeEach(() => {
		classList = makeClassList();
		storage = makeStorage();
	});

	it('adds "dark" class when isDark is true', () => {
		applyTheme(true, classList, storage);
		expect(classList.classes.has('dark')).toBe(true);
	});

	it('saves "dark" to storage when isDark is true', () => {
		applyTheme(true, classList, storage);
		expect(storage.data['theme']).toBe('dark');
	});

	it('removes "dark" class when isDark is false', () => {
		classList.add('dark'); // start in dark mode
		applyTheme(false, classList, storage);
		expect(classList.classes.has('dark')).toBe(false);
	});

	it('saves "light" to storage when isDark is false', () => {
		applyTheme(false, classList, storage);
		expect(storage.data['theme']).toBe('light');
	});

	it('does not add "dark" class when isDark is false', () => {
		applyTheme(false, classList, storage);
		expect(classList.classes.has('dark')).toBe(false);
	});

	it('does not remove "dark" class when isDark is true', () => {
		applyTheme(true, classList, storage);
		expect(classList.classes.has('dark')).toBe(true);
	});
});

describe('detectInitialTheme', () => {
	it('returns true when storage has "dark"', () => {
		const storage = makeStorage({ theme: 'dark' });
		expect(detectInitialTheme(storage, false)).toBe(true);
	});

	it('returns false when storage has "light"', () => {
		const storage = makeStorage({ theme: 'light' });
		expect(detectInitialTheme(storage, true)).toBe(false);
	});

	it('returns true when no storage entry and OS prefers dark', () => {
		const storage = makeStorage();
		expect(detectInitialTheme(storage, true)).toBe(true);
	});

	it('returns false when no storage entry and OS prefers light', () => {
		const storage = makeStorage();
		expect(detectInitialTheme(storage, false)).toBe(false);
	});

	it('storage "dark" takes priority over OS light preference', () => {
		const storage = makeStorage({ theme: 'dark' });
		expect(detectInitialTheme(storage, false)).toBe(true);
	});

	it('storage "light" takes priority over OS dark preference', () => {
		const storage = makeStorage({ theme: 'light' });
		expect(detectInitialTheme(storage, true)).toBe(false);
	});
});
