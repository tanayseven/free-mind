<script lang="ts">
    import { Switch } from "@/components/ui/switch";

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
</script>

<div class="flex flex-col items-center gap-5 w-full max-w-sm">
    <Switch
        checked={isBlocking}
        onCheckedChange={(checked) => (checked ? onStart() : onStop())}
        {disabled}
        size="sm"
    />
</div>
