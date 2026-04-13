<script lang="ts">
	import TimerIcon from '@lucide/svelte/icons/timer';
	import { Switch } from '@/components/ui/switch';

	let {
		isBlocking,
		onStart,
		onStop,
		disabled = false
	}: {
		isBlocking: boolean;
		onStart: () => void;
		onStop: () => void;
		disabled?: boolean;
	} = $props();

	const presets = [
		{ label: '15 min', minutes: 15 },
		{ label: '30 min', minutes: 30 },
		{ label: '1 hour', minutes: 60 },
		{ label: '2 hours', minutes: 120 }
	];

	let selectedMinutes = $state(30);
</script>

<div class="flex w-full max-w-sm flex-col items-center gap-5">
	<div class="flex w-full flex-col items-center gap-4">
		<div class="flex items-center gap-2 text-muted-foreground">
			<TimerIcon class="size-4" />
			<span class="text-sm font-medium">Block for a duration</span>
		</div>

		<div class="flex flex-wrap justify-center gap-2">
			{#each presets as preset}
				<button
					class="h-8 rounded-md border px-3 text-xs font-medium transition-colors
                        {selectedMinutes === preset.minutes
						? 'border-primary bg-primary text-primary-foreground'
						: 'border-border bg-background text-foreground hover:bg-muted'}"
					onclick={() => (selectedMinutes = preset.minutes)}
					{disabled}
				>
					{preset.label}
				</button>
			{/each}
		</div>

		<p class="text-xs text-muted-foreground">
			Timer mode coming soon — set a duration and block automatically.
		</p>

		<Switch
			checked={isBlocking}
			onCheckedChange={(checked) => (checked ? onStart() : onStop())}
			{disabled}
			size="sm"
		/>
	</div>
</div>
