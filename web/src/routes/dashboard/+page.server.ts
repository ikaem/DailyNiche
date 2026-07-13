import { getFeeds } from '$lib/server/api';
import type { PageServerLoad } from './$types';

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
