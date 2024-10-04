<script lang="ts">
import {Button} from "$lib/components/ui/button";
import {goto} from "$app/navigation";
import {ArrowBigLeft, Moon, type Settings, Sun} from "lucide-svelte";
import {toggleMode} from "mode-watcher";
import {Separator} from "$lib/components/ui/separator";
import {ScrollArea} from "$lib/components/ui/scroll-area";
import {loadSettings} from "../../settings";

let selectedWebsiteGroup = "";
let websiteGroups: Record<string, string[]>

const loadWebsiteGroups = async () => {
        console.log("loading website groups")
        const settings = await loadSettings()
        websiteGroups = settings.websiteBlockList
}

</script>
<div class="container h-full p-0">
        {#await loadWebsiteGroups()}
        {:then _}
        <div class="container p-4 flex flex-col space-y-10">
                <div class="grid grid-cols-[0rem_1fr_0rem]">
                        <Button variant="outline" class="flex content-center scale-100" size="icon" on:click={()=>goto('/')}>
                                <ArrowBigLeft class="h-6 w-6" />
                                <span class="sr-only">Go back</span>
                        </Button>
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
                <div>
                        <h2 class="text-xl font-bold">Preferences</h2>
                        <Separator class="my-4" />
                        <div class="grid grid-cols-[0.5fr_0.5fr] px-6 ml-14">
                                <div class="flex flex-col">
                                        <label for="timer" class="text-lg">Block Group</label>
                                        <ScrollArea  class="h-72 w-48 rounded-md border">
                                                <div class="p-4">
                                                        {#each Object.keys(websiteGroups) as websiteGroup}
                                                                <div class="text-sm">
                                                                        {#if selectedWebsiteGroup === websiteGroup}
                                                                                <Button variant="ghost" size="sm" disabled>{websiteGroup}</Button>
                                                                        {:else}
                                                                                <Button variant="outline" size="sm" on:click={() => selectedWebsiteGroup = websiteGroup}>{websiteGroup}</Button>
                                                                        {/if}

                                                                </div>
                                                                <Separator class="my-2" />
                                                        {/each}
                                                </div>
                                        </ScrollArea>
                                </div>
                                <div class="flex flex-col">
                                        <label for="timer" class="text-lg">Block Sites {#if selectedWebsiteGroup !== ''} in group: {selectedWebsiteGroup}{/if}</label>
                                        {#if selectedWebsiteGroup !== ''}
                                        <ScrollArea  class="h-72 w-48 rounded-md border">
                                                <div class="p-4">
                                                        {#each websiteGroups[selectedWebsiteGroup] as website}
                                                                <div class="text-sm">
                                                                        <Button variant="outline" size="sm">{website}</Button>
                                                                </div>
                                                                <Separator class="my-2" />
                                                        {/each}
                                                </div>
                                        </ScrollArea>
                                        {/if}
                                </div>
                        </div>
                </div>
        </div>
        {/await}
</div>