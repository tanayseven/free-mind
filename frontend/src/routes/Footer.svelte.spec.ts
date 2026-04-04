import { page, userEvent } from '@vitest/browser/context';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Footer from './Footer.svelte';

const mockBrowserOpenURL = vi.fn();

vi.mock('../../wailsjs/runtime/runtime', () => ({
	BrowserOpenURL: (url: string) => mockBrowserOpenURL(url)
}));

describe('Footer.svelte', () => {
	beforeEach(() => {
		mockBrowserOpenURL.mockClear();
	});

	describe('Tooltip for ❤️', () => {
		it('shows "love" tooltip on hover', async () => {
			render(Footer);
			const trigger = page.getByText('❤️');
			await userEvent.hover(trigger.element());
			await expect.element(page.getByText('love')).toBeVisible();
		});
	});

	describe('Tooltip for 😓', () => {
		it('shows "sweat" tooltip on hover', async () => {
			render(Footer);
			const trigger = page.getByText('😓');
			await userEvent.hover(trigger.element());
			await expect.element(page.getByText('sweat')).toBeVisible();
		});
	});

	describe('Link for "Tanay PrabhuDesai"', () => {
		it('calls BrowserOpenURL with https://tanay.tech when clicked', async () => {
			render(Footer);
			const button = page.getByRole('button', { name: 'Tanay PrabhuDesai' });
			await button.click();
			expect(mockBrowserOpenURL).toHaveBeenCalledWith('https://tanay.tech');
		});
	});

	describe('Link for "Free Mind"', () => {
		it('calls BrowserOpenURL with https://freemind.tanay.tech when clicked', async () => {
			render(Footer);
			const button = page.getByRole('button', { name: 'Free Mind' });
			await button.click();
			expect(mockBrowserOpenURL).toHaveBeenCalledWith('https://freemind.tanay.tech');
		});
	});
});
