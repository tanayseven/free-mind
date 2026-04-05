export interface ClassList {
	add(cls: string): void;
	remove(cls: string): void;
}

export interface Storage {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
}

/**
 * Applies or removes the dark theme, persisting the choice to storage.
 */
export function applyTheme(isDark: boolean, classList: ClassList, storage: Storage): void {
	if (isDark) {
		classList.add('dark');
		storage.setItem('theme', 'dark');
	} else {
		classList.remove('dark');
		storage.setItem('theme', 'light');
	}
}

/**
 * Detects the initial theme from storage or the OS colour-scheme preference.
 */
export function detectInitialTheme(storage: Storage, prefersDark: boolean): boolean {
	const saved = storage.getItem('theme');
	return saved === 'dark' || (!saved && prefersDark);
}
