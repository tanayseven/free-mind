import { page } from '@vitest/browser/context';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('/+page.svelte', () => {
	it('should render app title', async () => {
		render(Page);

		const title = page.getByText('Free Mind');
		await expect.element(title).toBeInTheDocument();
	});
});
