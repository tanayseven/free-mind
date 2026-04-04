import { page } from '@vitest/browser/context';
import { describe, expect, it, beforeEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Header from './Header.svelte';

describe('Header.svelte', () => {
	beforeEach(() => {
		localStorage.clear();
		document.documentElement.classList.remove('dark');
	});

	describe('Navigation buttons', () => {
		it('renders the Home button', async () => {
			render(Header);
			const homeButton = page.getByRole('button', { name: 'Home' });
			await expect.element(homeButton).toBeInTheDocument();
		});

		it('renders the About button', async () => {
			render(Header);
			const aboutButton = page.getByRole('button', { name: 'About' });
			await expect.element(aboutButton).toBeInTheDocument();
		});
	});

	describe('Dark/Light mode toggle', () => {
		it('starts in light mode by default when no localStorage entry exists', async () => {
			render(Header);
			const toggle = page.getByRole('switch');
			await expect.element(toggle).toHaveAttribute('aria-checked', 'false');
			expect(document.documentElement.classList.contains('dark')).toBe(false);
		});

		it('switches to dark mode when the toggle is clicked', async () => {
			render(Header);
			const toggle = page.getByRole('switch');
			await toggle.click();
			await expect.element(toggle).toHaveAttribute('aria-checked', 'true');
			expect(document.documentElement.classList.contains('dark')).toBe(true);
		});

		it('saves "dark" to localStorage when toggled on', async () => {
			render(Header);
			const toggle = page.getByRole('switch');
			await toggle.click();
			expect(localStorage.getItem('theme')).toBe('dark');
		});

		it('switches back to light mode when toggled off', async () => {
			render(Header);
			const toggle = page.getByRole('switch');
			await toggle.click(); // → dark
			await toggle.click(); // → light
			await expect.element(toggle).toHaveAttribute('aria-checked', 'false');
			expect(document.documentElement.classList.contains('dark')).toBe(false);
		});

		it('saves "light" to localStorage when toggled off', async () => {
			render(Header);
			const toggle = page.getByRole('switch');
			await toggle.click(); // → dark
			await toggle.click(); // → light
			expect(localStorage.getItem('theme')).toBe('light');
		});

		it('initializes to dark mode when localStorage has theme "dark"', async () => {
			localStorage.setItem('theme', 'dark');
			render(Header);
			const toggle = page.getByRole('switch');
			await expect.element(toggle).toHaveAttribute('aria-checked', 'true');
			expect(document.documentElement.classList.contains('dark')).toBe(true);
		});

		it('initializes to light mode when localStorage has theme "light"', async () => {
			localStorage.setItem('theme', 'light');
			render(Header);
			const toggle = page.getByRole('switch');
			await expect.element(toggle).toHaveAttribute('aria-checked', 'false');
			expect(document.documentElement.classList.contains('dark')).toBe(false);
		});
	});
});
