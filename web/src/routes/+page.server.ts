import { getPostsByDate, getPostsToday } from '$lib/server/api';
import {
	favoritePostAction,
	markReadLaterAction,
	unfavoritePostAction,
	unmarkReadLaterAction
} from '$lib/server/savedPostActions';
import type { Actions, PageServerLoad } from './$types';

// Today's date in UTC as YYYY-MM-DD, matching this app's established
// UTC-everywhere convention (see CLAUDE.md's Timestamps & Timezones note) -
// so DateNav always has a concrete "current date" to compute prev/next from,
// rather than needing to separately guess what "today" means.
function todayUTC(): string {
	return new Date().toISOString().slice(0, 10);
}

// Runs only on the server - the browser never talks to the Go API directly,
// so it's never subject to CORS. Errors are returned as data (not thrown)
// so +page.svelte can show an inline message without losing the header/nav.
//
// A ?date=YYYY-MM-DD query param selects a specific day's issue (set by
// DateNav's prev/next/date-input navigation); without one, this is the
// homepage's default "today" view, unchanged from before - still calling
// getPostsToday(), not getPostsByDate(todayUTC()), to keep that exact
// existing call path untouched.
export const load: PageServerLoad = async ({ url }) => {
	const dateParam = url.searchParams.get('date');
	const date = dateParam ?? todayUTC();
	try {
		const posts = dateParam ? await getPostsByDate(dateParam) : await getPostsToday();
		return { posts, error: null, date };
	} catch (err) {
		return {
			posts: [],
			error: err instanceof Error ? err.message : 'Failed to load posts',
			date
		};
	}
};

// All four are shared with the Saved page (see $lib/server/savedPostActions)
// - the actual logic lives there since SvelteKit requires each route to
// export its own `actions`, even when the implementation is identical.
export const actions: Actions = {
	favoritePost: favoritePostAction,
	unfavoritePost: unfavoritePostAction,
	markReadLater: markReadLaterAction,
	unmarkReadLater: unmarkReadLaterAction
};
