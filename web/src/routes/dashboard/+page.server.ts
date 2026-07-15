import { fail } from '@sveltejs/kit';
import { addFeed, ApiError, getFeeds } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

// Runs only on the server - same reasoning as the home page's load: no
// CORS, no client-side fetch. Errors are returned as data (not thrown) so
// +page.svelte can show an inline message without losing the header/nav.
export const load: PageServerLoad = async () => {
	try {
		const feeds = await getFeeds();
		return { feeds, error: null };
	} catch (err) {
		return { feeds: [], error: err instanceof Error ? err.message : 'Failed to load feeds' };
	}
};

// Mirrors the Go API's own validation messages (see feeds_handler.go) -
// checked here too so an obviously-invalid submission never wastes a
// request to the Go API, and so this action still validates even if the Go
// API's own checks ever change or are removed. Go's validation remains the
// final authority regardless - this is a defensive second layer, not a
// replacement.
function isValidAbsoluteUrl(value: string): boolean {
	try {
		new URL(value);
		return true;
	} catch {
		return false;
	}
}

export const actions: Actions = {
	// Returns nothing on success - use:enhance's default behavior already
	// resets the form and calls invalidateAll() for any successful action
	// result, whether or not data is returned (verified against
	// @sveltejs/kit's own enhance()/invalidateAll() source).
	addFeed: async ({ request }) => {
		const formData = await request.formData();
		const name = String(formData.get('name') ?? '').trim();
		const url = String(formData.get('url') ?? '').trim();

		if (!name) {
			return fail(400, { message: 'name is required' });
		}
		if (!url) {
			return fail(400, { message: 'url is required' });
		}
		if (!isValidAbsoluteUrl(url)) {
			return fail(400, { message: 'url must be a valid absolute URL' });
		}

		try {
			await addFeed(name, url);
		} catch (err) {
			if (err instanceof ApiError) {
				return fail(err.status, { message: err.message });
			}
			return fail(500, { message: 'Failed to add feed' });
		}
	}
};
