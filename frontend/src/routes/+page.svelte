<script lang="ts">
    import {FetchDaemonPort, SendBlockList, StartBlocking, StopBlocking} from "../../wailsjs/go/main/App";

    const websitesToBeBlocked = [
        "youtube.com",
        "www.youtube.com",
        "facebook.com",
        "www.facebook.com",
        "instagram.com",
        "www.instagram.com",
    ]

    // Function to call the Go Write function
    async function sendStartCommand() {
        try {
            await FetchDaemonPort()
            await SendBlockList(websitesToBeBlocked.join(","));
            await StartBlocking();
        } catch (error) {
            console.error("Error calling Write function:", error);
        }
    }

    // Function to call the Go Write function
    async function sendStopCommand() {
        try {
            await FetchDaemonPort()
            await StopBlocking();
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
            on:click={sendStartCommand}
        >
            Start
        </button>
        <button
            class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
            on:click={sendStopCommand}
        >
            Stop
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
