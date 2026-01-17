<script lang="ts">
    import {onMount} from "svelte";
    import {FetchDaemonPort, SendBlockList, StartBlocking, StopBlocking, InstallDaemonWithOneClick} from "../../wailsjs/go/main/App";

    let daemonStatus = "Loading... Please wait.";
    let isLoading = true;
    let showInstallButton = false;

    // Function to check daemon connection and handle errors
    async function checkDaemonConnection() {
        try {
            console.log("Starting checkDaemonConnection function");
            isLoading = true;
            // Try to connect to the daemon
            console.log("Calling FetchDaemonPort...");
            daemonStatus = await FetchDaemonPort();
            console.log("Daemon status:", daemonStatus);
            
            // Additional check to verify if the daemon is actually running
            // If the response contains "Error", it means the daemon is not running
            if (daemonStatus.includes("Error")) {
                console.error("Daemon connection failed:", daemonStatus);
                daemonStatus = "Error connecting to daemon. Daemon may not be installed or running.";
                isLoading = false;
                showInstallButton = true;
                return false;
            }
            
            // Try to send a test message to verify the connection
            try {
                console.log("Sending test message to daemon...");
                // Use SendBlockList as a test to verify the connection works
                const testResult = await SendBlockList("test");
                console.log("Test message result:", testResult);
                if (!testResult) {
                    console.error("Daemon test message failed");
                    daemonStatus = "Daemon is installed but not responding to commands.";
                    isLoading = false;
                    showInstallButton = true;
                    return false;
                }
            } catch (testError) {
                console.error("Error testing daemon connection:", testError);
                console.error("Error details:", JSON.stringify(testError));
                daemonStatus = "Daemon is installed but not responding to commands.";
                isLoading = false;
                showInstallButton = true;
                return false;
            }
            
            console.log("Daemon connection successful, setting isLoading to false");
            isLoading = false;
            showInstallButton = false;
            return true;
        } catch (error) {
            console.error("Error connecting to daemon:", error);
            daemonStatus = "Error connecting to daemon. Daemon may not be installed or running.";
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
            
            // Install the daemon
            console.log("Calling InstallDaemonWithOneClick...");
            const installResult = await InstallDaemonWithOneClick();
            
            // Now let's manually start the daemon
            console.log("Attempting to start the daemon...");
            try {
                // Use pkexec to run the daemon with elevated privileges
                const daemonPath = "/usr/bin/free-mind-daemon";
                
                // Create a temporary script to start the daemon
                const tempScript = `
                #!/bin/bash
                ${daemonPath} &
                `;
                
                // Write the script to a temporary file
                const scriptElement = document.createElement('a');
                const scriptBlob = new Blob([tempScript], {type: 'text/plain'});
                scriptElement.href = URL.createObjectURL(scriptBlob);
                const scriptPath = '/tmp/start-daemon.sh';
                scriptElement.download = 'start-daemon.sh';
                document.body.appendChild(scriptElement);
                scriptElement.click();
                document.body.removeChild(scriptElement);
                
                console.log("Created temporary script to start daemon");
                
                // Alert the user to run the script with elevated privileges
                alert("Please run the following command in a terminal with sudo privileges:\n\nchmod +x /tmp/start-daemon.sh && sudo /tmp/start-daemon.sh");
            } catch (startError) {
                console.error("Error trying to start daemon:", startError);
            }
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
            console.error("Error details:", JSON.stringify(error));
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
            <p class="mb-4 text-sm text-gray-600">{daemonStatus}</p>
        {/if}
        <div class="flex justify-center gap-4">
            <button
                class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
                on:click={sendStartCommand}
                disabled={isLoading || showInstallButton}
            >
                Start
            </button>
            <button
                class="h-10 rounded-md px-6 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
                on:click={sendStopCommand}
                disabled={isLoading || showInstallButton}
            >
                Stop
            </button>
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
