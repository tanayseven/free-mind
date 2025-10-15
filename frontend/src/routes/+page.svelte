<script lang="ts">
    import {Read, Write} from "../../wailsjs/go/main/App";
    
    // Function to call the Go Read function
    async function callRead() {
        try {
            const result = await Read("example");
            alert(result);
        } catch (error) {
            console.error("Error calling Read function:", error);
        }
    }
    
    // Function to call the Go Write function
    async function callWrite() {
        try {
            const result = await Write("User");
            alert(result);
        } catch (error) {
            console.error("Error calling Write function:", error);
        }
    }
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="Svelte demo app" />
</svelte:head>

<section class="w-full max-w-4xl mx-auto text-center py-8">
    <h1 class="text-3xl font-bold mb-8">Welcome to Free-Mind!</h1>
    <div class="flex justify-center gap-4">
        <button
            class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
            on:click={callRead}
        >
            Read
        </button>
        <button
            class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
            on:click={callWrite}
        >
            Write
        </button>
    </div>
</section>

<style>
    /* Responsive adjustments */
    @media (max-width: 640px) {
        h1 {
            font-size: 1.75rem;
            margin-bottom: 1.5rem;
        }
    }
</style>
