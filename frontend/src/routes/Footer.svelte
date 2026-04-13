<script lang="ts">
	import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
	import {
		Tooltip,
		TooltipContent,
		TooltipTrigger,
		TooltipProvider
	} from '@/components/ui/tooltip';
</script>

<footer class="w-full bg-secondary px-4 py-1.5 text-center text-xs text-secondary-foreground/60">
	<p class="mx-auto max-w-4xl">
		Made with <TooltipProvider
			><Tooltip><TooltipTrigger>❤️</TooltipTrigger><TooltipContent>love</TooltipContent></Tooltip
			></TooltipProvider
		> and <TooltipProvider
			><Tooltip><TooltipTrigger>😓</TooltipTrigger><TooltipContent>sweat</TooltipContent></Tooltip
			></TooltipProvider
		> by
		<button
			onclick={() => BrowserOpenURL('https://tanay.tech')}
			class="cursor-pointer text-primary hover:underline">Tanay PrabhuDesai</button
		>
		head to
		<button
			onclick={() => BrowserOpenURL('https://freemind.tanay.tech')}
			class="cursor-pointer text-primary hover:underline">Free Mind</button
		> to learn about this application.
	</p>
</footer>
