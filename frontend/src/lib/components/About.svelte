<script lang="ts">
	import { BrowserOpenURL } from '$lib/api';
	import {
		Tooltip,
		TooltipContent,
		TooltipTrigger,
		TooltipProvider
	} from '@/components/ui/tooltip';
</script>

<p class="text-sm text-muted-foreground">
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
	>. Learn more about
	<button
		onclick={() => BrowserOpenURL('https://freemind.tanay.tech')}
		class="cursor-pointer text-primary hover:underline">Free Mind</button
	>.
</p>
<p class="mt-1 text-xs text-muted-foreground/80">
	Works offline. Free Mind never collects any data.
</p>
