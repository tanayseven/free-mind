<script lang="ts">
	import BlockButton from './BlockButton.svelte';
	import UnblockCountdown from './UnblockCountdown.svelte';

	let {
		isBlocking,
		onStart,
		onStop,
		disabled = false,
		unblockWaiting = 30
	}: {
		isBlocking: boolean;
		onStart: () => void;
		onStop: () => void;
		disabled?: boolean;
		unblockWaiting?: number;
	} = $props();

	let showCountdown = $state(false);
	// Set when the countdown completes; cleared once isBlocking confirms false
	let stopInitiated = $state(false);

	// Derived: On during countdown, Off once stop is initiated, otherwise mirrors isBlocking
	const switchIsOn = $derived(showCountdown ? true : stopInitiated ? false : isBlocking);

	// Once isBlocking actually reflects the stopped state, clear the flag
	$effect(() => {
		if (!isBlocking) stopInitiated = false;
	});

	function handleSwitchOff() {
		showCountdown = true;
	}

	function handleCountdownComplete() {
		showCountdown = false;
		stopInitiated = true;
		onStop();
	}

	function handleCountdownCancel() {
		showCountdown = false;
		stopInitiated = false;
	}
</script>

{#if showCountdown}
	<UnblockCountdown
		{unblockWaiting}
		onComplete={handleCountdownComplete}
		onCancel={handleCountdownCancel}
	/>
{/if}

<div class="flex w-full max-w-sm flex-col items-center gap-5">
	<p class="text-sm text-muted-foreground">Block until you choose to stop.</p>
	<BlockButton
		isBlocking={switchIsOn}
		{disabled}
		onclick={() => (switchIsOn ? handleSwitchOff() : onStart())}
	/>
</div>
