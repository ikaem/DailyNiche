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
	it('keeps a post published on the viewed day itself', () => {
		// given: a post published on the same day as the viewed issue
		const posts = [makePost(1, '2026-09-07T10:00:00Z')];

		// when: filtering to 0 days back from that same day
		const result = filterPostsByDaysBack(posts, '2026-09-07', 0);

		// then: it's kept
		expect(result.map((p) => p.id)).toEqual([1]);
	});

	it('keeps a post published exactly daysBack days before the viewed day', () => {
		// given: a post published exactly 4 days before the viewed issue
		const posts = [makePost(1, '2026-09-03T10:00:00Z')];

		// when: filtering to 4 days back from 2026-09-07
		const result = filterPostsByDaysBack(posts, '2026-09-07', 4);

		// then: the boundary is inclusive - it's kept
		expect(result.map((p) => p.id)).toEqual([1]);
	});

	it('drops a post published one day older than the cutoff', () => {
		// given: a post published 5 days before the viewed issue
		const posts = [makePost(1, '2026-09-02T10:00:00Z')];

		// when: filtering to 4 days back from 2026-09-07
		const result = filterPostsByDaysBack(posts, '2026-09-07', 4);

		// then: it's dropped
		expect(result).toEqual([]);
	});

	it('keeps some posts and drops others in a mixed list', () => {
		// given: posts spanning both sides of a 4-day cutoff from 2026-09-07
		const posts = [
			makePost(1, '2026-09-07T10:00:00Z'), // viewed day itself
			makePost(2, '2026-09-03T10:00:00Z'), // exactly at the cutoff
			makePost(3, '2026-09-02T10:00:00Z'), // one day past the cutoff
			makePost(4, '2026-08-01T10:00:00Z') // a new feed's old backlog
		];

		// when: filtering to 4 days back
		const result = filterPostsByDaysBack(posts, '2026-09-07', 4);

		// then: only the two within range survive
		expect(result.map((p) => p.id)).toEqual([1, 2]);
	});

	it('is relative to the viewed day, not real-world today', () => {
		// given: a post published on a past issue's own day
		const posts = [makePost(1, '2026-01-05T10:00:00Z')];

		// when: filtering to 2 days back while viewing that past issue
		// (2026-01-05), regardless of whatever the real current date is
		const result = filterPostsByDaysBack(posts, '2026-01-06', 2);

		// then: it's kept - the cutoff was computed against the viewed
		// day, not against today
		expect(result.map((p) => p.id)).toEqual([1]);
	});
});
