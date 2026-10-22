<script lang="ts">
	let { isBlocking }: { isBlocking: boolean } = $props();
</script>

<!-- Fixed width so the header doesn't shift when the label changes. -->
<div class="flex w-24 items-center justify-end gap-2" role="status">
	<span class="relative flex size-2.5">
		{#if isBlocking}
			<span
				class="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary/60 [animation-duration:2s]"
			></span>
		{/if}
		<span
			class="relative inline-flex size-2.5 rounded-full {isBlocking
				? 'bg-primary'
				: 'bg-muted-foreground/40'}"
		></span>
	</span>
	<span class="text-sm font-medium {isBlocking ? 'text-primary' : 'text-muted-foreground'}">
		{isBlocking ? 'Focusing' : 'Idle'}
	</span>
</div>
