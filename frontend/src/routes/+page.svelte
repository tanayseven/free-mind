<script lang="ts">
    import {onMount} from "svelte";
    import {ConnectToDaemon, SendBlockList, StartBlocking, StopBlocking, InstallAndStartDaemon, CheckDaemonInstalled, CheckBlocking} from "../../wailsjs/go/main/App";
    import { Environment } from "../../wailsjs/runtime/runtime";
    import { Switch } from "@/components/ui/switch";
    import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
    import { Settings, List } from "@lucide/svelte";

    let daemonStatus = "Loading... Please wait.";
    let isLoading = true;
    let showInstallButton = false;
    let isBlocking = false;
    let isMac = false;

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
            const env = await Environment();
            isMac = env.platform === "darwin";

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

<div class="w-full flex-1 flex flex-col self-stretch">
    {#if isLoading}
        <div class="flex flex-1 items-center justify-center">
            <p>{daemonStatus}</p>
        </div>
    {:else if showInstallButton}
        <div class="flex flex-1 flex-col items-center justify-center gap-2">
            <p class="text-destructive">{daemonStatus}</p>
            <button
                class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
                on:click={installDaemon}
            >
                Install Daemon
            </button>
        </div>
    {:else}
        <Tabs value="free-mode" class="w-full flex flex-col flex-1">
            <div class="flex justify-center px-2 pt-1">
                <TabsList>
                    <TabsTrigger value="free-mode">Free Mode</TabsTrigger>
                    <TabsTrigger value="timer-mode">Timer Mode</TabsTrigger>
                    <TabsTrigger value="schedule-mode">Schedule Mode</TabsTrigger>
                    <TabsTrigger value="pomodoro-mode">Pomodoro Mode</TabsTrigger>
                    <TabsTrigger value="blocked-sites">
                        <List class="size-4" />
                        Blocked Sites
                    </TabsTrigger>
                    <TabsTrigger value="settings">
                        <Settings class="size-4" />
                        {isMac ? "Preferences" : "Settings"}
                    </TabsTrigger>
                </TabsList>
            </div>

            {#if daemonStatus}
                <p class="mt-2 px-4 text-sm text-muted-foreground text-center">{daemonStatus}</p>
            {/if}

            <TabsContent value="free-mode" class="flex flex-1 items-center justify-center">
                <div class="flex items-center gap-4">
                    <Switch
                        checked={isBlocking}
                        onCheckedChange={(checked) => checked ? sendStartCommand() : sendStopCommand()}
                        disabled={isLoading || showInstallButton}
                        size="lg"
                        class={isBlocking
                            ? "data-[state=checked]:bg-destructive"
                            : "data-[state=unchecked]:bg-primary"}
                    />
                    <span class="text-lg font-semibold {isBlocking ? 'text-destructive' : 'text-primary'}">
                        {isBlocking ? "Stop" : "Start"}
                    </span>
                </div>
            </TabsContent>

            <TabsContent value="timer-mode" class="flex flex-1 items-center justify-center">
                <p class="text-muted-foreground">Timer Mode — coming soon.</p>
            </TabsContent>

            <TabsContent value="schedule-mode" class="flex flex-1 items-center justify-center">
                <p class="text-muted-foreground">Schedule Mode — coming soon.</p>
            </TabsContent>

            <TabsContent value="pomodoro-mode" class="flex flex-1 items-center justify-center">
                <p class="text-muted-foreground">Pomodoro Mode — coming soon.</p>
            </TabsContent>

            <TabsContent value="blocked-sites" class="flex flex-1 items-center justify-center">
                <p class="text-muted-foreground">Blocked Sites — coming soon.</p>
            </TabsContent>

            <TabsContent value="settings" class="flex flex-1 items-center justify-center">
                <p class="text-muted-foreground">{isMac ? "Preferences" : "Settings"} — coming soon.</p>
            </TabsContent>
        </Tabs>
    {/if}
</div>
