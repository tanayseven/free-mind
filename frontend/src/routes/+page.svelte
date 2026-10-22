<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		ConnectToDaemon,
		SendBlockList,
		StartBlocking,
		StopBlocking,
		InstallAndStartDaemon,
		CheckDaemonInstalled,
		CheckBlocking,
		LoadBlockedWebsites,
		SaveBlockedWebsites,
		LoadSettings
	} from '$lib/api';
	import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
	import {
		House,
		List,
		Settings,
		Infinity as InfinityIcon,
		Timer,
		CalendarClock,
		AlarmClock,
		ShieldCheck
	} from '@lucide/svelte';
	import { applyTheme, loadThemePreference, type ThemePreference } from '$lib/theme';
	import { collapseWww } from '$lib/domains';
	import StatusDot from '$lib/components/modes/StatusDot.svelte';
	import FreeMode from '$lib/components/modes/FreeMode.svelte';
	import TimerMode from '$lib/components/modes/TimerMode.svelte';
	import ScheduleMode from '$lib/components/modes/ScheduleMode.svelte';
	import PomodoroMode from '$lib/components/modes/PomodoroMode.svelte';
	import WebsitesTab, { type WebsiteEntry } from '$lib/components/WebsitesTab.svelte';
	import SettingsTab from '$lib/components/SettingsTab.svelte';

	const modes = [
		{ value: 'free', label: 'Free', icon: InfinityIcon },
		{ value: 'timer', label: 'Timer', icon: Timer },
		{ value: 'schedule', label: 'Schedule', icon: CalendarClock },
		{ value: 'pomodoro', label: 'Pomodoro', icon: AlarmClock }
	] as const;

	let daemonStatus = $state('Starting Free Mind…');
	let isLoading = $state(true);
	let showInstallButton = $state(false);
	let isBlocking = $state(false);
	let theme = $state<ThemePreference>('system');
	let selectedMode = $state('free');
	let websites = $state<WebsiteEntry[]>([]);
	let websitesReady = $state(false);
	let unblockWaiting = $state(30);

	$effect(() => {
		if (!websitesReady) return;
		const json = JSON.stringify(websites);
		SaveBlockedWebsites(json).catch((e: unknown) =>
			console.error('Failed to save blocked websites:', e)
		);
	});

	const darkQuery = () => window.matchMedia('(prefers-color-scheme: dark)');

	function setTheme(preference: ThemePreference) {
		theme = preference;
		applyTheme(preference, darkQuery().matches, document.documentElement.classList, localStorage);
	}

	let enabledCount = $derived(websites.filter((w) => w.enabled).length);

	// When the current blocking session started. Kept in storage so the elapsed
	// time survives closing the window or starting from the tray.
	let blockingSince = $state<number | null>(null);
	let now = $state(Date.now());

	$effect(() => {
		if (isLoading || showInstallButton) return;
		if (isBlocking && blockingSince === null) {
			const saved = Number(localStorage.getItem('blockingSince'));
			blockingSince = saved > 0 ? saved : Date.now();
			localStorage.setItem('blockingSince', String(blockingSince));
		} else if (!isBlocking && blockingSince !== null) {
			blockingSince = null;
			localStorage.removeItem('blockingSince');
		}
	});

	$effect(() => {
		const id = setInterval(() => (now = Date.now()), 15_000);
		return () => clearInterval(id);
	});

	function formatElapsed(ms: number): string {
		const minutes = Math.floor(ms / 60_000);
		if (minutes < 1) return 'just now';
		if (minutes < 60) return `for ${minutes} min`;
		return `for ${Math.floor(minutes / 60)} h ${minutes % 60} min`;
	}

	async function checkDaemonConnection() {
		try {
			isLoading = true;
			const result = await ConnectToDaemon();

			if (result.includes('Error')) {
				daemonStatus = 'Free Mind needs a one-time permission to block websites.';
				isLoading = false;
				showInstallButton = true;
				return false;
			}

			daemonStatus = '';
			isLoading = false;
			showInstallButton = false;
			isBlocking = await CheckBlocking();
			return true;
		} catch (error) {
			daemonStatus = "Couldn't reach the Free Mind helper: " + error;
			isLoading = false;
			showInstallButton = true;
			return false;
		}
	}

	async function installDaemon() {
		try {
			console.log('Starting installDaemon function');
			isLoading = true;
			showInstallButton = false;
			daemonStatus = 'Waiting for permission… Enter your password in the system prompt.';

			console.log('Checking if daemon is already installed...');
			const installedCheck = await CheckDaemonInstalled();
			console.log('Daemon installed check result:', installedCheck);

			// Installed but unreachable or outdated: reinstalling also restarts it.
			if (installedCheck) {
				daemonStatus = 'Updating the Free Mind helper… Enter your password in the system prompt.';
			}

			console.log('Calling InstallAndStartDaemon...');
			const installResult = await InstallAndStartDaemon();

			console.log('Installation result:', installResult);
			if (installResult.includes('Error')) throw installResult;
			daemonStatus = 'Finishing setup…';

			console.log('Setting timeout to check daemon connection after installation...');
			setTimeout(async () => {
				console.log('Timeout elapsed, checking daemon connection...');
				const connected = await checkDaemonConnection();
				console.log('Daemon connection check result:', connected);
				if (connected) {
					console.log('Daemon connected successfully after installation');
					daemonStatus = '';
					isLoading = false;
				} else {
					console.error('Failed to connect to daemon after installation');
					daemonStatus =
						"Setup finished, but Free Mind can't reach its helper. Try again, or restart Free Mind.";
					showInstallButton = true;
					isLoading = false;
				}
			}, 3000);
		} catch (error) {
			console.error('Error installing daemon:', error);
			console.error('Full error details:', JSON.stringify(error));
			daemonStatus = 'Setup did not finish: ' + String(error).replace(/^Error[^:]*:\s*/, '');
			showInstallButton = true;
			isLoading = false;
		}
	}

	onMount(async () => {
		console.log('Component mounted, initializing application...');
		try {
			setTheme(loadThemePreference(localStorage));
			// Follow OS changes while the preference is "system".
			darkQuery().addEventListener('change', () => {
				if (theme === 'system') setTheme('system');
			});

			try {
				const json = await LoadBlockedWebsites();
				websites = collapseWww(JSON.parse(json));
			} catch (e: unknown) {
				console.error('Failed to load blocked websites:', e);
			}
			websitesReady = true;

			try {
				const settings = await LoadSettings();
				unblockWaiting = settings.unblockWaiting;
			} catch (e: unknown) {
				console.error('Failed to load settings:', e);
			}

			console.log('Checking daemon connection on page load...');
			const connected = await checkDaemonConnection();
			console.log('Initial daemon connection check result:', connected);

			if (!connected) {
				console.log('Daemon connection failed, attempting installation');
				await installDaemon();
			} else {
				console.log('Daemon connected successfully on page load');
			}
		} catch (error) {
			console.error('Error during initialization:', error);
			console.error('Full error details:', JSON.stringify(error));
			daemonStatus = 'Free Mind could not start: ' + error;
			isLoading = false;
			showInstallButton = true;
		}
		console.log(
			'Initialization complete. isLoading:',
			isLoading,
			'showInstallButton:',
			showInstallButton
		);
	});

	async function sendStartCommand() {
		try {
			console.log('Sending start command...');
			await ConnectToDaemon();
			const enabledDomains = websites.filter((w) => w.enabled).map((w) => w.domain);
			const blockResult = await SendBlockList(enabledDomains.join(','));
			console.log('SendBlockList result:', blockResult);
			if (blockResult) {
				const startResult = await StartBlocking();
				console.log('StartBlocking result:', startResult);
				if (startResult) isBlocking = true;
			}
		} catch (error) {
			console.error('Error calling commands:', error);
			console.error('Full error details:', JSON.stringify(error));
		}
	}

	async function sendStopCommand() {
		try {
			console.log('Sending stop command...');
			await ConnectToDaemon();
			const stopResult = await StopBlocking();
			console.log('StopBlocking result:', stopResult);
			isBlocking = await CheckBlocking();
		} catch (error) {
			console.error('Error calling Stop command:', error);
			console.error('Full error details:', JSON.stringify(error));
		}
	}

	let pollInterval: ReturnType<typeof setInterval> | null = null;

	$effect(() => {
		if (!isLoading && !showInstallButton) {
			pollInterval = setInterval(async () => {
				isBlocking = await CheckBlocking();
			}, 100);
		} else if (pollInterval !== null) {
			clearInterval(pollInterval);
			pollInterval = null;
		}
	});

	onDestroy(() => {
		if (pollInterval !== null) clearInterval(pollInterval);
	});
</script>

<svelte:head>
	<title>Free Mind</title>
	<meta name="description" content="Block distracting websites and stay focused" />
</svelte:head>

<div class="flex min-h-0 w-full flex-1 flex-col self-stretch">
	{#if isLoading || showInstallButton}
		<div class="flex w-full items-center border-b border-border/50 px-5 py-3">
			<span class="text-lg font-bold tracking-tight">Free Mind</span>
		</div>
		<div class="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
			{#if isLoading}
				<p class="text-sm text-muted-foreground" role="status">{daemonStatus}</p>
			{:else}
				<ShieldCheck class="size-12 text-primary" strokeWidth={1.5} />
				<div class="flex max-w-md flex-col gap-2">
					<h1 class="text-lg font-semibold">One-time setup</h1>
					<p class="text-sm text-muted-foreground">
						Free Mind blocks websites by editing a system file, which needs your permission once.
						After that, you can start and stop blocking without entering your password.
					</p>
					{#if !daemonStatus.startsWith('Free Mind needs')}
						<p class="text-sm text-destructive">{daemonStatus}</p>
					{/if}
				</div>
				<button
					class="inline-flex h-9 items-center justify-center rounded-md bg-primary px-5 text-sm font-medium text-primary-foreground shadow-xs transition-colors hover:bg-primary/90"
					onclick={installDaemon}
				>
					Grant permission
				</button>
			{/if}
		</div>
	{:else}
		<Tabs value="home" class="flex min-h-0 w-full flex-1 flex-col">
			<div class="flex w-full items-center gap-4 border-b border-border/50 px-5 py-3">
				<span class="w-24 shrink-0 text-lg font-bold tracking-tight">Free Mind</span>

				<div class="flex flex-1 justify-center">
					<TabsList variant="line" class="h-10">
						<TabsTrigger value="home" class="h-8 gap-1.5 px-4 text-sm">
							<House class="size-3.5" />
							Home
						</TabsTrigger>
						<TabsTrigger value="websites" class="h-8 gap-1.5 px-4 text-sm">
							<List class="size-3.5" />
							Websites
						</TabsTrigger>
						<TabsTrigger value="settings" class="h-8 gap-1.5 px-4 text-sm">
							<Settings class="size-3.5" />
							Settings
						</TabsTrigger>
					</TabsList>
				</div>

				<StatusDot {isBlocking} />
			</div>

			{#if daemonStatus}
				<p class="mt-2 px-4 text-center text-xs text-destructive">{daemonStatus}</p>
			{/if}

			<TabsContent
				value="home"
				class="flex flex-1 flex-col items-center gap-8 overflow-y-auto px-6 pt-12 pb-6"
			>
				<div class="flex flex-col items-center gap-2">
					<div
						role="radiogroup"
						aria-label="Blocking mode"
						class="inline-flex gap-1 rounded-lg bg-muted p-1"
					>
						{#each modes as mode (mode.value)}
							<button
								role="radio"
								aria-checked={selectedMode === mode.value}
								disabled={isBlocking}
								onclick={() => (selectedMode = mode.value)}
								class="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm font-medium transition-colors disabled:cursor-not-allowed
									{selectedMode === mode.value
									? 'bg-background text-foreground shadow-sm'
									: 'text-muted-foreground enabled:hover:text-foreground disabled:opacity-50'}"
							>
								<mode.icon class="size-4" />
								{mode.label}
							</button>
						{/each}
					</div>
					<p class="h-4 text-xs text-muted-foreground">
						{#if isBlocking}Stop blocking to change the mode.{/if}
					</p>
				</div>

				{#if selectedMode === 'free'}
					<FreeMode
						{isBlocking}
						onStart={sendStartCommand}
						onStop={sendStopCommand}
						{unblockWaiting}
					/>
				{:else if selectedMode === 'timer'}
					<TimerMode {isBlocking} onStart={sendStartCommand} onStop={sendStopCommand} />
				{:else if selectedMode === 'schedule'}
					<ScheduleMode {isBlocking} onStart={sendStartCommand} onStop={sendStopCommand} />
				{:else if selectedMode === 'pomodoro'}
					<PomodoroMode {isBlocking} onStart={sendStartCommand} onStop={sendStopCommand} />
				{/if}

				<p class="text-sm text-muted-foreground" role="status">
					{#if isBlocking}
						Blocking {enabledCount} website{enabledCount !== 1 ? 's' : ''}
						{#if blockingSince !== null}{formatElapsed(now - blockingSince)}{/if}
					{:else if enabledCount === 0}
						No websites selected. Add some in the Websites tab.
					{:else}
						{enabledCount} website{enabledCount !== 1 ? 's' : ''} ready to block
					{/if}
				</p>
			</TabsContent>

			<TabsContent value="websites" class="flex min-h-0 flex-1 flex-col overflow-hidden">
				<WebsitesTab bind:websites {isBlocking} />
			</TabsContent>

			<TabsContent value="settings" class="flex flex-1 flex-col overflow-y-auto">
				<SettingsTab {theme} onThemeChange={setTheme} bind:unblockWaiting />
			</TabsContent>
		</Tabs>
	{/if}
</div>
