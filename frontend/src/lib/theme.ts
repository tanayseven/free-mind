export interface ClassList {
	add(cls: string): void;
	remove(cls: string): void;
}

export interface Storage {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
}

/** `system` follows the OS colour-scheme preference. */
export type ThemePreference = 'system' | 'light' | 'dark';

/**
 * Reads the saved theme preference, defaulting to `system`.
 */
export function loadThemePreference(storage: Storage): ThemePreference {
	const saved = storage.getItem('theme');
	return saved === 'light' || saved === 'dark' ? saved : 'system';
}

/**
 * Whether the dark theme should be shown for a preference.
 */
export function resolveDark(preference: ThemePreference, prefersDark: boolean): boolean {
	return preference === 'dark' || (preference === 'system' && prefersDark);
}

/**
 * Applies the dark theme or removes it, saving the preference to storage.
 */
export function applyTheme(
	preference: ThemePreference,
	prefersDark: boolean,
	classList: ClassList,
	storage: Storage
): void {
	if (resolveDark(preference, prefersDark)) {
		classList.add('dark');
	} else {
		classList.remove('dark');
	}
	storage.setItem('theme', preference);
}
