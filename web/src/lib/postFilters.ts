import type { Post } from './types';
import { addDaysUTC } from './dateUtils';

// filterPostsByDaysBack keeps only posts whose real publish date falls
// within a `daysBack`-day window ending on the viewed issue (inclusive of
// that day), dropping older ones - the fix for a newly-added feed's first
// fetch dumping its entire back-catalog into one day's view. Relative to
// `viewedDate` (the issue being browsed), not real-world "today", so
// paging to a past issue with this filter on still shows that issue's own
// recent history, rather than being computed against today's date and
// showing nothing.
//
// daysBack is the *size* of that window, not how far back from
// viewedDate to start additionally counting - daysBack: 1 means "just
// viewedDate itself", not "viewedDate plus one more day back". Confirmed
// live (2026-09-07): the first version of this used `-daysBack` directly,
// so "limit to 1 day" actually showed a 2-day window (today + yesterday) -
// an off-by-one against what the label promises.
export function filterPostsByDaysBack(posts: Post[], viewedDate: string, daysBack: number): Post[] {
	const cutoff = addDaysUTC(viewedDate, -(daysBack - 1));
	return posts.filter((post) => post.publishedAt.slice(0, 10) >= cutoff);
}
