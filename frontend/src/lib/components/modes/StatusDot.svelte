<script lang="ts">
	let { isBlocking }: { isBlocking: boolean } = $props();
</script>

<div class="flex items-center gap-2.5">
	{#if isBlocking}
		<span class="relative flex size-4">
			<span
				class="absolute -inset-1 inline-flex animate-ping rounded-full bg-red-500/25 [animation-duration:1.6s]"
			></span>
			<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-500/70"
			></span>
			<span
				class="relative inline-flex size-4 animate-pulse rounded-full bg-red-500 shadow-[0_0_16px_6px_rgba(239,68,68,0.65)]"
			></span>
		</span>
	{:else}
		<span class="relative flex size-4">
			<span
				class="inline-flex size-4 rounded-full bg-green-500 shadow-[0_0_10px_3px_rgba(34,197,94,0.4)]"
			></span>
		</span>
	{/if}
	<span class="text-sm font-medium {isBlocking ? 'text-red-500' : 'text-green-500'}">
		{isBlocking ? 'Blocking' : 'Not blocking'}
	</span>
</div>
