<script lang="ts">
    import CalendarIcon from "@lucide/svelte/icons/calendar";
    import StatusDot from "./StatusDot.svelte";

    let {
        isBlocking,
    }: {
        isBlocking: boolean;
    } = $props();

    const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
    let selectedDays = $state(new Set(["Mon", "Tue", "Wed", "Thu", "Fri"]));

    function toggleDay(day: string) {
        const next = new Set(selectedDays);
        if (next.has(day)) {
            next.delete(day);
        } else {
            next.add(day);
        }
        selectedDays = next;
    }
</script>

<div class="flex flex-col items-center gap-5 w-full max-w-sm">
    <StatusDot {isBlocking} />

    <div class="flex flex-col items-center gap-4 w-full">
        <div class="flex items-center gap-2 text-muted-foreground">
            <CalendarIcon class="size-4" />
            <span class="text-sm font-medium">Block on a schedule</span>
        </div>

        <div class="flex gap-1.5 flex-wrap justify-center">
            {#each days as day}
                <button
                    class="size-9 rounded-md text-xs font-medium border transition-colors
                        {selectedDays.has(day)
                        ? 'bg-primary text-primary-foreground border-primary'
                        : 'bg-background text-foreground border-border hover:bg-muted'}"
                    onclick={() => toggleDay(day)}
                >
                    {day}
                </button>
            {/each}
        </div>

        <div class="flex items-center gap-3 text-sm">
            <div class="flex flex-col items-start gap-1">
                <label class="text-xs text-muted-foreground" for="start-time">From</label>
                <input
                    id="start-time"
                    type="time"
                    value="09:00"
                    class="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground"
                    disabled
                />
            </div>
            <span class="text-muted-foreground mt-4">–</span>
            <div class="flex flex-col items-start gap-1">
                <label class="text-xs text-muted-foreground" for="end-time">To</label>
                <input
                    id="end-time"
                    type="time"
                    value="17:00"
                    class="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground"
                    disabled
                />
            </div>
        </div>

        <p class="text-xs text-muted-foreground">
            Schedule mode coming soon — block automatically on selected days and times.
        </p>
    </div>
</div>
