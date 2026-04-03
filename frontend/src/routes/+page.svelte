<script lang="ts">
    import {onMount} from "svelte";
    import {ConnectToDaemon, SendBlockList, StartBlocking, StopBlocking, InstallAndStartDaemon, CheckDaemonInstalled, CheckBlocking} from "../../wailsjs/go/main/App";
    import { Switch } from "@/components/ui/switch";

    let daemonStatus = "Loading... Please wait.";
    let isLoading = true;
    let showInstallButton = false;
    let isBlocking = false;

    // Function to check daemon connection and handle errors
    async function checkDaemonConnection() {
        try {
            isLoading = true;
            const result = await ConnectToDaemon();

            if (result.includes("Error")) {
                daemonStatus = "Daemon is not installed or not running.";
                isLoading = false;
                showInstallButton = true;
                return false;
            }

            daemonStatus = "";
            isLoading = false;
            showInstallButton = false;
            isBlocking = await CheckBlocking();
            return true;
        } catch (error) {
            daemonStatus = "Error connecting to daemon: " + error;
            isLoading = false;
            showInstallButton = true;
            return false;
        }
    }

    // Function to install and start the daemon
    async function installDaemon() {
        try {
            console.log("Starting installDaemon function");
            isLoading = true;
            showInstallButton = false;
            daemonStatus = "Installing daemon...";
            
            // First check if daemon is already installed
            console.log("Checking if daemon is already installed...");
            const installedCheck = await CheckDaemonInstalled();
            console.log("Daemon installed check result:", installedCheck);
            
            if (installedCheck) {
                daemonStatus = "Daemon appears to be installed. Attempting to restart...";
                isLoading = false;
                return true; // Already installed
            }
            
            // Install and start the daemon
            console.log("Calling InstallAndStartDaemon...");
            const installResult = await InstallAndStartDaemon();
            
            console.log("Installation result:", installResult);
            daemonStatus = installResult;
            
            // Try to connect to the daemon after installation
            // Increase timeout to give daemon more time to start
            console.log("Setting timeout to check daemon connection after installation...");
            setTimeout(async () => {
                console.log("Timeout elapsed, checking daemon connection...");
                const connected = await checkDaemonConnection();
                console.log("Daemon connection check result:", connected);
                if (connected) {
                    console.log("Daemon connected successfully after installation");
                    daemonStatus = "Daemon installed and connected successfully";
                    isLoading = false;
                } else {
                    // If connection failed after installation, show a more detailed error message
                    console.error("Failed to connect to daemon after installation");
                    daemonStatus = "Daemon installed but connection failed. The daemon may not be running. Please restart the application or try manual installation.";
                    showInstallButton = true;
                    isLoading = false;
                }
            }, 3000); // Wait 3 seconds for the daemon to start (increased from 2 seconds)
            
        } catch (error) {
            console.error("Error installing daemon:", error);
            console.error("Full error details:", JSON.stringify(error));
            daemonStatus = "Error installing daemon: " + error;
            showInstallButton = true;
            isLoading = false;
        }
    }

    onMount(async () => {
        console.log("Component mounted, initializing application...");
        try {
            // Try to connect to the daemon on page load
            console.log("Checking daemon connection on page load...");
            const connected = await checkDaemonConnection();
            console.log("Initial daemon connection check result:", connected);
            
            // If connection failed, try to install the daemon automatically
            if (!connected) {
                console.log("Daemon connection failed, attempting installation");
                await installDaemon();
            } else {
                console.log("Daemon connected successfully on page load");
            }
        } catch (error) {
            // Ensure we're not stuck in loading state if there's an error
            console.error("Error during initialization:", error);
            console.error("Full error details:", JSON.stringify(error));
            daemonStatus = "Error initializing application: " + error;
            isLoading = false;
            showInstallButton = true;
        }
        console.log("Initialization complete. isLoading:", isLoading, "showInstallButton:", showInstallButton);
    });

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
            console.log("Sending start command...");
            await ConnectToDaemon();
            const blockResult = await SendBlockList(websitesToBeBlocked.join(","));
            console.log("SendBlockList result:", blockResult);
            if (blockResult) {
                const startResult = await StartBlocking();
                console.log("StartBlocking result:", startResult);
                if (startResult) isBlocking = true;
            }
        } catch (error) {
            console.error("Error calling commands:", error);
            console.error("Full error details:", JSON.stringify(error));
        }
    }

    // Function to call the Go Write function
    async function sendStopCommand() {
        try {
            console.log("Sending stop command...");
            await ConnectToDaemon();
            const stopResult = await StopBlocking();
            console.log("StopBlocking result:", stopResult);
            isBlocking = await CheckBlocking();
        } catch (error) {
            console.error("Error calling Stop command:", error);
            console.error("Full error details:", JSON.stringify(error));
        }
    }
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="Svelte demo app" />
</svelte:head>

<section class="w-full max-w-4xl mx-auto text-center py-8">
    <h1 class="text-3xl font-bold mb-8">Welcome to Free-Mind!</h1>
    
    {#if isLoading}
        <div class="mb-4">
            <p>{daemonStatus}</p>
        </div>
    {:else if showInstallButton}
        <div class="mb-4">
            <p class="text-red-500 mb-2">{daemonStatus}</p>
            <button
                class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-blue-500 text-white shadow-xs hover:bg-blue-600 mb-4"
                on:click={installDaemon}
            >
                Install Daemon
            </button>
        </div>
    {:else}
        {#if daemonStatus}
            <p class="mb-4 text-sm text-muted-foreground">{daemonStatus}</p>
        {/if}
        <div class="flex justify-center items-center gap-4">
            <Switch
                checked={isBlocking}
                onCheckedChange={(checked) => checked ? sendStartCommand() : sendStopCommand()}
                disabled={isLoading || showInstallButton}
                size="lg"
                class={isBlocking
                    ? "data-[state=checked]:bg-red-500"
                    : "data-[state=unchecked]:bg-green-500"}
            />
            <span class="text-lg font-semibold {isBlocking ? 'text-red-500' : 'text-green-500'}">
                {isBlocking ? "Stop" : "Start"}
            </span>
        </div>
    {/if}
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
