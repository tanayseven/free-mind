import { describe, it, expect } from 'vitest';
import { collapseWww, normalizeDomain, parseDomainList } from './domains';

describe('normalizeDomain', () => {
	it('accepts bare domains and lowercases them', () => {
		expect(normalizeDomain('Reddit.com')).toBe('reddit.com');
		expect(normalizeDomain('news.ycombinator.com')).toBe('news.ycombinator.com');
	});

	it('strips scheme, www, path, port and trailing dot', () => {
		expect(normalizeDomain('https://www.youtube.com/watch?v=1')).toBe('youtube.com');
		expect(normalizeDomain('http://example.com:8080/')).toBe('example.com');
		expect(normalizeDomain('bbc.co.uk.')).toBe('bbc.co.uk');
	});

	it('rejects things that are not domains', () => {
		expect(normalizeDomain('')).toBeNull();
		expect(normalizeDomain('localhost')).toBeNull();
		expect(normalizeDomain('bad_domain.com')).toBeNull();
		expect(normalizeDomain('-x.com')).toBeNull();
		expect(normalizeDomain('a..com')).toBeNull();
	});
});

describe('parseDomainList', () => {
	it('splits on commas, spaces and new lines and deduplicates', () => {
		expect(parseDomainList('x.com, reddit.com\nwww.x.com  tiktok.com')).toEqual({
			valid: ['x.com', 'reddit.com', 'tiktok.com'],
			invalid: []
		});
	});

	it('reports invalid entries', () => {
		expect(parseDomainList('x.com nope')).toEqual({ valid: ['x.com'], invalid: ['nope'] });
	});
});

describe('collapseWww', () => {
	it('drops www rows that duplicate a bare domain', () => {
		const entries = [{ domain: 'x.com' }, { domain: 'www.x.com' }, { domain: 'www.y.com' }];
		expect(collapseWww(entries)).toEqual([{ domain: 'x.com' }, { domain: 'y.com' }]);
	});
});
