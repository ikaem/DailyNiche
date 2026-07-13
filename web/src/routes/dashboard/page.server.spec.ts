import { describe, expect, it, vi } from 'vitest';
import type { Feed } from '$lib/types';

// See src/routes/page.server.spec.ts for why vi.hoisted + vi.mock are
// paired like this.
const { getFeeds } = vi.hoisted(() => ({ getFeeds: vi.fn() }));
vi.mock('$lib/server/api', () => ({ getFeeds }));

import { load } from './+page.server';

describe('dashboard +page.server load', () => {
	it('returns feeds and no error on success', async () => {
		// given: getFeeds resolves with an active and a disabled feed
		const feeds: Feed[] = [
			{ id: 1, name: 'Tech Blog', url: 'https://example.com/tech/feed.xml', disabledAt: null },
			{
				id: 2,
				name: 'Old Blog',
				url: 'https://example.com/old/feed.xml',
				disabledAt: '2026-05-01T00:00:00Z'
			}
		];
		getFeeds.mockResolvedValue(feeds);

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns the feeds with no error
		expect(result).toEqual({ feeds, error: null });
	});

	it('returns an empty list and an error message when the fetch fails', async () => {
		// given: getFeeds rejects
		getFeeds.mockRejectedValue(new Error('network down'));

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns an empty list with the error message, not a thrown error
		expect(result).toEqual({ feeds: [], error: 'network down' });
	});
});
