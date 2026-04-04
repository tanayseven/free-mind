<script lang="ts">
	import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
	import {
		Tooltip,
		TooltipContent,
		TooltipTrigger,
		TooltipProvider
	} from '@/components/ui/tooltip';
</script>

<footer class="w-full bg-secondary py-4 px-6 text-center text-secondary-foreground text-sm md:text-base">
	<p class="max-w-4xl mx-auto">
		Made with <TooltipProvider><Tooltip><TooltipTrigger>❤️</TooltipTrigger><TooltipContent>love</TooltipContent></Tooltip></TooltipProvider> and <TooltipProvider><Tooltip><TooltipTrigger>😓</TooltipTrigger><TooltipContent>sweat</TooltipContent></Tooltip></TooltipProvider> by <button onclick={() => BrowserOpenURL('https://tanay.tech')} class="text-primary hover:underline cursor-pointer">Tanay PrabhuDesai</button> head to <button onclick={() => BrowserOpenURL('https://freemind.tanay.tech')} class="text-primary hover:underline cursor-pointer">Free Mind</button> to learn about this application.
	</p>
</footer>
