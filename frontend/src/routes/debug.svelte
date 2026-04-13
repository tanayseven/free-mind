<script lang="ts">
	import { onMount } from 'svelte';

	let debugInfo = '';
	let isDebugging = false;

	async function runDiagnostic() {
		isDebugging = true;
		debugInfo = 'Running diagnostics...\n\n';

		try {
			// Check if we can access the Wails runtime
			debugInfo += 'Checking Wails runtime availability...\n';
			if (typeof window !== 'undefined' && window['go'] && window['go']['main']) {
				debugInfo += '✓ Wails runtime available\n';
			} else {
				debugInfo += '✗ Wails runtime not available\n';
			}

			// Try to call a simple function
			debugInfo += '\nTesting basic functionality...\n';
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
	<title>Debug - Free Mind</title>
</svelte:head>

<div class="mx-auto max-w-4xl p-4">
	<h1 class="mb-4 text-2xl font-bold">Free Mind Debug Information</h1>

	<div class="mb-4 rounded-lg bg-muted p-4">
		<button
			on:click={runDiagnostic}
			disabled={isDebugging}
			class="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
		>
			{#if isDebugging}Running Diagnostic...{:else}Run Diagnostic{/if}
		</button>
	</div>

	<div class="rounded-lg bg-gray-800 p-4 font-mono text-sm whitespace-pre-wrap text-green-400">
		{debugInfo}
	</div>
</div>
