<script lang="ts">
    import {onMount} from "svelte";
    import {ConnectToDaemon, SendBlockList, StartBlocking, StopBlocking, InstallAndStartDaemon, CheckDaemonInstalled, CheckBlocking, LoadBlockedWebsites, SaveBlockedWebsites, LoadSettings, SaveSettings} from "../../wailsjs/go/main/App";
    import { Environment } from "../../wailsjs/runtime/runtime";
    import { Switch } from "@/components/ui/switch";
    import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
    import * as ToggleGroup from "$lib/components/ui/toggle-group/index.js";
    import * as Alert from "$lib/components/ui/alert/index.js";
    import { House, LayoutGrid, List, Settings, Sun, Moon } from "@lucide/svelte";
    import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
    import { applyTheme, detectInitialTheme } from "$lib/theme";
    import StatusDot from "$lib/components/modes/StatusDot.svelte";
    import FreeMode from "$lib/components/modes/FreeMode.svelte";
    import TimerMode from "$lib/components/modes/TimerMode.svelte";
    import ScheduleMode from "$lib/components/modes/ScheduleMode.svelte";
    import PomodoroMode from "$lib/components/modes/PomodoroMode.svelte";
    import WebsitesTab, { type WebsiteEntry } from "$lib/components/WebsitesTab.svelte";

    const modeLabels: Record<string, string> = {
        free: "⛓️‍💥 Free",
        timer: "⏲️ Timer",
        schedule: "🗓️ Schedule",
        pomodoro: "⏰ Pomodoro",
    };

    let daemonStatus = $state("Loading... Please wait.");
    let isLoading = $state(true);
    let showInstallButton = $state(false);
    let isBlocking = $state(false);
    let isMac = $state(false);
    let isDark = $state(false);
    let selectedMode = $state("free");
    let websites = $state<WebsiteEntry[]>([]);
    let websitesReady = $state(false);
    let unblockWaiting = $state(30);

    $effect(() => {
        if (!websitesReady) return;
        const json = JSON.stringify(websites);
        SaveBlockedWebsites(json).catch((e: unknown) => console.error("Failed to save blocked websites:", e));
    });

    function toggleDark(checked: boolean) {
        isDark = checked;
        applyTheme(checked, document.documentElement.classList, localStorage);
    }

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

    async function installDaemon() {
        try {
            console.log("Starting installDaemon function");
            isLoading = true;
            showInstallButton = false;
            daemonStatus = "Installing daemon...";

            console.log("Checking if daemon is already installed...");
            const installedCheck = await CheckDaemonInstalled();
            console.log("Daemon installed check result:", installedCheck);

            if (installedCheck) {
                daemonStatus = "Daemon appears to be installed. Attempting to restart...";
                isLoading = false;
                return true;
            }

            console.log("Calling InstallAndStartDaemon...");
            const installResult = await InstallAndStartDaemon();

            console.log("Installation result:", installResult);
            daemonStatus = installResult;

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
                    console.error("Failed to connect to daemon after installation");
                    daemonStatus = "Daemon installed but connection failed. The daemon may not be running. Please restart the application or try manual installation.";
                    showInstallButton = true;
                    isLoading = false;
                }
            }, 3000);

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
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            isDark = detectInitialTheme(localStorage, prefersDark);
            if (isDark) {
                document.documentElement.classList.add('dark');
            }

            const env = await Environment();
            isMac = env.platform === "darwin";

            try {
                const json = await LoadBlockedWebsites();
                websites = JSON.parse(json);
            } catch (e: unknown) {
                console.error("Failed to load blocked websites:", e);
            }
            websitesReady = true;

            try {
                const settings = await LoadSettings();
                unblockWaiting = settings.unblockWaiting;
            } catch (e: unknown) {
                console.error("Failed to load settings:", e);
            }

            console.log("Checking daemon connection on page load...");
            const connected = await checkDaemonConnection();
            console.log("Initial daemon connection check result:", connected);

            if (!connected) {
                console.log("Daemon connection failed, attempting installation");
                await installDaemon();
            } else {
                console.log("Daemon connected successfully on page load");
            }
        } catch (error) {
            console.error("Error during initialization:", error);
            console.error("Full error details:", JSON.stringify(error));
            daemonStatus = "Error initializing application: " + error;
            isLoading = false;
            showInstallButton = true;
        }
        console.log("Initialization complete. isLoading:", isLoading, "showInstallButton:", showInstallButton);
    });

    async function sendStartCommand() {
        try {
            console.log("Sending start command...");
            await ConnectToDaemon();
            const enabledDomains = websites.filter((w) => w.enabled).map((w) => w.domain);
            const blockResult = await SendBlockList(enabledDomains.join(","));
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
    <title>Free Mind</title>
    <meta name="description" content="Block distracting websites and stay focused" />
</svelte:head>

<div class="w-full flex-1 min-h-0 flex flex-col self-stretch">
    {#if isLoading}
        <!-- Minimal header during loading -->
        <div class="w-full flex items-center justify-between px-4 py-2.5 border-b border-border/50">
            <span class="font-bold text-base tracking-tight">Free Mind</span>
        </div>
        <div class="flex flex-1 items-center justify-center">
            <p class="text-muted-foreground text-sm">{daemonStatus}</p>
        </div>
    {:else if showInstallButton}
        <!-- Minimal header during install -->
        <div class="w-full flex items-center justify-between px-4 py-2.5 border-b border-border/50">
            <span class="font-bold text-base tracking-tight">Free Mind</span>
        </div>
        <div class="flex flex-1 flex-col items-center justify-center gap-3">
            <p class="text-destructive text-sm">{daemonStatus}</p>
            <button
                class="h-8 rounded-md px-4 inline-flex items-center justify-center text-xs font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90 transition-colors"
                onclick={installDaemon}
            >
                Install Daemon
            </button>
        </div>
    {:else}
        <Tabs value="home" class="w-full flex flex-col flex-1 min-h-0">
            <!-- Unified header + tabs bar -->
            <div class="w-full flex items-center gap-4 px-5 py-3 border-b border-border/50">
                <span class="font-bold text-lg tracking-tight shrink-0">Free Mind</span>

                <div class="flex-1 flex justify-center">
                    <TabsList variant="line" class="h-10">
                        <TabsTrigger value="home" class="text-sm px-4 h-8 gap-1.5">
                            <House class="size-3.5" />
                            Home
                        </TabsTrigger>
                        <TabsTrigger value="modes" class="text-sm px-4 h-8 gap-1.5">
                            <LayoutGrid class="size-3.5" />
                            Modes
                        </TabsTrigger>
                        <TabsTrigger value="websites" class="text-sm px-4 h-8 gap-1.5">
                            <List class="size-3.5" />
                            Websites
                        </TabsTrigger>
                        <TabsTrigger value="settings" class="text-sm px-4 h-8 gap-1.5">
                            <Settings class="size-3.5" />
                            Settings
                        </TabsTrigger>
                    </TabsList>
                </div>

                <!-- Status + Dark mode toggle -->
                <div class="flex items-center gap-4 shrink-0">
                    <StatusDot {isBlocking} />
                    <div class="flex items-center gap-2">
                        <Sun class="size-4 text-muted-foreground transition-opacity {isDark ? 'opacity-40' : 'opacity-100'}" />
                        <Switch
                            checked={isDark}
                            onCheckedChange={toggleDark}
                            size="sm"
                            class="data-[state=checked]:bg-muted-foreground/40 data-[state=unchecked]:bg-muted-foreground/40"
                        />
                        <Moon class="size-4 text-muted-foreground transition-opacity {isDark ? 'opacity-100' : 'opacity-40'}" />
                    </div>
                </div>
            </div>

            {#if daemonStatus}
                <p class="mt-2 px-4 text-xs text-muted-foreground text-center">{daemonStatus}</p>
            {/if}

            <TabsContent value="home" class="flex flex-1 flex-col items-center justify-center gap-3 p-6">
                <span class="text-xs text-muted-foreground font-medium tracking-wide uppercase">
                    {modeLabels[selectedMode]}
                </span>
                {#if selectedMode === "free"}
                    <FreeMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                        {unblockWaiting}
                    />
                {:else if selectedMode === "timer"}
                    <TimerMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                    />
                {:else if selectedMode === "schedule"}
                    <ScheduleMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                    />
                {:else if selectedMode === "pomodoro"}
                    <PomodoroMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                    />
                {/if}
            </TabsContent>

            <TabsContent value="modes" class="flex flex-1 flex-col items-center gap-6 p-6">
                {#if isBlocking}
                    <Alert.Root variant="destructive" class="w-full max-w-sm">
                        <AlertCircleIcon />
                        <Alert.Title>Blocking is active</Alert.Title>
                        <Alert.Description>
                            Stop blocking before switching modes.
                        </Alert.Description>
                    </Alert.Root>
                {/if}

                <ToggleGroup.Root
                    type="single"
                    bind:value={selectedMode}
                    disabled={isBlocking}
                    class="flex-wrap justify-center"
                >
                    <ToggleGroup.Item value="free">⛓️‍💥 Free</ToggleGroup.Item>
                    <ToggleGroup.Item value="timer">⏲️ Timer</ToggleGroup.Item>
                    <ToggleGroup.Item value="schedule">🗓️ Schedule</ToggleGroup.Item>
                    <ToggleGroup.Item value="pomodoro">⏰ Pomodoro</ToggleGroup.Item>
                </ToggleGroup.Root>

                {#if selectedMode === "free"}
                    <FreeMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                        {unblockWaiting}
                    />
                {:else if selectedMode === "timer"}
                    <TimerMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                    />
                {:else if selectedMode === "schedule"}
                    <ScheduleMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                    />
                {:else if selectedMode === "pomodoro"}
                    <PomodoroMode
                        {isBlocking}
                        onStart={sendStartCommand}
                        onStop={sendStopCommand}
                        disabled={isLoading || showInstallButton}
                    />
                {/if}
            </TabsContent>

            <TabsContent value="websites" class="flex flex-1 min-h-0 flex-col overflow-hidden">
                <WebsitesTab bind:websites {isBlocking} />
            </TabsContent>

            <TabsContent value="settings" class="flex flex-1 flex-col gap-6 overflow-y-auto p-4">
                <div class="flex flex-col gap-1.5">
                    <p class="text-sm font-medium">Unblock waiting time</p>
                    <p class="text-muted-foreground text-xs">Seconds to wait before blocking is disabled in Free Mode.</p>
                    <div class="flex items-center gap-2 mt-1">
                        <input
                            type="number"
                            min="1"
                            max="300"
                            value={unblockWaiting}
                            oninput={(e) => {
                                const v = parseInt((e.target as HTMLInputElement).value, 10);
                                if (!isNaN(v) && v >= 1 && v <= 300) {
                                    unblockWaiting = v;
                                    SaveSettings({ unblockWaiting: v }).catch((err: unknown) =>
                                        console.error("Failed to save settings:", err)
                                    );
                                }
                            }}
                            class="border-input bg-background text-foreground focus-visible:ring-ring w-24 rounded-md border px-3 py-1.5 text-sm focus-visible:ring-1 focus-visible:outline-none"
                        />
                        <span class="text-muted-foreground text-sm">seconds</span>
                    </div>
                </div>
            </TabsContent>
        </Tabs>
    {/if}
</div>

<style>
    /* Custom glow animation handled via Tailwind + inline shadow utilities */
</style>
