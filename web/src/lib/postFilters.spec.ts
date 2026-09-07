import { describe, expect, it } from 'vitest';
import { filterPostsByDaysBack } from './postFilters';
import type { Post } from './types';

// Only publishedAt varies across these tests - everything else is filler
// the function under test never looks at.
function makePost(id: number, publishedAt: string): Post {
	return {
		id,
		title: `Post ${id}`,
		description: 'A post.',
		imageUrl: 'https://example.com/image.jpg',
		url: 'https://example.com/post',
		feedName: 'Some Feed',
		publishedAt,
		favoritedAt: null,
		readLaterAt: null
	};
}

describe('filterPostsByDaysBack', () => {
	it('keeps a post published on the viewed day itself when daysBack is 1', () => {
		// given: a post published on the same day as the viewed issue
		const posts = [makePost(1, '2026-09-07T10:00:00Z')];

		// when: filtering to a 1-day window (just the viewed day itself)
		const result = filterPostsByDaysBack(posts, '2026-09-07', 1);

		// then: it's kept
		expect(result.map((p) => p.id)).toEqual([1]);
	});

	it('drops a post published the day before when daysBack is 1', () => {
		// given: a post published one day before the viewed issue
		const posts = [makePost(1, '2026-09-06T10:00:00Z')];

		// when: filtering to a 1-day window
		const result = filterPostsByDaysBack(posts, '2026-09-07', 1);

		// then: it's dropped - a 1-day window means just the viewed day,
		// not the viewed day plus one more day back. Confirmed live
		// (2026-09-07): the first version of this got this exact case
		// wrong, showing both days when "limited to 1 day".
		expect(result).toEqual([]);
	});

	it('keeps a post published exactly at the edge of a larger window', () => {
		// given: a post published 3 days before the viewed issue
		const posts = [makePost(1, '2026-09-04T10:00:00Z')];

		// when: filtering to a 4-day window from 2026-09-07 (the viewed day
		// plus 3 days back)
		const result = filterPostsByDaysBack(posts, '2026-09-07', 4);

		// then: the boundary is inclusive - it's kept
		expect(result.map((p) => p.id)).toEqual([1]);
	});

	it('drops a post published one day older than a larger window', () => {
		// given: a post published 4 days before the viewed issue - one day
		// further back than a 4-day window reaches
		const posts = [makePost(1, '2026-09-03T10:00:00Z')];

		// when: filtering to a 4-day window from 2026-09-07
		const result = filterPostsByDaysBack(posts, '2026-09-07', 4);

		// then: it's dropped
		expect(result).toEqual([]);
	});

	it('keeps some posts and drops others in a mixed list', () => {
		// given: posts spanning both sides of a 4-day window from 2026-09-07
		// (which reaches back to and includes 2026-09-04)
		const posts = [
			makePost(1, '2026-09-07T10:00:00Z'), // viewed day itself
			makePost(2, '2026-09-04T10:00:00Z'), // exactly at the window's edge
			makePost(3, '2026-09-03T10:00:00Z'), // one day past the edge
			makePost(4, '2026-08-01T10:00:00Z') // a new feed's old backlog
		];

		// when: filtering to a 4-day window
		const result = filterPostsByDaysBack(posts, '2026-09-07', 4);

		// then: only the two within range survive
		expect(result.map((p) => p.id)).toEqual([1, 2]);
	});

	it('is relative to the viewed day, not real-world today', () => {
		// given: a post published on a past issue's own day
		const posts = [makePost(1, '2026-01-05T10:00:00Z')];

		// when: filtering to a 2-day window while viewing 2026-01-06,
		// regardless of whatever the real current date is
		const result = filterPostsByDaysBack(posts, '2026-01-06', 2);

		// then: it's kept - the window was computed against the viewed
		// day, not against today
		expect(result.map((p) => p.id)).toEqual([1]);
	});
});
