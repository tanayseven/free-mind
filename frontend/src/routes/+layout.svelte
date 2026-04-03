<script lang="ts">
	import Header from './Header.svelte';
	import '../app.css';
	import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
	import {
		Tooltip,
		TooltipContent,
		TooltipTrigger,
		TooltipProvider
	} from '@/components/ui/tooltip';

	let { children } = $props();
</script>

<div class="app min-h-screen flex flex-col">
	<Header />

	<main class="flex-1 flex flex-col items-center justify-center p-4 md:p-6 lg:p-8">
		{@render children()}
	</main>

	<footer class="w-full bg-secondary py-4 px-6 text-center text-secondary-foreground text-sm md:text-base">
		<p class="max-w-4xl mx-auto">
			Made with <TooltipProvider><Tooltip><TooltipTrigger>❤️</TooltipTrigger><TooltipContent>love</TooltipContent></Tooltip></TooltipProvider> and <TooltipProvider><Tooltip><TooltipTrigger>😓</TooltipTrigger><TooltipContent>sweat</TooltipContent></Tooltip></TooltipProvider> by <button onclick={() => BrowserOpenURL('https://tanay.tech')} class="text-primary hover:underline cursor-pointer">Tanay PrabhuDesai</button> visit <button onclick={() => BrowserOpenURL('https://freemind.tanay.tech')} class="text-primary hover:underline cursor-pointer">Freemind</button> to learn about this application.
		</p>
	</footer>
</div>

<style>
	/* Additional styles can be added here if needed */
</style>
