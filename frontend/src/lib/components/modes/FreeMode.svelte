<script lang="ts">
    import { onDestroy } from "svelte";
    import { Switch } from "@/components/ui/switch";
    import { Progress } from "$lib/components/ui/progress/index.js";

    let {
        isBlocking,
        onStart,
        onStop,
        disabled = false,
    }: {
        isBlocking: boolean;
        onStart: () => void;
        onStop: () => void;
        disabled?: boolean;
    } = $props();

    const TOTAL_STEPS = 1000;
    const INTERVAL_MS = 50;
    const TOTAL_DURATION_MS = 30000;
    const DECREMENT = TOTAL_STEPS / (TOTAL_DURATION_MS / INTERVAL_MS);

    let showCountdown = $state(false);
    let progressValue = $state(TOTAL_STEPS);
    let intervalId: ReturnType<typeof setInterval> | null = null;

    // Local state for the switch visual — lets us hold it at "on" during the countdown
    let switchIsOn = $state(false);
    $effect(() => {
        switchIsOn = isBlocking;
    });

    function cancelCountdown() {
        if (intervalId !== null) {
            clearInterval(intervalId);
            intervalId = null;
        }
        window.removeEventListener("blur", handleWindowBlur);
        showCountdown = false;
        progressValue = TOTAL_STEPS;
        // Restore the switch to reflect the actual blocking state
        switchIsOn = isBlocking;
    }

    function handleWindowBlur() {
        cancelCountdown();
    }

    function handleSwitchOff() {
        // Immediately snap the switch back to "on" — don't let bits-ui toggle it yet
        switchIsOn = true;
        showCountdown = true;
        progressValue = TOTAL_STEPS;

        window.addEventListener("blur", handleWindowBlur);

        intervalId = setInterval(() => {
            progressValue -= DECREMENT;
            if (progressValue <= 0) {
                progressValue = 0;
                clearInterval(intervalId!);
                intervalId = null;
                window.removeEventListener("blur", handleWindowBlur);
                showCountdown = false;
                onStop();
            }
        }, INTERVAL_MS);
    }

    onDestroy(() => {
        cancelCountdown();
    });
</script>

{#if showCountdown}
    <div class="fixed inset-0 z-50 bg-black/50 flex flex-col items-center justify-center gap-8 p-10">
        <div class="max-w-md text-center space-y-5">
            <p class="text-white/90 text-sm leading-relaxed">
                You are about to exit Focus Mode, where distracting websites are intentionally
                restricted to support your productivity.
            </p>
            <p class="text-white/90 text-sm leading-relaxed">
                Pause for a moment. Take a deep breath, and consider your intention carefully.
            </p>
            <p class="text-white/90 text-sm leading-relaxed">
                Stay committed to your focus—continue only if you are certain you want to proceed.
            </p>
        </div>
        <Progress value={progressValue} max={TOTAL_STEPS} class="w-full max-w-md [&_[data-slot=progress-indicator]]:transition-none" />
    </div>
{/if}

<div class="flex flex-col items-center gap-5 w-full max-w-sm">
    <Switch
        bind:checked={switchIsOn}
        onCheckedChange={(checked) => (checked ? onStart() : handleSwitchOff())}
        {disabled}
        size="sm"
    />
</div>
