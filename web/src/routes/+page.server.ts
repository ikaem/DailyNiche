import { fail } from '@sveltejs/kit';
import {
	ApiError,
	favoritePost,
	getPostsByDate,
	getPostsToday,
	markReadLater,
	unfavoritePost,
	unmarkReadLater
} from '$lib/server/api';
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

// Shared by all four actions below - a local, private helper (not exported
// or shared across routes) since all four need identical id parsing,
// matching how the dashboard's own actions validate a form field inline.
function parsePostId(formData: FormData): number | null {
	const raw = String(formData.get('id') ?? '');
	const id = Number(raw);
	if (!raw || Number.isNaN(id)) {
		return null;
	}
	return id;
}

export const actions: Actions = {
	// All four return nothing on success - use:enhance's default behavior
	// already calls invalidateAll() for any successful result, which reruns
	// load() and reflects the new saved state, same pattern the dashboard's
	// addFeed/deleteFeed/etc. actions already use.
	favoritePost: async ({ request }) => {
		const id = parsePostId(await request.formData());
		if (id === null) {
			return fail(400, { message: 'a valid post id is required' });
		}
		try {
			await favoritePost(id);
		} catch (err) {
			if (err instanceof ApiError) {
				return fail(err.status, { message: err.message });
			}
			return fail(500, { message: 'Failed to favorite post' });
		}
	},

	unfavoritePost: async ({ request }) => {
		const id = parsePostId(await request.formData());
		if (id === null) {
			return fail(400, { message: 'a valid post id is required' });
		}
		try {
			await unfavoritePost(id);
		} catch (err) {
			if (err instanceof ApiError) {
				return fail(err.status, { message: err.message });
			}
			return fail(500, { message: 'Failed to unfavorite post' });
		}
	},

	markReadLater: async ({ request }) => {
		const id = parsePostId(await request.formData());
		if (id === null) {
			return fail(400, { message: 'a valid post id is required' });
		}
		try {
			await markReadLater(id);
		} catch (err) {
			if (err instanceof ApiError) {
				return fail(err.status, { message: err.message });
			}
			return fail(500, { message: 'Failed to mark post read later' });
		}
	},

	unmarkReadLater: async ({ request }) => {
		const id = parsePostId(await request.formData());
		if (id === null) {
			return fail(400, { message: 'a valid post id is required' });
		}
		try {
			await unmarkReadLater(id);
		} catch (err) {
			if (err instanceof ApiError) {
				return fail(err.status, { message: err.message });
			}
			return fail(500, { message: 'Failed to unmark post read later' });
		}
	}
};
