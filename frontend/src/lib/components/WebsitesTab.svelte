<script lang="ts">
	import { Switch } from '@/components/ui/switch';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Search, Pencil, Check, X, Trash2, Plus, ChevronRight } from '@lucide/svelte';
	import InfoIcon from '@lucide/svelte/icons/info';
	import { SvelteSet } from 'svelte/reactivity';
	import { normalizeDomain, parseDomainList } from '$lib/domains';

	export type WebsiteEntry = {
		id: string;
		domain: string;
		category: string;
		enabled: boolean;
	};

	let {
		websites = $bindable<WebsiteEntry[]>([]),
		isBlocking = false
	}: {
		websites?: WebsiteEntry[];
		isBlocking?: boolean;
	} = $props();

	const BASE_CATEGORIES = ['Social', 'News', 'Video', 'Gaming', 'Shopping', 'Other'];
	let categories = $derived([
		...BASE_CATEGORIES,
		...[...new Set(websites.map((w) => w.category))]
			.filter((c) => !BASE_CATEGORIES.includes(c))
			.sort()
	]);

	let filterText = $state('');
	let collapsed = new SvelteSet<string>();

	let newDomains = $state('');
	let newCategory = $state('Other');
	let addError = $state('');

	let editingId = $state<string | null>(null);
	let editingDomain = $state('');
	let editError = $state('');
	let editInput = $state<HTMLInputElement | null>(null);

	let filtered = $derived.by(() => {
		const q = filterText.trim().toLowerCase();
		if (!q) return websites;
		return websites.filter((w) => w.domain.includes(q) || w.category.toLowerCase().includes(q));
	});

	// Categories with at least one (filtered) site, in display order.
	let groups = $derived(
		categories
			.map((category) => ({
				category,
				entries: filtered.filter((w) => w.category === category)
			}))
			.filter((g) => g.entries.length > 0)
	);

	let enabledCount = $derived(websites.filter((w) => w.enabled).length);

	function addEntries() {
		const { valid, invalid } = parseDomainList(newDomains);
		const existing = new Set(websites.map((w) => w.domain));
		const fresh = valid.filter((d) => !existing.has(d));
		websites = [
			...websites,
			...fresh.map((domain) => ({
				id: crypto.randomUUID(),
				domain,
				category: newCategory,
				enabled: true
			}))
		];
		const notes: string[] = [];
		if (invalid.length) notes.push(`Not a valid website: ${invalid.join(', ')}`);
		if (valid.length > fresh.length) notes.push('Some websites were already in the list.');
		addError = notes.join(' ');
		// Keep invalid input in the box so it can be fixed.
		newDomains = invalid.join(' ');
		collapsed.delete(newCategory);
		filterText = '';
	}

	function setCategoryEnabled(category: string, enabled: boolean) {
		websites = websites.map((w) => (w.category === category ? { ...w, enabled } : w));
	}

	function toggleCollapsed(category: string) {
		if (collapsed.has(category)) collapsed.delete(category);
		else collapsed.add(category);
	}

	function startEdit(entry: WebsiteEntry) {
		editingId = entry.id;
		editingDomain = entry.domain;
		editError = '';
		// Focus after Svelte renders
		setTimeout(() => editInput?.focus(), 0);
	}

	function commitEdit() {
		if (!editingId) return;
		const domain = normalizeDomain(editingDomain);
		if (domain === null) {
			editError = 'Not a valid website';
			return;
		}
		if (websites.some((w) => w.id !== editingId && w.domain === domain)) {
			editError = 'Already in the list';
			return;
		}
		websites = websites.map((w) => (w.id === editingId ? { ...w, domain } : w));
		cancelEdit();
	}

	function cancelEdit() {
		editingId = null;
		editingDomain = '';
		editError = '';
	}

	function deleteEntry(id: string) {
		websites = websites.filter((w) => w.id !== id);
	}

	function toggleEnabled(id: string, enabled: boolean) {
		websites = websites.map((w) => (w.id === id ? { ...w, enabled } : w));
	}
</script>

<div class="flex min-h-0 flex-1 flex-col gap-3 p-4">
	{#if isBlocking}
		<Alert.Root class="shrink-0">
			<InfoIcon />
			<Alert.Title>Changes apply to your next session</Alert.Title>
			<Alert.Description>Edits take effect the next time you start blocking.</Alert.Description>
		</Alert.Root>
	{/if}

	<!-- Add websites -->
	<form
		class="flex shrink-0 flex-col gap-1.5"
		onsubmit={(e) => {
			e.preventDefault();
			addEntries();
		}}
	>
		<div class="flex items-center gap-2">
			<input
				type="text"
				placeholder="Add websites, e.g. reddit.com — paste several separated by commas or spaces"
				bind:value={newDomains}
				aria-label="Websites to add"
				class="min-w-0 flex-1 rounded-md border border-input bg-background px-3 py-1.5 text-sm outline-none placeholder:text-muted-foreground/60 focus:ring-1 focus:ring-ring"
			/>
			<select
				bind:value={newCategory}
				aria-label="Category"
				class="rounded-md border border-input bg-background py-1.5 pr-8 pl-2 text-sm outline-none focus:ring-1 focus:ring-ring"
			>
				{#each categories as cat (cat)}
					<option value={cat}>{cat}</option>
				{/each}
			</select>
			<button
				type="submit"
				disabled={!newDomains.trim()}
				class="inline-flex h-8 shrink-0 items-center gap-1.5 rounded-md bg-primary px-3 text-xs font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
			>
				<Plus class="size-3.5" />
				Add
			</button>
		</div>
		{#if addError}
			<p class="text-xs text-destructive">{addError}</p>
		{/if}
	</form>

	<!-- Filter -->
	<div class="relative shrink-0">
		<Search
			class="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground"
		/>
		<input
			type="text"
			placeholder="Filter by website or category…"
			bind:value={filterText}
			aria-label="Filter websites"
			class="w-full rounded-md border border-input bg-background py-1.5 pr-3 pl-8 text-sm outline-none placeholder:text-muted-foreground/60 focus:ring-1 focus:ring-ring"
		/>
	</div>

	<!-- Grouped list -->
	<div class="min-h-0 flex-1 overflow-auto rounded-md border border-border">
		{#each groups as group (group.category)}
			{@const all = websites.filter((w) => w.category === group.category)}
			{@const on = all.filter((w) => w.enabled).length}
			{@const open = !collapsed.has(group.category) || filterText.trim() !== ''}
			<section class="border-b border-border last:border-b-0">
				<div class="sticky top-0 z-10 flex items-center gap-2 bg-muted px-3 py-2">
					<button
						onclick={() => toggleCollapsed(group.category)}
						aria-expanded={open}
						class="flex flex-1 items-center gap-1.5 text-left text-sm font-medium"
					>
						<ChevronRight class="size-4 transition-transform {open ? 'rotate-90' : ''}" />
						{group.category}
						<span class="text-xs font-normal text-muted-foreground">
							{on} of {all.length} on
						</span>
					</button>
					<Switch
						checked={on > 0}
						onCheckedChange={(checked) => setCategoryEnabled(group.category, checked)}
						aria-label="Block all {group.category} websites"
					/>
				</div>
				{#if open}
					<ul>
						{#each group.entries as entry (entry.id)}
							<li
								class="group flex items-center gap-3 border-t border-border/60 px-3 py-1.5 pl-9 hover:bg-muted/30"
							>
								{#if editingId === entry.id}
									<div class="flex min-w-0 flex-1 items-center gap-1.5">
										<input
											bind:this={editInput}
											type="text"
											bind:value={editingDomain}
											onkeydown={(e) => {
												if (e.key === 'Enter') commitEdit();
												if (e.key === 'Escape') cancelEdit();
											}}
											aria-label="Website"
											class="min-w-0 flex-1 rounded border border-input bg-background px-2 py-0.5 text-sm outline-none focus:ring-1 focus:ring-ring"
										/>
										{#if editError}
											<span class="shrink-0 text-xs text-destructive">{editError}</span>
										{/if}
										<button
											onclick={commitEdit}
											class="shrink-0 text-primary transition-colors hover:text-primary/80"
											title="Save"
										>
											<Check class="size-4" />
										</button>
										<button
											onclick={cancelEdit}
											class="shrink-0 text-muted-foreground transition-colors hover:text-foreground"
											title="Cancel"
										>
											<X class="size-4" />
										</button>
									</div>
								{:else}
									<span
										class="min-w-0 flex-1 truncate text-sm {entry.enabled
											? ''
											: 'text-muted-foreground line-through decoration-muted-foreground/50'}"
									>
										{entry.domain}
									</span>
									<button
										onclick={() => startEdit(entry)}
										class="shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-foreground focus-visible:opacity-100"
										title="Edit"
									>
										<Pencil class="size-3.5" />
									</button>
								{/if}
								<Switch
									checked={entry.enabled}
									onCheckedChange={(checked) => toggleEnabled(entry.id, checked)}
									aria-label="Block {entry.domain}"
								/>
								<button
									onclick={() => deleteEntry(entry.id)}
									class="ml-3 shrink-0 text-muted-foreground transition-colors hover:text-destructive"
									title="Remove {entry.domain}"
								>
									<Trash2 class="size-3.5" />
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		{:else}
			<p class="py-10 text-center text-sm text-muted-foreground">
				{filterText ? 'No websites match your filter.' : 'No websites added yet.'}
			</p>
		{/each}
	</div>

	<p class="shrink-0 text-xs text-muted-foreground">
		{websites.length} website{websites.length !== 1 ? 's' : ''} · {enabledCount} blocked
		{#if filterText.trim()}
			· {filtered.length} matching
		{/if}
		· <span class="text-muted-foreground/80">www. versions are blocked too</span>
	</p>
</div>
