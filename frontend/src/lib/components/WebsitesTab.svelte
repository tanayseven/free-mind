<script lang="ts">
    import { Switch } from "@/components/ui/switch";
    import * as Table from "@/components/ui/table";
    import * as Alert from "$lib/components/ui/alert/index.js";
    import { Search, Pencil, Check, X, Trash2, Plus } from "@lucide/svelte";
    import InfoIcon from "@lucide/svelte/icons/info";

    export type WebsiteEntry = {
        id: string;
        domain: string;
        category: string;
        enabled: boolean;
    };

    const CATEGORIES = ["Social", "Video", "News", "Gaming", "Shopping", "Other"];

    let {
        websites = $bindable<WebsiteEntry[]>([
            { id: "1", domain: "youtube.com", category: "Video", enabled: true },
            { id: "2", domain: "www.youtube.com", category: "Video", enabled: true },
            { id: "3", domain: "facebook.com", category: "Social", enabled: true },
            { id: "4", domain: "www.facebook.com", category: "Social", enabled: true },
            { id: "5", domain: "instagram.com", category: "Social", enabled: true },
            { id: "6", domain: "www.instagram.com", category: "Social", enabled: true },
        ]),
        isBlocking = false,
    }: {
        websites?: WebsiteEntry[];
        isBlocking?: boolean;
    } = $props();

    let filterText = $state("");
    let editingId = $state<string | null>(null);
    let editingDomain = $state("");
    let selectedIds = $state(new Set<string>());
    let editInput = $state<HTMLInputElement | null>(null);

    let filtered = $derived(
        filterText.trim() === ""
            ? websites
            : websites.filter((w) => {
                  const q = filterText.toLowerCase();
                  return (
                      w.domain.toLowerCase().includes(q) ||
                      w.category.toLowerCase().includes(q) ||
                      (w.enabled ? "enabled" : "disabled").includes(q)
                  );
              })
    );

    let allSelected = $derived(
        filtered.length > 0 && filtered.every((w) => selectedIds.has(w.id))
    );
    let someSelected = $derived(
        filtered.some((w) => selectedIds.has(w.id)) && !allSelected
    );

    function toggleSelectAll() {
        if (allSelected) {
            const next = new Set(selectedIds);
            filtered.forEach((w) => next.delete(w.id));
            selectedIds = next;
        } else {
            const next = new Set(selectedIds);
            filtered.forEach((w) => next.add(w.id));
            selectedIds = next;
        }
    }

    function toggleSelect(id: string) {
        const next = new Set(selectedIds);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        selectedIds = next;
    }

    function startEdit(entry: WebsiteEntry) {
        editingId = entry.id;
        editingDomain = entry.domain;
        // Focus after Svelte renders
        setTimeout(() => editInput?.focus(), 0);
    }

    function commitEdit() {
        if (!editingId) return;
        const trimmed = editingDomain.trim();
        if (trimmed) {
            websites = websites.map((w) =>
                w.id === editingId ? { ...w, domain: trimmed } : w
            );
        }
        editingId = null;
        editingDomain = "";
    }

    function cancelEdit() {
        editingId = null;
        editingDomain = "";
    }

    function deleteEntry(id: string) {
        websites = websites.filter((w) => w.id !== id);
        const next = new Set(selectedIds);
        next.delete(id);
        selectedIds = next;
    }

    function deleteSelected() {
        const toDelete = new Set(selectedIds);
        websites = websites.filter((w) => !toDelete.has(w.id));
        selectedIds = new Set();
    }

    function addEntry() {
        const id = crypto.randomUUID();
        const entry: WebsiteEntry = { id, domain: "example.com", category: "Other", enabled: true };
        websites = [...websites, entry];
        // Clear filter so the new row is visible
        filterText = "";
        startEdit(entry);
    }

    function updateCategory(id: string, category: string) {
        websites = websites.map((w) => (w.id === id ? { ...w, category } : w));
    }

    function toggleEnabled(id: string, enabled: boolean) {
        websites = websites.map((w) => (w.id === id ? { ...w, enabled } : w));
    }
</script>

<div class="flex flex-col h-full p-4 gap-3 overflow-hidden">
    {#if isBlocking}
        <Alert.Root variant="warning" class="shrink-0">
            <InfoIcon />
            <Alert.Title>Changes apply to your next session</Alert.Title>
            <Alert.Description>
                A blocking session is currently active. Any edits you make here will take effect when the next session starts.
            </Alert.Description>
        </Alert.Root>
    {/if}

    <!-- Filter bar + actions -->
    <div class="flex items-center gap-2 shrink-0">
        <div class="relative flex-1">
            <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground pointer-events-none" />
            <input
                type="text"
                placeholder="Filter by website, category, or status…"
                bind:value={filterText}
                class="w-full rounded-md border border-input bg-background pl-8 pr-3 py-1.5 text-sm outline-none focus:ring-1 focus:ring-ring placeholder:text-muted-foreground/60"
            />
        </div>
        {#if selectedIds.size > 0}
            <button
                onclick={deleteSelected}
                class="h-8 px-3 inline-flex items-center gap-1.5 rounded-md border border-destructive/60 text-destructive text-xs font-medium hover:bg-destructive/10 transition-colors shrink-0"
            >
                <Trash2 class="size-3.5" />
                Delete ({selectedIds.size})
            </button>
        {/if}
        <button
            onclick={addEntry}
            class="h-8 px-3 inline-flex items-center gap-1.5 rounded-md bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 transition-colors shrink-0"
        >
            <Plus class="size-3.5" />
            Add
        </button>
    </div>

    <!-- Table -->
    <div class="flex-1 overflow-auto rounded-md border border-border">
        <Table.Root>
            <Table.Header>
                <Table.Row class="bg-muted/40 hover:bg-muted/40">
                    <Table.Head class="w-10 px-3">
                        <input
                            type="checkbox"
                            checked={allSelected}
                            bind:indeterminate={someSelected}
                            onchange={toggleSelectAll}
                            class="size-4 rounded border-input accent-primary cursor-pointer"
                        />
                    </Table.Head>
                    <Table.Head>Website</Table.Head>
                    <Table.Head>Category</Table.Head>
                    <Table.Head class="text-center w-24">Enabled</Table.Head>
                    <Table.Head class="w-10"></Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#each filtered as entry (entry.id)}
                    <Table.Row class="group">
                        <Table.Cell class="px-3">
                            <input
                                type="checkbox"
                                checked={selectedIds.has(entry.id)}
                                onchange={() => toggleSelect(entry.id)}
                                class="size-4 rounded border-input accent-primary cursor-pointer"
                            />
                        </Table.Cell>
                        <Table.Cell class="min-w-48">
                            {#if editingId === entry.id}
                                <div class="flex items-center gap-1">
                                    <input
                                        bind:this={editInput}
                                        type="text"
                                        bind:value={editingDomain}
                                        onkeydown={(e) => {
                                            if (e.key === "Enter") commitEdit();
                                            if (e.key === "Escape") cancelEdit();
                                        }}
                                        class="flex-1 min-w-0 rounded border border-input bg-background px-2 py-0.5 text-sm outline-none focus:ring-1 focus:ring-ring"
                                    />
                                    <button
                                        onclick={commitEdit}
                                        class="text-primary hover:text-primary/80 transition-colors shrink-0"
                                        title="Save"
                                    >
                                        <Check class="size-3.5" />
                                    </button>
                                    <button
                                        onclick={cancelEdit}
                                        class="text-muted-foreground hover:text-foreground transition-colors shrink-0"
                                        title="Cancel"
                                    >
                                        <X class="size-3.5" />
                                    </button>
                                </div>
                            {:else}
                                <div class="flex items-center gap-1.5">
                                    <span class="text-sm">{entry.domain}</span>
                                    <button
                                        onclick={() => startEdit(entry)}
                                        class="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-foreground transition-opacity shrink-0"
                                        title="Edit"
                                    >
                                        <Pencil class="size-3" />
                                    </button>
                                </div>
                            {/if}
                        </Table.Cell>
                        <Table.Cell>
                            <select
                                value={entry.category}
                                onchange={(e) => updateCategory(entry.id, e.currentTarget.value)}
                                class="rounded border border-input bg-background px-2 py-1 text-xs outline-none focus:ring-1 focus:ring-ring cursor-pointer"
                            >
                                {#each CATEGORIES as cat}
                                    <option value={cat}>{cat}</option>
                                {/each}
                            </select>
                        </Table.Cell>
                        <Table.Cell class="text-center">
                            <Switch
                                checked={entry.enabled}
                                onCheckedChange={(checked) => toggleEnabled(entry.id, checked)}
                                size="sm"
                            />
                        </Table.Cell>
                        <Table.Cell>
                            <button
                                onclick={() => deleteEntry(entry.id)}
                                class="text-muted-foreground hover:text-destructive transition-colors"
                                title="Delete"
                            >
                                <Trash2 class="size-3.5" />
                            </button>
                        </Table.Cell>
                    </Table.Row>
                {/each}
                {#if filtered.length === 0}
                    <Table.Row class="hover:bg-transparent">
                        <Table.Cell colspan={5} class="text-center py-10 text-muted-foreground text-sm">
                            {filterText ? "No websites match your filter." : "No websites added yet."}
                        </Table.Cell>
                    </Table.Row>
                {/if}
            </Table.Body>
        </Table.Root>
    </div>

    <!-- Footer count -->
    <p class="text-xs text-muted-foreground shrink-0">
        {filtered.length} website{filtered.length !== 1 ? "s" : ""}
        {#if filterText} matching · {/if}
        {websites.filter((w) => w.enabled).length} enabled
    </p>
</div>
