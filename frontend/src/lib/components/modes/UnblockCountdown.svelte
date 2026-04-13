<script lang="ts">
    import { onDestroy } from "svelte";
    import { Progress } from "$lib/components/ui/progress/index.js";

    let {
        unblockWaiting = 30,
        onComplete,
        onCancel,
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
            window.removeEventListener("blur", handleBlur);
            onComplete();
        }
    }, INTERVAL_MS);

    function handleBlur() {
        clearInterval(intervalId);
        window.removeEventListener("blur", handleBlur);
        onCancel();
    }

    window.addEventListener("blur", handleBlur);

    onDestroy(() => {
        clearInterval(intervalId);
        window.removeEventListener("blur", handleBlur);
    });
</script>

<div class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm flex flex-col items-center justify-center gap-12 p-16">
    <div class="w-full text-center space-y-8">
        <p class="text-white/90 text-3xl font-semibold leading-snug">
            You are about to exit Focus Mode, where distracting websites are intentionally
            restricted to support your productivity.
        </p>
        <p class="text-white/90 text-3xl font-semibold leading-snug">
            Pause for a moment. Take a deep breath, and consider your intention carefully.
        </p>
        <p class="text-white/90 text-3xl font-semibold leading-snug">
            Stay committed to your focus—continue only if you are certain you want to proceed.
        </p>
    </div>
    <Progress value={progressValue} max={TOTAL_STEPS} class="w-full [&_[data-slot=progress-indicator]]:transition-none" />
</div>
