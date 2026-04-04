<script lang="ts">
    import AlarmClockIcon from "@lucide/svelte/icons/alarm-clock";
    import StatusDot from "./StatusDot.svelte";

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

    let workMinutes = $state(25);
    let breakMinutes = $state(5);
    let cycles = $state(4);
</script>

<div class="flex flex-col items-center gap-5 w-full max-w-sm">
    <StatusDot {isBlocking} />

    <div class="flex flex-col items-center gap-4 w-full">
        <div class="flex items-center gap-2 text-muted-foreground">
            <AlarmClockIcon class="size-4" />
            <span class="text-sm font-medium">Pomodoro timer</span>
        </div>

        <div class="grid grid-cols-3 gap-3 w-full text-center">
            <div class="flex flex-col items-center gap-1.5">
                <span class="text-xs text-muted-foreground">Work</span>
                <div class="flex items-center gap-1">
                    <button
                        class="size-6 rounded border border-border text-xs hover:bg-muted transition-colors disabled:opacity-50"
                        onclick={() => workMinutes > 5 && (workMinutes -= 5)}
                        {disabled}
                    >−</button>
                    <span class="text-sm font-medium w-8 text-center">{workMinutes}m</span>
                    <button
                        class="size-6 rounded border border-border text-xs hover:bg-muted transition-colors disabled:opacity-50"
                        onclick={() => workMinutes < 60 && (workMinutes += 5)}
                        {disabled}
                    >+</button>
                </div>
            </div>

            <div class="flex flex-col items-center gap-1.5">
                <span class="text-xs text-muted-foreground">Break</span>
                <div class="flex items-center gap-1">
                    <button
                        class="size-6 rounded border border-border text-xs hover:bg-muted transition-colors disabled:opacity-50"
                        onclick={() => breakMinutes > 1 && (breakMinutes -= 1)}
                        {disabled}
                    >−</button>
                    <span class="text-sm font-medium w-8 text-center">{breakMinutes}m</span>
                    <button
                        class="size-6 rounded border border-border text-xs hover:bg-muted transition-colors disabled:opacity-50"
                        onclick={() => breakMinutes < 30 && (breakMinutes += 1)}
                        {disabled}
                    >+</button>
                </div>
            </div>

            <div class="flex flex-col items-center gap-1.5">
                <span class="text-xs text-muted-foreground">Cycles</span>
                <div class="flex items-center gap-1">
                    <button
                        class="size-6 rounded border border-border text-xs hover:bg-muted transition-colors disabled:opacity-50"
                        onclick={() => cycles > 1 && (cycles -= 1)}
                        {disabled}
                    >−</button>
                    <span class="text-sm font-medium w-8 text-center">{cycles}</span>
                    <button
                        class="size-6 rounded border border-border text-xs hover:bg-muted transition-colors disabled:opacity-50"
                        onclick={() => cycles < 10 && (cycles += 1)}
                        {disabled}
                    >+</button>
                </div>
            </div>
        </div>

        <p class="text-xs text-muted-foreground">
            Pomodoro mode coming soon — block during work intervals, unblock during breaks.
        </p>

        <button
            class="h-8 rounded-md px-5 text-xs font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
            onclick={isBlocking ? onStop : onStart}
            {disabled}
        >
            {isBlocking ? "Stop blocking" : `Start ${cycles} × ${workMinutes}min session`}
        </button>
    </div>
</div>
