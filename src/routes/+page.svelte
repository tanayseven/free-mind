<script lang="ts">

  import {invoke} from '@tauri-apps/api/tauri';
  import {Button} from "$lib/components/ui/button";
  import {toggleMode} from "mode-watcher";
  import {Moon, Settings, Sun} from "lucide-svelte";
  import {Label} from "$lib/components/ui/label";
  import {Switch} from "$lib/components/ui/switch";
  import {Input} from "$lib/components/ui/input";
  import {loadSettings, saveSettings} from "../settings";
  import {goto} from "$app/navigation";

  const hostsFile = '/etc/hosts';
  const possibleSubdomains = ["www", "news", "blog", "web"]
  const distractingSites = ['facebook.com', 'twitter.com', 'instagram.com', 'reddit.com', 'x.com', 'linkedin.com', 'youtube.com', 'whatsapp.com'];
  const start = '#-freemind-blacklist-start-#';
  const end = '#-freemind-blacklist-end-#';
  const redirectUrl = '127.0.0.1';

  const areSitesBlocked = (hostsFileContents: string) => {
    return hostsFileContents.includes(start) && hostsFileContents.includes(end);
  }

  const addBlockedSites = (hostsFileContents: string) => {
    if (areSitesBlocked(hostsFileContents)) return hostsFileContents;
    const lines = hostsFileContents.split('\n');
    let newLines = [start];
    for (let distractingSite of distractingSites) {
      const line = `${redirectUrl} ${distractingSite}`;
      newLines.push(line);
      for (let subdomain of possibleSubdomains) {
        const line = `${redirectUrl} ${([subdomain, distractingSite].join('.'))}`;
        newLines.push(line);
      }
    }
    newLines.push(end);
    newLines.push('');
    return lines.concat(newLines).join('\n');
  }

  const removeBlockedSites = (hostsFileContents: string) => {
    if (!areSitesBlocked(hostsFileContents)) return hostsFileContents;
    const lines = hostsFileContents.split('\n');
    let newLines = [];
    let isInsideBlockedSitesSection = false;
    for (let line of lines) {
      if (line.includes(start)) {
        isInsideBlockedSitesSection = true;
        continue;
      }
      if (line.includes(end)) {
        isInsideBlockedSitesSection = false;
        continue;
      }
      if (!isInsideBlockedSitesSection) {
        newLines.push(line);
      }
    }
    if (newLines[newLines.length - 1] === '') newLines.pop();
    return newLines.join('\n');
  }

  const startFocus = async () => {
    console.log("Starting focus time")
    try {
      const hostContents: string = await invoke('read_file_contents', { filePath: hostsFile });
      const modifiedHostContents = addBlockedSites(hostContents)
      await invoke('write_file_contents', { filePath: hostsFile, contents: modifiedHostContents });
      await invoke('restart_network');
      focusMode = true;
    } catch (error) {
      console.error('Error reading hosts file:', error);
    }
  }

  type TimeRemaining = {
    minutes: number;
    seconds: number;
  }

  const timeRemaining: TimeRemaining = {
    minutes: 0,
    seconds: 0
  }

  const stopFocus = async () => {
    console.log("Stopping focus time")
    try {
      const hostContents: string = await invoke('read_file_contents', { filePath: hostsFile });
      const modifiedHostContents = removeBlockedSites(hostContents)
      await invoke('write_file_contents', { filePath: hostsFile, contents: modifiedHostContents });
      await invoke('restart_network');
      focusMode = false;
    } catch (error) {
      console.error('Error reading hosts file:', error);
    }
  }

  const isFocusEnabled = async () => {
    console.log("Checking focus time")
    try {
      const hostContents: string = await invoke('read_file_contents', { filePath: hostsFile });
      return areSitesBlocked(hostContents);
    } catch (error) {
      console.error('Error reading hosts file:', error);
      return false;
    }
  }

  let focusMode = false;
  const secondsInMinute = 60;
  let timerValue = 30;
  let isTimerRunning = false;
  let timeRemainingInSeconds = 0;

  const startTimer = () => {
    isTimerRunning = true;
    timeRemainingInSeconds = timerValue * secondsInMinute;
    startFocus();
    const interval = setInterval(() => {
      timeRemainingInSeconds -= 1;
      timeRemaining.minutes = Math.floor(timeRemainingInSeconds / secondsInMinute);
      timeRemaining.seconds = timeRemainingInSeconds % secondsInMinute;
      if (timeRemainingInSeconds <= 0) {
        clearInterval(interval);
        stopFocus();
        isTimerRunning = false;
      }
    }, 1000);
  }

  const toggleFocus = async () => {
    console.log("Toggling focus time")
    try {
      if (await isFocusEnabled()) {
        await stopFocus();
      } else {
        await startFocus();
      }
    } catch (error) {
      console.error('Error reading hosts file:', error);
    }
  }

  const isRoot = async () => {
    try {
      console.log("Checking if root...")
      const isElevated = await invoke('is_elevated');
      console.log("Is elevated: ", isElevated)
      return isElevated as boolean;
    } catch (error) {
      console.error('Error checking elevated privileges:', error);
      return false;
    }
  };

  const onLoad = async () => {
    focusMode = await isFocusEnabled();
    const currentSettings = await loadSettings();
    timerValue = currentSettings.timerValue;
  }

  const saveAllPreferences = async () => {
    const currentSettings = await loadSettings();
    console.log(`Current settings ${currentSettings}`)
    const newSettings = {
      ...currentSettings,
      timerValue,
    }
    await saveSettings(newSettings);
  }

</script>

<div class="container h-full p-0" use:onLoad>
  <div class="container p-4 flex flex-col space-y-10">
    <div class="grid grid-cols-[1fr_0rem]">
      <h1 class="grid self-center text-4xl font-extrabold tracking-tight lg:text-5xl justify-center">Free Mind</h1>
      <Button class="flex place-self-end content-center" on:click={toggleMode} variant="outline" size="icon">
        <Sun
          class="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0"
        />
        <Moon
          class="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
        />
        <span class="sr-only">Toggle theme</span>
      </Button>
    </div>

    {#await isRoot()}
      <div class="flex justify-center">
        <h3 class="text-yellow-700">Checking root privileges...</h3>
      </div>
    {:then isRoot}
      {#if !isRoot}
        <div class="flex justify-center text-red-700">
          <h3>Root privileges are not enabled, this app will not work, please restart as root!</h3>
        </div>
      {/if}
      <div class="flex items-center justify-center space-x-2">
        <Switch id="focus-mode" on:click={toggleFocus} disabled={!isRoot} bind:checked={focusMode} />
        <Label class="justify-center text-xl" for="focus-mode">Focus Mode</Label>
      </div>
    {:catch error}
      <h3>Error checking elevated privileges: {error}</h3>
    {/await}
    <div class="flex items-center justify-center">
      <Label for="till-date" class="text-xl mr-4">Focus for:</Label>
      <Input id="till-date" type="number" bind:value={timerValue} class="w-20 py-6 text-xl" disabled={isTimerRunning} on:change={saveAllPreferences} />
      <Label for="till-date" class="text-xl ml-2 mr-4">minutes</Label>
    </div>
    <div class="flex items-center justify-center">
      <Button class="text-xl" variant="outline" on:click={startTimer} disabled={isTimerRunning}>Start Timer</Button>
    </div>
    <div class="flex items-center justify-end space-x-2">
      <Button variant="outline" class="text-xl" on:click={()=>goto('/preferences')}>
        <Settings class="mr-2 h-6 w-6" />
        Preferences
      </Button>
    </div>
  </div>
  <div class="absolute bottom-0">
    <Label for="time-remaining" class="text-m ml-2 mr-4">Time remaining:</Label>
    <span id="time-remaining" class="text-muted-foreground text-m w-5 text-center">{timeRemaining.minutes}</span>
    <Label for="time-remaining" class="text-m ml-2 mr-4">minutes</Label>
    <span id="time-remaining" class="text-muted-foreground text-m w-5 text-center">{timeRemaining.seconds}</span>
    <Label for="time-remaining" class="text-m ml-2 mr-4">seconds</Label>
  </div>
</div>

<style>

</style>
