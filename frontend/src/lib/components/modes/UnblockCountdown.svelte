<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Progress } from '$lib/components/ui/progress/index.js';

	let {
		unblockWaiting = 30,
		onComplete,
		onCancel
	}: {
		unblockWaiting?: number;
		onComplete: () => void;
		onCancel: () => void;
	} = $props();

	const TOTAL_STEPS = 1000;
	const INTERVAL_MS = 50;
	const decrement = TOTAL_STEPS / ((unblockWaiting * 1000) / INTERVAL_MS);

	let progressValue = $state(TOTAL_STEPS);
	const intervalId = setInterval(() => {
		progressValue -= decrement;
		if (progressValue <= 0) {
			progressValue = 0;
			clearInterval(intervalId);
			window.removeEventListener('blur', handleBlur);
			onComplete();
		}
	}, INTERVAL_MS);

	function handleBlur() {
		clearInterval(intervalId);
		window.removeEventListener('blur', handleBlur);
		onCancel();
	}

	window.addEventListener('blur', handleBlur);

	onDestroy(() => {
		clearInterval(intervalId);
		window.removeEventListener('blur', handleBlur);
	});
</script>

<div
	class="fixed inset-0 z-50 flex flex-col items-center justify-center gap-12 bg-black/50 p-16 backdrop-blur-sm"
>
	<div class="w-full space-y-8 text-center">
		<p class="text-3xl leading-snug font-semibold text-white/90">
			You are about to exit Focus Mode, where distracting websites are intentionally restricted to
			support your productivity.
		</p>
		<p class="text-3xl leading-snug font-semibold text-white/90">
			Pause for a moment. Take a deep breath, and consider your intention carefully.
		</p>
		<p class="text-3xl leading-snug font-semibold text-white/90">
			Stay committed to your focus—continue only if you are certain you want to proceed.
		</p>
	</div>
	<Progress
		value={progressValue}
		max={TOTAL_STEPS}
		class="w-full [&_[data-slot=progress-indicator]]:transition-none"
	/>
</div>
