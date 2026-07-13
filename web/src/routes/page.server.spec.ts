import { describe, expect, it, vi } from 'vitest';
import type { Post } from '$lib/types';

// vi.mock's factory is hoisted above this file's imports, so it can't
// reference a normal variable declared further down (it wouldn't exist yet
// when the factory runs). vi.hoisted's own callback is hoisted too, and
// runs first, so its return value IS available inside vi.mock's factory.
// Together: every import of getPostsToday from $lib/server/api - including
// the one inside +page.server.ts below - resolves to this one shared fake.
const { getPostsToday } = vi.hoisted(() => ({ getPostsToday: vi.fn() }));
vi.mock('$lib/server/api', () => ({ getPostsToday }));

import { load } from './+page.server';

describe('+page.server load', () => {
	it('returns posts and no error on success', async () => {
		// given: getPostsToday resolves with one post
		const posts: Post[] = [
			{
				id: 1,
				title: 'Go 2.0 Announced',
				description: 'The Go team announces the next major version.',
				imageUrl: '',
				url: 'https://example.com/go-2-0-announced',
				feedName: 'Tech Blog',
				publishedAt: '2026-07-10T09:00:00Z'
			}
		];
		getPostsToday.mockResolvedValue(posts);

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns the posts with no error
		expect(result).toEqual({ posts, error: null });
	});

	it('returns an empty list and an error message when the fetch fails', async () => {
		// given: getPostsToday rejects
		getPostsToday.mockRejectedValue(new Error('network down'));

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns an empty list with the error message, not a thrown error
		expect(result).toEqual({ posts: [], error: 'network down' });
	});
});
