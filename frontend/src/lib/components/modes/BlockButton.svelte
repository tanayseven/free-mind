<script lang="ts">
	import { Power } from '@lucide/svelte';

	let {
		isBlocking,
		disabled = false,
		onclick
	}: {
		isBlocking: boolean;
		disabled?: boolean;
		onclick: () => void;
	} = $props();
</script>

<button
	{onclick}
	{disabled}
	role="switch"
	aria-checked={isBlocking}
	aria-label={isBlocking ? 'Stop blocking' : 'Start blocking'}
	class="group flex size-40 flex-col items-center justify-center gap-2 rounded-full border-4 transition-all duration-300 focus-visible:ring-4 focus-visible:ring-ring/40 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50
		{isBlocking
		? 'border-primary bg-primary text-primary-foreground shadow-[0_0_40px_-4px] shadow-primary/60 hover:bg-primary/90'
		: 'border-border bg-card text-muted-foreground hover:border-primary/60 hover:text-primary'}"
>
	<Power class="size-10" strokeWidth={2.25} />
	<span class="text-sm font-semibold tracking-wide uppercase">
		{isBlocking ? 'Stop' : 'Start'}
	</span>
</button>
