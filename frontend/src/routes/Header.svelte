<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Switch } from "@/components/ui/switch";
    import { Sun, Moon } from "@lucide/svelte";
    import { onMount } from "svelte";

    let isDark = $state(false);

    function toggleDark(checked: boolean) {
        isDark = checked;
        if (checked) {
            document.documentElement.classList.add('dark');
            localStorage.setItem('theme', 'dark');
        } else {
            document.documentElement.classList.remove('dark');
            localStorage.setItem('theme', 'light');
        }
    }

    onMount(() => {
        const saved = localStorage.getItem('theme');
        if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
            isDark = true;
            document.documentElement.classList.add('dark');
        }
    });
</script>

<header class="w-full bg-primary py-3 px-4 md:py-4 md:px-6 flex items-center justify-between">
    <div class="text-primary-foreground font-bold text-lg md:text-xl">Free-Mind</div>
    <nav class="flex gap-2 md:gap-4 items-center">
        <Button variant="secondary" size="sm" class="text-xs md:text-sm">Home</Button>
        <Button variant="secondary" size="sm" class="text-xs md:text-sm">About</Button>
        <div class="flex items-center gap-1.5 ml-2">
            <Sun
                class="size-4 text-primary-foreground transition-opacity {isDark ? 'opacity-40' : 'opacity-100'}"
            />
            <Switch
                checked={isDark}
                onCheckedChange={toggleDark}
                size="sm"
                class="data-[state=checked]:bg-primary-foreground/40 data-[state=unchecked]:bg-primary-foreground/40"
            />
            <Moon
                class="size-4 text-primary-foreground transition-opacity {isDark ? 'opacity-100' : 'opacity-40'}"
            />
        </div>
    </nav>
</header>

<style>
    /* Additional styles can be added here if needed */
</style>
