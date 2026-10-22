<script lang="ts">
	import { Monitor, Sun, Moon, Check } from '@lucide/svelte';
	import { SaveSettings } from '$lib/api';
	import type { ThemePreference } from '$lib/theme';
	import About from './About.svelte';

	let {
		theme,
		onThemeChange,
		unblockWaiting = $bindable(30)
	}: {
		theme: ThemePreference;
		onThemeChange: (theme: ThemePreference) => void;
		unblockWaiting?: number;
	} = $props();

	const themes: { value: ThemePreference; label: string; icon: typeof Sun }[] = [
		{ value: 'system', label: 'System', icon: Monitor },
		{ value: 'light', label: 'Light', icon: Sun },
		{ value: 'dark', label: 'Dark', icon: Moon }
	];
	const waitPresets = [10, 30, 60, 120];
	const MIN_WAIT = 1;
	const MAX_WAIT = 300;

	let saved = $state(false);
	let savedTimer: ReturnType<typeof setTimeout> | undefined;

	async function setWait(seconds: number) {
		if (isNaN(seconds) || seconds < MIN_WAIT || seconds > MAX_WAIT) return;
		unblockWaiting = seconds;
		try {
			await SaveSettings({ unblockWaiting: seconds });
			saved = true;
			clearTimeout(savedTimer);
			savedTimer = setTimeout(() => (saved = false), 1500);
		} catch (err: unknown) {
			console.error('Failed to save settings:', err);
		}
	}

	const segment =
		'inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm font-medium transition-colors';
	const segmentOn = 'bg-background text-foreground shadow-sm';
	const segmentOff = 'text-muted-foreground hover:text-foreground';
</script>

<div class="mx-auto flex w-full max-w-xl flex-col gap-8 p-6">
	<section class="flex flex-col gap-2">
		<h2 class="text-sm font-semibold">Appearance</h2>
		<div
			role="radiogroup"
			aria-label="Theme"
			class="inline-flex w-fit gap-1 rounded-lg bg-muted p-1"
		>
			{#each themes as t (t.value)}
				<button
					role="radio"
					aria-checked={theme === t.value}
					onclick={() => onThemeChange(t.value)}
					class="{segment} {theme === t.value ? segmentOn : segmentOff}"
				>
					<t.icon class="size-4" />
					{t.label}
				</button>
			{/each}
		</div>
	</section>

	<section class="flex flex-col gap-2">
		<div class="flex items-center gap-2">
			<h2 class="text-sm font-semibold">Wait before unblocking</h2>
			{#if saved}
				<span class="inline-flex items-center gap-1 text-xs text-primary" role="status">
					<Check class="size-3.5" /> Saved
				</span>
			{/if}
		</div>
		<p class="text-sm text-muted-foreground">
			When you stop blocking in Free mode, Free Mind makes you wait this long first. The pause gives
			you a moment to reconsider before distractions come back.
		</p>
		<div class="mt-1 flex flex-wrap items-center gap-3">
			<div
				role="radiogroup"
				aria-label="Wait time"
				class="inline-flex gap-1 rounded-lg bg-muted p-1"
			>
				{#each waitPresets as s (s)}
					<button
						role="radio"
						aria-checked={unblockWaiting === s}
						onclick={() => setWait(s)}
						class="{segment} {unblockWaiting === s ? segmentOn : segmentOff}"
					>
						{s < 60 ? `${s}s` : `${s / 60} min`}
					</button>
				{/each}
			</div>
			<label class="flex items-center gap-2 text-sm text-muted-foreground">
				or
				<input
					type="number"
					min={MIN_WAIT}
					max={MAX_WAIT}
					value={unblockWaiting}
					onchange={(e) => setWait(parseInt(e.currentTarget.value, 10))}
					class="w-20 rounded-md border border-input bg-background px-2 py-1 text-sm text-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none"
				/>
				seconds
			</label>
		</div>
	</section>

	<section class="flex flex-col gap-2">
		<h2 class="text-sm font-semibold">About</h2>
		<About />
	</section>
</div>
