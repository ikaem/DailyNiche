import type { Post } from './types';
import { addDaysUTC } from './dateUtils';

// filterPostsByDaysBack keeps only posts whose real publish date falls
// within `daysBack` days of the viewed issue (inclusive), dropping older
// ones - the fix for a newly-added feed's first fetch dumping its entire
// back-catalog into one day's view. Relative to `viewedDate` (the issue
// being browsed), not real-world "today", so paging to a past issue with
// this filter on still shows that issue's own recent history, rather than
// being computed against today's date and showing nothing.
export function filterPostsByDaysBack(posts: Post[], viewedDate: string, daysBack: number): Post[] {
	const cutoff = addDaysUTC(viewedDate, -daysBack);
	return posts.filter((post) => post.publishedAt.slice(0, 10) >= cutoff);
}
