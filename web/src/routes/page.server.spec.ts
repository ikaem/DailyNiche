import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Post } from '$lib/types';

// vi.mock's factory is hoisted above this file's imports, so it can't
// reference a normal variable declared further down (it wouldn't exist yet
// when the factory runs). vi.hoisted's own callback is hoisted too, and
// runs first, so its return value IS available inside vi.mock's factory.
// Together: every import of getPostsToday/getPostsByDate from
// $lib/server/api - including the one inside +page.server.ts below -
// resolves to these same shared fakes.
const { getPostsToday, getPostsByDate } = vi.hoisted(() => ({
	getPostsToday: vi.fn(),
	getPostsByDate: vi.fn()
}));
vi.mock('$lib/server/api', () => ({ getPostsToday, getPostsByDate }));

import { load } from './+page.server';

// Builds a fake LoadEvent carrying just the "url" property load() actually
// reads - everything else on the real LoadEvent is irrelevant here.
function loadEvent(path: string): Parameters<typeof load>[0] {
	return { url: new URL(`http://localhost${path}`) } as Parameters<typeof load>[0];
}

describe('+page.server load', () => {
	beforeEach(() => {
		// Freezes "today" so the date-less path resolves to a known, exact
		// date instead of whatever day the test happens to run on.
		vi.setSystemTime(new Date('2026-07-15T12:00:00Z'));
		// getPostsToday/getPostsByDate are shared vi.fn()s across every test
		// in this file (via vi.hoisted) - without clearing, call counts from
		// earlier tests leak into later ones.
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

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

		// when: the load function runs with no ?date param
		const result = await load(loadEvent('/'));

		// then: it returns the posts, no error, and today's date
		expect(result).toEqual({ posts, error: null, date: '2026-07-15' });
		expect(getPostsByDate).not.toHaveBeenCalled();
	});

	it('returns an empty list and an error message when the fetch fails', async () => {
		// given: getPostsToday rejects
		getPostsToday.mockRejectedValue(new Error('network down'));

		// when: the load function runs with no ?date param
		const result = await load(loadEvent('/'));

		// then: it returns an empty list with the error message, not a thrown error
		expect(result).toEqual({ posts: [], error: 'network down', date: '2026-07-15' });
	});

	it('fetches a specific date when ?date is present, instead of today', async () => {
		// given: getPostsByDate resolves with one post for a specific date
		const posts: Post[] = [
			{
				id: 2,
				title: 'Archived Post',
				description: 'From a previous day.',
				imageUrl: '',
				url: 'https://example.com/archived-post',
				feedName: 'Tech Blog',
				publishedAt: '2026-07-09T09:00:00Z'
			}
		];
		getPostsByDate.mockResolvedValue(posts);

		// when: the load function runs with ?date=2026-07-09
		const result = await load(loadEvent('/?date=2026-07-09'));

		// then: it calls getPostsByDate with that exact date, not getPostsToday
		expect(getPostsByDate).toHaveBeenCalledWith('2026-07-09');
		expect(getPostsToday).not.toHaveBeenCalled();
		expect(result).toEqual({ posts, error: null, date: '2026-07-09' });
	});

	it('returns an empty list and an error message when fetching a specific date fails', async () => {
		// given: getPostsByDate rejects
		getPostsByDate.mockRejectedValue(new Error('network down'));

		// when: the load function runs with ?date=2026-07-09
		const result = await load(loadEvent('/?date=2026-07-09'));

		// then: it returns an empty list with the error message and the requested date
		expect(result).toEqual({ posts: [], error: 'network down', date: '2026-07-09' });
	});
});
