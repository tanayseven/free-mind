<script lang="ts">
    import { onMount } from "svelte";
    
    let debugInfo = "";
    let isDebugging = false;
    
    async function runDiagnostic() {
        isDebugging = true;
        debugInfo = "Running diagnostics...\n\n";
        
        try {
            // Check if we can access the Wails runtime
            debugInfo += "Checking Wails runtime availability...\n";
            if (typeof window !== 'undefined' && window['go'] && window['go']['main']) {
                debugInfo += "✓ Wails runtime available\n";
            } else {
                debugInfo += "✗ Wails runtime not available\n";
            }
            
            // Try to call a simple function
            debugInfo += "\nTesting basic functionality...\n";
            try {
                const result = await window['go']['main']['App']['CheckDaemonInstalled']();
                debugInfo += `✓ CheckDaemonInstalled returned: ${result}\n`;
            } catch (error) {
                debugInfo += `✗ Error calling CheckDaemonInstalled: ${error}\n`;
            }
            
            // Try to get daemon binary path
            try {
                const result = await window['go']['main']['App']['GetDaemonBinaryPath']();
                debugInfo += `✓ GetDaemonBinaryPath returned: ${result}\n`;
            } catch (error) {
                debugInfo += `✗ Error calling GetDaemonBinaryPath: ${error}\n`;
            }
            
            // Try to get hosts file path
            try {
                const result = await window['go']['main']['App']['HostsFilePath']();
                debugInfo += `✓ HostsFilePath returned: ${result}\n`;
            } catch (error) {
                debugInfo += `✗ Error calling HostsFilePath: ${error}\n`;
            }
            
        } catch (error) {
            debugInfo += `\nError during diagnostics: ${error}\n`;
        }
        
        isDebugging = false;
    }
    
    onMount(() => {
        runDiagnostic();
    });
</script>

<svelte:head>
    <title>Debug - Free-Mind</title>
</svelte:head>

<div class="p-4 max-w-4xl mx-auto">
    <h1 class="text-2xl font-bold mb-4">Free-Mind Debug Information</h1>
    
    <div class="mb-4 p-4 bg-muted rounded-lg">
        <button 
            on:click={runDiagnostic}
            disabled={isDebugging}
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
            {#if isDebugging}Running Diagnostic...{/if}
            {:else}Run Diagnostic{/if}
        </button>
    </div>
    
    <div class="p-4 bg-gray-800 text-green-400 font-mono text-sm rounded-lg whitespace-pre-wrap">
        {debugInfo}
    </div>
</div>