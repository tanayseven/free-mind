// Helpers for the website list. The backend adds `www.` variants when it
// blocks, so the list keeps one row per site.

const LABEL = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;

/**
 * Turns user input (a URL or a hostname) into a bare hostname without `www.`,
 * or `null` if it isn't a valid domain.
 */
export function normalizeDomain(input: string): string | null {
	let host = input.trim().toLowerCase();
	host = host.replace(/^[a-z][a-z0-9+.-]*:\/\//, ''); // scheme
	host = host.split(/[/?#]/, 1)[0]; // path, query, fragment
	host = host.replace(/:\d+$/, ''); // port
	host = host.replace(/\.$/, '');
	host = host.replace(/^www\./, '');
	if (host.length === 0 || host.length > 249) return null;
	const labels = host.split('.');
	if (labels.length < 2) return null;
	if (!labels.every((l) => l.length <= 63 && LABEL.test(l))) return null;
	return host;
}

/**
 * Splits pasted text (commas, spaces or new lines) into normalized domains.
 * Returns the valid ones (deduplicated, in order) and the entries that weren't valid.
 */
export function parseDomainList(text: string): { valid: string[]; invalid: string[] } {
	const valid: string[] = [];
	const invalid: string[] = [];
	for (const part of text.split(/[\s,;]+/)) {
		if (!part) continue;
		const domain = normalizeDomain(part);
		if (domain === null) invalid.push(part);
		else if (!valid.includes(domain)) valid.push(domain);
	}
	return { valid, invalid };
}

/**
 * Drops `www.` rows whose bare domain is also listed, and strips `www.` from
 * the rest, for lists saved before rows were one per site.
 */
export function collapseWww<T extends { domain: string }>(entries: T[]): T[] {
	const bare = new Set(entries.map((e) => e.domain).filter((d) => !d.startsWith('www.')));
	const result: T[] = [];
	for (const entry of entries) {
		if (!entry.domain.startsWith('www.')) {
			result.push(entry);
			continue;
		}
		const stripped = entry.domain.slice(4);
		if (!bare.has(stripped)) {
			bare.add(stripped);
			result.push({ ...entry, domain: stripped });
		}
	}
	return result;
}
